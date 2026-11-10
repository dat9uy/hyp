package render

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// Show renders one node in full. ancestors lists the node's ancestors from
// its parent up to the Goal (index.Ancestors).
func Show(p replay.Projection, id string, ancestors []string) string {
	r := renderer{p: p}
	n := p.Nodes[id]
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }

	w("%s %s", n.Kind, r.id(id))
	if n.Alias != "" {
		w("alias: %s", n.Alias)
	}
	if n.Title != "" {
		w("title: %s", n.Title)
	}
	if n.Body != "" {
		w("body: %s", n.Body)
	}
	if n.Kind == event.KindGoal {
		for _, l := range p.Lines {
			w("line %s", r.lineLine(l))
		}
		for _, l := range p.Readings {
			for _, line := range readingLines(r, l) {
				w("%s", line)
			}
		}
	}
	crumbs := make([]string, 0, len(ancestors)+1)
	for i := len(ancestors) - 1; i >= 0; i-- {
		crumbs = append(crumbs, r.id(ancestors[i]))
	}
	w("path: %s", strings.Join(append(crumbs, r.id(id)), " > "))

	ns, _ := p.Display(id)
	status := ns.Word
	if ns.Reason != "" {
		status += "/" + ns.Reason
	}
	if x := p.Exps[id]; x != nil && x.Status == replay.ExpSuperseded {
		status += " (replaced by " + r.ids(x.ReplacedBy) + ")"
	}
	w("status: %s", status)
	if ns.Stale != nil {
		w("stale: %s", r.chain(ns.Stale))
	}
	if q := p.Questions[id]; q != nil && len(q.Candidates) > 0 {
		w("candidates: %s", r.named(q.Candidates))
	}
	if sub := p.Subjects[id]; len(sub) > 0 {
		var files []string
		for _, f := range sub {
			files = append(files, f.Path+" "+f.SHA256[:12])
		}
		w("subject: %s", strings.Join(files, ", "))
	}
	if n.Kind == event.KindDesignQuestion {
		rec, ok := p.CurrentSelect(id)
		if ok {
			kind := ""
			if k := SelectKind(p, id, rec); k != "" {
				kind = " (" + k + ")"
			}
			w("selected: %s%s — %s", r.id(rec.Hyp), kind, rec.Why)
			if body := p.Nodes[rec.Hyp].Body; body != "" {
				w("design: %s", body)
			}
		}
		if ls := p.LimitsOf(id); len(ls) > 0 {
			w("limits:")
			for _, l := range ls {
				w("  %s", r.limitLine(l))
			}
		}
		// Live rivals no Result tells apart: each pair may be one claim, or
		// a case that would separate them was never run on both.
		if pairs := p.Unseparated(id); len(pairs) > 0 {
			w("not separated:")
			for _, pr := range pairs {
				w("  %s and %s: no Result gives them different verdicts", r.withAlias(pr[0]), r.withAlias(pr[1]))
			}
		}
		r.experiments(w, id)
	}

	for _, e := range p.LiveEdges() {
		switch {
		case e.From == id:
			w("edge out: %s %s", e.Type, r.id(e.To))
		case e.To == id:
			w("edge in: %s %s", r.id(e.From), e.Type)
		}
	}

	if n.Kind == event.KindExperiment {
		w("recipe: %s", n.Recipe)
		if n.Rubric != "" {
			w("rubric: %s", n.Rubric)
		}
		w("shape: %s", p.Exps[id].Shape)
		if len(n.Targets) > 0 {
			w("targets: %s", r.ids(n.Targets))
		}
		for _, f := range n.Records {
			w("records: %s %s", f.Path, f.SHA256[:12])
		}
		for _, k := range n.Aims {
			w("aims: %s", r.limitLabel(event.KeyResult(k), k))
		}
	}
	if n.Kind == event.KindResult {
		w("observation: %s", n.Observation)
		w("verdicts: %s", r.verdicts(n.Verdicts))
		for _, e := range n.Evidence {
			w("evidence: %s", e)
		}
		for i, l := range n.Limits {
			w("limit #%d: %s", i+1, l)
		}
		for _, k := range n.Closes {
			w("closes: %s", r.limitLabel(event.KeyResult(k), k))
		}
		for _, k := range n.Applies {
			w("applies: %s", r.limitLabel(event.KeyResult(k), k))
		}
		for _, l := range p.Readings {
			if l.Result == id {
				for _, line := range readingLines(r, l) {
					w("%s", line)
				}
			}
		}
		if resOf(p, id).Questionable {
			w("questionable: recipe changed since this Result")
		}
		for _, why := range resOf(p, id).Uncontained {
			w("uncontained: %s", why)
		}
	}
	if len(n.Children) > 0 {
		w("children:")
		for _, c := range n.Children {
			cs, _ := p.Display(c)
			line := fmt.Sprintf("  %s %s", r.withAlias(c), cs.Word)
			if cs.Reason != "" {
				line += "/" + cs.Reason
			}
			if cs.Stale != nil {
				line += " STALE"
			}
			if cn := p.Nodes[c]; cn.Kind == event.KindResult {
				line += " " + r.verdicts(cn.Verdicts)
				if resOf(p, c).Questionable {
					line += " QUESTIONABLE"
				}
				if len(resOf(p, c).Uncontained) > 0 {
					line += " UNCONTAINED"
				}
			} else {
				line += " " + Cut(cn.Title, 60)
			}
			w("%s", line)
		}
	}
	if n.Kind == event.KindHypothesis {
		// A Hypothesis may be judged as another Experiment's target, so its
		// children do not list every Result that judged it.
		judged := false
		for _, res := range p.LiveResults() {
			rn := p.Nodes[res]
			verdict, ok := rn.Verdicts[id]
			if !ok {
				continue
			}
			if !judged {
				w("judged by:")
				judged = true
			}
			w("  %s (%s) %s — %s", r.id(res), r.id(rn.Parent), verdict, rn.Observation)
			for _, l := range p.Readings {
				if l.Result == res {
					for _, line := range readingLines(r, l) {
						w("    %s", line)
					}
				}
			}
		}
		var preds []string
		seen := map[string]bool{id: true}
		edges := p.LiveEdges()
		for queue := []string{id}; len(queue) > 0; queue = queue[1:] {
			for _, e := range edges {
				if e.From == queue[0] && e.Type == event.EdgeSupersedes && !seen[e.To] {
					seen[e.To] = true
					preds = append(preds, e.To)
					queue = append(queue, e.To)
				}
			}
		}
		var lines []string
		for _, res := range p.LiveResults() {
			rn := p.Nodes[res]
			for _, pr := range preds {
				if _, ok := rn.Verdicts[pr]; ok {
					line := fmt.Sprintf("  %s (%s) %s", r.id(res), r.id(rn.Parent), r.verdicts(rn.Verdicts))
					if _, own := rn.Verdicts[id]; own {
						line += " (also counted for this hypothesis)"
					}
					lines = append(lines, line)
					break
				}
			}
		}
		if len(lines) > 0 {
			w("predecessor evidence (not counted):")
			for _, l := range lines {
				w("%s", l)
			}
		}
	}

	amended := false
	for _, ev := range p.Applied {
		if ev.Type == event.TypeHypothesisAmend && ev.Node == id {
			amended = true
			break
		}
	}
	w("history:")
	for _, ev := range p.Applied {
		if touches(ev, id) {
			w("  %s", r.eventLine(ev, event.MaxWhy))
			if ev.Type == event.TypeCheckpoint {
				var s event.Checkpoint
				_ = json.Unmarshal(ev.Payload, &s)
				w("    summary: %s", Cut(s.Summary, event.MaxText))
				for _, k := range s.Closes {
					w("    closes: %s", r.limitLabel(event.KeyResult(k), k))
				}
				for _, k := range s.Revises {
					w("    revises: %s", r.limitLabel(event.KeyResult(k), k))
				}
				for _, k := range s.Accepts {
					w("    accepts: %s", r.limitLabel(event.KeyResult(k), k))
				}
				for _, k := range s.Refers {
					w("    refers: %s", r.limitLabel(event.KeyResult(k), k))
				}
				for _, k := range s.Returns {
					w("    returns: %s", r.limitLabel(event.KeyResult(k), k))
				}
			}
			if amended && (ev.Type == event.TypeNodeAdd || ev.Type == event.TypeHypothesisAmend) {
				var rev struct {
					Body string `json:"body"`
				}
				_ = json.Unmarshal(ev.Payload, &rev)
				w("    body: %s", rev.Body)
			}
		}
	}
	return b.String()
}

