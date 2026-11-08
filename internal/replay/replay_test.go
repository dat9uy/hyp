package replay

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

const (
	sup = event.VerdictSupports
	con = event.VerdictContradicts
	inc = event.VerdictInconclusive
)

func replayB(b *testutil.B) Projection { return Replay(b.Evs, nil, "") }

// wantStatus takes the status word of any node kind.
func wantStatus[S ~string](t *testing.T, p Projection, id string, status S, reason string) {
	t.Helper()
	got := statuses(p)[id]
	if got.Status != string(status) || got.Reason != reason {
		t.Fatalf("%s: got %s/%s, want %s/%s", id, got.Status, got.Reason, status, reason)
	}
}

// nodeStatus is one node's derived status, whatever its kind.
type nodeStatus struct {
	Status, Reason string
	Stale          *StaleCause
	Candidates     []string
}

// statuses is every node's derived status, so a test can compare them all.
func statuses(p Projection) map[string]nodeStatus {
	out := map[string]nodeStatus{}
	for id := range p.Nodes {
		s, _ := p.Display(id)
		ns := nodeStatus{Status: s.Word, Reason: s.Reason, Stale: s.Stale}
		if q := p.Questions[id]; q != nil {
			ns.Candidates = q.Candidates
		}
		out[id] = ns
	}
	return out
}

func wantStale(t *testing.T, p Projection, id string, want *StaleCause) {
	t.Helper()
	if got := statuses(p)[id].Stale; !reflect.DeepEqual(got, want) {
		t.Fatalf("%s stale: got %s, want %s", id, chain(got), chain(want))
	}
}

func chain(c *StaleCause) string {
	var parts []string
	for ; c != nil; c = c.Next {
		parts = append(parts, c.Node+" "+c.Reason)
	}
	if parts == nil {
		return "none"
	}
	return strings.Join(parts, " ← ")
}

func wantNoRejections(t *testing.T, p Projection) {
	t.Helper()
	if len(p.Rejections) > 0 {
		t.Fatalf("unexpected rejections: %+v", p.Rejections)
	}
}

func TestScenario01SupportedUnselected(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	b.Supports(h1)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, h1, HypSupported, "")
	wantStatus(t, p, q, DQOpen, ReasonSupportedUnsel)
	if !reflect.DeepEqual(p.Questions[q].Candidates, []string{h1}) {
		t.Fatalf("candidates = %v, want [%s]", p.Questions[q].Candidates, h1)
	}
}

func TestScenario02SelectAnswers(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	b.Supports(h1)
	b.As("a").Select(q, h1, "", false)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
	if p.Goal.Status != GoalComplete {
		t.Fatalf("goal = %s, want complete", p.Goal.Status)
	}
}

// The owner's choice between two candidates, recorded as the select, ends
// the trade-off: no confirm follows it.
func TestScenario03ChoiceOnTwoSupportedAnswers(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	h2 := b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	p := replayB(b)
	if len(p.Questions[q].Candidates) != 2 {
		t.Fatalf("candidates = %v, want both candidates (trade-off)", p.Questions[q].Candidates)
	}
	b.As("a").Select(q, h1, "", false)
	p = replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
}

type scenario04 struct {
	b                      *testutil.B
	q1, q2, q3, h1, h2, h3 string
	e3                     string
}

// buildScenario04 answers three chained questions: H1 depends_on H2 depends_on H3.
func buildScenario04() scenario04 {
	c := testutil.BuildChain()
	return scenario04{b: c.B, q1: c.Q1, q2: c.Q2, q3: c.Q3, h1: c.H1, h2: c.H2, h3: c.H3, e3: c.E3}
}

func TestScenario04ContradictionPropagatesAndRetractRestores(t *testing.T) {
	s := buildScenario04()
	b := s.b
	p := replayB(b)
	wantNoRejections(t, p)
	if p.Goal.Status != GoalComplete {
		t.Fatalf("goal = %s, want complete", p.Goal.Status)
	}
	before, beforeGoal := statuses(p), p.Goal

	r := b.Result(s.e3, map[string]string{s.h3: con})
	p = replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, s.h3, HypContradicted, "")
	byR := &StaleCause{Node: r, Reason: StaleReasonContradicts}
	wantStale(t, p, s.h2, &StaleCause{Node: s.h3, Reason: string(HypContradicted), Next: byR})
	wantStale(t, p, s.h1, &StaleCause{Node: s.h2, Reason: StaleReasonStale,
		Next: &StaleCause{Node: s.h3, Reason: string(HypContradicted), Next: byR}})
	wantStatus(t, p, s.q1, DQReopened, ReasonSelectedStale)
	wantStatus(t, p, s.q2, DQReopened, ReasonSelectedStale)
	wantStatus(t, p, s.q3, DQReopened, ReasonSelectedNotSupport)
	if p.Goal.Status != GoalInProgress {
		t.Fatalf("goal = %s, want in_progress", p.Goal.Status)
	}

	b.Retract(r)
	p = replayB(b)
	wantNoRejections(t, p)
	// The retracted Result is the only node that did not exist before; every
	// node that did must be exactly as it was.
	for id := range statuses(p) {
		if _, existed := before[id]; !existed && id != r {
			t.Fatalf("unexpected new node %s", id)
		}
	}
	after := map[string]nodeStatus{}
	for id := range before {
		after[id] = statuses(p)[id]
	}
	if !reflect.DeepEqual(before, after) {
		for id := range before {
			if !reflect.DeepEqual(before[id], after[id]) {
				t.Errorf("%s: before %+v, after %+v", id, before[id], after[id])
			}
		}
		t.Fatal("retract did not restore the earlier statuses exactly")
	}
	if !reflect.DeepEqual(beforeGoal, p.Goal) {
		t.Fatalf("goal before %+v, after %+v", beforeGoal, p.Goal)
	}
}

