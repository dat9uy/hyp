package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
)

// noRecord fails the test when the repository at top holds the record ref.
func noRecord(t *testing.T, top, when string) {
	t.Helper()
	if exec.Command("git", "-C", top, "rev-parse", "--verify", "-q", "refs/hyp/record").Run() == nil {
		t.Fatalf("%s: the repository holds the record ref, want nothing written", when)
	}
}

// refused fails the test unless r is a rule refusal, exit 1.
func refused(t *testing.T, r result, substr string) {
	t.Helper()
	if r.code != 1 || !strings.Contains(r.stderr, "refused:") || !strings.Contains(r.stderr, substr) {
		t.Fatalf("exit %d, stderr %q; want a refusal, exit 1, naming %q", r.code, r.stderr, substr)
	}
}

// storeBeforeGit makes, in a new folder top, a store s written while no git
// repository holds it, then makes top a git repository that commits it,
// the evidence included. It returns top, the Goal and the cited pointer.
func storeBeforeGit(t *testing.T) (top, g, ptr string) {
	t.Helper()
	gitConfig(t)
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(top, "s"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(top, "s"))
	must(t, "init")
	g, _, ptr = chain(t, "the bytes the Result cited\n")
	gitIn(t, top, "init", "-q")
	gitIn(t, top, "add", "-f", "-A")
	gitIn(t, top, "commit", "-qm", "store")
	return top, g, ptr
}

// Line 8: hyp import adds each event once, with the bytes of each file it
// names, or not at all. When the checkout holds a named file with other
// bytes than its event stores, or not at all, the import exits 1 and
// writes nothing; from a commit that holds the bytes it adds every event,
// and a second import adds nothing.
func TestImportRefusesAnEventWithoutItsBytes(t *testing.T) {
	for _, how := range []string{"changed", "removed"} {
		t.Run(how, func(t *testing.T) {
			top, g, ptr := storeBeforeGit(t)
			if how == "changed" {
				writeFile(t, ptr, "changed after the Result\n")
			} else if err := os.Remove(ptr); err != nil {
				t.Fatal(err)
			}
			refused(t, hyp(t, "import", "."), ptr)
			noRecord(t, top, "after a refused import")
			want(t, hyp(t, "import", "HEAD:."), 0, "imported 5 events")
			want(t, hyp(t, "show", ptr), 0, "the bytes the Result cited")
			want(t, hyp(t, "status", g), 0, "")
			tip := gitIn(t, top, "rev-parse", "refs/hyp/record")
			want(t, hyp(t, "import", "HEAD:."), 0, "holds every event\n")
			if again := gitIn(t, top, "rev-parse", "refs/hyp/record"); again != tip {
				t.Fatalf("a second import moved the record from %s to %s", tip, again)
			}
		})
	}
}

// Line 8: an import of two stores, of which the second lacks a named file,
// adds nothing of either.
func TestImportOfTwoStoresAddsNothingWhenOneIsRefused(t *testing.T) {
	top, _, _ := storeBeforeGit(t)
	other := filepath.Join(top, "t")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(other)
	must(t, "init")
	dropRecord(t, top)
	_, _, ptr := chain(t, "t's bytes\n")
	if err := os.Remove(ptr); err != nil {
		t.Fatal(err)
	}
	t.Chdir(top)
	refused(t, hyp(t, "import", "s", "t"), ptr)
	noRecord(t, top, "after a refused import of two stores")
}

// Line 8: an event that pins a bundle the source lacks is not imported, and
// nothing is; from a commit that holds the bundle with its pinned hash it
// is, with the bundle.
func TestImportRefusesAnEventWithoutItsBundle(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, "s", "README.md"), "r\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, "s"))
	must(t, "init")
	dropRecord(t, repo)
	bundleIn(t, filepath.Join(repo, "s"), "kit", "", nil)
	_, r := recipeChain(t)
	gitIn(t, repo, "add", "-f", "-A")
	gitIn(t, repo, "commit", "-qm", "store")
	if err := os.RemoveAll(filepath.Join(repo, "s", "recipes")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	refused(t, hyp(t, "import", "s"), "kit@1")
	noRecord(t, repo, "after a refused import")
	want(t, hyp(t, "import", "HEAD:s"), 0, "imported")
	if ls := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record"); !strings.Contains(ls, "s/recipes/kit@1/recipe.md") {
		t.Fatalf("the record lacks the bundle kit@1 after the import:\n%s", ls)
	}
	t.Chdir(filepath.Join(repo, "s"))
	if out := hyp(t, "show", r).stdout; strings.Contains(out, "questionable") {
		t.Fatalf("%s reads questionable after the import:\n%s", r, out)
	}
}

