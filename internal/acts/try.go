package acts

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

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
	ProbeTimeout time.Duration

	// Examples names a draft folder of examples, FOLDER/<line>/<example>,
	// for a Goal the owner has not added yet: Goal is empty, and Base is the
	// commit the Goal will name as H0.
	Examples string
	Base     string
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
	H0        bool   // it ran on H0
	Worktree  bool
	RowTree   string // the subject hash it ran on; empty with Worktree
	Out       probe.Out
	OutHash   string // the name the output would have under .tree/outputs
	ErrHash   string // the name the stderr would have under .tree/outputs
	Verdict   string
	Saved     bool // the cell was not in this clone's trace, and try saved it there
}

// Try runs one example's Probe and its compare as place and run do, and
// records no event and no run of the budget. It reads the example folder as
// it is on disk, committed or not, and runs it on H0's commit, on the commit
// named, or with Worktree on the files of this checkout as a commit of every
// change would hold them. When the folder is the example's current placed
// version and this clone's trace lacks its cell on that commit's row, it
// saves the cell there, marked as run again in this clone: the trace is a
// cache that a reader fills by running an example again. It is a read: it
// checks no guard, so it runs in every state of the Goal. An unknown Goal,
// line or folder is a bad argument; a Probe that fails is an environment
// problem, as in place and run.
func Try(o TryOpts) (TryReport, error) {
	lineKey, name, ok := strings.Cut(o.Key, "/")
	if !ok || !lineKeyRe.MatchString(lineKey) || !exampleNameRe.MatchString(name) {
		return TryReport{}, fault.BadArg(fault.BadExampleKey, o.Key)
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return TryReport{}, err
	}
	if o.Examples != "" {
		return tryDraft(root, o, lineKey, name)
	}
	if err := logstore.Find(root); err != nil {
		return TryReport{}, err
	}
	if o.Worktree && o.Commit != "" {
		return TryReport{}, fault.BadArg(fault.TryWorktreeCommit)
	}
	t, _, err := ReadTables(root)
	if err != nil {
		return TryReport{}, err
	}
	g, err := t.Find(o.Goal)
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
	rep := TryReport{Goal: g.EventID, ID: g.ID, Key: o.Key, Folder: rel, Words: ex.Words, Commit: g.Base, Worktree: o.Worktree,
		H0: !o.Worktree, RowTree: g.BaseTree}
	if rep.Hash, err = gitread.DirTreeHash(disk); err != nil {
		return TryReport{}, err
	}
	if fi, err := os.Stat(filepath.Join(disk, "probe")); err == nil && fi.IsDir() {
		if rep.ProbeHash, err = gitread.DirTreeHash(filepath.Join(disk, "probe")); err != nil {
			return TryReport{}, err
		}
	}
	switch {
	case o.Worktree:
		if rep.Commit, err = gitread.CommitOf(root, "HEAD"); err != nil {
			return TryReport{}, err
		}
		rep.RowTree = ""
	case o.Commit != "":
		if strings.HasPrefix(o.Commit, "-") {
			return TryReport{}, fault.BadArg(fault.TryNoCommit, o.Commit)
		}
		if ok, err := (gitread.Ancestry{Dir: root}).HasCommit(o.Commit); err != nil {
			return TryReport{}, err
		} else if !ok {
			return TryReport{}, fault.BadArg(fault.TryNoCommit, o.Commit)
		}
		if rep.Commit, err = gitread.CommitOf(root, o.Commit); err != nil {
			return TryReport{}, err
		}
		if rep.RowTree, err = gitread.RowTree(root, rep.Commit); err != nil {
			return TryReport{}, err
		}
		rep.H0 = rep.RowTree == g.BaseTree
	}

	run := rep.Commit
	if o.Worktree {
		run = g.Base
	}
	out, err := probe.Run(probe.Opts{
		Repo: root, Commit: run, ExampleDir: disk, Worktree: o.Worktree, Timeout: o.ProbeTimeout,
	})
	if err != nil {
		return TryReport{}, err
	}
	rep.Out = out
	rep.OutHash, rep.ErrHash = logstore.Hash(out.Stdout), logstore.Hash(out.Stderr)
	rep.Verdict = string(example.Compare(out.Stdout, ex))

	cur, placed := g.Current(o.Key)
	if o.Worktree || !placed || cur.Hash != rep.Hash || HasCell(g, o.Key, rep.RowTree) {
		return rep, nil
	}
	tr, err := logstore.TraceOf(root)
	if err != nil {
		return TryReport{}, err
	}
	for _, b := range [][]byte{out.Stdout, out.Stderr} {
		if _, err := tr.PutOutput(b); err != nil {
			return TryReport{}, err
		}
	}
	if err := tr.PutCell(logstore.Cell{
		Hash: rep.Hash, RowTree: rep.RowTree, Commit: rep.Commit, Probe: rep.ProbeHash,
		Output: rep.OutHash, Stderr: rep.ErrHash, Exit: out.Exit, Verdict: rep.Verdict, By: logstore.CellByTry,
	}); err != nil {
		return TryReport{}, err
	}
	rep.Saved = true
	return rep, nil
}

