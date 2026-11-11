package cli

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
)

// A subject has one shape: a base commit, plus what is given over it
// (ADR 0020). These tests hold the rules of G-D1CH1R line 3 in a store kept
// in a git repository, so a Hypothesis's --base names real commits.

// gitCommits makes dir a git repository with n commits and returns their
// full ids, oldest first.
func gitCommits(t *testing.T, dir string, n int) []string {
	t.Helper()
	run := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "--quiet")
	var ids []string
	for i := range n {
		run("commit", "--quiet", "--allow-empty", "-m", "c"+string(rune('0'+i)))
		ids = append(ids, run("rev-parse", "HEAD"))
	}
	return ids
}

// writeSandboxRecord writes a sandbox.json for one run at base, given these
// patches over it, under the Experiment's evidence directory, and returns
// its pointer.
func writeSandboxRecord(t *testing.T, dir, exp, run, base string, patches ...string) string {
	t.Helper()
	var ps []map[string]string
	for i, p := range patches {
		ps = append(ps, map[string]string{"arm": "arm" + string(rune('a'+i)), "sha256": p})
	}
	data, _ := json.Marshal(map[string]any{"format": 1, "id": run, "base": base, "patches": ps})
	rel := filepath.Join(".tree", "evidence", exp, run, "sandbox.json")
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(rel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, rel), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(rel)
}

// writeHashed writes body to name under dir and returns its sha256.
func writeHashed(t *testing.T, dir, name, body string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

// lastResult reads the payload of the store's latest result.record.
func lastResult(t *testing.T, dir string) map[string]json.RawMessage {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(dir, ".tree", "events", "*.jsonl"))
	var last map[string]json.RawMessage
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		sc := bufio.NewScanner(fh)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			var ev event.Event
			if json.Unmarshal(sc.Bytes(), &ev) != nil || ev.Type != event.TypeResultRecord {
				continue
			}
			var p map[string]json.RawMessage
			_ = json.Unmarshal(ev.Payload, &p)
			last = p
		}
		fh.Close()
	}
	return last
}

// candidate adds a supported Hypothesis under q, tied to this subject.
func candidate(t *testing.T, q, title string, subject ...string) string {
	t.Helper()
	h := supported(t, q, title)
	must(t, append(append([]string{"subject", h}, subject...), "--why", "w")...)
	return h
}

// runner adds a Hypothesis that the run under test judges, and returns the
// Experiment that runs the line's case on it.
func runner(t *testing.T, q string) (h, e string) {
	t.Helper()
	h = must(t, "add", "hyp", "--parent", q, "--title", "runner", "--why", "w")
	e = must(t, "add", "exp", "--parent", h, "--title", "runs the line", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	return h, e
}

// R3a. hyp record keeps one subject per cited run, {base, files}, in
// evidence order: a run at B given patch P and a run at C are two subjects,
// never one merged list.
func TestResultKeepsOneSubjectPerCitedRun(t *testing.T) {
	dir, _, q, keys := setupLines(t, "a run at B plus a patch never covers bare B")
	ids := gitCommits(t, dir, 2)
	b, c := ids[0], ids[1]
	p := writeHashed(t, dir, "arm.patch", "--- a\n+++ b\n")
	h, e := runner(t, q)
	r1 := writeSandboxRecord(t, dir, e, "run-b", b, p)
	r2 := writeSandboxRecord(t, dir, e, "run-c", c)
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "two runs", "--evidence", r1, "--evidence", r2, "--closes", keys[0], "--why", "w")
	got := lastResult(t, dir)
	if _, ok := got["fingerprints"]; ok {
		t.Fatalf("hyp record wrote fingerprints %s; a Result keeps one subject per cited run", got["fingerprints"])
	}
	var subjects []event.RunSubject
	_ = json.Unmarshal(got["subjects"], &subjects)
	wantSubjects := []event.RunSubject{{Base: b, Files: []string{p}}, {Base: c}}
	if !slices.EqualFunc(subjects, wantSubjects, func(x, y event.RunSubject) bool { return x.Base == y.Base && slices.Equal(x.Files, y.Files) }) {
		t.Fatalf("subjects = %+v, want one per cited run %+v", subjects, wantSubjects)
	}
	want(t, hyp(t, "check"), 0, "check: 0 errors")
}

