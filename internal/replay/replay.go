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
		Status:      st.Derived,
		Goal:        st.goalStatus(),
		Applied:     st.Applied,
		Limits:      st.limits(),
		RivalsAtAdd: st.RivalsAtAdd,
		Exps:        st.exps(),
		Results:     st.results(),
		Edges:       st.Edges,
		Selection:   st.Selection,
		Dossiers:    st.dossiers(),
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
	for id, ns := range st.Derived {
		if ns.Status == DQAnswered && st.Nodes[id].Kind == event.KindDesignQuestion {
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
	switch ev.Type {
	case event.TypeNodeAdd:
		var p event.NodeAdd
		_ = event.DecodePayload(ev, &p)
		if p.Kind == event.KindExperiment {
			// Derived still holds the state before this event.
			st.RivalsAtAdd[ev.ID] = st.LiveRivals(st.Nodes[p.Parent].Parent, p.Parent)
		}
		st.addNode(&Node{ID: ev.ID, Kind: p.Kind, Parent: p.Parent, Title: p.Title, Body: p.Body,
			Alias: p.Alias, Recipe: p.Recipe, RecipeHash: p.RecipeHash, DecidedBy: p.DecidedBy, Actor: ev.Actor,
			AddedBy: ev.ID, Targets: p.Targets})
		if p.Alias != "" {
			st.Aliases[p.Alias] = ev.ID
		}
		if p.Kind == event.KindGoal {
			st.GoalID = ev.ID
		}
		if p.Kind == event.KindDesignQuestion {
			st.Selection[ev.ID] = newSel()
		}
	case event.TypeResultRecord:
		var p event.ResultRecord
		_ = event.DecodePayload(ev, &p)
		st.addNode(&Node{ID: ev.ID, Kind: event.KindResult, Parent: p.Parent, Actor: ev.Actor,
			AddedBy: ev.ID, Verdicts: p.Verdicts, VerdictOrder: event.VerdictOrder(ev.Payload), Observation: p.Observation, Evidence: p.Evidence,
			EvidenceHashes: p.EvidenceHashes, Limits: p.Limits, Closes: p.Closes})
		for h := range p.Verdicts {
			st.ResultsByHyp[h] = append(st.ResultsByHyp[h], ev.ID)
		}
	case event.TypeResultRetract:
		st.Retracted[ev.Node] = true
	case event.TypeHypothesisAbandon:
		st.Abandoned[ev.Node] = true
	case event.TypeHypothesisAmend:
		var p event.HypothesisAmend
		_ = event.DecodePayload(ev, &p)
		n := st.Nodes[ev.Node]
		n.Body = p.Body
		// The owner confirm on the current select and its acceptances judged
		// the old body and stop counting; onAmend holds the verifier rules.
		if sel := st.Selection[n.Parent]; sel != nil && sel.Current != "" && sel.Selects[sel.Current].Hyp == ev.Node {
			delete(st.Confirms, sel.Current+"|"+event.RoleOwner)
			st.dropAcceptances(sel.Current)
			sel.onAmend()
		}
	case event.TypeDQSelect:
		var p event.DQSelect
		_ = event.DecodePayload(ev, &p)
		cands := st.candidates(ev.Node, func(c string) string { return st.Derived[c].Status },
			func(c string) bool { return st.Derived[c].Stale != nil })
		sel := st.Selection[ev.Node]
		sel.onSelect(sel.Current != "" && (p.Hypothesis == sel.Selects[sel.Current].Hyp || st.supersedes(p.Hypothesis, sel.Selects[sel.Current].Hyp)))
		sel.Selects[ev.ID] = SelectRec{ID: ev.ID, Hyp: p.Hypothesis, Actor: ev.Actor,
			OwnerRequired: len(cands) >= 2 || p.Tradeoff || p.ScopeChange, Tradeoff: p.Tradeoff,
			ScopeChange: p.ScopeChange, Why: ev.Why}
		sel.setCurrent(ev.ID, ev.ID, p.PrevSelect)
	case event.TypeDQUnselect:
		var p event.DQUnselect
		_ = event.DecodePayload(ev, &p)
		sel := st.Selection[ev.Node]
		sel.setCurrent(ev.ID, "", p.PrevSelect)
		sel.onUnselect()
	case event.TypeDQDefer:
		st.Hold[ev.Node] = DQDeferred
	case event.TypeDQWithdraw:
		st.Hold[ev.Node] = DQWithdrawn
	case event.TypeExperimentWithdraw:
		st.Withdrawn[ev.Node] = true
	case event.TypeDQResume:
		st.Hold[ev.Node] = ""
	case event.TypeConfirm:
		var p event.Confirm
		_ = event.DecodePayload(ev, &p)
		if p.Role == event.RoleVerifier {
			st.Selection[ev.Node].onConfirm(ev, p.Select, p.Verdict)
		}
		if p.Role == event.RoleOwner {
			st.Confirms[p.Select+"|"+p.Role] = p.Verdict
			// An owner's no withdraws their sign-off on this select's limits. Only
			// a sufficient owner confirm carries accepts, so nothing is re-added.
			if p.Verdict == event.ConfirmInsufficient {
				st.dropAcceptances(p.Select)
			}
			for _, k := range p.Accepts {
				st.Accepted[p.Select+"|"+k] = Acceptance{Select: p.Select, Actor: ev.Actor}
			}
		}
	case event.TypeDQRequestVerify:
		var p event.DQRequestVerify
		_ = event.DecodePayload(ev, &p)
		st.Selection[ev.Node].onRequest(ev, p.Select)
	case event.TypeEdgeAdd:
		var p event.Edge
		_ = event.DecodePayload(ev, &p)
		st.EdgesFrom[ev.Node] = append(st.EdgesFrom[ev.Node], len(st.Edges))
		st.Edges = append(st.Edges, Edge{ID: ev.ID, From: ev.Node, Type: p.Type, To: p.To, Live: true})
		if p.Type == event.EdgeSupersedes {
			st.SupersededN[p.To]++
		}
	case event.TypeEdgeRemove:
		var p event.Edge
		_ = event.DecodePayload(ev, &p)
		st.Edges[st.liveEdge(ev.Node, p.Type, p.To)].Live = false
		if p.Type == event.EdgeSupersedes {
			st.SupersededN[p.To]--
		}
	}
}

func (st *State) addNode(n *Node) {
	st.Nodes[n.ID] = n
	st.Order = append(st.Order, n.ID)
	if parent := st.Nodes[n.Parent]; parent != nil {
		parent.Children = append(parent.Children, n.ID)
	}
}

// limits derives each question's limits. A limit is closed when a
// non-retracted Result closes it, accepted when an owner confirm on the
// question's current select accepted it, and open otherwise. Limits of a
// retracted Result are left out.
func (st *State) limits() map[string][]Limit {
	closedBy := map[string]string{}
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] {
			continue
		}
		for _, k := range n.Closes {
			if _, ok := closedBy[k]; !ok {
				closedBy[k] = id
			}
		}
	}
	out := map[string][]Limit{}
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] || len(n.Limits) == 0 {
			continue
		}
		q := st.resultQuestion(id)
		for i, text := range n.Limits {
			l := Limit{Key: event.LimitKey(id, i+1), Result: id, Text: text, Hyps: sortedKeys(n.Verdicts), State: LimitOpen}
			if by, ok := closedBy[l.Key]; ok {
				l.State, l.By = LimitClosed, by
			} else if a, ok := st.Accepted[st.Selection[q].Current+"|"+l.Key]; ok {
				l.State, l.By = LimitAccepted, a.Actor
			}
			if cur := st.Selection[q].Current; cur != "" {
				l.Gating = l.Gates(st.Selection[q].Selects[cur].Hyp)
			}
			out[q] = append(out[q], l)
		}
	}
	return out
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