func TestScenario05ReaffirmCannotClearUnsupported(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q1, q2 := b.DQ(g, "q1"), b.DQ(g, "q2")
	h1, h2 := b.Hyp(q1, "h1"), b.Hyp(q2, "h2")
	b.Supports(h1)
	b.Edge(h1, event.EdgeDependsOn, h2)
	bad := b.Contradicts(h2)
	b.Reaffirm(h1)
	p := replayB(b)
	wantNoRejections(t, p)
	if p.Hyps[h1].Stale == nil {
		t.Fatal("reaffirm cleared staleness while the target is contradicted")
	}

	b.Retract(bad)
	b.Supports(h2)
	p = replayB(b)
	// The latest in-force basis of H1 (the reaffirm) recorded H2 as
	// contradicted, so H1 still reports the change since its basis.
	wantStale(t, p, h1, &StaleCause{Node: h2, Reason: StaleReasonChanged})
	b.Reaffirm(h1)
	p = replayB(b)
	wantNoRejections(t, p)
	wantStale(t, p, h1, nil)
}

func TestScenario06SupersedeSelected(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	b.Supports(h1)
	b.Answer(q, h1, "")
	q2 := b.DQ(g, "q2")
	h9 := b.Hyp(q2, "h9")
	b.Supports(h9)
	b.Edge(h9, event.EdgeDependsOn, h1)
	wantStatus(t, replayB(b), q, DQAnswered, "")

	h1b := b.Hyp(q, "h1b")
	b.Edge(h1b, event.EdgeSupersedes, h1)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, h1, HypSuperseded, "")
	wantStatus(t, p, q, DQReopened, ReasonSelectedNotSupport)
	wantStale(t, p, h9, &StaleCause{Node: h1, Reason: string(HypSuperseded)})

	b.Supports(h1b)
	p = replayB(b)
	if !reflect.DeepEqual(p.Questions[q].Candidates, []string{h1b}) {
		t.Fatalf("candidates = %v, want [%s]", p.Questions[q].Candidates, h1b)
	}
}

func TestScenario07AwaitingSubquestions(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	b.Supports(h1)
	qc := b.DQ(h1, "qc")
	wantStatus(t, replayB(b), h1, HypSupported, "")

	b.Answer(q, h1, "")
	wantStatus(t, replayB(b), q, DQOpen, ReasonAwaitingSubquestions)

	hc := b.Hyp(qc, "hc")
	b.Supports(hc)
	b.Answer(qc, hc, "")
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, h1, HypSupported, "")
	wantStatus(t, p, q, DQAnswered, "")
	if p.Goal.Status != GoalComplete {
		t.Fatalf("goal = %s, want complete", p.Goal.Status)
	}
}

func TestScenario08ConflictingSelections(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	b.As("a").Select(q, h1, "", false)
	sb := b.As("b").Select(q, h2, "", false)
	wantStatus(t, replayB(b), q, DQReopened, ReasonConflicting)
	b.As("a").Select(q, h1, sb, false)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
}

func TestScenario09DeferResumeWithdraw(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q2 := b.DQ(g, "q2")
	h2 := b.Hyp(q2, "h2")
	b.Supports(h2)
	b.Answer(q2, h2, "")
	q := b.DQ(g, "q")
	b.Hyp(q, "h1")
	wantStatus(t, replayB(b), q, DQOpen, ReasonUntested)
	b.Defer(q)
	wantStatus(t, replayB(b), q, DQDeferred, "")
	b.Resume(q)
	wantStatus(t, replayB(b), q, DQOpen, ReasonUntested)
	b.Withdraw(q)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQWithdrawn, "")
	if p.Goal.Status != GoalComplete {
		t.Fatalf("goal = %s, want complete (withdrawn DQ ignored)", p.Goal.Status)
	}
}

func TestScenario10CrossActorCycleRejected(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q1, q2 := b.DQ(g, "q1"), b.DQ(g, "q2")
	h1, h2 := b.Hyp(q1, "h1"), b.Hyp(q2, "h2")
	b.As("a").Edge(h1, event.EdgeDependsOn, h2)
	second := b.As("b").Edge(h2, event.EdgeDependsOn, h1)
	files := map[string][]event.Event{}
	for _, ev := range b.Evs {
		files[ev.Actor] = append(files[ev.Actor], ev)
	}
	p := Replay(Merge(files), nil, "")
	if len(p.Rejections) != 1 || p.Rejections[0].EventID != second {
		t.Fatalf("rejections = %+v, want only %s", p.Rejections, second)
	}
}

func TestScenario11RecipeChangeFlagsOnly(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	e := b.Exp(h1, "boot@1")
	r := b.Result(e, map[string]string{h1: sup})
	changed := Replay(b.Evs, map[string]string{"boot@1": "bbb"}, "")
	same := Replay(b.Evs, map[string]string{"boot@1": "aaa"}, "")
	wantNoRejections(t, changed)
	if !changed.Results[r].Questionable || same.Results[r].Questionable {
		t.Fatalf("questionable: changed=%v same=%v", changed.Results[r].Questionable, same.Results[r].Questionable)
	}
	if !reflect.DeepEqual(statuses(changed), statuses(same)) || !reflect.DeepEqual(changed.Goal, same.Goal) {
		t.Fatal("a recipe change altered a status")
	}
}

func TestNewerRecipeVersionLeavesOlderResultsClean(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	r := b.Result(b.Exp(h1, "boot@1"), map[string]string{h1: sup})
	if p := Replay(b.Evs, map[string]string{"boot@1": "aaa", "boot@2": "bbb"}, ""); p.Results[r].Questionable {
		t.Fatal("boot@2 beside boot@1 flagged the boot@1 Result")
	}
	if p := Replay(b.Evs, map[string]string{"boot@2": "aaa"}, ""); !p.Results[r].Questionable {
		t.Fatal("a boot@1 Result matched the hash of boot@2")
	}
}

