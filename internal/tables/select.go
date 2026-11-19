package tables

import (
	"strings"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/herr"
)

// metEverywhere reports whether the row with rowTree is met on every line:
// every line has a placed example and every current example is met there.
func (g *Goal) metEverywhere(rowTree string) bool {
	placed := g.Placed()
	if len(placed) == 0 {
		return false
	}
	lines := map[string]bool{}
	for _, p := range placed {
		if g.Verdict(p.Key, rowTree) != "met" {
			return false
		}
		lines[p.Line] = true
	}
	for _, l := range g.Lines {
		if !lines[l.Key] {
			return false
		}
	}
	return true
}

// Candidates are the rows of g that a select may name: leaf rows met on every
// line under the current placements. A row whose leaf mark is unknown, because
// this clone lacks a commit, is left out; GuardCandidate names that commit.
func Candidates(g *Goal, anc Ancestry) []Row {
	var out []Row
	for _, row := range g.Rows(anc) {
		if row.Leaf && row.Unknown == "" && g.metEverywhere(row.RowTree) {
			out = append(out, row)
		}
	}
	return out
}

// selectRoles are the roles that may select.
func selectRole(role string) bool { return role == "owner" || role == "runner" }

// GuardCandidate is the Candidate guard of the select act: the tree is a
// candidate, and with two or more candidates only the owner selects. tree is
// a row tree hash or a unique prefix of one. A row whose leaf mark needs a
// commit this clone lacks is an environment problem naming it.
func GuardCandidate(g *Goal, anc Ancestry, tree, role string) (Row, error) {
	if !selectRole(role) {
		return Row{}, herr.Refused("Candidate: the role is %q; only the runner or the owner selects", role)
	}
	if tree == "" {
		return Row{}, herr.BadArgs("select needs a row tree")
	}
	var named []Row
	for _, row := range g.Rows(anc) {
		if row.RowTree == tree {
			named = []Row{row}
			break
		}
		if len(tree) >= 7 && strings.HasPrefix(row.RowTree, tree) {
			named = append(named, row)
		}
	}
	if len(named) > 1 {
		return Row{}, herr.BadArgs("%s matches %d rows of %s; give more of the tree", tree, len(named), g.ID)
	}
	if len(named) == 0 {
		return Row{}, herr.Refused("Candidate: %s has no row %s; a candidate is a leaf row met on every line", g.ID, tree)
	}
	row := named[0]
	if row.Unknown != "" {
		return Row{}, herr.Env("Candidate: the leaf mark of row %s needs commit %s, which this clone lacks", row.RowTree, row.Unknown)
	}
	cands := Candidates(g, anc)
	for _, c := range cands {
		if c.RowTree != row.RowTree {
			continue
		}
		if len(cands) >= 2 && role != "owner" {
			return Row{}, herr.Refused("Candidate: %s has %d candidates; only the owner selects among them", g.ID, len(cands))
		}
		return row, nil
	}
	if !row.Leaf {
		return Row{}, herr.Refused("Candidate: row %s is not a leaf; a commit of another row descends from it", row.RowTree)
	}
	return Row{}, herr.Refused("Candidate: row %s is not met on every line", row.RowTree)
}

// GuardGoalActor is the Goal's actor guard: only the actor that added the
// Goal selects for the owner. Other roles pass.
func GuardGoalActor(g *Goal, actor, role string) error {
	if role == "owner" && actor != g.Actor {
		return herr.Refused("Goal's actor: %s was added by %s, not %s; only that actor selects for the owner", g.ID, g.Actor, actor)
	}
	return nil
}

// GuardSelect is the git-free check of a select event, called by Replay at the
// event's place in the merge order: not Selected, the table half of Candidate
// (the row is met on every line), and Goal's actor. The leaf mark reads git
// and is trusted as written, as Leaf only is. ev.Node is the goal.add event
// id.
func GuardSelect(t Tables, ev events.Event) error {
	var p events.Select
	if err := events.Decode(ev, &p); err != nil {
		return herr.Refused("Selected: %v", err)
	}
	g, ok := t.Goals[ev.Node]
	if !ok {
		return herr.Refused("no Goal %s to select on", GoalID(ev.Node))
	}
	if err := GuardNotSelected(g); err != nil {
		return err
	}
	if !selectRole(p.Role) {
		return herr.Refused("Candidate: the role is %q; only the runner or the owner selects", p.Role)
	}
	if !g.metEverywhere(p.RowTree) {
		return herr.Refused("Candidate: row %s is not met on every line of %s", p.RowTree, g.ID)
	}
	return GuardGoalActor(g, ev.Actor, p.Role)
}

// UnstableCell is one example a Stable refusal found unstable: the tree, and
// the output of the live cell it differed from and of the refused run.
type UnstableCell struct {
	Key, Tree, First, Second string
}

// unstableAt reads, at the Result's place in the merge order, the live cell
// each differing example was compared with, so a later retract or adjust does
// not change what the refusal shows.
func (g *Goal) unstableAt(p events.ResultRecord) []UnstableCell {
	live := g.LiveCells(p.RowTree)
	var out []UnstableCell
	for _, key := range p.Differing {
		u := UnstableCell{Key: key, Tree: p.RowTree}
		for _, c := range p.Cells {
			if p.Line+"/"+c.Example != key {
				continue
			}
			u.Second = c.Output
			for _, l := range live {
				if l.Key == key && l.Hash == c.Hash {
					u.First = l.Output
					break
				}
			}
		}
		out = append(out, u)
	}
	return out
}

// SplitLines are the keys of the lines whose examples give different verdicts
// on the row with rowTree: at least one met and at least one failed. An
// example that is not run is no verdict.
func (g *Goal) SplitLines(rowTree string) []string {
	met, failed := map[string]bool{}, map[string]bool{}
	for _, p := range g.Placed() {
		switch g.Verdict(p.Key, rowTree) {
		case "met":
			met[p.Line] = true
		case "failed":
			failed[p.Line] = true
		}
	}
	var out []string
	for _, l := range g.Lines {
		if met[l.Key] && failed[l.Key] {
			out = append(out, l.Key)
		}
	}
	return out
}

// CountedTrees are the distinct row trees the budget counts, in the order
// they were first run.
func (g *Goal) CountedTrees() []string { return g.countedTrees() }

// SortedGoals lists the Goals in the order of their goal.add events.
func SortedGoals(t Tables) []*Goal { return sortedGoals(t) }
