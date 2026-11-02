package sandbox

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/store"
)

func readChanges(t *testing.T, out string) Changes {
	t.Helper()
	var ch Changes
	data, err := os.ReadFile(filepath.Join(out, "changes.json"))
	if err != nil || json.Unmarshal(data, &ch) != nil {
		t.Fatalf("changes.json = %s (%v)", data, err)
	}
	return ch
}

// changeHost is a host repository holding a.txt, b.txt and c.txt.
func changeHost(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "init", "-q", "-b", "main")
	for _, f := range []string{"a.txt", "b.txt", "c.txt"} {
		writeFile(t, filepath.Join(dir, f), f+"\n")
	}
	run(t, dir, "add", ".")
	run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "init")
	return dir, run(t, dir, "rev-parse", "HEAD")
}

func TestRemoveWritesChanges(t *testing.T) {
	path := pathDir(t, "git", "lsof", "rm")
	setup(t)
	t.Setenv("PATH", path)
	h, base := changeHost(t)
	files := map[string]string{
		"recipe.md": recipeMD("boot", "## Drive\n\nRun: ./d.sh\n"),
		"d.sh":      "#!/bin/sh\ncd \"$HYP_SANDBOX_WT\"\necho new > new.txt\necho edited > b.txt\nrm c.txt\n",
	}
	root := recipeRoot(t, "boot", files, "d.sh")
	if _, _, err := createWith(t, "chg", CreateOptions{Host: h, Base: base, Refs: []string{"boot@1"}, Recipes: root}); err != nil {
		t.Fatal(err)
	}
	if _, err := Drive("chg", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "chg", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	ch := readChanges(t, out)
	if ch.Format != ChangesFormat || ch.Baseline != "manifest.json at create" {
		t.Errorf("changes = %+v", ch)
	}
	var got []string
	for _, c := range ch.Changes {
		copied := "-"
		if c.Copied != nil {
			copied = map[bool]string{true: "copied", false: "not copied"}[*c.Copied]
		}
		got = append(got, c.Path+" "+c.Change+" "+copied)
	}
	if want := "b.txt modified copied, c.txt deleted -, new.txt added copied"; strings.Join(got, ", ") != want {
		t.Errorf("changes = %q, want %q", strings.Join(got, ", "), want)
	}
	for f, want := range map[string]string{"new.txt": "new\n", "b.txt": "edited\n"} {
		if data, _ := os.ReadFile(filepath.Join(out, "changes", f)); string(data) != want {
			t.Errorf("changes/%s = %q, want %q", f, data, want)
		}
	}
	if exists(filepath.Join(out, "changes", "c.txt")) {
		t.Error("a deleted file was copied")
	}
}

func TestChangesCopyCap(t *testing.T) {
	old := copyCap
	copyCap = 4
	t.Cleanup(func() { copyCap = old })
	setup(t)
	driveBox(t, "cap", "Run: ./d.sh", map[string]string{
		"d.sh": "#!/bin/sh\nprintf 0123456789 > \"$HYP_SANDBOX_WT/big.txt\"\n",
	}, "d.sh")
	if _, err := Drive("cap", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "cap", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	ch := readChanges(t, out)
	if len(ch.Changes) != 1 || ch.Changes[0].Path != "big.txt" || ch.Changes[0].Size != 10 || ch.Changes[0].Copied == nil || *ch.Changes[0].Copied {
		t.Fatalf("changes = %+v", ch.Changes)
	}
	if exists(filepath.Join(out, "changes", "big.txt")) {
		t.Error("a file over the cap was copied")
	}
}

func TestRemoveWithoutManifestWarns(t *testing.T) {
	setup(t)
	_, box := create(t, "nom")
	if err := os.Remove(filepath.Join(box, "manifest.json")); err != nil {
		t.Fatal(err)
	}
	var warn bytes.Buffer
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "nom", OutDir: out, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(warn.String(), "sandbox: no manifest (created by an older hyp); changes.json not written") {
		t.Errorf("warn = %q", warn.String())
	}
	if exists(filepath.Join(out, "changes.json")) {
		t.Error("changes.json written without a manifest")
	}
}

func TestStoreCutIsBaseline(t *testing.T) {
	setup(t)
	src := t.TempDir()
	if err := store.Init(src); err != nil {
		t.Fatal(err)
	}
	ev, err := store.Append(src, "a", event.Event{Type: event.TypeNodeAdd, Payload: []byte(`{"kind":"Goal","parent":"","title":"g"}`), Why: "w"})
	if err != nil {
		t.Fatal(err)
	}
	rec, _, err := createWith(t, "cutb", CreateOptions{Refs: []string{"none"}, Store: src, Cut: ev.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(rec.WT, ".tree", "events")) {
		t.Fatal("no cut in wt/.tree")
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "cutb", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	if ch := readChanges(t, out); len(ch.Changes) != 0 {
		t.Errorf("changes = %+v, want none", ch.Changes)
	}
}

func TestCreateInputs(t *testing.T) {
	path := pathDir(t, "git", "lsof", "cat")
	setup(t)
	t.Setenv("PATH", path)
	if _, _, err := createWith(t, "in0", CreateOptions{Refs: []string{"none"}, Inputs: filepath.Join(t.TempDir(), "missing.json")}); code(err) != 2 || !strings.Contains(err.Error(), "is not a regular file") {
		t.Fatalf("missing inputs: err = %v", err)
	}
	in := filepath.Join(t.TempDir(), "params.json")
	writeFile(t, in, `{"n": 1}`+"\n")
	root := recipeRoot(t, "use", map[string]string{
		"recipe.md": bundle("use", "Run: ./launch.sh", "", ""),
		"launch.sh": "#!/bin/sh\ncat \"$HYP_SANDBOX_INPUTS\"\n",
	}, "launch.sh")
	rec, box, err := createWith(t, "in1", CreateOptions{Refs: []string{"use@1"}, Recipes: root, Inputs: in})
	if err != nil {
		t.Fatal(err)
	}
	copyPath := filepath.Join(box, "inputs", "params.json")
	sum, _ := fileSum(copyPath)
	if rec.Inputs == nil || rec.Inputs.Path != copyPath || rec.Inputs.SHA256 != sum || sum == "" {
		t.Fatalf("inputs = %+v, want the copy %s with sha256 %s", rec.Inputs, copyPath, sum)
	}
	if log, _ := os.ReadFile(filepath.Join(box, "recipe.log")); !strings.Contains(string(log), `{"n": 1}`) {
		t.Errorf("recipe.log = %q, want the inputs", log)
	}
	if _, err := Remove(RemoveOptions{ID: "in1", OutDir: filepath.Join(t.TempDir(), "x"), Keep: []string{"inputs"}}); code(err) != 2 {
		t.Errorf("--keep inputs: err = %v, want 2", err)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "in1", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(filepath.Join(out, "inputs", "params.json")); string(data) != `{"n": 1}`+"\n" {
		t.Errorf("OUT_DIR/inputs/params.json = %q", data)
	}
}