// Lines 8 and 2: a clone that lacks a subject's commit imports nothing;
// once the commit is fetched, the import adds every event and the record
// keeps the commit after its branch goes and git gc runs.
func TestImportRefusesAnEventWithoutItsSubjectCommit(t *testing.T) {
	gitConfig(t)
	a, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, a, "init", "-q")
	writeFile(t, filepath.Join(a, "s", "README.md"), "r\n")
	gitIn(t, a, "add", "-A")
	gitIn(t, a, "commit", "-qm", "base")
	gitIn(t, a, "checkout", "-q", "-b", "feat")
	writeFile(t, filepath.Join(a, "feat.txt"), "only on feat\n")
	gitIn(t, a, "add", "feat.txt")
	gitIn(t, a, "commit", "-qm", "feat")
	c := gitIn(t, a, "rev-parse", "HEAD")
	gitIn(t, a, "checkout", "-q", "main")
	t.Chdir(filepath.Join(a, "s"))
	must(t, "init")
	dropRecord(t, a)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	must(t, "subject", h, "--base", c, "--why", "w")
	gitIn(t, a, "add", "-f", "-A")
	gitIn(t, a, "commit", "-qm", "store")
	b := filepath.Join(t.TempDir(), "b")
	if out, err := exec.Command("git", "clone", "-q", "--no-local", "--single-branch", "--branch", "main", a, b).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v %s", err, out)
	}
	t.Chdir(b)
	refused(t, hyp(t, "import", "s"), c)
	noRecord(t, b, "after a refused import")
	gitIn(t, b, "fetch", "-q", "origin", "feat:refs/remotes/origin/feat")
	want(t, hyp(t, "import", "s"), 0, "imported 4 events")
	gitIn(t, b, "update-ref", "-d", "refs/remotes/origin/feat")
	gitIn(t, b, "reflog", "expire", "--expire=now", "--all")
	gitIn(t, b, "gc", "-q", "--prune=now")
	if exec.Command("git", "-C", b, "merge-base", "--is-ancestor", c, "refs/hyp/record").Run() != nil {
		t.Fatalf("after its branch is deleted and gc, the record does not reach the subject commit %s", c[:12])
	}
}

// Line 8: a work-tree store that holds an id another store of the record
// holds is not imported, and the record stays as it was.
func TestImportRefusesAnIDAnotherStoreHolds(t *testing.T) {
	repo := recordRepo(t)
	g, _, _ := chain(t, "s's bytes\n")
	tip := gitIn(t, repo, "rev-parse", "refs/hyp/record")
	data := gitIn(t, repo, "show", "refs/hyp/record:s/.tree/events/a.jsonl")
	writeFile(t, filepath.Join(repo, "w", ".tree", "events", "b.jsonl"), data+"\n")
	t.Chdir(repo)
	refused(t, hyp(t, "import", "w"), "store s")
	if again := gitIn(t, repo, "rev-parse", "refs/hyp/record"); again != tip {
		t.Fatalf("a refused import moved the record from %s to %s", tip, again)
	}
	t.Chdir(filepath.Join(repo, "s"))
	want(t, hyp(t, "show", g), 0, "")
}

// Line 4: a write's new id sorts after every id the record holds, also an
// id of another store that is newer than every id of the store written.
func TestNewIDSortsAfterEveryStoresID(t *testing.T) {
	repo := recordRepo(t)
	must(t, "add", "goal", "--title", "first", "--why", "w")
	// A work-tree store whose one event has an id two days ahead.
	other := t.TempDir()
	t.Chdir(other)
	t.Setenv("HYP_ACTOR", "b")
	must(t, "init")
	must(t, "add", "goal", "--title", "ahead", "--why", "w")
	id := lastID(t)
	ahead := ulid.MustNew(ulid.Timestamp(time.Now().Add(48*time.Hour)), bytes.NewReader(bytes.Repeat([]byte{7}, 16))).String()
	p := filepath.Join(other, ".tree", "events", "b.jsonl")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(id)) {
		t.Fatalf("setup: the events file does not hold %s", id)
	}
	writeFile(t, filepath.Join(repo, "w", ".tree", "events", "b.jsonl"), strings.ReplaceAll(string(b), id, ahead))
	t.Chdir(repo)
	want(t, hyp(t, "import", "w"), 0, "imported 1 events")
	t.Chdir(filepath.Join(repo, "s"))
	t.Setenv("HYP_ACTOR", "a")
	must(t, "add", "goal", "--title", "next", "--why", "w")
	if next := lastID(t); next <= ahead {
		t.Fatalf("the new id %s sorts before %s, the newest id the record holds (store w)", next, ahead)
	}
}

