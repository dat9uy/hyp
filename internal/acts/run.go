package acts

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/example"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/probe"
	"github.com/dat9uy/hyp/internal/tables"
)

// RunOpts is the input of run.
type RunOpts struct {
	Dir          string // where the command runs
	Actor        string
	Role         string
	Why          string
	Commit       string // a commit name
	Line         string // <Goal>/<line>, for example G-NXHHVA/line-3
	ProbeTimeout time.Duration

	// afterProbe, when set, runs after the Probes and before the second check;
	// a test uses it to land an event between the two checks.
	afterProbe func()
}

// RunCell is one example of a run.
type RunCell struct {
	Key     string // line-3/a
	Words   string // the example's words section, as the example holds it
	Verdict string
	Output  string // the output hash
	Exit    int
}

// RunReport is what run returns.
type RunReport struct {
	Goal    string
	Line    string
	Commit  string
	RowTree string
	Cells   []RunCell
	Used    int  // N of "runs: N of 3 used", which the state block prints
	LastRun bool // the next run is the last the Goal may make
}

// runChecked is what the guards that need no Probe found.
type runChecked struct {
	goal    *tables.Goal
	commit  string // the full sha
	rowTree string
	placed  []tables.Placement // the current placements of the line, in key order
	read    events.CommitRead  // what the run read of git about the commit, which its Result carries
}

// Run records one Result for every placed example of a line on a commit, or
// refuses with the first failing guard, in the order its entry in hyp spec names: Not
// ended (always holds), Selected, Bytes, Open (not for H0), Own path, Leaf
// only, Budget, then Stable. Everything but Stable is checked before a Probe
// runs, so a refused run costs nothing; the Probes run with the lock
// released, and every guard is checked again under the lock before the
// outputs are stored and the event is appended, so a refused run leaves no
// file. A run that Stable refuses keeps its outputs and records a Result
// marked refused, and returns the refusal: its event names those outputs.
func Run(o RunOpts) (RunReport, error) {
	if err := CheckActor(o.Actor, o.Role); err != nil {
		return RunReport{}, err
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return RunReport{}, fault.BadArg(fault.WhyTooLong, n, logstore.MaxWhy)
	}
	if o.Commit == "" {
		return RunReport{}, fault.BadArg(fault.NoCommit)
	}
	cut := strings.LastIndex(o.Line, "/")
	if cut <= 0 || !lineKeyRe.MatchString(o.Line[cut+1:]) {
		return RunReport{}, fault.BadArg(fault.BadGoalLine, o.Line)
	}
	goalName, lineKey := o.Line[:cut], o.Line[cut+1:]
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return RunReport{}, err
	}

	var first runChecked
	err = logstore.WithLock(root, func(evs []events.Event) error {
		var err error
		first, err = checkRun(root, o, evs, goalName, lineKey)
		return err
	})
	if err != nil {
		return RunReport{}, err
	}

	cells, outs, words, err := runProbes(root, first, o.ProbeTimeout)
	if err != nil {
		return RunReport{}, err
	}
	if o.afterProbe != nil {
		o.afterProbe()
	}

	var rep RunReport
	var stable error
	ev, err := logstore.Append(root, o.Actor, func(evs []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeResultRecord, Why: o.Why}
		again, err := checkRun(root, o, evs, goalName, lineKey)
		if err != nil {
			return ev, err
		}
		if err := sameBasis(first, again); err != nil {
			return ev, err
		}
		payload := events.ResultRecord{Role: o.Role, Line: lineKey, Commit: again.commit, RowTree: again.rowTree, Cells: cells, Read: &again.read}
		if diff := tables.StableDiff(again.goal, lineKey, again.rowTree, cells); len(diff) > 0 {
			payload.Refused = "Stable"
			for _, d := range diff {
				payload.Differing = append(payload.Differing, d.Key)
			}
			stable = stableRefusal(root, again.goal, diff)
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeResultRecord, err)
		}
		ev.Node, ev.Payload = again.goal.EventID, raw
		if err := tables.GuardRun(tables.Replay(evs), ev); err != nil {
			return ev, err
		}
		if err := putOutputs(root, outs...); err != nil {
			return ev, err
		}
		rep = RunReport{Goal: again.goal.ID, Line: lineKey, Commit: again.commit, RowTree: again.rowTree}
		return ev, nil
	})
	if err != nil {
		return RunReport{}, err
	}
	evs, _, err := logstore.Load(root)
	if err != nil {
		return RunReport{}, err
	}
	g := tables.Replay(evs).Goals[ev.Node]
	if g == nil {
		return RunReport{}, fault.EnvFault(fault.GoalVanished, rep.Goal)
	}
	rep.Used = tables.RunsUsed(g)
	rep.LastRun = rep.Used == tables.MaxHypotheses-1
	if stable != nil {
		return rep, stable
	}
	for i, c := range cells {
		rep.Cells = append(rep.Cells, RunCell{Key: lineKey + "/" + c.Example, Words: words[i], Verdict: c.Verdict, Output: c.Output, Exit: c.Exit})
	}
	return rep, nil
}

