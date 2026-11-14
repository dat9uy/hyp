// Package render turns a replay.Projection into text. Renderers are pure
// functions of their arguments; they never read files.
package render

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// DefaultMaxLines is the default length budget of Status.
const DefaultMaxLines = 60

// line kinds, used to decide what truncation may drop.
const (
	kKeep   = iota // never dropped
	kRecent        // dropped first
	kExp           // then planned Experiments under the frontier
	kHyp           // then Hypotheses under the frontier
	kStale         // then STALE lines beyond the first five
	kOwed          // then OWED lines beyond the first five
)

type line struct {
	kind int
	text string
}

// Status renders the orientation screen for a fresh agent. maxLines <= 0
// means unlimited.
func Status(p replay.Projection, maxLines int) string { return StatusLive(p, maxLines, false) }

// StatusLive is Status; with live set, FRONTIER leaves out each contradicted
// Hypothesis that holds no marked Experiment, and counts them under their
// question. The owner reads it; a runner reads the full Status.
func StatusLive(p replay.Projection, maxLines int, live bool) string {
	return StatusOf(p, "", maxLines, live)
}

// StatusOf renders the status of Goal g, or of the store when g is "". A
// store with one Goal prints that Goal as it always did; a store with many
// prints a GOALS table and one AWAITING OWNER list, grouped by Goal, and
// hyp status G-… prints one Goal (ADR 0020).
func StatusOf(p replay.Projection, g string, maxLines int, live bool) string {
	if len(p.Goals) == 0 {
		return "GOAL none (hyp add goal --title T --why W)\n"
	}
	if g == "" {
		if len(p.Goals) > 1 {
			return join(truncate(goalsTable(p), maxLines))
		}
		g = p.Goals[0].GoalID
	}
	r := renderer{p: p}
	allOwed := p.Owed()
	out, awaitingOwner := r.goalBlock(g, allOwed, live)
	add := func(kind int, format string, a ...any) { out = append(out, line{kind, fmt.Sprintf(format, a...)}) }

	add(kKeep, "AWAITING OWNER")
	for _, s := range r.ownerLines(g, allOwed, awaitingOwner) {
		add(kKeep, "%s", s)
	}

	// OPEN LIMITS is a ledger, not open work: it lists every open limit of the
	// Goal, whatever the state of its question or whether FRONTIER prints it.
	var open []string
	for _, ev := range p.Applied {
		if ev.Type != event.TypeNodeAdd || p.Nodes[ev.ID].Kind != event.KindDesignQuestion || p.GoalOf(ev.ID) != g {
			continue
		}
		for _, l := range p.LimitsOf(ev.ID) {
			if l.Open() {
				open = append(open, fmt.Sprintf("  %s %s — %s", r.id(ev.ID), r.limitLabel(l.Result, l.Key), Cut(l.Text, 80)))
			}
		}
	}
	if len(open) > 0 {
		add(kKeep, "OPEN LIMITS")
		for _, s := range open {
			add(kKeep, "%s", s)
		}
	}

	add(kKeep, "RECENT")
	shown := 0
	for i := len(p.Applied) - 1; i >= 0 && shown < 10; i-- {
		ev := p.Applied[i]
		if ev.Type == event.TypeNodeAdd || ev.Type == event.TypeCheckpoint || !p.EventOf(ev, g) {
			continue
		}
		add(kRecent, "  %s", r.eventLine(ev, 80))
		shown++
	}

	if cp := p.LastCheckpoint[g]; cp != nil {
		var s event.Checkpoint
		_ = json.Unmarshal(cp.Payload, &s)
		add(kKeep, "CHECKPOINT %s %s: %s", Minute(cp.TS), cp.Actor, OneLine(s.Summary))
	} else {
		add(kKeep, "CHECKPOINT none")
	}
	if q := questionable(p, g); q > 0 {
		add(kKeep, "QUESTIONABLE %d results", q)
	}
	if n := uncontained(p, g); n > 0 {
		add(kKeep, "UNCONTAINED %d results", n)
	}
	if len(p.Rejections) > 0 {
		add(kKeep, "REJECTED %d events (run hyp check)", len(p.Rejections))
	}
	return join(truncate(out, maxLines))
}

