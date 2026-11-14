package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// pinsOf lists the sha256 each applied event pins for the records file rel.
func pinsOf(t *testing.T, rel string) []string {
	t.Helper()
	res := hyp(t, "events", "--json")
	if res.code != 0 {
		t.Fatalf("hyp events --json: exit %d, %s", res.code, res.stderr)
	}
	var doc struct {
		Events []struct {
			Payload struct {
				Records []struct{ Path, SHA256 string } `json:"records"`
			} `json:"payload"`
		} `json:"events"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &doc); err != nil {
		t.Fatalf("parse events json: %v", err)
	}
	var out []string
	for _, ev := range doc.Events {
		for _, f := range ev.Payload.Records {
			if f.Path == rel && !slices.Contains(out, f.SHA256) {
				out = append(out, f.SHA256)
			}
		}
	}
	return out
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Line 5: two clones each write an event that names the same store-relative
// path with other bytes, then each runs hyp sync. Each record holds the
// bytes of both events: hyp show prints each event's bytes by its sha256,
// no Experiment reads questionable, and hyp show of the bare path names
// both contents.
func TestSyncKeepsEachSidesNamedBytes(t *testing.T) {
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
	bytesOf := map[string]string{"a": "bytes of a\n", "b": "bytes of b\n"}
	exps := map[string]string{}
	for side, dir := range map[string]string{"a": a, "b": b} {
		t.Chdir(filepath.Join(dir, "s"))
		writeFile(t, shared, bytesOf[side])
		exps[side] = must(t, "add", "exp", "--parent", h, "--title", "E"+side, "--recipe", "none", "--rubric", "reading", "--records", shared, "--why", "w")
	}
	t.Chdir(filepath.Join(a, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	t.Chdir(filepath.Join(b, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	t.Chdir(filepath.Join(a, "s"))
	want(t, hyp(t, "sync"), 0, "synced")

	for _, side := range []string{a, b} {
		t.Chdir(filepath.Join(side, "s"))
		writeFile(t, shared, "a later edit of the work tree\n")
		pins := pinsOf(t, shared)
		if len(pins) != 2 {
			t.Fatalf("%s: the events pin %d contents of %s, want 2", filepath.Base(side), len(pins), shared)
		}
		for _, pin := range pins {
			shown := hyp(t, "show", shared+"@"+pin[:12])
			if shown.code != 0 {
				t.Fatalf("%s: hyp show %s@%s: exit %d %s", filepath.Base(side), shared, pin[:12], shown.code, shown.stderr)
			}
			if got := sha256Hex(shown.stdout); got != pin {
				t.Errorf("%s: an event pins %s at %s, but the record prints bytes %q (sha256 %s)", filepath.Base(side), shared, pin[:12], shown.stdout, got[:12])
			}
		}
		want(t, hyp(t, "show", shared), 2, "different contents")
		for side, e := range exps {
			r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran "+side, "--why", "w")
			if out := hyp(t, "show", r); out.code != 0 || strings.Contains(out.stdout, "questionable") {
				t.Errorf("%s: %s on E%s reads questionable after sync, though the record keeps the bytes its Experiment pins:\n%s", filepath.Base(side), r, side, out.stdout)
			}
		}
	}
}

// A file an event names that changes after hyp hashed it and before the
// write keeps it is refused, so the sha256 the event stores is always the
// sha256 of the bytes the record keeps.
func TestRecordRefusesEvidenceChangedDuringTheWrite(t *testing.T) {
	recordRepo(t)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
	ptr := ".tree/evidence/" + e + "/out.txt"
	writeFile(t, ptr, "hashed bytes\n")
	recordHashed = func() { writeFile(t, ptr, "bytes written meanwhile\n") }
	t.Cleanup(func() { recordHashed = nil })
	want(t, hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w"), 1, "changed while hyp wrote it")
	recordHashed = nil
	want(t, hyp(t, "show", ptr), 2, "no file an event")
}
