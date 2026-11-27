package acts

import (
	"slices"

	"github.com/dat9uy/hyp/internal/fault"
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

// Standing is what hyp calculates about a Goal from the tables alone: what
// each run read of git is in its Result.
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
// its land names; an ended Goal waits for no one. It reads no git.
func Stand(g *tables.Goal) Standing {
	st := Standing{Stop: StopNone, Waits: "none", Used: tables.RunsUsed(g), Rows: g.Rows()}
	st.Candidates = tables.Candidates(g)
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
	case g.State == tables.StateWritten:
		st.Waits = "owner"
	}
	return st
}

// selectCause is the act that caused the Selected stop: a select event of
// the older form, or for a calculated select the run whose Result made its
// row met on every line.
func selectCause(sel *tables.Select) Cause {
	if sel.Calculated {
		return Cause{Act: ActRun, ID: sel.ID, Time: sel.Time}
	}
	return Cause{Act: ActSelect, ID: sel.ID, Time: sel.Time}
}

// lastRun is the run that caused a stop of the runs: the latest Result, in
// merge order, that counts, not refused by Stable. Only a
// run adds a row or a cell, so the latest one is the run after which the stop
// held.
func lastRun(g *tables.Goal) Cause {
	for i := len(g.Results) - 1; i >= 0; i-- {
		if r := g.Results[i]; r.Refused == "" {
			return Cause{Act: ActRun, ID: r.ID, Time: r.Time}
		}
	}
	return Cause{}
}

// causeOf is the cause a land event names, as git alone holds it: the
// Goal's select event of the older form, or else the id alone. A Result is
// in a clone's trace, not in git, so the act and time of a Result that
// caused the stop are not printed: every clone prints a landed Goal the same.
func causeOf(g *tables.Goal, id string) Cause {
	if g.Select != nil && g.Select.ID == id && !g.Select.Calculated {
		return Cause{Act: ActSelect, ID: id, Time: g.Select.Time}
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
func Allowed(t tables.Tables) []Ready {
	var out []Ready
	for _, g := range tables.SortedGoals(t) {
		out = append(out, AllowedFor(g)...)
	}
	return out
}

// AllowedFor lists the allowed acts of one Goal: only acts its state's entry
// in the spec allows (spec.States), and of those only the ones whose guards
// that read only the tables hold; Kind and Stable need a Probe run
// and are checked when the act runs. Acts the binary lacks are never listed.
// Within a role the lines keep the order of the Goal's lines.
func AllowedFor(g *tables.Goal) []Ready {
	return slices.DeleteFunc(allowedFor(g), func(r Ready) bool { return !spec.Allows(g.State, r.Act) })
}

func allowedFor(g *tables.Goal) []Ready {
	if g.State == tables.StateEnded {
		return nil
	}
	if g.Select != nil {
		// a selected Goal allows only the owner's land and unselect of the
		// selected commit, and close; a select this clone's Results show,
		// with no select event, also allows the owner's place
		c := g.Select.Commit
		out := []Ready{{Goal: g.ID, Role: "owner", Act: ActLand, Commit: c}, {Goal: g.ID, Role: "owner", Act: ActUnselect, Commit: c}}
		if g.Select.Calculated {
			out = append(out, Ready{Goal: g.ID, Role: "owner", Act: ActPlace})
		}
		return append(out, Ready{Goal: g.ID, Role: "owner", Act: ActClose})
	}
	st := Stand(g)
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
			if selectable(g, c) {
				add(Ready{Role: "owner", Act: ActSelect, Tree: c.RowTree})
			}
		}
	}
	if st.Stop == StopTwoCandidates || st.Stop == StopBudgetUsed {
		add(Ready{Role: "owner", Act: ActLand})
	}
	// one place line, for the Goal's whole example set: a place is the
	// owner's act. A selected Goal returned above.
	add(Ready{Role: "owner", Act: ActPlace})
	add(Ready{Role: "owner", Act: ActClose})

	// runner: runs only; this clone's Results show the select
	if st.Stop == StopNone {
		// A lone candidate that fails Spec or Own path is fixed by a commit
		// that changes only goals/, so its subject, and its Hypothesis, stay.
		fix := len(st.Candidates) == 1 && !selectable(g, st.Candidates[0])
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
			if !slices.Contains(counted, row.RowTree) || !row.Leaf {
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

// selectable checks the select guards that the act checks on git, Spec and
// Own path, on what the Result of the commit the select would record read
// of it. A candidate that fails them is not listed; HeldBack names the guard.
func selectable(g *tables.Goal, c tables.Row) bool {
	return gitGuards(g, c) == nil
}

// gitGuards is the first select guard that the act checks on git and that
// refuses the candidate c, as hyp select would print it, on the commit the
// select would record: Spec, on what the Result of that commit carries. Own
// path held when that Result was recorded, since run checks it on the same
// commit and H0. It is nil when both hold.
func gitGuards(g *tables.Goal, c tables.Row) error {
	commit := latestCommit(g, c.RowTree)
	if commit == "" {
		return fault.Refuse(fault.GuardCandidate, fault.NoResultOnRow, c.RowTree)
	}
	return SpecGuard(g, commit, ReadOf(g, commit))
}

// Hold is a candidate whose select hyp holds back: its row and the refusal
// hyp select would print for it.
type Hold struct {
	Row tables.Row
	Err error
}

// HeldBack lists the candidates of an open Goal whose select hyp holds back,
// in the order of the candidates: each candidate of a Goal at no stop with
// one candidate, or at Two candidates, that fails a select guard that reads
// git in the act. A selected or ended Goal holds none back.
func HeldBack(g *tables.Goal) []Hold {
	if g.State == tables.StateEnded || g.Select != nil {
		return nil
	}
	st := Stand(g)
	if !(st.Stop == StopTwoCandidates || (st.Stop == StopNone && len(st.Candidates) == 1)) {
		return nil
	}
	var out []Hold
	for _, c := range st.Candidates {
		if err := gitGuards(g, c); err != nil {
			out = append(out, Hold{Row: c, Err: err})
		}
	}
	return out
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
