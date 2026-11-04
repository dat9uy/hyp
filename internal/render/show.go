package render

import (
	"encoding/json"
	"fmt"
	"slices"
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
	crumbs := make([]string, 0, len(ancestors)+1)
	for i := len(ancestors) - 1; i >= 0; i-- {
		crumbs = append(crumbs, r.id(ancestors[i]))
	}
	w("path: %s", strings.Join(append(crumbs, r.id(id)), " > "))

	ns := p.Status[id]
	status := ns.Status
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
	if len(ns.Suggested) > 0 {
		w("suggested: %s", r.ids(ns.Suggested))
	}
	if n.Kind == event.KindDesignQuestion {
		rec, ok := p.CurrentSelect(id)
		if ok {
			kind := ""
			if k := SelectKind(rec); k != "" {
				kind = " (" + k + ")"
			}
			w("selected: %s%s — %s", r.id(rec.Hyp), kind, rec.Why)
			if body := p.Nodes[rec.Hyp].Body; body != "" {
				w("design: %s", body)
			}
		}
		if rec.Verifier != "" {
			w("verifier: %s — %s", rec.Verifier, rec.VerifierWhy)
		}
		if rec.RequestWhy != "" {
			w("re-verify requested: %s", rec.RequestWhy)
		}
		if ls := p.Limits[id]; len(ls) > 0 {
			w("limits:")
			for _, l := range ls {
				w("  %s", r.limitLine(l, rec.Hyp))
			}
		}
		r.experiments(w, id)
		r.dossier(w, id, rec.Hyp)
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
		if n.DecidedBy != "" {
			w("decided by: %s", n.DecidedBy)
		}
		if len(n.Targets) > 0 {
			w("targets: %s", r.ids(n.Targets))
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
			res, _, _ := event.ParseLimitKey(k)
			w("closes: %s", r.limitLabel(res, k))
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
			cs := p.Status[c]
			line := fmt.Sprintf("  %s %s", r.id(c), cs.Status)
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
		for _, ev := range p.Applied {
			if ev.Type != event.TypeResultRecord || p.Status[ev.ID].Status == replay.ResRetracted {
				continue
			}
			rn := p.Nodes[ev.ID]
			for _, pr := range preds {
				if _, ok := rn.Verdicts[pr]; ok {
					line := fmt.Sprintf("  %s (%s) %s", r.id(ev.ID), r.id(rn.Parent), r.verdicts(rn.Verdicts))
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
		old, had := before.Status[ev.ID]
		cur := after.Status[ev.ID]
		oldText := "absent"
		if had {
			oldText = statusText(r, old, before.Results[ev.ID])
		}
		curText := statusText(r, cur, after.Results[ev.ID])
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

// statusText is a node's status as one phrase; res is its Result record, or
// nil for any other node.
func statusText(r renderer, s replay.NodeStatus, res *replay.Res) string {
	t := s.Status
	if s.Reason != "" {
		t += "/" + s.Reason
	}
	if s.Stale != nil {
		t += " stale ← " + r.chain(s.Stale)
	}
	if res != nil && res.Questionable {
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

// limitLine is one limit as hyp show and render --markdown print it:
// R-…#n, its state, and its text. hyp is the question's selected Hypothesis.
func (r renderer) limitLine(l replay.Limit, hyp string) string {
	state := l.State
	switch l.State {
	case replay.LimitOpen:
		if hyp != "" && !l.Gating {
			state += ", does not gate (judges " + r.ids(l.Hyps) + ")"
		}
	case replay.LimitClosed:
		state += " by " + r.id(l.By)
	case replay.LimitAccepted:
		state += " by " + l.By
	}
	return fmt.Sprintf("%s %s — %s", r.limitLabel(l.Result, l.Key), state, l.Text)
}

// dossier lists question q's Dossier: each live Hypothesis with its status,
// the Results with a verdict on it, and the links going out from both. It
// marks a contradicts that alone keeps a rival from being a trade-off
// candidate, so the verifier sees which verdict decides whether the owner must
// sign.
func (r renderer) dossier(w func(string, ...any), q, selected string) {
	d := r.p.Dossiers[q]
	if d == nil {
		return
	}
	w("dossier:")
	links := func(from string) {
		for _, it := range d.Items {
			if it.Kind == replay.DossierLink && it.Node == from {
				w("    link %s %s %s", r.id(from), it.Word, r.id(it.To))
			}
		}
	}
	for _, it := range d.Items {
		if it.Kind != replay.DossierHyp {
			continue
		}
		h := it.Node
		line := fmt.Sprintf("  %s %s", r.id(h), r.p.Status[h].Status)
		if r.p.Status[h].Stale != nil {
			line += " STALE"
		}
		if h == selected {
			line += " (selected)"
		}
		w("%s %s", line, Cut(r.p.Nodes[h].Title, 60))
		links(h)
		for _, res := range d.Items {
			if res.Kind != replay.DossierResult || res.To != h {
				continue
			}
			mark := ""
			if slices.Contains(d.Deciding[h], res.Node) {
				mark = " [decides the owner requirement: alone keeps this rival from being a candidate]"
			}
			w("    %s %s%s", r.id(res.Node), res.Word, mark)
			links(res.Node)
		}
	}
}

// experiments lists every Experiment under question q, grouped by parent
// Hypothesis. The tests line names the parent and its targets; the judged
// line gives each one's verdict from the Experiment's non-retracted Results,
// "unjudged" for a tested Hypothesis no Result judged, then the round's marks
// and each uncontained Result.
func (r renderer) experiments(w func(string, ...any), q string) {
	p := r.p
	first := true
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
			tested := append([]string{h}, en.Targets...)
			w("  %s tests %s", r.id(e), r.ids(tested))
			x := p.Exps[e]
			got := x.Judged
			if len(x.Live) == 0 {
				judged := "pending"
				if x.Status != replay.ExpPlanned {
					judged = x.Status
				}
				w("    judged: %s", judged)
				continue
			}
			var parts []string
			for _, t := range tested {
				v, ok := got[t]
				if !ok {
					v = "unjudged"
				}
				parts = append(parts, r.id(t)+"="+v)
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
