package tables

import (
	"path"
	"slices"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/spec"
)

// StopNone is the stop of a Goal at none of the three stops.
const StopNone = "none"

// VariantsDir is the folder, below a Goal's spec folder, that holds one
// variant config per potential Hypothesis.
const VariantsDir = "variants/"

// SpecDir is the spec folder of Goal id, as git names it.
func SpecDir(id string) string { return path.Join("goals", id, "specs") + "/" }

// Cause is the event of the act that caused a stop: a select, or the
// result.record of the run after which the stop held.
type Cause struct {
	Act  string // spec.ActSelect or spec.ActRun
	ID   string // the event id
	Time string // the event's time
}

// Standing is what hyp calculates about a Goal from the tables alone: what
// each run read of git is in its Result.
type Standing struct {
	Stop       string
	Cause      Cause  // the act that caused the stop; empty for StopNone
	Waits      string // who the Goal waits for: grill, owner or none
	Used       int    // N of "runs: N of 3 used"
	Rows       []Row
	Candidates []Row
}

// Calculate gives each Goal the select this clone's Results show, and
// returns t. No event holds that select: git holds only the owner's acts, so
// a Goal is selected on a commit when the Results of this clone's trace show
// its row as the one candidate, met on every line, that no unselect
// rejected, that holds a spec and writes no other Goal's folder (Spec and
// Own path), and, after an unselect, that a run recorded a new cell for
// (New cell). A clone that lacks those Results shows the Goal open. A Goal
// whose select an event of the older form holds, and an ended Goal, keep
// what replay gave them.
func Calculate(t Tables) Tables {
	for _, g := range t.Goals {
		calculateSelect(g)
	}
	return t
}

func calculateSelect(g *Goal) {
	if g.State != StateOpen || g.Select != nil {
		return
	}
	st := Stand(g)
	if st.Stop != StopNone || len(st.Candidates) != 1 || !g.HasNewCell() {
		return
	}
	c := st.Candidates[0]
	if !Selectable(g, c) {
		return
	}
	sel := &Select{RowTree: c.RowTree, Calculated: true}
	for _, r := range g.Results {
		if r.RowTree == c.RowTree && r.Refused == "" {
			sel.ID, sel.Commit, sel.Time, sel.Actor = r.ID, r.Commit, r.Time, r.Actor
		}
	}
	g.Select = sel
	g.State = StateSelected
}

// Stand calculates the standing of a Goal: its rows and candidates, and the
// stop with the act that caused it. A landed Goal keeps the stop
// its land names; an ended Goal waits for no one. It reads no git.
func Stand(g *Goal) Standing {
	st := Standing{Stop: StopNone, Waits: "none", Used: RunsUsed(g), Rows: g.Rows()}
	st.Candidates = Candidates(g)
	switch {
	case g.Land != nil:
		st.Stop, st.Cause = g.Land.Stop, causeOf(g, g.Land.Cause)
	case g.Close != nil && g.Select != nil:
		st.Stop = StopSelected
		st.Cause = selectCause(g.Select)
	case g.Close != nil:
		// a closed Goal keeps the stop its rows hold, and waits for no one
		if len(st.Candidates) >= 2 {
			st.Stop, st.Cause = StopTwoCandidates, lastRun(g)
		} else if len(st.Candidates) == 0 && budgetUsed(g) {
			st.Stop, st.Cause = StopBudgetUsed, lastRun(g)
		}
	case g.Select != nil:
		st.Stop, st.Waits = StopSelected, "owner"
		st.Cause = selectCause(g.Select)
	case len(st.Candidates) >= 2:
		st.Stop, st.Waits, st.Cause = StopTwoCandidates, "owner", lastRun(g)
	case len(st.Candidates) == 0 && budgetUsed(g):
		st.Stop, st.Waits, st.Cause = StopBudgetUsed, "owner", lastRun(g)
	case g.State == StateWritten:
		st.Waits = "owner"
	}
	return st
}

// selectCause is the act that caused the Selected stop: a select event of
// the older form, or for a calculated select the run whose Result made its
// row met on every line.
func selectCause(sel *Select) Cause {
	if sel.Calculated {
		return Cause{Act: spec.ActRun, ID: sel.ID, Time: sel.Time}
	}
	return Cause{Act: spec.ActSelect, ID: sel.ID, Time: sel.Time}
}

// lastRun is the run that caused a stop of the runs: the latest Result, in
// merge order, that counts, not refused by Stable. Only a
// run adds a row or a cell, so the latest one is the run after which the stop
// held.
func lastRun(g *Goal) Cause {
	for i := len(g.Results) - 1; i >= 0; i-- {
		if r := g.Results[i]; r.Refused == "" {
			return Cause{Act: spec.ActRun, ID: r.ID, Time: r.Time}
		}
	}
	return Cause{}
}

// LastRunOn is the latest Result on the row with rowTree, in merge order,
// that counts.
func LastRunOn(g *Goal, rowTree string) Cause {
	for i := len(g.Results) - 1; i >= 0; i-- {
		if r := g.Results[i]; r.RowTree == rowTree && r.Refused == "" {
			return Cause{Act: spec.ActRun, ID: r.ID, Time: r.Time}
		}
	}
	return Cause{}
}