// R3b. A Hypothesis's subject may name a base. A run at another base never
// covers it; a run at its base does.
func TestBaseSubjectIsCoveredOnlyAtItsBase(t *testing.T) {
	dir, g, q, keys := setupLines(t, "the commit runs the case")
	ids := gitCommits(t, dir, 2)
	b, c := ids[0], ids[1]
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	bad := must(t, "add", "hyp", "--parent", q2, "--title", "bad", "--why", "w")
	want(t, hyp(t, "subject", bad, "--base", "no-such-commit", "--why", "w"), 2, "--base no-such-commit is not a commit")
	hc := candidate(t, q2, "commit C", "--base", "HEAD")
	want(t, hyp(t, "show", hc), 0, "subject: base "+c[:12]+"\n")

	h, e := runner(t, q)
	atB := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran at B", "--evidence", writeSandboxRecord(t, dir, e, "run-b", b), "--closes", keys[0], "--why", "w")
	want(t, hyp(t, "status"), 0, "uncovered on "+q2+": "+hc)

	h2, e2 := runner(t, q)
	res := must(t, "record", "--exp", e2, "--verdict", h2+"=supports", "--observation", "ran at C", "--evidence", writeSandboxRecord(t, dir, e2, "run-c", c), "--closes", keys[0], "--why", "w")
	want(t, hyp(t, "show", keys[0]), 0, "ran in "+atB+": ran on "+h+": supports\n")
	want(t, hyp(t, "show", keys[0]), 0, "ran in "+res+": ran on "+h2+": supports; ran on "+hc+"\n")
	notIn(t, "uncovered on "+q2)
}

// R3c. An old Result, written with one merged fingerprints list, reads as
// one subject with no base: it still covers a Hypothesis whose files it
// holds, as before ADR 0020.
func TestOldResultReadsAsOneSubjectWithNoBase(t *testing.T) {
	dir, g, q, keys := setupLines(t, "the file runs the case")
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	sum := writeHashed(t, dir, "h-row.py", "row = 1\n")
	h2 := candidate(t, q2, "the file", "h-row.py")
	h, e := runner(t, q)
	res := rawResult(t, "a", map[string]any{"parent": fullID(t, e), "verdicts": map[string]string{fullID(t, h): "supports"}, "observation": "old",
		"closes": []string{fullID(t, keys[0][:strings.Index(keys[0], "#")]) + keys[0][strings.Index(keys[0], "#"):]}, "fingerprints": []string{sum}})
	want(t, hyp(t, "check"), 0, "check: 0 errors")
	want(t, hyp(t, "show", keys[0]), 0, "ran in R-"+res[len(res)-6:]+": ran on "+h+": supports; ran on "+h2+"\n")
	notIn(t, "uncovered on "+q2)
}

// R3d. A run at B plus a patch never covers a Hypothesis whose subject is
// bare B; it covers one whose subject is B and that patch.
func TestPatchedRunNeverCoversBareBase(t *testing.T) {
	dir, g, q, keys := setupLines(t, "a run at B plus a patch never covers bare B")
	b := gitCommits(t, dir, 1)[0]
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	p := writeHashed(t, dir, "arm.patch", "--- a\n+++ b\n")
	bare := candidate(t, q2, "bare B", "--base", b)
	patched := candidate(t, q2, "B and the patch", "arm.patch", "--base", b)
	h, e := runner(t, q)
	res := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran at B with the patch", "--evidence", writeSandboxRecord(t, dir, e, "run-bp", b, p), "--closes", keys[0], "--why", "w")
	want(t, hyp(t, "status"), 0, "uncovered on "+q2+": "+bare+";")
	want(t, hyp(t, "show", keys[0]), 0, "ran in "+res+": ran on "+h+": supports; ran on "+patched+"\n")
}

