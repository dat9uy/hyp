package replay

import (
	"slices"
	"sort"

	"github.com/dat9uy/hyp/internal/event"
)

// Merge combines per-actor event files by a stable k-way merge: it repeatedly
// takes the head event with the smallest id, breaking ties by file name. Order
// inside one file is never changed.
func Merge(files map[string][]event.Event) []event.Event {
	names := make([]string, 0, len(files))
	total := 0
	for name, evs := range files {
		names = append(names, name)
		total += len(evs)
	}
	sort.Strings(names)
	cursor := make([]int, len(names))
	out := make([]event.Event, 0, total)
	for len(out) < total {
		best := -1
		for i, name := range names {
			if cursor[i] >= len(files[name]) {
				continue
			}
			if best < 0 || files[name][cursor[i]].ID < files[names[best]][cursor[best]].ID {
				best = i
			}
		}
		out = append(out, files[names[best]][cursor[best]])
		cursor[best]++
	}
	return out
}

// Replay applies events in order and returns the derived projection. recipeHashes
// maps each recipe ref (name@version) to its current hash. Refused events are
// skipped and reported. When asOf is not empty, replay stops after the event
// with that id.
func Replay(events []event.Event, recipeHashes map[string]string, asOf string) Projection {
	_, p := Build(events, recipeHashes, asOf)
	return p
}

// Build is Replay that also returns the final State, which Validate needs to
// check a new event before it is appended.
func Build(events []event.Event, recipeHashes map[string]string, asOf string) (*State, Projection) {
	st := NewState()
	var rejections []Rejection
	var lastCheckpoint *event.Event
	for _, ev := range events {
		if err := st.Step(ev, recipeHashes); err != nil {
			rejections = append(rejections, Rejection{EventID: ev.ID, Reason: err.Error()})
		} else if ev.Type == event.TypeCheckpoint {
			cp := ev
			lastCheckpoint = &cp
		}
		if asOf != "" && ev.ID == asOf {
			break
		}
	}
	p := st.Projection()
	p.Rejections, p.LastCheckpoint = rejections, lastCheckpoint
	return st, p
}

// Projection derives the projection of the state as stepped so far. It has
// no Rejections or LastCheckpoint: only the caller that stepped knows those.
func (st *State) Projection() Projection {
	return Projection{
		Nodes:       st.Nodes,
		Hyps:        st.Hyps,
		Questions:   st.Questions,
		Goal:        st.goalStatus(),
		Applied:     st.Applied,
		RivalsAtAdd: st.RivalsAtAdd,
		Exps:        st.exps(),
		Results:     st.results(),
		Edges:       st.Edges,
		Decisions:   st.decisions(),
		Readings:    st.readings(st.Applied),
		Lines:       st.lines(),
		Subjects:    st.Subjects,
	}
}

// Step validates ev and, when it is allowed, applies it and recomputes every
// derived field (spec §7 steps 1–5).
func (st *State) Step(ev event.Event, recipeHashes map[string]string) error {
	if err := Validate(st, ev); err != nil {
		return err
	}
	st.apply(ev)
	st.recompute(recipeHashes)
	if basis := st.basisNodes(ev); len(basis) > 0 {
		for _, n := range basis {
			st.snapshot(n, ev.ID)
		}
		st.recompute(recipeHashes)
	}
	for id, q := range st.Questions {
		if q.State == DQAnswered {
			st.EverAnswered[id] = true
		}
	}
	st.Applied = append(st.Applied, ev)
	return nil
}

// apply changes stored state only; every derived field is left to recompute.
// Payloads were already decoded successfully by Validate.
func (st *State) apply(ev event.Event) {
	// Only applied ids are reserved: a refused event never enters the state,
	// so its id stays free.
	st.SeenIDs[ev.ID] = true
	rules[ev.Type].apply(st, ev)
}

func (st *State) addNode(n *Node) {
	st.Nodes[n.ID] = n
	st.Order = append(st.Order, n.ID)
	if parent := st.Nodes[n.Parent]; parent != nil {
		parent.Children = append(parent.Children, n.ID)
	}
}

