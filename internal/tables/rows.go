package tables

import (
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
)

// MaxHypotheses is the number of distinct Hypotheses a Goal may run since its
// boundary, H0 never counted. It is three for every Goal.
const MaxHypotheses = 3

// Ancestry answers the two questions about git that Rows and the run guards
// ask. gitread.Ancestry is the real one.
type Ancestry interface {
	// IsAncestor reports whether a is an ancestor of b, or b itself.
	IsAncestor(a, b string) (bool, error)
	// HasCommit reports whether this clone holds the commit.
	HasCommit(commit string) (bool, error)
}

// Row is a row of the Rows table: one subject tree of one Goal.
type Row struct {
	Goal     string   // the goal.add event id
	RowTree  string   // the subject tree hash
	Commits  []string // the commits it was seen at, in order
	Leaf     bool
	Unknown  string            // a commit this clone lacks: the leaf mark is unknown
	Verdicts map[string]string // by example key: the live cell's verdict, or NotRun
}

// rowCommits lists the Goal's rows in order, H0's first, as row tree and the
// commits it was seen at. A Result refused by Stable counts for nothing, so
// it makes no row; a retracted one still does.
func (g *Goal) rowCommits() (trees []string, commits map[string][]string) {
	commits = map[string][]string{g.BaseTree: {g.Base}}
	trees = []string{g.BaseTree}
	for _, r := range g.Results {
		if r.Refused != "" {
			continue
		}
		cs, seen := commits[r.RowTree]
		if !seen {
			trees = append(trees, r.RowTree)
		}
		if !slices.Contains(cs, r.Commit) {
			commits[r.RowTree] = append(cs, r.Commit)
		}
	}
	return trees, commits
}

// Rows calculates the rows of every Goal, H0's row first. A row is a leaf when
// none of its commits has a descendant among the commits of the Goal's other
// rows. A commit this clone lacks leaves the mark unknown and names the
// commit.
func Rows(t Tables, anc Ancestry) []Row {
	var out []Row
	for _, g := range sortedGoals(t) {
		out = append(out, g.Rows(anc)...)
	}
	return out
}

// Rows calculates the rows of one Goal, H0's row first.
func (g *Goal) Rows(anc Ancestry) []Row {
	var out []Row
	trees, commits := g.rowCommits()
	for _, tree := range trees {
		row := Row{Goal: g.EventID, RowTree: tree, Commits: commits[tree], Verdicts: map[string]string{}}
		for _, p := range g.Placed() {
			row.Verdicts[p.Key] = g.Verdict(p.Key, tree)
		}
		row.Leaf, row.Unknown = leafMark(anc, trees, commits, tree)
		out = append(out, row)
	}
	return out
}

func sortedGoals(t Tables) []*Goal {
	var gs []*Goal
	for _, g := range t.Goals {
		gs = append(gs, g)
	}
	slices.SortFunc(gs, func(a, b *Goal) int { return strings.Compare(a.EventID, b.EventID) })
	return gs
}

// leafMark reports whether the row with tree is a leaf, or the first commit
// this clone lacks.
func leafMark(anc Ancestry, trees []string, commits map[string][]string, tree string) (leaf bool, unknown string) {
	for _, c := range commits[tree] {
		for _, other := range trees {
			if other == tree {
				continue
			}
			for _, d := range commits[other] {
				if c == d {
					continue
				}
				desc, err := anc.IsAncestor(c, d)
				if err != nil {
					return false, missingOf(anc, c, d)
				}
				if desc {
					return false, ""
				}
			}
		}
	}
	return true, ""
}

// missingOf names the commit of a failed ancestry question that this clone
// lacks, a first, else b.
func missingOf(anc Ancestry, a, b string) string {
	if ok, err := anc.HasCommit(a); err != nil || !ok {
		return a
	}
	return b
}

// GuardOpen is the Open guard: the Goal is open, unless the commit's row tree
// is H0's. A line with no placed example has nothing to run, so it is refused
// by Open even on H0.
func GuardOpen(g *Goal, rowTree, line string) error {
	if rowTree == g.BaseTree {
		for _, p := range g.Placed() {
			if p.Line == line {
				return nil
			}
		}
		return fault.Refuse(fault.GuardOpen, fault.NothingPlaced, g.ID, line)
	}
	if g.State != StateOpen {
		return fault.Refuse(fault.GuardOpen, fault.NotOpen, g.ID, g.State)
	}
	return nil
}

// GuardBytes is the Bytes guard: the commit exists in this clone.
func GuardBytes(anc Ancestry, commit string) error {
	ok, err := anc.HasCommit(commit)
	if err != nil {
		return err
	}
	if !ok {
		return fault.Refuse(fault.GuardBytes, fault.CommitNotHere, commit)
	}
	return nil
}

// GuardLeafOnly is the Leaf only guard: the run is on H0, or on a commit that
// no row of the Goal descends from. A commit of a row that this clone lacks
// is an environment problem naming it.
func GuardLeafOnly(g *Goal, anc Ancestry, commit string) error {
	if commit == g.Base {
		return nil
	}
	trees, commits := g.rowCommits()
	for _, tree := range trees {
		for _, c := range commits[tree] {
			if c == commit {
				continue
			}
			below, err := anc.IsAncestor(commit, c)
			if err != nil {
				return fault.EnvFault(fault.LeafNeedsCommit, g.ID, missingOf(anc, commit, c), err)
			}
			if below {
				return fault.Refuse(fault.GuardLeafOnly, fault.AncestorOfRow, commit, c, g.ID)
			}
		}
	}
	return nil
}

