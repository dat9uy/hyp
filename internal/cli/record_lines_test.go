package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

// The tests in this file state the record: in a git repository, every
// store's events and the files they name live on a hidden ref that no
// branch holds. They drive hyp only through Run and git, in temporary
// repositories.

// TestMain lets a test run this binary as hyp in another process, so two
// writes can run at once from two working directories.
func TestMain(m *testing.M) {
	if os.Getenv("HYP_TEST_AS_HYP") == "1" {
		os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

// gitConfig isolates git from the user's configuration and gives commits
// an identity.
func gitConfig(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	writeFile(t, cfg, "[user]\n\tname = t\n\temail = t@t\n[init]\n\tdefaultBranch = main\n[gc]\n\tauto = 0\n")
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("HYP_ACTOR", "a")
}

// recordRepo makes a git repository whose folder s, committed with a
// README, holds a store that hyp init made there, and changes to s.
func recordRepo(t *testing.T) string {
	t.Helper()
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, "s", "README.md"), "a store\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, "s"))
	must(t, "init")
	return repo
}

// chain adds a Goal, a question, a Hypothesis, an Experiment and its Result,
// which cites the evidence file out.txt holding data, from the working
// directory's store. It returns the Goal, the Result and the evidence
// pointer.
func chain(t *testing.T, data string) (g, r, ptr string) {
	t.Helper()
	g = must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
	ptr = ".tree/evidence/" + e + "/out.txt"
	writeFile(t, ptr, data)
	r = must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
	return g, r, ptr
}

// views is what hyp prints for the chain from the working directory.
func views(t *testing.T, g, r, ptr string) []string {
	t.Helper()
	var out []string
	for _, args := range [][]string{{"status", g}, {"show", g}, {"show", r}, {"show", ptr}, {"events", "--json"}} {
		res := hyp(t, args...)
		if res.code != 0 {
			t.Fatalf("hyp %v: exit %d, stderr %q", args, res.code, res.stderr)
		}
		out = append(out, res.stdout)
	}
	return out
}

func sameViews(t *testing.T, when string, got, want []string) {
	t.Helper()
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: view %d differs:\n%s\nwant\n%s", when, i, got[i], want[i])
		}
	}
}

// Line 1: an event a write exited 0 on, and the bytes of the file it names,
// print the same in every worktree of the clone after every branch and
// worktree act, and a later change of the work-tree file changes nothing.
func TestRecordOutlivesEveryWorktreeAct(t *testing.T) {
	repo := recordRepo(t)
	must(t, "add", "goal", "--title", "first", "--why", "w")
	wt := filepath.Join(t.TempDir(), "w")
	gitIn(t, repo, "worktree", "add", "-q", "-b", "work", wt)
	t.Chdir(filepath.Join(wt, "s"))
	g, r, ptr := chain(t, "the bytes\n")
	want := views(t, g, r, ptr)
	if !strings.Contains(want[3], "the bytes") {
		t.Fatalf("hyp show %s printed %q, want the file's bytes", ptr, want[3])
	}
	writeFile(t, ptr, "changed later\n")
	sameViews(t, "after the work-tree file changed", views(t, g, r, ptr), want)
	if err := os.Remove(ptr); err != nil {
		t.Fatal(err)
	}
	sameViews(t, "after the work-tree file was removed", views(t, g, r, ptr), want)

	gitIn(t, wt, "add", "-A")
	gitIn(t, wt, "commit", "-qm", "work", "--allow-empty")
	gitIn(t, repo, "merge", "-q", "--squash", "work")
	gitIn(t, repo, "commit", "-qm", "squash", "--allow-empty")
	gitIn(t, repo, "reset", "-q", "--hard", "HEAD~1")
	t.Chdir(repo)
	gitIn(t, repo, "worktree", "remove", "--force", wt)
	gitIn(t, repo, "branch", "-q", "-D", "work")
	gitIn(t, repo, "checkout", "-q", "-b", "other")
	gitIn(t, repo, "checkout", "-q", "main")
	gitIn(t, repo, "clean", "-fdxq")
	gitIn(t, repo, "reflog", "expire", "--expire=now", "--all")
	gitIn(t, repo, "gc", "-q", "--prune=now")
	t.Chdir(filepath.Join(repo, "s"))
	sameViews(t, "in the main worktree after squash, reset, branch delete, worktree remove, clean and gc", views(t, g, r, ptr), want)

	third := filepath.Join(t.TempDir(), "third")
	gitIn(t, repo, "worktree", "add", "-q", "-b", "third", third)
	t.Chdir(filepath.Join(third, "s"))
	sameViews(t, "in a new worktree", views(t, g, r, ptr), want)
}

