package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func sumOf(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// G-SZCBXZ#6 (and the Goal body: "old .tree/ folders stay as frozen copies,
// cited files tracked"): hyp import untracks only the evidence files no
// event names. An uncited file whose name holds a glob character must not
// take a cited file with it.
func TestReviewImportKeepsCitedFileTrackedBesideAGlobName(t *testing.T) {
	repo := recordRepo(t)
	other := t.TempDir()
	t.Chdir(other)
	must(t, "init")
	_, _, ptr := chain(t, "cited\n") // .tree/evidence/E-x/out.txt
	dir := filepath.Dir(ptr)
	if os.Getenv("REVIEW_CONTROL") == "" {
		writeFile(t, filepath.Join(other, dir, "ou[t].txt"), "uncited\n")
	}
	dst := filepath.Join(repo, "w")
	if err := os.CopyFS(dst, os.DirFS(other)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dst, ".tree", ".gitignore")); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "add", "-A", "w")
	gitIn(t, repo, "commit", "-qm", "work-tree store")
	t.Chdir(dst)
	want(t, hyp(t, "import", "."), 0, "imported")
	tracked := gitIn(t, repo, "ls-files", "w/.tree/evidence")
	if !strings.Contains(tracked, "w/"+ptr) {
		t.Fatalf("hyp import untracked the cited file %s; tracked evidence now:\n%s", ptr, tracked)
	}
}

// storeHistory writes, in a folder outside git, a store whose Result R1
// cites P with "one", then an Experiment E2 that pins P as a record after
// the file was overwritten with "two", as a work-tree store allows. It copies the store
// into repo/name after each event and commits it, and returns the first
// commit and P.
func storeHistory(t *testing.T, repo, name string) (first, p string) {
	t.Helper()
	other := t.TempDir()
	t.Chdir(other)
	must(t, "init")
	_, h, _, ptr := chainH(t, "one\n")
	dst := filepath.Join(repo, name)
	cp := func(msg string) string {
		if err := os.RemoveAll(filepath.Join(dst, ".tree")); err != nil {
			t.Fatal(err)
		}
		if err := os.CopyFS(dst, os.DirFS(other)); err != nil {
			t.Fatal(err)
		}
		_ = os.Remove(filepath.Join(dst, ".tree", ".gitignore"))
		gitIn(t, repo, "add", "-A", name)
		gitIn(t, repo, "commit", "-qm", msg)
		return gitIn(t, repo, "rev-parse", "HEAD")
	}
	first = cp("R1")
	t.Chdir(other)
	writeFile(t, ptr, "two\n")
	must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--rubric", "reading", "--records", ptr, "--why", "w")
	cp("E2")
	return first, ptr
}

// G-SZCBXZ#1 and #8: after hyp import, hyp show PATH@SHA prints the bytes
// with that sha256 or refuses; it never prints other bytes under it.
func TestReviewShowNeverPrintsOtherBytesForAPickedSha(t *testing.T) {
	repo := recordRepo(t)
	_, ptr := storeHistory(t, repo, "w")
	t.Chdir(filepath.Join(repo, "w"))
	want(t, hyp(t, "import", "."), 0, "imported")
	one := sumOf("one\n")
	r := hyp(t, "show", ptr+"@"+one[:12])
	t.Logf("hyp show %s@%s: exit %d, stdout %q, stderr %q", ptr, one[:12], r.code, r.stdout, r.stderr)
	if r.code == 0 && r.stdout != "one\n" {
		t.Fatalf("hyp show %s@%s printed %q, whose sha256 is %s, not %s", ptr, one[:12], r.stdout, sumOf(r.stdout)[:12], one[:12])
	}
}

// G-SZCBXZ#8: hyp import from REF and then from the checkout adds every
// event the record lacks; running it again adds nothing.
func TestReviewImportFromRefThenCheckoutAddsTheLaterEvent(t *testing.T) {
	repo := recordRepo(t)
	first, _ := storeHistory(t, repo, "w")
	t.Chdir(repo)
	want(t, hyp(t, "import", first+":w"), 0, "imported")
	r := hyp(t, "import", "w")
	t.Logf("hyp import w after %s:w: exit %d, stdout %q, stderr %q", first[:12], r.code, r.stdout, r.stderr)
	want(t, r, 0, "imported 1 events")
}

// G-SZCBXZ#1 and #9: in a record repository, an event a write exited 0 on is
// printed by hyp status; a write either reaches the record or fails. With
// no git on PATH hyp cannot read the record, which is an environment
// problem (exit 3), not a cue to read and write the work-tree store.
func TestReviewNoGitOnPathInARecordRepository(t *testing.T) {
	recordRepo(t)
	g := must(t, "add", "goal", "--title", "in the record", "--why", "w")
	path := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	st := hyp(t, "status")
	t.Logf("hyp status with no git: exit %d, stdout %q, stderr %q", st.code, st.stdout, st.stderr)
	if st.code == 0 && !strings.Contains(st.stdout, g) {
		t.Errorf("hyp status with no git on PATH exits 0 and does not print %s, which a write exited 0 on", g)
	}
	w := hyp(t, "add", "goal", "--title", "written with no git", "--why", "w")
	t.Logf("hyp add goal with no git: exit %d, stdout %q, stderr %q", w.code, w.stdout, w.stderr)
	t.Setenv("PATH", path)
	if w.code == 0 {
		lost := strings.Fields(w.stdout)[0]
		if out := hyp(t, "status").stdout; !strings.Contains(out, lost) {
			t.Errorf("hyp add goal exited 0 with no git on PATH, and with git back hyp status does not print %s:\n%s", lost, out)
		}
		next := hyp(t, "add", "goal", "--title", "next", "--why", "w")
		t.Logf("the next write with git back: exit %d, stderr %q", next.code, next.stderr)
	}
}

// G-SZCBXZ#4: writes that each name a file, run at once from two worktrees,
// all apply.
func TestReviewTwoWorktreesRecordEvidenceAtOnce(t *testing.T) {
	repo := recordRepo(t)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	wt := filepath.Join(t.TempDir(), "w")
	gitIn(t, repo, "worktree", "add", "-q", "-b", "work", wt)
	dirs := []string{filepath.Join(repo, "s"), filepath.Join(wt, "s")}
	const n = 6
	exps := make([]string, n)
	for i := range n {
		t.Chdir(dirs[i%2])
		exps[i] = must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
		writeFile(t, filepath.Join(dirs[i%2], ".tree", "evidence", exps[i], "out.txt"), exps[i]+"\n")
	}
	var wg sync.WaitGroup
	errs := make([]string, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "record", "--exp", exps[i], "--verdict", h+"=supports", "--observation", "ran", "--evidence", ".tree/evidence/"+exps[i]+"/out.txt", "--why", "w")
			cmd.Dir = dirs[i%2]
			cmd.Env = append(os.Environ(), "HYP_TEST_AS_HYP=1", "HYP_ACTOR=a")
			if out, err := cmd.CombinedOutput(); err != nil {
				errs[i] = string(out) + err.Error()
			}
		}()
	}
	wg.Wait()
	for i, e := range errs {
		if e != "" {
			t.Errorf("write %d from %s failed: %s", i, dirs[i%2], e)
		}
	}
	for i := range n {
		t.Chdir(dirs[(i+1)%2])
		want(t, hyp(t, "show", ".tree/evidence/"+exps[i]+"/out.txt"), 0, exps[i])
	}
}
