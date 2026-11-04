// Package signals computes the waste tripwire's signals from an event log:
// wasted, no-movement, one-rival, D1, spin, verifier-loop and unjudged. Each signal line
// names the statuses or verdicts its rule used. It is pure, like replay: no
// file I/O and no clock.
package signals

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// Report is what Compute returns. Field names match the rounds JSON that
// count-wasted.py wrote before the signals moved into hyp.
type Report struct {
	Rounds  []Round  `json:"rounds"`
	Signals []Signal `json:"signals"`
}

// Round is one Experiment, judged by its Results.
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
	Pending    bool       `json:"pending"`
	Wasted     bool       `json:"wasted"`
	Closed     string     `json:"closed,omitempty"` // "superseded by E-…" or "withdrawn": closed with no Result, so neither pending nor wasted
	Unjudged   []string   `json:"unjudged,omitempty"` // targets live at add that no Result judged
	Line       string     `json:"line"`               // the round as one line: its judgement, then add-time signals
}

// Result is one Result of a round that is not retracted.
type Result struct {
	ID    string     `json:"id"`
	Event string     `json:"event"`
	Moves [][]string `json:"moves"` // [H, status before, status after]
	Moved bool       `json:"moved"`
}

// Signal is one fired signal. Key is stable across calls, so a caller can
// report each signal once.
type Signal struct {
	Key    string `json:"key"`
	Round  int    `json:"round"`
	Detail string `json:"detail"`
}

// result is one result.record that is not retracted, for spin. Its statuses
// are from when it was recorded, as for a Round's Results.
type result struct {
	id, exp       string
	targets       []string
	before, after map[string]replay.NodeStatus
}

