package store

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
)

// TestAppendSortsAfterNewest: the newest event's id is from a clock far ahead
// of this one, as on a branch merged in from a machine whose clock ran fast.
// Each appended id still sorts after every id the store holds.
func TestAppendSortsAfterNewest(t *testing.T) {
	root := newStore(t)
	g, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "goal", "why"))
	if err != nil {
		t.Fatal(err)
	}
	// 7ZZZZZZZZZ is the greatest ULID time, year 10889.
	future := "7ZZZZZZZZZ0000000000000000"
	f, err := os.OpenFile(filepath.Join(root, ".tree", "events", "z.jsonl"), os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"v":1,"id":"` + future + `","ts":"2026-10-27T00:00:00Z","actor":"z","type":"node.add","node":"` + future + `","payload":{"kind":"DesignQuestion","parent":"` + g.ID + `","title":"q"},"why":"w"}` + "\n")
	f.Close()
	var got []string
	for _, actor := range []string{"a", "b", "a"} {
		ev, err := Append(root, actor, nodeAdd(t, event.KindDesignQuestion, g.ID, "q", "why"))
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, ev.ID)
	}
	evs, _, err := Events(root)
	if err != nil {
		t.Fatal(err)
	}
	var all []string
	for _, ev := range evs {
		all = append(all, ev.ID)
	}
	if want := append([]string{g.ID, future}, got...); !slices.Equal(all, want) {
		t.Fatalf("merged ids %v, want %v: each appended id after %s", all, want, future)
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

// TestUnionMergeKeepsBothBranches: init writes .tree/.gitattributes holding
// one line; with it, git merges two branches that each appended to the same
// actor's events file with exit 0, and the file holds every line of both
// branches once and no other line.
func TestUnionMergeKeepsBothBranches(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	if err := Init(repo); err != nil {
		t.Fatal(err)
	}
	ga, err := os.ReadFile(filepath.Join(repo, ".tree", ".gitattributes"))
	if err != nil || string(ga) != "events/*.jsonl merge=union\n" {
		t.Fatalf(".tree/.gitattributes = %q (%v), want exactly the line events/*.jsonl merge=union", ga, err)
	}
	g, err := Append(repo, "a", nodeAdd(t, event.KindGoal, "", "goal", "why"))
	if err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", "-A")
	gitRun(t, repo, "commit", "-qm", "goal")
	file := filepath.Join(repo, ".tree", "events", "a.jsonl")
	read := func() []string {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		return strings.SplitAfter(strings.TrimSuffix(string(data), "\n"), "\n")
	}
	gitRun(t, repo, "checkout", "-q", "-b", "side")
	for range 2 {
		if _, err := Append(repo, "a", nodeAdd(t, event.KindDesignQuestion, g.ID, "side", "why")); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, repo, "commit", "-qam", "side")
	side := read()
	gitRun(t, repo, "checkout", "-q", "main")
	for range 2 {
		if _, err := Append(repo, "a", nodeAdd(t, event.KindDesignQuestion, g.ID, "main", "why")); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, repo, "commit", "-qam", "main")
	main := read()
	cmd := exec.Command("git", "-c", "user.name=t", "-c", "user.email=t@t", "merge", "--no-edit", "side")
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git merge: %v: %s", err, out)
	}
	merged := read()
	want := map[string]bool{}
	for _, l := range append(side, main...) {
		want[strings.TrimSuffix(l, "\n")] = true
	}
	got := map[string]int{}
	for _, l := range merged {
		got[strings.TrimSuffix(l, "\n")]++
	}
	if len(merged) != len(want) {
		t.Errorf("merged file has %d lines, want %d", len(merged), len(want))
	}
	for l := range want {
		if got[l] != 1 {
			t.Errorf("line held %d times, want once: %s", got[l], l)
		}
	}
	for l := range got {
		if !want[l] {
			t.Errorf("line of neither branch: %q", l)
		}
	}
}
