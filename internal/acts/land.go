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

// LandOpts is the input of land.
type LandOpts struct {
	Dir          string // where the command runs
	Actor        string
	Role         string
	Why          string
	Goal         string // the Goal's id
	Commit       string // the commit to land; empty for the one the Goal is selected on, or none
	ProbeTimeout time.Duration
}

// LandReport is what land returns: the Goal, now landed, the stop it ended
// and the act that caused that stop; for a land on a commit, the commit, its
// tree and row tree, each example's verdict there, and the examples land ran
// because this clone held no Result for them.
type LandReport struct {
	Goal     string
	Stop     string
	Cause    Cause
	Commit   string
	Tree     string
	RowTree  string
	Verdicts []events.LandVerdict
	Ran      []string // the example keys land ran on the commit
}

// Land records the owner's land of a Goal, or refuses with the first failing
// guard, in the order its entry in hyp spec names: Not ended, Owner lands,
// Goal's actor, Evidence, Spec, Own path, then Stop. A land on a commit, the
// one named or the one this clone shows the Goal selected on, carries its
// evidence: the commit, its tree, its row tree and each current example's
// verdict there. For an example this clone holds no Result of on that row,
// land runs it first, and it refuses when one is not met, or when an
// unselect rejected the row. Every clone then ends the Goal from git alone.
// A Goal that is not selected and is given no commit lands at Two candidates
// or Budget used, with no commit. The land ends the Goal and names its
// snapshot: the Goal's events replay applied, which later Goals cite and
// which replay holds the ended Goal to in every clone. The examples run with
// the lock released, and every guard is checked again under the lock.
func Land(o LandOpts) (LandReport, error) {
	if err := CheckActor(o.Actor, o.Role); err != nil {
		return LandReport{}, err
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return LandReport{}, fault.BadArg(fault.WhyTooLong, n, logstore.MaxWhy)
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return LandReport{}, err
	}

	// the first check finds the commit and runs what this clone lacks
	var first landPlan
	err = logstore.WithLock(root, func(evs []events.Event) error {
		var err error
		first, err = planLand(root, o, evs)
		return err
	})
	if err != nil {
		return LandReport{}, err
	}
	ran := map[string]string{}
	var ranKeys []string
	if first.commit != "" {
		for _, p := range first.goal.Placed() {
			if first.goal.Verdict(p.Key, first.rowTree) != tables.NotRun {
				continue
			}
			ex, _, err := readExample(root, p.Hash)
			if err != nil {
				return LandReport{}, err
			}
			out, err := probe.Run(probe.Opts{Repo: root, Commit: first.commit, ExampleHash: p.Hash, Timeout: o.ProbeTimeout})
			if err != nil {
				return LandReport{}, err
			}
			ran[p.Key+"@"+p.Hash] = string(example.Compare(out.Stdout, ex))
			ranKeys = append(ranKeys, p.Key)
		}
	}

	var rep LandReport
	_, err = logstore.Append(root, o.Actor, func(evs []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeLand, Why: o.Why, Actor: o.Actor}
		again, err := planLand(root, o, evs)
		if err != nil {
			return ev, err
		}
		t, g := again.tables, again.goal
		p := events.Land{Role: o.Role, Seen: g.Seen()}
		rep = LandReport{Goal: g.ID}
		if again.commit == "" {
			st := Stand(g)
			if st.Stop == StopNone || st.Stop == StopSelected {
				return ev, fault.Refuse(fault.GuardStop, fault.NoStop, g.ID)
			}
			p.Stop, p.Cause = st.Stop, st.Cause.ID
			rep.Stop, rep.Cause = st.Stop, st.Cause
		} else {
			if again.commit != first.commit {
				return ev, fault.Refuse(fault.GuardEvidence, fault.EvidenceMissing, again.commit, "")
			}
			for _, x := range g.Placed() {
				v := g.Verdict(x.Key, again.rowTree)
				if v == tables.NotRun {
					v = ran[x.Key+"@"+x.Hash]
				}
				if v == "" {
					return ev, fault.Refuse(fault.GuardEvidence, fault.EvidenceMissing, again.commit, x.Key)
				}
				p.Verdicts = append(p.Verdicts, events.LandVerdict{Line: x.Line, Example: x.Example, Hash: x.Hash, Verdict: v})
			}
			p.Stop, p.Commit, p.Tree, p.RowTree = StopSelected, again.commit, again.tree, again.rowTree
			if err := tables.GuardEvidence(g, p); err != nil {
				return ev, err
			}
			read, err := readSpec(root, g, again.commit)
			if err != nil {
				return ev, err
			}
			if err := SpecGuard(g, again.commit, &read); err != nil {
				return ev, err
			}
			p.Read = &read
			changes, err := gitread.OwnChange(root, again.commit, g.Base)
			if err != nil {
				return ev, err
			}
			if err := tables.GuardOwnPath(g, changes); err != nil {
				return ev, err
			}
			cause := Cause{}
			if g.Select != nil && g.Select.RowTree == again.rowTree {
				cause = Cause{Act: ActRun, ID: g.Select.ID, Time: g.Select.Time}
				if !g.Select.Calculated {
					cause.Act = ActSelect
				}
			} else if c := lastRunOn(g, again.rowTree); c.ID != "" {
				cause = c
			}
			p.Cause = cause.ID
			rep.Stop, rep.Cause = StopSelected, cause
			rep.Commit, rep.Tree, rep.RowTree, rep.Verdicts, rep.Ran = p.Commit, p.Tree, p.RowTree, p.Verdicts, ranKeys
		}
		payload, err := json.Marshal(p)
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeLand, err)
		}
		ev.Node, ev.Payload = g.EventID, payload
		if err := tables.GuardLand(t, ev); err != nil {
			return ev, err
		}
		return ev, nil
	})
	if err != nil {
		return LandReport{}, err
	}
	return rep, nil
}

