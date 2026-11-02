package recipe

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/herr"
)

// writeRecipe writes the single file recipes/<name>.md, which is not a recipe.
func writeRecipe(t *testing.T, root, name, body string) {
	t.Helper()
	dir := filepath.Join(root, "recipes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const bootV1 = "---\nname: boot\nversion: 1\nformat: 1\n---\n\nStart the app and load the home page.\n"

func TestResolveMatch(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@1", map[string]string{"recipe.md": bootV1})
	name, hash, err := Resolve(root, "boot@1")
	if err != nil || name != "boot" || len(hash) != 64 {
		t.Fatalf("Resolve = %q %q %v", name, hash, err)
	}
	if Hashes(root)["boot@1"] != hash {
		t.Fatal("Hashes disagrees with Resolve")
	}
}

// A single file recipes/<name>.md is not a recipe: a ref only it could answer
// is not found (exit 2), and beside a bundle of that ref it is ignored.
func TestSingleFileIsNotARecipe(t *testing.T) {
	root := t.TempDir()
	writeRecipe(t, root, "boot", bootV1)
	_, err := Load(root, "boot@1")
	if herr.KindOf(err) != herr.BadArguments || !strings.Contains(err.Error(), "recipe boot@1 not found; searched "+filepath.Join(root, "recipes")) {
		t.Fatalf("Load = %v, want not found", err)
	}
	if h, problems := Scan(root); len(h) != 0 || len(problems) != 0 {
		t.Fatalf("Scan = %v %v, want nothing", h, problems)
	}
	writeBundle(t, root, "boot@1", map[string]string{"recipe.md": bootV1 + "A bundle.\n"})
	r, err := Load(root, "boot@1")
	if err != nil || r.Path != filepath.Join(root, "recipes", "boot@1") {
		t.Fatalf("Load = %+v %v, want the bundle", r, err)
	}
	if h, problems := Scan(root); len(h) != 1 || h["boot@1"] != r.Hash || len(problems) != 0 {
		t.Fatalf("Scan = %v %v, want the bundle alone", h, problems)
	}
}

func TestLoadRefusesPathRefs(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "sub/b", map[string]string{"recipe.md": "---\nname: b\nversion: 1\nformat: 1\n---\n"})
	for _, ref := range []string{"sub/b@1", "../recipes/sub/b@1", "..@1"} {
		if _, err := Load(root, ref); err == nil || !strings.Contains(err.Error(), "invalid recipe ref") {
			t.Errorf("Load(%s) err = %v, want invalid recipe ref", ref, err)
		}
	}
}

func TestResolveMissingFile(t *testing.T) {
	if _, _, err := Resolve(t.TempDir(), "boot@1"); err == nil {
		t.Fatal("missing recipe resolved")
	}
	if len(Hashes(t.TempDir())) != 0 {
		t.Fatal("Hashes of a store without recipes is not empty")
	}
}

func TestHashChangesAfterEdit(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@1", map[string]string{"recipe.md": bootV1})
	before := Hashes(root)["boot@1"]
	writeBundle(t, root, "boot@1", map[string]string{"recipe.md": bootV1 + "Also check the console.\n"})
	if after := Hashes(root)["boot@1"]; after == before {
		t.Fatal("hash did not change after edit")
	}
}

// writeBundle writes files under root/recipes/<name>/, with mode 0o755 for the
// names in exec.
func writeBundle(t *testing.T, root, name string, files map[string]string, exec ...string) {
	t.Helper()
	dir := filepath.Join(root, "recipes", name)
	for rel, body := range files {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		mode := os.FileMode(0o644)
		if slices.Contains(exec, rel) {
			mode = 0o755
		}
		if err := os.WriteFile(p, []byte(body), mode); err != nil {
			t.Fatal(err)
		}
	}
}

const bootBundle = "---\nname: boot\nversion: 1\nformat: 1\n---\n\n## Launch\n\nRun: ./launch.sh a b\n\n## Doctor\n\nRun: ./bin/doctor.sh\n\n## Cleanup\n\nRun: ./stop.sh\n"

func bootFiles() map[string]string {
	return map[string]string{
		"recipe.md":     bootBundle,
		"launch.sh":     "#!/bin/sh\n",
		"bin/doctor.sh": "#!/bin/sh\n",
		"stop.sh":       "#!/bin/sh\n",
	}
}

var bootExec = []string{"launch.sh", "bin/doctor.sh", "stop.sh"}

func TestCaseInputDotEntriesAreHashed(t *testing.T) {
	root := t.TempDir()
	files := bootFiles()
	files["known-answers/a/input/.tree/x"] = "1\n"
	files["known-answers/a/.notes"] = "skipped\n"
	writeBundle(t, root, "boot@1", files, bootExec...)
	before := Hashes(root)["boot@1"]
	writeBundle(t, root, "boot@1", map[string]string{"known-answers/a/.notes": "still skipped\n"})
	if Hashes(root)["boot@1"] != before {
		t.Fatal("a dot-entry outside input/ changed the hash")
	}
	writeBundle(t, root, "boot@1", map[string]string{"known-answers/a/input/.tree/x": "2\n"})
	if after := Hashes(root)["boot@1"]; before == "" || after == before {
		t.Fatalf("editing known-answers/a/input/.tree/x left the hash %q", after)
	}
}

func TestCaseInputMayHoldAt(t *testing.T) {
	root := t.TempDir()
	files := bootFiles()
	files["known-answers/a/input/recipes/x@1/recipe.md"] = "1\n"
	writeBundle(t, root, "boot@1", files, bootExec...)
	before := Hashes(root)["boot@1"]
	writeBundle(t, root, "boot@1", map[string]string{"known-answers/a/input/recipes/x@1/recipe.md": "2\n"})
	if after := Hashes(root)["boot@1"]; before == "" || after == before {
		t.Fatalf("editing known-answers/a/input/recipes/x@1/recipe.md left the hash %q", after)
	}
	for _, rel := range []string{"notes@1.txt", "known-answers/a@1/input/x", "known-answers/a/expect/x@1"} {
		t.Run(rel, func(t *testing.T) {
			root := t.TempDir()
			files := bootFiles()
			files[rel] = "x\n"
			writeBundle(t, root, "boot@1", files, bootExec...)
			_, err := bundleHash(filepath.Join(root, "recipes", "boot@1"))
			if err == nil || !strings.Contains(err.Error(), "outside [A-Za-z0-9._-]; only a path under known-answers/NAME/input/ may also hold @") {
				t.Fatalf("bundleHash err = %v", err)
			}
		})
	}
}

func TestBundleResolves(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@1", bootFiles(), bootExec...)
	r, err := Load(root, "boot@1")
	if err != nil {
		t.Fatal(err)
	}
	if r.Dir == "" || len(r.Hash) != 64 || r.Hash != Hashes(root)["boot@1"] {
		t.Fatalf("Load = %+v, Hashes = %v", r, Hashes(root))
	}
	if len(r.Launch) != 1 || len(r.Doctor) != 1 || len(r.Cleanup) != 1 {
		t.Fatalf("commands = %v %v %v", r.Launch, r.Doctor, r.Cleanup)
	}
	if !reflect.DeepEqual(r.Launch[0][1:], []string{"a", "b"}) {
		t.Fatalf("Launch = %v", r.Launch)
	}
	for _, argv := range slices.Concat(r.Launch, r.Doctor, r.Cleanup) {
		if !filepath.IsAbs(argv[0]) || !strings.HasPrefix(argv[0], r.Dir+string(filepath.Separator)) {
			t.Fatalf("argv[0] %s is not absolute inside %s", argv[0], r.Dir)
		}
	}
}

func TestBundleHashCoversEveryFile(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot", bootFiles(), bootExec...)
	hash := func() string {
		t.Helper()
		h, err := bundleHash(filepath.Join(root, "recipes", "boot"))
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	prev := hash()
	for _, edit := range []map[string]string{
		{"recipe.md": bootBundle + "More prose.\n"},
		{"bin/doctor.sh": "#!/bin/sh\nexit 0\n"},
		{"notes.txt": "new\n"},
	} {
		writeBundle(t, root, "boot", edit)
		h := hash()
		if h == prev {
			t.Fatalf("hash unchanged after %v", edit)
		}
		prev = h
	}
	for _, ignored := range []map[string]string{
		{".DS_Store": "x"},
		{filepath.Join(".git", "x"): "x"},
		{"__pycache__/m.pyc": "x"},
	} {
		writeBundle(t, root, "boot", ignored)
		if h := hash(); h != prev {
			t.Fatalf("hash changed after %v", ignored)
		}
	}
}

func TestBundleHashMatchesManifest(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "two", map[string]string{"a.sh": "A\n", "sub/b.txt": "B\n"})
	manifest := sum([]byte("A\n")) + "  a.sh\n" + sum([]byte("B\n")) + "  sub/b.txt\n"
	got, err := bundleHash(filepath.Join(root, "recipes", "two"))
	if err != nil || got != sum([]byte(manifest)) {
		t.Fatalf("bundleHash = %q %v, want %s", got, err, sum([]byte(manifest)))
	}
}

func TestBundleRefusals(t *testing.T) {
	recipeWith := func(body string) string { return "---\nname: boot\nversion: 1\nformat: 1\n---\n\n" + body }
	outside := func(t *testing.T, root string) {
		writeBundle(t, root, "", map[string]string{"outside.sh": "#!/bin/sh\n"}, "outside.sh")
	}
	cases := []struct {
		name  string
		setup func(t *testing.T, root string)
		want  string // "" means Load succeeds with no Launch commands
	}{
		{"outside", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ../outside.sh\n")})
			outside(t, root)
		}, "line 8: Run: ../outside.sh: ../outside.sh resolves outside the bundle (move the script into the bundle)"},
		{"absolute", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: /bin/echo hi\n")})
		}, "Run: /bin/echo hi: a command is a relative path inside the bundle, not an absolute path"},
		{"no slash", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: launch.sh\n"), "launch.sh": "#!/bin/sh\n"}, "launch.sh")
		}, "Run: launch.sh: a command is a relative path containing / (write ./launch.sh)"},
		{"missing", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ./lauch.sh\n"), "launch.sh": "#!/bin/sh\n"}, "launch.sh")
		}, "Run: ./lauch.sh: ./lauch.sh does not exist in the bundle"},
		{"not executable", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ./data.txt\n"), "data.txt": "x\n"})
		}, "Run: ./data.txt: ./data.txt is not executable (chmod +x data.txt)"},
		{"directory", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ./bin\n"), "bin/a.sh": "#!/bin/sh\n"}, "bin/a.sh")
		}, "Run: ./bin: ./bin is not a regular file"},
		{"bundle root", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ./\n")})
		}, "Run: ./: ./ is not a regular file"},
		{"symlink", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ./link.sh\n")})
			outside(t, root)
			if err := os.Symlink("../outside.sh", filepath.Join(root, "recipes", "boot@1", "link.sh")); err != nil {
				t.Fatal(err)
			}
		}, "not a regular file"},
		{"symlinked bundle", func(t *testing.T, root string) {
			writeBundle(t, root, "real", map[string]string{"recipe.md": bootV1})
			if err := os.Symlink("real", filepath.Join(root, "recipes", "boot@1")); err != nil {
				t.Fatal(err)
			}
		}, "is a symlink"},
		{"unhashed script", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ./.hid/x.sh\n"), ".hid/x.sh": "#!/bin/sh\n"}, ".hid/x.sh")
		}, "Run: ./.hid/x.sh: ./.hid/x.sh is under a dot or __pycache__ path the bundle hash skips"},
		{"pycache script", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ./__pycache__/x\n"), "__pycache__/x": "#!/bin/sh\n"}, "__pycache__/x")
		}, "./__pycache__/x is under a dot or __pycache__ path the bundle hash skips"},
		{"symlink in dot dir", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": bootV1, ".hid/keep": "x"})
			if err := os.Symlink("/etc/passwd", filepath.Join(root, "recipes", "boot@1", ".hid", "ln")); err != nil {
				t.Fatal(err)
			}
		}, "not a regular file"},
		{"variable", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ./a.sh $HOME\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, "not a plain command"},
		{"pipe", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\nRun: ./a.sh | tee x\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, "not a plain command"},
		{"other section", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Known answers\nRun: ./a.sh\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, "allowed only under Launch, Doctor, Drive, Evidence, Cleanup and Standalone"},
		{"plain standalone", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Standalone\nRun: ./a.sh\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, "a Standalone command is named"},
		{"labelled drive", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Drive\nRun go: ./a.sh\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, "is allowed only under Standalone"},
		{"labelled no section", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("Run go: ./a.sh\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, "is allowed only under Standalone"},
		{"bad label", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Standalone\nRun Go: ./a.sh\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, "must match [a-z][a-z0-9-]*"},
		{"label twice", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Standalone\nRun go: ./a.sh\nRun go: ./a.sh x\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, "is already used"},
		{"labelled shell", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Standalone\nRun go: ./a.sh $X\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, "is not a plain command"},
		{"fenced", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": recipeWith("## Launch\n```\nRun: ./a.sh\n```\n"), "a.sh": "#!/bin/sh\n"}, "a.sh")
		}, ""},
		{"bad name", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": bootV1, "a b.sh": "x\n"})
		}, "outside [A-Za-z0-9._-]"},
		{"name mismatch", func(t *testing.T, root string) {
			writeBundle(t, root, "boot@1", map[string]string{"recipe.md": "---\nname: other\nversion: 1\nformat: 1\n---\n"})
		}, `recipe boot@1: recipe.md has name "other"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			c.setup(t, root)
			r, err := Load(root, "boot@1")
			if c.want == "" {
				if err != nil || len(r.Launch) != 0 {
					t.Fatalf("Load = %+v %v, want success with no Launch", r, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
		})
	}
}

func TestOpenReadsEveryStep(t *testing.T) {
	root := t.TempDir()
	md := "---\nname: boot\nversion: 1\nformat: 1\n---\n\n## Launch\nRun: ./l.sh\n\n## Doctor\nRun: ./d.sh\n\n## Drive\nRun: ./v.sh\n\n" +
		"## Evidence\nRun: ./e.sh\n\n## Cleanup\nRun: ./c.sh\n\n## Standalone\nRun snap: ./s.sh one\nRun re-test: ./r.sh\n"
	files := map[string]string{"recipe.md": md}
	scripts := []string{"l.sh", "d.sh", "v.sh", "e.sh", "c.sh", "s.sh", "r.sh"}
	for _, f := range scripts {
		files[f] = "#!/bin/sh\n"
	}
	writeBundle(t, root, "boot@1", files, scripts...)
	r, err := Load(root, "boot@1")
	if err != nil {
		t.Fatal(err)
	}
	for i, step := range [][][]string{r.Launch, r.Doctor, r.Drive, r.Evidence, r.Cleanup} {
		if len(step) != 1 || !strings.HasSuffix(step[0][0], "/"+scripts[i]) {
			t.Fatalf("step %d = %v, want one %s", i, step, scripts[i])
		}
	}
	if len(r.Standalone) != 2 {
		t.Fatalf("Standalone = %v, want 2", r.Standalone)
	}
	if n := r.Standalone[0]; n.Name != "snap" || !strings.HasSuffix(n.Argv[0], "/s.sh") || !reflect.DeepEqual(n.Argv[1:], []string{"one"}) {
		t.Fatalf("Standalone[0] = %v", n)
	}
	if n := r.Standalone[1]; n.Name != "re-test" || !strings.HasSuffix(n.Argv[0], "/r.sh") || len(n.Argv) != 1 {
		t.Fatalf("Standalone[1] = %v", n)
	}
}

// A directory recipes/<name>/ without a version is not a recipe: Load does not
// find it and Scan neither hashes nor reports it.
func TestUnversionedBundleIsNotARecipe(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot", bootFiles(), bootExec...)
	if _, err := Load(root, "boot@1"); err == nil || !strings.Contains(err.Error(), "recipe boot@1 not found; searched") {
		t.Fatalf("err = %v, want not found", err)
	}
	if h, problems := Scan(root); len(h) != 0 || len(problems) != 0 {
		t.Fatalf("Scan = %v %v, want nothing", h, problems)
	}
}

// Every bundle in the repository is recipes/<name>@<version>/ and opens.
// .git and .tree are skipped: event logs and evidence are records. So is a
// nested repository or worktree (a directory below the root holding .git),
// which may be a checkout of an older commit.
func TestRepositoryBundlesStillOpen(t *testing.T) {
	var found []string
	err := filepath.WalkDir("../..", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == ".tree") {
			return fs.SkipDir
		}
		if d.IsDir() && p != "../.." {
			if _, err := os.Lstat(filepath.Join(p, ".git")); err == nil {
				return fs.SkipDir
			}
		}
		if d.Name() != "recipe.md" || filepath.Base(filepath.Dir(filepath.Dir(p))) != "recipes" {
			return nil
		}
		dir := filepath.Dir(p)
		found = append(found, dir)
		if !strings.Contains(filepath.Base(dir), "@") {
			t.Errorf("%s: a bundle directory needs @<version>", dir)
			return nil
		}
		if _, err := Open(dir); err != nil {
			t.Errorf("Open(%s): %v", dir, err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// The two test-data bundles: the synthetic tree's count-check@1 and the
	// sandbox tests' scratch-pg@1.
	if len(found) < 2 {
		t.Fatalf("found %d bundles, want at least 2", len(found))
	}
}

// count-check@1 is frozen: the synthetic tree pins this hash, so Open must
// still give it.
func TestSynthBundleKeepsItsPin(t *testing.T) {
	r, err := Open("../testutil/testdata/synth/recipes/count-check@1")
	if err != nil {
		t.Fatal(err)
	}
	if want := "3a281aefcaba87f666b249aede7b3865161efec0011320de26ccee21106186a3"; r.Hash != want {
		t.Fatalf("hash = %s, want %s", r.Hash, want)
	}
}

func TestOpenCopiedBundle(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@1", bootFiles(), bootExec...)
	r, err := Load(root, "boot@1")
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "boot")
	if err := os.CopyFS(dst, os.DirFS(r.Dir)); err != nil {
		t.Fatal(err)
	}
	c, err := OpenAs(dst, "boot@1")
	if err != nil || c.Hash != r.Hash {
		t.Fatalf("OpenAs(copy) = %+v %v, want hash %s", c, err, r.Hash)
	}
}

func TestOpenRefusesFolderWithoutVersion(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@1", bootFiles(), bootExec...)
	dst := filepath.Join(t.TempDir(), "boot")
	if err := os.CopyFS(dst, os.DirFS(filepath.Join(root, "recipes", "boot@1"))); err != nil {
		t.Fatal(err)
	}
	_, err := Open(dst)
	if herr.KindOf(err) != herr.Refusal || !strings.Contains(err.Error(), "the directory name has no @version") {
		t.Fatalf("Open(boot/) = %v, want a refusal", err)
	}
}

func TestOpenAsChecksTheRef(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@1", bootFiles(), bootExec...)
	dst := filepath.Join(t.TempDir(), "anything")
	if err := os.CopyFS(dst, os.DirFS(filepath.Join(root, "recipes", "boot@1"))); err != nil {
		t.Fatal(err)
	}
	r, err := OpenAs(dst, "boot@1")
	if err != nil || r.Name != "boot" || r.Version != "1" {
		t.Fatalf("OpenAs(anything, boot@1) = %+v %v", r, err)
	}
	for ref, want := range map[string]string{
		"other@1": `recipe other@1: recipe.md has name "boot"`,
		"boot@2":  "recipe boot@2: recipe.md has version 1",
	} {
		if _, err := OpenAs(dst, ref); herr.KindOf(err) != herr.Refusal || !strings.Contains(err.Error(), want) {
			t.Fatalf("OpenAs(anything, %s) = %v, want %q", ref, err, want)
		}
	}
}

func mkGit(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// repoWithStore makes repo/.git and an empty store directory repo/a/b.
func repoWithStore(t *testing.T) (repo, store string) {
	t.Helper()
	repo = t.TempDir()
	mkGit(t, repo)
	store = filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(store, 0o755); err != nil {
		t.Fatal(err)
	}
	return repo, store
}

func bootFilesV(version string) map[string]string {
	f := bootFiles()
	f["recipe.md"] = strings.Replace(bootBundle, "version: 1", "version: "+version, 1)
	return f
}

func TestDirsStopsAtRepoRoot(t *testing.T) {
	repo, store := repoWithStore(t)
	want := []string{filepath.Join(store, "recipes"), filepath.Join(repo, "a", "recipes"), filepath.Join(repo, "recipes")}
	if got := Dirs(store); !reflect.DeepEqual(got, want) {
		t.Fatalf("Dirs = %v, want %v", got, want)
	}
}

func TestDirsWithoutGitIsOnlyBesideStore(t *testing.T) {
	parent := t.TempDir()
	store := filepath.Join(parent, "store")
	if err := os.MkdirAll(store, 0o755); err != nil {
		t.Fatal(err)
	}
	writeBundle(t, parent, "boot@1", bootFiles(), bootExec...)
	if got := Dirs(store); !reflect.DeepEqual(got, []string{filepath.Join(store, "recipes")}) {
		t.Fatalf("Dirs = %v", got)
	}
	_, err := Load(store, "boot@1")
	if err == nil || !strings.Contains(err.Error(), "not found; searched "+filepath.Join(store, "recipes")) ||
		strings.Contains(err.Error(), filepath.Join(parent, "recipes")+string(filepath.Separator)) {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadFromRepoRootLibrary(t *testing.T) {
	repo, store := repoWithStore(t)
	writeBundle(t, repo, "boot@1", bootFiles(), bootExec...)
	r, err := Load(store, "boot@1")
	if err != nil || r.Path != filepath.Join(repo, "recipes", "boot@1") || r.Name != "boot" || r.Version != "1" {
		t.Fatalf("Load = %+v %v", r, err)
	}
	if Hashes(store)["boot@1"] != r.Hash {
		t.Fatalf("Hashes = %v, want boot@1 %s", Hashes(store), r.Hash)
	}
}

func TestVersionedBundlesSideBySide(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@1", bootFiles(), bootExec...)
	writeBundle(t, root, "boot@2", bootFilesV("2"), bootExec...)
	r1, err1 := Load(root, "boot@1")
	r2, err2 := Load(root, "boot@2")
	if err1 != nil || err2 != nil || r1.Hash == r2.Hash {
		t.Fatalf("Load = %+v %v, %+v %v", r1, err1, r2, err2)
	}
	h, problems := Scan(root)
	if !reflect.DeepEqual(h, map[string]string{"boot@1": r1.Hash, "boot@2": r2.Hash}) || len(problems) != 0 {
		t.Fatalf("Scan = %v %v", h, problems)
	}
}

func TestOpenRefusesWrongDirectoryVersion(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@2", bootFiles(), bootExec...) // recipe.md says version 1
	if _, err := Load(root, "boot@2"); err == nil || !strings.Contains(err.Error(), "recipe boot@2: recipe.md has version 1") {
		t.Fatalf("err = %v", err)
	}
	h, problems := Scan(root)
	if len(h) != 0 || len(problems) != 1 || !strings.Contains(problems[0], "does not open") {
		t.Fatalf("Scan = %v %v", h, problems)
	}
}

func TestLoadSameHashTwiceReturnsNearest(t *testing.T) {
	repo, store := repoWithStore(t)
	writeBundle(t, repo, "boot@1", bootFiles(), bootExec...)
	writeBundle(t, store, "boot@1", bootFiles(), bootExec...)
	r, err := Load(store, "boot@1")
	if err != nil || r.Path != filepath.Join(store, "recipes", "boot@1") {
		t.Fatalf("Load = %+v %v", r, err)
	}
	if _, problems := Scan(store); len(problems) != 0 {
		t.Fatalf("problems = %v", problems)
	}
}

func TestLoadRefusesDifferingCopies(t *testing.T) {
	repo, store := repoWithStore(t)
	near, far := filepath.Join(store, "recipes", "boot@1"), filepath.Join(repo, "recipes", "boot@1")
	writeBundle(t, repo, "boot@1", bootFiles(), bootExec...)
	changed := bootFiles()
	changed["note.txt"] = "x\n"
	writeBundle(t, store, "boot@1", changed, bootExec...)
	_, err := Load(store, "boot@1")
	if err == nil || !strings.Contains(err.Error(), near+" and "+far+" differ") {
		t.Fatalf("err = %v", err)
	}
	h, problems := Scan(store)
	if _, ok := h["boot@1"]; ok || len(problems) != 1 || !strings.Contains(problems[0], near+" and "+far+" differ") {
		t.Fatalf("Scan = %v %v", h, problems)
	}
}

func TestMissingRefNamesEveryDirectory(t *testing.T) {
	repo, store := repoWithStore(t)
	writeRecipe(t, repo, "boot", bootV1)
	_, err := Load(store, "boot@2")
	want := "recipe boot@2 not found; searched " + strings.Join(Dirs(store), ", ")
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %s", err, want)
	}
}

func TestBrokenOlderVersionDoesNotBlockNewer(t *testing.T) {
	repo, store := repoWithStore(t)
	writeBundle(t, store, "boot@6", bootFilesV("6")) // no execute bits: does not open
	writeBundle(t, repo, "boot@7", bootFilesV("7"), bootExec...)
	if _, err := Load(store, "boot@6"); err == nil || !strings.Contains(err.Error(), "is not executable") {
		t.Fatalf("Load(boot@6) err = %v, want not executable", err)
	}
	if r, err := Load(store, "boot@7"); err != nil || r.Version != "7" {
		t.Fatalf("Load = %+v %v", r, err)
	}
}

func TestScanReportsBundleThatDoesNotOpen(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@1", bootFiles()) // no execute bits
	h, problems := Scan(root)
	dir := filepath.Join(root, "recipes", "boot@1")
	// Hashing never parses, so the bundle still hashes and its pins verify.
	if h["boot@1"] == "" || len(problems) != 1 ||
		!strings.Contains(problems[0], "recipe "+dir+" does not open: ") ||
		!strings.Contains(problems[0], "is not executable (chmod +x") ||
		!strings.Contains(problems[0], "its pins still verify") {
		t.Fatalf("Scan = %v %v", h, problems)
	}
}

// Every pin is written with Load and checked against Scan, so the two must
// give one hash per ref in every layout, beside the store and in a parent
// library.
func TestLoadAndScanAgreeOnEveryRef(t *testing.T) {
	for _, tc := range []struct {
		name  string
		write func(t *testing.T, dir string)
		ref   string
	}{
		{"versioned bundle", func(t *testing.T, dir string) { writeBundle(t, dir, "boot@2", bootFilesV("2"), bootExec...) }, "boot@2"},
	} {
		for _, level := range []string{"beside the store", "parent library"} {
			t.Run(tc.name+" "+level, func(t *testing.T) {
				repo, store := repoWithStore(t)
				dir := store
				if level == "parent library" {
					dir = repo
				}
				tc.write(t, dir)
				r, err := Load(store, tc.ref)
				if err != nil {
					t.Fatal(err)
				}
				h, problems := Scan(store)
				if len(problems) != 0 || len(h) != 1 || h[tc.ref] != r.Hash {
					t.Fatalf("Scan = %v %v, Load(%s).Hash = %s", h, problems, tc.ref, r.Hash)
				}
			})
		}
	}
}

func TestRepoRoot(t *testing.T) {
	repo := t.TempDir()
	mkGit(t, repo)
	tree := filepath.Join(repo, "tree")
	if err := os.MkdirAll(tree, 0o755); err != nil {
		t.Fatal(err)
	}

	real := func(p string) string {
		t.Helper()
		r, err := filepath.EvalSymlinks(p)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	if got, ok := RepoRoot(tree); !ok || real(got) != real(repo) {
		t.Fatalf("RepoRoot(tree) = %q, %v; want %q", got, ok, repo)
	}
	if got, ok := RepoRoot(t.TempDir()); ok || got != "" {
		t.Fatalf("RepoRoot outside a repository = %q, %v", got, ok)
	}
}

// A script no command names is hashed content: the bundle opens, and editing
// the script changes the hash.
func TestUnnamedScriptIsHashedContent(t *testing.T) {
	root := t.TempDir()
	files := bootFiles()
	files["helper.py"] = "print(1)\n"
	writeBundle(t, root, "boot@1", files, bootExec...)
	dir := filepath.Join(root, "recipes", "boot@1")
	r, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	writeBundle(t, root, "boot@1", map[string]string{"helper.py": "print(2)\n"})
	r2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if r.Hash == r2.Hash {
		t.Fatal("editing helper.py left the hash unchanged")
	}
}

// A bundle whose recipe.md has no format: 1 line is refused like any other
// missing required field, before its commands are parsed. It still hashes,
// so its pins still verify.
func TestLoadRefusesMissingFormat(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, "boot@1", map[string]string{"recipe.md": "---\nname: boot\nversion: 1\n---\n\nRun: ./nope.sh\n"})
	_, err := Load(root, "boot@1")
	if herr.KindOf(err) != herr.Refusal || !strings.Contains(err.Error(), "recipe boot@1: recipe.md has no format: 1 line") {
		t.Fatalf("Load = %v, want the missing format line refused", err)
	}
	if h := Hashes(root); h["boot@1"] == "" {
		t.Fatalf("Hashes = %v, want boot@1 hashed", h)
	}
}
