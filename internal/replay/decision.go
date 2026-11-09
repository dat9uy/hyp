package replay

import (
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// Decision is the record of one DesignQuestion's selection (spec §7.4), the
// one owner of its select history, conflicts, owner Confirms, owner
// requirement, limit acceptances and reason order. Only the methods in this
// file write it. Render, signals, check and the CLI read it, through
// Projection.Decisions.
type Decision struct {
	Current   string
	Conflicts []string
	Selects   map[string]SelectRec
	// Confirms holds every applied owner confirm on each select, in log
	// order, each with the Dossier at its place in the log. A verifier
	// confirm in an old log applies nothing (ADR 0015).
	Confirms map[string][]ConfirmRec
	// Limits and Dossier are derived when a Projection is built; a record
	// stored in State holds neither.
	Limits  []Limit  // the limits recorded under the question, in creation order
	Dossier *Dossier // of the current select, or nil when there is none
}

// ConfirmRec is one applied confirm on a select. Items is the Dossier as it
// stood at the confirm's place in the merged log.
type ConfirmRec struct {
	ID, Role, Verdict, Actor, Why string
	Accepts                       []string
	Items                         []DossierItem
}

func newDecision() *Decision {
	return &Decision{Selects: map[string]SelectRec{}, Confirms: map[string][]ConfirmRec{}}
}

// setCurrent applies the conflict rule of spec §7.4 for the select or
// unselect event evID: a write whose prev_select is not the current selection
// records both as conflicting; a write that names the current selection clears
// the conflict set. The write still applies and current becomes next.
func (s *Decision) setCurrent(evID, next, prev string) {
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

// onSelect applies dq.select rec, whose Candidates replay stored at apply.
func (s *Decision) onSelect(rec SelectRec, prev string) {
	s.Selects[rec.ID] = rec
	s.setCurrent(rec.ID, rec.ID, prev)
}

// onUnselect applies a dq.unselect.
func (s *Decision) onUnselect(evID, prev string) { s.setCurrent(evID, "", prev) }

// onConfirm applies owner confirm ev on select p.Select, whose Dossier just
// then was items.
func (s *Decision) onConfirm(ev event.Event, p event.Confirm, items []DossierItem) {
	s.Confirms[p.Select] = append(s.Confirms[p.Select], ConfirmRec{ID: ev.ID, Role: p.Role, Verdict: p.Verdict,
		Actor: ev.Actor, Why: ev.Why, Accepts: p.Accepts, Items: items})
}

// counting is the Confirm by role that counts on select id when the Dossier is
// now: the latest one that judged this Dossier (spec §7.4), or nil. A Confirm
// lapses when the Dossier differs from the one it judged and counts again if
// the Dossier returns to it.
func (s *Decision) counting(id, role string, now []DossierItem) *ConfirmRec {
	cs := s.Confirms[id]
	for i := len(cs) - 1; i >= 0; i-- {
		if cs[i].Role == role && slices.Equal(cs[i].Items, now) {
			return &cs[i]
		}
	}
	return nil
}

// latest is the last Confirm by role on select id, counting or not, or nil.
func (s *Decision) latest(id, role string) *ConfirmRec {
	cs := s.Confirms[id]
	for i := len(cs) - 1; i >= 0; i-- {
		if cs[i].Role == role {
			return &cs[i]
		}
	}
	return nil
}

// Counting is the Confirm by role that counts on the current select of a
// projected record, or nil.
func (s *Decision) Counting(role string) *ConfirmRec {
	if s == nil || s.Dossier == nil {
		return nil
	}
	return s.counting(s.Current, role, s.Dossier.Items)
}

// Latest is the last Confirm by role on the current select, or nil.
func (s *Decision) Latest(role string) *ConfirmRec {
	if s == nil {
		return nil
	}
	return s.latest(s.Current, role)
}

// ChangedSince lists what the current Dossier of a projected record holds and
// c's does not, then the reverse.
func (s *Decision) ChangedSince(c *ConfirmRec) (added, removed []DossierItem) {
	return diffItems(c.Items, s.Dossier.Items)
}

// accepted maps each limit key an owner Confirm accepted on select id with
// the Dossier now to that owner's actor: a sufficient owner Confirm with this
// Dossier accepted it, and no later owner insufficient has this Dossier.
func (s *Decision) accepted(id string, now []DossierItem) map[string]string {
	out := map[string]string{}
	for _, c := range s.Confirms[id] {
		if c.Role != event.RoleOwner || !slices.Equal(c.Items, now) {
			continue
		}
		if c.Verdict == event.ConfirmInsufficient {
			clear(out)
		}
		for _, k := range c.Accepts {
			out[k] = c.Actor
		}
	}
	return out
}

// newCandidate reports whether a candidate now was not a candidate when
// select rec applied. A candidate that fell since does not count: the chosen
// one still stands.
func newCandidate(rec SelectRec, candidates []string) bool {
	return slices.ContainsFunc(candidates, func(c string) bool { return !slices.Contains(rec.Candidates, c) })
}

// reason is the first reason question id is not answered (spec §7.4), or ""
// when it is. live lists its live child Hypotheses, candidates its trade-off
// candidates, and stale says the question itself is stale. The Dossier is read
// only for a scope-change select, since this runs on every event.
func (d *derivation) reason(id string, live, candidates []string, stale bool) string {
	st := d.st
	s := st.Decisions[id]
	if len(s.Conflicts) > 0 {
		return ReasonConflicting
	}
	if s.Current == "" {
		return d.candidateReason(live)
	}
	rec := s.Selects[s.Current]
	switch {
	case d.hypStatus(rec.Hyp) != HypSupported:
		return ReasonSelectedNotSupport
	case d.staleOf(rec.Hyp) != nil || stale:
		return ReasonSelectedStale
	case st.awaitsSubquestions(rec.Hyp, func(c string) QuestionState { return d.dqState(c).State }):
		return ReasonAwaitingSubquestions
	case newCandidate(rec, candidates):
		return ReasonNewCandidate
	}
	// Only a scope change needs the owner's confirm: a trade-off ends with
	// the owner's choice, recorded as the select (ADR 0015).
	if rec.ScopeChange {
		if o := s.counting(rec.ID, event.RoleOwner, st.dossierItems(id, rec.Hyp)); o == nil || o.Verdict != event.ConfirmSufficient {
			return ReasonAwaitingOwner
		}
	}
	return ""
}

// candidateReason gives the open reason of a DQ with no current selection.
func (d *derivation) candidateReason(live []string) string {
	if len(live) == 0 {
		return ReasonNoCandidates
	}
	count := map[HypStatus]int{}
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

// decisions builds what Projection.Decisions holds: each question's stored
// record with its limits and the Dossier of its current select. Each is a
// copy, so a projection keeps its limits and Dossier as the log goes on.
func (st *State) decisions() map[string]*Decision {
	limits := st.limits()
	out := make(map[string]*Decision, len(st.Decisions))
	for q, s := range st.Decisions {
		c := *s
		c.Limits = limits[q]
		if s.Current != "" {
			c.Dossier = st.dossier(q, s.Selects[s.Current].Hyp)
		}
		out[q] = &c
	}
	return out
}

// readings derives every reading, in creation order. The owner checkpoint that
// closes a reading gives its first answer, and each later one that revises it
// gives the answer in force from then on. Readings of a retracted Result are
// left out.
func (st *State) readings(applied []event.Event) []Reading {
	answers := map[string][]Answer{}
	for _, ev := range applied {
		if ev.Type != event.TypeCheckpoint {
			continue
		}
		var c event.Checkpoint
		_ = event.DecodePayload(ev, &c)
		for _, k := range c.Closes {
			if len(answers[k]) == 0 {
				answers[k] = []Answer{{By: ev.ID, Text: c.Summary}}
			}
		}
		for _, k := range c.Revises {
			if len(answers[k]) > 0 {
				answers[k] = append(answers[k], Answer{By: ev.ID, Text: c.Summary})
			}
		}
	}
	aimed := map[string]bool{}
	for _, id := range st.Order {
		if n := st.Nodes[id]; n.Kind == event.KindExperiment {
			for _, k := range n.Aims {
				aimed[k] = true
			}
		}
	}
	pos := map[string]int{}
	for i, ev := range applied {
		pos[ev.ID] = i
	}
	var out []Reading
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] {
			continue
		}
		for i, item := range n.Readings {
			k := event.ReadingKey(id, i+1)
			l := Reading{Key: k, Result: id, Text: item.Text, Of: item.Of, Question: st.resultQuestion(id)}
			l.Reach = st.reach(id, l.Question)
			as := slices.Clone(answers[k])
			for j := range as {
				until := -1
				if j+1 < len(as) {
					until = pos[as[j+1].By]
				}
				as[j].Applied = st.applierBetween(k, pos, pos[as[j].By], until)
				// Only a first answer is applied without applies, so a tree
				// written before applies keeps its meaning; a revised answer
				// waits for a Result that names it.
				if j == 0 && as[j].Applied == "" && !aimed[k] {
					as[j].Applied = st.resultBetween(func(r string) bool { return st.resultQuestion(r) == l.Question }, pos, pos[as[j].By], until)
				}
			}
			if len(as) > 0 {
				last := as[len(as)-1]
				l.By, l.Answer, l.Answers, l.Applied = as[0].By, last.Text, as, last.Applied
			}
			out = append(out, l)
		}
	}
	return out
}

// applierBetween returns the first non-retracted Result whose applies names
// key k and whose position is after at and, when until >= 0, before until.
func (st *State) applierBetween(k string, pos map[string]int, at, until int) string {
	return st.resultBetween(func(r string) bool { return slices.Contains(st.Nodes[r].Applies, k) }, pos, at, until)
}

// resultBetween returns the first non-retracted Result for which keep holds
// and whose position is after at and, when until >= 0, before until, or "".
func (st *State) resultBetween(keep func(string) bool, pos map[string]int, at, until int) string {
	best, bestPos := "", -1
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] || !keep(id) {
			continue
		}
		p, ok := pos[id]
		if ok && p > at && (until < 0 || p < until) && (bestPos < 0 || p < bestPos) {
			best, bestPos = id, p
		}
	}
	return best
}

