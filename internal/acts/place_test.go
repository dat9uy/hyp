package acts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/example"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
	"github.com/dat9uy/hyp/internal/testutil"
)

// fixtureProbe is the Probe of the Probe contract in hyp spec.
const fixtureProbe = `#!/bin/sh
o=$(mktemp); sh -c "$HYP_RUN" >"$o"; c=$?; cat "$o"; [ -s "$o" ] && [ -n "$(tail -c1 "$o")" ] && echo; rm -f "$o"; echo "exit $c"
`

// The run prints "hi" on H0: a floor "nope" fails there, a floor "hi" is met.
const (
	failingOnH0 = "-- words --\nw\n-- run --\necho hi\n-- floor: contains --\nnope\n"
	metOnH0     = "-- words --\nw\n-- run --\necho hi\n-- floor: contains --\nhi\n"
)

type placeEnv struct {
	r    *testutil.Repo
	goal string // the Goal id
	root string
}

// newPlaceEnv commits a subject, adds a Goal with lines on it as the owner
// and returns the environment.
func newPlaceEnv(t *testing.T, lines ...LineInput) *placeEnv {
	t.Helper()
	r := testutil.NewRepo(t)
	base := r.Commit(t, map[string]string{"main.go": "package main\n"})
	if len(lines) == 0 {
		lines = []LineInput{gapLine}
	}
	rep, err := AddGoal(opts(r.Dir, base, lines...))
	if err != nil {
		t.Fatal(err)
	}
	return &placeEnv{r: r, goal: rep.Goal, root: r.Dir}
}

