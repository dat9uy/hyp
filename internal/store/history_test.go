package store

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// TestAppendSortsAfterHistory: an appended event's id sorts after every
// event id the history holds when the append takes its lock, in one clone,
// also when two actors append to two stores at once. Store a holds an id
// from a clock far ahead; appends to store b sort after it, and appends to
// a and b at once, each numbered under the lock, come out of the merge in
// the order they took it.
func TestAppendSortsAfterHistory(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	dirs := map[string]string{"a": filepath.Join(repo, "a"), "b": filepath.Join(repo, "b")}
	goals := map[string]string{}
	for name, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := Init(dir); err != nil {
			t.Fatal(err)
		}
		g, err := Append(dir, name, nodeAdd(t, event.KindGoal, "", "goal "+name, "why"))
		if err != nil {
			t.Fatal(err)
		}
		goals[name] = g.ID
	}
	// 7ZZZZZZZZZ is the greatest ULID time, year 10889.
	future := "7ZZZZZZZZZ0000000000000000"
	f, err := os.OpenFile(filepath.Join(dirs["a"], ".tree", "events", "z.jsonl"), os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(f, `{"v":1,"id":%q,"ts":"2026-10-27T00:00:00Z","actor":"z","type":"node.add","node":%q,"payload":{"kind":"DesignQuestion","parent":%q,"title":"q"},"why":"w"}`+"\n", future, future, goals["a"])
	f.Close()
	gitRun(t, repo, "add", "-A")

	hists := map[string]*History{}
	for name, dir := range dirs {
		h, err := OpenHistory(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(h.Stores) != 2 || h.Current != name {
			t.Fatalf("history from %s: stores %v, current %q", name, h.Stores, h.Current)
		}
		hists[name] = h
	}

	var (
		mu   sync.Mutex
		seq  int
		wg   sync.WaitGroup
		errs = make(chan error, 40)
	)
	for _, name := range []string{"a", "b"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 15 {
				_, err := hists[name].AppendWith(name, func(*replay.State, replay.Projection) (event.Event, error) {
					mu.Lock()
					seq++
					n := seq
					mu.Unlock()
					return nodeAdd(t, event.KindDesignQuestion, goals[name], "q", "seq "+strconv.Itoa(n)), nil
				})
				if err != nil {
					errs <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}

	whole, err := OpenHistory(repo)
	if err != nil {
		t.Fatal(err)
	}
	evs, _, err := whole.Events()
	if err != nil {
		t.Fatal(err)
	}
	var order []int
	for _, ev := range evs {
		n, ok := strings.CutPrefix(ev.Why, "seq ")
		if !ok {
			continue
		}
		if ev.ID <= future {
			t.Errorf("appended id %s (%s, store %s) sorts at or before %s, an id of store a", ev.ID, ev.Why, ev.Store, future)
		}
		i, _ := strconv.Atoi(n)
		order = append(order, i)
	}
	if len(order) != 30 || !slices.IsSorted(order) {
		t.Errorf("merged appends in lock order %v, want 1 to 30 in order", order)
	}
}
