package acts

import (
	"github.com/dat9uy/hyp/internal/tables"
)

// Calculate gives each Goal the select this clone's Results show, and
// returns t. No event holds that select: git holds only the owner's acts, so
// a Goal is selected on a commit when the Results of this clone's trace show
// its row as the one candidate, met on every line, that no unselect
// rejected, that holds a spec and writes no other Goal's folder (Spec and
// Own path), and, after an unselect, that a run recorded a new cell for
// (New cell). A clone that lacks those Results shows the Goal open. A Goal
// whose select an event of the older form holds, and an ended Goal, keep
// what replay gave them.
func Calculate(t tables.Tables) tables.Tables {
	for _, g := range t.Goals {
		calculateSelect(g)
	}
	return t
}

func calculateSelect(g *tables.Goal) {
	if g.State != tables.StateOpen || g.Select != nil {
		return
	}
	st := Stand(g)
	if st.Stop != StopNone || len(st.Candidates) != 1 || !g.HasNewCell() {
		return
	}
	c := st.Candidates[0]
	if !selectable(g, c) {
		return
	}
	sel := &tables.Select{RowTree: c.RowTree, Calculated: true}
	for _, r := range g.Results {
		if r.RowTree == c.RowTree && r.Refused == "" {
			sel.ID, sel.Commit, sel.Time, sel.Actor = r.ID, r.Commit, r.Time, r.Actor
		}
	}
	g.Select = sel
	g.State = tables.StateSelected
}