// GuardOwnPath is the Own path guard: the runner's own change writes no path
// under goals/<another Goal>/.
func GuardOwnPath(g *Goal, changes []gitread.OwnPath) error {
	for _, c := range changes {
		rest, ok := strings.CutPrefix(c.Path, "goals/")
		if !ok {
			continue
		}
		if name, _, _ := strings.Cut(rest, "/"); name != g.ID {
			return fault.Refuse(fault.GuardOwnPath, fault.OtherGoalPath, c.Commit, c.Path, g.ID)
		}
	}
	return nil
}

// countedTrees are the distinct row trees, H0's excluded, of the Goal's
// Results recorded since the boundary. A retracted Result counts; one refused
// by Stable does not; one on a tree that was H0 when it ran never does, even
// after a Draft replace moved the base.
func (g *Goal) countedTrees() []string {
	var trees []string
	for _, r := range g.Results {
		if r.Refused != "" || r.epoch != g.epoch || r.wasH0 || r.RowTree == g.BaseTree || slices.Contains(trees, r.RowTree) {
			continue
		}
		trees = append(trees, r.RowTree)
	}
	return trees
}

// RunsUsed is N of "runs: N of 3 used".
func RunsUsed(g *Goal) int { return len(g.countedTrees()) }

// GuardBudget is the Budget guard: refused when three distinct trees have run
// since the boundary and the row tree is not among them. H0 never counts.
func GuardBudget(g *Goal, rowTree string) error {
	if rowTree == g.BaseTree {
		return nil
	}
	counted := g.countedTrees()
	if len(counted) >= MaxHypotheses && !slices.Contains(counted, rowTree) {
		return fault.Refuse(fault.GuardBudget, fault.BudgetUsed, g.ID, len(counted), MaxHypotheses)
	}
	return nil
}

// Unstable is an example whose new verdict differs from a live cell on the
// same row tree under the same example hash.
type Unstable struct {
	Key string
	Was LiveCell // the first live cell, whose verdict stands
	Now events.ExampleCell
}

// StableDiff compares the cells of a run of line on the row with rowTree with
// the live cells there. It returns one entry per example whose verdict
// differs.
func StableDiff(g *Goal, line, rowTree string, cells []events.ExampleCell) []Unstable {
	live := g.LiveCells(rowTree)
	var out []Unstable
	for _, c := range cells {
		key := line + "/" + c.Example
		for _, l := range live {
			if l.Key == key && l.Hash == c.Hash {
				if l.Verdict != c.Verdict {
					out = append(out, Unstable{Key: key, Was: l, Now: c})
				}
				break // the first live cell's verdict stands
			}
		}
	}
	return out
}

// GuardStable is the Stable guard: two runs on the same bytes give one
// verdict.
func GuardStable(g *Goal, line, rowTree string, cells []events.ExampleCell) error {
	diff := StableDiff(g, line, rowTree, cells)
	if len(diff) == 0 {
		return nil
	}
	keys := make([]string, len(diff))
	for i, d := range diff {
		keys[i] = d.Key
	}
	return fault.Refuse(fault.GuardStable, fault.OtherVerdict, g.ID, keys)
}

// GuardRunAtReplay is the check of a result.record event that Replay makes
// at the event's place in the merge order: Not ended, not Selected and Open,
// which read only events stored in git. Budget and Stable read the Results
// of the clone's trace, and the act alone checks them.
func GuardRunAtReplay(t Tables, ev events.Event) error {
	var p events.ResultRecord
	if err := events.Decode(ev, &p); err != nil {
		return fault.Refuse(fault.GuardStable, fault.PayloadInvalid, err)
	}
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToRun, GoalID(ev.Node))
	}
	if err := GuardNotEnded(g); err != nil {
		return err
	}
	if err := GuardNotSelected(g); err != nil {
		return err
	}
	return GuardOpen(g, p.RowTree, p.Line)
}

// GuardRun is the git-free check of a result.record event, called by the run
// act under the lock: not Selected, Open, Budget and Stable. The run act
// calls the same guards in spec order around the guards that read git. A Result already marked refused by Stable is kept and
// counts for nothing, so it passes Stable. ev.Node is the goal.add event id.
func GuardRun(t Tables, ev events.Event) error {
	var p events.ResultRecord
	if err := events.Decode(ev, &p); err != nil {
		return fault.Refuse(fault.GuardStable, fault.PayloadInvalid, err)
	}
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToRun, GoalID(ev.Node))
	}
	if err := GuardNotEnded(g); err != nil {
		return err
	}
	if err := GuardNotSelected(g); err != nil {
		return err
	}
	if err := GuardOpen(g, p.RowTree, p.Line); err != nil {
		return err
	}
	if err := GuardBudget(g, p.RowTree); err != nil {
		return err
	}
	if p.Refused != "" {
		return nil
	}
	return GuardStable(g, p.Line, p.RowTree, p.Cells)
}
