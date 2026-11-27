package acts

import (
	"slices"
	"testing"
)

// catTLA is a kept model whose one step takes the command cat, the command
// of counterEx's run section.
const catTLA = "---- MODULE kept ----\n\\* command: cat\nRead == n' = n\n====\n"

func TestCommitViewWalksOnTheKeptModelOfItsCommit(t *testing.T) {
	e := newRunEnv(t)
	hash := e.goalRow(t).Placed()[0].Hash
	if kept, err := CommitAt(e.root, e.base).Keeps(); err != nil || kept {
		t.Fatalf("H0 keeps = %v, %v; want false", kept, err)
	}
	if w, err := CommitAt(e.root, e.base).Walk(hash); err != nil || w.Kept || len(w.Steps) != 0 {
		t.Fatalf("walk on H0 = %+v, %v; want no walk", w, err)
	}
	c := e.child(t, e.base, map[string]string{KeptDir + "kept.tla": catTLA})
	v := CommitAt(e.root, c)
	if kept, err := v.Keeps(); err != nil || !kept {
		t.Fatalf("keeps = %v, %v; want true", kept, err)
	}
	w, err := v.Walk(hash)
	if err != nil {
		t.Fatal(err)
	}
	want := []WalkStep{{Command: "cat counter", Step: "Read"}}
	if !w.Kept || !slices.Equal(w.Steps, want) {
		t.Fatalf("walk = %+v, want %+v", w, want)
	}
	if _, err := v.Walk("0000000000000000000000000000000000000000"); err == nil {
		t.Fatal("walk of an example this clone lacks read")
	}
}

func TestGoalViewWithNoSelectWalksOnH0AndReadsNoSelectedCommit(t *testing.T) {
	e := newRunEnv(t)
	g := e.shown(t)
	v := ViewOf(e.root, g)
	if commit, selected := v.WalkCommit(); commit != e.base || selected {
		t.Fatalf("walk commit = %s, %v; want H0", commit, selected)
	}
	if v.OnNoRef() {
		t.Fatal("a Goal with no select reported on no ref")
	}
	if kept, err := v.Kept(); err != nil || len(kept) != 0 {
		t.Fatalf("kept = %+v, %v; want none", kept, err)
	}
	if charts, err := v.Charts(); err != nil || len(charts) != 0 {
		t.Fatalf("charts = %+v, %v; want none", charts, err)
	}
	if kept, err := v.Keeps(); err != nil || kept {
		t.Fatalf("keeps = %v, %v; want false", kept, err)
	}
	words, err := v.ExampleWords(g.Placed()[0].Hash)
	if err != nil || words != "w" {
		t.Fatalf("words = %q, %v; want w", words, err)
	}
}

func TestGoalViewOfASelectedGoalIsWhatItsCommitRead(t *testing.T) {
	e := newRunEnv(t)
	both := "SPECIFICATION Spec\nINVARIANTS OldRule NewRule\n"
	if err := e.selectKept(t, map[string]string{"kept.tla": keptTLA, "kept.cfg": keptCfg},
		map[string]string{"variants/both.cfg": both, "chart.mmd": chartMmd}); err != nil {
		t.Fatal(err)
	}
	g := e.shown(t)
	if g.Select == nil {
		t.Fatal("no select")
	}
	v := ViewOf(e.root, g)
	read, err := SelectedRead(e.root, g)
	if err != nil {
		t.Fatal(err)
	}
	if commit, selected := v.WalkCommit(); commit != g.Select.Commit || !selected {
		t.Fatalf("walk commit = %s, %v; want the selected commit", commit, selected)
	}
	got, err := v.Selected()
	if err != nil {
		t.Fatal(err)
	}
	again, _ := v.Selected()
	if got != again {
		t.Fatal("the selected commit was read twice")
	}
	kept, err := v.Kept()
	if err != nil || !slices.Equal(kept, read.Kept) {
		t.Fatalf("kept = %+v, %v; want %+v", kept, err, read.Kept)
	}
	charts, err := v.Charts()
	if err != nil || len(charts) != 1 || charts[0].Text != chartMmd {
		t.Fatalf("charts = %+v, %v", charts, err)
	}
	keeps, err := v.Keeps()
	if err != nil || keeps != read.Keeps {
		t.Fatalf("keeps = %v, %v; want %v", keeps, err, read.Keeps)
	}
	w, err := v.Walk(g.Placed()[0].Hash)
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := WalkRead(e.root, read, g.Placed()[0].Hash); w.Kept != want.Kept || !slices.Equal(w.Steps, want.Steps) {
		t.Fatalf("walk = %+v, want %+v", w, want)
	}
}