// readingLines is one reading as hyp show prints it: the reading with the
// answer in force, then one indented line per earlier answer it replaced.
func readingLines(r renderer, l replay.Reading) []string {
	state := "open"
	if !l.Open() {
		revised := ""
		if n := len(l.Answers); n > 1 {
			revised = ", revised by owner checkpoint " + l.Answers[n-1].By
		}
		state = fmt.Sprintf("closed by owner checkpoint %s%s, answer: %q", l.By, revised, l.Answer)
	}
	out := []string{fmt.Sprintf("reading %s%s %s — %s", r.limitLabel(l.Result, l.Key), r.of(l), state, l.Text)}
	for _, a := range l.Answers[:max(len(l.Answers)-1, 0)] {
		applied := ", not applied"
		if a.Applied != "" {
			applied = ", applied by " + r.id(a.Applied)
		}
		out = append(out, fmt.Sprintf("  earlier answer: %q (owner checkpoint %s%s)", a.Text, a.By, applied))
	}
	return out
}

// of names the answer a reading reads, as " of R-W6V06D#r1", or "".
func (r renderer) of(l replay.Reading) string {
	if l.Of == "" {
		return ""
	}
	return " of " + r.limitLabel(event.KeyResult(l.Of), l.Of)
}

// limitLabel shows limit key k of Result res as R-W6V06D#1.
func (r renderer) limitLabel(res, k string) string {
	return r.id(res) + strings.TrimPrefix(k, res)
}

