package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// G-SZCBXZ#4: two writes run at once from two worktrees of one record
// repository both apply. Here one Experiment ran once in each worktree and
// each run kept its log at the Experiment's evidence path, with its own
// bytes, as two branches could before the record. Each Result names the
// bytes it cites by sha256, so the record could keep both (as hyp sync does
// under kept/). The control shows that two Results of one Experiment,
// citing one path with other bytes, both apply in a work-tree store.
func TestReviewTwoWorktreesCiteOnePathAtOnce(t *testing.T) {
	repo := recordRepo(t)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
	wt := filepath.Join(t.TempDir(), "w")
	gitIn(t, repo, "worktree", "add", "-q", "-b", "work", wt)
	ptr := ".tree/evidence/" + e + "/run.log"
	dirs := []string{filepath.Join(repo, "s"), filepath.Join(wt, "s")}
	for i, d := range dirs {
		writeFile(t, filepath.Join(d, ptr), "run from worktree "+string(rune('a'+i))+"\n")
	}
	var wg sync.WaitGroup
	outs := make([]string, 2)
	errs := make([]error, 2)
	for i := range dirs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
			cmd.Dir = dirs[i]
			cmd.Env = append(os.Environ(), "HYP_TEST_AS_HYP=1", "HYP_ACTOR=a")
			out, err := cmd.CombinedOutput()
			outs[i], errs[i] = string(out), err
		}()
	}
	wg.Wait()
	for i := range dirs {
		if errs[i] != nil {
			t.Errorf("write from %s: %v\n%s", dirs[i], errs[i], outs[i])
		}
	}

	// Control: a work-tree store (no git) takes both Results.
	plain := t.TempDir()
	t.Chdir(plain)
	must(t, "init")
	g = must(t, "add", "goal", "--title", "G", "--why", "w")
	q = must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h = must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e = must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
	ptr = ".tree/evidence/" + e + "/run.log"
	for _, tag := range []string{"a", "b"} {
		writeFile(t, ptr, "run "+tag+"\n")
		must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
	}
}

// G-SZCBXZ#7 (and #5): the record keeps each bundle an event pins, so a
// Result's questionable flag is the same from every checkout. Two clones
// each made and pinned a bundle kit@1, each with its own bytes; neither
// Result is questionable before hyp sync. After sync joins the records,
// one clone's bundle survives only under kept/, where no bundle lookup
// reads, so that clone's Result turns questionable on both sides.
func TestReviewSyncKeepsEachSidesPinnedBundle(t *testing.T) {
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

	pinAndRecord := func(clone, tag string) string {
		t.Chdir(filepath.Join(clone, "s"))
		bundleIn(t, filepath.Join(clone, "s"), "kit", "", map[string]string{"run.sh": "#!/bin/sh\necho " + tag + "\n"})
		q := must(t, "add", "dq", "--parent", g, "--title", "Q "+tag, "--why", "w")
		h := must(t, "add", "hyp", "--parent", q, "--title", "H "+tag, "--why", "w")
		e := must(t, "add", "exp", "--parent", h, "--title", "E "+tag, "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
		ptr := ".tree/evidence/" + e + "/out.txt"
		writeFile(t, ptr, "ran "+tag+"\n")
		return must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
	}
	ra := pinAndRecord(a, "a")
	rb := pinAndRecord(b, "b")
	for _, c := range []struct{ clone, r string }{{a, ra}, {b, rb}} {
		t.Chdir(filepath.Join(c.clone, "s"))
		if out := hyp(t, "show", c.r).stdout; strings.Contains(out, "questionable") {
			t.Fatalf("before sync, %s reads questionable in its own clone:\n%s", c.r, out)
		}
	}
	for _, clone := range []string{a, b, a} {
		t.Chdir(filepath.Join(clone, "s"))
		want(t, hyp(t, "sync"), 0, "synced")
	}
	for _, clone := range []string{a, b} {
		t.Chdir(filepath.Join(clone, "s"))
		for _, r := range []string{ra, rb} {
			if out := hyp(t, "show", r).stdout; strings.Contains(out, "questionable") {
				t.Errorf("after sync, %s reads questionable from clone %s:\n%s", r, filepath.Base(clone), out)
			}
		}
	}
}

// none (the Goal body's decision "a write is durable once it exits 0"; also
// G-SZCBXZ#1 and #3 under an unclean shutdown): the base fsyncs the events
// file before a write exits. A record write instead adds loose objects and
// moves a ref, which git by default does not harden (core.fsync defaults to
// committed,-loose-object, and reference is not in it). This test logs every
// git run of one record write and asks that each one that writes an object
// or the ref ran with core.fsync naming loose objects and refs.
func TestReviewRecordWriteHardensObjectsAndRef(t *testing.T) {
	recordRepo(t)
	real, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not on PATH")
	}
	bin := t.TempDir()
	log := filepath.Join(t.TempDir(), "git.log")
	script := "#!/bin/sh\nprintf '%s|%s|%s=%s|%s\\n' \"$GIT_CONFIG_PARAMETERS\" \"$GIT_CONFIG_COUNT\" \"$GIT_CONFIG_KEY_0\" \"$GIT_CONFIG_VALUE_0\" \"$*\" >> " + log + "\nexec " + real + " \"$@\"\n"
	writeFile(t, filepath.Join(bin, "git"), script)
	if err := os.Chmod(filepath.Join(bin, "git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	must(t, "add", "goal", "--title", "durable", "--why", "w")
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	cfg, _ := exec.Command(real, "config", "--get", "core.fsync").Output()
	hardened := func(line string) bool {
		v := strings.ToLower(line + " " + string(cfg))
		if !strings.Contains(v, "core.fsync") && len(cfg) == 0 {
			return false
		}
		return strings.Contains(v, "all") || (strings.Contains(v, "loose-object") || strings.Contains(v, "objects") || strings.Contains(v, "committed")) && strings.Contains(v, "reference")
	}
	writers := 0
	for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		args := l[strings.LastIndex(l, "|")+1:]
		if !strings.Contains(args, "hash-object -w") && !strings.Contains(args, "write-tree") && !strings.Contains(args, "commit-tree") && !strings.Contains(args, "update-ref") {
			continue
		}
		writers++
		if !hardened(l) {
			t.Errorf("git %s ran with no core.fsync for loose objects and refs; a write that exits 0 may be lost on an unclean shutdown", args)
		}
	}
	if writers == 0 {
		t.Fatalf("no git run wrote an object or the ref; log:\n%s", data)
	}
}
