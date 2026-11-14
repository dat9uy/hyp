package store

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
)

func newStore(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := Init(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func nodeAdd(t *testing.T, kind, parent, title, why string) event.Event {
	t.Helper()
	raw, err := json.Marshal(event.NodeAdd{Kind: kind, Parent: parent, Title: title})
	if err != nil {
		t.Fatal(err)
	}
	return event.Event{Type: event.TypeNodeAdd, Payload: raw, Why: why}
}

func lines(t *testing.T, root, actor string) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".tree", "events", actor+".jsonl"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return bytes.Count(data, []byte("\n"))
}

func TestInitRefusesExistingStore(t *testing.T) {
	root := newStore(t)
	if err := Init(root); err == nil {
		t.Fatal("second init succeeded")
	}
	gi, _ := os.ReadFile(filepath.Join(root, ".tree", ".gitignore"))
	if string(gi) != "index.sqlite\nindex.sqlite-*\nevidence/\n" {
		t.Fatalf(".gitignore = %q", gi)
	}
	sub := filepath.Join(root, "a", "b")
	os.MkdirAll(sub, 0o755)
	if got, err := FindRoot(sub); err != nil || got != root {
		t.Fatalf("FindRoot = %q %v", got, err)
	}
	if _, err := FindRoot(t.TempDir()); err != ErrNoStore {
		t.Fatalf("FindRoot without store = %v", err)
	}
}

func TestAppendRefusesAndWritesNothing(t *testing.T) {
	root := newStore(t)
	g, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "goal", "why"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Append(root, "b", nodeAdd(t, event.KindHypothesis, g.ID, "h", "why"))
	if err == nil || !strings.HasPrefix(err.Error(), "refused:") {
		t.Fatalf("err = %v, want refused", err)
	}
	if n := lines(t, root, "b"); n != 0 {
		t.Fatalf("refused event wrote %d lines", n)
	}
}

func TestAppendRefusesMissingWhy(t *testing.T) {
	root := newStore(t)
	_, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "goal", ""))
	if err == nil || !strings.Contains(err.Error(), "refused: missing field: why") {
		t.Fatalf("err = %v", err)
	}
	if n := lines(t, root, "a"); n != 0 {
		t.Fatalf("refused event wrote %d lines", n)
	}
}

func TestTwoActorsMerge(t *testing.T) {
	root := newStore(t)
	g, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "goal", "why"))
	if err != nil {
		t.Fatal(err)
	}
	q, err := Append(root, "b", nodeAdd(t, event.KindDesignQuestion, g.ID, "q", "why"))
	if err != nil {
		t.Fatal(err)
	}
	p, lineErrs, err := Load(root)
	if err != nil || len(lineErrs) > 0 || len(p.Rejections) > 0 {
		t.Fatalf("load: %v %v %v", err, lineErrs, p.Rejections)
	}
	if p.Nodes[g.ID] == nil || p.Nodes[q.ID] == nil {
		t.Fatal("load did not see both actors")
	}
}

func TestConcurrentAppendSameActor(t *testing.T) {
	root := newStore(t)
	g, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "goal", "why"))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := Append(root, "w", nodeAdd(t, event.KindDesignQuestion, g.ID, "q", "why"))
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	files, lineErrs, err := ReadAll(root)
	if err != nil || len(lineErrs) > 0 || len(files["w"]) != 20 {
		t.Fatalf("w has %d valid lines, errs %v %v", len(files["w"]), lineErrs, err)
	}
}

func TestReadAllReportsBadLine(t *testing.T) {
	root := newStore(t)
	if _, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "goal", "why")); err != nil {
		t.Fatal(err)
	}
	f, _ := os.OpenFile(filepath.Join(root, ".tree", "events", "a.jsonl"), os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString("{not json\n")
	f.Close()
	files, lineErrs, err := ReadAll(root)
	if err != nil || len(files["a"]) != 1 || len(lineErrs) != 1 || lineErrs[0].Line != 2 {
		t.Fatalf("files=%d lineErrs=%+v err=%v", len(files["a"]), lineErrs, err)
	}
}