// evidenceResult records a Result under e citing ptrs, with hashes as its
// evidence_hashes.
func evidenceResult(b *testutil.B, e, h string, ptrs []string, hashes map[string]string) string {
	return b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: e,
		Verdicts: map[string]string{h: sup}, Observation: "o", Evidence: ptrs, EvidenceHashes: hashes})
}

func TestEvidencePinFlagsResult(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	e := b.Exp(h1, event.RecipeNone)
	ptr := "recipes/cases@1/known.txt"
	r := evidenceResult(b, e, h1, []string{".tree/evidence/E-1/log.txt", ptr}, map[string]string{ptr: "aaa"})
	plain := b.Result(e, map[string]string{h1: sup})

	same := Replay(b.Evs, map[string]string{"cases@1": "aaa"}, "")
	wantNoRejections(t, same)
	if same.Results[r].Questionable {
		t.Fatal("an unchanged evidence pin flagged its Result")
	}
	for name, hashes := range map[string]map[string]string{
		"changed": {"cases@1": "bbb"},
		"missing": {"cases@2": "aaa"},
	} {
		p := Replay(b.Evs, hashes, "")
		if !p.Results[r].Questionable {
			t.Fatalf("%s: the evidence pin did not flag its Result", name)
		}
		if p.Results[plain].Questionable {
			t.Fatalf("%s: a Result without pins was flagged", name)
		}
		if !reflect.DeepEqual(statuses(p), statuses(same)) || !reflect.DeepEqual(p.Goal, same.Goal) {
			t.Fatalf("%s: an evidence pin altered a status", name)
		}
	}
}

func TestPinsListsExperimentThenEvidence(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	e := b.Exp(h1, "boot@1")
	a, c := "recipes/boot@1/a.txt", "./recipes/cases@2/c.txt"
	r := evidenceResult(b, e, h1, []string{c, "https://x", a, c}, map[string]string{a: "aaa", c: "ccc"})
	p := Replay(b.Evs, nil, "")
	wantNoRejections(t, p)
	boot, cases := event.Ref{Name: "boot", Version: "1"}, event.Ref{Name: "cases", Version: "2"}
	want := []Pin{{Ref: boot, Hash: "aaa"}, {Ref: cases, Hash: "ccc", Pointer: c}, {Ref: boot, Hash: "aaa", Pointer: a}}
	if got := p.Results[r].Pins; !reflect.DeepEqual(got, want) {
		t.Fatalf("Pins = %+v, want %+v", got, want)
	}
}

