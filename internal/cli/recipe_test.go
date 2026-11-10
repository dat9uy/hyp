package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/store"
)

// recordedHash is the recipe_hash of the one Experiment in the store at tree.
func recordedHash(t *testing.T, tree string) string {
	t.Helper()
	p, _, err := store.Load(tree)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range p.Nodes {
		if n.Kind == event.KindExperiment {
			return n.RecipeHash
		}
	}
	t.Fatal("no Experiment")
	return ""
}

func TestRecipeHashMatchesAddExp(t *testing.T) {
	repo, tree, q := libraryStore(t)
	dir := filepath.Join(repo, "recipes", "boot@1")
	if err := os.WriteFile(filepath.Join(dir, "check.py"), []byte("print('ok')\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "recipe.md"), "---\nname: boot\nversion: 1\nformat: 1\n---\n\n## Standalone\n\nRun check: ./check.py\n")
	writeFile(t, filepath.Join(dir, "known-answers", "ok", "expect.json"), `{"step": "standalone", "command": "check", "exit": 0, "stdout": "ok\n"}`)
	writeFile(t, filepath.Join(dir, "known-answers", "ok", "input", "a"), "")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--rubric", "reading", "--why", "w")
	rec := recordedHash(t, tree)

	r := hyp(t, "bundle", "hash", dir)
	want(t, r, 0, rec)
	if r.stdout != rec+"\n" {
		t.Fatalf("stdout %q, want the hash alone", r.stdout)
	}
	r = hyp(t, "bundle", "hash", "--json", dir)
	var got recipeHash
	if err := json.Unmarshal([]byte(r.stdout), &got); err != nil || r.code != 0 {
		t.Fatalf("--json: exit %d, %v: %q", r.code, err, r.stdout)
	}
	if want := (recipeHash{Ref: "boot@1", Name: "boot", Version: "1", Path: dir, SHA256: rec}); got != want {
		t.Fatalf("--json = %+v, want %+v", got, want)
	}
	// The directory name is not hashed: a legacy copy hashes the same.
	legacy := filepath.Join(t.TempDir(), "boot")
	if err := os.CopyFS(legacy, os.DirFS(dir)); err != nil {
		t.Fatal(err)
	}
	want(t, hyp(t, "bundle", "hash", legacy), 0, rec)
	t.Chdir(dir)
	want(t, hyp(t, "bundle", "hash", "."), 0, rec)

	want(t, hyp(t, "bundle", "hash"), 2, "usage: hyp bundle hash DIR [--json]")
	want(t, hyp(t, "bundle", "list"), 2, "usage: hyp bundle hash DIR [--json]")
	want(t, hyp(t, "bundle", "hash", t.TempDir()), 2, "recipe.md")
}

// A folder without @version hashes without a version check, even one whose
// recipe.md gives no version.
func TestRecipeHashFolderWithoutVersion(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "x")
	writeFile(t, filepath.Join(dir, "recipe.md"), "---\nname: x\n---\n")
	_, _, hash, err := recipe.HashOnly(dir)
	if err != nil {
		t.Fatal(err)
	}
	want(t, hyp(t, "bundle", "hash", dir), 0, hash)
	writeFile(t, filepath.Join(dir, "recipe.md"), "---\nname: y\n---\n")
	want(t, hyp(t, "bundle", "hash", dir), 1, "recipe name mismatch")
}

func TestEvidenceInLibraryRecipe(t *testing.T) {
	repo, tree, q := libraryStore(t)
	lib := filepath.Join(repo, "recipes", "boot@1")
	if err := os.MkdirAll(filepath.Join(lib, "cases"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lib, "cases", "known.txt"), []byte("k\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	rec := func(ev string) result {
		return hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--evidence", ev, "--why", "w")
	}

	want(t, rec("recipes/boot@1/cases/known.txt"), 0, "result.record")
	if _, err := os.Stat(filepath.Join(tree, "recipes")); !os.IsNotExist(err) {
		t.Fatalf("the store has its own recipes/ (%v)", err)
	}
	p, _, err := store.Load(tree)
	if err != nil {
		t.Fatal(err)
	}
	var cited []string
	for _, n := range p.Nodes {
		if n.Kind == event.KindResult {
			cited = append(cited, n.Evidence...)
		}
	}
	if strings.Join(cited, ",") != "recipes/boot@1/cases/known.txt" {
		t.Fatalf("evidence recorded as %v", cited)
	}
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")

	// A ref found nowhere names both directories searched.
	r := rec("recipes/nope@1/a.txt")
	want(t, r, 2, "evidence recipes/nope@1/a.txt: recipe nope@1 not found; searched "+filepath.Join(tree, "recipes")+", ")
	want(t, r, 2, filepath.Join(repo, "recipes"))
	want(t, rec("recipes/boot@1/cases/missing.txt"), 2, "evidence recipes/boot@1/cases/missing.txt not found in recipe boot@1 at "+lib)
	want(t, rec("recipes/boot@1/../boot@1/cases/known.txt"), 0, "result.record")
	for _, ev := range []string{"recipes/boot@1", "recipes/boot@1/cases"} {
		want(t, rec(ev), 2, "evidence "+ev+" is a directory; cite a file in it")
	}
	want(t, rec("recipes/../../recipes/boot@1/cases/known.txt"), 2, "not found under .tree/evidence or recipes")

	// The same ref beside the store with other content is refused, naming both.
	writeLib(t, tree, "boot@1", "---\nname: boot\nversion: 1\nformat: 1\n---\n\nA local edit.\n")
	r = rec("recipes/boot@1/recipe.md")
	want(t, r, 1, filepath.Join("tree", "recipes", "boot@1")+" and ")
	want(t, r, 1, " differ; one ref must hash the same everywhere")
}

// writeFile writes data to path, making its directory.
func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

// bundleHash is the hash recipe.Open gives the bundle in dir.
func bundleHash(t *testing.T, dir string) string {
	t.Helper()
	r, err := recipe.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	return r.Hash
}

func TestEvidencePinsLibraryBundle(t *testing.T) {
	repo, tree, q := libraryStore(t)
	boot := filepath.Join(repo, "recipes", "boot@1")
	cases := filepath.Join(repo, "recipes", "cases@1")
	writeFile(t, filepath.Join(boot, "known.txt"), "k\n")
	writeLib(t, repo, "cases@1", "---\nname: cases\nversion: 1\nformat: 1\n---\n")
	writeFile(t, filepath.Join(cases, "a.txt"), "a\n")
	bootHash, casesHash := bundleHash(t, boot), bundleHash(t, cases)

	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	rec := func(evidence ...string) result {
		args := []string{"record", "--exp", e, "--verdict", h + "=supports", "--observation", "o", "--why", "w"}
		for _, ev := range evidence {
			args = append(args, "--evidence", ev)
		}
		return hyp(t, args...)
	}
	evDir := filepath.Join(tree, ".tree", "evidence", e)
	writeFile(t, filepath.Join(evDir, "log.txt"), "l\n")
	r := rec("recipes/cases@1/a.txt", "recipes/boot@1/known.txt", ".tree/evidence/"+e+"/log.txt", "https://example.com/x")
	want(t, r, 0, "result.record")
	res := strings.Fields(r.stdout)[0]

	p, _, err := store.Load(tree)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	for _, n := range p.Nodes {
		if n.Kind == event.KindResult {
			got = n.EvidenceHashes
		}
	}
	wantHashes := map[string]string{"recipes/cases@1/a.txt": casesHash, "recipes/boot@1/known.txt": bootHash}
	if !reflect.DeepEqual(got, wantHashes) {
		t.Fatalf("evidence_hashes = %v, want %v", got, wantHashes)
	}
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")

	// An in-place edit of a cited bundle flags the Result, naming the pointer
	// and both hashes; recording more Results under boot@1 still works.
	writeFile(t, filepath.Join(cases, "a.txt"), "edited\n")
	edited := bundleHash(t, cases)
	r = hyp(t, "check")
	want(t, r, 0, "warn: questionable evidence "+res+" (recipes/cases@1/a.txt: recorded "+casesHash+", now "+edited+")")
	want(t, r, 0, "check: 0 errors, 1 warnings")
	want(t, hyp(t, "status"), 0, "QUESTIONABLE 1 results")
	want(t, rec("recipes/cases@1/a.txt"), 0, "result.record")

	// A cited bundle that no longer resolves reads as missing.
	if err := os.RemoveAll(cases); err != nil {
		t.Fatal(err)
	}
	want(t, hyp(t, "check"), 0, "(recipes/cases@1/a.txt: recorded "+casesHash+", now missing)")
}

func TestRecordRefusesChangedExperimentRecipe(t *testing.T) {
	repo, tree, q := libraryStore(t)
	boot := filepath.Join(repo, "recipes", "boot@1")
	writeFile(t, filepath.Join(boot, "known.txt"), "k\n")
	pinned := bundleHash(t, boot)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	res := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o",
		"--evidence", "recipes/boot@1/known.txt", "--why", "w")

	writeFile(t, filepath.Join(boot, "known.txt"), "edited\n")
	now := bundleHash(t, boot)
	r := hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w")
	want(t, r, 1, "refused: recipe boot@1 changed since "+e+" pinned it (recorded "+pinned+", now "+now+"); restore it, or add a new Experiment or a new recipe version")
	r = hyp(t, "check")
	want(t, r, 0, "warn: questionable evidence "+res+" (recipe boot changed)\nwarn: questionable evidence "+res+" (recipes/boot@1/known.txt: recorded "+pinned+", now "+now+")\n")
	want(t, r, 0, "check: 0 errors, 2 warnings")

	// A new version is the remedy: a new Experiment under boot@2 records.
	writeLib(t, repo, "boot@2", "---\nname: boot\nversion: 2\nformat: 1\n---\n")
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "boot@2", "--decided-by", "reading", "--why", "w")
	must(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "o", "--why", "w")

	// A recipe that no longer loads is refused too.
	if err := os.RemoveAll(boot); err != nil {
		t.Fatal(err)
	}
	r = hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w")
	want(t, r, 1, "refused: recipe boot@1 of "+e+" no longer loads (recipe boot@1 not found; searched "+filepath.Join(tree, "recipes"))
}