// Line 2: a commit a Hypothesis subject names stays reachable from a ref,
// and a sandbox can be made at it, after its branch is deleted and git gc.
func TestSubjectCommitStaysReachable(t *testing.T) {
	repo := recordRepo(t)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	gitIn(t, repo, "checkout", "-q", "-b", "feat")
	writeFile(t, filepath.Join(repo, "feat.txt"), "only on feat\n")
	gitIn(t, repo, "add", "feat.txt")
	gitIn(t, repo, "commit", "-qm", "feat")
	c := gitIn(t, repo, "rev-parse", "HEAD")
	gitIn(t, repo, "checkout", "-q", "main")
	must(t, "subject", h, "--base", c, "--why", "w")
	gitIn(t, repo, "branch", "-q", "-D", "feat")
	gitIn(t, repo, "reflog", "expire", "--expire=now", "--all")
	gitIn(t, repo, "gc", "-q", "--prune=now")
	if out, err := exec.Command("git", "-C", repo, "cat-file", "-e", c+"^{commit}").CombinedOutput(); err != nil {
		t.Fatalf("commit %s is gone after its branch was deleted and gc ran: %v %s", c, err, out)
	}
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	t.Setenv("HOME", t.TempDir())
	stub := filepath.Join(t.TempDir(), "hyp")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho hyp 0.0.0-test\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	want(t, hyp(t, "sandbox", "create", "at-subject", "--host", repo, "--base", c, "--hyp", stub, "--recipe", "none"), 0, "")
	want(t, hyp(t, "sandbox", "remove", "at-subject", "--discard"), 0, "")
}

// Line 4: writes run at once from two worktrees all apply, each new id
// sorts after every id the store held before it, and the refusal when no
// ULID sorts after the newest id stays.
func TestTwoWorktreesWriteAtOnce(t *testing.T) {
	repo := recordRepo(t)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	wt := filepath.Join(t.TempDir(), "w")
	gitIn(t, repo, "worktree", "add", "-q", "-b", "work", wt)
	dirs := []string{filepath.Join(repo, "s"), filepath.Join(wt, "s")}
	const n = 8
	var wg sync.WaitGroup
	errs := make([]string, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
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
			t.Fatalf("write %d from %s failed: %s", i, dirs[i%2], e)
		}
	}
	for _, d := range dirs {
		t.Chdir(d)
		var out struct{ Events []struct{ ID string } }
		if err := json.Unmarshal([]byte(hyp(t, "events", "--json").stdout), &out); err != nil {
			t.Fatal(err)
		}
		if len(out.Events) != n+1 {
			t.Fatalf("hyp events in %s lists %d events, want %d", d, len(out.Events), n+1)
		}
	}
	// The record's events file holds the lines in the order they were
	// written, so each id sorts after every id before it.
	data := gitIn(t, repo, "show", "refs/hyp/record:s/.tree/events/a.jsonl")
	var ids []string
	for _, l := range strings.Split(data, "\n") {
		var ev struct{ ID string }
		if json.Unmarshal([]byte(l), &ev) == nil {
			ids = append(ids, ev.ID)
		}
	}
	if len(ids) != n+1 || !slices.IsSorted(ids) || len(slices.Compact(slices.Clone(ids))) != n+1 {
		t.Fatalf("ids in write order %v, want %d ids each after the last", ids, n+1)
	}

	// A store whose newest id is the greatest ULID refuses the next write.
	other := t.TempDir()
	t.Chdir(other)
	must(t, "init")
	max := "7ZZZZZZZZZZZZZZZZZZZZZZZZZ"
	writeFile(t, filepath.Join(other, ".tree", "events", "z.jsonl"), `{"v":1,"id":"`+max+`","ts":"2026-10-27T00:00:00Z","actor":"z","type":"node.add","node":"`+max+`","payload":{"kind":"Goal","parent":"","title":"g"},"why":"w"}`+"\n")
	writeFile(t, filepath.Join(repo, "m", "README.md"), "m\n")
	if err := os.CopyFS(filepath.Join(repo, "m"), os.DirFS(other)); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(repo, "m"))
	want(t, hyp(t, "import", "."), 0, "imported 1 events")
	want(t, hyp(t, "add", "goal", "--title", "next", "--why", "w"), 1, "no ULID sorts after "+max)
}