// goalBlock renders Goal g's GOAL, LINES, FRONTIER, STALE and OWED, from the
// nodes whose parents lead to it, and returns with them the owner's choose
// lines its FRONTIER walk found.
func (r renderer) goalBlock(gid string, allOwed []replay.Owed, live bool) (out []line, awaitingOwner []string) {
	p := r.p
	add := func(kind int, format string, a ...any) { out = append(out, line{kind, fmt.Sprintf(format, a...)}) }
	gs := p.GoalStatusOf(gid)
	g := p.Nodes[gs.GoalID]
	c := gs.Counts
	add(kKeep, "GOAL %s %s — %s (answered %d, open %d, reopened %d, deferred %d, withdrawn %d, stale %d)",
		r.id(g.ID), Cut(g.Title, 60), gs.Status, c[string(replay.DQAnswered)], c[string(replay.DQOpen)],
		c[string(replay.DQReopened)], c[string(replay.DQDeferred)], c[string(replay.DQWithdrawn)], c[replay.StaleReasonStale])

	// LINES are the Goal's lines: whose turn each is, the Results that ran
	// its case, and a mark when no one closer ran the selected Hypotheses
	// together. The ledger owes a line on each question (OWED).
	if lines := p.LinesOf(g.ID); len(lines) > 0 {
		add(kKeep, "LINES")
		for _, l := range lines {
			add(kKeep, "  %s", Cut(r.lineLine(l), 160))
			if sel := p.NotTogether(l); len(sel) > 0 && !l.Accepted() {
				add(kKeep, "    (no one closer covers %s together)", strings.ReplaceAll(r.ids(sel), ", ", " and "))
			}
		}
	}

	owed := slices.DeleteFunc(slices.Clone(allOwed), func(o replay.Owed) bool { return p.GoalOf(o.Question) != g.ID })
	add(kKeep, "FRONTIER")
	// FRONTIER is the tree pruned to open work. The walk visits every question
	// once. A node prints only beneath its parent, and a sub-question prints only
	// from an open or reopened question and through a rival. An open question
	// that does not print is counted, so the Goal line's counts stay accounted for.
	unprinted := 0
	edges := p.LiveEdges()
	onFrontier := func(dq string) bool {
		s := p.Questions[dq].State
		return s.Open()
	}
	// blocker reports a sub-question that a printed supported Hypothesis still
	// owes and that would otherwise not print, so nothing would name the work.
	blocker := func(h, sub string) bool {
		return p.Hyps[h].Status == replay.HypSupported && p.Questions[sub].State == replay.DQDeferred
	}
	printQuestion := func(dq, indent string) {
		ns := p.Questions[dq]
		add(kKeep, "%s%s %s/%s %s", indent, r.id(dq), ns.State, ns.Reason, Cut(p.Nodes[dq].Title, 60))
		// The candidates line says whether one or several are left; with none,
		// the question's reason already says so.
		// Two or more candidates are a trade-off, a state named where it is
		// computed; the owner's turn to choose prints under AWAITING OWNER only
		// when the ledger owes it.
		if len(ns.Candidates) > 0 {
			state := ""
			if len(ns.Candidates) >= 2 {
				state = " (trade-off)"
			}
			add(kKeep, "%s  candidates (%d): %s%s", indent, len(ns.Candidates), r.named(ns.Candidates), state)
			// A premise a candidate rests on prints beside it, so the reader
			// sees which other answer the choice assumes.
			for _, pr := range premises(p, ns.Candidates) {
				by := ""
				switch {
				case len(pr.by) < len(ns.Candidates):
					by = " — by " + r.ids(pr.by)
				case len(ns.Candidates) >= 2:
					by = " — by every candidate"
				}
				where := ""
				if q := p.Nodes[pr.to].Parent; q != dq && p.Questions[q] != nil {
					where = fmt.Sprintf(", in %s %s", r.id(q), p.Questions[q].State)
				}
				add(kKeep, "%s  assumes %s %s%s%s", indent, r.withAlias(pr.to), p.Hyps[pr.to].Status, where, by)
			}
		}
		switch {
		case slices.ContainsFunc(owed, func(o replay.Owed) bool { return o.Kind == replay.OwedChoose && o.Key == dq }):
			mark := ""
			if !resultSeparates(p, ns.Candidates) {
				mark = " (no Result separates them)"
			}
			awaitingOwner = append(awaitingOwner, fmt.Sprintf("  %s choose: %s%s", r.id(dq), r.named(ns.Candidates), mark))
		}
		// When the question's turn waits on something other than runner work
		// (spec §8.1), a line names each thing, so the owner sees why the
		// runner owes nothing there. Runner work prints under OWED, and a
		// premise also has its assumes line above.
		if t := p.Turn(dq); t.Kind == replay.TurnWaits && t.Waits[0].Kind != replay.WaitRunner {
			for _, w := range t.Waits {
				switch w.Kind {
				case replay.WaitReading:
					if len(w.Reach) == 0 {
						add(kKeep, "%s  waits on open reading %s for the owner; it holds %s's decision", indent, r.limitLabel(event.KeyResult(w.Key), w.Key), r.id(dq))
					} else {
						add(kKeep, "%s  waits on open reading %s for the owner; its answer can move %s", indent, r.limitLabel(event.KeyResult(w.Key), w.Key), r.named(w.Reach))
					}
				case replay.WaitReferred:
					if len(ns.Candidates) > 0 {
						add(kKeep, "%s  waits on referred limit %s for the owner; an accept releases the select, a return sends its case to run", indent, r.limitLabel(event.KeyResult(w.Key), w.Key))
					}
				case replay.WaitSubquestion:
					add(kKeep, "%s  waits on a sub-question of %s not answered", indent, r.withAlias(w.Key))
				}
			}
		}
		if chosen := chosenCandidate(p, dq); chosen != "" {
			var rivals []string
			for _, h := range ns.Candidates {
				if h != chosen {
					rivals = append(rivals, h)
				}
			}
			add(kKeep, "%s  trade-off decided: %s over %s", indent, r.id(chosen), r.ids(rivals))
		}
	}
	printHypothesis := func(h, indent string, kind int) {
		hs := p.Hyps[h]
		stale := ""
		if hs.Stale != nil {
			stale = " STALE"
		}
		sup := ""
		for _, e := range edges {
			if e.From == h && e.Type == event.EdgeSupersedes {
				sup += " (supersedes " + r.id(e.To) + ")"
			}
		}
		add(kind, "%s  %s %s%s%s %s", indent, r.withAlias(h), hs.Status, stale, sup, Cut(p.Nodes[h].Title, 60))
		for _, e := range p.Nodes[h].Children {
			if p.Nodes[e].Kind == event.KindExperiment && p.Exps[e].Status == replay.ExpPlanned && len(r.marks(e)) == 0 {
				add(kExp, "%s    %s planned %s", indent, r.id(e), Cut(p.Nodes[e].Title, 60))
			}
		}
		// A marked Experiment is dropped with its Hypothesis, not before it.
		// A planned one keeps its title on the same line, and a closed one
		// has no line.
		for _, e := range p.Nodes[h].Children {
			if p.Nodes[e].Kind != event.KindExperiment {
				continue
			}
			if m := r.marks(e); len(m) > 0 && !p.Exps[e].Closed() {
				title := ""
				if p.Exps[e].Status == replay.ExpPlanned {
					title = " " + Cut(p.Nodes[e].Title, 60)
				}
				add(kHyp, "%s    %s %s [%s]%s", indent, r.id(e), p.Exps[e].Status, strings.Join(m, "; "), title)
			}
		}
	}
	// inPlay says the branch above dq lets it print.
	var walk func(dq string, depth int, inPlay bool)
	walk = func(dq string, depth int, inPlay bool) {
		indent := strings.Repeat("  ", depth+1)
		prints := inPlay && onFrontier(dq)
		if prints {
			printQuestion(dq, indent)
		} else if onFrontier(dq) {
			unprinted++
		}
		hidden := 0
		for _, h := range p.Nodes[dq].Children {
			if p.Nodes[h].Kind == event.KindExperiment {
				// A measuring Experiment under the question prints while
				// it is planned, as one under a Hypothesis does.
				if prints && p.Exps[h].Status == replay.ExpPlanned {
					add(kExp, "%s  %s planned (%s) %s", indent, r.id(h), replay.ShapeMeasure, Cut(p.Nodes[h].Title, 60))
				}
				continue
			}
			listed := prints && replay.LiveHypStatus(p.Hyps[h].Status)
			subInPlay := listed && replay.RivalStatus(p.Hyps[h].Status)
			var subs []string
			for _, c := range p.Nodes[h].Children {
				if p.Nodes[c].Kind == event.KindDesignQuestion {
					subs = append(subs, c)
				}
			}
			if listed {
				// A sub-question's line means nothing without the Hypothesis that
				// holds it, so truncation keeps that Hypothesis.
				kind := kHyp
				if slices.ContainsFunc(subs, func(sub string) bool { return subInPlay && onFrontier(sub) || blocker(h, sub) }) {
					kind = kKeep
				}
				if live && kind == kHyp && p.Hyps[h].Status == replay.HypContradicted && !r.marked(h) {
					hidden++
				} else {
					printHypothesis(h, indent, kind)
				}
			}
			for _, sub := range subs {
				if listed && blocker(h, sub) {
					add(kKeep, "%s    %s %s %s", indent, r.id(sub), p.Questions[sub].State, Cut(p.Nodes[sub].Title, 60))
				}
				walk(sub, depth+2, subInPlay)
			}
		}
		if hidden > 0 {
			add(kKeep, "%s  (%d contradicted hidden)", indent, hidden)
		}
	}
	for _, dq := range g.Children {
		walk(dq, 0, true)
	}
	if unprinted == 1 {
		add(kKeep, "  (1 open question out of play; hyp render --markdown lists every question)")
	} else if unprinted > 1 {
		add(kKeep, "  (%d open questions out of play; hyp render --markdown lists every question)", unprinted)
	}

	add(kKeep, "STALE")
	n := 0
	for _, ev := range p.Applied {
		if ev.Type != event.TypeNodeAdd || p.GoalOf(ev.ID) != g.ID {
			continue
		}
		shown, _ := p.Display(ev.ID)
		if s := shown.Stale; s != nil {
			kind := kKeep
			if n >= 5 {
				kind = kStale
			}
			add(kind, "  %s stale ← %s", r.id(ev.ID), r.chain(s))
			n++
		}
	}

	// OWED is the ledger: what the runner owes on each question. Its last
	// line is the stop clause a run reads.
	add(kKeep, "OWED")
	n = 0
	for _, o := range owed {
		if o.Who == replay.OwedOwner {
			continue
		}
		kind := kKeep
		if n >= 5 {
			kind = kOwed
		}
		add(kind, "  %s: %s", o.Who, r.owedLine(o))
		n++
	}
	if !replay.RunnerOwes(owed) {
		add(kKeep, "  runner owes nothing: stop and report to the owner")
	}

	return out, awaitingOwner
}

