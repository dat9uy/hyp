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

// Ancestry answers the two questions about git that the run guards ask.
// gitread.Ancestry is the real one. Rows read none: the run's Result
// carries what it read.
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
	Verdicts map[string]string // by example key: the live cell's verdict, or NotRun
}

// rowCommits lists the Goal's rows in order, H0's first, as row tree and the
// commits it was seen at. A Result refused by Stable counts for nothing, so
// it makes no row.
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

// Rows calculates the rows of every Goal, H0's row first. A row is a leaf
// when no Result on another row names one of its commits among the commits
// its run's commit descends from (CommitRead.Above): the run read git, and
// its Result carries what it read, so no row reads git.
func Rows(t Tables) []Row {
	var out []Row
	for _, g := range sortedGoals(t) {
		out = append(out, g.Rows()...)
	}
	return out
}

// Rows calculates the rows of one Goal, H0's row first.
func (g *Goal) Rows() []Row {
	var out []Row
	trees, commits := g.rowCommits()
	for _, tree := range trees {
		row := Row{Goal: g.EventID, RowTree: tree, Commits: commits[tree], Verdicts: map[string]string{}}
		for _, p := range g.Placed() {
			row.Verdicts[p.Key] = g.Verdict(p.Key, tree)
		}
		row.Leaf = g.leaf(tree, commits[tree])
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

// leaf reports whether the row with tree, seen at commits, is a leaf: no
// Result that makes a row on another tree carries one of commits in Above.
// Leaf only refuses a run on an ancestor of a row, so a row recorded later
// never lies below one recorded before it, and the Result of the lower row
// is always the earlier one; the later Result names it.
//
// A move above a select, a run on a merge of the selected commit, makes
// the rows below it no leaf only once its row is met on every line: until
// then the Goal stays selected on the commit it moves from.
func (g *Goal) leaf(tree string, commits []string) bool {
	for _, r := range g.Results {
		if r.Refused != "" || r.RowTree == tree || r.Read == nil {
			continue
		}
		if r.Move != "" && !g.metEverywhere(r.RowTree) {
			continue
		}
		for _, c := range r.Read.Above {
			if slices.Contains(commits, c) {
				return false
			}
		}
	}
	return true
}

// Above is each commit of g's rows, H0's included, other than commit, that
// commit descends from, as anc answers: what a run carries in its Result.
func (g *Goal) Above(anc Ancestry, commit string) ([]string, error) {
	var out []string
	trees, commits := g.rowCommits()
	for _, tree := range trees {
		for _, c := range commits[tree] {
			if c == commit || slices.Contains(out, c) {
				continue
			}
			below, err := anc.IsAncestor(c, commit)
			if err != nil {
				return nil, fault.EnvFault(fault.LeafNeedsCommit, g.ID, missingOf(anc, c, commit), err)
			}
			if below {
				out = append(out, c)
			}
		}
	}
	return out, nil
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

// GoalTree is what Own path reads of a run's commit: the tree hash at a
// path, and who wrote the bytes at a path.
type GoalTree interface {
	TreeHash(path string) (string, error)
	Writer(path string) (gitread.Writer, error)
}

// GuardOwnPath is the Own path guard. It checks each Goal folder of the
// run's commit against the store t, where the commit's goals/ tree differs
// from H0's (changes): a file under goals/X/ of another Goal X holds bytes
// X's owner acts recorded, and no file is added or changed under a goals/X/
// folder of no Goal. An example folder of X, goals/X/<line>/<example>/,
// holds a tree whose hash an add or a place of X recorded for that example;
// any other file of X's folder, such as its spec, was written by a commit
// that adds the file of an event of X the store holds, as the owner's
// record of an act on X does, or by a commit H0's history holds. So it reads what the runner wrote, whatever
// branch brought the owner's acts on other Goals. goals/model/, the
// repository's kept model, is no Goal's folder: a Goal's model is a change
// to it. A file the commit lacks is no write.
func GuardOwnPath(g *Goal, t Tables, tree GoalTree, changes []gitread.OwnPath) error {
	for _, c := range changes {
		rest, ok := strings.CutPrefix(c.Path, "goals/")
		if !ok || c.Deleted {
			continue
		}
		name, inner, _ := strings.Cut(rest, "/")
		if name == g.ID || name == "model" {
			continue
		}
		other := t.ByID(name)
		if other != nil {
			ok, err := recordedPath(other, tree, c.Path, inner)
			if err != nil {
				return err
			}
			if ok {
				continue
			}
		}
		w, err := tree.Writer(c.Path)
		if err != nil {
			return err
		}
		if other != nil {
			return fault.Refuse(fault.GuardOwnPath, fault.OtherGoalPath, w.Commit, c.Path, g.ID)
		}
		return fault.Refuse(fault.GuardOwnPath, fault.NoGoalFolder, w.Commit, c.Path)
	}
	return nil
}

// recordedPath reports whether the file at path, inner within the folder
// of Goal x, holds bytes x's owner acts recorded.
func recordedPath(x *Goal, tree GoalTree, path, inner string) (bool, error) {
	parts := strings.SplitN(inner, "/", 3)
	if len(parts) == 3 {
		if _, isLine := x.line(parts[0]); isLine {
			key := parts[0] + "/" + parts[1]
			h, err := tree.TreeHash("goals/" + x.ID + "/" + key)
			if err != nil {
				return false, err
			}
			for _, p := range x.Placements[key] {
				if h != "" && p.Hash == h {
					return true, nil
				}
			}
			return false, nil
		}
	}
	w, err := tree.Writer(path)
	if err != nil || w.Old {
		return w.Old, err
	}
	for _, id := range w.Events {
		if slices.Contains(x.seen, id) {
			return true, nil
		}
	}
	return false, nil
}

// countedTrees are the distinct row trees, H0's excluded, of the Goal's
// Results recorded since the boundary. A Result refused
// by Stable does not count; one on a tree that was H0 when it ran never does, even
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

// GuardRun is the git-free check of a result.record event that the run act
// calls before it appends: not Ended, not Selected, Open, Budget and Stable.
// The act calls the same guards in spec order around the guards that read
// git. A Result already marked refused by Stable is kept and counts for
// nothing, so it passes Stable. ev.Node is the goal.add event id.
// p is the payload of ev.
func GuardRun(t Tables, ev events.Event, p *events.ResultRecord) error {
	g, err := guardRunGit(t, ev, p)
	if err != nil {
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

// guardRunAtReplay is what replay checks of a result.record: the run guards
// that read only events stored in git, Not ended, Selected and Open. Budget
// and Stable read the Results of the clone's trace, so replay does not check
// them; the act did.
// p is the payload of ev.
func guardRunAtReplay(t Tables, ev events.Event, p *events.ResultRecord) error {
	_, err := guardRunGit(t, ev, p)
	return err
}

// guardRunGit checks the run guards that read only events stored in git:
// Not ended, Selected and Open. p is the payload of ev.
func guardRunGit(t Tables, ev events.Event, p *events.ResultRecord) (*Goal, error) {
	g, ok := t.Goals[ev.Node]
	if !ok {
		return nil, fault.Refuse(fault.NoGuard, fault.NoGoalToRun, GoalID(ev.Node))
	}
	if err := GuardNotEnded(g); err != nil {
		return nil, err
	}
	if IsMove(g, p.Move) {
		return g, nil
	}
	if err := GuardNotSelected(g); err != nil {
		return nil, err
	}
	if g.openAtReplay() {
		return g, nil
	}
	if err := GuardOpen(g, p.RowTree, p.Line); err != nil {
		return nil, err
	}
	return g, nil
}

// RefusedStable and RefusedSameBase are the marks of a Result its run
// recorded although a guard refused it: it counts for nothing.
const (
	RefusedStable   = "Stable"
	RefusedSameBase = "Same base"
)

// RowBases maps each commit of g's rows, H0's excluded, to the base its
// latest Result recorded, "" for H0: the walk of a later run's base goes
// through them (gitread.Bases).
func (g *Goal) RowBases() map[string]string {
	out := map[string]string{}
	for _, r := range g.Results {
		if r.Refused != "" || r.Commit == g.Base {
			continue
		}
		out[r.Commit] = ""
		if r.Base != nil {
			out[r.Commit] = r.Base.Commit
		}
	}
	return out
}

// BaseOf is the base the latest Result on commit that counts recorded, and
// whether such a Result exists; the base is nil when the commit stood on H0.
func (g *Goal) BaseOf(commit string) (*events.BaseRead, bool) {
	var base *events.BaseRead
	found := false
	for _, r := range g.Results {
		if r.Commit == commit && r.Refused == "" {
			base, found = r.Base, true
		}
	}
	return base, found
}

// RefusedMoves are the Results of g refused by Same base, in merge order:
// the moves onto a newer base that wait for the owner.
func (g *Goal) RefusedMoves() []*Result {
	var out []*Result
	for _, r := range g.Results {
		if r.Refused == RefusedSameBase {
			out = append(out, r)
		}
	}
	return out
}
