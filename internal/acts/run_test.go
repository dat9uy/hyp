package acts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
	"github.com/dat9uy/hyp/internal/testutil"
)

// markProbe is the Probe of the Probe contract in hyp spec, that also touches
// the file MARKER_FILE names, so a test sees whether a Probe ran.
const markProbe = `#!/bin/sh
[ -n "$MARKER_FILE" ] && touch "$MARKER_FILE"
o=$(mktemp); sh -c "$HYP_RUN" >"$o"; c=$?; cat "$o"; [ -s "$o" ] && [ -n "$(tail -c1 "$o")" ] && echo; rm -f "$o"; echo "exit $c"
`

// The subject prints its counter file; the floor "ok" is met when the
// counter holds ok, and failed on the base, which holds bad.
const counterEx = "-- words --\nw\n-- run --\ncat counter\n-- floor: contains --\nok\n-- external --\nMARKER_FILE\n"

// flakyEx prints a number that grows on every run, whatever the commit, so
// two runs on one tree differ: the floor "1" is met on the first run on a tree
// and failed on the second (the placement's own run on H0 printed 0).
const flakyEx = "-- words --\nflaky\n-- run --\nn=$(cat \"$COUNTER_FILE\"); echo $n; echo $((n+1)) > \"$COUNTER_FILE\"\n-- floor: contains --\n1\n-- external --\nCOUNTER_FILE\n"

type runEnv struct {
	*placeEnv
	base string // H0
}