// truncate drops lines, in kind order, until out fits maxLines including the
// trailer line it then appends.
func truncate(out []line, maxLines int) []line {
	if maxLines <= 0 || len(out) <= maxLines {
		return out
	}
	excess := len(out) - (maxLines - 1)
	dropped := 0
	drop := make([]bool, len(out))
	for _, kind := range []int{kRecent, kExp, kHyp, kStale, kOwed} {
		for i := len(out) - 1; i >= 0 && dropped < excess; i-- {
			if out[i].kind == kind && !drop[i] {
				drop[i] = true
				dropped++
			}
		}
	}
	if dropped == 0 {
		// Only frontier lines are left; they are never dropped.
		return out
	}
	var kept []line
	for i, l := range out {
		if !drop[i] {
			kept = append(kept, l)
		}
	}
	return append(kept, line{kKeep, fmt.Sprintf("… %d lines truncated (hyp status --max-lines 0 for all)", dropped)})
}

func join(lines []line) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.text)
		b.WriteByte('\n')
	}
	return b.String()
}

// renderer holds the projection for the small formatting helpers.
type renderer struct{ p replay.Projection }

func (r renderer) id(id string) string {
	if n := r.p.Nodes[id]; n != nil {
		return event.DisplayID(n.Kind, id)
	}
	if id == "" {
		return "none"
	}
	return id
}

