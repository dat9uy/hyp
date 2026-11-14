package cli

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// kitWithStoreInput writes bundle kit@1 under dir whose known answer's
// input holds a store folder, as a case input often does. The bundle hash
// covers every file under known-answers/NAME/input/, dot entries included.
func kitWithStoreInput(t *testing.T, dir string) {
	t.Helper()
	bundleIn(t, dir, "kit", "", map[string]string{
		"known-answers/ok/input/.tree/events/a.jsonl": "{\"v\":1}\n",
	})
}

// recipeChain adds a Goal down to a Result on an Experiment pinning kit@1
// and returns the Experiment and the Result.
func recipeChain(t *testing.T) (e, r string) {
	t.Helper()
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e = must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
	ptr := ".tree/evidence/" + e + "/out.txt"
	writeFile(t, ptr, "ran\n")
	r = must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
	return e, r
}

// G-SZCBXZ#7: the record must keep each bundle an event pins, so a Result's
// questionable flag reads the same as in a work-tree store. A bundle whose
// known-answer input holds a store folder loses that input in the record's
// copy of the store, and its Result reads questionable in the record.
func TestReviewBundleWithStoreInputKeepsItsHash(t *testing.T) {
	// Control: no git, the work-tree store, as on the base.
	plain := t.TempDir()
	gitConfig(t)
	t.Chdir(plain)
	must(t, "init")
	kitWithStoreInput(t, plain)
	_, r0 := recipeChain(t)
	if out := hyp(t, "show", r0).stdout; strings.Contains(out, "questionable") {
		t.Fatalf("control: the work-tree store already reads %s questionable:\n%s", r0, out)
	}

	repo := recordRepo(t)
	kitWithStoreInput(t, filepath.Join(repo, "s"))
	_, r := recipeChain(t)
	files := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record")
	if !strings.Contains(files, "s/recipes/kit@1/known-answers/ok/input/.tree/events/a.jsonl") {
		t.Fatalf("the record lacks the bundle's known-answer input:\n%s", files)
	}
	if out := hyp(t, "show", r).stdout; strings.Contains(out, "questionable") {
		t.Errorf("in a record repository %s reads questionable, though the record holds every file of kit@1 and the work-tree store reads it fine:\n%s", r, out)
	}
	if out := hyp(t, "check").stdout; strings.Contains(out, r) {
		t.Errorf("hyp check names %s in a record repository:\n%s", r, out)
	}
}

// G-SZCBXZ#5 (and #1): after hyp sync, each event that names a path finds
// the bytes it named. A named evidence file whose path has the shape of an
// events file (a case input holding a store, kept as evidence) is joined
// line by line like an events file, so neither event finds its bytes.
func TestReviewSyncKeepsNamedFileShapedLikeAnEventsFile(t *testing.T) {
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

	shared := ".tree/evidence/case/input/.tree/events/x.jsonl"
	bytesOf := map[string]string{"a": "{\"id\":\"A\"}\n", "b": "{\"id\":\"B\"}\n"}
	for side, dir := range map[string]string{"a": a, "b": b} {
		t.Chdir(filepath.Join(dir, "s"))
		writeFile(t, shared, bytesOf[side])
		must(t, "add", "exp", "--parent", h, "--title", "E"+side, "--recipe", "none", "--rubric", "reading", "--records", shared, "--why", "w")
	}
	t.Chdir(filepath.Join(a, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	t.Chdir(filepath.Join(b, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	t.Chdir(filepath.Join(a, "s"))
	want(t, hyp(t, "sync"), 0, "synced")

	for _, side := range []string{a, b} {
		t.Chdir(filepath.Join(side, "s"))
		pins := pinsOf(t, shared)
		if len(pins) != 2 {
			t.Fatalf("%s: the events pin %d contents of %s, want 2", filepath.Base(side), len(pins), shared)
		}
		for _, pin := range pins {
			shown := hyp(t, "show", shared+"@"+pin[:12])
			if got := sha256Hex(shown.stdout); shown.code != 0 || got != pin {
				t.Errorf("%s: an event pins %s at %s, but hyp show prints exit %d bytes %q (sha256 %.12s) %s", filepath.Base(side), shared, pin[:12], shown.code, shown.stdout, got, shown.stderr)
			}
		}
	}
}

// Goal body, the Experiment rubric: "runs hyp import in a sandbox clone of
// this repository". The candidate refuses hyp import in every clone hyp
// sandbox create makes, so that rubric step cannot run as written.
func TestReviewImportInASandboxClone(t *testing.T) {
	repo := recordRepo(t)
	must(t, "add", "goal", "--title", "G", "--why", "w")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "store")
	clone := filepath.Join(t.TempDir(), "c")
	gitIn(t, repo, "clone", "-q", repo, clone)
	if out, err := exec.Command("git", "-C", clone, "rev-parse", "--verify", "-q", "refs/hyp/record").CombinedOutput(); err == nil {
		t.Fatalf("a plain clone holds the record: %s", out)
	}
	// Mark it the way hyp sandbox create marks its clone.
	if out, err := exec.Command("sh", "-c", "touch \"$(git -C "+clone+" rev-parse --path-format=absolute --git-common-dir)/hyp-sandbox\"").CombinedOutput(); err != nil {
		t.Fatalf("mark: %v %s", err, out)
	}
	t.Chdir(filepath.Join(clone, "s"))
	res := hyp(t, "import", ".")
	if res.code != 0 {
		t.Errorf("hyp import in a sandbox clone: exit %d, stderr %q", res.code, res.stderr)
	}
}