// supersedes reports whether h reaches old through live supersedes edges.
func (st *State) supersedes(h, old string) bool {
	seen := map[string]bool{h: true}
	for queue := []string{h}; len(queue) > 0; queue = queue[1:] {
		for _, i := range st.EdgesFrom[queue[0]] {
			e := st.Edges[i]
			if !e.Live || e.Type != event.EdgeSupersedes || seen[e.To] {
				continue
			}
			if e.To == old {
				return true
			}
			seen[e.To] = true
			queue = append(queue, e.To)
		}
	}
	return false
}

// basisNodes lists the nodes for which ev is a basis event (spec §7.3).
func (st *State) basisNodes(ev event.Event) []string {
	if basis := rules[ev.Type].basis; basis != nil {
		return basis(&st.Facts, ev)
	}
	return nil
}

// snapshot records (status, stale) of every live depends_on target of n.
func (st *State) snapshot(n, eventID string) {
	targets := map[string]SnapVal{}
	for _, t := range st.dependsOn(n) {
		var v SnapVal
		if h := st.Hyps[t]; h != nil {
			v = SnapVal{Status: h.Status, Stale: h.Stale != nil}
		}
		targets[t] = v
	}
	st.Snaps[n] = append(st.Snaps[n], Snap{EventID: eventID, Targets: targets})
}

// dependsOn lists n's live depends_on targets in edge-add order.
func (f *Facts) dependsOn(n string) []string {
	var out []string
	for _, i := range f.EdgesFrom[n] {
		if e := f.Edges[i]; e.Live && e.Type == event.EdgeDependsOn {
			out = append(out, e.To)
		}
	}
	return out
}

// LiveHypStatus reports whether a Hypothesis with derived status s is live
// (spec §7.1): neither superseded nor abandoned. A contradicted Hypothesis is
// still live.
func LiveHypStatus(s HypStatus) bool { return s != HypSuperseded && s != HypAbandoned }

// RivalStatus reports whether a Hypothesis with derived status s is a live
// rival (spec §7.1): live and not contradicted.
func RivalStatus(s HypStatus) bool { return LiveHypStatus(s) && s != HypContradicted }

// NotRivals lists the Hypotheses in ids, once each in first-seen order,
// whose current derived status is not a live rival. The D1 signal and the
// hyp add exp warning both read it, so they cannot disagree.
func (st *State) NotRivals(ids []string) []string {
	var out, seen []string
	for _, h := range ids {
		if !slices.Contains(seen, h) {
			seen = append(seen, h)
			if !RivalStatus(st.hypWord(h)) {
				out = append(out, h)
			}
		}
	}
	return out
}

// Contradictions lists the Results that are not retracted and contradict
// Hypothesis h, in creation order.
func (st *State) Contradictions(h string) []string {
	var out []string
	for _, r := range st.resultsFor(h) {
		if st.Nodes[r].Verdicts[h] == event.VerdictContradicts {
			out = append(out, r)
		}
	}
	return out
}

// Superseders lists the nodes (Hypotheses, or Experiments for an
// Experiment) with a live supersedes edge to h, in edge-add order.
func (st *State) Superseders(h string) []string {
	var out []string
	for _, e := range st.Edges {
		if e.Live && e.Type == event.EdgeSupersedes && e.To == h {
			out = append(out, e.From)
		}
	}
	return out
}

// LiveRivals lists the child Hypotheses of DesignQuestion q, other than
// except, whose current derived status is a live rival, in creation order.
// The untargeted-rival refusal and the one-rival signal read it.
func (st *State) LiveRivals(q, except string) []string {
	var out []string
	for _, c := range st.Nodes[q].Children {
		if c != except && st.Nodes[c].Kind == event.KindHypothesis && RivalStatus(st.hypWord(c)) {
			out = append(out, c)
		}
	}
	return out
}

// candidates lists the candidates of DesignQuestion q (spec §7.4):
// live child Hypotheses that rule 3 of §7.1 gives supported. It reads only
// facts the Dossier lists, so the owner requirement changes only when the
// Dossier does.
func (st *State) candidates(q string) []string {
	var out []string
	for _, c := range st.Nodes[q].Children {
		if st.isLiveHyp(c) && st.verdictStatus(c, "") == HypSupported {
			out = append(out, c)
		}
	}
	return out
}

// hypWord is Hypothesis id's derived status, or "" for any other id.
func (st *State) hypWord(id string) HypStatus {
	if h := st.Hyps[id]; h != nil {
		return h.Status
	}
	return ""
}