// A sandbox record holds one run per Arm, so it keeps one subject per Arm,
// each at its base and given that Arm's one patch: a Hypothesis tied to B
// and Arm a's patch is covered, and bare B is not.
func TestSandboxKeepsOneSubjectPerArm(t *testing.T) {
	dir, g, q, keys := setupLines(t, "each Arm runs the case")
	b := gitCommits(t, dir, 1)[0]
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	pa := writeHashed(t, dir, "a.patch", "--- a\n+++ a\n")
	pb := writeHashed(t, dir, "b.patch", "--- b\n+++ b\n")
	armA := candidate(t, q2, "B and Arm a", "a.patch", "--base", b)
	bare := candidate(t, q2, "bare B", "--base", b)
	h, e := runner(t, q)
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "two Arms", "--evidence", writeSandboxRecord(t, dir, e, "run-ab", b, pa, pb), "--closes", keys[0], "--why", "w")
	var subjects []event.RunSubject
	_ = json.Unmarshal(lastResult(t, dir)["subjects"], &subjects)
	if len(subjects) != 2 || subjects[0].Base != b || !slices.Equal(subjects[0].Files, []string{pa}) || !slices.Equal(subjects[1].Files, []string{pb}) {
		t.Fatalf("subjects = %+v, want one per Arm at %s", subjects, b)
	}
	want(t, hyp(t, "status"), 0, "uncovered on "+q2+": "+bare+";")
	notIn(t, armA+";")
}

// Rule 26 on a base: a bare-B subject that only a run at B plus a patch met
// never ran, so it may be tied again; once a run at exactly B holds it, a
// different subject is a new Hypothesis.
func TestBaseSubjectLocksOnlyWhenItRan(t *testing.T) {
	dir, _, q, _ := setupLines(t, "the commit runs the case")
	ids := gitCommits(t, dir, 3)
	p := writeHashed(t, dir, "arm.patch", "--- a\n+++ b\n")
	h := candidate(t, q, "commit", "--base", ids[0])
	e := must(t, "add", "exp", "--parent", h, "--title", "patched", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "B plus a patch", "--evidence", writeSandboxRecord(t, dir, e, "run-bp", ids[0], p), "--why", "w")
	must(t, "subject", h, "--base", ids[1], "--why", "w")
	e2 := must(t, "add", "exp", "--parent", h, "--title", "bare", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	res := must(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "exactly B", "--evidence", writeSandboxRecord(t, dir, e2, "run-b", ids[1]), "--why", "w")
	want(t, hyp(t, "subject", h, "--base", ids[2], "--why", "w"), 1, "bad subject: "+h+"'s subject ran in "+res)
}

// A Hypothesis tied with a base and two paths of the same bytes is held by
// a run given those bytes once.
func TestBaseSubjectWithOneFileTwiceIsHeld(t *testing.T) {
	dir, g, q, keys := setupLines(t, "the patch runs the case")
	b := gitCommits(t, dir, 1)[0]
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	p := writeHashed(t, dir, "one.patch", "--- a\n+++ b\n")
	writeHashed(t, dir, "copy.patch", "--- a\n+++ b\n")
	twice := candidate(t, q2, "two paths", "one.patch", "copy.patch", "--base", b)
	h, e := runner(t, q)
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "B plus the patch", "--evidence", writeSandboxRecord(t, dir, e, "run-bp", b, p), "--closes", keys[0], "--why", "w")
	notIn(t, "uncovered on "+q2)
	_ = twice
}

// git missing from PATH is an environment problem, exit 3; a rev git cannot
// resolve is a bad argument, exit 2.
func TestSubjectBaseExitCodes(t *testing.T) {
	dir, _, q, _ := setupLines(t, "x")
	gitCommits(t, dir, 1)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	want(t, hyp(t, "subject", h, "--base", "no-such", "--why", "w"), 2, "--base no-such is not a commit")
	t.Setenv("PATH", t.TempDir())
	want(t, hyp(t, "subject", h, "--base", "HEAD", "--why", "w"), 3, "--base HEAD: git is not on PATH")
}