// limits derives each question's limits. A limit is closed when a
// non-retracted Result closes it, accepted when a counting owner confirm on
// the question's current select accepted it (Decision.accepted), and open
// otherwise. Limits of a retracted Result are left out.
func (st *State) limits() map[string][]Limit {
	closers := map[string][]string{}
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] {
			continue
		}
		for _, k := range n.Closes {
			closers[k] = append(closers[k], id)
		}
	}
	// An owner checkpoint accepts the limits its accepts names, any
	// checkpoint refers the limits its refers names to the owner, and an
	// owner checkpoint returns the referred limits its returns names to the
	// runner. The latest refer or return names whose turn an open limit is.
	ownerAccepted, referred, returned := map[string]string{}, map[string]string{}, map[string]string{}
	for _, ev := range st.Applied {
		if ev.Type != event.TypeCheckpoint {
			continue
		}
		var c event.Checkpoint
		_ = event.DecodePayload(ev, &c)
		for _, k := range c.Accepts {
			if _, seen := ownerAccepted[k]; !seen {
				ownerAccepted[k] = ev.Actor
			}
		}
		for _, k := range c.Refers {
			referred[k], returned[k] = ev.ID, ""
		}
		for _, k := range c.Returns {
			referred[k], returned[k] = "", ev.ID
		}
	}
	out := map[string][]Limit{}
	accepted := map[string]map[string]string{}
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] || len(n.Limits) == 0 {
			continue
		}
		q := st.resultQuestion(id)
		s := st.Decisions[q]
		if _, ok := accepted[q]; !ok && s.Current != "" {
			accepted[q] = s.accepted(s.Current, st.dossierItems(q, s.Selects[s.Current].Hyp))
		}
		for i, text := range n.Limits {
			l := Limit{Key: event.LimitKey(id, i+1), Result: id, Question: q, Text: text, Hyps: sortedKeys(n.Verdicts), State: LimitOpen, Referred: referred[event.LimitKey(id, i+1)], Returned: returned[event.LimitKey(id, i+1)]}
			if cs := closers[l.Key]; len(cs) > 0 {
				l.State, l.By, l.Closers = LimitClosed, cs[0], cs
			} else if actor, ok := accepted[q][l.Key]; ok {
				l.State, l.By = LimitAccepted, actor
			} else if actor, ok := ownerAccepted[l.Key]; ok {
				l.State, l.By = LimitAccepted, actor
			}
			if s.Current != "" {
				l.Gating = l.Gates(s.Selects[s.Current].Hyp)
			}
			out[q] = append(out[q], l)
		}
	}
	return out
}

// reach lists the Hypotheses under question q that an answer to a reading on
// Result r can move: a candidate r judged, which a changed verdict could
// drop, and a contradicted rival whose only contradicts is r's, which a
// changed verdict could revive. Only r's verdicts count, since a reading
// states how those verdicts read the owner's words.
func (st *State) reach(r, q string) []string {
	var out []string
	n := st.Nodes[r]
	for _, h := range n.VerdictOrder {
		hn, hs := st.Nodes[h], st.Hyps[h]
		if hn == nil || hs == nil || hn.Parent != q {
			continue
		}
		switch hs.Status {
		case HypSupported:
			out = append(out, h)
		case HypContradicted:
			if cs := st.Contradictions(h); len(cs) == 1 && cs[0] == r {
				out = append(out, h)
			}
		}
	}
	return out
}
