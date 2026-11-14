package record

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestMain runs one record write and stops the process after the step that
// HYP_RECORD_STOP_AFTER names, when it is set, as a kill there would.
func TestMain(m *testing.M) {
	if id := os.Getenv("HYP_RECORD_WRITE_READ"); id != "" {
		// Write one event, then read the store, which writes the new tip's
		// snapshot and prunes the old ones.
		dir := os.Getenv("HYP_RECORD_DIR")
		if err := writeOnce(dir, id); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		s, err := Locate(dir)
		if err == nil {
			_, err = s.Root()
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(0)
	}
	if step := os.Getenv("HYP_RECORD_STOP_AFTER"); step != "" {
		afterStep = func(s string) {
			if s == step {
				os.Exit(9)
			}
		}
		if err := writeOnce(os.Getenv("HYP_RECORD_DIR"), "second"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// writeOnce appends one line named id to the store at dir, with the
// evidence file out-<id>.txt it names.
func writeOnce(dir, id string) error {
	s, err := Locate(dir)
	if err != nil {
		return err
	}
	release, err := s.Lock()
	if err != nil {
		return err
	}
	defer release()
	rel := EvidenceDir + "E-1/out-" + id + ".txt"
	src := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(src, []byte("bytes of "+id+"\n"), 0o644); err != nil {
		return err
	}
	return s.Append(map[string][][]byte{"a": {[]byte(`{"id":"` + id + `"}`)}}, []File{s.StoreFile(rel, src, "")}, nil)
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

// Line 3: a write stopped after any of its steps leaves the record holding
// the whole event, with the bytes of every file it names, or nothing of it.
func TestStoppedWriteLeavesAllOrNothing(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "none"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, step := range []string{"objects", "tree", "commit", "ref"} {
		t.Run(step, func(t *testing.T) {
			repo, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			git(t, repo, "init", "-q")
			dir := filepath.Join(repo, "s")
			if err := Init(dir); err != nil {
				t.Fatal(err)
			}
			if err := writeOnce(dir, "first"); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^$")
			cmd.Env = append(os.Environ(), "HYP_RECORD_STOP_AFTER="+step, "HYP_RECORD_DIR="+dir)
			out, err := cmd.CombinedOutput()
			if code := cmd.ProcessState.ExitCode(); code != 9 {
				t.Fatalf("the stopped write exited %d (%v): %s", code, err, out)
			}
			// The work-tree files the write named are gone, as after a
			// clean: the record alone holds what a reader sees.
			if err := os.RemoveAll(EvidenceRoot(dir)); err != nil {
				t.Fatal(err)
			}
			s, err := Locate(dir)
			if err != nil {
				t.Fatal(err)
			}
			files, err := s.EventFiles()
			if err != nil {
				t.Fatal(err)
			}
			lines := bytes.Count(files["a"], []byte("\n"))
			second, ferr := s.ReadFile(EvidenceDir + "E-1/out-second.txt")
			switch {
			case step == "ref":
				if lines != 2 || ferr != nil || string(second) != "bytes of second\n" {
					t.Fatalf("after the ref moved the record holds %d lines and file %q (%v), want the whole event", lines, second, ferr)
				}
			default:
				if lines != 1 || ferr == nil {
					t.Fatalf("stopped after %s, the record holds %d lines and file %q, want nothing of the event", step, lines, second)
				}
			}
			first, err := s.ReadFile(EvidenceDir + "E-1/out-first.txt")
			if err != nil || string(first) != "bytes of first\n" {
				t.Fatalf("the first event's file reads %q (%v)", first, err)
			}
		})
	}
}

// Line 3: a prune never removes a snapshot that a running read uses. This
// process reads the record and holds its snapshot; a later command writes
// and reads, and so prunes old snapshots: the held one stays, whatever its
// age, and one no running read holds goes.
func TestPruneLeavesASnapshotAReadHolds(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "none"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	git(t, repo, "init", "-q")
	dir := filepath.Join(repo, "s")
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}
	if err := writeOnce(dir, "first"); err != nil {
		t.Fatal(err)
	}
	s, err := Locate(dir)
	if err != nil {
		t.Fatal(err)
	}
	root, err := s.Root()
	if err != nil {
		t.Fatal(err)
	}
	held := snapRoot(t, root)
	old := time.Now().Add(-2 * snapshotKeep)
	run := func(id string) {
		t.Helper()
		cmd := exec.Command(os.Args[0], "-test.run=^$")
		cmd.Env = append(os.Environ(), "HYP_RECORD_WRITE_READ="+id, "HYP_RECORD_DIR="+dir)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("the command that writes %s and reads: %v: %s", id, err, out)
		}
	}
	snaps := func() []string {
		t.Helper()
		ents, err := os.ReadDir(filepath.Dir(held))
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, e := range ents {
			out = append(out, filepath.Join(filepath.Dir(held), e.Name()))
		}
		return out
	}
	if err := os.Chtimes(held, old, old); err != nil {
		t.Fatal(err)
	}
	run("second")
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(EvidenceDir+"E-1/out-first.txt"))); err != nil {
		t.Fatalf("a prune removed the snapshot this read holds: %v", err)
	}
	// The second command's snapshot: no running read holds it now.
	var free string
	for _, p := range snaps() {
		if p != held {
			free = p
		}
	}
	if free == "" {
		t.Fatal("the second command wrote no snapshot")
	}
	if err := os.Chtimes(free, old, old); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(held, old, old); err != nil {
		t.Fatal(err)
	}
	run("third")
	if _, err := os.Stat(free); err == nil {
		t.Fatalf("the old snapshot %s, which no read holds, was not pruned", free)
	}
	if _, err := os.Stat(held); err != nil {
		t.Fatalf("a prune removed the snapshot this read holds: %v", err)
	}
}

// snapRoot is the snapshot folder that holds the store root root.
func snapRoot(t *testing.T, root string) string {
	t.Helper()
	for d := root; filepath.Dir(d) != d; d = filepath.Dir(d) {
		if filepath.Base(filepath.Dir(d)) == "snap" {
			return d
		}
	}
	t.Fatalf("%s is in no snapshot", root)
	return ""
}