// closed reports whether an Experiment is superseded or withdrawn.
// ruledOut names each contradicted live Hypothesis under question dq with the
// Results that contradict it, so a new rival's author sees what it must
// survive: "h-module by R-5JYRA0, h-path by R-9SFQMB".
func (r renderer) ruledOut(dq string) string {
	var parts []string
	for _, h := range r.p.Nodes[dq].Children {
		if hs := r.p.Hyps[h]; hs == nil || hs.Status != replay.HypContradicted {
			continue
		}
		var by []string
		for _, res := range r.p.LiveResults() {
			if r.p.Nodes[res].Verdicts[h] == event.VerdictContradicts {
				by = append(by, r.id(res))
			}
		}
		name := r.p.Nodes[h].Alias
		if name == "" {
			name = r.id(h)
		}
		parts = append(parts, name+" by "+strings.Join(by, ", "))
	}
	return strings.Join(parts, "; ")
}

// reach says what an answer to reading l can touch on its Result's question:
// the question's candidates, and each Hypothesis the Result judged with its
// verdict and its status now.
func (r renderer) reach(l replay.Reading) string {
	cands := "no candidate"
	if c := r.p.Questions[l.Question].Candidates; len(c) > 0 {
		cands = "candidates " + r.named(c)
	}
	n := r.p.Nodes[l.Result]
	var judged []string
	for _, h := range n.VerdictOrder {
		judged = append(judged, fmt.Sprintf("%s %s, now %s", r.withAlias(h), n.Verdicts[h], r.p.Hyps[h].Status))
	}
	return fmt.Sprintf("%s %s; judged %s", r.id(l.Question), cands, strings.Join(judged, "; "))
}

