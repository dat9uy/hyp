package acts

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/example"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/herr"
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
	Words   string
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

// String is the report as the command prints it.
func (r RunReport) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "ran %s %s at %s\n", r.Goal, r.Line, r.Commit)
	fmt.Fprintf(&b, "row: %s\n", r.RowTree)
	for _, c := range r.Cells {
		fmt.Fprintf(&b, "%s %q: %s (exit %d, output %s)\n", c.Key, c.Words, c.Verdict, c.Exit, c.Output)
	}
	return b.String()
}

// runChecked is what the guards that need no Probe found.
type runChecked struct {
	goal    *tables.Goal
	commit  string // the full sha
	rowTree string
	placed  []tables.Placement // the current placements of the line, in key order
}

// Run records one Result for every placed example of a line on a commit, or
// refuses with the first failing guard of the state machine, section 8: Not
// ended (always holds), Selected, Bytes, Open (not for H0), Own path, Leaf
// only, Budget, then Stable. Everything but Stable is checked before a Probe
// runs, so a refused run costs nothing; the Probes run with the lock
// released, and every guard is checked again under the lock before the event
// is appended. A run that Stable refuses keeps its outputs and records a
// Result marked refused, and returns the refusal.
func Run(o RunOpts) (RunReport, error) {
	if !logstore.ValidActor(o.Actor) {
		return RunReport{}, herr.BadArgs("actor %q must match ^[a-z0-9][a-z0-9-]{0,39}$ (--actor or $HYP_ACTOR)", o.Actor)
	}
	if !ValidRole(o.Role) {
		return RunReport{}, herr.BadArgs("role %q must be one of %s (--role or $HYP_ROLE)", o.Role, strings.Join(Roles, ", "))
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return RunReport{}, herr.BadArgs("why is %d runes; the limit is %d", n, logstore.MaxWhy)
	}
	if o.Commit == "" {
		return RunReport{}, herr.BadArgs("run needs a COMMIT")
	}
	cut := strings.LastIndex(o.Line, "/")
	if cut <= 0 || !lineKeyRe.MatchString(o.Line[cut+1:]) {
		return RunReport{}, herr.BadArgs("%q is not <Goal>/<line>, for example G-NXHHVA/line-1", o.Line)
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

	cells, words, err := runProbes(root, first, o.ProbeTimeout)
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
		payload := events.ResultRecord{Role: o.Role, Line: lineKey, Commit: again.commit, RowTree: again.rowTree, Cells: cells}
		if diff := tables.StableDiff(again.goal, lineKey, again.rowTree, cells); len(diff) > 0 {
			payload.Refused = "Stable"
			for _, d := range diff {
				payload.Differing = append(payload.Differing, d.Key)
			}
			stable = stableRefusal(root, again.goal, diff)
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return ev, herr.Env("encode result.record: %v", err)
		}
		ev.Node, ev.Payload = again.goal.EventID, raw
		if err := tables.GuardRun(tables.Replay(evs), ev); err != nil {
			return ev, err
		}
		rep = RunReport{Goal: again.goal.ID, Line: lineKey, Commit: again.commit, RowTree: again.rowTree}
		return ev, nil
	})
	if err != nil {
		return RunReport{}, err
	}
	evs, _, err := logstore.Events(root)
	if err != nil {
		return RunReport{}, err
	}
	g := tables.Replay(evs).Goals[ev.Node]
	if g == nil {
		return RunReport{}, herr.Env("Goal %s vanished from the log", rep.Goal)
	}
	rep.Used = tables.RunsUsed(g)
	rep.LastRun = rep.Used == tables.MaxTrees-1
	if stable != nil {
		return rep, stable
	}
	for i, c := range cells {
		rep.Cells = append(rep.Cells, RunCell{Key: lineKey + "/" + c.Example, Words: words[i], Verdict: c.Verdict, Output: c.Output, Exit: c.Exit})
	}
	return rep, nil
}