// Line 5: hyp sync makes both records the union of both, a fresh clone
// syncs to what the syncing clone prints, a remote with no record gets the
// local one, and a remote that refuses the ref gives exit 3.
func TestSyncJoinsTwoRecords(t *testing.T) {
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
	g, r, ptr := chain(t, "from a\n")
	want(t, hyp(t, "sync"), 0, "synced")

	b := filepath.Join(top, "b")
	gitIn(t, top, "clone", "-q", remote, b)
	t.Chdir(filepath.Join(b, "s"))
	want(t, hyp(t, "sync", "origin"), 0, "synced")
	t.Chdir(filepath.Join(a, "s"))
	aViews := views(t, g, r, ptr)
	t.Chdir(filepath.Join(b, "s"))
	sameViews(t, "a fresh clone after hyp sync", views(t, g, r, ptr), aViews)

	// Each side writes what the other lacks, with a subject commit that
	// only a holds and a file that only b holds.
	t.Chdir(filepath.Join(a, "s"))
	q := must(t, "add", "dq", "--parent", g, "--title", "from a", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	gitIn(t, a, "checkout", "-q", "-b", "local")
	writeFile(t, filepath.Join(a, "local.txt"), "only in a\n")
	gitIn(t, a, "add", "local.txt")
	gitIn(t, a, "commit", "-qm", "local")
	c := gitIn(t, a, "rev-parse", "HEAD")
	gitIn(t, a, "checkout", "-q", "main")
	must(t, "subject", h, "--base", c, "--why", "w")
	t.Chdir(filepath.Join(b, "s"))
	_, r2, ptr2 := chain(t, "from b\n")
	t.Chdir(filepath.Join(a, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	t.Chdir(filepath.Join(b, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	t.Chdir(filepath.Join(a, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	if ta, tb := gitIn(t, a, "rev-parse", "refs/hyp/record"), gitIn(t, b, "rev-parse", "refs/hyp/record"); ta != tb {
		t.Fatalf("records differ after sync: a %s, b %s", ta, tb)
	}
	if tr := gitIn(t, remote, "rev-parse", "refs/hyp/record"); tr != gitIn(t, a, "rev-parse", "refs/hyp/record") {
		t.Fatalf("the remote's record %s is not the joined record", tr)
	}
	if out, err := exec.Command("git", "-C", b, "cat-file", "-e", c+"^{commit}").CombinedOutput(); err != nil {
		t.Fatalf("b lacks a's subject commit %s after sync: %s", c, out)
	}
	aViews = views(t, g, r2, ptr2)
	t.Chdir(filepath.Join(b, "s"))
	sameViews(t, "b after the join", views(t, g, r2, ptr2), aViews)

	fresh := filepath.Join(top, "c")
	gitIn(t, top, "clone", "-q", remote, fresh)
	t.Chdir(filepath.Join(fresh, "s"))
	want(t, hyp(t, "sync", "origin"), 0, "synced")
	sameViews(t, "a fresh clone of the joined remote", views(t, g, r2, ptr2), aViews)

	// A remote whose update hook refuses the ref.
	refusing := filepath.Join(top, "refusing.git")
	gitIn(t, top, "init", "-q", "--bare", refusing)
	writeFile(t, filepath.Join(refusing, "hooks", "update"), "#!/bin/sh\ncase \"$1\" in refs/hyp/*) exit 1;; esac\nexit 0\n")
	if err := os.Chmod(filepath.Join(refusing, "hooks", "update"), 0o755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, a, "remote", "add", "refusing", refusing)
	t.Chdir(filepath.Join(a, "s"))
	want(t, hyp(t, "sync", "refusing"), 3, "refused")
}

// Line 6: the record holds the events and the files they name and no other;
// hyp init and hyp import make git ignore the evidence folder, and hyp
// import untracks each evidence file no event names, leaving it on disk.
func TestRecordHoldsOnlyNamedFiles(t *testing.T) {
	repo := recordRepo(t)
	bundleIn(t, filepath.Join(repo, "s"), "kit", "", map[string]string{"run.sh": "#!/bin/sh\n"})
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
	cited := ".tree/evidence/" + e + "/cited.txt"
	writeFile(t, cited, "cited\n")
	writeFile(t, ".tree/evidence/"+e+"/uncited.txt", "uncited\n")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", cited, "--why", "w")
	got := strings.Split(gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record"), "\n")
	wantFiles := []string{"s/" + cited, "s/.tree/events/a.jsonl", "s/recipes/kit@1/recipe.md", "s/recipes/kit@1/run.sh"}
	slices.Sort(got)
	slices.Sort(wantFiles)
	if !slices.Equal(got, wantFiles) {
		t.Fatalf("the record holds %v, want %v", got, wantFiles)
	}
	if out, err := exec.Command("git", "-C", repo, "check-ignore", "-q", "s/.tree/evidence/"+e+"/uncited.txt").CombinedOutput(); err != nil {
		t.Fatalf("git does not ignore an evidence file after hyp init: %v %s", err, out)
	}

	// A work-tree store whose evidence git tracks, made with no record.
	other := t.TempDir()
	t.Chdir(other)
	t.Setenv("HYP_ACTOR", "b")
	must(t, "init")
	g2, _, ptr := chain(t, "kept\n")
	_ = g2
	writeFile(t, filepath.Join(other, ".tree", "evidence", "stray.txt"), "stray\n")
	if err := os.Remove(filepath.Join(other, ".tree", ".gitignore")); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(repo, "w")
	if err := os.CopyFS(dst, os.DirFS(other)); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "add", "-A", "w")
	gitIn(t, repo, "commit", "-qm", "work-tree store")
	t.Chdir(dst)
	want(t, hyp(t, "import", "."), 0, "imported")
	tracked := gitIn(t, repo, "ls-files", "w/.tree/evidence")
	if !strings.Contains(tracked, "w/"+ptr) || strings.Contains(tracked, "stray.txt") {
		t.Fatalf("tracked evidence after import:\n%s\nwant the cited file and not stray.txt", tracked)
	}
	if _, err := os.Stat(filepath.Join(dst, ".tree", "evidence", "stray.txt")); err != nil {
		t.Fatalf("import removed the untracked file from disk: %v", err)
	}
	if out, err := exec.Command("git", "-C", repo, "check-ignore", "-q", "w/.tree/evidence/new.txt").CombinedOutput(); err != nil {
		t.Fatalf("git does not ignore an evidence file after hyp import: %v %s", err, out)
	}
}

// Line 7: a Result's flags and the warnings hyp check prints for it are the
// same from every checkout, whatever its recipes/ and evidence hold.
func TestFlagsSameFromEveryCheckout(t *testing.T) {
	repo := recordRepo(t)
	bare := gitIn(t, repo, "rev-parse", "HEAD")
	bundleIn(t, filepath.Join(repo, "s"), "kit", "", map[string]string{"run.sh": "#!/bin/sh\n"})
	gitIn(t, repo, "add", "s/recipes")
	gitIn(t, repo, "commit", "-qm", "kit")
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
	sha := strings.Repeat("ab", 20)
	ptr := ".tree/evidence/" + e + "/log.txt"
	writeFile(t, ptr, "built "+sha+"\n")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "built "+sha, "--evidence", ptr, "--why", "w")
	flags := func() []string {
		var out []string
		for _, l := range strings.Split(hyp(t, "check").stdout, "\n") {
			if strings.Contains(l, r) && !strings.Contains(l, "checked-out") {
				out = append(out, l)
			}
		}
		return append(out, hyp(t, "show", r).stdout, hyp(t, "status", g).stdout)
	}
	want1 := flags()
	wt := filepath.Join(t.TempDir(), "w")
	gitIn(t, repo, "worktree", "add", "-q", "-b", "bare", wt, bare)
	t.Chdir(filepath.Join(wt, "s"))
	sameViews(t, "a checkout with no recipes/ and no evidence", flags(), want1)
	t.Chdir(filepath.Join(repo, "s"))
	writeFile(t, filepath.Join(repo, "s", "recipes", "kit@1", "run.sh"), "#!/bin/sh\necho edited\n")
	writeFile(t, ptr, "other bytes\n")
	sameViews(t, "a checkout whose bundle and evidence changed", flags(), want1)
}

// Line 8: hyp import adds every event a work-tree store holds that the
// record lacks, from the checkout or from REF, with the files they name and
// their subject commits; again it adds nothing; statuses equal the store's.
func TestImportAddsEachMissingEvent(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, "README.md"), "r\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	base := gitIn(t, repo, "rev-parse", "HEAD")
	// Two work-tree stores, written while git holds no record.
	var statuses, goals, hyps []string
	for _, name := range []string{"one", "two"} {
		t.Chdir(repo)
		if err := os.Mkdir(name, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(filepath.Join(repo, name))
		must(t, "init")
		dropRecord(t, repo)
		g, _, _ := chain(t, name+"\n")
		q := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
		h := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
		must(t, "subject", h, "--base", base, "--why", "w")
		statuses = append(statuses, hyp(t, "status", g).stdout+hyp(t, "show", h).stdout)
		goals, hyps = append(goals, g), append(hyps, h)
	}
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "stores")
	at := gitIn(t, repo, "rev-parse", "HEAD")
	// two is read from the commit: its store folder is gone from the checkout.
	gitIn(t, repo, "rm", "-rq", "two/.tree")
	writeFile(t, filepath.Join(repo, "two", "README.md"), "two\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "drop two's store")
	t.Chdir(repo)
	want(t, hyp(t, "import", "one", at+":two"), 0, "imported")
	tip := gitIn(t, repo, "rev-parse", "refs/hyp/record")
	want(t, hyp(t, "import", "one", at+":two"), 0, "holds every event")
	if again := gitIn(t, repo, "rev-parse", "refs/hyp/record"); again != tip {
		t.Fatalf("a second import moved the record from %s to %s", tip, again)
	}
	for i, name := range []string{"one", "two"} {
		t.Chdir(filepath.Join(repo, name))
		if got := hyp(t, "status", goals[i]).stdout + hyp(t, "show", hyps[i]).stdout; got != statuses[i] {
			t.Fatalf("store %s after import prints\n%s\nwant\n%s", name, got, statuses[i])
		}
	}
}

// dropRecord deletes the record ref, if there is one, so the repository
// keeps its stores in the work tree.
func dropRecord(t *testing.T, repo string) {
	t.Helper()
	_ = exec.Command("git", "-C", repo, "update-ref", "-d", "refs/hyp/record").Run()
}

// Line 9: a write where discovery finds a work-tree store holding an event
// the record lacks is refused, naming hyp import; hyp check reports it and a
// checked-out bundle whose hash differs from a pin; hyp init makes a git
// repository a record repository; with no record ref, or no git, hyp reads
// and writes the work-tree store.
func TestUnimportedWorkTreeStore(t *testing.T) {
	repo := recordRepo(t)
	if out, err := exec.Command("git", "-C", repo, "rev-parse", "--verify", "-q", "refs/hyp/record").CombinedOutput(); err != nil {
		t.Fatalf("hyp init made no record: %v %s", err, out)
	}
	bundleIn(t, filepath.Join(repo, "s"), "kit", "", nil)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
	writeFile(t, filepath.Join(repo, "s", "recipes", "kit@1", "recipe.md"), "---\nname: kit\nversion: 1\nformat: 1\n---\n\nedited\n")
	want(t, hyp(t, "check"), 0, "checked-out bundle kit@1 hashes")

	other := t.TempDir()
	t.Chdir(other)
	must(t, "init")
	chain(t, "w\n")
	dst := filepath.Join(repo, "w")
	if err := os.CopyFS(dst, os.DirFS(other)); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dst)
	want(t, hyp(t, "add", "goal", "--title", "lost", "--why", "w"), 1, "hyp import")
	want(t, hyp(t, "check"), 0, "run hyp import "+dst)

	// No record ref: the work-tree store, as before the record.
	dropRecord(t, repo)
	want(t, hyp(t, "add", "goal", "--title", "kept", "--why", "w"), 0, "")
	if data, err := os.ReadFile(filepath.Join(dst, ".tree", "events", "a.jsonl")); err != nil || !bytes.Contains(data, []byte(`"title":"kept"`)) {
		t.Fatalf("with no record the write did not reach the work-tree events file: %v", err)
	}
	// No .git at or above: the work-tree store.
	t.Chdir(other)
	want(t, hyp(t, "add", "goal", "--title", "plain", "--why", "w"), 0, "")
	if data, err := os.ReadFile(filepath.Join(other, ".tree", "events", "a.jsonl")); err != nil || !bytes.Contains(data, []byte(`"title":"plain"`)) {
		t.Fatalf("with no git the write did not reach the work-tree events file: %v", err)
	}
}