// marked says Hypothesis h holds an open Experiment FRONTIER marks one-rival,
// so --live keeps it and that mark in view.
func (r renderer) marked(h string) bool {
	for _, e := range r.p.Nodes[h].Children {
		if r.p.Nodes[e].Kind == event.KindExperiment && len(r.marks(e)) > 0 && !r.p.Exps[e].Closed() {
			return true
		}
	}
	return false
}

// marks names what Experiment e's round left untested: "one-rival" when its
// parent had rivals at its add and it targeted none. A target left
// without a verdict is refused at record (ADR 0018), so no mark names it.
func (r renderer) marks(e string) []string {
	var m []string
	x := r.p.Exps[e]
	if x.OneRival {
		m = append(m, "one-rival")
	}
	return m
}

// withAlias shows a node as H-DFKM5P (h-path), or as its id when it has no alias.
func (r renderer) withAlias(id string) string {
	if n := r.p.Nodes[id]; n != nil && n.Alias != "" {
		return r.id(id) + " (" + n.Alias + ")"
	}
	return r.id(id)
}

// named is ids with each node's alias beside it.
func (r renderer) named(ids []string) string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = r.withAlias(id)
	}
	return strings.Join(out, ", ")
}

func (r renderer) ids(ids []string) string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = r.id(id)
	}
	return strings.Join(out, ", ")
}

