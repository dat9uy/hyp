package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// Two clones each pin kit@1 with their own bytes, and each Result cites a
// file of that bundle as evidence (recipes/kit@1/run.sh). After hyp sync
// joins the records, hyp show of that evidence pointer must still print,
// in each clone, the bytes that clone's own Result named: the record keeps
// both copies, one at the ref's path and one under recipes/.kept/.
func TestReviewShowBundleEvidenceAfterSync(t *testing.T) {
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
	want(t, hyp(t, "sync"), 0, "synced")
	b := filepath.Join(top, "b")
	gitIn(t, top, "clone", "-q", remote, b)
	t.Chdir(filepath.Join(b, "s"))
	want(t, hyp(t, "sync"), 0, "synced")

	const ptr = "recipes/kit@1/run.sh"
	type side struct{ clone, tag, r string }
	sides := []*side{{clone: a, tag: "a"}, {clone: b, tag: "b"}}
	for _, s := range sides {
		t.Chdir(filepath.Join(s.clone, "s"))
		bundleIn(t, filepath.Join(s.clone, "s"), "kit", "", map[string]string{"run.sh": "#!/bin/sh\necho " + s.tag + "\n"})
		q := must(t, "add", "dq", "--parent", g, "--title", "Q "+s.tag, "--why", "w")
		h := must(t, "add", "hyp", "--parent", q, "--title", "H "+s.tag, "--why", "w")
		e := must(t, "add", "exp", "--parent", h, "--title", "E "+s.tag, "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
		s.r = must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
		if out := hyp(t, "show", ptr).stdout; !strings.Contains(out, "echo "+s.tag) {
			t.Fatalf("before sync, hyp show %s in clone %s prints %q", ptr, s.tag, out)
		}
	}
	for _, clone := range []string{a, b, a} {
		t.Chdir(filepath.Join(clone, "s"))
		want(t, hyp(t, "sync"), 0, "synced")
	}
	for _, s := range sides {
		t.Chdir(filepath.Join(s.clone, "s"))
		res := hyp(t, "show", ptr)
		if res.code != 0 || !strings.Contains(res.stdout, "echo "+s.tag) {
			t.Errorf("after sync, hyp show %s in clone %s (whose Result %s cites it) exits %d and prints %q, not the bytes %s named (echo %s); stderr %q",
				ptr, s.tag, s.r, res.code, res.stdout, s.r, s.tag, res.stderr)
		}
	}
}
