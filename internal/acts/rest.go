package acts

import (
	"path/filepath"
	"slices"
	"time"

	"github.com/dat9uy/hyp/internal/example"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/probe"
	"github.com/dat9uy/hyp/internal/tables"
)

// measureRests measures, for every Goal in work that rests on a premise
// selected or landed on a commit, each placed example that this clone's
// trace holds no cell of on that commit, and saves the cells in the trace:
// the Goal then opens on the commit, or waits for the owner, with no act of
// the owner. The run that selects the premise, the land of the premise and
// the add of a resting Goal call it. It uses no run of any budget.
func measureRests(root string, timeout time.Duration) error {
	t, _, err := load(root, calculatedView)
	if err != nil {
		return err
	}
	for _, g := range tables.SortedGoals(t) {
		r := g.Rest
		if r == nil || r.Commit == "" || len(r.Missing) == 0 {
			continue
		}
		var cells []logstore.Cell
		var outs [][]byte
		for _, p := range g.Placed() {
			if !slices.Contains(r.Missing, p.Key) {
				continue
			}
			opts := probe.Opts{Repo: root, Commit: r.Commit, ExampleHash: p.Hash, Timeout: timeout}
			ex, probeHash, err := readExample(root, p.Hash)
			if err != nil {
				// the Goal's folder is not committed yet: the add copied
				// the bytes it hashed there
				dir := filepath.Join(root, "goals", g.ID, p.Line, p.Example)
				if h, herr := gitread.DirTreeHash(dir); herr != nil || h != p.Hash {
					return err
				}
				if ex, probeHash, err = readExampleDir(dir); err != nil {
					return err
				}
				opts.ExampleDir = dir
			}
			out, err := probe.Run(opts)
			if err != nil {
				return err
			}
			cells = append(cells, logstore.Cell{
				Hash: p.Hash, RowTree: r.RowTree, Commit: r.Commit, Probe: probeHash,
				Output: logstore.Hash(out.Stdout), Stderr: logstore.Hash(out.Stderr), Exit: out.Exit,
				Verdict: string(example.Compare(out.Stdout, ex)), By: logstore.CellByRest,
			})
			outs = append(outs, out.Stdout, out.Stderr)
		}
		if err := putOutputs(root, outs...); err != nil {
			return err
		}
		if err := putCells(root, cells...); err != nil {
			return err
		}
	}
	return nil
}