// checkRun runs the guards that need no Probe, in the order of the state
// machine: Selected, Bytes, Open, Own path, Leaf only, then Budget. An unknown
// Goal or line is a bad argument.
func checkRun(root string, o RunOpts, evs []events.Event, goalName, lineKey string) (runChecked, error) {
	t := tables.Replay(evs)
	g, err := t.Find(goalName)
	if err != nil {
		return runChecked{}, err
	}
	c := runChecked{goal: g}
	if _, ok := lineOf(g, lineKey); !ok {
		return runChecked{}, herr.BadArgs("%s has no line %s", g.ID, lineKey)
	}
	for _, p := range g.Placed() {
		if p.Line == lineKey {
			c.placed = append(c.placed, p)
		}
	}
	if err := tables.GuardNotSelected(g); err != nil {
		return runChecked{}, err
	}
	anc := gitread.Ancestry{Dir: root}
	if strings.HasPrefix(o.Commit, "-") {
		return runChecked{}, herr.Refused("Bytes: %q is not a commit", o.Commit)
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
		return herr.Refused("Open: H0 of %s moved from %s to %s while the Probes ran; run again", again.goal.ID, first.goal.Base, again.goal.Base)
	}
	if len(again.placed) != len(first.placed) {
		return herr.Refused("Open: the examples of the line changed while the Probes ran; run again")
	}
	for i, p := range first.placed {
		if q := again.placed[i]; q.Key != p.Key || q.Hash != p.Hash {
			return herr.Refused("Open: %s changed from %s to %s while the Probes ran; run again", p.Key, p.Hash, q.Hash)
		}
	}
	return nil
}

// runProbes runs the Probe of every placed example of the line on the commit
// and stores both outputs by hash. It returns the cells and each example's
// words, in key order.
func runProbes(root string, c runChecked, timeout time.Duration) ([]events.ExampleCell, []string, error) {
	var cells []events.ExampleCell
	var words []string
	for _, p := range c.placed {
		ex, probeHash, err := readExample(root, p.Hash)
		if err != nil {
			return nil, nil, err
		}
		out, err := probe.Run(probe.Opts{Repo: root, Commit: c.commit, ExampleHash: p.Hash, Timeout: timeout})
		if err != nil {
			return nil, nil, err
		}
		outHash, err := logstore.PutOutput(root, out.Stdout)
		if err != nil {
			return nil, nil, err
		}
		errHash, err := logstore.PutOutput(root, out.Stderr)
		if err != nil {
			return nil, nil, err
		}
		cells = append(cells, events.ExampleCell{
			Example: p.Example, Hash: p.Hash, Probe: probeHash, Output: outHash, Stderr: errHash,
			Exit: out.Exit, Verdict: string(example.Compare(out.Stdout, ex)),
		})
		words = append(words, ex.Words)
	}
	return cells, words, nil
}

// maxShown is how much of one stored output a Stable refusal prints.
const maxShown = 4096

// stableRefusal is the Stable refusal: for each example whose verdict
// differed, both verdicts and both outputs.
func stableRefusal(root string, g *tables.Goal, diff []tables.Unstable) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Stable: two runs of %s on one tree gave different verdicts; the first stands", g.ID)
	for _, d := range diff {
		fmt.Fprintf(&b, "\n%s: %s, then %s", d.Key, d.Was.Verdict, d.Now.Verdict)
		fmt.Fprintf(&b, "\n  first output %s:\n%s", d.Was.Output, indent(readOutput(root, d.Was.Output)))
		fmt.Fprintf(&b, "\n  second output %s:\n%s", d.Now.Output, indent(readOutput(root, d.Now.Output)))
	}
	return herr.Refused("%s", b.String())
}

// readOutput reads a stored output for display, at most maxShown bytes.
func readOutput(root, hash string) string { return readOutputMax(root, hash, maxShown) }

// readOutputMax reads a stored output for display, at most max bytes.
func readOutputMax(root, hash string, max int) string {
	f, err := os.Open(logstore.OutputPath(root, hash))
	if err != nil {
		return "(unreadable: " + err.Error() + ")"
	}
	defer f.Close()
	buf, err := io.ReadAll(io.LimitReader(f, int64(max)+1))
	if err != nil {
		return "(unreadable: " + err.Error() + ")"
	}
	if len(buf) > max {
		return string(buf[:max]) + "\n(truncated)"
	}
	return string(buf)
}
