package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// G-SZCBXZ#1: a write that exits 0 keeps every earlier event. A record
// write stages its files with `git update-index --index-info`, one
// "mode sha\tpath" line per file, without -z. An evidence path that holds a
// newline therefore adds a second index line of the path's choosing. Here
// the cited file's path holds a newline followed by the line that stages
// the cited file's own blob at s/.tree/events/a.jsonl, so the write
// replaces the store's events file in the record, and every earlier event,
// the Goal and the first Result among them, is gone from hyp show and hyp
// status in every worktree.
func TestReviewEvidencePathWithANewlineRewritesTheEventsFile(t *testing.T) {
	repo := recordRepo(t)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
	first := ".tree/evidence/" + e + "/first.txt"
	writeFile(t, first, "first run\n")
	r1 := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "first ran", "--evidence", first, "--why", "w")
	want(t, hyp(t, "show", g), 0, "G")

	content := "not an event\n"
	cmd := exec.Command("git", "hash-object", "--stdin")
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.TrimSpace(string(out))
	dir := ".tree/evidence/" + e + "/x\n100644 " + sha + "\ts"
	crafted := dir + "/.tree/events/a.jsonl"
	writeFile(t, crafted, content)

	before := gitIn(t, repo, "show", "refs/hyp/record:s/.tree/events/a.jsonl")
	// The write runs in its own process: hyp panics after it in this case.
	rec := exec.Command(os.Args[0], "record", "--exp", e, "--verdict", h+"=supports", "--observation", "second ran", "--evidence", crafted, "--why", "w")
	rec.Dir = filepath.Join(repo, "s")
	rec.Env = append(os.Environ(), "HYP_TEST_AS_HYP=1", "HYP_ACTOR=a")
	recOut, recErr := rec.CombinedOutput()
	t.Logf("second hyp record: err %v\n%s", recErr, firstLines(string(recOut), 3))

	after := gitIn(t, repo, "show", "refs/hyp/record:s/.tree/events/a.jsonl")
	if !strings.HasPrefix(after, strings.TrimSpace(before)) {
		t.Errorf("a write whose evidence path holds a newline replaced the record's events file:\nbefore %d bytes, after %q", len(before), after)
	}
	if r := hyp(t, "show", g); r.code != 0 {
		t.Errorf("after that write, hyp show %s: exit %d, %q; want the Goal", g, r.code, strings.TrimSpace(r.stderr))
	}
	if r := hyp(t, "show", r1); r.code != 0 || !strings.Contains(r.stdout, "first ran") {
		t.Errorf("after that write, hyp show %s: exit %d, %q; want the first Result", r1, r.code, strings.TrimSpace(r.stderr))
	}
	if r := hyp(t, "status"); !strings.Contains(r.stdout, g) {
		t.Errorf("after that write, hyp status no longer shows %s:\n%s", g, firstLines(r.stdout+r.stderr, 3))
	}
}

// none (closest G-SZCBXZ#1): a work-tree store takes an evidence file whose
// name holds a newline; the record refuses it with exit 3, an environment
// code, and git's own message, not a rule refusal that names the path.
func TestReviewEvidenceNameWithANewlineExitsThree(t *testing.T) {
	recordRepo(t)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
	p := ".tree/evidence/" + e + "/a\nb.txt"
	writeFile(t, p, "bytes\n")
	r := hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", p, "--why", "w")
	if r.code != 0 {
		t.Errorf("hyp record citing a file whose name holds a newline: exit %d, stderr %q; a work-tree store takes it", r.code, strings.TrimSpace(r.stderr))
	}
}

func firstLines(s string, n int) string {
	l := strings.SplitN(s, "\n", n+1)
	if len(l) > n {
		l = l[:n]
	}
	return strings.Join(l, "\n")
}
