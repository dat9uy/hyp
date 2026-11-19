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
	Commit    string // H0's commit, or HEAD's with Worktree
	Worktree  bool
	Out       probe.Out
	OutHash   string // the name the output would have under .tree/outputs
	ErrHash   string // the name the stderr would have under .tree/outputs
	Verdict   string
}

// Try runs one example's Probe and its compare as place and run do, and
// records nothing: no event, no output file, no run of the budget. It reads
// the example folder as it is on disk, committed or not, and runs it on H0's
// commit, or with Worktree on the files of this checkout as a commit of every
// change would hold them. It is a read: it checks no guard, so it runs in
// every state of the Goal. An unknown Goal, line or folder is a bad argument;
// a Probe that fails is an environment problem, as in place and run.
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
	rep := TryReport{Goal: g.EventID, ID: g.ID, Key: o.Key, Folder: rel, Words: ex.Words, Commit: g.Base, Worktree: o.Worktree}
	if rep.Hash, err = gitread.DirTreeHash(disk); err != nil {
		return TryReport{}, err
	}
	if fi, err := os.Stat(filepath.Join(disk, "probe")); err == nil && fi.IsDir() {
		if rep.ProbeHash, err = gitread.DirTreeHash(filepath.Join(disk, "probe")); err != nil {
			return TryReport{}, err
		}
	}
	if o.Worktree {
		if rep.Commit, err = gitread.CommitOf(root, "HEAD"); err != nil {
			return TryReport{}, err
		}
	}

	out, err := probe.Run(probe.Opts{
		Repo: root, Commit: g.Base, ExampleDir: disk, Worktree: o.Worktree, Timeout: o.ProbeTimeout,
	})
	if err != nil {
		return TryReport{}, err
	}
	rep.Out = out
	rep.OutHash, rep.ErrHash = logstore.Hash(out.Stdout), logstore.Hash(out.Stderr)
	rep.Verdict = string(example.Compare(out.Stdout, ex))
	return rep, nil
}