func TestEvidenceHashesValidated(t *testing.T) {
	ptr := "recipes/cases@1/known.txt"
	for _, tc := range []struct {
		name     string
		evidence []string
		hashes   map[string]string
		want     string
	}{
		{"key not cited", []string{ptr}, map[string]string{"recipes/cases@1/other.txt": "aaa"}, "bad evidence_hashes: recipes/cases@1/other.txt is not in evidence"},
		{"no version", []string{"recipes/cases/known.txt"}, map[string]string{"recipes/cases/known.txt": "aaa"}, "is not in a versioned recipe"},
		{"no file", []string{"recipes/cases@1"}, map[string]string{"recipes/cases@1": "aaa"}, "is not in a versioned recipe"},
		{"trailing slash", []string{"recipes/cases@1/"}, map[string]string{"recipes/cases@1/": "aaa"}, "is not in a versioned recipe"},
		{"store evidence", []string{".tree/evidence/E-1/a.txt"}, map[string]string{".tree/evidence/E-1/a.txt": "aaa"}, "is not in a versioned recipe"},
		{"not hex", []string{ptr}, map[string]string{ptr: "XYZ"}, `invalid evidence_hashes value for recipes/cases@1/known.txt: "XYZ"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := testutil.New()
			g := b.Goal("g")
			q := b.DQ(g, "q")
			h1 := b.Hyp(q, "h1")
			e := b.Exp(h1, event.RecipeNone)
			r := evidenceResult(b, e, h1, tc.evidence, tc.hashes)
			wantRejected(t, Replay(b.Evs, nil, ""), r, tc.want)
		})
	}
}

func TestScenario13ChildUnderAbandonedDoesNotBlock(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	qc := b.DQ(h2, "qc")
	b.Abandon(h2)
	b.Supports(h1)
	b.Answer(q, h1, "")
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, qc, DQOpen, ReasonNoCandidates)
	if p.Goal.Status != GoalComplete {
		t.Fatalf("goal = %s, want complete", p.Goal.Status)
	}
}

func TestScenario14BadEventsSkipped(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	clean := len(b.Evs)
	b.Emit("bogus.type", g, event.Empty{})
	b.Hyp(g, "hyp under goal")
	b.Evs = append(b.Evs, b.Evs[1])
	b.Supports(h1)
	p := replayB(b)
	if len(p.Rejections) != 3 {
		t.Fatalf("rejections = %+v, want 3", p.Rejections)
	}
	for i, want := range []string{"unknown event type", "wrong parent kind", "duplicate id"} {
		if !strings.Contains(p.Rejections[i].Reason, want) {
			t.Errorf("rejection %d = %q, want %q", i, p.Rejections[i].Reason, want)
		}
	}
	// The three bad events sit at clean..clean+2; drop them by position, since
	// the duplicate is byte-identical to the event it copies.
	good := append(append([]event.Event(nil), b.Evs[:clean]...), b.Evs[clean+3:]...)
	want := Replay(good, nil, "")
	if len(want.Rejections) != 0 {
		t.Fatalf("clean log rejected: %+v", want.Rejections)
	}
	if !reflect.DeepEqual(statuses(p), statuses(want)) || !reflect.DeepEqual(p.Goal, want.Goal) ||
		!reflect.DeepEqual(p.Applied, want.Applied) {
		t.Fatal("bad events changed the statuses")
	}
}

func TestScenario15AsOfEqualsTruncation(t *testing.T) {
	s := buildScenario04()
	r := s.b.Result(s.e3, map[string]string{s.h3: con})
	s.b.Retract(r)
	evs := s.b.Evs
	for i := range evs {
		a := Replay(evs, nil, evs[i].ID)
		w := Replay(evs[:i+1], nil, "")
		if !reflect.DeepEqual(statuses(a), statuses(w)) || !reflect.DeepEqual(a.Goal, w.Goal) {
			t.Fatalf("asOf %d differs from truncation", i)
		}
	}
}

func TestScenario16InconclusiveReason(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	b.Result(b.Exp(h1, event.RecipeNone), map[string]string{h1: inc})
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, h1, HypInconclusive, "")
	wantStatus(t, p, q, DQOpen, ReasonInconclusive)
}

func TestMergeKeepsFileOrder(t *testing.T) {
	id := func(last string) string { return "01JB000000000000000000000" + last }
	files := map[string][]event.Event{
		"a": {{ID: id("5")}, {ID: id("3")}},
		"b": {{ID: id("4")}},
	}
	var got []string
	for _, ev := range Merge(files) {
		got = append(got, ev.ID)
	}
	want := []string{id("4"), id("5"), id("3")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merge = %v, want %v", got, want)
	}
}

func TestRefusals(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	q2 := b.DQ(g, "q2")
	h1 := b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindHypothesis, Parent: q, Title: "h1", Alias: "h-one"})
	h2, h5 := b.Hyp(q, "h2"), b.Hyp(q, "h5")
	h3 := b.Hyp(q2, "h3")
	h4 := b.Hyp(q, "h4")
	b.Abandon(h4)
	e1 := b.Exp(h1, event.RecipeNone)
	r1 := b.Result(e1, map[string]string{h1: sup})
	b.Retract(r1)
	sel := b.As("a").Select(q, h1, "", false)
	b.As("a").Edge(h5, event.EdgeSupersedes, h2)
	b.Edge(h1, event.EdgeDependsOn, h3)
	st, p := Build(b.Evs, nil, "")
	wantNoRejections(t, p)

	n := 0
	mk := func(typ, node, actor string, payload any) event.Event {
		saved := b.Evs
		b.N = 1000 + n
		n++
		b.As(actor)
		id := b.Emit(typ, node, payload)
		ev := b.Evs[len(b.Evs)-1]
		b.Evs = saved
		_ = id
		return ev
	}
	node := func(kind, parent, title string) event.Event {
		return mk(event.TypeNodeAdd, "", "a", event.NodeAdd{Kind: kind, Parent: parent, Title: title})
	}
	with := func(ev event.Event, f func(*event.Event)) event.Event { f(&ev); return ev }

	rows := []struct {
		rule string
		ev   event.Event
		want string
	}{
		{"1 version", with(node(event.KindHypothesis, q, "x"), func(e *event.Event) { e.V = 2 }), "unsupported version"},
		{"1 unknown type", mk("node.edit", h1, "a", event.Empty{}), "unknown event type"},
		{"1 missing why", with(node(event.KindHypothesis, q, "x"), func(e *event.Event) { e.Why = "" }), "missing field: why"},
		{"1 unknown payload field", with(node(event.KindHypothesis, q, "x"), func(e *event.Event) {
			e.Payload = []byte(`{"kind":"Hypothesis","parent":"` + q + `","title":"x","status":"supported"}`)
		}), "unknown field"},
		{"1 ts not UTC", with(node(event.KindHypothesis, q, "x"), func(e *event.Event) { e.TS = "2026-10-24T07:00:00+07:00" }), "not UTC"},
		{"1 over-cap title", node(event.KindHypothesis, q, strings.Repeat("ế", 121)), "over limit: title"},
		{"2 duplicate id", with(node(event.KindHypothesis, q, "x"), func(e *event.Event) { e.ID = h1; e.Node = h1 }), "duplicate id"},
		{"3 unknown node", node(event.KindHypothesis, "01JB9999999999999999999999", "x"), "unknown node"},
		{"4 wrong parent kind", node(event.KindHypothesis, g, "x"), "wrong parent kind"},
		{"4 second goal", node(event.KindGoal, "", "g2"), "only one Goal"},
		{"5 target in another DQ", mk(event.TypeNodeAdd, "", "a", event.NodeAdd{Kind: event.KindExperiment, Parent: h1,
			Title: "x", Recipe: event.RecipeNone, Targets: []string{h3}}), "bad target"},
		{"5 decided_by on a Hypothesis", mk(event.TypeNodeAdd, "", "a", event.NodeAdd{Kind: event.KindHypothesis, Parent: q,
			Title: "x", DecidedBy: event.DecidedReading}), "unknown field"},
		{"5 unknown decided_by", mk(event.TypeNodeAdd, "", "a", event.NodeAdd{Kind: event.KindExperiment, Parent: h1,
			Title: "x", Recipe: event.RecipeNone, DecidedBy: "measure"}), "invalid decided_by"},
		{"6 verdict not parent or target", mk(event.TypeResultRecord, "", "a", event.ResultRecord{Parent: e1,
			Verdicts: map[string]string{h2: sup}, Observation: "o"}), "bad verdict"},
		{"7 retract twice", mk(event.TypeResultRetract, r1, "a", event.Empty{}), "already retracted"},
		{"7 abandon twice", mk(event.TypeHypothesisAbandon, h4, "a", event.Empty{}), "already abandoned"},
		{"8 select non-child", mk(event.TypeDQSelect, q, "a", event.DQSelect{Hypothesis: h3}), "bad select"},
		{"8 select abandoned", mk(event.TypeDQSelect, q, "a", event.DQSelect{Hypothesis: h4}), "bad select"},
		{"8 select superseded", mk(event.TypeDQSelect, q, "a", event.DQSelect{Hypothesis: h2}), "bad select"},
		{"9 confirm unknown select", mk(event.TypeConfirm, q2, "b", event.Confirm{Select: sel,
			Role: event.RoleOwner, Verdict: event.ConfirmSufficient}), "bad confirm"},
		{"10 supersedes self", mk(event.TypeEdgeAdd, h1, "a", event.Edge{Type: event.EdgeSupersedes, To: h1}), "self edge"},
		{"10 cross-DQ supersedes", mk(event.TypeEdgeAdd, h1, "a", event.Edge{Type: event.EdgeSupersedes, To: h3}), "cross-DQ supersedes"},
		{"10 supersedes cycle", mk(event.TypeEdgeAdd, h2, "a", event.Edge{Type: event.EdgeSupersedes, To: h5}), "supersedes cycle"},
		{"11 depends_on from Experiment", mk(event.TypeEdgeAdd, e1, "a", event.Edge{Type: event.EdgeDependsOn, To: h3}), "wrong kind"},
		{"11 depends_on to DQ", mk(event.TypeEdgeAdd, h1, "a", event.Edge{Type: event.EdgeDependsOn, To: q2}), "wrong kind"},
		{"11 depends_on cycle", mk(event.TypeEdgeAdd, h3, "a", event.Edge{Type: event.EdgeDependsOn, To: h1}), "depends_on cycle"},
		{"11 cycle through parent links", mk(event.TypeEdgeAdd, h1, "a", event.Edge{Type: event.EdgeDependsOn, To: h1}), "self edge"},
		{"12 duplicate edge", mk(event.TypeEdgeAdd, h1, "a", event.Edge{Type: event.EdgeDependsOn, To: h3}), "duplicate edge"},
		{"12 remove non-live edge", mk(event.TypeEdgeRemove, h1, "a", event.Edge{Type: event.EdgeDependsOn, To: h2}), "edge not live"},
		{"12 unknown edge type", mk(event.TypeEdgeAdd, h1, "a", event.Edge{Type: "blocks", To: h2}), "unknown edge type"},
		{"13 alias used", mk(event.TypeNodeAdd, "", "a", event.NodeAdd{Kind: event.KindHypothesis, Parent: q,
			Title: "x", Alias: "h-one"}), "alias already used"},
	}
	for _, row := range rows {
		err := Validate(st, row.ev)
		if err == nil || !strings.Contains(err.Error(), row.want) {
			t.Errorf("rule %s: got %v, want error containing %q", row.rule, err, row.want)
		}
	}
	if err := Validate(st, event.Event{}); err == nil {
		t.Error("empty event accepted")
	}
}

func TestGoalWithNoAnsweredQuestionIsInProgress(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	p := replayB(b)
	wantNoRejections(t, p)
	if p.Goal.Status != GoalInProgress {
		t.Fatalf("goal with no questions = %s, want in_progress", p.Goal.Status)
	}
	q := b.DQ(g, "q")
	b.Withdraw(q)
	p = replayB(b)
	wantNoRejections(t, p)
	if p.Goal.Status != GoalInProgress {
		t.Fatalf("goal with only a withdrawn question = %s, want in_progress", p.Goal.Status)
	}
}

func wantRejected(t *testing.T, p Projection, id, want string) {
	t.Helper()
	for _, r := range p.Rejections {
		if r.EventID == id {
			if !strings.Contains(r.Reason, want) {
				t.Fatalf("%s rejected with %q, want %q", id, r.Reason, want)
			}
			return
		}
	}
	t.Fatalf("%s not rejected, want %q", id, want)
}

func wantLimit(t *testing.T, p Projection, q string, i int, state, by string) {
	t.Helper()
	ls := p.LimitsOf(q)
	if len(ls) <= i {
		t.Fatalf("limits under %s = %+v, want at least %d", q, ls, i+1)
	}
	if ls[i].State != state || ls[i].By != by {
		t.Fatalf("limit %s: got %s by %q, want %s by %q", ls[i].Key, ls[i].State, ls[i].By, state, by)
	}
}

func amend(b *testutil.B, h, body string) string {
	return b.Emit(event.TypeHypothesisAmend, h, event.HypothesisAmend{Body: body})
}

func limitResult(b *testutil.B, h string, limits, closes []string) string {
	return b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h, event.RecipeNone),
		Verdicts: map[string]string{h: sup}, Observation: "obs", Limits: limits, Closes: closes})
}

func TestAmendReplacesBody(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindHypothesis, Parent: q, Title: "h", Body: "old"})
	b.Supports(h)
	b.As("a").Select(q, h, "", false)
	amend(b, h, "new")
	p := replayB(b)
	wantNoRejections(t, p)
	if got := p.Nodes[h].Body; got != "new" {
		t.Fatalf("body = %q, want new", got)
	}
	wantStatus(t, p, q, DQAnswered, "")
}

func TestAmendRefusals(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindHypothesis, Parent: q, Title: "h", Body: "old"})
	h2 := b.Hyp(q, "h2")
	b.Abandon(h2)
	same := amend(b, h, "old")
	gone := amend(b, h2, "new")
	p := replayB(b)
	wantRejected(t, p, same, "body is unchanged")
	wantRejected(t, p, gone, "superseded or abandoned")
}

func TestLimitStates(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	r1 := limitResult(b, h, []string{"l1"}, nil)
	p := replayB(b)
	if ls := p.LimitsOf(q); len(ls) != 1 || ls[0].Key != event.LimitKey(r1, 1) || ls[0].Text != "l1" {
		t.Fatalf("limits = %+v, want one l1 on %s", ls, r1)
	}
	wantLimit(t, p, q, 0, LimitOpen, "")
	r2 := limitResult(b, h, nil, []string{event.LimitKey(r1, 1)})
	p = replayB(b)
	wantNoRejections(t, p)
	wantLimit(t, p, q, 0, LimitClosed, r2)
	b.Retract(r2)
	p = replayB(b)
	wantNoRejections(t, p)
	wantLimit(t, p, q, 0, LimitOpen, "")
}

func TestLimitHyps(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h2, event.RecipeNone, h1),
		Verdicts: map[string]string{h2: event.VerdictSupports, h1: event.VerdictContradicts}, Observation: "o", Limits: []string{"l1"}})
	p := replayB(b)
	wantNoRejections(t, p)
	if got := p.LimitsOf(q)[0].Hyps; !slices.Equal(got, []string{h1, h2}) {
		t.Fatalf("Hyps = %v, want [%s %s]", got, h1, h2)
	}
}

func TestLimitAcceptance(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	r1 := limitResult(b, h, []string{"l1"}, nil)
	sel := b.As("a").Select(q, h, "", true)
	b.As("o").Emit(event.TypeConfirm, q, event.Confirm{Select: sel, Role: event.RoleOwner,
		Verdict: event.ConfirmSufficient, Accepts: []string{event.LimitKey(r1, 1)}})
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
	wantLimit(t, p, q, 0, LimitAccepted, "o")
	amend(b.As("a"), h, "new")
	p = replayB(b)
	wantNoRejections(t, p)
	wantLimit(t, p, q, 0, LimitOpen, "")
}

func TestInsufficientOwnerConfirmCancelsAcceptance(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	r1 := limitResult(b, h, []string{"l1"}, nil)
	sel := b.As("a").Select(q, h, "", true)
	accept := func() {
		b.As("o").Emit(event.TypeConfirm, q, event.Confirm{Select: sel, Role: event.RoleOwner,
			Verdict: event.ConfirmSufficient, Accepts: []string{event.LimitKey(r1, 1)}})
	}
	accept()
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
	wantLimit(t, p, q, 0, LimitAccepted, "o")

	// An old verifier's no is history: the owner's acceptance stays.
	b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	p = replayB(b)
	wantNoRejections(t, p)
	wantLimit(t, p, q, 0, LimitAccepted, "o")

	// The owner's no withdraws their acceptance; the status is unchanged by it.
	b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmInsufficient)
	p = replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQReopened, ReasonAwaitingOwner)
	wantLimit(t, p, q, 0, LimitOpen, "")

	accept()
	p = replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
	wantLimit(t, p, q, 0, LimitAccepted, "o")
}

func TestLimitRefusals(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q, q2 := b.DQ(g, "q"), b.DQ(g, "q2")
	h, h2 := b.Hyp(q, "h"), b.Hyp(q2, "h2")
	r1 := limitResult(b, h, []string{"l1"}, nil)
	sel := b.As("a").Select(q, h, "", false)
	bad := limitResult(b, h, nil, []string{"nope"})
	tooBig := limitResult(b, h, nil, []string{event.LimitKey(r1, 2)})
	other := limitResult(b, h2, nil, []string{event.LimitKey(r1, 1)})
	no := b.As("o").Emit(event.TypeConfirm, q, event.Confirm{Select: sel, Role: event.RoleOwner,
		Verdict: event.ConfirmInsufficient, Accepts: []string{event.LimitKey(r1, 1)}})
	// An old verifier confirm that accepts a limit is history: replay takes
	// it and it accepts nothing.
	b.As("v").Emit(event.TypeConfirm, q, event.Confirm{Select: sel, Role: event.RoleVerifier,
		Verdict: event.ConfirmSufficient, Accepts: []string{event.LimitKey(r1, 1)}})
	p := replayB(b)
	wantRejected(t, p, bad, "bad limit ref")
	wantRejected(t, p, tooBig, "bad limit ref")
	wantRejected(t, p, other, "bad limit ref")
	wantRejected(t, p, no, "only a sufficient owner confirm accepts limits")
	if len(p.Rejections) != 4 {
		t.Fatalf("rejections = %+v, want the four above", p.Rejections)
	}
	wantLimit(t, p, q, 0, LimitOpen, "")
}

func TestLimitAcceptanceOnStaleSelectDoesNotOverride(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	r1 := limitResult(b, h, []string{"l1"}, nil)
	old := b.As("a").Select(q, h, "", true)
	cur := b.As("a").Select(q, h, old, true)
	accept := func(sel string) {
		b.As("o").Emit(event.TypeConfirm, q, event.Confirm{Select: sel, Role: event.RoleOwner,
			Verdict: event.ConfirmSufficient, Accepts: []string{event.LimitKey(r1, 1)}})
	}
	accept(cur)
	accept(old)
	p := replayB(b)
	wantNoRejections(t, p)
	wantLimit(t, p, q, 0, LimitAccepted, "o")
	b.As("a").Select(q, h, cur, true)
	wantLimit(t, replayB(b), q, 0, LimitOpen, "")
}

// A stale supported sibling is still a candidate, so selecting h1 is the
// owner's choice between two, and it answers.
func TestStaleSupportedSiblingIsACandidate(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	q2 := b.DQ(g, "q2")
	h3 := b.Hyp(q2, "h3")
	b.Supports(h1)
	b.Supports(h3)
	b.Edge(h2, event.EdgeDependsOn, h3)
	b.Supports(h2)
	b.Contradicts(h3)
	p := replayB(b)
	wantStatus(t, p, h2, HypSupported, "")
	if p.Hyps[h2].Stale == nil {
		t.Fatalf("h2 is not stale; the test needs a stale supported sibling")
	}
	if !reflect.DeepEqual(p.Questions[q].Candidates, []string{h1, h2}) {
		t.Fatalf("candidates = %v, want [%s %s]", p.Questions[q].Candidates, h1, h2)
	}
	b.As("a").Select(q, h1, "", false)
	p = replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
}

func TestLoneCandidateStaysACandidateWhenNotReady(t *testing.T) {
	build := func() (b *testutil.B, q, h, tgt string) {
		b = testutil.New()
		g := b.Goal("g")
		q = b.DQ(g, "q")
		h = b.Hyp(q, "h")
		tgt = b.Hyp(b.DQ(g, "q2"), "t")
		b.Supports(tgt)
		b.Edge(h, event.EdgeDependsOn, tgt)
		b.Supports(h)
		return b, q, h, tgt
	}
	b, q, h, tgt := build()
	if got := replayB(b).Questions[q].Candidates; !reflect.DeepEqual(got, []string{h}) {
		t.Fatalf("ready: candidates = %v, want [%s]", got, h)
	}
	b.Contradicts(tgt)
	if got := replayB(b).Questions[q].Candidates; !reflect.DeepEqual(got, []string{h}) {
		t.Fatalf("stale: candidates = %v, want [%s]", got, h)
	}
	b, q, h, _ = build()
	b.DQ(h, "sub")
	if got := replayB(b).Questions[q].Candidates; !reflect.DeepEqual(got, []string{h}) {
		t.Fatalf("waiting: candidates = %v, want [%s]", got, h)
	}
}

// twoCandidates builds two supported rivals with h2 depending on a supported
// t in a second question.
func twoCandidates() (b *testutil.B, q, h1, h2, tgt string) {
	b = testutil.New()
	g := b.Goal("g")
	q = b.DQ(g, "q")
	h1, h2 = b.Hyp(q, "h1"), b.Hyp(q, "h2")
	tgt = b.Hyp(b.DQ(g, "q2"), "t")
	b.Supports(h1)
	b.Supports(h2)
	b.Supports(tgt)
	b.Edge(h2, event.EdgeDependsOn, tgt)
	return b, q, h1, h2, tgt
}

func TestTwoCandidatesWithOneNotReadyStayATradeoff(t *testing.T) {
	b, q, h1, h2, tgt := twoCandidates()
	b.As("a").Contradicts(tgt)
	if got := replayB(b).Questions[q].Candidates; !reflect.DeepEqual(got, []string{h1, h2}) {
		t.Fatalf("one stale: candidates = %v, want both", got)
	}
	b, q, h1, h2, _ = twoCandidates()
	b.As("a").DQ(h2, "sub")
	if got := replayB(b).Questions[q].Candidates; !reflect.DeepEqual(got, []string{h1, h2}) {
		t.Fatalf("one waiting: candidates = %v, want both", got)
	}
}

func TestLiveRivalsSkipDeadHypotheses(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2, h3, h4, h5 := b.Hyp(q, "h1"), b.Hyp(q, "h2"), b.Hyp(q, "h3"), b.Hyp(q, "h4"), b.Hyp(q, "h5")
	b.Contradicts(h2)
	b.Abandon(h3)
	b.Edge(h5, event.EdgeSupersedes, h4)
	st, _ := Build(b.Evs, nil, "")
	if got, want := st.LiveRivals(q, h1), []string{h5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LiveRivals = %v, want %v", got, want)
	}
	if !LiveHypStatus(HypContradicted) || RivalStatus(HypContradicted) {
		t.Fatal("a contradicted Hypothesis must be a live node and not a live rival")
	}
}

func TestLimitGatingFollowsCurrentSelect(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	limitResult(b, h1, []string{"only h1"}, nil)
	limitResult(b, h2, []string{"only h2"}, nil)
	if p := replayB(b); p.LimitsOf(q)[0].Gating || p.LimitsOf(q)[1].Gating {
		t.Fatal("with no select, no limit gates")
	}
	b.As("a").Select(q, h1, "", false)
	p := replayB(b)
	if !p.LimitsOf(q)[0].Gating || p.LimitsOf(q)[1].Gating {
		t.Fatalf("after selecting h1: gating = %v, %v; want true, false", p.LimitsOf(q)[0].Gating, p.LimitsOf(q)[1].Gating)
	}
}

func TestUnjudgedReadsRivalsAtAdd(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2, h3 := b.Hyp(q, "h1"), b.Hyp(q, "h2"), b.Hyp(q, "h3")
	e := b.Exp(h1, event.RecipeNone, h2, h3)
	p := replayB(b)
	if got := p.Exps[e].Unjudged; got != nil {
		t.Fatalf("planned Experiment: Unjudged = %v, want nil", got)
	}
	if p.Exps[e].OneRival {
		t.Fatal("an Experiment that names live rivals is not one-rival")
	}
	b.Result(e, map[string]string{h1: sup, h2: con})
	p = replayB(b)
	if got, want := p.Exps[e].Unjudged, []string{h3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Unjudged = %v, want %v", got, want)
	}
	if got := p.UnjudgedAfter(e, map[string]string{h3: inc}); got != nil {
		t.Fatalf("UnjudgedAfter with a verdict for h3 = %v, want nil", got)
	}
	b.Contradicts(h3)
	if got, want := replayB(b).Exps[e].Unjudged, []string{h3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("after h3 is contradicted later: Unjudged = %v, want %v (liveness is read at add)", got, want)
	}
	e2 := b.Exp(h1, event.RecipeNone, h2)
	b.Result(e2, map[string]string{h1: sup})
	if got := replayB(b).Exps[e2].Unjudged; got != nil {
		t.Fatalf("h2 was contradicted before e2's add: Unjudged = %v, want nil", got)
	}
	h4 := b.Hyp(q, "h4")
	e3 := b.Exp(h1, event.RecipeNone)
	if !replayB(b).Exps[e3].OneRival {
		t.Fatalf("e3 names no target while %s is live: want one-rival", h4)
	}
}

// why sets the why of b's last event and returns its id.
func why(b *testutil.B, why string) string {
	b.Evs[len(b.Evs)-1].Why = why
	return b.Evs[len(b.Evs)-1].ID
}

func currentSelect(t *testing.T, p Projection, q string) SelectRec {
	t.Helper()
	wantNoRejections(t, p)
	rec, ok := p.CurrentSelect(q)
	if !ok {
		t.Fatalf("%s has no current select", q)
	}
	return rec
}

func TestSelectRecKeepsScopeChangeAndWhy(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	b.Emit(event.TypeDQSelect, q, event.DQSelect{Hypothesis: h, ScopeChange: true})
	sel := why(b, "pick h")
	rec := currentSelect(t, replayB(b), q)
	if rec.ID != sel || rec.Hyp != h || !rec.ScopeChange || rec.Why != "pick h" {
		t.Fatalf("current select = %+v, want %s on %s, scope-change, why %q", rec, sel, h, "pick h")
	}
}

func TestDecisionKeepsLatestOwnerConfirm(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindHypothesis, Parent: q, Title: "h", Body: "old"})
	b.Supports(h)
	sel := b.As("a").Select(q, h, "", true)
	b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmInsufficient)
	why(b, "thin")
	if c := replayB(b).Decisions[q].Latest(event.RoleOwner); c.Verdict != event.ConfirmInsufficient || c.Why != "thin" {
		t.Fatalf("after an insufficient confirm: %+v, want insufficient %q", c, "thin")
	}
	b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmSufficient)
	why(b, "good")
	if c := replayB(b).Decisions[q].Latest(event.RoleOwner); c.Verdict != event.ConfirmSufficient || c.Why != "good" {
		t.Fatalf("after a sufficient confirm: %+v, want sufficient %q", c, "good")
	}

	// An amend stops the sufficient confirm counting, but the record keeps it:
	// hyp show still prints the latest owner confirm.
	amend(b.As("a"), h, "new")
	dec := replayB(b).Decisions[q]
	if c := dec.Counting(event.RoleOwner); c != nil {
		t.Fatalf("after an amend a confirm still counts: %+v", c)
	}
	if c := dec.Latest(event.RoleOwner); c.Verdict != event.ConfirmSufficient || c.Why != "good" {
		t.Fatalf("after an amend: %+v, want sufficient %q kept", c, "good")
	}
}

func TestConfirmOnEarlierSelectChangesOnlyIt(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	s1 := b.As("a").Select(q, h1, "", true)
	s2 := b.Select(q, h2, s1, true)
	b.As("o").Confirm(q, s1, event.RoleOwner, event.ConfirmInsufficient)
	why(b, "late")
	p := replayB(b)
	if rec, dec := currentSelect(t, p, q), p.Decisions[q]; rec.ID != s2 || dec.Latest(event.RoleOwner) != nil {
		t.Fatalf("current select = %+v with owner %+v, want %s with none", rec, dec.Latest(event.RoleOwner), s2)
	}
	if cs := p.Decisions[q].Confirms[s1]; len(cs) != 1 || cs[0].Verdict != event.ConfirmInsufficient || cs[0].Why != "late" {
		t.Fatalf("earlier select's confirms = %+v, want insufficient %q", cs, "late")
	}
}

func TestCurrentSelectFalseAfterUnselect(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	if _, ok := replayB(b).CurrentSelect(q); ok {
		t.Fatal("no select yet: want ok false")
	}
	sel := b.As("a").Select(q, h, "", false)
	currentSelect(t, replayB(b), q)
	b.Unselect(q, sel)
	p := replayB(b)
	wantNoRejections(t, p)
	if rec, ok := p.CurrentSelect(q); ok {
		t.Fatalf("after an unselect: current select %+v, want none", rec)
	}
}

func TestLiveEdgesDropRemovedKeepOrder(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2, h3 := b.Hyp(q, "h1"), b.Hyp(q, "h2"), b.Hyp(q, "h3")
	e1 := b.Edge(h1, event.EdgeDependsOn, h2)
	b.Edge(h2, event.EdgeDependsOn, h3)
	e3 := b.Edge(h1, event.EdgeDependsOn, h3)
	b.Unedge(h2, event.EdgeDependsOn, h3)
	p := replayB(b)
	wantNoRejections(t, p)
	var got []string
	for _, e := range p.LiveEdges() {
		got = append(got, e.ID)
	}
	if want := []string{e1, e3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LiveEdges = %v, want %v", got, want)
	}
}

// An Experiment's decided_by is kept on its node; one written without it,
// as every Experiment before the field was, replays exactly as before.
func TestDecidedByReplays(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	old := b.Exp(h, event.RecipeNone)
	read := b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindExperiment, Parent: h, Title: "e",
		Recipe: event.RecipeNone, DecidedBy: event.DecidedReading})
	st, p := Build(b.Evs, nil, "")
	wantNoRejections(t, p)
	if got := st.Nodes[old].DecidedBy; got != "" {
		t.Fatalf("Experiment with no decided_by: decided_by %q, want none", got)
	}
	if got := st.Nodes[read].DecidedBy; got != event.DecidedReading {
		t.Fatalf("declared Experiment: decided_by %q, want %q", got, event.DecidedReading)
	}
	if a, b := statuses(p)[old], statuses(p)[read]; a.Status != b.Status || a.Reason != b.Reason {
		t.Fatalf("statuses differ: %+v and %+v", a, b)
	}
}

// Exps holds each Experiment's status, its non-retracted Results and their
// verdicts, a later Result winning; Status agrees with it.
func TestExpRecordFollowsLiveResults(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	e := b.Exp(h1, event.RecipeNone, h2)
	if x := replayB(b).Exps[e]; x.Status != ExpPlanned || x.Live != nil || len(x.Judged) != 0 {
		t.Fatalf("planned: %+v", x)
	}
	r1 := b.Result(e, map[string]string{h1: sup, h2: con})
	r2 := b.Result(e, map[string]string{h1: inc})
	p := replayB(b)
	x := p.Exps[e]
	if x.Status != ExpRun || statuses(p)[e].Status != string(x.Status) {
		t.Fatalf("status = %q (Status %q), want run", x.Status, statuses(p)[e].Status)
	}
	if !reflect.DeepEqual(x.Live, []string{r1, r2}) {
		t.Fatalf("Live = %v, want %v", x.Live, []string{r1, r2})
	}
	if want := map[string]string{h1: inc, h2: con}; !reflect.DeepEqual(x.Judged, want) {
		t.Fatalf("Judged = %v, want %v", x.Judged, want)
	}
	b.Retract(r1)
	b.Retract(r2)
	if x := replayB(b).Exps[e]; x.Status != ExpPlanned || x.Live != nil || x.Unjudged != nil {
		t.Fatalf("every Result retracted: %+v, want planned with nothing live", x)
	}
}
