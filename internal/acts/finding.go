package acts

import (
	"encoding/json"
	"unicode/utf8"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// FindingOpts is the input of the acts on a finding. Any actor runs them,
// with any role or none: Role is recorded only when it is one of the four.
type FindingOpts struct {
	Dir      string // where the command runs
	Actor    string
	Role     string
	Why      string
	Goal     string   // finding add: the Goal's id
	Fact     string   // finding add
	Evidence string   // finding add
	Findings []string // finding fold: one or more; finding refute: one
	Issue    int      // finding fold
}

// FindingReport is what an act on a finding returns: the Goal of each
// finding it wrote about, and each finding with its state after the act.
type FindingReport struct {
	Goals    []string // Goal ids, in the order of Findings, without repeats
	Findings []FindingLine
}

// FindingLine is one finding of a report.
type FindingLine struct {
	ID    string // F-…
	Goal  string // G-…
	State string // spec.FindingStates
	Issue int
}

// FindingAdd records a finding on a Goal in any state. A Goal the store does
// not hold is a refusal.
func FindingAdd(o FindingOpts) (FindingReport, error) {
	if o.Fact == "" {
		return FindingReport{}, fault.BadArg(fault.EmptyFact)
	}
	if o.Evidence == "" {
		return FindingReport{}, fault.BadArg(fault.EmptyEvidence)
	}
	return findingAct(o, events.TypeFindingAdd, func(t tables.Tables, ev *events.Event) error {
		g, err := t.Find(o.Goal)
		if err != nil {
			if fault.KindOf(err) == fault.BadArguments {
				if e, ok := fault.As(err); ok && e.Code == fault.NoGoal {
					return fault.Refuse(fault.NoGuard, fault.NoGoalForFinding, o.Goal)
				}
			}
			return err
		}
		raw, err := json.Marshal(events.FindingAdd{Role: o.Role, Fact: o.Fact, Evidence: o.Evidence})
		if err != nil {
			return fault.EnvFault(fault.EncodeEvent, events.TypeFindingAdd, err)
		}
		ev.Node, ev.Payload = g.EventID, raw
		return tables.GuardFindingAdd(t, *ev)
	})
}

// FindingFold folds one or more findings into an issue, in one event: the
// fold records all of them or none.
func FindingFold(o FindingOpts) (FindingReport, error) {
	return findingAct(o, events.TypeFindingFold, func(t tables.Tables, ev *events.Event) error {
		if err := tables.GuardIssue(o.Issue); err != nil {
			return err
		}
		p := events.FindingFold{Role: o.Role, Issue: o.Issue}
		for _, name := range o.Findings {
			f, err := t.FindFinding(name)
			if err != nil {
				return err
			}
			if err := tables.GuardNotRefuted(f); err != nil {
				return err
			}
			p.Findings = append(p.Findings, f.EventID)
		}
		raw, err := json.Marshal(p)
		if err != nil {
			return fault.EnvFault(fault.EncodeEvent, events.TypeFindingFold, err)
		}
		ev.Payload = raw
		return tables.GuardFindingFold(t, *ev)
	})
}

// FindingRefute refutes one finding. Refuted is final.
func FindingRefute(o FindingOpts) (FindingReport, error) {
	return findingAct(o, events.TypeFindingRefute, func(t tables.Tables, ev *events.Event) error {
		f, err := t.FindFinding(o.Findings[0])
		if err != nil {
			return err
		}
		if err := tables.GuardNotRefuted(f); err != nil {
			return err
		}
		raw, err := json.Marshal(events.FindingRefute{Role: o.Role})
		if err != nil {
			return fault.EnvFault(fault.EncodeEvent, events.TypeFindingRefute, err)
		}
		ev.Node, ev.Payload = f.EventID, raw
		return tables.GuardFindingRefute(t, *ev)
	})
}

// findingAct writes one finding event of type typ under the store's lock and
// the trace's: build checks its guards on the tables of every event and
// fills the event in. The event goes to the clone's trace, as its kind's
// entry says; nothing under .tree/ changes. The report reads the findings
// after the event, from the tables replayed with it.
func findingAct(o FindingOpts, typ string, build func(tables.Tables, *events.Event) error) (FindingReport, error) {
	if !logstore.ValidActor(o.Actor) {
		if o.Actor == "" {
			return FindingReport{}, fault.BadArg(fault.NoActor)
		}
		return FindingReport{}, fault.BadArg(fault.BadActor, o.Actor)
	}
	if !ValidRole(o.Role) {
		o.Role = ""
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return FindingReport{}, fault.BadArg(fault.WhyTooLong, n, logstore.MaxWhy)
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return FindingReport{}, err
	}
	var before []events.Event
	ev, err := logstore.Append(root, o.Actor, func(evs []events.Event) (events.Event, error) {
		ev := events.Event{Type: typ, Why: o.Why}
		if err := build(tables.Replay(evs), &ev); err != nil {
			return ev, err
		}
		before = evs
		return ev, nil
	})
	if err != nil {
		return FindingReport{}, err
	}
	t := tables.Replay(append(before, ev))
	var ids []string
	switch typ {
	case events.TypeFindingAdd:
		ids = []string{ev.ID}
	case events.TypeFindingFold:
		var p events.FindingFold
		if err := events.Decode(ev, &p); err != nil {
			return FindingReport{}, err
		}
		ids = p.Findings
	case events.TypeFindingRefute:
		ids = []string{ev.Node}
	}
	var rep FindingReport
	seen := map[string]bool{}
	for _, id := range ids {
		f, err := t.FindFinding(id)
		if err != nil {
			return FindingReport{}, err
		}
		g := tables.GoalID(f.Goal)
		if !seen[g] {
			seen[g] = true
			rep.Goals = append(rep.Goals, g)
		}
		rep.Findings = append(rep.Findings, FindingLine{ID: f.ID, Goal: g, State: f.State(), Issue: f.Issue})
	}
	return rep, nil
}
