package acts

import (
	"slices"

	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/spec"
	"github.com/dat9uy/hyp/internal/tables"
)

// The stops of a Goal.
const (
	StopNone          = "none"
	StopSelected      = tables.StopSelected
	StopTwoCandidates = tables.StopTwoCandidates
	StopBudgetUsed    = tables.StopBudgetUsed
)

// The acts a ready line names: the acts of the spec's state entries.
const (
	ActAddGoal  = spec.ActAddGoal
	ActPlace    = spec.ActPlace
	ActRun      = spec.ActRun
	ActSelect   = spec.ActSelect
	ActLand     = spec.ActLand
	ActUnselect = spec.ActUnselect
	ActClose    = spec.ActClose
)

// Ready is one allowed act a role may run, with every key hyp knows. A key
// hyp cannot know is empty; the channel prints it as a placeholder.
type Ready struct {
	Goal   string // the Goal id
	Role   string
	Act    string // ActAddGoal, ActPlace, ActRun, ActSelect, ActLand, ActUnselect or ActClose
	Line   string // place and run: the line key
	Commit string // run: the commit, or empty for any leaf
	Tree   string // select: the candidate's row tree
}

// Cause is the event of the act that caused a stop: a select, or the
// result.record of the run after which the stop held.
type Cause struct {
	Act  string // ActSelect or ActRun
	ID   string // the event id
	Time string // the event's time
}

// Standing is what hyp calculates about a Goal from the tables and git.
type Standing struct {
	Stop       string
	Cause      Cause  // the act that caused the stop; empty for StopNone
	Waits      string // who the Goal waits for: grill, owner or none
	Used       int    // N of "runs: N of 3 used"
	Rows       []tables.Row
	Candidates []tables.Row
}

// Stand calculates the standing of a Goal: its rows and candidates, and the
// stop with the act that caused it. A landed Goal keeps the stop
// its land names; an ended Goal waits for no one. repo is the git top level.
func Stand(g *tables.Goal, repo string) Standing {
	anc := gitread.Ancestry{Dir: repo}
	st := Standing{Stop: StopNone, Waits: "none", Used: tables.RunsUsed(g), Rows: g.Rows(anc)}
	st.Candidates = tables.Candidates(g, anc)
	switch {
	case g.Land != nil:
		st.Stop, st.Cause = g.Land.Stop, causeOf(g, g.Land.Cause)
	case g.Close != nil && g.Select != nil:
		st.Stop = StopSelected
		st.Cause = Cause{Act: ActSelect, ID: g.Select.ID, Time: g.Select.Time}
	case g.Close != nil:
		// a closed Goal keeps the stop its rows hold, and waits for no one
		if len(st.Candidates) >= 2 {
			st.Stop, st.Cause = StopTwoCandidates, lastRun(g)
		} else if len(st.Candidates) == 0 && budgetUsed(g) {
			st.Stop, st.Cause = StopBudgetUsed, lastRun(g)
		}
	case g.Select != nil:
		st.Stop, st.Waits = StopSelected, "owner"
		st.Cause = Cause{Act: ActSelect, ID: g.Select.ID, Time: g.Select.Time}
	case len(st.Candidates) >= 2:
		st.Stop, st.Waits, st.Cause = StopTwoCandidates, "owner", lastRun(g)
	case len(st.Candidates) == 0 && budgetUsed(g):
		st.Stop, st.Waits, st.Cause = StopBudgetUsed, "owner", lastRun(g)
	case g.State == tables.StateWritten:
		st.Waits = "grill"
	}
	return st
}

// lastRun is the run that caused a stop of the runs: the latest Result, in
// merge order, that counts, neither retracted nor refused by Stable. Only a
// run adds a row or a cell, so the latest one is the run after which the stop
// held.
func lastRun(g *tables.Goal) Cause {
	for i := len(g.Results) - 1; i >= 0; i-- {
		if r := g.Results[i]; !r.Retracted && r.Refused == "" {
			return Cause{Act: ActRun, ID: r.ID, Time: r.Time}
		}
	}
	return Cause{}
}

// causeOf is the cause a land event names: the Goal's select, or one of its
// Results.
func causeOf(g *tables.Goal, id string) Cause {
	if g.Select != nil && g.Select.ID == id {
		return Cause{Act: ActSelect, ID: id, Time: g.Select.Time}
	}
	for _, r := range g.Results {
		if r.ID == id {
			return Cause{Act: ActRun, ID: id, Time: r.Time}
		}
	}
	return Cause{ID: id}
}

// budgetUsed reports the stop Budget used: three trees counted, the third
// with a live cell for every placed example. The caller checks no candidate.
func budgetUsed(g *tables.Goal) bool {
	counted := g.CountedTrees()
	if len(counted) < tables.MaxHypotheses {
		return false
	}
	third := counted[tables.MaxHypotheses-1]
	for _, p := range g.Placed() {
		if g.Verdict(p.Key, third) == tables.NotRun {
			return false
		}
	}
	return true
}

// Allowed lists the allowed acts of every Goal, in the order of the Goals'
// events.
func Allowed(t tables.Tables, repo string) []Ready {
	var out []Ready
	for _, g := range tables.SortedGoals(t) {
		out = append(out, AllowedFor(g, repo)...)
	}
	return out
}