func (r renderer) verdicts(v map[string]string) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = r.id(k) + "=" + v[k]
	}
	return strings.Join(parts, " ")
}

// touches reports whether ev acts on node id or points an edge at it.
func touches(ev event.Event, id string) bool {
	if ev.Node == id {
		return true
	}
	if ev.Type == event.TypeEdgeAdd || ev.Type == event.TypeEdgeRemove {
		var e event.Edge
		_ = json.Unmarshal(ev.Payload, &e)
		return e.To == id
	}
	return false
}

// Diff lists every node whose derived status differs between before and
// after, then the events applied after the before point.
func Diff(before, after replay.Projection) string {
	r := renderer{p: after}
	var b strings.Builder
	for _, ev := range after.Applied {
		if ev.Type != event.TypeNodeAdd && ev.Type != event.TypeResultRecord {
			continue
		}
		_, had := before.Display(ev.ID)
		oldText := "absent"
		if had {
			oldText = statusText(r, before, ev.ID, true)
		}
		curText := statusText(r, after, ev.ID, true)
		if !had || oldText != curText {
			fmt.Fprintf(&b, "%s %s → %s\n", r.id(ev.ID), oldText, curText)
		}
	}
	b.WriteString("EVENTS\n")
	for _, ev := range after.Applied[min(len(before.Applied), len(after.Applied)):] {
		fmt.Fprintf(&b, "  %s\n", r.eventLine(ev, 80))
	}
	return b.String()
}

// statusText is node id's status in p as one phrase, naming a questionable
// Result when flagged.
func statusText(r renderer, p replay.Projection, id string, flagged bool) string {
	s, _ := p.Display(id)
	t := s.Word
	if s.Reason != "" {
		t += "/" + s.Reason
	}
	if s.Stale != nil {
		t += " stale ← " + r.chain(s.Stale)
	}
	if flagged && resOf(p, id).Questionable {
		t += " questionable"
	}
	return t
}

// resOf returns Result id's record, or the zero record when p has none.
func resOf(p replay.Projection, id string) replay.Res {
	if res := p.Results[id]; res != nil {
		return *res
	}
	return replay.Res{}
}

// closers names each Result that closes a limit with its verdicts, as
// "R-5JYRA0 (H-G1Z4MR (h-module) contradicts)". A close records that the
// case ran, whatever the verdicts, so the verdicts say what the case did to
// each Hypothesis it judged.
func (r renderer) closers(ids []string) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		n := r.p.Nodes[id]
		var vs []string
		for _, h := range n.VerdictOrder {
			vs = append(vs, r.withAlias(h)+" "+n.Verdicts[h])
		}
		parts[i] = r.id(id) + " (" + strings.Join(vs, "; ") + ")"
	}
	return strings.Join(parts, ", ")
}

// lineLine is one line of the Goal as hyp show and hyp status print it:
// G-…#n, its facts, and its text.
func (r renderer) lineLine(l replay.Limit) string {
	return fmt.Sprintf("%s %s — %s", r.limitLabel(l.Result, l.Key), r.lineState(l), l.Text)
}

