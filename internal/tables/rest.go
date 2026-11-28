package tables

import (
	"slices"

	"github.com/dat9uy/hyp/internal/fault"
)

// A Goal the owner adds with --from a Goal in work rests on that Goal's
// select. It stays written until the Goal it rests on, its premise, is
// selected or landed on a commit; it then opens on that commit, as its H0,
// when each of its examples gives there the verdict its line needs. The
// verdicts are cells of the clone's trace, which the run that selects the
// premise measures. A Goal added with --from an ended Goal does not rest:
// it opens on its own H0.

// The states of a rest.
const (
	RestWaits      = "waits"      // the premise is in work with no select
	RestEnded      = "ended"      // the premise was closed, or landed with no commit
	RestUnmeasured = "unmeasured" // an example has no cell on the premise's commit
	RestWrong      = "wrong"      // an example gives the wrong verdict on the premise's commit
	RestOn         = "on"         // every example gives its verdict there: the Goal stands on it
)

// RestCut marks a Result of a resting Goal that stands on no commit the
// premise is selected or landed on now: the owner cut the select it ran on.
// Like a refused Result, it counts for nothing.
const RestCut = "Cut"

// Rest is what a resting Goal reads of its premise: the premise, the commit
// it is selected or landed on, and how the Goal's examples measure there.
type Rest struct {
	Premise string // the premise's Goal id
	State   string // RestWaits, RestEnded, RestUnmeasured, RestWrong or RestOn
	Landed  bool   // the commit is the premise's land, not its select
	Closed  bool   // RestEnded by a close
	Commit  string
	RowTree string
	Wrong   []RestVerdict // RestWrong: each example whose verdict is not the one its line needs
	Missing []string      // RestUnmeasured: the keys of the examples with no cell there
}

// RestVerdict is one example's verdict on the premise's commit, and the
// verdict its line needs.
type RestVerdict struct {
	Key, Verdict, Needs string
}

// Premise is the Goal g rests on, or nil when g rests on none.
func (t Tables) Premise(g *Goal) *Goal {
	if !g.Rests {
		return nil
	}
	return t.Goals[g.From]
}

// premiseCommit is the commit the premise p holds for a Goal that rests on
// it: its land's commit, else its select's, with its row tree.
func premiseCommit(p *Goal) (commit, rowTree string, landed, ok bool) {
	switch {
	case p.Land != nil && p.Land.Commit != "":
		return p.Land.Commit, p.Land.RowTree, true, true
	case p.State == StateEnded:
		return "", "", false, false
	case p.Select != nil:
		return p.Select.Commit, p.Select.RowTree, false, true
	}
	return "", "", false, false
}

// rest calculates the rest of g, a Goal in work that rests on a premise,
// after the premise's select is calculated. A Goal on its premise's commit
// takes that commit as its H0, each example's cell there as its H0 cell,
// and keeps only the Results that stand on that commit; any other rest keeps
// the Goal written, and every Result it holds counts for nothing.
func (t Tables) rest(g *Goal) {
	p := t.Premise(g)
	if p == nil || g.State == StateEnded {
		return
	}
	r := &Rest{Premise: p.ID}
	g.Rest = r
	commit, tree, landed, ok := premiseCommit(p)
	if !ok {
		r.State, r.Closed = RestWaits, p.Close != nil
		if p.State == StateEnded {
			r.State = RestEnded
		}
		g.cutResults("")
		g.State = StateWritten
		return
	}
	r.Commit, r.RowTree, r.Landed = commit, tree, landed
	cells := map[string]LiveCell{}
	for _, x := range g.Placed() {
		l, _ := g.line(x.Line)
		tc, ok := g.TraceCell(x.Hash, tree)
		if !ok {
			r.Missing = append(r.Missing, x.Key)
			continue
		}
		if tc.Verdict != Needs(l.Kind) {
			r.Wrong = append(r.Wrong, RestVerdict{Key: x.Key, Verdict: tc.Verdict, Needs: Needs(l.Kind)})
		}
		cells[x.Key] = LiveCell{Key: x.Key, Hash: x.Hash, Probe: tc.Probe, Output: tc.Output, Stderr: tc.Stderr,
			Exit: tc.Exit, Verdict: tc.Verdict, By: tc.By}
	}
	switch {
	case len(r.Wrong) > 0:
		r.State = RestWrong
	case len(r.Missing) > 0:
		r.State = RestUnmeasured
	default:
		r.State = RestOn
	}
	if r.State != RestOn {
		g.cutResults("")
		g.State = StateWritten
		return
	}
	g.Base, g.BaseTree, g.premise = commit, tree, cells
	g.cutResults(commit)
	g.State = StateWritten
	if g.everyLinePlaced() {
		g.State = StateOpen
	}
}

// everyLinePlaced reports whether every line of g holds a placed example.
func (g *Goal) everyLinePlaced() bool {
	placed := map[string]bool{}
	for _, p := range g.Placed() {
		placed[p.Line] = true
	}
	for _, l := range g.Lines {
		if !placed[l.Key] {
			return false
		}
	}
	return len(g.Lines) > 0
}

// cutResults marks RestCut every Result of g that does not stand on commit:
// a Result on another commit whose run did not read it among the commits
// its commit descends from. With no commit, it marks every Result. A Result
// that stands on commit is measured again against it as H0.
func (g *Goal) cutResults(commit string) {
	for _, r := range g.Results {
		if r.Refused != "" {
			continue
		}
		on := commit != "" && (r.Commit == commit || (r.Read != nil && slices.Contains(r.Read.Above, commit)))
		if !on {
			r.Refused = RestCut
			continue
		}
		r.wasH0 = r.RowTree == g.BaseTree
	}
}

// GuardLandedFirst is the Landed first guard of land: a Goal that rests on
// a premise lands only after the premise landed on a commit. holds reports
// whether the commit landed holds the premise's land commit; replay passes
// nil, since that half reads git and the act checked it.
func GuardLandedFirst(t Tables, g *Goal, commit string, holds func(premise string) (bool, error)) error {
	p := t.Premise(g)
	if p == nil {
		return nil
	}
	if p.Land == nil || p.Land.Commit == "" {
		return fault.Refuse(fault.GuardLandedFirst, fault.PremiseNotLanded, g.ID, p.ID, p.State)
	}
	if holds == nil || commit == "" {
		return nil
	}
	ok, err := holds(p.Land.Commit)
	if err != nil {
		return err
	}
	if !ok {
		return fault.Refuse(fault.GuardLandedFirst, fault.LandWithoutPremise, commit, p.ID, p.Land.Commit)
	}
	return nil
}

// UseRestBases gives each Goal of t that rests on a premise the H0 the
// calculated tables c show it standing on: the commit its premise is
// selected or landed on, when there is one. A Goal whose premise has no such
// commit keeps the base its goal.add names.
func (t Tables) UseRestBases(c Tables) {
	for id, g := range t.Goals {
		cg := c.Goals[id]
		if !g.Rests || cg == nil || cg.Rest == nil || cg.Rest.Commit == "" {
			continue
		}
		g.Base, g.BaseTree = cg.Rest.Commit, cg.Rest.RowTree
	}
}

// openAtReplay reports whether replay lets an act of g pass Open: replay
// cannot see the select of the premise a Goal rests on, which a clone's
// Results calculate, so for such a Goal the act checked Open where the
// Goal stands.
func (g *Goal) openAtReplay() bool { return g.Rests && g.Rest == nil }