// isLiveHyp is LiveHypStatus read from stored state, for use while the
// derivation that sets the status is still running.
func (st *State) isLiveHyp(id string) bool {
	n := st.Nodes[id]
	return n != nil && n.Kind == event.KindHypothesis && !st.superseded(id) && !st.Abandoned[id]
}

func (st *State) isLiveDQ(id string) bool {
	n := st.Nodes[id]
	return n != nil && n.Kind == event.KindDesignQuestion && st.Hold[id] != DQWithdrawn
}

// derivation holds the memo of one recompute call.
type derivation struct {
	st      *State
	hyp     map[string]HypStatus
	dq      map[string]*Question
	stale   map[string]*StaleCause
	staleOK map[string]bool
	// busy guards against a cycle that Validate should have refused: a node
	// met again while being derived gets a neutral value instead of looping.
	busy map[string]bool
}

func (d *derivation) enter(key string) bool {
	if d.busy[key] {
		return false
	}
	d.busy[key] = true
	return true
}

// recompute rebuilds the derived records from the stored state.
func (st *State) recompute(recipeHashes map[string]string) {
	st.recipeHashes = recipeHashes
	d := &derivation{st: st, hyp: map[string]HypStatus{}, dq: map[string]*Question{},
		stale: map[string]*StaleCause{}, staleOK: map[string]bool{}, busy: map[string]bool{}}
	hyps, questions := map[string]*Hyp{}, map[string]*Question{}
	for _, id := range st.Order {
		n := st.Nodes[id]
		switch n.Kind {
		case event.KindGoal:
			st.goal = st.goalState(d)
		case event.KindDesignQuestion:
			questions[id] = d.dqState(id)
		case event.KindHypothesis:
			hyps[id] = &Hyp{Status: d.hypStatus(id), Stale: d.staleOf(id)}
		}
	}
	st.Hyps, st.Questions = hyps, questions
}

// hypStatus follows spec §7.1.
func (d *derivation) hypStatus(id string) HypStatus {
	if s, ok := d.hyp[id]; ok {
		return s
	}
	if !d.enter("h|" + id) {
		return HypProposed
	}
	st := d.st
	var s HypStatus
	switch {
	case st.superseded(id):
		s = HypSuperseded
	case st.Abandoned[id]:
		s = HypAbandoned
	default:
		s = st.verdictStatus(id, "")
	}
	d.hyp[id] = s
	return s
}

// verdictStatus is rule 3 of spec §7.1: the status of Hypothesis id from the
// verdicts of its non-retracted Results, leaving out the Result except when it
// is not "".
func (st *State) verdictStatus(id, except string) HypStatus {
	var supports, contradicts, any bool
	for _, r := range st.resultsFor(id) {
		if r == except {
			continue
		}
		any = true
		switch st.Nodes[r].Verdicts[id] {
		case event.VerdictSupports:
			supports = true
		case event.VerdictContradicts:
			contradicts = true
		}
	}
	switch {
	case !any:
		return HypProposed
	case contradicts:
		return HypContradicted
	case supports:
		return HypSupported
	}
	return HypInconclusive
}

// awaitsSubquestions is the obligation of spec §7.4: a sub-question of Hypothesis
// id that is not withdrawn is not answered. status reads a question's state.
func (st *State) awaitsSubquestions(id string, status func(string) QuestionState) bool {
	for _, c := range st.Nodes[id].Children {
		if st.Nodes[c].Kind == event.KindDesignQuestion && st.isLiveDQ(c) && status(c) != DQAnswered {
			return true
		}
	}
	return false
}

// HasOpenSubquestion reports whether Hypothesis h has a sub-question that is
// not withdrawn and not answered, so a deferred one counts (spec §7.4).
func (p Projection) HasOpenSubquestion(h string) bool {
	for _, c := range p.Nodes[h].Children {
		if q := p.Questions[c]; q != nil && q.State != DQWithdrawn && q.State != DQAnswered {
			return true
		}
	}
	return false
}

// resultsFor lists non-retracted Results with a verdict for hypothesis h, in
// creation order.
func (st *State) resultsFor(h string) []string {
	var out []string
	for _, id := range st.ResultsByHyp[h] {
		if !st.Retracted[id] {
			out = append(out, id)
		}
	}
	return out
}