// chain renders a stale cause as "H-2 stale ← H-3 contradicted ← R-9".
func (r renderer) chain(c *replay.StaleCause) string {
	var parts []string
	for ; c != nil; c = c.Next {
		if c.Reason == replay.StaleReasonContradicts {
			parts = append(parts, r.id(c.Node))
			continue
		}
		parts = append(parts, r.id(c.Node)+" "+c.Reason)
	}
	return strings.Join(parts, " ← ")
}

func (r renderer) eventLine(ev event.Event, width int) string {
	if ev.Type == event.TypeAnswer {
		return fmt.Sprintf("%s %s %s %s", Minute(ev.TS), ev.Actor, ev.Type, Cut(r.answerKeys(ev), width))
	}
	return fmt.Sprintf("%s %s %s %s — %s", Minute(ev.TS), ev.Actor, ev.Type, r.id(ev.Node), Cut(ev.Why, width))
}

// SelectedHyp returns the Hypothesis of the DQ's current selection ("" when
// none).
func SelectedHyp(p replay.Projection, dq string) string {
	rec, _ := p.CurrentSelect(dq)
	return rec.Hyp
}

// SelectKind names what a select needs the owner for: "trade-off" when its
// Hypothesis is one of two or more candidates, or "" for a plain select.
func SelectKind(p replay.Projection, dq string, rec replay.SelectRec) string {
	var kinds []string
	if cands := p.Questions[dq].Candidates; len(cands) >= 2 && slices.Contains(cands, rec.Hyp) {
		kinds = append(kinds, "trade-off")
	}
	return strings.Join(kinds, ", ")
}

// chosenCandidate returns the Hypothesis of the DQ's current select when that
// select records the owner's choice: its Hypothesis is one of two or more
// candidates ("" otherwise).
func chosenCandidate(p replay.Projection, dq string) string {
	cands := p.Questions[dq].Candidates
	if rec, ok := p.CurrentSelect(dq); ok && len(cands) >= 2 && slices.Contains(cands, rec.Hyp) {
		return rec.Hyp
	}
	return ""
}

// questionable counts Goal g's Results with questionable evidence.
func questionable(p replay.Projection, g string) int {
	n := 0
	for id, res := range p.Results {
		if res.Questionable && p.GoalOf(id) == g {
			n++
		}
	}
	return n
}

// uncontained counts Goal g's Results with uncontained evidence.
func uncontained(p replay.Projection, g string) int {
	n := 0
	for id, res := range p.Results {
		if len(res.Uncontained) > 0 && p.GoalOf(id) == g {
			n++
		}
	}
	return n
}

// ownerLines are the lines AWAITING OWNER prints for Goal g: the choose
// lines its FRONTIER walk found, each other item the ledger owes the owner
// on g's questions, and each open reading of g's Results.
func (r renderer) ownerLines(g string, allOwed []replay.Owed, awaiting []string) []string {
	p := r.p
	out := slices.Clone(awaiting)
	for _, o := range allOwed {
		// A choose item prints with its question, from the walk.
		if o.Who == replay.OwedOwner && o.Kind != replay.OwedChoose && p.GoalOf(o.Question) == g {
			out = append(out, "  "+r.owedLine(o))
		}
	}
	// Open readings that ask the same question print as one line, each key
	// with the question it holds, so the owner answers them once.
	printed := map[string]bool{}
	for _, l := range p.Readings {
		if !l.Open() || printed[l.Key] || p.GoalOf(l.Question) != g {
			continue
		}
		var open []string
		for _, k := range p.SameAsk(l.Key) {
			if m, _ := p.Reading(k); m.Open() && p.GoalOf(m.Question) == g {
				open = append(open, k)
			}
		}
		if len(open) == 1 {
			out = append(out, fmt.Sprintf("  %s%s reading: %s", r.limitLabel(l.Result, l.Key), r.of(l), l.Text))
			continue
		}
		labels := make([]string, len(open))
		for i, k := range open {
			printed[k] = true
			m, _ := p.Reading(k)
			labels[i] = fmt.Sprintf("%s (%s)", r.limitLabel(m.Result, k), r.id(m.Question))
		}
		out = append(out, fmt.Sprintf("  %s%s reading: %s", strings.Join(labels, ", "), r.of(l), l.Text))
		// Each copy holds its own Result's question: name what the answer
		// can touch there, from stored verdicts and derived candidates.
		for _, k := range open {
			m, _ := p.Reading(k)
			out = append(out, fmt.Sprintf("    %s: %s", r.limitLabel(m.Result, k), r.reach(m)))
		}
	}
	return out
}