// newRunEnv makes a Goal with one gap line on a base whose counter reads bad,
// and places line-1/a, a gap example that fails there. A Goal with more lines
// gets line-1/a placed the old way, one example per event, as a log written
// before the whole-set place holds it: it stays written.
func newRunEnv(t *testing.T, lines ...LineInput) *runEnv {
	t.Helper()
	if len(lines) == 0 {
		lines = []LineInput{gapLine}
	}
	r := testutil.NewRepo(t)
	// The store is not committed, so a checkout of another commit never meets
	// a modified events file.
	if err := os.WriteFile(filepath.Join(r.Dir, ".git", "info", "exclude"), []byte(".tree/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	base := r.Commit(t, map[string]string{"counter": "bad\n"})
	rep, err := AddGoal(opts(r.Dir, base, lines...))
	if err != nil {
		t.Fatal(err)
	}
	e := &runEnv{placeEnv: &placeEnv{r: r, goal: rep.Goal, root: r.Dir}, base: base}
	if len(lines) == 1 {
		e.add(t, "line-1/a", counterEx)
	} else {
		e.write(t, "line-1/a", counterEx, markProbe)
		e.commit(t)
		e.placeOldForm(t, "line-1/a")
	}
	return e
}

// placeOldForm places one committed example with an example.place event, the
// form a log written before the whole-set place holds. H0's cell is the one
// try shows for it.
func (e *placeEnv) placeOldForm(t *testing.T, key string) {
	t.Helper()
	tr, err := Try(TryOpts{Dir: e.root, Goal: e.goal, Key: key, ProbeTimeout: 20 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	line, name, _ := strings.Cut(key, "/")
	_, tb := e.log(t)
	g := tb.ByID(e.goal)
	payload, err := json.Marshal(events.ExamplePlace{
		Role: "grill", Line: line, Example: name, Hash: tr.Hash,
		H0: events.Cell{Commit: g.Base, RowTree: g.BaseTree, Probe: tr.ProbeHash, Output: tr.OutHash,
			Stderr: tr.ErrHash, Exit: tr.Out.Exit, Verdict: tr.Verdict},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := putOutputs(e.root, tr.Out.Stdout, tr.Out.Stderr); err != nil {
		t.Fatal(err)
	}
	if _, err := logstore.Append(e.root, "grill-1", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeExamplePlace, Node: g.EventID, Payload: payload}, nil
	}); err != nil {
		t.Fatal(err)
	}
}

// add commits an example folder and places it.
func (e *runEnv) add(t *testing.T, key, txtar string) {
	t.Helper()
	e.write(t, key, txtar, markProbe)
	e.commit(t)
	if _, err := e.place(); err != nil {
		t.Fatalf("place %s: %v", key, err)
	}
}

// child commits files on top of from, leaves HEAD on main, and returns the
// new commit.
func (e *runEnv) child(t *testing.T, from string, files map[string]string) string {
	t.Helper()
	testutil.Git(t, e.root, "checkout", "-q", "--detach", from)
	c := e.r.Commit(t, files)
	testutil.Git(t, e.root, "checkout", "-q", "main")
	return c
}

// leaf commits a counter value on top of from.
func (e *runEnv) leaf(t *testing.T, from, value string) string {
	t.Helper()
	return e.child(t, from, map[string]string{"counter": value + "\n"})
}

func (e *runEnv) opts(commit string) RunOpts {
	return RunOpts{Dir: e.root, Actor: "runner-1", Role: "runner", Commit: commit, Line: e.goal + "/line-1", ProbeTimeout: 20 * time.Second}
}

func (e *runEnv) run(commit string) (RunReport, error) { return Run(e.opts(commit)) }

func (e *runEnv) results(t *testing.T) []events.ResultRecord {
	t.Helper()
	evs, _ := e.log(t)
	var out []events.ResultRecord
	for _, ev := range evs {
		if ev.Type != events.TypeResultRecord {
			continue
		}
		var p events.ResultRecord
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	return out
}

func (e *runEnv) goalRow(t *testing.T) *tables.Goal {
	t.Helper()
	_, tb := e.log(t)
	g := tb.ByID(e.goal)
	if g == nil {
		t.Fatal("no Goal")
	}
	return g
}

func sha(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestRunRecordsResultWithCells(t *testing.T) {
	e := newRunEnv(t)
	c := e.leaf(t, e.base, "ok")
	rep, err := e.run(c)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Goal != e.goal || rep.Commit != c || len(rep.Cells) != 1 || rep.Cells[0].Key != "line-1/a" || rep.Cells[0].Verdict != "met" || rep.Cells[0].Words != "w" {
		t.Fatalf("report = %+v", rep)
	}
	res := e.results(t)
	if len(res) != 1 || res[0].Commit != c || res[0].Line != "line-1" || res[0].Role != "runner" || res[0].Refused != "" {
		t.Fatalf("results = %+v", res)
	}
	hash := testutil.Git(t, e.root, "rev-parse", "HEAD:goals/"+e.goal+"/line-1/a")
	cell := res[0].Cells[0]
	if cell.Example != "a" || cell.Hash != hash || cell.Verdict != "met" || cell.Exit != 0 || cell.Output != sha("ok") || cell.Output != rep.Cells[0].Output {
		t.Fatalf("cell = %+v", cell)
	}
	if cell.Probe == "" || cell.Stderr != sha("") {
		t.Fatalf("cell = %+v, want the Probe hash and the stderr hash", cell)
	}
	out, err := os.ReadFile(logstore.OutputPath(e.root, cell.Output))
	if err != nil || string(out) != "ok" {
		t.Fatalf("stored output = %q, %v; want the stdout with one newline dropped", out, err)
	}
	if g := e.goalRow(t); g.Verdict("line-1/a", res[0].RowTree) != "met" {
		t.Fatalf("the cell is not live on its row")
	}
}

// twoLines makes a Goal with a gap line and a guard line, with only line-1
// placed, so the Goal is written.
func twoLines(t *testing.T) *runEnv {
	t.Helper()
	e := newRunEnv(t, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	if e.goalRow(t).State != tables.StateWritten {
		t.Fatal("test setup: the Goal is open")
	}
	return e
}

func TestRunOnH0BeforeOpen(t *testing.T) {
	e := twoLines(t)
	rep, err := e.run(e.base)
	if err != nil {
		t.Fatalf("a run on H0 before the Goal opens was refused: %v", err)
	}
	if rep.Cells[0].Verdict != "failed" || rep.Used != 0 {
		t.Fatalf("report = %+v, want failed and no budget used", rep)
	}
	// line-2 has no placed example: nothing to run, even on H0.
	o := e.opts(e.base)
	o.Line = e.goal + "/line-2"
	_, err = Run(o)
	refusedBy(t, err, "Open")
}

func TestRunRefusesBeforeOpen(t *testing.T) {
	e := twoLines(t)
	_, err := e.run(e.leaf(t, e.base, "ok"))
	refusedBy(t, err, "Open")
	if n := len(e.results(t)); n != 0 {
		t.Fatalf("%d results after a refusal", n)
	}
}

func TestRunAfterSelectRefusedBySelectedState(t *testing.T) {
	e := newRunEnv(t)
	c := e.leaf(t, e.base, "ok")
	if _, err := e.run(c); err != nil {
		t.Fatal(err)
	}
	tree, err := gitread.RowTree(e.root, c)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := json.Marshal(events.Select{Role: "runner", Commit: c, RowTree: tree})
	if _, err := logstore.Append(e.root, "runner-1", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeSelect, Node: e.goalRow(t).EventID, Payload: p}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if e.goalRow(t).State != tables.StateSelected {
		t.Fatal("test setup: the Goal is not selected")
	}
	_, err = e.run(e.leaf(t, e.base, "ok2"))
	refusedBy(t, err, "Selected")
}

func TestRunRefusesAncestorRow(t *testing.T) {
	e := newRunEnv(t)
	c1 := e.leaf(t, e.base, "ok")
	c2 := e.leaf(t, c1, "ok2")
	if _, err := e.run(c2); err != nil {
		t.Fatalf("a leaf was refused: %v", err)
	}
	_, err := e.run(c1)
	refusedBy(t, err, "Leaf only")
	if _, err := e.run(e.base); err != nil {
		t.Fatalf("a run on H0 was refused: %v", err)
	}
}

func TestRunRepeatSameVerdictRecords(t *testing.T) {
	e := newRunEnv(t)
	c := e.leaf(t, e.base, "ok")
	if _, err := e.run(c); err != nil {
		t.Fatal(err)
	}
	rep, err := e.run(c)
	if err != nil {
		t.Fatalf("a repeat with the same verdict was refused: %v", err)
	}
	if n := len(e.results(t)); n != 2 || rep.Used != 1 {
		t.Fatalf("results = %d, used = %d; want 2 results and one tree used", n, rep.Used)
	}
}

// withFlaky places the flaky example, whose counter file lives in a temp dir.
func (e *runEnv) withFlaky(t *testing.T) {
	t.Helper()
	counter := filepath.Join(t.TempDir(), "n")
	if err := os.WriteFile(counter, []byte("0"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("COUNTER_FILE", counter)
	e.add(t, "line-1/b", flakyEx)
}

func TestRunRepeatDifferingVerdictRefused(t *testing.T) {
	e := newRunEnv(t)
	e.withFlaky(t)
	c := e.leaf(t, e.base, "ok")
	first, err := e.run(c)
	if err != nil {
		t.Fatal(err)
	}
	if first.Cells[1].Verdict != "met" {
		t.Fatalf("first run = %+v, want line-1/b met", first.Cells)
	}
	_, err = e.run(c)
	refusedBy(t, err, "Stable")
	if !testutil.Mentions(err, "line-1/b") {
		t.Fatalf("err = %v, want it to name line-1/b", err)
	}
	// The second output stays, and the Result is marked refused.
	if _, statErr := os.Stat(logstore.OutputPath(e.root, sha("2"))); statErr != nil {
		t.Fatalf("the second output file is gone: %v", statErr)
	}
	res := e.results(t)
	if len(res) != 2 || res[1].Refused != "Stable" || len(res[1].Differing) != 1 || res[1].Differing[0] != "line-1/b" {
		t.Fatalf("results = %+v, want the second marked refused by Stable naming line-1/b", res)
	}
	if !testutil.Mentions(err, sha("1")[:12]) || !testutil.Mentions(err, sha("2")[:12]) {
		t.Fatalf("err = %v, want both output hashes", err)
	}
	// The refused Result counts nothing and has no live cell.
	g := e.goalRow(t)
	if n := tables.RunsUsed(g); n != 1 {
		t.Fatalf("RunsUsed = %d, want 1", n)
	}
	if v := g.Verdict("line-1/b", res[0].RowTree); v != "met" {
		t.Fatalf("verdict = %s, want the first verdict to stand", v)
	}
}

func TestRunRunsEveryExampleOfTheLine(t *testing.T) {
	e := newRunEnv(t)
	e.withFlaky(t)
	rep, err := e.run(e.leaf(t, e.base, "ok"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Cells) != 2 || rep.Cells[0].Key != "line-1/a" || rep.Cells[1].Key != "line-1/b" {
		t.Fatalf("cells = %+v, want both examples in key order", rep.Cells)
	}
	if res := e.results(t); len(res) != 1 || len(res[0].Cells) != 2 {
		t.Fatalf("results = %+v, want one Result with two cells", res)
	}
}

func TestRunRefusesFourthTree(t *testing.T) {
	e := newRunEnv(t)
	var leaves []string
	for _, v := range []string{"ok1", "ok2", "ok3", "ok4"} {
		leaves = append(leaves, e.leaf(t, e.base, v))
	}
	for _, c := range leaves[:3] {
		if _, err := e.run(c); err != nil {
			t.Fatalf("one of the three trees was refused: %v", err)
		}
	}
	_, err := e.run(leaves[3])
	refusedBy(t, err, "Budget")
	if n := len(e.results(t)); n != 3 {
		t.Fatalf("%d results, want 3: a refused run costs nothing", n)
	}
	// H0 never counts, and a counted tree runs again.
	if _, err := e.run(e.base); err != nil {
		t.Fatalf("H0 was refused by Budget: %v", err)
	}
	if _, err := e.run(leaves[0]); err != nil {
		t.Fatalf("a counted tree was refused by Budget: %v", err)
	}
}

func TestRunPrintsRunsUsed(t *testing.T) {
	e := newRunEnv(t)
	rep, err := e.run(e.leaf(t, e.base, "ok1"))
	if err != nil || rep.Used != 1 || rep.LastRun {
		t.Fatalf("report = %+v, %v; want one tree used and no warning", rep, err)
	}
	if len(rep.Cells) != 1 || rep.Cells[0].Key != "line-1/a" || rep.Cells[0].Words == "" {
		t.Fatalf("report cells = %+v", rep.Cells)
	}
}

func TestRunWarnsBeforeLastRun(t *testing.T) {
	e := newRunEnv(t)
	var rep RunReport
	for _, v := range []string{"ok1", "ok2"} {
		var err error
		if rep, err = e.run(e.leaf(t, e.base, v)); err != nil {
			t.Fatal(err)
		}
	}
	if !rep.LastRun || rep.Used != 2 {
		t.Fatalf("report = %+v, want the last-run mark at two trees", rep)
	}
	rep, err := e.run(e.leaf(t, e.base, "ok3"))
	if err != nil || rep.LastRun {
		t.Fatalf("report = %+v, %v; want no warning on the last run", rep, err)
	}
}

func TestRunWritesNoGitState(t *testing.T) {
	e := newRunEnv(t)
	c := e.leaf(t, e.base, "ok")
	before := testutil.Snapshot(t, e.root)
	rep, err := e.run(c)
	if err != nil || rep.Cells[0].Verdict != "met" {
		t.Fatalf("run = %+v, %v; it must have recorded", rep, err)
	}
	if testutil.Snapshot(t, e.root) != before {
		t.Fatal("run changed the repository's git state")
	}
	_, err = e.run("no-such-commit")
	refusedBy(t, err, "Bytes")
	if testutil.Snapshot(t, e.root) != before {
		t.Fatal("a refused run changed the repository's git state")
	}
}

func TestRefusedRunRunsNoProbe(t *testing.T) {
	e := twoLines(t) // placed with MARKER_FILE unset: no marker yet
	marker := filepath.Join(t.TempDir(), "marker")
	t.Setenv("MARKER_FILE", marker)
	// Positive first: a run that is allowed runs the Probe and leaves the marker.
	if _, err := e.run(e.base); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("an allowed run did not run the Probe: %v", err)
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	// Open refuses a leaf before any Probe runs.
	_, err := e.run(e.leaf(t, e.base, "ok"))
	refusedBy(t, err, "Open")
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Fatal("a refused run ran a Probe")
	}
}

func TestRunChecksAgainAfterTheProbes(t *testing.T) {
	e := newRunEnv(t)
	c := e.leaf(t, e.base, "ok")
	o := e.opts(c)
	o.afterProbe = func() {
		// a new version of the example lands while the Probes run
		e.write(t, "line-1/a", counterEx+"-- ceiling --\nnever\n", markProbe)
		e.commit(t)
		if _, err := e.place(); err != nil {
			t.Errorf("the placement that lands between the checks failed: %v", err)
		}
	}
	_, err := Run(o)
	refusedBy(t, err, "Open")
	if n := len(e.results(t)); n != 0 {
		t.Fatalf("%d results recorded on a stale placement", n)
	}
}

func TestRunRefusesOtherGoalPath(t *testing.T) {
	e := newRunEnv(t)
	good := e.leaf(t, e.base, "ok")
	if _, err := e.run(good); err != nil {
		t.Fatal(err)
	}
	bad := e.child(t, e.base, map[string]string{"counter": "ok9\n", "goals/G-OTHER00/line-1/a/example.txtar": "x"})
	_, err := e.run(bad)
	refusedBy(t, err, "Own path")
	if !testutil.Mentions(err, "goals/G-OTHER00/line-1/a/example.txtar") || !testutil.Mentions(err, bad) {
		t.Fatalf("err = %v, want the path and the commit", err)
	}
}

func TestRunLeafMarkNeedsEveryCommit(t *testing.T) {
	e := newRunEnv(t)
	if _, err := e.run(e.leaf(t, e.base, "ok")); err != nil {
		t.Fatal(err)
	}
	// a Result at a commit this clone lacks, as a merged log can hold
	missing := strings.Repeat("0", 40)
	rowTree := strings.Repeat("1", 40)
	p, _ := json.Marshal(events.ResultRecord{Role: "runner", Line: "line-1", Commit: missing, RowTree: rowTree})
	if _, err := logstore.Append(e.root, "runner-2", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeResultRecord, Node: e.goalRow(t).EventID, Payload: p}, nil
	}); err != nil {
		t.Fatal(err)
	}
	_, err := e.run(e.leaf(t, e.base, "ok2"))
	if err == nil || fault.KindOf(err) != fault.Environment || !testutil.Mentions(err, missing) {
		t.Fatalf("err = %v, want an environment error naming %s", err, missing)
	}
}

func TestRunBadArguments(t *testing.T) {
	e := newRunEnv(t)
	c := e.leaf(t, e.base, "ok")
	for name, mod := range map[string]func(*RunOpts){
		"no slash":  func(o *RunOpts) { o.Line = "line-1" },
		"bad line":  func(o *RunOpts) { o.Line = e.goal + "/nope" },
		"unknown":   func(o *RunOpts) { o.Line = "G-NOPE00/line-1" },
		"no line":   func(o *RunOpts) { o.Line = e.goal + "/line-9" },
		"bad role":  func(o *RunOpts) { o.Role = "king" },
		"bad actor": func(o *RunOpts) { o.Actor = "Bad Actor" },
		"empty":     func(o *RunOpts) { o.Commit = "" },
	} {
		o := e.opts(c)
		mod(&o)
		if _, err := Run(o); err == nil || fault.KindOf(err) != fault.BadArguments {
			t.Errorf("%s: err = %v, want bad arguments", name, err)
		}
	}
}

func TestRunWithoutStoreIsRefusal(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	_, err := Run(RunOpts{Dir: r.Dir, Actor: "runner-1", Role: "runner", Commit: "HEAD", Line: "G-ABCDEF/line-1"})
	if fault.KindOf(err) != fault.Refusal || fault.CodeOf(err) != fault.NoStore {
		t.Fatalf("err = %v, want a refusal naming no store", err)
	}
}

// TestRunRefusedAfterTheProbesStoresNoOutput: a run refused by the second
// check, after its Probes ran, leaves no output of its own in the store.
func TestRunRefusedAfterTheProbesStoresNoOutput(t *testing.T) {
	e := newRunEnv(t)
	const printed = "only this run prints me"
	c := e.leaf(t, e.base, printed)
	o := e.opts(c)
	o.afterProbe = func() {
		e.write(t, "line-1/a", counterEx+"-- ceiling --\nnever\n", markProbe)
		e.commit(t)
		if _, err := e.place(); err != nil {
			t.Errorf("the placement that lands between the checks failed: %v", err)
		}
	}
	_, err := Run(o)
	refusedBy(t, err, "Open")
	own := logstore.Hash([]byte(printed))
	if slices.Contains(outputFiles(t, e.root), own) {
		t.Fatalf("a refused run stored its output %s", own)
	}
}