// staleOf follows spec §7.3.
func (d *derivation) staleOf(id string) *StaleCause {
	if d.staleOK[id] {
		return d.stale[id]
	}
	if !d.enter("s|" + id) {
		return nil
	}
	st := d.st
	var cause *StaleCause
	if st.isLiveHyp(id) || st.isLiveDQ(id) {
		for _, t := range st.dependsOn(id) {
			ts := d.hypStatus(t)
			if ts != HypSupported {
				cause = &StaleCause{Node: t, Reason: string(ts)}
				if ts == HypContradicted {
					cause.Next = &StaleCause{Node: st.latestContradiction(t), Reason: StaleReasonContradicts}
				}
				break
			}
			if next := d.staleOf(t); next != nil {
				cause = &StaleCause{Node: t, Reason: StaleReasonStale, Next: next}
				break
			}
			if snap, ok := st.effectiveSnap(id, t); ok && snap != (SnapVal{Status: HypSupported}) {
				cause = &StaleCause{Node: t, Reason: StaleReasonChanged}
				break
			}
		}
	}
	d.stale[id] = cause
	d.staleOK[id] = true
	return cause
}

// effectiveSnap returns the snapshot of target t from n's latest in-force
// basis event that recorded t. A retracted Result is no longer in force.
func (st *State) effectiveSnap(n, t string) (SnapVal, bool) {
	snaps := st.Snaps[n]
	for i := len(snaps) - 1; i >= 0; i-- {
		if st.Retracted[snaps[i].EventID] {
			continue
		}
		if v, ok := snaps[i].Targets[t]; ok {
			return v, true
		}
	}
	return SnapVal{}, false
}

func (st *State) latestContradiction(h string) string {
	if c := st.Contradictions(h); len(c) > 0 {
		return c[len(c)-1]
	}
	return ""
}

// dqState follows spec §7.4.
func (d *derivation) dqState(id string) *Question {
	if s, ok := d.dq[id]; ok {
		return s
	}
	if !d.enter("q|" + id) {
		return &Question{State: DQOpen}
	}
	st := d.st
	q := &Question{Stale: d.staleOf(id)}
	switch st.Hold[id] {
	case DQWithdrawn:
		q = &Question{State: DQWithdrawn}
		d.dq[id] = q
		return q
	case DQDeferred:
		q.State = DQDeferred
	}

	var live []string
	for _, c := range st.Nodes[id].Children {
		if st.isLiveHyp(c) {
			live = append(live, c)
		}
	}
	candidates := st.candidates(id)
	q.Candidates = candidates
	if q.State == DQDeferred {
		d.dq[id] = q
		return q
	}

	reason := d.reason(id, live, candidates, q.Stale != nil)
	if reason == "" {
		q.State = DQAnswered
		d.dq[id] = q
		return q
	}
	q.State = DQOpen
	if st.EverAnswered[id] {
		q.State = DQReopened
	}
	q.Reason = reason
	d.dq[id] = q
	return q
}

// goalState follows spec §7.5, with one addition: a Goal needs at least one
// answered live top-level DQ to be complete, so a new Goal is in_progress.
func (st *State) goalState(d *derivation) GoalState {
	g := st.Nodes[st.GoalID]
	if g == nil {
		return ""
	}
	answered := 0
	for _, c := range g.Children {
		if !st.isLiveDQ(c) {
			continue
		}
		if d.dqState(c).State != DQAnswered {
			return GoalInProgress
		}
		answered++
	}
	if answered == 0 {
		return GoalInProgress
	}
	return GoalComplete
}

func (st *State) goalStatus() GoalStatus {
	gs := GoalStatus{GoalID: st.GoalID, Counts: map[string]int{
		string(DQOpen): 0, string(DQReopened): 0, string(DQAnswered): 0, string(DQDeferred): 0, string(DQWithdrawn): 0, StaleReasonStale: 0,
	}}
	if st.GoalID == "" {
		return gs
	}
	gs.Status = st.goal
	for _, id := range st.Order {
		var stale *StaleCause
		switch st.Nodes[id].Kind {
		case event.KindDesignQuestion:
			q := st.Questions[id]
			gs.Counts[string(q.State)]++
			stale = q.Stale
		case event.KindHypothesis:
			stale = st.Hyps[id].Stale
		}
		if stale != nil && (st.isLiveHyp(id) || st.isLiveDQ(id)) {
			gs.Counts[StaleReasonStale]++
		}
	}
	return gs
}

func appendUnique(list []string, s string) []string {
	if contains(list, s) {
		return list
	}
	return append(list, s)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
