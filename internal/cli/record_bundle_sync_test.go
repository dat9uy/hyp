package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// G-SZCBXZ#7 (and #5): the record keeps each bundle an event pins, so a
// Result's questionable flag is the same from every checkout. Two clones
// each make and pin a bundle kit@1 with their own bytes; neither Result is
// questionable before hyp sync. After sync joins the two records, each
// Result must read as before in both clones, and each clone must still
// record a Result on its own Experiment from the bundle it holds.
func TestSyncKeepsEachSidesPinnedBundle(t *testing.T) {
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

	type side struct{ clone, tag, h, e, r string }
	sides := []*side{{clone: a, tag: "a"}, {clone: b, tag: "b"}}
	for _, s := range sides {
		t.Chdir(filepath.Join(s.clone, "s"))
		bundleIn(t, filepath.Join(s.clone, "s"), "kit", "", map[string]string{"run.sh": "#!/bin/sh\necho " + s.tag + "\n"})
		q := must(t, "add", "dq", "--parent", g, "--title", "Q "+s.tag, "--why", "w")
		s.h = must(t, "add", "hyp", "--parent", q, "--title", "H "+s.tag, "--why", "w")
		s.e = must(t, "add", "exp", "--parent", s.h, "--title", "E "+s.tag, "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
		ptr := ".tree/evidence/" + s.e + "/out1.txt"
		writeFile(t, ptr, "ran "+s.tag+"1\n")
		s.r = must(t, "record", "--exp", s.e, "--verdict", s.h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
		if out := hyp(t, "show", s.r).stdout; strings.Contains(out, "questionable") {
			t.Fatalf("before sync, %s reads questionable in its own clone:\n%s", s.r, out)
		}
	}
	for _, clone := range []string{a, b, a} {
		t.Chdir(filepath.Join(clone, "s"))
		want(t, hyp(t, "sync"), 0, "synced")
	}
	files := gitIn(t, a, "ls-tree", "-r", "--name-only", "refs/hyp/record")
	if strings.Count(files, "/kit@1/run.sh") != 2 || !strings.Contains(files, "s/recipes/.kept/") || strings.Contains(files, "s/.tree/kept/") {
		t.Errorf("the joined record keeps kit@1 as other than one copy at its path and one whole copy under recipes/.kept/:\n%s", files)
	}
	for _, c := range sides {
		t.Chdir(filepath.Join(c.clone, "s"))
		for _, s := range sides {
			if out := hyp(t, "show", s.r).stdout; strings.Contains(out, "questionable") {
				t.Errorf("after sync, %s reads questionable from clone %s:\n%s", s.r, c.tag, out)
			}
			if out := hyp(t, "check").stdout; strings.Contains(out, "questionable evidence "+s.r) {
				t.Errorf("after sync, hyp check in clone %s calls %s questionable:\n%s", c.tag, s.r, out)
			}
		}
	}
	for _, s := range sides {
		t.Chdir(filepath.Join(s.clone, "s"))
		ptr := ".tree/evidence/" + s.e + "/out2.txt"
		writeFile(t, ptr, "ran "+s.tag+"2\n")
		r := hyp(t, "record", "--exp", s.e, "--verdict", s.h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
		if r.code != 0 {
			t.Errorf("after sync, clone %s cannot record a Result on %s from the kit@1 it pinned: exit %d %s", s.tag, s.e, r.code, r.stderr)
			continue
		}
		id := strings.TrimSpace(r.stdout)
		if out := hyp(t, "show", id).stdout; strings.Contains(out, "questionable") {
			t.Errorf("after sync, clone %s's new Result %s reads questionable:\n%s", s.tag, id, out)
		}
	}
}
