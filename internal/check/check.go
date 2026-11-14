// Package check implements hyp check: unparseable lines, rejected events and
// warnings about the log.
package check

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/dat9uy/hyp/internal/claim"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/store"
)

// Run checks the store at root.
func Run(root string) (report []string, errorsFound int) { return RunHistory(store.Single(root)) }

// RunHistory checks a history: every store of it, read as one log. Errors
// are unparseable lines and rejected events; everything else is a warning.
// The last line is the summary.
func RunHistory(h *store.History) (report []string, errorsFound int) {
	evs, lineErrs, err := h.Events()
	if err != nil {
		return []string{"error: " + err.Error(), "check: 1 errors, 0 warnings"}, 1
	}
	p := h.Replay(evs, h.Hashes(), "")
	display := func(id string) string {
		if n := p.Nodes[id]; n != nil {
			return event.DisplayID(n.Kind, id)
		}
		return id
	}

	for _, le := range lineErrs {
		report = append(report, fmt.Sprintf("error: %s:%d %v", le.File, le.Line, le.Err))
	}
	for _, r := range p.Rejections {
		in := ""
		if len(h.Stores) > 1 {
			in = " in store " + replay.StoreName(r.Store)
		}
		report = append(report, fmt.Sprintf("error: rejected %s%s: %s", display(r.EventID), in, r.Reason))
	}
	errorsFound = len(report)

	var warns []string
	for _, ev := range evs {
		for _, ref := range references(ev) {
			if ref > ev.ID {
				warns = append(warns, fmt.Sprintf("warn: %s references %s, which sorts later", ev.ID, ref))
			}
		}
	}
	for _, ev := range p.Applied {
		if ev.Type == event.TypeReaffirm && reaffirmNoEffect(p, ev.Node) {
			warns = append(warns, fmt.Sprintf("warn: reaffirm %s had no effect (target still unsupported)", ev.ID))
		}
	}
	// A Result applied under a closed Experiment in a merged log: the live
	// Result wins, so the Experiment reads run (spec §7.6).
	var closedRun []string
	for id, x := range p.Exps {
		if x.Status == replay.ExpRun && x.Close != "" {
			closedRun = append(closedRun, id)
		}
	}
	sort.Strings(closedRun)
	for _, id := range closedRun {
		x := p.Exps[id]
		warns = append(warns, fmt.Sprintf("warn: %s is %s but reads run: its Result %s applied after the close",
			display(id), x.Close, display(x.Live[0])))
	}
	// One line per pin that disagrees with the current recipe hashes of
	// the Result's own store.
	hashes, problems := map[string]map[string]string{}, []string{}
	for _, s := range h.Stores {
		m, pr := recipe.Scan(h.Dir(s))
		hashes[s] = m
		for _, x := range pr {
			if !slices.Contains(problems, x) {
				problems = append(problems, x)
			}
		}
	}
	for _, id := range resultIDs(p) {
		for _, pin := range p.Results[id].Pins {
			now, ok := hashes[p.Nodes[id].Store][pin.Ref.String()]
			if ok && now == pin.Hash {
				continue
			}
			if pin.Pointer == "" {
				warns = append(warns, fmt.Sprintf("warn: questionable evidence %s (recipe %s changed)", display(id), pin.Ref.Name))
				continue
			}
			if !ok {
				now = "missing"
			}
			warns = append(warns, fmt.Sprintf("warn: questionable evidence %s (%s: recorded %s, now %s)", display(id), pin.Pointer, pin.Hash, now))
		}
	}
	// A Result whose verdict script hyp cannot vouch for (spec §7.6).
	for _, id := range resultIDs(p) {
		for _, why := range p.Results[id].Uncontained {
			warns = append(warns, fmt.Sprintf("warn: uncontained %s (%s)", display(id), why))
		}
	}
	// A Result whose text disagrees with the files it cites (spec §7.6).
	for _, id := range resultIDs(p) {
		warns = append(warns, claim.Warnings(h.StoreDir(p, id), p, id)...)
	}
	// Two questions whose candidates assume each other's Hypotheses hold
	// each other's select (spec §8.1).
	for _, c := range p.PremiseCycles() {
		warns = append(warns, fmt.Sprintf("warn: premise cycle between %s and %s; remove a depends_on edge to break it", display(c[0]), display(c[1])))
	}
	// A line's cite keeps what it meant when the line was written; a later
	// change in the cited key's state is only this warning (ADR 0020).
	for _, l := range p.Lines {
		for _, c := range l.Cites {
			if c.Now != c.Meaning {
				name := p.KeyLabel(c.Key)
				if p.Nodes[c.Key] != nil {
					name = display(c.Key)
				}
				warns = append(warns, fmt.Sprintf("warn: %s cites %s, which was %s when the line was written and is %s now", p.KeyLabel(l.Key), name, c.Meaning, c.Now))
			}
		}
	}
	// A recipe left out of the hashes makes its Results questionable; say why.
	for _, pr := range problems {
		warns = append(warns, "warn: "+pr)
	}
	report = append(report, warns...)
	report = append(report, fmt.Sprintf("check: %d errors, %d warnings", errorsFound, len(warns)))
	return report, errorsFound
}

// references lists the ids an event points at, besides its own.
func references(ev event.Event) []string {
	var refs []string
	if ev.Node != ev.ID && ev.Node != "" {
		refs = append(refs, ev.Node)
	}
	var p struct {
		Parent     string            `json:"parent"`
		Targets    []string          `json:"targets"`
		Hypothesis string            `json:"hypothesis"`
		Select     string            `json:"select"`
		PrevSelect string            `json:"prev_select"`
		To         string            `json:"to"`
		Type       string            `json:"type"`
		Verdicts   map[string]string `json:"verdicts"`
	}
	if json.Unmarshal(ev.Payload, &p) != nil {
		return refs
	}
	refs = append(refs, p.Parent, p.Hypothesis, p.Select, p.PrevSelect)
	refs = append(refs, p.Targets...)
	if p.Type != event.EdgeExternalRef {
		refs = append(refs, p.To)
	}
	for h := range p.Verdicts {
		refs = append(refs, h)
	}
	out := refs[:0]
	for _, r := range refs {
		if r != "" {
			out = append(out, r)
		}
	}
	return out
}

// reaffirmNoEffect reports whether node is still stale because a depends_on
// target is not supported, which a reaffirm cannot fix (spec §7.3).
func reaffirmNoEffect(p replay.Projection, node string) bool {
	shown, _ := p.Display(node)
	s := shown.Stale
	if s == nil {
		return false
	}
	return s.Reason != replay.StaleReasonStale && s.Reason != replay.StaleReasonChanged
}

func resultIDs(p replay.Projection) []string {
	var out []string
	for _, ev := range p.Applied {
		if ev.Type == event.TypeResultRecord {
			out = append(out, ev.ID)
		}
	}
	return out
}
