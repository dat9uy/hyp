package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// eventsJSON returns the applied events hyp events --json prints.
func eventsJSON(t *testing.T) []map[string]any {
	t.Helper()
	res := hyp(t, "events", "--json")
	if res.code != 0 {
		t.Fatalf("hyp events --json: exit %d, %s", res.code, res.stderr)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(res.stdout), &doc); err != nil {
		t.Fatalf("parse events json: %v\n%s", err, res.stdout)
	}
	var out []map[string]any
	for _, v := range doc["events"].([]any) {
		out = append(out, v.(map[string]any))
	}
	return out
}

func sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// recordPins lists, for every node.add that pins records, path -> sha256.
func recordPins(t *testing.T) [][2]string {
	var pins [][2]string
	for _, ev := range eventsJSON(t) {
		p, _ := ev["payload"].(map[string]any)
		recs, _ := p["records"].([]any)
		for _, r := range recs {
			m := r.(map[string]any)
			pins = append(pins, [2]string{m["path"].(string), m["sha256"].(string)})
		}
	}
	return pins
}

// Probe for line 5: two clones each write an event naming the same
// store-relative path with other bytes. After hyp sync each side must hold
// every named file of both, so each event's pinned sha256 must match the
// bytes hyp show prints for its path, in both clones.
func TestProbeSyncKeepsEachSidesNamedBytes(t *testing.T) {
	gitConfig(t)
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	remote := filepath.Join(top, "remote.git")
	gitIn(t, top, "init", "-q", "--bare", remote)
	a := filepath.Join(top, "a")
	gitIn(t, top, "clone", "-q", remote, a)
	writeFile(t, filepath.Join(a, "s", "README.md"), "a store\n")
	gitIn(t, a, "add", "-A")
	gitIn(t, a, "commit", "-qm", "base")
	gitIn(t, a, "push", "-q", "origin", "main")
	t.Chdir(filepath.Join(a, "s"))
	must(t, "init")
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	want(t, hyp(t, "sync"), 0, "synced")

	b := filepath.Join(top, "b")
	gitIn(t, top, "clone", "-q", remote, b)
	t.Chdir(filepath.Join(b, "s"))
	want(t, hyp(t, "sync"), 0, "synced")

	shared := ".tree/evidence/shared/in.txt"
	t.Chdir(filepath.Join(a, "s"))
	writeFile(t, shared, "bytes of a\n")
	must(t, "add", "exp", "--parent", h, "--title", "EA", "--recipe", "none", "--rubric", "reading", "--records", shared, "--why", "w")
	t.Chdir(filepath.Join(b, "s"))
	writeFile(t, shared, "bytes of b\n")
	must(t, "add", "exp", "--parent", h, "--title", "EB", "--recipe", "none", "--rubric", "reading", "--records", shared, "--why", "w")

	t.Chdir(filepath.Join(a, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	t.Chdir(filepath.Join(b, "s"))
	sb := hyp(t, "sync")
	t.Logf("b sync stderr: %q", sb.stderr)
	t.Chdir(filepath.Join(a, "s"))
	want(t, hyp(t, "sync"), 0, "synced")

	for _, side := range []string{a, b} {
		t.Chdir(filepath.Join(side, "s"))
		shown := hyp(t, "show", shared)
		if shown.code != 0 {
			t.Fatalf("%s: hyp show %s: exit %d %s", side, shared, shown.code, shown.stderr)
		}
		for _, pin := range recordPins(t) {
			if pin[0] == shared && pin[1] != sum(shown.stdout) {
				t.Errorf("%s: an event pins %s at sha256 %s, but the record holds bytes %q (sha256 %s): one side's named file is lost by sync",
					filepath.Base(side), shared, pin[1][:12], shown.stdout, sum(shown.stdout)[:12])
			}
		}
	}
}

// Probe for the design key "an evidence pointer stores its file's sha256":
// the sha256 hyp record stores must be the sha256 of the bytes the record
// keeps for that pointer. hyp record hashes the evidence before it takes
// the lock and reads the bytes again in the write.
func TestProbeEvidenceSHA256MatchesRecordedBytes(t *testing.T) {
	recordRepo(t)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
	ptr := ".tree/evidence/" + e + "/out.txt"
	writeFile(t, ptr, "hashed bytes\n")
	recordHashed = func() { writeFile(t, ptr, "bytes written meanwhile\n") }
	t.Cleanup(func() { recordHashed = nil })
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
	recordHashed = nil
	var stored string
	for _, ev := range eventsJSON(t) {
		if ev["type"] != "result.record" {
			continue
		}
		p := ev["payload"].(map[string]any)
		if m, ok := p["evidence_sha256"].(map[string]any); ok {
			stored, _ = m[ptr].(string)
		}
	}
	shown := hyp(t, "show", ptr)
	if shown.code != 0 {
		t.Fatalf("hyp show %s: exit %d %s", ptr, shown.code, shown.stderr)
	}
	if stored != sum(shown.stdout) {
		t.Fatalf("%s stores evidence_sha256 %s for %s, but the record keeps bytes %q with sha256 %s",
			r, stored, ptr, shown.stdout, sum(shown.stdout))
	}
}

// Probe (no line): re-tying a Hypothesis subject to an edited patch at the
// same path. The base accepts it (the subject is named by sha256); the
// candidate refuses it in a record repository.
func TestProbeSubjectRetieAfterPatchEdit(t *testing.T) {
	repo := recordRepo(t)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	base := gitIn(t, repo, "rev-parse", "HEAD")
	writeFile(t, "arms/a.patch", "first\n")
	must(t, "subject", h, "--base", base, "--patch", "arms/a.patch", "--why", "w")
	writeFile(t, "arms/a.patch", "second\n")
	res := hyp(t, "subject", h, "--base", base, "--patch", "arms/a.patch", "--why", "w")
	if res.code != 0 {
		t.Fatalf("re-tie after editing the patch: exit %d, stderr %q", res.code, strings.TrimSpace(res.stderr))
	}
}