// Compute replays evs once and returns every round and every signal it
// shows. Only applied events count; a refused event changes nothing. Facts
// that hold only at the end of the log (retracts, one-rival, unjudged, the
// current select) are read from the final state's projection.
func Compute(evs []event.Event) Report {
	st := replay.NewState()
	display := func(id string) string { return event.DisplayID(st.Nodes[id].Kind, id) }
	var rounds []*Round
	var exps []string // Experiment id of each round
	byExp := map[string]*Round{}
	// recorded is one result.record as stepped; it counts once the log shows
	// it was not retracted.
	type recorded struct {
		node  string
		round *Round
		res   Result
		spin  result
	}
	var recs []recorded
	selAt := map[string]*selState{} // result.record event → its question's select state just after it
	prev := ""
	for _, ev := range evs {
		// Read the state before ev; a refused ev discards what was read.
		before := st.Derived
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
				r.AtAdd = append(r.AtAdd, []string{display(h), before[h].Status})
			}
			for _, h := range d1 {
				r.D1 = append(r.D1, []string{display(h), before[h].Status})
			}
			rounds = append(rounds, r)
			exps = append(exps, ev.Node)
			byExp[ev.Node] = r
		case ev.Type == event.TypeResultRecord:
			targets := st.Nodes[ev.Node].VerdictOrder
			after := st.Derived
			r := byExp[rec.Parent]
			selAt[ev.ID] = viewOf(r.QuestionID, st.Decisions[r.QuestionID])
			for _, t := range targets {
				if !slices.Contains(live, t) {
					live = append(live, t)
				}
			}
			res := Result{ID: display(ev.Node), Event: ev.ID, Moves: [][]string{}}
			for _, h := range live {
				b, a := before[h].Status, after[h].Status
				res.Moves = append(res.Moves, []string{display(h), b, a})
				res.Moved = res.Moved || b != a
			}
			recs = append(recs, recorded{ev.Node, r, res,
				result{id: display(ev.Node), exp: rec.Parent, targets: targets, before: before, after: after}})
		}
		prev = ev.ID
	}

	full := st.Projection()
	nodes := full.Nodes
	retracted := map[string]bool{}
	for _, ev := range full.Applied {
		if ev.Type == event.TypeResultRetract {
			retracted[ev.Node] = true
		}
	}
	var results []result
	for _, x := range recs {
		if !retracted[x.node] {
			x.round.Results = append(x.round.Results, x.res)
			results = append(results, x.spin)
		}
	}
	for i, r := range rounds {
		x := full.Exps[exps[i]]
		r.OneRival = x.OneRival
		switch x.Status {
		case replay.ExpSuperseded:
			var by []string
			for _, e := range x.ReplacedBy {
				by = append(by, display(e))
			}
			r.Closed = "superseded by " + strings.Join(by, ",")
		case replay.ExpWithdrawn:
			r.Closed = "withdrawn"
		}
		r.Pending = len(x.Live) == 0 && r.Closed == ""
		for _, h := range x.Unjudged {
			r.Unjudged = append(r.Unjudged, display(h))
		}
	}

	found := map[string]Signal{}
	add := func(key string, round int, detail string) { found[key] = Signal{key, round, detail} }
	roundOf := map[string]int{}
	out := Report{Rounds: []Round{}, Signals: []Signal{}}
	for i, r := range rounds {
		roundOf[r.Exp] = i + 1
		r.Wasted = !r.Pending && r.Closed == ""
		for _, x := range r.Results {
			r.Wasted = r.Wasted && !x.Moved
		}
		if len(r.Unjudged) > 0 {
			add("unjudged:"+r.Exp, i+1, "unjudged "+r.Exp+" targets "+strings.Join(r.Unjudged, ",")+", which no Result judged")
		}
		parts := roundParts(r)
		line := strings.TrimRight(r.Exp+" "+parts[0][0]+" "+parts[0][1], " ")
		for _, pt := range parts[1:] {
			line += "; " + pt[0] + " " + pt[1]
			add(pt[0]+":"+r.Exp, i+1, pt[0]+" "+r.Exp+" "+pt[1])
		}
		if parts[0][0] == "wasted" {
			add("wasted:"+r.Exp, i+1, "wasted "+r.Exp+" "+parts[0][1])
		}
		r.Line = line
	}

	var judged []*Round
	for _, r := range rounds {
		if len(r.Results) > 0 {
			judged = append(judged, r)
		}
	}
	sort.SliceStable(judged, func(i, j int) bool {
		return last(judged[i]).Event < last(judged[j]).Event
	})
	for i := 1; i < len(judged); i++ {
		x, y := judged[i-1], judged[i]
		if !x.Wasted || !y.Wasted {
			continue
		}
		xr, yr := last(x), last(y)
		detail := fmt.Sprintf("no-movement %s,%s: %s, %s moved no live rival", x.Exp, y.Exp, xr.ID, yr.ID)
		if s := selAt[yr.Event]; s != nil {
			detail += "; " + s.describe(display) + "; the next experiment may not change anything"
		}
		add("no-movement:"+yr.ID, roundOf[y.Exp], detail)
	}

	// Only the Experiment's parent counts: rivals it merely targets are being tested, not spun on.
	type step struct{ rid, before, after string }
	byHyp := map[string][]step{}
	var hypOrder []string
	for _, x := range results {
		parent := nodes[x.exp].Parent
		if !slices.Contains(x.targets, parent) {
			continue
		}
		if _, ok := byHyp[parent]; !ok {
			hypOrder = append(hypOrder, parent)
		}
		byHyp[parent] = append(byHyp[parent], step{x.id, x.before[parent].Status, x.after[parent].Status})
	}
	expOf := map[string]string{}
	for _, x := range results {
		expOf[x.id] = display(x.exp)
	}
	for _, h := range hypOrder {
		seq := byHyp[h]
		for i := 2; i < len(seq); i++ {
			w := seq[i-2 : i+1]
			if w[0].before != w[0].after || w[1].before != w[1].after || w[2].before != w[2].after {
				continue
			}
			var moves []string
			for _, s := range w {
				moves = append(moves, fmt.Sprintf("%s→%s at %s", s.before, s.after, s.rid))
			}
			add("spin:"+display(h)+":"+w[2].rid, roundOf[expOf[w[2].rid]], "spin "+display(h)+" "+strings.Join(moves, ", "))
		}
	}

	under := func(n, q string) bool {
		for n != "" {
			if n == q {
				return true
			}
			nd := nodes[n]
			if nd == nil {
				return false
			}
			n = nd.Parent
		}
		return false
	}
	var questions []string
	for _, ev := range full.Applied {
		if ev.Type == event.TypeDQSelect && !slices.Contains(questions, ev.Node) {
			questions = append(questions, ev.Node)
		}
	}
	for _, q := range questions {
		s := viewOf(q, full.Decisions[q])
		if s == nil {
			continue
		}
		if len(s.chain) >= 2 {
			add("verifier-loop:"+display(q)+":"+s.chain[len(s.chain)-1].id, len(rounds), "verifier-loop "+s.describe(display))
		}
		if s.request == "" {
			continue
		}
		lastVerdict := s.chain[len(s.chain)-1].id
		fresh := false
		for _, ev := range full.Applied {
			if ev.ID <= lastVerdict || ev.ID >= s.request {
				continue
			}
			switch ev.Type {
			case event.TypeResultRecord, event.TypeResultRetract:
				fresh = fresh || under(ev.Node, q)
			case event.TypeEdgeAdd:
				// A link counts from either end: an external_ref from a Hypothesis under q points outside the tree.
				var p event.Edge
				_ = event.DecodePayload(ev, &p)
				fresh = fresh || under(ev.Node, q) || under(p.To, q)
			}
		}
		if !fresh {
			add("verifier-loop:"+display(q)+":"+s.request, len(rounds),
				"verifier-loop "+s.describe(display)+"; request-verify with no new Result, retract or link")
		}
	}

	for _, r := range rounds {
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

func last(r *Round) Result { return r.Results[len(r.Results)-1] }

// roundParts gives a round's judgement first, then its add-time signals, as
// (signal, text) pairs.
func roundParts(r *Round) [][2]string {
	var parts [][2]string
	switch {
	case r.Wasted:
		x := last(r)
		var moves []string
		for _, m := range x.Moves {
			moves = append(moves, fmt.Sprintf("%s %s→%s", m[0], m[1], m[2]))
		}
		parts = append(parts, [2]string{"wasted", x.ID + ": " + strings.Join(moves, ", ")})
	case r.Pending:
		parts = append(parts, [2]string{"pending", ""})
	case r.Closed != "":
		word, rest, _ := strings.Cut(r.Closed, " ")
		parts = append(parts, [2]string{word, rest})
	default:
		parts = append(parts, [2]string{"ok", ""})
	}
	if r.OneRival {
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
		text += "; targets: " + targets + "; could target: " + strings.Join(rivals, ",")
		parts = append(parts, [2]string{"one-rival", text})
	}
	if len(r.D1) > 0 {
		var st []string
		for _, hs := range r.D1 {
			st = append(st, hs[0]+"="+hs[1])
		}
		parts = append(parts, [2]string{"D1", "at add: " + strings.Join(st, ", ")})
	}
	return parts
}
