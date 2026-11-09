// Package signals lists a tree's design marks from its event log: one-rival,
// D1 and unjudged. Each mark is also a write-time refusal or warning, or a
// mark that hyp status and hyp show print; this package lists them for the
// whole tree. Progress on a question is the ledger's, not a signal's (ADR
// 0017). It is pure, like replay: no file I/O and no clock.
package signals

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// Report is what Compute returns.
type Report struct {
	Rounds  []Round  `json:"rounds"`
	Signals []Signal `json:"signals"`
}

// Round is one Experiment, with the facts read at its add and its Results.
type Round struct {
	Exp        string     `json:"exp"`
	Cut        string     `json:"cut"` // the event just before the Experiment was added
	Parent     string     `json:"parent"`
	Targets    []string   `json:"targets"`
	Question   string     `json:"question"`
	QuestionID string     `json:"question_id"`
	AtAdd      [][]string `json:"at_add"` // [H, status] for the parent, then each live rival
	OneRival   bool       `json:"one_rival"`
	D1         [][]string `json:"d1"`
	Results    []Result   `json:"results"`
	Unjudged   []string   `json:"unjudged,omitempty"` // targets live at add that no Result judged
}

// Result is one Result of a round that is not retracted.
type Result struct {
	ID    string     `json:"id"`
	Event string     `json:"event"`
	Moves [][]string `json:"moves"` // [H, status before, status after]
}

// Signal is one fired signal. Key is stable across calls, so a caller can
// report each signal once.
type Signal struct {
	Key    string `json:"key"`
	Round  int    `json:"round"`
	Detail string `json:"detail"`
}

// word is Hypothesis h's status in hyps, or "" when hyps does not hold it.
func word(hyps map[string]*replay.Hyp, h string) string {
	if x := hyps[h]; x != nil {
		return string(x.Status)
	}
	return ""
}

// Compute replays evs once and returns every round and every signal it
// shows. Only applied events count; a refused event changes nothing. Facts
// that hold only at the end of the log (retracts, one-rival, unjudged) are
// read from the final state's projection.
func Compute(evs []event.Event) Report {
	st := replay.NewState()
	display := func(id string) string { return event.DisplayID(st.Nodes[id].Kind, id) }
	var rounds []*Round
	var exps []string // Experiment id of each round
	byExp := map[string]*Round{}
	type recorded struct {
		node  string
		round *Round
		res   Result
	}
	var recs []recorded
	prev := ""
	for _, ev := range evs {
		// Read the state before ev; a refused ev discards what was read.
		before := st.Hyps
		var live, d1 []string
		var add event.NodeAdd
		var rec event.ResultRecord
		switch ev.Type {
		case event.TypeNodeAdd:
			_ = event.DecodePayload(ev, &add)
			if add.Kind == event.KindExperiment {
				d1 = st.NotRivals(append([]string{add.Parent}, add.Targets...))
			}
		case event.TypeResultRecord:
			_ = event.DecodePayload(ev, &rec)
			if r := byExp[rec.Parent]; r != nil {
				live = st.LiveRivals(r.QuestionID, "")
			}
		}
		if st.Step(ev, nil) != nil {
			continue
		}
		switch {
		case ev.Type == event.TypeNodeAdd && add.Kind == event.KindExperiment:
			q := st.Nodes[add.Parent].Parent
			live = st.RivalsAtAdd[ev.Node]
			r := &Round{Exp: display(ev.Node), Cut: prev, Parent: display(add.Parent), Targets: []string{},
				Question: display(q), QuestionID: q, AtAdd: [][]string{}, D1: [][]string{}, Results: []Result{}}
			for _, t := range add.Targets {
				r.Targets = append(r.Targets, display(t))
			}
			for _, h := range append([]string{add.Parent}, live...) {
				r.AtAdd = append(r.AtAdd, []string{display(h), word(before, h)})
			}
			for _, h := range d1 {
				r.D1 = append(r.D1, []string{display(h), word(before, h)})
			}
			rounds = append(rounds, r)
			exps = append(exps, ev.Node)
			byExp[ev.Node] = r
		case ev.Type == event.TypeResultRecord:
			after := st.Hyps
			r := byExp[rec.Parent]
			for _, t := range st.Nodes[ev.Node].VerdictOrder {
				if !slices.Contains(live, t) {
					live = append(live, t)
				}
			}
			res := Result{ID: display(ev.Node), Event: ev.ID, Moves: [][]string{}}
			for _, h := range live {
				res.Moves = append(res.Moves, []string{display(h), word(before, h), word(after, h)})
			}
			recs = append(recs, recorded{ev.Node, r, res})
		}
		prev = ev.ID
	}

	full := st.Projection()
	for _, x := range recs {
		if full.Results[x.node].Status != replay.ResRetracted {
			x.round.Results = append(x.round.Results, x.res)
		}
	}

	found := map[string]Signal{}
	add := func(key string, round int, detail string) { found[key] = Signal{key, round, detail} }
	out := Report{Rounds: []Round{}, Signals: []Signal{}}
	for i, r := range rounds {
		x := full.Exps[exps[i]]
		r.OneRival = x.OneRival
		for _, h := range x.Unjudged {
			r.Unjudged = append(r.Unjudged, display(h))
		}
		if r.OneRival {
			add("one-rival:"+r.Exp, i+1, "one-rival "+r.Exp+" "+oneRivalText(r))
		}
		if len(r.D1) > 0 {
			var sts []string
			for _, hs := range r.D1 {
				sts = append(sts, hs[0]+"="+hs[1])
			}
			add("D1:"+r.Exp, i+1, "D1 "+r.Exp+" at add: "+strings.Join(sts, ", "))
		}
		if len(r.Unjudged) > 0 {
			add("unjudged:"+r.Exp, i+1, "unjudged "+r.Exp+" targets "+strings.Join(r.Unjudged, ",")+", which no Result judged")
		}
		out.Rounds = append(out.Rounds, *r)
	}
	keys := make([]string, 0, len(found))
	for k := range found {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out.Signals = append(out.Signals, found[k])
	}
	return out
}

// oneRivalText names the statuses at add, the targets and the rivals the
// Experiment could have targeted.
func oneRivalText(r *Round) string {
	text := fmt.Sprintf("at add: %s=%s (parent)", r.AtAdd[0][0], r.AtAdd[0][1])
	var rivals []string
	for _, hs := range r.AtAdd[1:] {
		text += fmt.Sprintf(", %s=%s", hs[0], hs[1])
		rivals = append(rivals, hs[0])
	}
	targets := strings.Join(r.Targets, ",")
	if targets == "" {
		targets = "none"
	}
	return text + "; targets: " + targets + "; could target: " + strings.Join(rivals, ",")
}
