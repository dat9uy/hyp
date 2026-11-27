package acts

import (
	"slices"

	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/spec"
	"github.com/dat9uy/hyp/internal/tables"
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
	st := tables.Stand(g)
	var out []Ready
	add := func(r Ready) {
		r.Goal = g.ID
		out = append(out, r)
	}

	// owner
	if g.State == tables.StateWritten {
		add(Ready{Role: "owner", Act: ActAddGoal})
	}
	if st.Stop == tables.StopTwoCandidates || (st.Stop == tables.StopNone && len(st.Candidates) == 1) {
		for _, c := range st.Candidates {
			if tables.Selectable(g, c) {
				add(Ready{Role: "owner", Act: ActSelect, Tree: c.RowTree})
			}
		}
	}
	if st.Stop == tables.StopTwoCandidates || st.Stop == tables.StopBudgetUsed {
		add(Ready{Role: "owner", Act: ActLand})
	}
	// one place line, for the Goal's whole example set: a place is the
	// owner's act. A selected Goal returned above.
	add(Ready{Role: "owner", Act: ActPlace})
	add(Ready{Role: "owner", Act: ActClose})

	// runner: runs only; this clone's Results show the select
	if st.Stop == tables.StopNone {
		// A lone candidate that fails Spec or Own path is fixed by a commit
		// that changes only goals/, so its subject, and its Hypothesis, stay.
		fix := len(st.Candidates) == 1 && !tables.Selectable(g, st.Candidates[0])
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
func runLines(g *tables.Goal, st tables.Standing, fix bool) []Ready {
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