// checkRun runs the guards that need no Probe, in the order of the state
// machine: Selected, Bytes, Open, Own path, Leaf only, then Budget. Then it
// reads what the Result carries of the commit (ReadCommit). An unknown Goal
// or line is a bad argument.
func checkRun(root string, o RunOpts, evs []events.Event, goalName, lineKey string) (runChecked, error) {
	t := tables.Replay(evs)
	g, err := t.Find(goalName)
	if err != nil {
		return runChecked{}, err
	}
	c := runChecked{goal: g}
	if _, ok := lineOf(g, lineKey); !ok {
		return runChecked{}, fault.BadArg(fault.NoLine, g.ID, lineKey)
	}
	for _, p := range g.Placed() {
		if p.Line == lineKey {
			c.placed = append(c.placed, p)
		}
	}
	if err := tables.GuardNotEnded(g); err != nil {
		return runChecked{}, err
	}
	if err := tables.GuardNotSelected(g); err != nil {
		return runChecked{}, err
	}
	anc := gitread.Ancestry{Dir: root}
	if strings.HasPrefix(o.Commit, "-") {
		return runChecked{}, fault.Refuse(fault.GuardBytes, fault.NotACommit, o.Commit)
	}
	if err := tables.GuardBytes(anc, o.Commit); err != nil {
		return runChecked{}, err
	}
	if c.commit, err = gitread.CommitOf(root, o.Commit); err != nil {
		return runChecked{}, err
	}
	if c.rowTree, err = gitread.RowTree(root, c.commit); err != nil {
		return runChecked{}, err
	}
	if err := tables.GuardOpen(g, c.rowTree, lineKey); err != nil {
		return runChecked{}, err
	}
	changes, err := gitread.OwnChange(root, c.commit, g.Base)
	if err != nil {
		return runChecked{}, err
	}
	if err := tables.GuardOwnPath(g, changes); err != nil {
		return runChecked{}, err
	}
	if err := tables.GuardLeafOnly(g, anc, c.commit); err != nil {
		return runChecked{}, err
	}
	if err := tables.GuardBudget(g, c.rowTree); err != nil {
		return runChecked{}, err
	}
	if c.read, err = ReadCommit(root, g, c.commit); err != nil {
		return runChecked{}, err
	}
	return c, nil
}

func lineOf(g *tables.Goal, key string) (tables.Line, bool) {
	for _, l := range g.Lines {
		if l.Key == key {
			return l, true
		}
	}
	return tables.Line{}, false
}

// sameBasis refuses a run whose placements or H0 moved while the Probes ran:
// the cells it holds measured examples that are no longer the current ones.
func sameBasis(first, again runChecked) error {
	if again.goal.Base != first.goal.Base {
		return fault.Refuse(fault.GuardOpen, fault.H0MovedWhileRun, again.goal.ID, first.goal.Base, again.goal.Base)
	}
	if len(again.placed) != len(first.placed) {
		return fault.Refuse(fault.GuardOpen, fault.ExamplesChanged)
	}
	for i, p := range first.placed {
		if q := again.placed[i]; q.Key != p.Key || q.Hash != p.Hash {
			return fault.Refuse(fault.GuardOpen, fault.ExampleChanged, p.Key, p.Hash, q.Hash)
		}
	}
	return nil
}

// runProbes runs the Probe of every placed example of the line on the commit.
// It returns the cells, which name both outputs by hash, the outputs' bytes,
// stdout then stderr of each example, and each example's words, in key
// order. It stores nothing: Run stores the outputs once its guards hold.
func runProbes(root string, c runChecked, timeout time.Duration) ([]events.ExampleCell, [][]byte, []string, error) {
	var cells []events.ExampleCell
	var outs [][]byte
	var words []string
	for _, p := range c.placed {
		ex, probeHash, err := readExample(root, p.Hash)
		if err != nil {
			return nil, nil, nil, err
		}
		out, err := probe.Run(probe.Opts{Repo: root, Commit: c.commit, ExampleHash: p.Hash, Timeout: timeout})
		if err != nil {
			return nil, nil, nil, err
		}
		cells = append(cells, events.ExampleCell{
			Example: p.Example, Hash: p.Hash, Probe: probeHash,
			Output: logstore.Hash(out.Stdout), Stderr: logstore.Hash(out.Stderr),
			Exit: out.Exit, Verdict: string(example.Compare(out.Stdout, ex)),
		})
		outs = append(outs, out.Stdout, out.Stderr)
		words = append(words, ex.Words)
	}
	return cells, outs, words, nil
}

// StableFacts is what a Stable refusal shows: the Goal, and for each example
// whose verdict differed, the first verdict, which stands, and the second,
// each with the hash of its output.
type StableFacts struct {
	Goal  string
	Diffs []StableDiff
}

// StableDiff is one example a Stable refusal found unstable.
type StableDiff struct {
	Key                   string
	FirstVerdict, Verdict string
	FirstOutput, Output   string // output hashes, under .tree/outputs/
}

// stableRefusal is the Stable refusal: for each example whose verdict
// differed, both verdicts and both outputs' hashes.
func stableRefusal(root string, g *tables.Goal, diff []tables.Unstable) error {
	facts := StableFacts{Goal: g.ID}
	for _, d := range diff {
		facts.Diffs = append(facts.Diffs, StableDiff{
			Key: d.Key, FirstVerdict: d.Was.Verdict, Verdict: d.Now.Verdict,
			FirstOutput: d.Was.Output, Output: d.Now.Output,
		})
	}
	return fault.Refuse(fault.GuardStable, fault.VerdictsDiffered, facts)
}
