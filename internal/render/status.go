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
)

type line struct {
	kind int
	text string
}

// Status renders the orientation screen for a fresh agent. maxLines <= 0
// means unlimited.
func Status(p replay.Projection, maxLines int) string {
	if p.Goal.GoalID == "" {
		return "GOAL none (hyp add goal --title T --why W)\n"
	}
	r := renderer{p: p}
	var out []line
	add := func(kind int, format string, a ...any) { out = append(out, line{kind, fmt.Sprintf(format, a...)}) }

	g := p.Nodes[p.Goal.GoalID]
	c := p.Goal.Counts
	add(kKeep, "GOAL %s %s — %s (answered %d, open %d, reopened %d, deferred %d, withdrawn %d, stale %d)",
		r.id(g.ID), Cut(g.Title, 60), p.Goal.Status, c[replay.DQAnswered], c[replay.DQOpen],
		c[replay.DQReopened], c[replay.DQDeferred], c[replay.DQWithdrawn], c[replay.StaleReasonStale])

	add(kKeep, "FRONTIER")
	// FRONTIER is the tree pruned to open work. The walk visits every question
	// once. A node prints only beneath its parent, and a sub-question prints only
	// from an open or reopened question and through a live rival. An open question
	// that does not print is counted, so the Goal line's counts stay accounted for.
	unprinted := 0
	edges := p.LiveEdges()
	onFrontier := func(dq string) bool {
		s := p.Status[dq].Status
		return s == replay.DQOpen || s == replay.DQReopened
	}
	// blocker reports a question that keeps its printed Hypothesis waiting and
	// that would otherwise not print, so nothing would name the cause.
	blocker := func(h, sub string) bool {
		return p.Status[h].Status == replay.HypAwaitingSubquestions && p.Status[sub].Status == replay.DQDeferred
	}
	printQuestion := func(dq, indent string) {
		ns := p.Status[dq]
		add(kKeep, "%s%s %s/%s %s", indent, r.id(dq), ns.Status, ns.Reason, Cut(p.Nodes[dq].Title, 60))
		if ns.Reason == replay.ReasonVerifierInsufficient {
			add(kKeep, "%s  verifier: %s", indent, Cut(p.Decisions[dq].Counting(event.RoleVerifier).Why, event.MaxWhy))
		}
		if pick := tradeoffPick(p, dq); pick != "" {
			var rivals []string
			for _, h := range ns.Suggested {
				if h != pick {
					rivals = append(rivals, h)
				}
			}
			over := ""
			if len(rivals) > 0 {
				over = " over " + r.ids(rivals)
			}
			add(kKeep, "%s  trade-off decided: %s%s", indent, r.id(pick), over)
		}
	}
	printHypothesis := func(h, indent string, kind int) {
		hs := p.Status[h]
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
		add(kind, "%s  %s %s%s%s %s", indent, r.id(h), hs.Status, stale, sup, Cut(p.Nodes[h].Title, 60))
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
		for _, h := range p.Nodes[dq].Children {
			listed := prints && replay.LiveHypStatus(p.Status[h].Status)
			subInPlay := listed && replay.RivalStatus(p.Status[h].Status)
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
				printHypothesis(h, indent, kind)
			}
			for _, sub := range subs {
				if listed && blocker(h, sub) {
					add(kKeep, "%s    %s %s %s", indent, r.id(sub), p.Status[sub].Status, Cut(p.Nodes[sub].Title, 60))
				}
				walk(sub, depth+2, subInPlay)
			}
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
		if s := p.Status[ev.ID].Stale; s != nil {
			kind := kKeep
			if n >= 5 {
				kind = kStale
			}
			add(kind, "  %s stale ← %s", r.id(ev.ID), r.chain(s))
			n++
		}
	}

	add(kKeep, "AWAITING OWNER")
	for _, ev := range p.Applied {
		if ev.Type != event.TypeNodeAdd || p.Nodes[ev.ID].Kind != event.KindDesignQuestion {
			continue
		}
		ns := p.Status[ev.ID]
		if ns.Reason == replay.ReasonAwaitingOwner {
			add(kKeep, "  %s awaiting_owner: %s", r.id(ev.ID), r.id(SelectedHyp(p, ev.ID)))
		}
		if len(ns.Suggested) >= 2 && ns.Status != replay.DQAnswered && ns.Status != replay.DQWithdrawn &&
			tradeoffPick(p, ev.ID) == "" {
			add(kKeep, "  %s trade-off: %s", r.id(ev.ID), r.ids(ns.Suggested))
		}
	}

	var open []string
	for _, ev := range p.Applied {
		if ev.Type != event.TypeNodeAdd || p.Nodes[ev.ID].Kind != event.KindDesignQuestion {
			continue
		}
		for _, l := range p.LimitsOf(ev.ID) {
			if l.State == replay.LimitOpen {
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
	for _, kind := range []int{kRecent, kExp, kHyp, kStale} {
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

// SelectKind names what a select needs the owner for: "trade-off",
// "scope-change", both joined by ", ", or "" for a plain select.
func SelectKind(rec replay.SelectRec) string {
	var kinds []string
	if rec.Tradeoff {
		kinds = append(kinds, "trade-off")
	}
	if rec.ScopeChange {
		kinds = append(kinds, "scope-change")
	}
	return strings.Join(kinds, ", ")
}

// tradeoffPick returns the Hypothesis of the DQ's current select when that
// select is a trade-off and its pick is still suggested ("" otherwise).
func tradeoffPick(p replay.Projection, dq string) string {
	if rec, ok := p.CurrentSelect(dq); ok && rec.Tradeoff && slices.Contains(p.Status[dq].Suggested, rec.Hyp) {
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