// tryDraft runs one example of a draft folder on the commit the Goal will
// name as H0, before the owner adds the Goal with add goal --examples: the
// folder as it is on disk, committed or not. It records nothing, and needs no
// store.
func tryDraft(root string, o TryOpts, lineKey, name string) (TryReport, error) {
	if o.Worktree || o.Commit != "" {
		return TryReport{}, fault.BadArg(fault.TryDraftOnH0)
	}
	dir, err := exampleFolder(o.Examples)
	if err != nil {
		return TryReport{}, err
	}
	base, baseTree, err := resolveBase(root, o.Base)
	if err != nil {
		return TryReport{}, err
	}
	rel := path.Join(dir, lineKey, name)
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
	rep := TryReport{ID: "the draft " + dir, Key: o.Key, Folder: rel, Words: ex.Words, Commit: base, H0: true, RowTree: baseTree}
	if rep.Hash, err = gitread.DirTreeHash(disk); err != nil {
		return TryReport{}, err
	}
	if fi, err := os.Stat(filepath.Join(disk, "probe")); err == nil && fi.IsDir() {
		if rep.ProbeHash, err = gitread.DirTreeHash(filepath.Join(disk, "probe")); err != nil {
			return TryReport{}, err
		}
	}
	out, err := probe.Run(probe.Opts{Repo: root, Commit: base, ExampleDir: disk, Timeout: o.ProbeTimeout})
	if err != nil {
		return TryReport{}, err
	}
	rep.Out = out
	rep.OutHash, rep.ErrHash = logstore.Hash(out.Stdout), logstore.Hash(out.Stderr)
	rep.Verdict = string(example.Compare(out.Stdout, ex))
	return rep, nil
}

// HasCell reports whether this clone holds the cell of the current placement
// of key on the row with rowTree with its outputs: a live cell of a Result,
// H0's cell of a place whose outputs this clone holds, or a cell a try saved.
func HasCell(g *tables.Goal, key, rowTree string) bool {
	for _, c := range g.LiveCells(rowTree) {
		if c.Key == key && (c.By != "" || c.Result != "") {
			return true
		}
	}
	cur, ok := g.Current(key)
	if !ok {
		return false
	}
	_, ok = g.TraceCell(cur.Hash, rowTree)
	return ok
}

// ReadTables reads every event of the store and of the clone's trace and
// calculates the tables, with the cells the trace keeps as files. It also
// returns the lines replay could not read.
func ReadTables(root string) (tables.Tables, []logstore.LineError, error) {
	evs, unread, err := logstore.Load(root)
	if err != nil {
		return tables.Tables{}, nil, err
	}
	t := tables.Replay(evs)
	tr, err := logstore.TraceOf(root)
	if err != nil {
		return tables.Tables{}, nil, err
	}
	cells, err := tr.Cells()
	if err != nil {
		return tables.Tables{}, nil, err
	}
	t.UseCells(cells)
	return Calculate(t, root), unread, nil
}