// lineState prints a line's facts, each only when it has one, joined by
// "; ": its hand, its coverage on each live top-level question with a
// candidate, and the Results that ran its case. A line is never open or
// closed, so neither word prints.
func (r renderer) lineState(l replay.Limit) string {
	var parts []string
	if h := r.handWord(l); h != "" {
		parts = append(parts, h)
	}
	var covered []string
	var uncovered []string
	for _, q := range r.p.TopQuestions() {
		if len(r.p.Questions[q].Candidates) == 0 {
			continue
		}
		if un := r.p.Uncovered(l, q); len(un) > 0 {
			uncovered = append(uncovered, "uncovered on "+r.id(q)+": "+r.named(un))
		} else {
			covered = append(covered, r.id(q))
		}
	}
	if len(covered) > 0 {
		parts = append(parts, "covered on "+strings.Join(covered, ", "))
	}
	parts = append(parts, uncovered...)
	if len(l.Closers) > 0 {
		parts = append(parts, "ran in "+r.ids(l.Closers))
	}
	if len(parts) == 0 {
		return "no candidate"
	}
	return strings.Join(parts, "; ")
}

// handWord is a line's latest hand in words, or "" when none.
func (r renderer) handWord(l replay.Limit) string {
	switch {
	case l.Accepted():
		return "accepted by " + l.Hand.Actor
	case l.Referred():
		return "referred: the owner accepts or returns it"
	case l.Returned():
		return "returned"
	}
	return ""
}

// limitState is a Result's limit's state word: closed by its closers,
// accepted by the owner, or open, neither.
func (r renderer) limitState(l replay.Limit) string {
	switch {
	case l.Closed():
		return "closed by " + r.closers(l.Closers)
	case l.Accepted():
		return "accepted by " + l.Hand.Actor
	}
	return "open"
}

// ranOn names, for one closer of a line, each Hypothesis it ran on: its
// verdict when it judged it, and none when only its subject ran. A verdict
// is per Hypothesis, never per line, so it never says which line failed.
func (r renderer) ranOn(res string, hyps []string) string {
	n := r.p.Nodes[res]
	var parts []string
	for _, h := range hyps {
		switch v, ok := n.Verdicts[h]; {
		case ok:
			parts = append(parts, "ran on "+r.withAlias(h)+": "+v)
		case r.p.Covers(res, h):
			parts = append(parts, "ran on "+r.withAlias(h))
		}
	}
	return strings.Join(parts, "; ")
}

// limitLine is one limit as hyp show and render --markdown print it:
// R-…#n, its state, and its text.
func (r renderer) limitLine(l replay.Limit) string {
	return fmt.Sprintf("%s %s — %s", r.limitLabel(l.Result, l.Key), r.limitState(l), l.Text)
}

// experiments lists every Experiment under question q, grouped by parent
// Hypothesis, then each measuring Experiment under q itself. The tests line
// names its targets; the judged line gives each one's verdict from the
// Experiment's non-retracted Results, then the round's marks and each
// uncontained Result. A measuring Experiment has a measures line instead.
func (r renderer) experiments(w func(string, ...any), q string) {
	p := r.p
	first := true
	for _, e := range p.Nodes[q].Children {
		if p.Nodes[e].Kind != event.KindExperiment {
			continue
		}
		if first {
			w("experiments:")
			first = false
		}
		x := p.Exps[e]
		w("  %s measures (%s)", r.id(e), x.Status)
	}
	for _, h := range p.Nodes[q].Children {
		if p.Nodes[h].Kind != event.KindHypothesis {
			continue
		}
		for _, e := range p.Nodes[h].Children {
			en := p.Nodes[e]
			if en.Kind != event.KindExperiment {
				continue
			}
			if first {
				w("experiments:")
				first = false
			}
			x := p.Exps[e]
			tested := x.Targets
			w("  %s tests %s", r.id(e), r.ids(tested))
			got := x.Judged
			if len(x.Live) == 0 {
				judged := "pending"
				if x.Status != replay.ExpPlanned {
					judged = string(x.Status)
				}
				w("    judged: %s", judged)
				continue
			}
			var parts []string
			for _, t := range tested {
				if v, ok := got[t]; ok {
					parts = append(parts, r.id(t)+"="+v)
				}
			}
			line := "    judged: " + strings.Join(parts, ", ")
			m := r.marks(e)
			for _, res := range x.Live {
				if len(resOf(p, res).Uncontained) > 0 {
					m = append(m, "uncontained "+r.id(res))
				}
			}
			if len(m) > 0 {
				line += " [" + strings.Join(m, "; ") + "]"
			}
			w("%s", line)
		}
	}
}