// AllowedFor lists the allowed acts of one Goal: only acts its state's entry
// in the spec allows (spec.States), and of those only the ones whose guards
// that read only the tables and git hold; Kind and Stable need a Probe run
// and are checked when the act runs. Acts the binary lacks are never listed.
// Within a role the lines keep the order of the Goal's lines.
func AllowedFor(g *tables.Goal, repo string) []Ready {
	return slices.DeleteFunc(allowedFor(g, repo), func(r Ready) bool { return !spec.Allows(g.State, r.Act) })
}

func allowedFor(g *tables.Goal, repo string) []Ready {
	if g.State == tables.StateEnded {
		return nil
	}
	if g.Select != nil {
		// a selected Goal allows only the owner's land, unselect and close
		return []Ready{{Goal: g.ID, Role: "owner", Act: ActLand}, {Goal: g.ID, Role: "owner", Act: ActUnselect},
			{Goal: g.ID, Role: "owner", Act: ActClose}}
	}
	st := Stand(g, repo)
	var out []Ready
	add := func(r Ready) {
		r.Goal = g.ID
		out = append(out, r)
	}

	// owner
	if g.State == tables.StateWritten {
		add(Ready{Role: "owner", Act: ActAddGoal})
	}
	if st.Stop == StopTwoCandidates || (st.Stop == StopNone && len(st.Candidates) == 1) {
		for _, c := range st.Candidates {
			if selectable(g, repo, c) {
				add(Ready{Role: "owner", Act: ActSelect, Tree: c.RowTree})
			}
		}
	}
	if st.Stop == StopTwoCandidates || st.Stop == StopBudgetUsed {
		add(Ready{Role: "owner", Act: ActLand})
	}
	add(Ready{Role: "owner", Act: ActClose})

	// grill: one place line, for the Goal's whole example set. A selected
	// Goal returned above.
	add(Ready{Role: "grill", Act: ActPlace})

	// runner
	switch {
	case st.Stop == StopNone && len(st.Candidates) == 1 && g.HasNewCell() && selectable(g, repo, st.Candidates[0]):
		add(Ready{Role: "runner", Act: ActSelect, Tree: st.Candidates[0].RowTree})
	}
	if st.Stop == StopNone {
		// A lone candidate that fails Spec or Own path is fixed by a commit
		// that changes only goals/, so its subject, and its Hypothesis, stay.
		fix := len(st.Candidates) == 1 && !selectable(g, repo, st.Candidates[0])
		out = append(out, runLines(g, st, fix)...)
	}
	return orderRoles(out)
}

// roleOrder is the order roles print in: the order of the roles in hyp help.
var roleOrder = []string{"owner", "grill", "probe-writer", "runner"}

func orderRoles(ls []Ready) []Ready {
	slices.SortStableFunc(ls, func(a, b Ready) int {
		return slices.Index(roleOrder, a.Role) - slices.Index(roleOrder, b.Role)
	})
	return ls
}

// runLines are the runner's run lines. On a Goal that is written only a run on
// H0 is open, and only where an example's H0 cell is not run. On an open Goal
// a run on any leaf is open, so the commit is left empty; once three trees
// are counted, only the lines the counted leaf rows lack a cell for. With
// fix, the lone candidate cannot be selected on its commit: a run of any line
// on a new commit of the same Hypothesis costs no budget, so every line is
// open on a commit hyp cannot know yet.
func runLines(g *tables.Goal, st Standing, fix bool) []Ready {
	var out []Ready
	add := func(commit, line string) {
		out = append(out, Ready{Goal: g.ID, Role: "runner", Act: ActRun, Commit: commit, Line: line})
	}
	if g.State == tables.StateWritten {
		for _, l := range g.Lines {
			for _, p := range g.Placed() {
				if p.Line == l.Key && g.H0Verdict(p.Key) == tables.NotRun {
					add(g.Base, l.Key)
					break
				}
			}
		}
		return out
	}
	if g.State != tables.StateOpen {
		return nil
	}
	counted := g.CountedTrees()
	for _, l := range g.Lines {
		if !slices.ContainsFunc(g.Placed(), func(p tables.Placement) bool { return p.Line == l.Key }) {
			continue
		}
		if len(counted) < tables.MaxHypotheses || fix {
			add("", l.Key)
			continue
		}
		for _, row := range st.Rows {
			if !slices.Contains(counted, row.RowTree) || !row.Leaf || row.Unknown != "" {
				continue
			}
			for _, p := range g.Placed() {
				if p.Line == l.Key && row.Verdicts[p.Key] == tables.NotRun {
					add(row.Commits[len(row.Commits)-1], l.Key)
					break
				}
			}
		}
	}
	return out
}

// selectable checks the select guards that read only git: Spec and Own path
// on the commit the select would record. A candidate that fails them is not
// listed; the select act names the guard.
func selectable(g *tables.Goal, repo string, c tables.Row) bool {
	commit := latestCommit(g, c.RowTree)
	if commit == "" || guardSpec(repo, g, commit) != nil {
		return false
	}
	changes, err := gitread.OwnChange(repo, commit, g.Base)
	return err == nil && tables.GuardOwnPath(g, changes) == nil
}

// OnNoRef reports whether commit is in this clone and on no branch or tag:
// no ref under refs/heads, refs/remotes or refs/tags contains it. Another
// commit with the same subject hash on a branch does not count: git keeps
// the commit a select names only while a ref reaches it. A commit this clone
// lacks is not reported: this clone cannot tell where it is.
func OnNoRef(repo, commit string) bool {
	if ok, err := (gitread.Ancestry{Dir: repo}).HasCommit(commit); err != nil || !ok {
		return false
	}
	on, err := gitread.OnARef(repo, commit)
	return err == nil && !on
}
