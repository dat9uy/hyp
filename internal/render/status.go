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
	if p.Goal.GoalID == "" {
		return "GOAL none (hyp add goal --title T --why W)\n"
	}
	r := renderer{p: p}
	var out []line
	add := func(kind int, format string, a ...any) { out = append(out, line{kind, fmt.Sprintf(format, a...)}) }

	g := p.Nodes[p.Goal.GoalID]
	c := p.Goal.Counts
	add(kKeep, "GOAL %s %s — %s (answered %d, open %d, reopened %d, deferred %d, withdrawn %d, stale %d)",
		r.id(g.ID), Cut(g.Title, 60), p.Goal.Status, c[string(replay.DQAnswered)], c[string(replay.DQOpen)],
		c[string(replay.DQReopened)], c[string(replay.DQDeferred)], c[string(replay.DQWithdrawn)], c[replay.StaleReasonStale])

	// LINES are the Goal's lines: whose turn each is, the Results that ran
	// its case, and a mark when no one closer ran the selected Hypotheses
	// together. The ledger owes a line on each question (OWED).
	if len(p.Lines) > 0 {
		add(kKeep, "LINES")
		for _, l := range p.Lines {
			add(kKeep, "  %s", Cut(r.lineLine(l), 160))
			if sel := p.NotTogether(l); len(sel) > 0 && !l.Accepted() {
				add(kKeep, "    (no one closer covers %s together)", strings.ReplaceAll(r.ids(sel), ", ", " and "))
			}
		}
	}

	owed := p.Owed()
	add(kKeep, "FRONTIER")
	// FRONTIER is the tree pruned to open work. The walk visits every question
	// once. A node prints only beneath its parent, and a sub-question prints only
	// from an open or reopened question and through a live rival. An open question
	// that does not print is counted, so the Goal line's counts stay accounted for.
	unprinted := 0
	edges := p.LiveEdges()
	onFrontier := func(dq string) bool {
		s := p.Questions[dq].State
		return s == replay.DQOpen || s == replay.DQReopened
	}
	// blocker reports a sub-question that a printed supported Hypothesis still
	// owes and that would otherwise not print, so nothing would name the work.
	blocker := func(h, sub string) bool {
		return p.Hyps[h].Status == replay.HypSupported && p.Questions[sub].State == replay.DQDeferred
	}
	// awaitingOwner collects, for each question FRONTIER prints, the one step
	// due from the owner: choose between the candidates, or confirm a
	// scope-change select.
	var awaitingOwner []string
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
		case ns.Reason == replay.ReasonAwaitingOwner:
			awaitingOwner = append(awaitingOwner, fmt.Sprintf("  %s awaiting_owner: %s (scope change)", r.id(dq), r.id(SelectedHyp(p, dq))))
		case slices.ContainsFunc(owed, func(o replay.Owed) bool { return o.Kind == replay.OwedChoose && o.Key == dq }):
			mark := ""
			if !resultSeparates(p, ns.Candidates) {
				mark = " (no Result separates them)"
			}
			awaitingOwner = append(awaitingOwner, fmt.Sprintf("  %s choose: %s%s", r.id(dq), r.named(ns.Candidates), mark))
		}
		// An open reading whose answer can move a Hypothesis on the question
		// holds its rivals, select and choice (spec §8.1). When the question
		// owes no runner item and no choice, the line names each such reading
		// and what its answer can move, so the owner sees why the runner owes
		// nothing there. A premise hold has its assumes line above.
		if !slices.ContainsFunc(owed, func(o replay.Owed) bool {
			return o.Question == dq && (o.Who == replay.OwedRunner || o.Kind == replay.OwedChoose)
		}) {
			for _, l := range p.Readings {
				if l.Question == dq && l.By == "" && len(l.Reach) > 0 {
					add(kKeep, "%s  held: open reading %s waits for the owner; its answer can move %s", indent, r.limitLabel(l.Result, l.Key), r.named(l.Reach))
				}
			}
			if k := p.ReferredLimit(dq); k != "" && len(ns.Candidates) > 0 {
				add(kKeep, "%s  held: referred limit %s waits for the owner; an accept releases the select, a return sends its case to run", indent, r.limitLabel(event.KeyResult(k), k))
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
			if m := r.marks(e); len(m) > 0 && !closed(p.Exps[e]) {
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
		if ev.Type != event.TypeNodeAdd {
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

	add(kKeep, "AWAITING OWNER")
	for _, s := range awaitingOwner {
		add(kKeep, "%s", s)
	}
	for _, o := range owed {
		// A choose item prints with its question, from the walk above.
		if o.Who == replay.OwedOwner && o.Kind != replay.OwedChoose {
			add(kKeep, "  %s", r.owedLine(o))
		}
	}
	// Open readings that ask the same question print as one line, each key
	// with the question it holds, so the owner answers them once.
	printed := map[string]bool{}
	for _, l := range p.Readings {
		if l.By != "" || printed[l.Key] {
			continue
		}
		var open []string
		for _, k := range p.SameAsk(l.Key) {
			if m, _ := p.Reading(k); m.By == "" {
				open = append(open, k)
			}
		}
		if len(open) == 1 {
			add(kKeep, "  %s%s reading: %s", r.limitLabel(l.Result, l.Key), r.of(l), l.Text)
			continue
		}
		labels := make([]string, len(open))
		for i, k := range open {
			printed[k] = true
			m, _ := p.Reading(k)
			labels[i] = fmt.Sprintf("%s (%s)", r.limitLabel(m.Result, k), r.id(m.Question))
		}
		add(kKeep, "  %s%s reading: %s", strings.Join(labels, ", "), r.of(l), l.Text)
		// Each copy holds its own Result's question: name what the answer
		// can touch there, from stored verdicts and derived candidates.
		for _, k := range open {
			m, _ := p.Reading(k)
			add(kKeep, "    %s: %s", r.limitLabel(m.Result, k), r.reach(m))
		}
	}

	// OPEN LIMITS is a ledger, not open work: it lists every open limit, whatever
	// the state of its question or whether FRONTIER prints it.
	var open []string
	for _, ev := range p.Applied {
		if ev.Type != event.TypeNodeAdd || p.Nodes[ev.ID].Kind != event.KindDesignQuestion {
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
		if ev.Type == event.TypeNodeAdd || ev.Type == event.TypeCheckpoint {
			continue
		}
		add(kRecent, "  %s", r.eventLine(ev, 80))
		shown++
	}

	if cp := p.LastCheckpoint; cp != nil {
		var s event.Checkpoint
		_ = json.Unmarshal(cp.Payload, &s)
		sum := Cut(s.Summary, 200)
		if sum != Cut(s.Summary, event.MaxText) {
			sum += " (full: hyp show " + r.id(cp.Node) + ")"
		}
		add(kKeep, "CHECKPOINT %s %s: %s", Minute(cp.TS), cp.Actor, sum)
	} else {
		add(kKeep, "CHECKPOINT none")
	}
	if q := questionable(p); q > 0 {
		add(kKeep, "QUESTIONABLE %d results", q)
	}
	if n := uncontained(p); n > 0 {
		add(kKeep, "UNCONTAINED %d results", n)
	}
	if len(p.Rejections) > 0 {
		add(kKeep, "REJECTED %d events (run hyp check)", len(p.Rejections))
	}
	return join(truncate(out, maxLines))
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
		for _, ev := range r.p.Applied {
			if ev.Type == event.TypeResultRecord && r.p.Results[ev.ID].Status != replay.ResRetracted && r.p.Nodes[ev.ID].Verdicts[h] == event.VerdictContradicts {
				by = append(by, r.id(ev.ID))
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

// marked says Hypothesis h holds an open Experiment FRONTIER marks one-rival
// or unjudged, so --live keeps it and that mark in view.
func (r renderer) marked(h string) bool {
	for _, e := range r.p.Nodes[h].Children {
		if r.p.Nodes[e].Kind == event.KindExperiment && len(r.marks(e)) > 0 && !closed(r.p.Exps[e]) {
			return true
		}
	}
	return false
}

func closed(x *replay.Exp) bool {
	return x.Status == replay.ExpSuperseded || x.Status == replay.ExpWithdrawn
}

// marks names what Experiment e's round left untested: "one-rival" when its
// parent had live rivals at its add and it targeted none, and "unjudged H"
// for each target live at its add that no Result judged.
func (r renderer) marks(e string) []string {
	var m []string
	x := r.p.Exps[e]
	if x.OneRival {
		m = append(m, "one-rival")
	}
	for _, h := range x.Unjudged {
		m = append(m, "unjudged "+r.id(h))
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
	return fmt.Sprintf("%s %s %s %s — %s", Minute(ev.TS), ev.Actor, ev.Type, r.id(ev.Node), Cut(ev.Why, width))
}

// SelectedHyp returns the Hypothesis of the DQ's current selection ("" when
// none).
func SelectedHyp(p replay.Projection, dq string) string {
	rec, _ := p.CurrentSelect(dq)
	return rec.Hyp
}

// SelectKind names what a select needs the owner for: "trade-off" when its
// Hypothesis is one of two or more candidates, "scope-change", both joined by
// ", ", or "" for a plain select.
func SelectKind(p replay.Projection, dq string, rec replay.SelectRec) string {
	var kinds []string
	if cands := p.Questions[dq].Candidates; len(cands) >= 2 && slices.Contains(cands, rec.Hyp) {
		kinds = append(kinds, "trade-off")
	}
	if rec.ScopeChange {
		kinds = append(kinds, "scope-change")
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

func questionable(p replay.Projection) int {
	n := 0
	for _, res := range p.Results {
		if res.Questionable {
			n++
		}
	}
	return n
}

func uncontained(p replay.Projection) int {
	n := 0
	for _, res := range p.Results {
		if len(res.Uncontained) > 0 {
			n++
		}
	}
	return n
}

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
		what := map[string]string{replay.OwedLimit: "limit", replay.OwedReading: "apply answer", replay.OwedReferred: "limit referred, accept or return"}[o.Kind]
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
	}
	return q + " " + o.Kind
}

func resultSeparates(p replay.Projection, candidates []string) bool {
	for _, ev := range p.Applied {
		if ev.Type != event.TypeResultRecord || p.Results[ev.ID].Status == replay.ResRetracted {
			continue
		}
		verdicts := p.Nodes[ev.ID].Verdicts
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