// setCurrent applies the conflict rule of spec §7.4 for the select or
// unselect event evID: a write whose prev_select is not the current selection
// records both as conflicting; a write that names the current selection clears
// the conflict set. The write still applies and current becomes next.
func (s *Sel) setCurrent(evID, next, prev string) {
	if prev == s.Current {
		s.Conflicts = nil
	} else {
		if s.Current != "" {
			s.Conflicts = appendUnique(s.Conflicts, s.Current)
		}
		s.Conflicts = appendUnique(s.Conflicts, evID)
	}
	s.Current = next
}

// basisNodes lists the nodes for which ev is a basis event (spec §7.3).
func (st *State) basisNodes(ev event.Event) []string {
	switch ev.Type {
	case event.TypeResultRecord:
		var out []string
		for _, h := range sortedKeys(st.Nodes[ev.ID].Verdicts) {
			if st.Nodes[ev.ID].Verdicts[h] == event.VerdictSupports {
				out = append(out, h)
			}
		}
		return out
	case event.TypeDQSelect, event.TypeReaffirm:
		return []string{ev.Node}
	case event.TypeEdgeAdd:
		if st.Edges[len(st.Edges)-1].Type == event.EdgeDependsOn {
			return []string{ev.Node}
		}
	}
	return nil
}

