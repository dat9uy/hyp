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

func TestAnchor(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		root, p, want string
		ok            bool
	}{
		{root, filepath.Join(root, "a", "b"), "repo/a/b", true},
		{root, root, "repo", true},
		{root, filepath.Join(root, "missing", "f.txt"), "repo/missing/f.txt", true},
		{root, filepath.Dir(root), "", false},
		{root, filepath.Join(root, "..", "x"), "", false},
		{root, "/elsewhere", "", false},
		{link, filepath.Join(root, "a"), "repo/a", true},
		{root, filepath.Join(link, "a", "b"), "repo/a/b", true},
	} {
		got, ok := anchor(c.root, "repo", c.p)
		if got != c.want || ok != c.ok {
			t.Errorf("anchor(%s, repo, %s) = %q, %v; want %q, %v", c.root, c.p, got, ok, c.want, c.ok)
		}
	}
}

// A record names no host path: host is a directory name, a store
// inside the host is anchored to it, and every live path is derived.
func TestRecordNamesNoHostPath(t *testing.T) {
	setup(t)
	home := os.Getenv("HOME")
	h, base := host(t)
	src := filepath.Join(h, "inv")
	if err := store.Init(src); err != nil {
		t.Fatal(err)
	}
	ev, err := store.Append(src, "a", event.Event{Type: event.TypeNodeAdd, Payload: []byte(`{"kind":"Goal","parent":"","title":"g"}`), Why: "w"})
	if err != nil {
		t.Fatal(err)
	}
	in := filepath.Join(t.TempDir(), "params.json")
	writeFile(t, in, "{}\n")
	rec, box, err := createWith(t, "np", CreateOptions{Host: h, Base: base, Refs: []string{"none"}, Store: src, Cut: ev.ID, Inputs: in})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(box, "sandbox.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, prefix := range []string{home, filepath.Dir(h), os.TempDir()} {
		if r, _ := filepath.EvalSymlinks(prefix); strings.Contains(string(raw), prefix) || (r != "" && strings.Contains(string(raw), r)) {
			t.Errorf("sandbox.json names %s:\n%s", prefix, raw)
		}
	}
	name := filepath.Base(h)
	if rec.Host != name || rec.Store != name+"/inv" || rec.Format != 4 {
		t.Errorf("host = %q, store = %q, format = %d", rec.Host, rec.Store, rec.Format)
	}
	if rec.WTDir() != filepath.Join(box, "wt") || rec.InputsFile() != filepath.Join(box, "inputs", "params.json") {
		t.Errorf("wt = %s, inputs = %s", rec.WTDir(), rec.InputsFile())
	}
}

// A sandbox whose record a format-2 hyp wrote, with every live path in it,
// still drives and removes, and its old fields are copied out as read.
func TestFormat2RecordDrivesAndRemoves(t *testing.T) {
	setup(t)
	in := filepath.Join(t.TempDir(), "params.json")
	writeFile(t, in, `{"n": 2}`+"\n")
	files := map[string]string{
		"recipe.md": recipeMD("boot", "## Drive\n\nRun: ./drive.sh\n"),
		"drive.sh":  "#!/bin/sh\nwhile IFS= read -r l; do echo \"$l\"; done < \"$HYP_SANDBOX_INPUTS\" > \"$HYP_SANDBOX_WT/seen.json\"\n",
	}
	root := recipeRoot(t, "boot", files, "drive.sh")
	rec, box, err := createWith(t, "f2", CreateOptions{Refs: []string{"boot@1"}, Recipes: root, Inputs: in})
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]any{
		"format": 2, "id": "f2", "host": "/somewhere/" + rec.Host, "base": rec.Base,
		"wt": rec.WTDir(), "tmp": rec.TmpDir(), "bin": rec.BinDir(),
		"hyp_sha256": rec.HypSHA256, "hyp_version": rec.HypVersion, "created": rec.Created,
		"inputs": map[string]any{"source": in, "path": rec.InputsFile(), "sha256": rec.Inputs.SHA256},
		"recipes": []map[string]any{{
			"ref": "boot@1", "name": "boot", "version": "1", "sha256": rec.Recipes[0].SHA256,
			"path": filepath.Join(box, "recipes", "boot"), "setup": "none",
		}},
	}
	data, _ := json.Marshal(old)
	writeFile(t, filepath.Join(box, "sandbox.json"), string(data))
	var buf bytes.Buffer
	if _, err := Drive("f2", &buf); err != nil {
		t.Fatalf("drive: %v\n%s", err, buf.String())
	}
	if got, _ := os.ReadFile(filepath.Join(rec.WTDir(), "seen.json")); string(got) != `{"n": 2}`+"\n" {
		t.Errorf("drive saw inputs %q", got)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "f2", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(out, "inputs", "params.json")); string(got) != `{"n": 2}`+"\n" {
		t.Errorf("OUT_DIR/inputs/params.json = %q", got)
	}
	saved, _ := os.ReadFile(filepath.Join(out, "sandbox.json"))
	for _, k := range []string{`"wt": "` + rec.WTDir() + `"`, `"format": 2`, `"path": "` + filepath.Join(box, "recipes", "boot") + `"`} {
		if !strings.Contains(string(saved), k) {
			t.Errorf("saved record lacks %s:\n%s", k, saved)
		}
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}
