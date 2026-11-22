package acts

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/example"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/probe"
	"github.com/dat9uy/hyp/internal/tables"
)

// TryOpts is the input of try.
type TryOpts struct {
	Dir          string // where the command runs
	Goal         string // the Goal's id
	Key          string // LINE/EXAMPLE, for example line-3/a
	Worktree     bool   // run on the checkout's files, not on H0
	Commit       string // run on this commit, not on H0
	Actor        string // whose trace file a saved cell goes to
	Role         string // the role the command declared, or empty
	ProbeTimeout time.Duration
}

// TryReport is what try returns: the example as it is on disk, what its
// Probe printed, and the verdict hyp would record for it.
type TryReport struct {
	Goal      string // the Goal's event id
	ID        string // the Goal's short id
	Key       string
	Folder    string // goals/<G>/<line>/<example>
	Hash      string // the folder's tree hash on disk
	ProbeHash string // the tree hash of its probe/ folder on disk
	Words     string
	Commit    string // H0's commit, the commit named, or HEAD's with Worktree
	Worktree  bool
	OnCommit  bool   // the commit was named with --commit
	RowTree   string // the row tree the example ran on
	Saved     bool   // the cell was missing from the trace, and try saved it
	Out       probe.Out
	OutHash   string // the name the output would have under .tree/outputs
	ErrHash   string // the name the stderr would have under .tree/outputs
	Verdict   string
}

// Try runs one example's Probe and its compare as place and run do. It reads
// the example folder as it is on disk, committed or not, and runs it on H0's
// commit, on the commit named, or with Worktree on the files of this checkout
// as a commit of every change would hold them. It is a read: it checks no
// guard, so it runs in every state of the Goal, and it writes nothing to git
// and uses no run of the budget. It records nothing either, except one case:
// when the folder on disk is the current placement and the clone's trace
// lacks the cell that hyp status prints for it (H0's cell, or the selected
// row's), try saves the cell it measured in the trace, marked as run again
// in this clone. An unknown Goal, line or folder is a bad argument; a Probe
// that fails is an environment problem, as in place and run.
func Try(o TryOpts) (TryReport, error) {
	if o.Worktree && o.Commit != "" {
		return TryReport{}, fault.BadArg(fault.TryTwoSubjects)
	}
	lineKey, name, ok := strings.Cut(o.Key, "/")
	if !ok || !lineKeyRe.MatchString(lineKey) || !exampleNameRe.MatchString(name) {
		return TryReport{}, fault.BadArg(fault.BadExampleKey, o.Key)
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return TryReport{}, err
	}
	if err := logstore.Find(root); err != nil {
		return TryReport{}, err
	}
	evs, _, err := logstore.All(root)
	if err != nil {
		return TryReport{}, err
	}
	g, err := tables.Replay(evs).Find(o.Goal)
	if err != nil {
		return TryReport{}, err
	}
	found := false
	for _, l := range g.Lines {
		found = found || l.Key == lineKey
	}
	if !found {
		return TryReport{}, fault.BadArg(fault.NoLine, g.ID, lineKey)
	}

	rel := path.Join("goals", g.ID, lineKey, name)
	disk := filepath.Join(root, filepath.FromSlash(rel))
	if fi, err := os.Stat(disk); err != nil || !fi.IsDir() {
		return TryReport{}, fault.BadArg(fault.NotAFolder, rel)
	}
	data, err := os.ReadFile(filepath.Join(disk, "example.txtar"))
	if err != nil {
		return TryReport{}, fault.BadArg(fault.NoExampleTxtar, rel)
	}
	ex, err := example.Parse(data)
	if err != nil {
		return TryReport{}, err
	}
	rep := TryReport{Goal: g.EventID, ID: g.ID, Key: o.Key, Folder: rel, Words: ex.Words, Commit: g.Base, Worktree: o.Worktree}
	if rep.Hash, err = gitread.DirTreeHash(disk); err != nil {
		return TryReport{}, err
	}
	if fi, err := os.Stat(filepath.Join(disk, "probe")); err == nil && fi.IsDir() {
		if rep.ProbeHash, err = gitread.DirTreeHash(filepath.Join(disk, "probe")); err != nil {
			return TryReport{}, err
		}
	}
	rep.RowTree = g.BaseTree
	switch {
	case o.Worktree:
		if rep.Commit, err = gitread.CommitOf(root, "HEAD"); err != nil {
			return TryReport{}, err
		}
	case o.Commit != "":
		if strings.HasPrefix(o.Commit, "-") {
			return TryReport{}, fault.BadArg(fault.BadCommit, o.Commit)
		}
		if ok, err := gitread.HasCommit(root, o.Commit); err != nil {
			return TryReport{}, err
		} else if !ok {
			return TryReport{}, fault.BadArg(fault.BadCommit, o.Commit)
		}
		if rep.Commit, err = gitread.CommitOf(root, o.Commit); err != nil {
			return TryReport{}, err
		}
		if rep.RowTree, err = gitread.RowTree(root, rep.Commit); err != nil {
			return TryReport{}, err
		}
		rep.OnCommit = true
	}

	out, err := probe.Run(probe.Opts{
		Repo: root, Commit: rep.Commit, ExampleDir: disk, Worktree: o.Worktree, Timeout: o.ProbeTimeout,
	})
	if err != nil {
		return TryReport{}, err
	}
	if o.Worktree {
		rep.RowTree = out.RowTree
	}
	rep.Out = out
	rep.OutHash, rep.ErrHash = logstore.Hash(out.Stdout), logstore.Hash(out.Stderr)
	rep.Verdict = string(example.Compare(out.Stdout, ex))
	if missingCell(g, o.Key, rep.Hash, rep.RowTree) {
		cell := events.SavedCell{
			Line: lineKey, Example: name, Hash: rep.Hash, Probe: rep.ProbeHash,
			Output: rep.OutHash, Stderr: rep.ErrHash, Exit: out.Exit, Verdict: rep.Verdict,
		}
		p := events.CellsSave{Role: o.Role, Commit: rep.Commit, RowTree: rep.RowTree, Again: true, Cells: []events.SavedCell{cell}}
		if err := saveCells(root, o.Actor, g.EventID, p, [][]byte{out.Stdout, out.Stderr}); err != nil {
			return TryReport{}, err
		}
		rep.Saved = true
	}
	return rep, nil
}

// MissingCell reports whether hyp status prints the cell of key on the row
// with rowTree as not in this clone's trace: the row is H0's or the selected
// row, key is placed, and this clone has no live cell of it there.
func MissingCell(g *tables.Goal, key, rowTree string) bool {
	p, ok := g.Current(key)
	if !ok {
		return false
	}
	return missingCell(g, key, p.Hash, rowTree)
}

// missingCell is MissingCell for the example folder with hash: only the
// current placement's folder fills its cell.
func missingCell(g *tables.Goal, key, hash, rowTree string) bool {
	p, ok := g.Current(key)
	if !ok || p.Hash != hash {
		return false
	}
	if rowTree != g.BaseTree && (g.Select == nil || g.Select.RowTree != rowTree) {
		return false
	}
	for _, c := range g.LiveCells(rowTree) {
		if c.Key == key {
			return false
		}
	}
	return true
}