// goalsTable is the status of a store with many Goals: one row per Goal, in
// creation order, with its state, what its runner owes and its last runner
// checkpoint, then one AWAITING OWNER list grouped by Goal. hyp status G-…
// prints one Goal.
func goalsTable(p replay.Projection) []line {
	r := renderer{p: p}
	var out []line
	add := func(format string, a ...any) { out = append(out, line{kKeep, fmt.Sprintf(format, a...)}) }
	allOwed := p.Owed()
	add("GOALS (hyp status G-… prints one Goal)")
	awaiting := map[string][]string{}
	for _, gs := range p.Goals {
		g := gs.GoalID
		c := gs.Counts
		runner := 0
		for _, o := range allOwed {
			if o.Who == replay.OwedRunner && p.GoalOf(o.Question) == g {
				runner++
			}
		}
		owes := "runner owes nothing"
		if runner > 0 {
			owes = fmt.Sprintf("runner owes %d", runner)
		}
		cp := "checkpoint none"
		if ev := p.LastCheckpoint[g]; ev != nil {
			cp = fmt.Sprintf("checkpoint %s %s", Minute(ev.TS), ev.Actor)
		}
		add("  %s %s — %s (answered %d, open %d, reopened %d, deferred %d, withdrawn %d, stale %d); %s; %s",
			r.id(g), Cut(p.Nodes[g].Title, 60), gs.Status, c[string(replay.DQAnswered)], c[string(replay.DQOpen)],
			c[string(replay.DQReopened)], c[string(replay.DQDeferred)], c[string(replay.DQWithdrawn)], c[replay.StaleReasonStale], owes, cp)
		_, choose := r.goalBlock(g, allOwed, false)
		awaiting[g] = r.ownerLines(g, allOwed, choose)
	}
	add("AWAITING OWNER")
	for _, gs := range p.Goals {
		if len(awaiting[gs.GoalID]) == 0 {
			continue
		}
		add("  %s %s", r.id(gs.GoalID), Cut(p.Nodes[gs.GoalID].Title, 60))
		for _, s := range awaiting[gs.GoalID] {
			add("  %s", s)
		}
	}
	if len(p.Rejections) > 0 {
		add("REJECTED %d events (run hyp check)", len(p.Rejections))
	}
	return out
}

// OneLine prints s whole on one line: each line break becomes a space, so a
// summary's own lines never read as a section or a field of the view. A line
// break is CR LF, or one of LF, CR, VT, FF, NEL, LS and PS.
func OneLine(s string) string {
	return lineBreaks.Replace(s)
}

var lineBreaks = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\v", " ", "\f", " ", "\u0085", " ", "\u2028", " ", "\u2029", " ")

