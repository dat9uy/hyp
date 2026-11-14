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
	// A new Experiment on the clone's own kit@1, which the record holds at
	// its path only for the other clone, and a Result on it.
	for _, s := range sides {
		t.Chdir(filepath.Join(s.clone, "s"))
		e := hyp(t, "add", "exp", "--parent", s.h, "--title", "E2 "+s.tag, "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
		if e.code != 0 {
			t.Errorf("after sync, clone %s cannot add an Experiment pinning its kit@1: exit %d %s", s.tag, e.code, e.stderr)
			continue
		}
		eid := strings.Fields(e.stdout)[0]
		ptr := ".tree/evidence/" + eid + "/out.txt"
		writeFile(t, ptr, "ran "+s.tag+"3\n")
		id := must(t, "record", "--exp", eid, "--verdict", s.h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
		for _, c := range sides {
			t.Chdir(filepath.Join(c.clone, "s"))
			if c != s {
				want(t, hyp(t, "sync"), 0, "synced")
			}
			if out := hyp(t, "show", id).stdout; strings.Contains(out, "questionable") {
				t.Errorf("%s, on a new Experiment of clone %s, reads questionable from clone %s:\n%s", id, s.tag, c.tag, out)
			}
		}
		t.Chdir(filepath.Join(s.clone, "s"))
		want(t, hyp(t, "sync"), 0, "synced")
	}
}

// G-SZCBXZ#1 (and #5): an event names the bytes of each file it names. Two
// clones each pin kit@1 with their own bytes, and each Result cites the
// bundle's run.sh. After hyp sync the record keeps both copies, so the
// pointer names two contents: as for any file named with several
// contents, hyp show of the pointer exits 2 naming each, and each pick
// prints the bytes of one copy.
func TestShowBundleEvidenceAfterSync(t *testing.T) {
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
	for _, s := range []struct{ clone, tag string }{{a, "a"}, {b, "b"}} {
		t.Chdir(filepath.Join(s.clone, "s"))
		bundleIn(t, filepath.Join(s.clone, "s"), "kit", "", map[string]string{"run.sh": "#!/bin/sh\necho " + s.tag + "\n"})
		q := must(t, "add", "dq", "--parent", g, "--title", "Q "+s.tag, "--why", "w")
		h := must(t, "add", "hyp", "--parent", q, "--title", "H "+s.tag, "--why", "w")
		e := must(t, "add", "exp", "--parent", h, "--title", "E "+s.tag, "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
		must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
		if out := hyp(t, "show", ptr); out.code != 0 || out.stdout != "#!/bin/sh\necho "+s.tag+"\n" {
			t.Fatalf("before sync, hyp show %s in clone %s: exit %d %q %s", ptr, s.tag, out.code, out.stdout, out.stderr)
		}
	}
	for _, clone := range []string{a, b, a} {
		t.Chdir(filepath.Join(clone, "s"))
		want(t, hyp(t, "sync"), 0, "synced")
	}
	for _, clone := range []string{a, b} {
		t.Chdir(filepath.Join(clone, "s"))
		res := hyp(t, "show", ptr)
		if res.code != 2 || !strings.Contains(res.stderr, "2 different contents") {
			t.Errorf("after sync, hyp show %s in %s: exit %d stdout %q stderr %q, want exit 2 naming both contents", ptr, filepath.Base(clone), res.code, res.stdout, res.stderr)
			continue
		}
		got := map[string]bool{}
		for _, f := range strings.Fields(res.stderr) {
			f = strings.TrimSuffix(f, ",")
			if !strings.HasPrefix(f, ptr+"@") {
				continue
			}
			one := hyp(t, "show", f)
			if one.code != 0 || sha256Hex(one.stdout) != "" && !strings.HasPrefix(sha256Hex(one.stdout), strings.TrimPrefix(f, ptr+"@")) {
				t.Errorf("hyp show %s in %s: exit %d %q %s", f, filepath.Base(clone), one.code, one.stdout, one.stderr)
			}
			got[one.stdout] = true
		}
		if !got["#!/bin/sh\necho a\n"] || !got["#!/bin/sh\necho b\n"] {
			t.Errorf("in %s the picks of %s print %v, want the bytes of both clones' copies", filepath.Base(clone), ptr, got)
		}
	}
}

// G-SZCBXZ#7: the record keeps each bundle an event pins, so a pin stays
// held whatever another write adds. Two worktrees of one record repository
// each pin kit@1 with their own bytes, one beside the store and one in the
// repository's top library; each Result reads as it did after its write,
// from both worktrees.
func TestPinnedBundleAtTwoLibraryLevels(t *testing.T) {
	gitConfig(t)
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	w1 := filepath.Join(top, "w1")
	gitIn(t, top, "init", "-q", "-b", "main", w1)
	writeFile(t, filepath.Join(w1, "s", "README.md"), "a store\n")
	gitIn(t, w1, "add", "-A")
	gitIn(t, w1, "commit", "-qm", "base")
	t.Chdir(filepath.Join(w1, "s"))
	must(t, "init")
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	w2 := filepath.Join(top, "w2")
	gitIn(t, w1, "worktree", "add", "-q", "-b", "other", w2)

	pin := func(store, lib, tag string) string {
		t.Chdir(store)
		bundleIn(t, lib, "kit", "", map[string]string{"run.sh": "#!/bin/sh\necho " + tag + "\n"})
		q := must(t, "add", "dq", "--parent", g, "--title", "Q "+tag, "--why", "w")
		h := must(t, "add", "hyp", "--parent", q, "--title", "H "+tag, "--why", "w")
		e := must(t, "add", "exp", "--parent", h, "--title", "E "+tag, "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
		ptr := ".tree/evidence/" + e + "/out.txt"
		writeFile(t, ptr, "ran "+tag+"\n")
		r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
		if out := hyp(t, "show", r).stdout; strings.Contains(out, "questionable") {
			t.Fatalf("%s reads questionable right after its write:\n%s", r, out)
		}
		return r
	}
	r1 := pin(filepath.Join(w1, "s"), filepath.Join(w1, "s"), "near")
	r2 := pin(filepath.Join(w2, "s"), w2, "far")
	for _, wt := range []string{w1, w2} {
		t.Chdir(filepath.Join(wt, "s"))
		for _, id := range []string{r1, r2} {
			if out := hyp(t, "show", id).stdout; strings.Contains(out, "questionable") {
				t.Errorf("from %s, %s reads questionable after both writes exited 0:\n%s", filepath.Base(wt), id, out)
			}
			if out := hyp(t, "check").stdout; strings.Contains(out, "questionable evidence "+id) {
				t.Errorf("from %s, hyp check calls %s questionable:\n%s", filepath.Base(wt), id, out)
			}
		}
	}
}
