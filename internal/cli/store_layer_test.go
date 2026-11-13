package cli

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/store"
)

// treeFiles maps each regular file under dir, slash-relative, to its bytes.
func treeFiles(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		data, err := os.ReadFile(p)
		out[filepath.ToSlash(rel)] = data
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// eventIDs returns the id of each line of the events file of actor, in file order.
func eventIDs(t *testing.T, root, actor string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".tree", "events", actor+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var ev struct{ ID string }
		if err := json.Unmarshal([]byte(l), &ev); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, ev.ID)
	}
	return ids
}

// layerStore makes a store in a new directory with events of two actors and
// two Experiments, each with an evidence folder, and returns the directory and
// the full id of the first Experiment's add.
func layerStore(t *testing.T) (dir, e1 string) {
	t.Helper()
	dir = t.TempDir()
	t.Chdir(dir)
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	g := must(t, "add", "goal", "--title", "G", "--line", "a case", "--why", "start")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "try", "--actor", "b")
	x1 := must(t, "add", "exp", "--parent", h, "--title", "E1", "--recipe", "none", "--rubric", "reading", "--why", "run")
	e1 = eventIDs(t, dir, "a")[2]
	must(t, "checkpoint", "--goal", g, "--summary", "after E1", "--why", "mid", "--actor", "b")
	x2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--rubric", "reading", "--why", "run again")
	for _, x := range []string{x1, x2} {
		p := filepath.Join(dir, ".tree", "evidence", x, "run.log")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(x+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir, e1
}

// TestSandboxCutOverBaseTree: on a base whose root holds .tree, a store cut
// succeeds and replaces that .tree whole: the clone's .tree is the cut and
// nothing else, its status is the source's status as of the cut, and no
// evidence folder of a later Experiment is in it.
func TestSandboxCutOverBaseTree(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	src, cut := layerStore(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	h := t.TempDir()
	gitIn(t, h, "init", "-q", "-b", "main")
	for rel, body := range map[string]string{
		"a.txt":                       "a\n",
		".tree/events/old.jsonl":      `{"v":1,"id":"01JB0000000000000000000001","ts":"2026-10-27T00:00:00Z","actor":"old","type":"node.add","node":"01JB0000000000000000000001","payload":{"kind":"Goal","parent":"","title":"old"},"why":"w"}` + "\n",
		".tree/.gitignore":            "index.sqlite\n",
		".tree/evidence/E-000001/x":   "old\n",
		".tree/evidence/notes-old.md": "old\n",
	} {
		p := filepath.Join(h, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitIn(t, h, "add", "-A")
	gitIn(t, h, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "base with a tree")
	sha := gitIn(t, h, "rev-parse", "HEAD")
	stub := filepath.Join(t.TempDir(), "hyp")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho hyp 0.0.0-test\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	want(t, hyp(t, "sandbox", "create", "t1", "--host", h, "--base", sha, "--hyp", stub, "--recipe", "none", "--store", src, "--cut", cut), 0, "sandbox t1:")
	wt := filepath.Join(home, ".hyp-sandbox", "t1", "wt")

	c, err := store.CutAt(src, cut)
	if err != nil {
		t.Fatal(err)
	}
	expect := t.TempDir()
	if err := c.Write(expect); err != nil {
		t.Fatal(err)
	}
	got, wantFiles := treeFiles(t, filepath.Join(wt, ".tree")), treeFiles(t, filepath.Join(expect, ".tree"))
	if !maps.EqualFunc(got, wantFiles, bytes.Equal) {
		t.Errorf("clone .tree holds %v, want the cut's %v", slices.Sorted(maps.Keys(got)), slices.Sorted(maps.Keys(wantFiles)))
	}
	x2 := "E-" + eventIDs(t, src, "a")[3][20:]
	if _, err := os.Stat(filepath.Join(wt, ".tree", "evidence", x2)); !os.IsNotExist(err) {
		t.Errorf("clone holds the evidence of %s, added after the cut: %v", x2, err)
	}

	t.Chdir(wt)
	inClone := hyp(t, "status")
	t.Chdir(src)
	asOf := hyp(t, "status", "--as-of", cut)
	if inClone.code != 0 || inClone != asOf {
		t.Errorf("status in the clone:\n%+v\nstatus --as-of %s in the source:\n%+v", inClone, cut, asOf)
	}
	want(t, hyp(t, "sandbox", "remove", "t1", "--discard"), 0, "removed t1; discarded")
}

// TestInitRefusedInsideStore: hyp init where store discovery finds a store,
// at the directory or above it, is refused with exit 1, names the store's
// directory and creates nothing.
func TestInitRefusedInsideStore(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	want(t, hyp(t, "init"), 0, "initialised .tree")
	before := treeFiles(t, root)
	want(t, hyp(t, "init"), 1, root)
	sub := filepath.Join(root, "investigations", "topic")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	r := hyp(t, "init")
	want(t, r, 1, root)
	if !strings.HasPrefix(r.stderr, "refused: ") {
		t.Errorf("stderr %q, want a refusal", r.stderr)
	}
	if _, err := os.Lstat(filepath.Join(sub, ".tree")); !os.IsNotExist(err) {
		t.Errorf("refused init made %s/.tree: %v", sub, err)
	}
	if after := treeFiles(t, root); !maps.EqualFunc(before, after, bytes.Equal) {
		t.Errorf("refused init changed files: %v, was %v", slices.Sorted(maps.Keys(after)), slices.Sorted(maps.Keys(before)))
	}
}

// TestDiscoveryStopsAtRepositoryRoot: discovery walks up from the current
// directory and stops at the first directory holding .git, a directory or a
// worktree's file, so a store above it is not found (exit 3) and hyp init
// there is admitted. With no .git on the way, the store above is found.
func TestDiscoveryStopsAtRepositoryRoot(t *testing.T) {
	outer := t.TempDir()
	t.Chdir(outer)
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	must(t, "add", "goal", "--title", "outer", "--why", "start")
	for _, c := range []struct {
		name string
		mk   func(repo string) error
	}{
		{"git directory", func(repo string) error { return os.Mkdir(filepath.Join(repo, ".git"), 0o755) }},
		{"git file", func(repo string) error {
			return os.WriteFile(filepath.Join(repo, ".git"), []byte("gitdir: /elsewhere\n"), 0o644)
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			repo := filepath.Join(outer, strings.ReplaceAll(c.name, " ", "-"))
			sub := filepath.Join(repo, "a", "b")
			if err := os.MkdirAll(sub, 0o755); err != nil {
				t.Fatal(err)
			}
			t.Chdir(sub)
			want(t, hyp(t, "status"), 0, "outer")
			if err := c.mk(repo); err != nil {
				t.Fatal(err)
			}
			want(t, hyp(t, "status"), 3, "no .tree store found")
			t.Chdir(repo)
			want(t, hyp(t, "status"), 3, "no .tree store found")
			want(t, hyp(t, "init"), 0, "initialised .tree")
			t.Chdir(sub)
			if r := hyp(t, "status"); r.code != 0 || strings.Contains(r.stdout, "outer") {
				t.Errorf("status below the new store: %+v", r)
			}
		})
	}
}

// TestReplayIgnoresLineOrder: the same lines in another order inside each
// events file print the same status and check, and cut to the same bytes,
// each events file's kept lines in id order.
func TestReplayIgnoresLineOrder(t *testing.T) {
	src, cut := layerStore(t)
	shuffled := t.TempDir()
	t.Chdir(shuffled)
	must(t, "init")
	for _, actor := range []string{"a", "b"} {
		data, err := os.ReadFile(filepath.Join(src, ".tree", "events", actor+".jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.SplitAfter(strings.TrimSuffix(string(data), "\n"), "\n")
		lines[len(lines)-1] += "\n"
		slices.Reverse(lines)
		if err := os.WriteFile(filepath.Join(shuffled, ".tree", "events", actor+".jsonl"), []byte(strings.Join(lines, "")), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.CopyFS(filepath.Join(shuffled, ".tree", "evidence"), os.DirFS(filepath.Join(src, ".tree", "evidence"))); err != nil {
		t.Fatal(err)
	}
	if ids := eventIDs(t, shuffled, "a"); slices.IsSorted(ids) {
		t.Fatalf("shuffled a.jsonl is in id order: %v", ids)
	}
	for _, args := range [][]string{{"status"}, {"check"}, {"status", "--as-of", cut}} {
		t.Chdir(src)
		inOrder := hyp(t, args...)
		t.Chdir(shuffled)
		other := hyp(t, args...)
		if inOrder != other {
			t.Errorf("hyp %v differs:\nin order %+v\nshuffled %+v", args, inOrder, other)
		}
	}
	var cuts [2]map[string][]byte
	for i, dir := range []string{src, shuffled} {
		c, err := store.CutAt(dir, cut)
		if err != nil {
			t.Fatal(err)
		}
		out := t.TempDir()
		if err := c.Write(out); err != nil {
			t.Fatal(err)
		}
		cuts[i] = treeFiles(t, filepath.Join(out, ".tree"))
		for _, actor := range []string{"a", "b"} {
			if _, ok := cuts[i]["events/"+actor+".jsonl"]; ok {
				if ids := eventIDs(t, out, actor); !slices.IsSorted(ids) {
					t.Errorf("cut of %s: %s.jsonl not in id order: %v", dir, actor, ids)
				}
			}
		}
	}
	if !maps.EqualFunc(cuts[0], cuts[1], bytes.Equal) {
		t.Errorf("cuts differ: %v vs %v", slices.Sorted(maps.Keys(cuts[0])), slices.Sorted(maps.Keys(cuts[1])))
	}
}
