// Package check implements hyp check: unparseable lines, rejected events and
// warnings about the log.
package check

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/dat9uy/hyp/internal/claim"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/record"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/store"
)

// Run checks the store at root. Errors are unparseable lines and rejected
// events; everything else is a warning. The last line is the summary.
func Run(root string) (report []string, errorsFound int) { return run(root, nil) }

// RunStore checks the store discovery found. In a record repository it
// checks the record's copy and also warns when the store's work-tree folder
// holds an event the record lacks, which hyp import adds, and for each
// checked-out bundle whose hash differs from a pin of the same ref.
func RunStore(s *record.Store) (report []string, errorsFound int) {
	root, err := s.Root()
	if err != nil {
		return []string{"error: " + err.Error(), "check: 1 errors, 0 warnings"}, 1
	}
	if !s.RecordMode() {
		return run(root, nil)
	}
	var extra []string
	if evs, _, err := store.Events(root); err == nil {
		have := map[string]bool{}
		for _, ev := range evs {
			have[ev.ID] = s.InRecord
		}
		if wt, err := s.WorkTreeEventFiles(); err == nil && len(wt) > 0 {
			lacks := 0
			for actor, data := range wt {
				for _, ev := range store.ParseEvents(actor, data) {
					if !have[ev.ID] {
						lacks++
					}
				}
			}
			if lacks > 0 {
				extra = append(extra, fmt.Sprintf("warn: the work-tree store at %s holds %d events the record lacks; run hyp import %s", s.Dir, lacks, s.Dir))
			}
		}
	}
	if p, _, err := store.Load(root); err == nil {
		checkedOut, problems := recipe.Scan(s.Dir)
		seen := map[string]bool{}
		drift := func(by, ref, pinned string) {
			now, ok := checkedOut[ref]
			if !ok || pinned == "" || now == pinned || seen[ref+" "+now] {
				return
			}
			seen[ref+" "+now] = true
			extra = append(extra, fmt.Sprintf("warn: checked-out bundle %s hashes %s, not %s as %s pins it", ref, now, pinned, by))
		}
		for _, ev := range p.Applied {
			n := p.Nodes[ev.ID]
			if n == nil {
				continue
			}
			switch n.Kind {
			case event.KindExperiment:
				by := event.DisplayID(n.Kind, n.ID)
				if n.Recipe != event.RecipeNone {
					drift(by, n.Recipe, n.RecipeHash)
				}
				if n.Rubric != "" && n.Rubric != event.RubricReading {
					drift(by, n.Rubric, n.RubricHash)
				}
			case event.KindResult:
				for _, pin := range p.Results[n.ID].Pins {
					if pin.Pointer != "" {
						drift(event.DisplayID(n.Kind, n.ID), pin.Ref.String(), pin.Hash)
					}
				}
			}
		}
		for _, pr := range problems {
			extra = append(extra, "warn: checked-out "+pr)
		}
	}
	return run(root, extra)
}

func run(root string, extra []string) (report []string, errorsFound int) {
	evs, lineErrs, err := store.Events(root)
	if err != nil {
		return []string{"error: " + err.Error(), "check: 1 errors, 0 warnings"}, 1
	}
	p, _, err := store.Load(root)
	if err != nil {
		return []string{"error: " + err.Error(), "check: 1 errors, 0 warnings"}, 1
	}
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
		report = append(report, fmt.Sprintf("error: rejected %s: %s", display(r.EventID), r.Reason))
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
	// One line per pin that disagrees with the current recipe hashes.
	hashes, problems := recipe.Scan(root)
	for _, id := range resultIDs(p) {
		for _, pin := range p.Results[id].Pins {
			now, ok := hashes[pin.Ref.String()]
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
		warns = append(warns, claim.Warnings(root, p, id)...)
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
	warns = append(warns, extra...)
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