// write puts an example folder on disk, uncommitted.
func (e *placeEnv) write(t *testing.T, key, txtar, probe string) {
	t.Helper()
	dir := filepath.Join("goals", e.goal, key)
	if err := os.MkdirAll(filepath.Join(e.root, dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(e.root, dir, "example.txtar"), []byte(txtar), 0o644); err != nil {
		t.Fatal(err)
	}
	e.r.WriteExec(t, filepath.Join(dir, "probe", "run"), probe)
}

func (e *placeEnv) commit(t *testing.T) {
	t.Helper()
	testutil.Git(t, e.root, "add", "-A")
	testutil.Git(t, e.root, "commit", "-q", "--allow-empty", "-m", "example")
}

// example writes and commits an example folder.
func (e *placeEnv) example(t *testing.T, key, txtar string) {
	t.Helper()
	e.write(t, key, txtar, fixtureProbe)
	e.commit(t)
}

func (e *placeEnv) opts() PlaceOpts {
	return PlaceOpts{Dir: e.root, Actor: "grill-1", Role: "grill", Goal: e.goal, ProbeTimeout: 20 * time.Second}
}

// place places the Goal's whole example set.
func (e *placeEnv) place() (PlaceReport, error) { return Place(e.opts()) }

// log returns the store's events and the tables they replay to.
func (e *placeEnv) log(t *testing.T) ([]events.Event, tables.Tables) {
	t.Helper()
	evs, errs, err := logstore.Load(e.root)
	if err != nil || len(errs) != 0 {
		t.Fatalf("Events: %v %v", errs, err)
	}
	return evs, tables.Replay(evs)
}

// placements lists the store's place events.
func (e *placeEnv) placements(t *testing.T) []events.Event {
	t.Helper()
	evs, _ := e.log(t)
	var out []events.Event
	for _, ev := range evs {
		if ev.Type == events.TypeExamplesPlace || ev.Type == events.TypeExamplePlace {
			out = append(out, ev)
		}
	}
	return out
}

func refusedBy(t *testing.T, err error, guard string) {
	t.Helper()
	if err == nil || fault.KindOf(err) != fault.Refusal || string(fault.GuardOf(err)) != guard {
		t.Fatalf("err = %v, want a %s refusal", err, guard)
	}
}

// causes returns the causes of a refusal that placed none of the set, by key.
func causes(t *testing.T, err error) map[string]error {
	t.Helper()
	e, ok := fault.As(err)
	if !ok || e.Code != fault.SetNotPlaced || len(e.Args) != 1 {
		t.Fatalf("err = %v, want a refusal that places none of the set", err)
	}
	f, ok := e.Args[0].(SetFacts)
	if !ok {
		t.Fatalf("facts = %#v", e.Args[0])
	}
	out := map[string]error{}
	for _, c := range f.Causes {
		out[c.Key] = c.Err
	}
	return out
}

// causedBy checks that the cause of key exists and was refused by guard.
func causedBy(t *testing.T, cs map[string]error, key, guard string) error {
	t.Helper()
	err, ok := cs[key]
	if !ok {
		t.Fatalf("causes = %v, want one for %s", cs, key)
	}
	if string(fault.GuardOf(err)) != guard {
		t.Fatalf("cause of %s = %v, want a %s refusal", key, err, guard)
	}
	return err
}

// payload decodes a place event.
func payload(t *testing.T, ev events.Event) events.ExamplesPlace {
	t.Helper()
	var p events.ExamplesPlace
	if err := events.Decode(ev, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPlaceGapExampleFailingOnH0Records(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	rep, err := e.place()
	if err != nil {
		t.Fatal(err)
	}
	hash := testutil.Git(t, e.root, "rev-parse", "HEAD:goals/"+e.goal+"/line-1/a")
	if rep.Goal != e.goal || len(rep.Placed) != 1 || rep.Placed[0] != (PlacedRow{Key: "line-1/a", Version: 1, Hash: hash, Verdict: "failed", Exit: 0}) {
		t.Fatalf("report = %+v, want version 1 of %s failing on H0", rep, hash)
	}
	pl := e.placements(t)
	if len(pl) != 1 || pl[0].Type != events.TypeExamplesPlace {
		t.Fatalf("place events = %+v, want one examples.place", pl)
	}
	p := payload(t, pl[0])
	evs, tb := e.log(t)
	g := tb.ByID(e.goal)
	if pl[0].Node != g.EventID || pl[0].Actor != "grill-1" || p.Role != "grill" || len(p.Examples) != 1 {
		t.Fatalf("event = %+v payload = %+v", pl[0], p)
	}
	x := p.Examples[0]
	if x.Line != "line-1" || x.Example != "a" || x.Hash != hash {
		t.Fatalf("example = %+v", x)
	}
	probeHash := testutil.Git(t, e.root, "rev-parse", "HEAD:goals/"+e.goal+"/line-1/a/probe")
	if x.H0.Commit != g.Base || x.H0.RowTree != g.BaseTree || x.H0.Probe != probeHash || x.H0.Verdict != "failed" || x.H0.Exit != 0 {
		t.Fatalf("H0 cell = %+v, want commit %s row tree %s probe %s", x.H0, g.Base, g.BaseTree, probeHash)
	}
	// git holds the hash and H0's verdict; H0's outputs are in the trace
	if x.H0.Output != "" || x.H0.Stderr != "" {
		t.Fatalf("the place names an output in git: %+v", x.H0)
	}
	if _, err := os.Stat(filepath.Join(e.root, ".tree", "outputs")); !os.IsNotExist(err) {
		t.Fatalf(".tree/outputs: %v, want none", err)
	}
	tr, err := logstore.TraceOf(e.root)
	if err != nil {
		t.Fatal(err)
	}
	cells, err := tr.Cells()
	if err != nil {
		t.Fatal(err)
	}
	tc := cells[logstore.CellKey(hash, g.BaseTree)]
	if out, err := os.ReadFile(logstore.OutputPath(e.root, tc.Output)); err != nil || string(out) != "hi" {
		t.Fatalf("output file = %q, %v; want hi", out, err)
	}
	if out, err := os.ReadFile(logstore.OutputPath(e.root, tc.Stderr)); err != nil || len(out) != 0 {
		t.Fatalf("stderr file = %q, %v; want empty", out, err)
	}
	if len(tb.RefusedAtReplay) != 0 || len(evs) != 2 {
		t.Fatalf("log = %d events, refused at replay %+v", len(evs), tb.RefusedAtReplay)
	}
}

// TestPlacePlacesTheWholeSetInOneEvent: every example of every line goes
// into one event, and the Goal opens with it.
func TestPlacePlacesTheWholeSetInOneEvent(t *testing.T) {
	e := newPlaceEnv(t, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	e.write(t, "line-1/a", failingOnH0, fixtureProbe)
	e.write(t, "line-1/b", failingOnH0+"-- ceiling --\nzzz\n", fixtureProbe)
	e.write(t, "line-2/a", metOnH0, fixtureProbe)
	e.commit(t)
	rep, err := e.place()
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, p := range rep.Placed {
		keys = append(keys, p.Key)
	}
	if !slices.Equal(keys, []string{"line-1/a", "line-1/b", "line-2/a"}) || rep.State != tables.StateOpen {
		t.Fatalf("report = %+v, want three examples and the Goal open", rep)
	}
	if pl := e.placements(t); len(pl) != 1 || len(payload(t, pl[0]).Examples) != 3 {
		t.Fatalf("place events = %+v, want one holding three examples", pl)
	}
	if _, tb := e.log(t); tb.ByID(e.goal).State != tables.StateOpen {
		t.Fatalf("replayed state = %s, want open", tb.ByID(e.goal).State)
	}
	// an open Goal is frozen
	_, err = AddGoal(AddGoalOpts{Dir: e.root, Actor: "owner-1", Role: "owner", Goal: e.goal, Words: "new", Base: "HEAD", Lines: []LineInput{gapLine}})
	refusedBy(t, err, "Draft")
}

func TestPlaceRefusesWholeSetWhenOneExampleDisagrees(t *testing.T) {
	e := newPlaceEnv(t, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	e.write(t, "line-1/a", failingOnH0, fixtureProbe)
	e.write(t, "line-1/b", metOnH0, fixtureProbe) // a gap example met on H0
	e.write(t, "line-2/a", metOnH0, fixtureProbe)
	e.write(t, "line-2/b", failingOnH0, fixtureProbe) // a guard example failing on H0
	e.commit(t)
	_, err := e.place()
	refusedBy(t, err, "Kind")
	cs := causes(t, err)
	if len(cs) != 2 {
		t.Fatalf("causes = %v, want line-1/b and line-2/b", cs)
	}
	causedBy(t, cs, "line-1/b", "Kind")
	causedBy(t, cs, "line-2/b", "Kind")
	if n := len(e.placements(t)); n != 0 {
		t.Fatalf("%d place events, want none: the set places whole or not at all", n)
	}
	if _, tb := e.log(t); tb.ByID(e.goal).State != tables.StateWritten {
		t.Fatalf("state = %s, want written", tb.ByID(e.goal).State)
	}
}

func TestPlaceKindRefusalShowsH0Output(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", metOnH0+"-- floor: equals --\nhi\n-- ceiling --\nzzz\n")
	_, err := e.place()
	refusedBy(t, err, "Kind")
	// the cause carries H0's output beside the floors and ceilings; the
	// channel prints them
	e2, _ := fault.As(causedBy(t, causes(t, err), "line-1/a", "Kind"))
	f, ok := e2.Args[0].(KindFacts)
	if !ok || e2.Code != fault.WrongVerdictOnH0 || f.Line.Key != "line-1" || f.Verdict != "met" ||
		string(f.Out.Stdout) != "hi" || f.Out.Exit != 0 || len(f.Floors) != 2 ||
		f.Floors[1] != (example.Floor{Op: example.OpEquals, Text: "hi"}) || len(f.Ceilings) != 1 || f.Ceilings[0] != "zzz" {
		t.Fatalf("Kind refusal = %#v", e2)
	}
}

func TestPlaceRefusesUncommittedFolder(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, fixtureProbe)
	_, err := e.place()
	refusedBy(t, err, "Committed")
	if err := causedBy(t, causes(t, err), "line-1/a", "Committed"); !testutil.Mentions(err, "line-1/a") {
		t.Fatalf("err = %v, want it to name the folder", err)
	}
	e.commit(t)
	if _, err := e.place(); err != nil {
		t.Fatalf("a committed folder was refused: %v", err)
	}
	// an edit after the commit is not the committed folder
	e.write(t, "line-1/b", failingOnH0, fixtureProbe)
	e.commit(t)
	if err := os.WriteFile(filepath.Join(e.root, "goals", e.goal, "line-1/b/example.txtar"), []byte(failingOnH0+"-- ceiling --\nx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = e.place()
	if err := causedBy(t, causes(t, err), "line-1/b", "Committed"); !testutil.Mentions(err, "example.txtar") {
		t.Fatalf("err = %v, want it to name the first differing path, example.txtar", err)
	}
	// a stray file is a difference too
	testutil.Git(t, e.root, "checkout", "--", ".")
	if err := os.WriteFile(filepath.Join(e.root, "goals", e.goal, "line-1/b/stray"), []byte("s"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = e.place()
	if err := causedBy(t, causes(t, err), "line-1/b", "Committed"); !testutil.Mentions(err, "stray") {
		t.Fatalf("err = %v, want it to name stray", err)
	}
	// a folder committed and then removed from disk is still in the set
	if err := os.Remove(filepath.Join(e.root, "goals", e.goal, "line-1/b/stray")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(e.root, "goals", e.goal, "line-1/b")); err != nil {
		t.Fatal(err)
	}
	_, err = e.place()
	causedBy(t, causes(t, err), "line-1/b", "Committed")
	if n := len(e.placements(t)); n != 1 {
		t.Fatalf("%d placements after refusals, want only the first", n)
	}
}

func TestUncommittedPlaceRunsNoProbe(t *testing.T) {
	e := newPlaceEnv(t)
	marker := filepath.Join(t.TempDir(), "ran")
	probe := strings.Replace(fixtureProbe, "#!/bin/sh\n", "#!/bin/sh\ntouch "+marker+"\n", 1)
	e.write(t, "line-1/a", failingOnH0, probe)
	e.commit(t)
	e.write(t, "line-1/b", failingOnH0, probe) // never committed
	_, err := e.place()
	refusedBy(t, err, "Committed")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("a refused place ran a Probe: %v", err)
	}
}

func TestPlaceRefusesLineWithNoExample(t *testing.T) {
	e := newPlaceEnv(t, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	e.example(t, "line-1/a", failingOnH0)
	_, err := e.place()
	refusedBy(t, err, "Whole set")
	cs := causes(t, err)
	if len(cs) != 1 {
		t.Fatalf("causes = %v, want only line-2", cs)
	}
	causedBy(t, cs, "line-2", "Whole set")
	if n := len(e.placements(t)); n != 0 {
		t.Fatalf("%d place events, want none", n)
	}
}

func TestPlaceRefusesPlacedExampleGone(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, fixtureProbe)
	e.write(t, "line-1/b", failingOnH0, fixtureProbe)
	e.commit(t)
	if _, err := e.place(); err != nil {
		t.Fatal(err)
	}
	testutil.Git(t, e.root, "rm", "-q", "-r", "goals/"+e.goal+"/line-1/b")
	e.example(t, "line-1/c", failingOnH0)
	_, err := e.place()
	refusedBy(t, err, "Whole set")
	cs := causes(t, err)
	if len(cs) != 1 {
		t.Fatalf("causes = %v, want only line-1/b", cs)
	}
	causedBy(t, cs, "line-1/b", "Whole set")
	if n := len(e.placements(t)); n != 1 {
		t.Fatalf("%d place events, want the first only", n)
	}
}

func TestPlaceRefusesFolderThatIsNoExampleName(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, fixtureProbe)
	e.write(t, "line-1/Bad", failingOnH0, fixtureProbe)
	e.commit(t)
	_, err := e.place()
	causedBy(t, causes(t, err), "line-1/Bad", "Whole set")
}

func TestPlaceNamesEveryCauseBeforeTheProbes(t *testing.T) {
	e := newPlaceEnv(t, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	e.write(t, "line-1/run", failingOnH0, fixtureProbe)
	e.write(t, "line-2/a", metOnH0, fixtureProbe)
	e.commit(t)
	if _, err := e.place(); err != nil {
		t.Fatal(err)
	}
	e.example(t, "line-1/run", strings.Replace(failingOnH0, "echo hi", "echo ho", 1)+"-- ceiling --\nzzz\n")
	e.write(t, "line-2/new", metOnH0, fixtureProbe) // not committed
	_, err := e.place()
	cs := causes(t, err)
	if len(cs) != 2 {
		t.Fatalf("causes = %v, want two", cs)
	}
	causedBy(t, cs, "line-1/run", "Adds only")
	causedBy(t, cs, "line-2/new", "Committed")
}

func TestPlaceRefusesRunnerRole(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, fixtureProbe) // uncommitted: Grill places comes first
	for _, role := range []string{"runner", "owner", "probe-writer"} {
		o := e.opts()
		o.Role = role
		_, err := Place(o)
		refusedBy(t, err, "Grill places")
	}
	if n := len(e.placements(t)); n != 0 {
		t.Fatalf("%d placements after Grill places refusals", n)
	}
}

func TestPlaceRefusesNewVersionDroppingCeiling(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0+"-- ceiling --\nzzz\n")
	if _, err := e.place(); err != nil {
		t.Fatal(err)
	}
	e.example(t, "line-1/a", failingOnH0+"-- floor: contains --\nmore\n")
	_, err := e.place()
	refusedBy(t, err, "Adds only")
	if err := causedBy(t, causes(t, err), "line-1/a", "Adds only"); !testutil.Mentions(err, "ceiling") {
		t.Fatalf("err = %v, want it to name the dropped ceiling", err)
	}
	if n := len(e.placements(t)); n != 1 {
		t.Fatalf("%d placements, want 1", n)
	}
}

func TestPlaceRefusesNewVersionChangingProbe(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	if _, err := e.place(); err != nil {
		t.Fatal(err)
	}
	e.write(t, "line-1/a", failingOnH0+"-- ceiling --\nzzz\n", fixtureProbe+"# changed\n")
	e.commit(t)
	_, err := e.place()
	if err := causedBy(t, causes(t, err), "line-1/a", "Adds only"); !testutil.Mentions(err, "probe/") {
		t.Fatalf("err = %v, want it to name probe/", err)
	}
}

func TestPlaceRefusesNewVersionChangingRun(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	if _, err := e.place(); err != nil {
		t.Fatal(err)
	}
	e.example(t, "line-1/a", strings.Replace(failingOnH0, "echo hi", "echo ho", 1)+"-- ceiling --\nzzz\n")
	_, err := e.place()
	if err := causedBy(t, causes(t, err), "line-1/a", "Adds only"); !testutil.Mentions(err, "run") {
		t.Fatalf("err = %v, want it to name the run section", err)
	}
}

func TestPlaceAddingAFloorIsTheNextVersion(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, fixtureProbe)
	e.write(t, "line-1/b", failingOnH0, fixtureProbe)
	e.commit(t)
	first, err := e.place()
	if err != nil {
		t.Fatal(err)
	}
	e.example(t, "line-1/a", failingOnH0+"-- ceiling --\nzzz\n")
	second, err := e.place()
	if err != nil {
		t.Fatal(err)
	}
	// only the changed example is in the second place
	if len(first.Placed) != 2 || len(second.Placed) != 1 || second.Placed[0].Key != "line-1/a" ||
		second.Placed[0].Version != 2 || second.Placed[0].Hash == first.Placed[0].Hash {
		t.Fatalf("first %+v, second %+v", first, second)
	}
	_, tb := e.log(t)
	cur, _ := tb.ByID(e.goal).Current("line-1/a")
	if cur.Hash != second.Placed[0].Hash || cur.Version != 2 {
		t.Fatalf("current = %+v, want the second version", cur)
	}
	if b, _ := tb.ByID(e.goal).Current("line-1/b"); b.Version != 1 {
		t.Fatalf("line-1/b = %+v, want version 1 kept", b)
	}
}

// TestPlaceRefusesUnchangedSet: a place that changes no example is refused,
// a change under specs/ alone included, and the budget stays.
func TestPlaceRefusesUnchangedSet(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	if _, err := e.place(); err != nil {
		t.Fatal(err)
	}
	_, err := e.place()
	refusedBy(t, err, "Adds only")
	if fault.CodeOf(err) != fault.SetUnchanged {
		t.Fatalf("err = %v, want the unchanged-set refusal", err)
	}
	e.r.Commit(t, map[string]string{"goals/" + e.goal + "/specs/note.txt": "note\n"})
	_, err = e.place()
	if fault.CodeOf(err) != fault.SetUnchanged {
		t.Fatalf("after a specs change: err = %v, want the unchanged-set refusal", err)
	}
	if n := len(e.placements(t)); n != 1 {
		t.Fatalf("%d placements, want 1", n)
	}
}

func TestPlaceBadArgumentsAreExit2(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	for name, mod := range map[string]func(*PlaceOpts){
		"unknown goal": func(o *PlaceOpts) { o.Goal = "G-NOPE00" },
		"bad actor":    func(o *PlaceOpts) { o.Actor = "Not Valid" },
		"bad role":     func(o *PlaceOpts) { o.Role = "king" },
		"long why":     func(o *PlaceOpts) { o.Why = strings.Repeat("x", 601) },
	} {
		o := e.opts()
		mod(&o)
		if _, err := Place(o); err == nil || fault.KindOf(err) != fault.BadArguments {
			t.Errorf("%s: err = %v, want bad arguments", name, err)
		}
	}
	if n := len(e.placements(t)); n != 0 {
		t.Fatalf("%d placements after bad arguments", n)
	}
}

func TestPlaceWithoutStoreIsRefusal(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	_, err := Place(PlaceOpts{Dir: r.Dir, Actor: "grill-1", Role: "grill", Goal: "G-AAAAAA"})
	if fault.KindOf(err) != fault.Refusal || fault.CodeOf(err) != fault.NoStore {
		t.Fatalf("err = %v, want the no-store refusal", err)
	}
}

func TestPlaceProbeFailureRecordsNothing(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, fixtureProbe)
	e.write(t, "line-1/b", failingOnH0, "#!/bin/sh\nexit 7\n")
	e.commit(t)
	_, err := e.place()
	if err == nil || fault.KindOf(err) != fault.Environment {
		t.Fatalf("err = %v, want an environment error", err)
	}
	if n := len(e.placements(t)); n != 0 {
		t.Fatalf("%d placements after a failed Probe", n)
	}
}

func TestPlaceChecksAgainAfterTheProbe(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	// another actor places the same set while these Probes run
	o := e.opts()
	o.afterProbe = func() {
		other := e.opts()
		other.Actor = "grill-2"
		if _, err := Place(other); err != nil {
			t.Errorf("the place that lands between the checks was refused: %v", err)
		}
	}
	_, err := Place(o)
	refusedBy(t, err, "Adds only")
	pl := e.placements(t)
	if len(pl) != 1 || pl[0].Actor != "grill-2" {
		t.Fatalf("placements = %+v, want only the one that landed between the checks", pl)
	}
}

func TestPlaceRefusesWhenSetChangesDuringTheProbe(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	o := e.opts()
	o.afterProbe = func() { e.example(t, "line-1/b", failingOnH0) }
	_, err := Place(o)
	refusedBy(t, err, "Committed")
	if fault.CodeOf(err) != fault.SetChangedWhileProbe {
		t.Fatalf("err = %v, want it to say the set changed", err)
	}
	if n := len(e.placements(t)); n != 0 {
		t.Fatalf("%d placements recorded for a set that changed", n)
	}
}

func TestPlaceRefusesWhenH0MovesDuringTheProbe(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	o := e.opts()
	o.afterProbe = func() {
		// the owner replaces the written Goal with another base
		_, err := AddGoal(AddGoalOpts{Dir: e.root, Actor: "owner-1", Role: "owner", Goal: e.goal, Words: "moved", Base: "HEAD", Lines: []LineInput{gapLine}})
		if err != nil {
			t.Errorf("replace between the checks: %v", err)
		}
	}
	_, err := Place(o)
	refusedBy(t, err, "Kind")
	if fault.CodeOf(err) != fault.H0MovedWhilePlacing {
		t.Fatalf("err = %v, want it to say H0 moved", err)
	}
	if n := len(e.placements(t)); n != 0 {
		t.Fatalf("%d placements recorded on a stale H0", n)
	}
}

func TestPlaceWritesNoGitState(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	before := testutil.Snapshot(t, e.root)
	if _, err := e.place(); err != nil {
		t.Fatalf("place: %v; it must have recorded", err)
	}
	if testutil.Snapshot(t, e.root) != before {
		t.Fatal("place changed the repository's git state")
	}
	// a refused place changes nothing either
	_, err := e.place()
	refusedBy(t, err, "Adds only")
	if testutil.Snapshot(t, e.root) != before {
		t.Fatal("a refused place changed the repository's git state")
	}
}

func TestPlaceRefusesNewVersionChangingAnotherFile(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	helper := filepath.Join(e.root, "goals", e.goal, "line-1/a/helper")
	if err := os.WriteFile(helper, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.commit(t)
	if _, err := e.place(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(helper, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.write(t, "line-1/a", failingOnH0+"-- ceiling --\nzzz\n", fixtureProbe)
	e.commit(t)
	_, err := e.place()
	if err := causedBy(t, causes(t, err), "line-1/a", "Adds only"); !testutil.Mentions(err, "helper") {
		t.Fatalf("err = %v, want it to name helper", err)
	}
	// the same folder with only the ceiling added passes
	if err := os.WriteFile(helper, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.commit(t)
	if rep, err := e.place(); err != nil || rep.Placed[0].Version != 2 {
		t.Fatalf("place = %+v, %v; want version 2", rep, err)
	}
}

func TestPlaceCurrentVersionMissingFromCloneIsEnvironment(t *testing.T) {
	e := newPlaceEnv(t)
	_, tb := e.log(t)
	g := tb.ByID(e.goal)
	payload, err := json.Marshal(events.ExamplePlace{
		Role: "grill", Line: "line-1", Example: "a", Hash: strings.Repeat("ab", 20),
		H0: events.Cell{Commit: g.Base, RowTree: g.BaseTree, Verdict: "failed"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = logstore.Append(e.root, "grill-2", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeExamplePlace, Node: g.EventID, Payload: payload}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	e.example(t, "line-1/a", failingOnH0)
	_, err = e.place()
	if err == nil || fault.KindOf(err) != fault.Environment || !testutil.Mentions(err, strings.Repeat("ab", 20)) {
		t.Fatalf("err = %v, want an environment error naming the missing object", err)
	}
}

// outputFiles lists the files of the store's outputs folder.
func outputFiles(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(logstore.OutputsDir(root))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// TestKindRefusalStoresNoOutput: a place the Kind guard refuses, after its
// Probes ran, leaves no file in the store: the outputs are stored only with
// the event that names them.
func TestKindRefusalStoresNoOutput(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, fixtureProbe)
	e.write(t, "line-1/b", metOnH0, fixtureProbe)
	e.commit(t)
	before := outputFiles(t, e.root)
	_, err := e.place()
	refusedBy(t, err, "Kind")
	if after := outputFiles(t, e.root); !slices.Equal(after, before) {
		t.Fatalf("outputs after a Kind refusal = %v, were %v", after, before)
	}
}
