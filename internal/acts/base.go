package acts

import (
	"time"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/example"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/probe"
	"github.com/dat9uy/hyp/internal/tables"
)

// A run's commit can stand on code newer than H0: a fix another Goal or PR
// merged after the Goal was added. The run measures every placed example of
// the Goal on that base before it runs the line, and moves the Goal onto the
// base only when each example gives there the floors and ceilings it gives
// on H0. Measuring a base uses no run of the budget: its cells go to the
// clone's trace, not into a Result.

// BaseFacts is what a Same base refusal shows: the Goal, the commit run, its
// base, and each example that measures differently there than on H0.
type BaseFacts struct {
	Goal, Commit, Base string
	Diffs              []BaseDiff
}

// BaseDiff is one example that measures differently on the base than on H0:
// each verdict, the floors and ceilings that changed, and whether H0 also
// measures it differently now than at its place, so the environment changed
// and not the code.
type BaseDiff struct {
	Key                    string
	H0Verdict, BaseVerdict string
	Changed                []example.Mark // as they hold on H0
	Environment            bool
}

// baseCheck is what measuring a run's base found.
type baseCheck struct {
	read  *events.BaseRead // nil when the commit stands on no code newer than H0
	diffs []BaseDiff
	cells []logstore.Cell // the cells measured, which the trace keeps
	outs  [][]byte        // their outputs
}

// measureBase finds the base of the run's commit and, when its subject is
// not H0's, measures every placed example of the Goal there. Cells a run
// measured on the base before, which agree with H0, are not measured again.
// An example that measures differently on the base is measured on H0 again,
// so the refusal says whether the environment changed.
func measureBase(root string, c runChecked, timeout time.Duration) (baseCheck, error) {
	g := c.goal
	bases, err := gitread.Bases(root, c.commit, g.Base, g.ID, g.RowBases())
	if err != nil {
		return baseCheck{}, err
	}
	if len(bases) == 0 {
		return baseCheck{}, nil
	}
	if len(bases) > 1 {
		return baseCheck{}, fault.Refuse(fault.GuardSameBase, fault.TwoBases, c.commit, g.ID, bases)
	}
	base := bases[0]
	tree, err := gitread.RowTree(root, base)
	if err != nil {
		return baseCheck{}, err
	}
	if tree == g.BaseTree {
		return baseCheck{}, nil
	}
	bc := baseCheck{read: &events.BaseRead{Commit: base, RowTree: tree}}
	for _, p := range g.Placed() {
		ex, probeHash, err := readExample(root, p.Hash)
		if err != nil {
			return baseCheck{}, err
		}
		measure := func(commit, rowTree, by string) (logstore.Cell, []byte, error) {
			out, err := probeRun(root, commit, p.Hash, timeout)
			if err != nil {
				return logstore.Cell{}, nil, err
			}
			cell := logstore.Cell{
				Hash: p.Hash, RowTree: rowTree, Commit: commit, Probe: probeHash,
				Output: logstore.Hash(out.Stdout), Stderr: logstore.Hash(out.Stderr), Exit: out.Exit,
				Verdict: string(example.Compare(out.Stdout, ex)), By: by,
			}
			return cell, out.Stdout, nil
		}
		keep := func(cell logstore.Cell, stdout []byte) {
			bc.cells = append(bc.cells, cell)
			bc.outs = append(bc.outs, stdout)
		}
		// H0 as the place measured it, or as this clone measures it now
		// when its trace lacks the place's output
		h0Verdict, h0Out, ok := traceOutput(root, g, p.Hash, g.BaseTree)
		if !ok {
			cell, stdout, err := measure(g.Base, g.BaseTree, logstore.CellByTry)
			if err != nil {
				return baseCheck{}, err
			}
			keep(cell, stdout)
			h0Verdict, h0Out = cell.Verdict, stdout
		}
		h0Marks := example.Marks(h0Out, ex)
		if _, out, ok := traceOutput(root, g, p.Hash, tree); ok && example.SameMarks(example.Marks(out, ex), h0Marks) {
			continue
		}
		cell, stdout, err := measure(base, tree, logstore.CellByBase)
		if err != nil {
			return baseCheck{}, err
		}
		keep(cell, stdout)
		baseMarks := example.Marks(stdout, ex)
		if example.SameMarks(baseMarks, h0Marks) {
			continue
		}
		d := BaseDiff{Key: p.Key, H0Verdict: h0Verdict, BaseVerdict: cell.Verdict}
		for i, m := range h0Marks {
			if baseMarks[i] != m {
				d.Changed = append(d.Changed, m)
			}
		}
		now, err := probeRun(root, g.Base, p.Hash, timeout)
		if err != nil {
			return baseCheck{}, err
		}
		d.Environment = !example.SameMarks(example.Marks(now.Stdout, ex), h0Marks)
		bc.diffs = append(bc.diffs, d)
		bc.read.Differing = append(bc.read.Differing, p.Key)
		if d.Environment {
			bc.read.Environment = append(bc.read.Environment, p.Key)
		}
	}
	return bc, nil
}

// traceOutput is the verdict and the stdout of the cell the clone's trace
// keeps for the example with hash on the row with rowTree, if it keeps one
// whose output it can read.
func traceOutput(root string, g *tables.Goal, hash, rowTree string) (string, []byte, bool) {
	tc, ok := g.TraceCell(hash, rowTree)
	if !ok || !logstore.HasOutput(root, tc.Output) {
		return "", nil, false
	}
	data, _, err := logstore.ReadOutput(root, tc.Output, 0)
	if err != nil {
		return "", nil, false
	}
	return tc.Verdict, data, true
}

// probeRun runs the Probe of the example with hash on commit.
func probeRun(root, commit, hash string, timeout time.Duration) (probe.Out, error) {
	return probe.Run(probe.Opts{Repo: root, Commit: commit, ExampleHash: hash, Timeout: timeout})
}

// refusal is the Same base refusal of the run of commit, with the
// differences bc found.
func (bc baseCheck) refusal(g *tables.Goal, commit string) error {
	return fault.Refuse(fault.GuardSameBase, fault.BaseDiffers, BaseFacts{Goal: g.ID, Commit: commit, Base: bc.read.Commit, Diffs: bc.diffs})
}