// Cut shortens s to at most n runes, ending in … when it was cut.
func Cut(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// Minute formats an RFC3339 timestamp as "YYYY-MM-DD HH:MM" in UTC.
func Minute(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	return t.UTC().Format("2006-01-02 15:04")
}

// resultSeparates reports whether a non-retracted Result supports one candidate
// and gives another candidate a different verdict.
// owedLine is one ledger item as hyp status prints it.
func (r renderer) owedLine(o replay.Owed) string {
	q := r.id(o.Question)
	switch o.Kind {
	case replay.OwedRivals:
		if out := r.ruledOut(o.Question); out != "" {
			return q + " add rival Hypotheses; ruled out: " + out
		}
		return q + " add rival Hypotheses"
	case replay.OwedUntested:
		return q + " test " + r.withAlias(o.Key)
	case replay.OwedUndecided:
		return q + " decide " + r.withAlias(o.Key) + ", inconclusive with no planned Experiment"
	case replay.OwedPremise:
		t := r.p.Nodes[o.To].Parent
		return q + " premise " + r.id(o.Key) + " assumes " + r.id(o.To) + ", but " + r.id(t) + " selected " + r.id(SelectedHyp(r.p, t))
	case replay.OwedPlanned:
		return q + " run " + r.id(o.Key) + " " + Cut(r.p.Nodes[o.Key].Title, 60)
	case replay.OwedLimit, replay.OwedReading, replay.OwedReferred:
		res := event.KeyResult(o.Key)
		text := ""
		if l, ok := r.p.Limit(o.Key); ok {
			text = l.Text
		}
		for _, l := range r.p.Readings {
			if l.Key == o.Key {
				text = l.Answer
			}
		}
		what := map[replay.OwedKind]string{replay.OwedLimit: "limit", replay.OwedReading: "apply answer", replay.OwedReferred: "limit referred, accept or return"}[o.Kind]
		state := ""
		if o.Kind != replay.OwedReferred {
			state = " [" + o.State
			if len(o.AimedBy) > 0 {
				state += " by " + r.ids(o.AimedBy)
			}
			state += "]"
		}
		return fmt.Sprintf("%s %s %s%s — %s", q, what, r.limitLabel(res, o.Key), state, Cut(text, 80))
	case replay.OwedChoose:
		return q + " choose: " + r.named(r.p.Questions[o.Question].Candidates)
	case replay.OwedSelect:
		return q + " select " + r.withAlias(o.Key) + ", the only candidate"
	case replay.OwedStale:
		return q + " stale " + r.withAlias(o.Key) + ": reaffirm it, or re-run what its select rests on"
	}
	return q + " " + string(o.Kind)
}

func resultSeparates(p replay.Projection, candidates []string) bool {
	for _, res := range p.LiveResults() {
		verdicts := p.Nodes[res].Verdicts
		supports, other := false, false
		for _, h := range candidates {
			switch v, ok := verdicts[h]; {
			case ok && v == event.VerdictSupports:
				supports = true
			case ok:
				other = true
			}
		}
		if supports && other {
			return true
		}
	}
	return false
}

// premise is a Hypothesis that candidates rest on through depends_on.
type premise struct {
	to string
	by []string
}

// premises lists the live depends_on targets of the candidates, in the order
// the candidates and their edges give, each with the candidates that assume it.
func premises(p replay.Projection, candidates []string) []premise {
	var out []premise
	for _, h := range candidates {
		for _, e := range p.Edges {
			if !e.Live || e.Type != event.EdgeDependsOn || e.From != h {
				continue
			}
			i := slices.IndexFunc(out, func(pr premise) bool { return pr.to == e.To })
			if i < 0 {
				out = append(out, premise{to: e.To})
				i = len(out) - 1
			}
			if !slices.Contains(out[i].by, h) {
				out[i].by = append(out[i].by, h)
			}
		}
	}
	return out
}

// answerKeys lists the keys an answer settles, by field: "answers R-…#r1;
// accepts G-…#2".
func (r renderer) answerKeys(ev event.Event) string {
	var a event.Answer
	_ = event.DecodePayload(ev, &a)
	var parts []string
	for _, f := range []struct {
		items []event.KeyText
		verb  string
	}{{a.Answers, "answers"}, {a.Accepts, "accepts"}, {a.Returns, "returns"}} {
		var keys []string
		for _, it := range f.items {
			keys = append(keys, r.limitLabel(event.KeyResult(it.Key), it.Key))
		}
		if len(keys) > 0 {
			parts = append(parts, f.verb+" "+strings.Join(keys, ", "))
		}
	}
	return strings.Join(parts, "; ")
}