// An edit undone between the evidence hashing and the locked check would pin
// two hashes for one ref; record refuses and writes nothing.
func TestRecordRefusesSplitEvidencePin(t *testing.T) {
	repo, tree, q := libraryStore(t)
	boot := filepath.Join(repo, "recipes", "boot@1")
	writeFile(t, filepath.Join(boot, "known.txt"), "k\n")
	pinned := bundleHash(t, boot)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")

	writeFile(t, filepath.Join(boot, "known.txt"), "edited\n")
	edited := bundleHash(t, boot)
	recordHashed = func() { writeFile(t, filepath.Join(boot, "known.txt"), "k\n") }
	t.Cleanup(func() { recordHashed = nil })
	args := []string{"record", "--exp", e, "--verdict", h + "=supports", "--observation", "o",
		"--evidence", "recipes/boot@1/known.txt", "--why", "w"}
	want(t, hyp(t, args...), 1, "refused: recipe boot@1 changed while recording: its evidence hashed "+edited+", "+e+" pinned "+pinned)

	recordHashed = nil
	want(t, hyp(t, args...), 0, "result.record")
	p, _, err := store.Load(tree)
	if err != nil {
		t.Fatal(err)
	}
	var got []map[string]string
	for _, n := range p.Nodes {
		if n.Kind == event.KindResult {
			got = append(got, n.EvidenceHashes)
		}
	}
	wantHashes := []map[string]string{{"recipes/boot@1/known.txt": pinned}}
	if !reflect.DeepEqual(got, wantHashes) {
		t.Fatalf("results' evidence_hashes = %v, want %v", got, wantHashes)
	}
}