// causeOf is the cause a land event names, as git alone holds it: the
// Goal's select event of the older form, or else the id alone. A Result is
// in a clone's trace, not in git, so the act and time of a Result that
// caused the stop are not printed: every clone prints a landed Goal the same.
func causeOf(g *Goal, id string) Cause {
	if g.Select != nil && g.Select.ID == id && !g.Select.Calculated {
		return Cause{Act: spec.ActSelect, ID: id, Time: g.Select.Time}
	}
	return Cause{ID: id}
}

// budgetUsed reports the stop Budget used: three trees counted, the third
// with a live cell for every placed example. The caller checks no candidate.
func budgetUsed(g *Goal) bool {
	counted := g.CountedTrees()
	if len(counted) < MaxHypotheses {
		return false
	}
	third := counted[MaxHypotheses-1]
	for _, p := range g.Placed() {
		if g.Verdict(p.Key, third) == NotRun {
			return false
		}
	}
	return true
}

// LatestCommit is the commit of the latest Result, in merge order, that
// counts on the row with rowTree: not refused by Stable.
func LatestCommit(g *Goal, rowTree string) string {
	commit := ""
	for _, r := range g.Results {
		if r.RowTree == rowTree && r.Refused == "" {
			commit = r.Commit
		}
	}
	return commit
}

// ReadOf is what the latest Result on commit that counts, not
// refused by Stable, read of it; nil when no such Result
// carries a read.
func ReadOf(g *Goal, commit string) *events.CommitRead {
	var read *events.CommitRead
	for _, r := range g.Results {
		if r.Commit == commit && r.Refused == "" {
			read = r.Read
		}
	}
	return read
}

// Selectable checks the select guards that the act checks on git, Spec and
// Own path, on what the Result of the commit the select would record read
// of it. A candidate that fails them is not listed; HeldBack names the guard.
func Selectable(g *Goal, c Row) bool {
	return readGuards(g, c) == nil
}

// readGuards is the first select guard that the act checks on git and that
// refuses the candidate c, as hyp select would print it, on the commit the
// select would record: Spec, on what the Result of that commit carries. Own
// path held when that Result was recorded, since run checks it on the same
// commit and H0. It is nil when both hold.
func readGuards(g *Goal, c Row) error {
	commit := LatestCommit(g, c.RowTree)
	if commit == "" {
		return fault.Refuse(fault.GuardCandidate, fault.NoResultOnRow, c.RowTree)
	}
	return GuardSpec(g, commit, ReadOf(g, commit))
}

// GuardSpec is the Spec guard on what a run read of the commit: the commit
// holds a spec under goals/<G>/specs/. When the spec is a model (a .tla
// file), every rule its check configs check is named by a variant config
// under specs/variants/ (or the older specs/switches/). When the commit
// changes the kept model (goals/model/), every rule of it that is new, or
// whose text differs, against H0, whatever commit the clone has checked
// out, is named by such a variant config too; the other rules of the kept
// model need none. A model or a kept change needs a chart (.mmd) in the
// folder. hyp reads the files and runs no model checker. A nil read is a
// Result recorded before runs carried what they read: a run of a line on
// the commit records one.
func GuardSpec(g *Goal, commit string, read *events.CommitRead) error {
	dir := SpecDir(g.ID)
	if read == nil {
		return fault.Refuse(fault.GuardSpec, fault.NoRead, commit, g.ID)
	}
	if read.SpecFiles == 0 {
		return fault.Refuse(fault.GuardSpec, fault.NoSpec, commit, dir)
	}
	var kept []string
	for _, it := range read.Kept {
		if it.Rule {
			kept = append(kept, it.Name)
		}
	}
	if !read.Model && len(kept) == 0 {
		return nil
	}
	var missing []string
	if read.Model {
		missing = append(missing, read.Unnamed...)
	}
	for _, r := range kept {
		if !slices.Contains(read.Named, r) && !slices.Contains(missing, r) {
			missing = append(missing, r)
		}
	}
	if len(missing) > 0 {
		return fault.Refuse(fault.GuardSpec, fault.RuleNoVariant, commit, missing, dir+VariantsDir)
	}
	if len(read.Charts) == 0 {
		return fault.Refuse(fault.GuardSpec, fault.NoChart, commit, dir)
	}
	return nil
}

// Hold is a candidate whose select hyp holds back: its row and the refusal
// hyp select would print for it.
type Hold struct {
	Row Row
	Err error
}

// HeldBack lists the candidates of an open Goal whose select hyp holds back,
// in the order of the candidates: each candidate of a Goal at no stop with
// one candidate, or at Two candidates, that fails a select guard that reads
// git in the act. A selected or ended Goal holds none back.
func HeldBack(g *Goal) []Hold {
	if g.State == StateEnded || g.Select != nil {
		return nil
	}
	st := Stand(g)
	if !(st.Stop == StopTwoCandidates || (st.Stop == StopNone && len(st.Candidates) == 1)) {
		return nil
	}
	var out []Hold
	for _, c := range st.Candidates {
		if err := readGuards(g, c); err != nil {
			out = append(out, Hold{Row: c, Err: err})
		}
	}
	return out
}
