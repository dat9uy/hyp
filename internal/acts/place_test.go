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

// fixtureProbe is the Probe of the state machine, section 10.
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

func (e *placeEnv) opts(key string) PlaceOpts {
	return PlaceOpts{Dir: e.root, Actor: "grill-1", Role: "grill", Goal: e.goal, Key: key, ProbeTimeout: 20 * time.Second}
}

func (e *placeEnv) place(key string) (PlaceReport, error) { return Place(e.opts(key)) }

// log returns the store's events and the tables they replay to.
func (e *placeEnv) log(t *testing.T) ([]events.Event, tables.Tables) {
	t.Helper()
	evs, errs, err := logstore.Events(e.root)
	if err != nil || len(errs) != 0 {
		t.Fatalf("Events: %v %v", errs, err)
	}
	return evs, tables.Replay(evs)
}

func (e *placeEnv) placements(t *testing.T) []events.Event {
	t.Helper()
	evs, _ := e.log(t)
	var out []events.Event
	for _, ev := range evs {
		if ev.Type == events.TypeExamplePlace {
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

func TestPlaceGapExampleFailingOnH0Records(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	rep, err := e.place("line-1/a")
	if err != nil {
		t.Fatal(err)
	}
	hash := testutil.Git(t, e.root, "rev-parse", "HEAD:goals/"+e.goal+"/line-1/a")
	if rep.Goal != e.goal || rep.Key != "line-1/a" || rep.Version != 1 || rep.Hash != hash || rep.Verdict != "failed" || rep.Exit != 0 {
		t.Fatalf("report = %+v, want version 1 of %s failing on H0", rep, hash)
	}
	pl := e.placements(t)
	if len(pl) != 1 {
		t.Fatalf("%d example.place events, want 1", len(pl))
	}
	var p events.ExamplePlace
	if err := events.Decode(pl[0], &p); err != nil {
		t.Fatal(err)
	}
	evs, tb := e.log(t)
	g := tb.ByID(e.goal)
	if pl[0].Node != g.EventID || pl[0].Actor != "grill-1" || p.Role != "grill" || p.Line != "line-1" || p.Example != "a" || p.Hash != hash {
		t.Fatalf("event = %+v payload = %+v", pl[0], p)
	}
	probeHash := testutil.Git(t, e.root, "rev-parse", "HEAD:goals/"+e.goal+"/line-1/a/probe")
	if p.H0.Commit != g.Base || p.H0.RowTree != g.BaseTree || p.H0.Probe != probeHash || p.H0.Verdict != "failed" || p.H0.Exit != 0 {
		t.Fatalf("H0 cell = %+v, want commit %s row tree %s probe %s", p.H0, g.Base, g.BaseTree, probeHash)
	}
	if out, err := os.ReadFile(logstore.OutputPath(e.root, p.H0.Output)); err != nil || string(out) != "hi" {
		t.Fatalf("output file = %q, %v; want hi", out, err)
	}
	if out, err := os.ReadFile(logstore.OutputPath(e.root, p.H0.Stderr)); err != nil || len(out) != 0 {
		t.Fatalf("stderr file = %q, %v; want empty", out, err)
	}
	if len(tb.RefusedAtReplay) != 0 || len(evs) != 2 {
		t.Fatalf("log = %d events, refused at replay %+v", len(evs), tb.RefusedAtReplay)
	}
}

func TestPlaceRefusesUncommittedFolder(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, fixtureProbe)
	_, err := e.place("line-1/a")
	refusedBy(t, err, "Committed")
	if !testutil.Mentions(err, "line-1/a") {
		t.Fatalf("err = %v, want it to name the folder", err)
	}
	e.commit(t)
	if _, err := e.place("line-1/a"); err != nil {
		t.Fatalf("a committed folder was refused: %v", err)
	}
	// an edit after the commit is not the committed folder
	e.write(t, "line-1/b", failingOnH0, fixtureProbe)
	e.commit(t)
	if err := os.WriteFile(filepath.Join(e.root, "goals", e.goal, "line-1/b/example.txtar"), []byte(failingOnH0+"-- ceiling --\nx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = e.place("line-1/b")
	refusedBy(t, err, "Committed")
	if !testutil.Mentions(err, "example.txtar") {
		t.Fatalf("err = %v, want it to name the first differing path, example.txtar", err)
	}
	// a stray file is a difference too
	e.write(t, "line-1/c", failingOnH0, fixtureProbe)
	e.commit(t)
	if err := os.WriteFile(filepath.Join(e.root, "goals", e.goal, "line-1/c/stray"), []byte("s"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = e.place("line-1/c")
	refusedBy(t, err, "Committed")
	if !testutil.Mentions(err, "stray") {
		t.Fatalf("err = %v, want it to name stray", err)
	}
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
	if _, err := e.place("line-1/a"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("the committed example's Probe left no marker: %v", err)
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	e.write(t, "line-1/b", failingOnH0, probe) // never committed
	_, err := e.place("line-1/b")
	refusedBy(t, err, "Committed")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("a refused placement ran the Probe: %v", err)
	}
}

func TestPlaceRefusesRunnerRole(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, fixtureProbe) // uncommitted: Grill places comes first
	for _, role := range []string{"runner", "owner", "probe-writer"} {
		o := e.opts("line-1/a")
		o.Role = role
		_, err := Place(o)
		refusedBy(t, err, "Grill places")
	}
	if n := len(e.placements(t)); n != 0 {
		t.Fatalf("%d placements after Grill places refusals", n)
	}
}

func TestPlaceRefusesGuardExampleFailingOnH0(t *testing.T) {
	e := newPlaceEnv(t, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	e.example(t, "line-2/a", failingOnH0)
	_, err := e.place("line-2/a")
	refusedBy(t, err, "Kind")
	e.example(t, "line-2/b", metOnH0)
	rep, err := e.place("line-2/b")
	if err != nil || rep.Verdict != "met" {
		t.Fatalf("a guard example met on H0: %+v, %v", rep, err)
	}
	// and a gap example met on H0 is refused
	e.example(t, "line-1/a", metOnH0)
	_, err = e.place("line-1/a")
	refusedBy(t, err, "Kind")
	if n := len(e.placements(t)); n != 1 {
		t.Fatalf("%d placements, want only the met guard example", n)
	}
}

func TestPlaceKindRefusalShowsH0Output(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", metOnH0+"-- floor: equals --\nhi\n-- ceiling --\nzzz\n")
	_, err := e.place("line-1/a")
	refusedBy(t, err, "Kind")
	// the refusal carries H0's output beside the floors and ceilings; the
	// channel prints them
	e2, _ := fault.As(err)
	f, ok := e2.Args[0].(KindFacts)
	if !ok || e2.Code != fault.WrongVerdictOnH0 || f.Line.Key != "line-1" || f.Verdict != "met" ||
		string(f.Out.Stdout) != "hi" || f.Out.Exit != 0 || len(f.Floors) != 2 ||
		f.Floors[1] != (example.Floor{Op: example.OpEquals, Text: "hi"}) || len(f.Ceilings) != 1 || f.Ceilings[0] != "zzz" {
		t.Fatalf("Kind refusal = %#v", e2)
	}
}

func TestPlaceRefusesNewVersionDroppingCeiling(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0+"-- ceiling --\nzzz\n")
	if _, err := e.place("line-1/a"); err != nil {
		t.Fatal(err)
	}
	e.example(t, "line-1/a", failingOnH0+"-- floor: contains --\nmore\n")
	_, err := e.place("line-1/a")
	refusedBy(t, err, "Adds only")
	if !testutil.Mentions(err, "ceiling") {
		t.Fatalf("err = %v, want it to name the dropped ceiling", err)
	}
	if n := len(e.placements(t)); n != 1 {
		t.Fatalf("%d placements, want 1", n)
	}
}

func TestPlaceRefusesNewVersionChangingProbe(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	if _, err := e.place("line-1/a"); err != nil {
		t.Fatal(err)
	}
	e.write(t, "line-1/a", failingOnH0+"-- ceiling --\nzzz\n", fixtureProbe+"# changed\n")
	e.commit(t)
	_, err := e.place("line-1/a")
	refusedBy(t, err, "Adds only")
	if !testutil.Mentions(err, "probe/") {
		t.Fatalf("err = %v, want it to name probe/", err)
	}
}

func TestPlaceRefusesNewVersionChangingRun(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	if _, err := e.place("line-1/a"); err != nil {
		t.Fatal(err)
	}
	e.example(t, "line-1/a", strings.Replace(failingOnH0, "echo hi", "echo ho", 1)+"-- ceiling --\nzzz\n")
	_, err := e.place("line-1/a")
	refusedBy(t, err, "Adds only")
	if !testutil.Mentions(err, "run") {
		t.Fatalf("err = %v, want it to name the run section", err)
	}
}

func TestPlaceAddingAFloorIsTheNextVersion(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	first, err := e.place("line-1/a")
	if err != nil {
		t.Fatal(err)
	}
	e.example(t, "line-1/a", failingOnH0+"-- ceiling --\nzzz\n")
	second, err := e.place("line-1/a")
	if err != nil {
		t.Fatal(err)
	}
	if first.Version != 1 || second.Version != 2 || second.Hash == first.Hash {
		t.Fatalf("versions %d then %d, hashes %s then %s", first.Version, second.Version, first.Hash, second.Hash)
	}
	_, tb := e.log(t)
	cur, _ := tb.ByID(e.goal).Current("line-1/a")
	if cur.Hash != second.Hash || cur.Version != 2 {
		t.Fatalf("current = %+v, want the second version", cur)
	}
}

func TestPlaceRefusesRepeatOfCurrentVersion(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	if _, err := e.place("line-1/a"); err != nil {
		t.Fatal(err)
	}
	_, err := e.place("line-1/a")
	refusedBy(t, err, "Adds only")
	if n := len(e.placements(t)); n != 1 {
		t.Fatalf("%d placements, want 1", n)
	}
}

func TestPlaceOpensGoal(t *testing.T) {
	e := newPlaceEnv(t, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	e.example(t, "line-1/a", failingOnH0)
	rep, err := e.place("line-1/a")
	if err != nil || rep.State != tables.StateWritten {
		t.Fatalf("after line-1: %+v, %v; want written", rep, err)
	}
	e.example(t, "line-2/a", metOnH0)
	rep, err = e.place("line-2/a")
	if err != nil || rep.State != tables.StateOpen {
		t.Fatalf("after line-2: %+v, %v; want open", rep, err)
	}
	if _, tb := e.log(t); tb.ByID(e.goal).State != tables.StateOpen {
		t.Fatalf("replayed state = %s, want open", tb.ByID(e.goal).State)
	}
	// an open Goal is frozen
	_, err = AddGoal(AddGoalOpts{Dir: e.root, Actor: "owner-1", Role: "owner", Goal: e.goal, Words: "new", Base: "HEAD", Lines: []LineInput{gapLine}})
	refusedBy(t, err, "Draft")
}

func TestPlaceBadArgumentsAreExit2(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	for name, mod := range map[string]func(*PlaceOpts){
		"unknown goal":   func(o *PlaceOpts) { o.Goal = "G-NOPE00" },
		"unknown line":   func(o *PlaceOpts) { o.Key = "line-9/a" },
		"no slash":       func(o *PlaceOpts) { o.Key = "line-1" },
		"bad example":    func(o *PlaceOpts) { o.Key = "line-1/A" },
		"dash example":   func(o *PlaceOpts) { o.Key = "line-1/-a" },
		"nested example": func(o *PlaceOpts) { o.Key = "line-1/a/b" },
		"bad actor":      func(o *PlaceOpts) { o.Actor = "Not Valid" },
		"bad role":       func(o *PlaceOpts) { o.Role = "king" },
		"long why":       func(o *PlaceOpts) { o.Why = strings.Repeat("x", 601) },
	} {
		o := e.opts("line-1/a")
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
	_, err := Place(PlaceOpts{Dir: r.Dir, Actor: "grill-1", Role: "grill", Goal: "G-AAAAAA", Key: "line-1/a"})
	if fault.KindOf(err) != fault.Refusal || fault.CodeOf(err) != fault.NoStore {
		t.Fatalf("err = %v, want the no-store refusal", err)
	}
}

func TestPlaceProbeFailureRecordsNothing(t *testing.T) {
	e := newPlaceEnv(t)
	e.write(t, "line-1/a", failingOnH0, "#!/bin/sh\nexit 7\n")
	e.commit(t)
	_, err := e.place("line-1/a")
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
	// another actor places the same version while this Probe runs
	o := e.opts("line-1/a")
	o.afterProbe = func() {
		other := e.opts("line-1/a")
		other.Actor = "grill-2"
		if _, err := Place(other); err != nil {
			t.Errorf("the placement that lands between the checks was refused: %v", err)
		}
	}
	_, err := Place(o)
	refusedBy(t, err, "Adds only")
	pl := e.placements(t)
	if len(pl) != 1 || pl[0].Actor != "grill-2" {
		t.Fatalf("placements = %+v, want only the one that landed between the checks", pl)
	}
}

func TestPlaceRefusesWhenH0MovesDuringTheProbe(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", failingOnH0)
	o := e.opts("line-1/a")
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
	rep, err := e.place("line-1/a")
	if err != nil || rep.Verdict != "failed" {
		t.Fatalf("place = %+v, %v; it must have recorded", rep, err)
	}
	if testutil.Snapshot(t, e.root) != before {
		t.Fatal("place changed the repository's git state")
	}
	// a refused place changes nothing either
	_, err = e.place("line-1/a")
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
	if _, err := e.place("line-1/a"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(helper, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.write(t, "line-1/a", failingOnH0+"-- ceiling --\nzzz\n", fixtureProbe)
	e.commit(t)
	_, err := e.place("line-1/a")
	refusedBy(t, err, "Adds only")
	if !testutil.Mentions(err, "helper") {
		t.Fatalf("err = %v, want it to name helper", err)
	}
	// the same folder with only the ceiling added passes
	if err := os.WriteFile(helper, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.commit(t)
	if rep, err := e.place("line-1/a"); err != nil || rep.Version != 2 {
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
	_, err = e.place("line-1/a")
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
// Probe ran, leaves no file in the store: the outputs are stored only with
// the event that names them.
func TestKindRefusalStoresNoOutput(t *testing.T) {
	e := newPlaceEnv(t)
	e.example(t, "line-1/a", metOnH0)
	before := outputFiles(t, e.root)
	_, err := e.place("line-1/a")
	refusedBy(t, err, "Kind")
	if after := outputFiles(t, e.root); !slices.Equal(after, before) {
		t.Fatalf("outputs after a Kind refusal = %v, were %v", after, before)
	}
}