// A store's own recipes/ path without a version keeps its old meaning.
func TestEvidenceInStoreRecipesUnchanged(t *testing.T) {
	dir, _, q := setup(t)
	if err := os.MkdirAll(filepath.Join(dir, "recipes", "boot"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "recipes", "boot", "notes.txt"), []byte("n\n"), 0o644)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	want(t, hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o",
		"--evidence", "recipes/boot/notes.txt", "--why", "w"), 0, "result.record")
}

func TestRecipeCheckFindsDriftInEveryTree(t *testing.T) {
	repo, tree, q := libraryStore(t)
	boot := filepath.Join(repo, "recipes", "boot@1")
	writeFile(t, filepath.Join(boot, "known.txt"), "k\n")
	old := bundleHash(t, boot)

	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	res := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--evidence", "recipes/boot@1/known.txt", "--why", "w")
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")

	// A second store elsewhere in the repository.
	other := filepath.Join(repo, "b", "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(other)
	must(t, "init")
	g2 := must(t, "add", "goal", "--title", "G", "--why", "start")
	q2 := must(t, "add", "dq", "--parent", g2, "--title", "Q", "--why", "ask")
	h2 := must(t, "add", "hyp", "--parent", q2, "--title", "H", "--why", "w")
	must(t, "add", "exp", "--parent", h2, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	t.Chdir(tree)

	// A nested repository holding a copy of the store: boot@1 resolves nowhere
	// from inside it, so it would drift if it were walked.
	nested := filepath.Join(repo, "wt")
	if err := os.MkdirAll(filepath.Join(nested, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(filepath.Join(nested, ".tree"), os.DirFS(filepath.Join(tree, ".tree"))); err != nil {
		t.Fatal(err)
	}

	r := hyp(t, "bundle", "check")
	want(t, r, 0, "recipe check: 2 trees, 4 pins, 0 drift")

	writeFile(t, filepath.Join(boot, "known.txt"), "edited\n")
	now := bundleHash(t, boot)
	r = hyp(t, "bundle", "check")
	if r.code != 1 {
		t.Fatalf("exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
	for _, s := range []string{
		"drift: tree " + res + " recipes/boot@1/: recorded " + old + ", now " + now,
		"drift: tree " + res + " recipes/boot@1/known.txt: recorded " + old + ", now " + now,
		"drift: tree " + e2 + " recipes/boot@1/: recorded " + old + ", now " + now,
		"drift: b/other E-",
		"recipe check: 2 trees, 4 pins, 4 drift",
	} {
		if !strings.Contains(r.stdout, s) {
			t.Errorf("stdout %q lacks %q", r.stdout, s)
		}
	}
	if strings.Contains(r.stdout, "wt") {
		t.Errorf("stdout %q names a nested repository", r.stdout)
	}

	t.Chdir(repo)
	want(t, hyp(t, "bundle", "check"), 1, "4 drift")

	t.Chdir(t.TempDir())
	if r = hyp(t, "bundle", "check"); r.code != 3 {
		t.Fatalf("outside a repository: exit %d, stderr %q", r.code, r.stderr)
	}
	want(t, hyp(t, "bundle", "check", "x"), 2, "usage: hyp bundle check")
}

// A leftover recipes/boot/ that hashes as a pin to boot@1 gets one rename
// hint, however many pins it explains; one that hashes differently gets none.
func TestRecipeCheckHintsRename(t *testing.T) {
	repo, tree, q := libraryStore(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	for range 2 {
		must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	}
	rec := recordedHash(t, tree)
	lib := filepath.Join(repo, "recipes")
	if err := os.Rename(filepath.Join(lib, "boot@1"), filepath.Join(lib, "boot")); err != nil {
		t.Fatal(err)
	}
	hint := "hint: recipes/boot/ holds boot@1 (same hash) in a folder whose name has no version; rename it: git mv recipes/boot recipes/boot@1\n"

	r := hyp(t, "bundle", "check")
	want(t, r, 1, "recipe check found 2 failures; rename each folder a hint names")
	lines := strings.Split(r.stdout, "\n")
	if !strings.HasPrefix(lines[0], "drift: tree E-") || !strings.HasSuffix(lines[0], "recorded "+rec+", now missing") ||
		lines[1]+"\n" != hint || !strings.HasPrefix(lines[2], "drift: tree E-") {
		t.Fatalf("stdout %q, want drift, hint, drift", r.stdout)
	}
	if n := strings.Count(r.stdout, "hint: "); n != 1 {
		t.Fatalf("stdout has %d hints, want 1: %q", n, r.stdout)
	}

	writeFile(t, filepath.Join(lib, "boot", "notes.txt"), "edited\n")
	r = hyp(t, "bundle", "check")
	want(t, r, 1, "recipe check found 2 failures; a changed bundle needs a new version (hyp bundle new)")
	if strings.Contains(r.stdout, "hint: ") {
		t.Fatalf("stdout %q hints a folder that hashes differently", r.stdout)
	}
}

func TestRecipeNewCopiesNextVersion(t *testing.T) {
	repo, tree, q := libraryStore(t)
	boot := filepath.Join(repo, "recipes", "boot@1")
	writeFile(t, filepath.Join(boot, "run.sh"), "#!/bin/sh\necho ok\n")
	if err := os.Chmod(filepath.Join(boot, "run.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(boot, "cases", "a.txt"), "a\n")
	writeFile(t, filepath.Join(boot, ".hidden"), "h\n")
	writeFile(t, filepath.Join(boot, "__pycache__", "x.pyc"), "p")
	writeFile(t, filepath.Join(boot, "recipe.md"), "---\nname: boot\nversion: 1\nsubject: s\n---\n\n# boot\n\n## Standalone\n\nRun ok: ./run.sh\n")
	writeFile(t, filepath.Join(boot, "known-answers", "ok", "expect.json"), `{"step": "standalone", "command": "ok", "exit": 0, "stdout": "ok\n"}`)
	writeFile(t, filepath.Join(boot, "known-answers", "ok", "input", "a"), "")
	h1 := bundleHash(t, boot)

	r := hyp(t, "bundle", "new", "boot@2", "--from", "boot@1")
	want(t, r, 0, "from: boot@1 "+h1)
	if first := strings.SplitN(r.stdout, "\n", 2)[0]; !strings.HasSuffix(first, filepath.Join("recipes", "boot@2")) {
		t.Fatalf("first line %q", first)
	}
	two := filepath.Join(repo, "recipes", "boot@2")
	md, err := os.ReadFile(filepath.Join(two, "recipe.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got, w := string(md), "---\nname: boot\nversion: 2\nfrom: boot@1 "+h1+"\nformat: 1\nsubject: s\n---\n\n# boot\n\n## Standalone\n\nRun ok: ./run.sh\nExternal: TODO\n"; got != w {
		t.Fatalf("recipe.md = %q, want %q", got, w)
	}
	for rel, perm := range map[string]os.FileMode{"run.sh": 0o755, "cases/a.txt": 0o644} {
		a, err := os.ReadFile(filepath.Join(boot, rel))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(two, rel))
		if err != nil {
			t.Fatal(err)
		}
		fi, err := os.Stat(filepath.Join(two, rel))
		if err != nil {
			t.Fatal(err)
		}
		if string(a) != string(b) || fi.Mode().Perm() != perm {
			t.Errorf("%s: %q mode %v, want %q mode %v", rel, b, fi.Mode().Perm(), a, perm)
		}
	}
	for _, rel := range []string{".hidden", "__pycache__"} {
		if _, err := os.Lstat(filepath.Join(two, rel)); !os.IsNotExist(err) {
			t.Errorf("%s was copied: %v", rel, err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(repo, "recipes"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".boot@2.new-") {
			t.Errorf("temporary directory %s left behind", e.Name())
		}
	}
	if rc, err := recipe.Open(two); err != nil || rc.Version != "2" {
		t.Fatalf("Open(boot@2) = %v, %v", rc, err)
	}
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@2", "--rubric", "reading", "--why", "w")

	want(t, hyp(t, "bundle", "new", "boot@3", "--from", "boot@2"), 0, "from: boot@2 ")
	md, err = os.ReadFile(filepath.Join(repo, "recipes", "boot@3", "recipe.md"))
	if err != nil {
		t.Fatal(err)
	}
	var froms []string
	for _, line := range strings.Split(string(md), "\n") {
		if strings.HasPrefix(line, "from:") {
			froms = append(froms, line)
		}
	}
	if w := "from: boot@2 " + bundleHash(t, two); len(froms) != 1 || froms[0] != w {
		t.Fatalf("from lines %q, want [%q]", froms, w)
	}

	t.Chdir(repo)
	want(t, hyp(t, "bundle", "new", "boot@4", "--from", "boot@3"), 0, "from: boot@3 ")
	_ = tree
}

func TestRecipeNewRefusals(t *testing.T) {
	repo, _, _ := libraryStore(t)
	want(t, hyp(t, "bundle", "new", "boot@2", "--from", "boot@1"), 0, "from: boot@1 ")

	want(t, hyp(t, "bundle", "new", "boot@2", "--from", "boot@1"), 1, "exists")
	want(t, hyp(t, "bundle", "new", "boot@3", "--from", "boot@1"), 1, "not the next version after boot@1")
	writeLib(t, repo, "boot@5", "---\nname: boot\nversion: 5\n---\n")
	want(t, hyp(t, "bundle", "new", "boot@3", "--from", "boot@2"), 1, "boot@5 exists; the next version is boot@6")

	for _, args := range [][]string{
		{"nope", "--from", "boot@1"},
		{"boot@x", "--from", "boot@1"},
		{"boot@0", "--from", "boot@-1"},
		{"boot@10", "--from", "boot@9"},
	} {
		want(t, hyp(t, append([]string{"bundle", "new"}, args...)...), 2, "bad recipe ref")
	}
	want(t, hyp(t, "bundle", "new", "other@2", "--from", "boot@1"), 2, "name different recipes")
	want(t, hyp(t, "bundle", "new", "boot@2"), 2, "missing --from")
	writeFile(t, filepath.Join(repo, "recipes", "legacy.md"), "---\nname: legacy\nversion: 1\nformat: 1\n---\n")
	want(t, hyp(t, "bundle", "new", "legacy@2", "--from", "legacy@1"), 2, "bad recipe ref: bundle legacy@1 not found; searched ")

	entries, err := os.ReadDir(filepath.Join(repo, "recipes"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if w := []string{"boot@1", "boot@2", "boot@5", "legacy.md"}; !reflect.DeepEqual(names, w) {
		t.Fatalf("recipes = %v, want %v", names, w)
	}
}

// An Experiment cannot follow an instrument (a recipe with an Evidence or
// Standalone command) that has no known-answer case; add exp only looks for
// a case, and the gate runs the cases before each run.
func TestAddExpRefusesInstrumentWithoutKnownAnswers(t *testing.T) {
	repo, tree, q := libraryStore(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	addExp := func(ref string) result {
		t.Helper()
		return hyp(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", ref, "--rubric", "reading", "--why", "w")
	}
	tool := filepath.Join(repo, "recipes", "tool@1")
	writeFile(t, filepath.Join(tool, "recipe.md"), "---\nname: tool\nversion: 1\nformat: 1\n---\n\n## Standalone\n\nRun ok: ./ok.sh\n")
	writeFile(t, filepath.Join(tool, "ok.sh"), "#!/bin/sh\necho ok\n")
	if err := os.Chmod(filepath.Join(tool, "ok.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	const refused = "refused: recipe tool@1 is an instrument with no known answers; add known-answers/NAME/ cases before an Experiment uses it"
	want(t, addExp("tool@1"), 1, refused)

	// An empty known-answers/ holds no case.
	if err := os.MkdirAll(filepath.Join(tool, "known-answers"), 0o755); err != nil {
		t.Fatal(err)
	}
	want(t, addExp("tool@1"), 1, refused)
	p, _, err := store.Load(tree)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range p.Nodes {
		if n.Kind == event.KindExperiment {
			t.Fatalf("Experiment %s added despite the refusals", n.ID)
		}
	}

	writeFile(t, filepath.Join(tool, "known-answers", "ok", "expect.json"), `{"step": "standalone", "command": "ok", "exit": 0, "stdout": "ok\n"}`)
	writeFile(t, filepath.Join(tool, "known-answers", "ok", "input", "a"), "")
	if r := addExp("tool@1"); r.code != 0 || r.stderr != "" {
		t.Fatalf("with a case: %+v", r)
	}
	// A recipe with no command, and none, are not instruments: they need
	// --decided-by reading, and are written with it.
	for _, ref := range []string{"boot@1", "none"} {
		r := hyp(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", ref, "--decided-by", "reading", "--why", "w")
		if r.code != 0 || r.stderr != "" {
			t.Fatalf("%s: %+v", ref, r)
		}
	}
}

// A nearer copy of an old version cannot branch past a newer version in a
// library further up the lookup.
func TestRecipeNewSeesHigherVersionInEveryLibrary(t *testing.T) {
	repo, tree, _ := libraryStore(t)
	writeLib(t, tree, "boot@1", "---\nname: boot\nversion: 1\nformat: 1\n---\n")
	writeLib(t, repo, "boot@3", "---\nname: boot\nversion: 3\n---\n")
	want(t, hyp(t, "bundle", "new", "boot@2", "--from", "boot@1"), 1, "boot@3 exists; the next version is boot@4")
	for _, d := range []string{filepath.Join(tree, "recipes", "boot@2"), filepath.Join(repo, "recipes", "boot@2")} {
		if _, err := os.Lstat(d); !os.IsNotExist(err) {
			t.Fatalf("%s was created: %v", d, err)
		}
	}
}

// bundleIn writes recipes/<name>@1/ under dir: recipe.md with body, and
// each file of scripts (a path below the bundle) executable.
func bundleIn(t *testing.T, dir, name, body string, scripts map[string]string) {
	t.Helper()
	libIn(t, dir, "recipes", name, body, scripts)
}

// libIn writes bundle name@1 into library lib (recipes or rubrics) under dir.
func libIn(t *testing.T, dir, lib, name, body string, scripts map[string]string) {
	t.Helper()
	b := filepath.Join(dir, lib, name+"@1")
	if err := os.MkdirAll(b, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "---\nname: " + name + "\nversion: 1\nformat: 1\n---\n\n" + body
	if err := os.WriteFile(filepath.Join(b, "recipe.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	for f, s := range scripts {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(b, f)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(b, f), []byte(s), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func toolBundle(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	bundleIn(t, dir, "tool", "## Standalone\n\nRun fail: ./fail.sh\nRun ok: ./ok.sh\n", map[string]string{
		"fail.sh":                      "#!/bin/sh\necho boom\nexit 30\n",
		"ok.sh":                        "#!/bin/sh\necho \"got $*\"\n",
		"known-answers/ok/expect.json": `{"step": "standalone", "command": "ok", "args": ["x"], "exit": 0, "stdout": "got x\n"}`,
		"known-answers/ok/input/a":     "",
	})
}

func TestRecipeRunPassesExitCode(t *testing.T) {
	toolBundle(t)
	r := hyp(t, "bundle", "run", "--record", "r1", "tool@1", "ok", "a", "--all")
	if r.code != 0 || !strings.Contains(r.stdout, "got a --all") || !strings.Contains(r.stdout, "record -> r1") {
		t.Fatalf("ok: %+v", r)
	}
	r = hyp(t, "bundle", "run", "--record", "r2", "tool@1", "fail")
	if r.code != 30 || !strings.Contains(r.stdout, "boom") || !strings.Contains(r.stderr, "tool@1 standalone: ./fail.sh exited 30") {
		t.Fatalf("fail: %+v", r)
	}
}

func TestRecipeRunUsage(t *testing.T) {
	toolBundle(t)
	must(t, "bundle", "run", "--record", "r1", "tool@1", "ok")
	for _, args := range [][]string{
		{"tool@1", "ok"},
		{"--record", "r", "tool@1"},
		{"--record", "r3", "tool@1", "nope"},
		{"--record", "r1", "tool@1", "ok"},
	} {
		if r := hyp(t, append([]string{"bundle", "run"}, args...)...); r.code != 2 {
			t.Errorf("hyp bundle run %v: %+v, want exit 2", args, r)
		}
	}
}

func TestRecipeEvidence(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	bundleIn(t, dir, "chk", "## Evidence\n\nRun: ./check.sh\n", map[string]string{
		"check.sh":                       "#!/bin/sh\necho \"kept $1\"\nexit 30\n",
		"known-answers/kept/expect.json": `{"step": "evidence", "command": "./check.sh", "exit": 30, "stdout": "kept $CASE/input\n"}`,
		"known-answers/kept/input/a":     "",
	})
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "sandbox.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs("out")
	if r := hyp(t, "bundle", "evidence", "--record", "e1", "chk@1", "out"); r.code != 30 || !strings.Contains(r.stdout, "kept "+abs) {
		t.Fatalf("evidence: %+v", r)
	}
	if r := hyp(t, "bundle", "evidence", "--record", "e2", "chk@1", "nowhere"); r.code != 2 {
		t.Errorf("missing OUT_DIR: %+v", r)
	}
	if r := hyp(t, "bundle", "evidence", "--record", "e3", "chk@1"); r.code != 2 {
		t.Errorf("no OUT_DIR: %+v", r)
	}
}

func TestRecipeRecordInsideBundleRefused(t *testing.T) {
	toolBundle(t)
	if r := hyp(t, "bundle", "run", "--record", "recipes/tool@1/out", "tool@1", "ok"); r.code != 2 || !strings.Contains(r.stderr, "recipe library") {
		t.Errorf("run: %+v, want exit 2 naming the library", r)
	}
	if _, err := os.Stat(filepath.Join("recipes", "tool@1", "out")); err == nil {
		t.Error("run: record written inside the bundle")
	}

	dir := t.TempDir()
	t.Chdir(dir)
	bundleIn(t, dir, "chk", "## Evidence\n\nRun: ./check.sh\n", map[string]string{
		"check.sh": "#!/bin/sh\nexit 30\n",
	})
	if err := os.MkdirAll("out", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("out", "sandbox.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := hyp(t, "bundle", "evidence", "--record", "out/rec", "chk@1", "out"); r.code != 2 || !strings.Contains(r.stderr, "OUT_DIR") {
		t.Errorf("evidence in OUT_DIR: %+v, want exit 2", r)
	}
	if r := hyp(t, "bundle", "evidence", "--record", "recipes/chk@1/rec", "chk@1", "out"); r.code != 2 || !strings.Contains(r.stderr, "recipe library") {
		t.Errorf("evidence in bundle: %+v, want exit 2 naming the library", r)
	}
	for _, p := range []string{filepath.Join("out", "rec"), filepath.Join("recipes", "chk@1", "rec")} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("evidence: record %s written", p)
		}
	}
}

func TestRecipeKnownAnswers(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	bundleIn(t, dir, "tool", "## Standalone\n\nRun hi: ./hi.sh\n", map[string]string{"hi.sh": "#!/bin/sh\necho hi\n"})
	bundleIn(t, dir, "bare", "## Standalone\n\nRun hi: ./hi.sh\n", map[string]string{"hi.sh": "#!/bin/sh\necho hi\n"})
	bundleIn(t, dir, "plain", "Nothing to run.\n", nil)
	kase := func(name, stdout string) {
		c := filepath.Join(dir, "recipes", "tool@1", "known-answers", name)
		if err := os.MkdirAll(filepath.Join(c, "input"), 0o755); err != nil {
			t.Fatal(err)
		}
		body := `{"step": "standalone", "command": "hi", "exit": 0, "stdout": "` + stdout + `"}`
		if err := os.WriteFile(filepath.Join(c, "expect.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	kase("a", `hi\n`)
	kase("b", `hi\n`)
	r := hyp(t, "bundle", "known-answers", "tool@1")
	want(t, r, 0, "ok a\nok b\n")

	kase("c", `bye\n`)
	r = hyp(t, "bundle", "known-answers", "tool@1")
	want(t, r, 1, "FAIL c: stdout differs")
	want(t, r, 1, "refused: recipe tool@1: 1 of 3 known answers failed")

	r = hyp(t, "bundle", "known-answers", "bare@1")
	want(t, r, 1, "refused: recipe bare@1 is an instrument with no known answers")

	r = hyp(t, "bundle", "known-answers", "plain@1")
	want(t, r, 0, "bundle plain@1 has no Evidence or Standalone command; no known answers to run")

	r = hyp(t, "bundle", "known-answers")
	if r.code != 2 {
		t.Errorf("no REF: exit %d, want 2", r.code)
	}
}

func TestRecipeNewWritesFormat(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	bundleIn(t, dir, "x", "## Standalone\n\nRun ok: ./ok.sh\n", map[string]string{"ok.sh": "#!/bin/sh\n"})
	// The source has no format: 1 line.
	writeFile(t, filepath.Join(dir, "recipes", "x@1", "recipe.md"), "---\nname: x\nversion: 1\n---\n\n## Standalone\n\nRun ok: ./ok.sh\n")
	r := hyp(t, "bundle", "new", "x@2", "--from", "x@1")
	want(t, r, 0, "from: x@1 ")
	md, err := os.ReadFile(filepath.Join(dir, "recipes", "x@2", "recipe.md"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(md), "\n")
	if lines[1] != "name: x" || lines[2] != "version: 2" || !strings.HasPrefix(lines[3], "from: x@1 ") || lines[4] != "format: 1" {
		t.Fatalf("frontmatter = %q", lines[:5])
	}
	if _, err := recipe.Load(dir, "x@1"); err == nil || !strings.Contains(err.Error(), "recipe.md has no format: 1 line") {
		t.Fatalf("Load source: %v, want the missing format refused", err)
	}
	if _, err := recipe.Load(dir, "x@2"); err != nil {
		t.Fatalf("Load copy: %v", err)
	}
}

// A source that hashes but does not open is still copied, with a warning,
// so the copy can be fixed.
func TestRecipeNewWarnsWhenCopyDoesNotOpen(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	bundleIn(t, dir, "x", "Run: ./ok.sh\n", map[string]string{"ok.sh": "#!/bin/sh\n"})
	r := hyp(t, "bundle", "new", "x@2", "--from", "x@1")
	want(t, r, 0, "from: x@1 ")
	if !strings.Contains(r.stderr, "warn: the copy does not open (") || !strings.Contains(r.stderr, "fix it before using it") {
		t.Fatalf("stderr %q, want the does-not-open warning", r.stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, "recipes", "x@2", "recipe.md")); err != nil {
		t.Fatalf("the copy is missing: %v", err)
	}
}

// recipe check names each bundle frozen (a tree pins its hash) or draft.
func TestRecipeCheckFrozenAndDraft(t *testing.T) {
	repo := t.TempDir()
	gitIn(t, repo, "init", "-q")
	bundleIn(t, repo, "pinned", "", nil)
	bundleIn(t, repo, "loose", "", nil)
	t.Setenv("HYP_ACTOR", "a")
	t.Chdir(repo)
	must(t, "init")
	g := must(t, "add", "goal", "--title", "G", "--why", "start")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "pinned@1", "--decided-by", "reading", "--why", "w")
	r := hyp(t, "bundle", "check")
	want(t, r, 0, "draft: recipes/loose@1\nfrozen: recipes/pinned@1\nrecipe check: 1 trees, 1 pins, 0 drift\n")

	// An edit makes the pinned bundle drift, and no longer frozen.
	writeFile(t, filepath.Join(repo, "recipes", "pinned@1", "x.txt"), "x\n")
	want(t, hyp(t, "bundle", "check"), 1, "draft: recipes/pinned@1\n")
}

// hyp bundle check warns about each current instrument whose known answers
// are weak, before its summary line, and never fails for it. A case input
// that holds a store is a fixture, not a tree.
func TestRecipeCheckWarnsWeakCases(t *testing.T) {
	repo := t.TempDir()
	gitIn(t, repo, "init", "-q")
	pass := `{"step": "evidence", "command": "./ev.sh", "exit": 0, "stdout": ""}`
	fail := `{"step": "evidence", "command": "./ev.sh", "exit": 3, "stdout": ""}`
	bundleIn(t, repo, "env", "Set nothing up.\n", nil)
	bundleIn(t, repo, "none", "## Standalone\n\nRun a: ./a.sh\nExternal: none\n", map[string]string{"a.sh": "#!/bin/sh\n"})
	bundleIn(t, repo, "half", "## Evidence\n\nRun: ./ev.sh\nExternal: none\n\n## Standalone\n\nRun b: ./b.sh\nExternal: none\n", map[string]string{
		"ev.sh": "#!/bin/sh\n", "b.sh": "#!/bin/sh\n",
		"known-answers/p/expect.json": pass, "known-answers/p/input/a": "",
	})
	bundleIn(t, repo, "full", "## Evidence\n\nRun: ./ev.sh\nExternal: none\n", map[string]string{
		"ev.sh":                       "#!/bin/sh\n",
		"known-answers/p/expect.json": pass, "known-answers/p/input/a": "",
		"known-answers/f/expect.json": fail, "known-answers/f/input/.tree/events/a.jsonl": "",
	})
	writeLib(t, repo, "old@1", "---\nname: old\nversion: 1\n---\n\n## Standalone\n\nRun a: ./a.sh\n")
	t.Setenv("HYP_ACTOR", "a")
	t.Chdir(repo)
	must(t, "init")
	want(t, hyp(t, "bundle", "check"), 0, "draft: recipes/env@1\ndraft: recipes/full@1\ndraft: recipes/half@1\ndraft: recipes/none@1\ndraft: recipes/old@1\n"+
		"warn: recipes/half@1 ./ev.sh: no must-fail case\nwarn: recipes/half@1 b: no must-pass case\nwarn: recipes/half@1 b: no must-fail case\n"+
		"warn: recipes/none@1: instrument with no known answers\nrecipe check: 1 trees, 0 pins, 0 drift\n")
}

// A .go file in a bundle breaks go vet ./... in a repository whose root holds
// go.mod, so hyp bundle check and hyp add exp warn about it, exit 0; with no
// go.mod at the root they say nothing.
func TestGoFileInBundleWarns(t *testing.T) {
	repo, _, q := libraryStore(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	bundleIn(t, repo, "env", "Set nothing up.\n", map[string]string{"sub/probe_test.go": "package probe\n"})
	const warn = "warn: recipe env@1: sub/probe_test.go is a .go file, and go.mod at the repository root makes go vet ./... read this bundle as a package outside the main module; give it a non-.go name, such as probe_test.go.txt\n"
	addExp := func() result {
		t.Helper()
		return hyp(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "env@1", "--decided-by", "reading", "--why", "w")
	}
	if r := addExp(); r.code != 0 || r.stderr != "" {
		t.Fatalf("add exp with no go.mod: %+v", r)
	}
	if r := hyp(t, "bundle", "check"); r.code != 0 || strings.Contains(r.stdout, ".go file") {
		t.Fatalf("recipe check with no go.mod: %+v", r)
	}
	writeFile(t, filepath.Join(repo, "go.mod"), "module x\n")
	if r := addExp(); r.code != 0 || r.stderr != warn {
		t.Fatalf("add exp with go.mod: %+v", r)
	}
	r := hyp(t, "bundle", "check")
	if line := "warn: recipes/env@1/" + strings.TrimPrefix(warn, "warn: recipe env@1: "); r.code != 0 || !strings.Contains(r.stdout, line) {
		t.Fatalf("recipe check with go.mod: %+v, want %q", r, line)
	}
}

// An Experiment whose only Result is retracted is planned again, but its pin
// is checked once, through the Result, as before the Experiment record.
func TestRecipeCheckRetractedResultChecksPinOnce(t *testing.T) {
	repo, _, q := libraryStore(t)
	boot := filepath.Join(repo, "recipes", "boot@1")
	writeFile(t, filepath.Join(boot, "known.txt"), "k\n")
	old := bundleHash(t, boot)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	res := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w")
	must(t, "retract", res, "--why", "w")

	want(t, hyp(t, "bundle", "check"), 0, "recipe check: 1 trees, 1 pins, 0 drift")
	writeFile(t, filepath.Join(boot, "known.txt"), "edited\n")
	now := bundleHash(t, boot)
	r := hyp(t, "bundle", "check")
	if r.code != 1 {
		t.Fatalf("exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
	line := "drift: tree " + res + " recipes/boot@1/: recorded " + old + ", now " + now
	if !strings.Contains(r.stdout, line) || strings.Contains(r.stdout, "drift: tree "+e) || !strings.Contains(r.stdout, "1 pins, 1 drift") {
		t.Fatalf("stdout %q: want only %q", r.stdout, line)
	}
}

// A superseded or withdrawn Experiment still pins its recipe: recipe check
// counts the pin and names the bundle frozen.
func TestRecipeCheckCountsClosedExperimentPins(t *testing.T) {
	repo, _, q := libraryStore(t)
	boot := filepath.Join(repo, "recipes", "boot@1")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	old := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	repl := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	gone := must(t, "add", "exp", "--parent", h, "--title", "E3", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	must(t, "supersede", repl, "--old", old, "--why", "w")
	must(t, "withdraw", gone, "--why", "w")
	want(t, hyp(t, "bundle", "check"), 0, "frozen: recipes/boot@1\nrecipe check: 1 trees, 2 pins, 0 drift")
	writeFile(t, filepath.Join(boot, "extra.txt"), "x\n")
	r := hyp(t, "bundle", "check")
	if r.code != 1 || !strings.Contains(r.stdout, "drift: tree "+old+" ") || !strings.Contains(r.stdout, "drift: tree "+gone+" ") {
		t.Fatalf("exit %d, stdout %q: want drift on %s and %s", r.code, r.stdout, old, gone)
	}
}

// hyp bundle check warns about each Evidence and Standalone command whose
// External: line is missing or TODO, and never fails for it. hyp bundle new
// writes External: TODO under each command that lacks the line, and a frozen
// bundle without one still records.
func TestRecipeCheckWarnsExternal(t *testing.T) {
	repo, _, q := libraryStore(t)
	pass := `{"step": "evidence", "command": "./ev.sh", "exit": 0, "stdout": ""}`
	fail := `{"step": "evidence", "command": "./ev.sh", "exit": 3, "stdout": ""}`
	cases := map[string]string{"ev.sh": "#!/bin/sh\n", "b.sh": "#!/bin/sh\n",
		"known-answers/p/expect.json": pass, "known-answers/p/input/a": "",
		"known-answers/f/expect.json": fail, "known-answers/f/input/a": "",
		"known-answers/bp/expect.json": `{"step": "standalone", "command": "b", "exit": 0, "stdout": ""}`, "known-answers/bp/input/a": "",
		"known-answers/bf/expect.json": `{"step": "standalone", "command": "b", "exit": 1, "stdout": ""}`, "known-answers/bf/input/a": "",
	}
	bundleIn(t, repo, "said", "## Evidence\n\nRun: ./ev.sh\nExternal: reads OUT_DIR only\n\n## Standalone\n\nRun b: ./b.sh\n  External: none\n", cases)
	bundleIn(t, repo, "mute", "## Evidence\n\nRun: ./ev.sh\n\nExternal: too far down\n\n## Standalone\n\nRun b: ./b.sh\nExternal: TODO\n", cases)
	r := hyp(t, "bundle", "check")
	want(t, r, 0, "warn: recipes/mute@1 ./ev.sh: no External: line\nwarn: recipes/mute@1 b: External: not stated (TODO)\n")
	if strings.Contains(r.stdout, "said@1 ") {
		t.Fatalf("a bundle that states each command's reach is warned about:\n%s", r.stdout)
	}

	// A bundle a tree pins records without the line.
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "mute@1", "--rubric", "reading", "--why", "w")
	want(t, hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w"), 0, "result.record")

	want(t, hyp(t, "bundle", "new", "mute@2", "--from", "mute@1"), 0, "from: mute@1 ")
	md, err := os.ReadFile(filepath.Join(repo, "recipes", "mute@2", "recipe.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(md), "Run: ./ev.sh\nExternal: TODO\n\nExternal: too far down\n") || !strings.Contains(string(md), "Run b: ./b.sh\nExternal: TODO\n") ||
		strings.Count(string(md), "External: TODO") != 2 {
		t.Fatalf("recipe.md of the copy = %q", md)
	}
	want(t, hyp(t, "bundle", "check"), 0, "warn: recipes/mute@2 ./ev.sh: External: not stated (TODO)\nwarn: recipes/mute@2 b: External: not stated (TODO)\n")
}
