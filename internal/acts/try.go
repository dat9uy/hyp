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
	Commit       string // run on this commit, not on H0; empty for H0
	Actor        string // the actor a saved cell names in the trace
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
	H0        string // H0's commit
	Commit    string // H0's commit, the commit of --commit, or HEAD's with Worktree
	RowTree   string // the subject hash of Commit; empty with Worktree
	Worktree  bool
	Saved     bool // the cell was missing from the clone's trace, and try saved it there
	Out       probe.Out
	OutHash   string // the name the output would have under .tree/outputs
	ErrHash   string // the name the stderr would have under .tree/outputs
	Verdict   string
}

// Try runs one example's Probe and its compare as place and run do, and
// records no event in git and no run of the budget. It reads the example
// folder as it is on disk, committed or not, and runs it on H0's commit, on
// Commit, or with Worktree on the files of this checkout as a commit of every
// change would hold them. It is a read: it checks no guard, so it runs in
// every state of the Goal. When the folder on disk is the example's current
// placement and the clone's trace keeps no cell of it on the commit's row,
// try saves the cell there, marked as run again in this clone, so hyp status
// prints it: the trace is a cache that a reader fills by running an example
// again. A try with Worktree saves nothing. An unknown Goal, line or folder
// is a bad argument; a Probe that fails is an environment problem, as in
// place and run.
func Try(o TryOpts) (TryReport, error) {
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
	evs, _, err := logstore.Events(root)
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
	rep := TryReport{Goal: g.EventID, ID: g.ID, Key: o.Key, Folder: rel, Words: ex.Words, H0: g.Base, Commit: g.Base, Worktree: o.Worktree}
	if rep.Hash, err = gitread.DirTreeHash(disk); err != nil {
		return TryReport{}, err
	}
	if fi, err := os.Stat(filepath.Join(disk, "probe")); err == nil && fi.IsDir() {
		if rep.ProbeHash, err = gitread.DirTreeHash(filepath.Join(disk, "probe")); err != nil {
			return TryReport{}, err
		}
	}
	switch {
	case o.Worktree && o.Commit != "":
		return TryReport{}, fault.BadArg(fault.TryWorktreeCommit)
	case o.Worktree:
		if rep.Commit, err = gitread.CommitOf(root, "HEAD"); err != nil {
			return TryReport{}, err
		}
	case o.Commit != "":
		if ok, err := gitread.HasCommit(root, o.Commit); err != nil || !ok {
			return TryReport{}, fault.BadArg(fault.TryNoCommit, o.Commit)
		}
		if rep.Commit, err = gitread.CommitOf(root, o.Commit); err != nil {
			return TryReport{}, err
		}
	}
	if !o.Worktree {
		if rep.RowTree, err = gitread.RowTree(root, rep.Commit); err != nil {
			return TryReport{}, err
		}
	}

	out, err := probe.Run(probe.Opts{
		Repo: root, Commit: rep.Commit, ExampleDir: disk, Worktree: o.Worktree, Timeout: o.ProbeTimeout,
	})
	if err != nil {
		return TryReport{}, err
	}
	rep.Out = out
	rep.OutHash, rep.ErrHash = logstore.Hash(out.Stdout), logstore.Hash(out.Stderr)
	rep.Verdict = string(example.Compare(out.Stdout, ex))
	if o.Worktree || o.Actor == "" {
		return rep, nil
	}
	cur, ok := g.Current(o.Key)
	if !ok || cur.Hash != rep.Hash || CellInTrace(root, g, o.Key, cur.Hash, rep.RowTree) {
		return rep, nil
	}
	if err := putOutputs(root, out.Stdout, out.Stderr); err != nil {
		return TryReport{}, err
	}
	cache := events.CellsCache{Role: "", Again: true, Cells: []events.CachedCell{{
		Key: o.Key, Hash: cur.Hash, Commit: rep.Commit, RowTree: rep.RowTree, Probe: rep.ProbeHash,
		Output: rep.OutHash, Stderr: rep.ErrHash, Exit: out.Exit, Verdict: rep.Verdict,
	}}}
	if err := saveCells(root, o.Actor, g.EventID, cache); err != nil {
		return TryReport{}, err
	}
	rep.Saved = true
	return rep, nil
}

// CellInTrace reports whether this clone holds the outputs of the cell of
// example key at hash on the row with rowTree: a live cell that names its
// output, from a Result or from a place written before the trace kept
// outputs, or a cell its trace keeps (tables.Cached).
func CellInTrace(root string, g *tables.Goal, key, hash, rowTree string) bool {
	_, ok := CellOutputs(root, g, key, hash, rowTree)
	return ok
}

// ShownCell is a cell hyp status prints with its outputs: its verdict, exit
// code and output hashes, and whether a try ran it again in this clone.
type ShownCell struct {
	Verdict, Output, Stderr string
	Exit                    int
	Again                   bool
}

// CellOutputs returns the cell of example key at hash on the row with
// rowTree whose outputs this clone holds: first the first live cell, when it
// names its output, then a cell its trace keeps. It returns false when the
// clone holds none: the cell is not in this clone's trace. A live cell of a
// place names no output: the trace of the clone that placed keeps it.
func CellOutputs(root string, g *tables.Goal, key, hash, rowTree string) (ShownCell, bool) {
	for _, c := range g.LiveCells(rowTree) {
		if c.Key == key && c.Hash == hash {
			if c.Output == "" {
				break
			}
			return ShownCell{Verdict: c.Verdict, Output: c.Output, Stderr: c.Stderr, Exit: c.Exit}, true
		}
	}
	if c, ok := g.CachedCell(key, hash, rowTree); ok {
		return ShownCell{Verdict: c.Verdict, Output: c.Output, Stderr: c.Stderr, Exit: c.Exit, Again: c.Again}, true
	}
	return ShownCell{}, false
}
