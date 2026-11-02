package recipe

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/herr"
)

// No directory of the embedded template is empty: git keeps no empty
// directory, and a known answer's input/ must exist.
func TestTemplateHasNoEmptyDirectory(t *testing.T) {
	err := fs.WalkDir(template, "template", func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		files := 0
		fs.WalkDir(template, p, func(_ string, e fs.DirEntry, _ error) error {
			if e != nil && e.Type().IsRegular() {
				files++
			}
			return nil
		})
		if files == 0 {
			t.Errorf("%s holds no file", p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// firstRepo makes a repository with a store directory in it and returns both.
func firstRepo(t *testing.T) (repo, root string) {
	t.Helper()
	repo = t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return repo, filepath.Join(repo, "tree")
}

func TestNewFirstWritesTheTemplate(t *testing.T) {
	_, root := firstRepo(t)
	dir, err := NewFirst(root, "g@1")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "recipes", "g@1"); dir != want {
		t.Fatalf("dir %s, want %s", dir, want)
	}
	modes := map[string]os.FileMode{}
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			rel, _ := filepath.Rel(dir, p)
			info, _ := d.Info()
			modes[filepath.ToSlash(rel)] = info.Mode().Perm()
		}
		return nil
	})
	want := map[string]os.FileMode{
		"recipe.md":                                0o644,
		"check.sh":                                 0o755,
		"known-answers/must-pass/expect.json":      0o644,
		"known-answers/must-pass/input/README.txt": 0o644,
		"known-answers/must-fail/expect.json":      0o644,
		"known-answers/must-fail/input/README.txt": 0o644,
	}
	if !reflect.DeepEqual(modes, want) {
		t.Fatalf("files %v, want %v", modes, want)
	}
	r, err := Open(dir)
	if err != nil {
		t.Fatalf("the first bundle does not open: %v", err)
	}
	if r.Name != "g" || r.Version != "1" || r.From != "" || len(r.Standalone) != 1 || r.Standalone[0].Name != "check" {
		t.Fatalf("opened %+v", r)
	}
	if ext := r.Standalone[0].External; !ext.Line || ext.Text != ExternalTODO {
		t.Fatalf("check External %+v, want TODO", ext)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "recipe.md"))
	for _, s := range []string{"name: g\n", "format: " + Format + "\n", "# g\n",
		"## Launch\n", "## Doctor\n", "## Drive\n", "## Evidence\n", "## Standalone\n", "## Cleanup\n"} {
		if !strings.Contains(string(data), s) {
			t.Errorf("recipe.md lacks %q", s)
		}
	}
	if strings.Contains(string(data), nameToken) || strings.Contains(string(data), "from:") {
		t.Errorf("recipe.md still holds %s or a from: line", nameToken)
	}
	got, err := HasPlaceholder(dir)
	if err != nil {
		t.Fatal(err)
	}
	if w := []string{"check.sh", "recipe.md"}; !reflect.DeepEqual(got, w) {
		t.Fatalf("placeholder in %v, want %v", got, w)
	}
	if left, _ := filepath.Glob(filepath.Join(root, "recipes", ".*")); len(left) != 0 {
		t.Fatalf("temporary directories left: %v", left)
	}
}

func TestNewFirstRefusals(t *testing.T) {
	repo, root := firstRepo(t)
	for _, ref := range []string{"g@2", "g", "g@x/y", "../g@1"} {
		if _, err := NewFirst(root, ref); herr.KindOf(err) != herr.BadArguments {
			t.Errorf("%s: %v; want a bad argument", ref, err)
		}
	}
	if _, err := NewFirst(root, "g@2"); !strings.Contains(herr.Text(err), "missing --from") {
		t.Errorf("g@2: %v", err)
	}
	// A version of the name in a parent library is refused, since a next
	// version must start from it.
	if err := os.MkdirAll(filepath.Join(repo, "recipes", "g@3"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := NewFirst(root, "g@1")
	if herr.KindOf(err) != herr.Refusal || !strings.Contains(herr.Text(err), "recipe g exists at "+filepath.Join(repo, "recipes", "g@3")) || !strings.Contains(herr.Text(err), "--from g@3") {
		t.Fatalf("existing name: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "recipes")); !os.IsNotExist(err) {
		t.Fatalf("a refusal wrote recipes/ (%v)", err)
	}
	// A first version in the store's own library is refused the same way.
	if _, err := NewFirst(root, "h@1"); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFirst(root, "h@1"); herr.KindOf(err) != herr.Refusal {
		t.Fatalf("h@1 twice: %v", err)
	}
}

// The inputs of known answers may hold the marker, and the files the hash
// skips are not read either.
func TestHasPlaceholderSkipsInputs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "x@1")
	for rel, body := range map[string]string{
		"recipe.md":                       "---\nname: x\n---\n",
		"known-answers/a/input/store.txt": Placeholder,
		"known-answers/a/expect.json":     "{}",
		".notes":                          Placeholder,
		"lib/helper.py":                   "# " + Placeholder,
	} {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := HasPlaceholder(dir)
	if err != nil {
		t.Fatal(err)
	}
	if w := []string{"lib/helper.py"}; !reflect.DeepEqual(got, w) {
		t.Fatalf("got %v, want %v", got, w)
	}
}