// lastID is the greatest event id that hyp events prints from the working
// directory's store.
func lastID(t *testing.T) string {
	t.Helper()
	last := ""
	for _, e := range eventsJSON(t, hyp(t, "events", "--json").stdout) {
		var head struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(e), &head); err != nil {
			t.Fatal(err)
		}
		last = max(last, head.ID)
	}
	return last
}

// Line 6: hyp init and hyp import exit 1 and write nothing for a store
// folder whose name is not plain: a character outside ASCII letters,
// digits, '.', '_', '-', '@' and '/', or a part that differs only in case
// from another name in its folder. A plain name is taken.
func TestStoreFolderNameIsPlain(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	for _, name := range []string{"café", "a b", "x:y", "q~r", "ok/new\nline"} {
		dir := filepath.Join(repo, filepath.FromSlash(name))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(dir)
		refused(t, hyp(t, "init"), "store folder")
		if _, err := os.Stat(filepath.Join(dir, ".tree")); err == nil {
			t.Fatalf("a refused hyp init in %q wrote a store folder", name)
		}
		noRecord(t, repo, "after a refused hyp init in "+name)
		// The same folder, holding a work-tree store, is not imported.
		writeFile(t, filepath.Join(dir, ".tree", "events", "a.jsonl"), "")
		t.Chdir(repo)
		refused(t, hyp(t, "import", dir), "store folder")
		noRecord(t, repo, "after a refused hyp import of "+name)
		if err := os.RemoveAll(filepath.Join(dir, ".tree")); err != nil {
			t.Fatal(err)
		}
	}
	// A part named as git's folder: discovery takes the folder above it as
	// the repository root, and the name below that root starts with it.
	gitPart := filepath.Join(repo, "g", ".git", "x")
	if err := os.MkdirAll(gitPart, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(gitPart)
	refused(t, hyp(t, "init"), "store folder")
	if _, err := os.Stat(filepath.Join(gitPart, ".tree")); err == nil {
		t.Fatal("a refused hyp init below a folder named as git's wrote a store folder")
	}
	noRecord(t, repo, "after a refused hyp init below a folder named as git's")
	// Two names that differ only in case, on a filesystem that keeps both.
	if err := os.Mkdir(filepath.Join(repo, "Data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(repo, "data"), 0o755); err == nil {
		t.Chdir(filepath.Join(repo, "data"))
		refused(t, hyp(t, "init"), "Data")
		noRecord(t, repo, "after a refused hyp init beside a folder of other case")
	}
	plain := filepath.Join(repo, "ok-1.2_@v", "sub")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(plain)
	must(t, "init")
	want(t, hyp(t, "add", "goal", "--title", "plain", "--why", "w"), 0, "")
}

// Line 9: a store folder reached through a link inside the repository whose
// target is outside it is no store of the repository: a command run there,
// hyp init there and hyp import naming it each exit 1.
func TestStoreFolderOutsideTheRepository(t *testing.T) {
	repo := recordRepo(t)
	outside, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(outside)
	must(t, "init")
	g := must(t, "add", "goal", "--title", "outside", "--why", "w")
	if err := os.Symlink(outside, filepath.Join(repo, "ln")); err != nil {
		t.Fatal(err)
	}
	empty, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(empty, filepath.Join(repo, "ln2")); err != nil {
		t.Fatal(err)
	}
	tip := gitIn(t, repo, "rev-parse", "refs/hyp/record")
	in := filepath.Join(repo, "ln")
	t.Chdir(in)
	t.Setenv("PWD", in)
	refused(t, hyp(t, "status", g), "outside the repository")
	refused(t, hyp(t, "add", "goal", "--title", "x", "--why", "w"), "outside the repository")
	in2 := filepath.Join(repo, "ln2")
	t.Chdir(in2)
	t.Setenv("PWD", in2)
	refused(t, hyp(t, "init"), "outside the repository")
	if _, err := os.Stat(filepath.Join(empty, ".tree")); err == nil {
		t.Fatal("a refused hyp init wrote a store folder outside the repository")
	}
	t.Chdir(repo)
	t.Setenv("PWD", repo)
	refused(t, hyp(t, "import", "ln"), "outside the repository")
	if again := gitIn(t, repo, "rev-parse", "refs/hyp/record"); again != tip {
		t.Fatalf("a refused command moved the record from %s to %s", tip, again)
	}
}

// Line 1: a named file whose path holds a newline is kept at that path. A
// path that ends in a newline and an index line naming the store's events
// file changes no event: every earlier event stays printed, and the
// file's bytes print, after git clean -fdx; hyp sync joins two records
// that each hold such a path.
func TestNamedPathWithANewline(t *testing.T) {
	repo := recordRepo(t)
	g, r, ptr := chain(t, "first\n")
	before := hyp(t, "events", "--json").stdout
	shownR := hyp(t, "show", r).stdout
	empty := gitIn(t, repo, "hash-object", "-t", "blob", "/dev/null")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--rubric", "reading", "--why", "w")
	weird := []string{
		".tree/evidence/" + e + "/a\nb.txt",
		".tree/evidence/" + e + "/c\n100644 " + empty + "\ts/.tree/events/a.jsonl",
	}
	var rs []string
	for i, w := range weird {
		writeFile(t, w, "weird "+string(rune('a'+i))+"\n")
		rs = append(rs, must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", w, "--why", "w"))
	}
	earlier := func(when string) {
		t.Helper()
		after := map[string]bool{}
		for _, e := range eventsJSON(t, hyp(t, "events", "--json").stdout) {
			after[string(e)] = true
		}
		for _, e := range eventsJSON(t, before) {
			if !after[string(e)] {
				t.Fatalf("%s: hyp events no longer prints the earlier event %s", when, e)
			}
		}
		want(t, hyp(t, "show", r), 0, "")
		if got := hyp(t, "show", r).stdout; got != shownR {
			t.Fatalf("%s: hyp show %s prints\n%s\nwant\n%s", when, r, got, shownR)
		}
		want(t, hyp(t, "show", ptr), 0, "first")
		want(t, hyp(t, "status", g), 0, "")
	}
	check := func(when string) {
		t.Helper()
		earlier(when)
		for i, w := range weird {
			want(t, hyp(t, "show", w), 0, "weird "+string(rune('a'+i)))
			want(t, hyp(t, "show", rs[i]), 0, "")
		}
	}
	check("after the writes")
	gitIn(t, repo, "clean", "-fdxq")
	check("after git clean -fdx")

	// A remote, and a second clone that writes its own such path.
	remote := filepath.Join(t.TempDir(), "remote")
	gitIn(t, repo, "init", "-q", "--bare", remote)
	gitIn(t, repo, "remote", "add", "origin", remote)
	gitIn(t, repo, "push", "-q", "origin", "main")
	want(t, hyp(t, "sync"), 0, "synced")
	other := filepath.Join(t.TempDir(), "b")
	// A clone by the pack protocol, never by copying the remote's files.
	gitIn(t, repo, "clone", "-q", "--no-local", remote, other)
	t.Chdir(filepath.Join(other, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	t.Setenv("HYP_ACTOR", "b")
	wb := ".tree/evidence/" + e + "/d\ne.txt"
	writeFile(t, wb, "from b\n")
	rb := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", wb, "--why", "w")
	t.Chdir(filepath.Join(repo, "s"))
	t.Setenv("HYP_ACTOR", "a")
	writeFile(t, ".tree/evidence/"+e+"/f\ng.txt", "from a\n")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ".tree/evidence/"+e+"/f\ng.txt", "--why", "w")
	want(t, hyp(t, "sync"), 0, "synced")
	t.Chdir(filepath.Join(other, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	t.Chdir(filepath.Join(repo, "s"))
	want(t, hyp(t, "sync"), 0, "synced")
	for _, dir := range []string{filepath.Join(repo, "s"), filepath.Join(other, "s")} {
		t.Chdir(dir)
		earlier("after hyp sync in " + dir)
		want(t, hyp(t, "show", wb), 0, "from b")
		want(t, hyp(t, "show", ".tree/evidence/"+e+"/f\ng.txt"), 0, "from a")
		want(t, hyp(t, "show", rb), 0, "")
		for i, w := range weird {
			want(t, hyp(t, "show", w), 0, "weird "+string(rune('a'+i)))
		}
	}
}

// eventsJSON is each event hyp events --json prints, as its bytes.
func eventsJSON(t *testing.T, out string) []string {
	t.Helper()
	var v struct {
		Events []json.RawMessage `json:"events"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("hyp events --json: %v: %s", err, out)
	}
	var evs []string
	for _, e := range v.Events {
		evs = append(evs, string(e))
	}
	return evs
}

// Line 6: on a filesystem that folds letter case, a folder named .GIT is
// git's folder: discovery takes the folder above it as the repository
// root. hyp init below it exits 1 and writes nothing, and a write in a
// work-tree store copied there exits 1, so no write exits 0 on an event
// the record cannot hold.
func TestStoreBelowAFolderNamedAsGitsInOtherCase(t *testing.T) {
	repo := recordRepo(t)
	dir := filepath.Join(repo, "g", ".GIT", "x")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(repo, "g", ".git")); err != nil {
		t.Skip("this filesystem keeps letter case apart")
	}
	tip := gitIn(t, repo, "rev-parse", "refs/hyp/record")
	t.Chdir(dir)
	refused(t, hyp(t, "init"), "store folder")
	if _, err := os.Stat(filepath.Join(dir, ".tree")); err == nil {
		t.Fatal("a refused hyp init below g/.GIT wrote a store folder")
	}
	writeFile(t, filepath.Join(dir, ".tree", "events", "a.jsonl"), "")
	refused(t, hyp(t, "add", "goal", "--title", "x", "--why", "w"), "store folder")
	if again := gitIn(t, repo, "rev-parse", "refs/hyp/record"); again != tip {
		t.Fatalf("a refused command moved the record from %s to %s", tip, again)
	}
}

// Line 9: hyp sandbox create --store naming, through a link inside the host
// repository, a store folder outside it exits 1.
func TestSandboxStoreOutsideTheRepository(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	t.Setenv("HOME", t.TempDir())
	repo := recordRepo(t)
	outside, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(outside)
	must(t, "init")
	g := must(t, "add", "goal", "--title", "outside", "--why", "w")
	if err := os.Symlink(outside, filepath.Join(repo, "ln")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	t.Setenv("PWD", repo)
	refused(t, hyp(t, "sandbox", "create", "sb1", "--host", repo, "--base", "HEAD", "--recipe", "none", "--store", filepath.Join(repo, "ln"), "--cut", g), "outside the repository")
}

// Lines 1 and 6: git's index takes a part that is .git in any letter case
// with trailing dots, such as .git. or .git.., as git's folder (a filesystem
// that drops trailing dots reads it so): update-index ignores a path below it
// and still exits 0. hyp init in such a store folder exits 1 and writes
// nothing, and a write in a work-tree store copied there exits 1, so no write
// exits 0 on an event the record cannot hold.
func TestStoreBelowAFolderGitsIndexTakesAsGits(t *testing.T) {
	repo := recordRepo(t)
	tip := gitIn(t, repo, "rev-parse", "refs/hyp/record")
	for i, part := range []string{".git.", ".git..", ".GIT."} {
		t.Run(part, func(t *testing.T) {
			// Each case has its own parent: .git. and .GIT. are one folder
			// on a filesystem that folds case.
			dir := filepath.Join(repo, "a"+strings.Repeat("a", i), part, "x")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			t.Chdir(dir)
			refused(t, hyp(t, "init"), "store folder")
			if _, err := os.Stat(filepath.Join(dir, ".tree")); err == nil {
				t.Fatalf("a refused hyp init below %s wrote a store folder", part)
			}
			writeFile(t, filepath.Join(dir, ".tree", "events", "a.jsonl"), "")
			refused(t, hyp(t, "add", "goal", "--title", "x", "--why", "w"), "store folder")
			if again := gitIn(t, repo, "rev-parse", "refs/hyp/record"); again != tip {
				t.Fatalf("a refused command moved the record from %s to %s", tip, again)
			}
		})
	}
}