// landPlan is what the guards of a land that need no Probe found: the Goal,
// and the commit it lands with its tree and row tree, or no commit.
type landPlan struct {
	tables                tables.Tables
	goal                  *tables.Goal
	commit, tree, rowTree string
}

// planLand runs Not ended, Owner lands and Goal's actor, and finds the
// commit the land names: the one given, else the one this clone shows the
// Goal selected on, else none.
func planLand(root string, o LandOpts, evs []events.Event) (landPlan, error) {
	t := Calculate(tables.Replay(evs))
	g, err := t.Find(o.Goal)
	if err != nil {
		return landPlan{}, err
	}
	if err := tables.GuardNotEnded(g); err != nil {
		return landPlan{}, err
	}
	if err := tables.GuardOwnerLands(o.Role); err != nil {
		return landPlan{}, err
	}
	if err := tables.GuardGoalActor(g, o.Actor, o.Role); err != nil {
		return landPlan{}, err
	}
	lp := landPlan{tables: t, goal: g}
	name := o.Commit
	if name == "" && g.Select != nil {
		name = g.Select.Commit
	}
	if name == "" {
		return lp, nil
	}
	if lp.commit, lp.rowTree, err = commitAndRow(root, name); err != nil {
		return landPlan{}, err
	}
	out, err := gitread.Git(root, "rev-parse", "--verify", lp.commit+"^{tree}")
	if err != nil {
		return landPlan{}, err
	}
	lp.tree = strings.TrimSpace(string(out))
	if g.State == tables.StateWritten {
		return landPlan{}, fault.Refuse(fault.GuardEvidence, fault.LandNotOpen, g.ID, g.State)
	}
	if g.Rejected[lp.rowTree] {
		return landPlan{}, fault.Refuse(fault.GuardEvidence, fault.CommitUnselected, lp.commit, g.ID)
	}
	return lp, nil
}

// lastRunOn is the latest Result on the row with rowTree, in merge order,
// that counts.
func lastRunOn(g *tables.Goal, rowTree string) Cause {
	for i := len(g.Results) - 1; i >= 0; i-- {
		if r := g.Results[i]; r.RowTree == rowTree && !r.Retracted && r.Refused == "" {
			return Cause{Act: ActRun, ID: r.ID, Time: r.Time}
		}
	}
	return Cause{}
}