// snapshot records (status, stale) of every live depends_on target of n.
func (st *State) snapshot(n, eventID string) {
	targets := map[string]SnapVal{}
	for _, t := range st.dependsOn(n) {
		d := st.Derived[t]
		targets[t] = SnapVal{Status: d.Status, Stale: d.Stale != nil}
	}
	st.Snaps[n] = append(st.Snaps[n], Snap{EventID: eventID, Targets: targets})
}

// dependsOn lists n's live depends_on targets in edge-add order.
func (st *State) dependsOn(n string) []string {
	var out []string
	for _, i := range st.EdgesFrom[n] {
		if e := st.Edges[i]; e.Live && e.Type == event.EdgeDependsOn {
			out = append(out, e.To)
		}
	}
	return out
}

// LiveHypStatus reports whether a Hypothesis with derived status s is live
// (spec §7.1): neither superseded nor abandoned. A contradicted Hypothesis is
// still live.
func LiveHypStatus(s string) bool { return s != HypSuperseded && s != HypAbandoned }

// RivalStatus reports whether a Hypothesis with derived status s is a live
// rival (spec §7.1): live and not contradicted.
func RivalStatus(s string) bool { return LiveHypStatus(s) && s != HypContradicted }

// NotRivals lists the Hypotheses in ids, once each in first-seen order,
// whose current derived status is not a live rival. The D1 signal and the
// hyp add exp warning both read it, so they cannot disagree.
func (st *State) NotRivals(ids []string) []string {
	var out, seen []string
	for _, h := range ids {
		if !slices.Contains(seen, h) {
			seen = append(seen, h)
			if !RivalStatus(st.Derived[h].Status) {
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
		if c != except && st.Nodes[c].Kind == event.KindHypothesis && RivalStatus(st.Derived[c].Status) {
			out = append(out, c)
		}
	}
	return out
}

// candidates lists the trade-off candidates of DesignQuestion q (spec §7.4):
// live child Hypotheses that are supported and not stale. They are the
// Suggested set, and two or more of them make a select owner_required.
// status and stale read the derivation in progress or the last one.
func (st *State) candidates(q string, status func(string) string, stale func(string) bool) []string {
	var out []string
	for _, c := range st.Nodes[q].Children {
		if st.isLiveHyp(c) && status(c) == HypSupported && !stale(c) {
			out = append(out, c)
		}
	}
	return out
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
	hyp     map[string]string
	dq      map[string]NodeStatus
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

// recompute rebuilds st.Derived from the stored state.
func (st *State) recompute(recipeHashes map[string]string) {
	st.recipeHashes = recipeHashes
	d := &derivation{st: st, hyp: map[string]string{}, dq: map[string]NodeStatus{},
		stale: map[string]*StaleCause{}, staleOK: map[string]bool{}, busy: map[string]bool{}}
	out := make(map[string]NodeStatus, len(st.Order))
	for _, id := range st.Order {
		n := st.Nodes[id]
		switch n.Kind {
		case event.KindGoal:
			out[id] = NodeStatus{Status: st.goalState(d)}
		case event.KindDesignQuestion:
			out[id] = d.dqState(id)
		case event.KindHypothesis:
			out[id] = NodeStatus{Status: d.hypStatus(id), Stale: d.staleOf(id)}
		case event.KindExperiment:
			out[id] = NodeStatus{Status: st.expStatus(id)}
		case event.KindResult:
			status := ResRecorded
			if st.Retracted[id] {
				status = ResRetracted
			}
			out[id] = NodeStatus{Status: status}
		}
	}
	st.Derived = out
}

// hypStatus follows spec §7.1.
func (d *derivation) hypStatus(id string) string {
	if s, ok := d.hyp[id]; ok {
		return s
	}
	if !d.enter("h|" + id) {
		return HypProposed
	}
	st := d.st
	var s string
	switch {
	case st.superseded(id):
		s = HypSuperseded
	case st.Abandoned[id]:
		s = HypAbandoned
	default:
		s = st.verdictStatus(id, "")
		if s == HypSupported && st.awaitsSubquestions(id, func(c string) string { return d.dqState(c).Status }) {
			s = HypAwaitingSubquestions
		}
	}
	d.hyp[id] = s
	return s
}

// verdictStatus is rule 3 of spec §7.1: the status of Hypothesis id from the
// verdicts of its non-retracted Results, leaving out the Result except when it
// is not "".
func (st *State) verdictStatus(id, except string) string {
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

// awaitsSubquestions is rule 4 of spec §7.1: a child DQ of Hypothesis id that
// is not withdrawn is not answered. status reads a question's state.
func (st *State) awaitsSubquestions(id string, status func(string) string) bool {
	for _, c := range st.Nodes[id].Children {
		if st.Nodes[c].Kind == event.KindDesignQuestion && st.isLiveDQ(c) && status(c) != DQAnswered {
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
				cause = &StaleCause{Node: t, Reason: ts}
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
func (d *derivation) dqState(id string) NodeStatus {
	if s, ok := d.dq[id]; ok {
		return s
	}
	if !d.enter("q|" + id) {
		return NodeStatus{Status: DQOpen}
	}
	st := d.st
	ns := NodeStatus{Stale: d.staleOf(id)}
	switch st.Hold[id] {
	case DQWithdrawn:
		ns = NodeStatus{Status: DQWithdrawn}
		d.dq[id] = ns
		return ns
	case DQDeferred:
		ns.Status = DQDeferred
	}

	var live []string
	for _, c := range st.Nodes[id].Children {
		if st.isLiveHyp(c) {
			live = append(live, c)
		}
	}
	ns.Suggested = st.candidates(id, d.hypStatus, func(c string) bool { return d.staleOf(c) != nil })
	if ns.Status == DQDeferred {
		d.dq[id] = ns
		return ns
	}

	sel := st.Selection[id]
	reason := ""
	if len(sel.Conflicts) > 0 {
		reason = ReasonConflicting
	} else if sel.Current != "" {
		rec := sel.Selects[sel.Current]
		switch {
		case d.hypStatus(rec.Hyp) != HypSupported:
			reason = ReasonSelectedNotSupport
		case d.staleOf(rec.Hyp) != nil || ns.Stale != nil:
			reason = ReasonSelectedStale
		case sel.Verdict(rec.ID) == event.ConfirmInsufficient && !sel.Requested(rec.ID):
			reason = ReasonVerifierInsufficient
		case sel.Verdict(rec.ID) != event.ConfirmSufficient:
			reason = ReasonAwaitingVerifier
		case rec.OwnerRequired && st.Confirms[rec.ID+"|"+event.RoleOwner] != event.ConfirmSufficient:
			reason = ReasonAwaitingOwner
		}
		if reason == "" {
			ns.Status = DQAnswered
			d.dq[id] = ns
			return ns
		}
	} else {
		reason = d.candidateReason(live)
	}
	ns.Status = DQOpen
	if st.EverAnswered[id] {
		ns.Status = DQReopened
	}
	ns.Reason = reason
	d.dq[id] = ns
	return ns
}

// candidateReason gives the open reason of a DQ with no current selection.
func (d *derivation) candidateReason(live []string) string {
	if len(live) == 0 {
		return ReasonNoCandidates
	}
	count := map[string]int{}
	for _, c := range live {
		count[d.hypStatus(c)]++
	}
	switch {
	case count[HypProposed] == len(live):
		return ReasonUntested
	case count[HypContradicted] == len(live):
		return ReasonAllContradicted
	case count[HypSupported] == 0:
		return ReasonInconclusive
	}
	return ReasonSupportedUnsel
}

// goalState follows spec §7.5, with one addition: a Goal needs at least one
// answered live top-level DQ to be complete, so a new Goal is in_progress.
func (st *State) goalState(d *derivation) string {
	g := st.Nodes[st.GoalID]
	if g == nil {
		return ""
	}
	answered := 0
	for _, c := range g.Children {
		if !st.isLiveDQ(c) {
			continue
		}
		if d.dqState(c).Status != DQAnswered {
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
		DQOpen: 0, DQReopened: 0, DQAnswered: 0, DQDeferred: 0, DQWithdrawn: 0, StaleReasonStale: 0,
	}}
	if st.GoalID == "" {
		return gs
	}
	gs.Status = st.Derived[st.GoalID].Status
	for _, id := range st.Order {
		n := st.Nodes[id]
		ns := st.Derived[id]
		if n.Kind == event.KindDesignQuestion {
			gs.Counts[ns.Status]++
		}
		if ns.Stale != nil && (st.isLiveHyp(id) || st.isLiveDQ(id)) {
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

// dropAcceptances forgets the limit acceptances recorded on sel.
func (st *State) dropAcceptances(sel string) {
	for k, a := range st.Accepted {
		if a.Select == sel {
			delete(st.Accepted, k)
		}
	}
}
