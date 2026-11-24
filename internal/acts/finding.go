package acts

import (
	"encoding/json"
	"strings"
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
	Fact     string   // finding add
	Evidence string   // finding add
	Draft    string   // finding add: an add goal command line for the owner, or empty
	Findings []string // finding fold: one or more; finding refute: one
	Goal     string   // finding fold: the Goal to fold into
}

// FindingReport is what an act on a finding returns: the Goal a fold names,
// and each finding with its state after the act.
type FindingReport struct {
	Goal     string // G-… of a fold; empty for an add or a refute
	Findings []*tables.Finding
}

// DraftPrefix starts every draft a finding carries: a draft is an add goal
// command line, which the owner may edit and run.
const DraftPrefix = "hyp add goal "

// FindingAdd records a finding: a fact and its evidence, with no Goal, and
// an add goal the agent drafted for the owner when Draft is set.
func FindingAdd(o FindingOpts) (FindingReport, error) {
	if o.Fact == "" {
		return FindingReport{}, fault.BadArg(fault.EmptyFact)
	}
	if o.Evidence == "" {
		return FindingReport{}, fault.BadArg(fault.EmptyEvidence)
	}
	if o.Draft != "" && !strings.HasPrefix(o.Draft, DraftPrefix) {
		return FindingReport{}, fault.BadArg(fault.DraftNotAddGoal)
	}
	return findingAct(o, events.TypeFindingAdd, func(t tables.Tables, ev *events.Event) error {
		raw, err := json.Marshal(events.FindingAdd{Role: o.Role, Fact: o.Fact, Evidence: o.Evidence, Draft: o.Draft})
		if err != nil {
			return fault.EnvFault(fault.EncodeEvent, events.TypeFindingAdd, err)
		}
		ev.Payload = raw
		return tables.GuardFindingAdd(t, *ev)
	})
}

// FindingFold folds one or more findings into one Goal in work, in one
// event: the fold records all of them or none. A Goal the store does not
// hold, or an ended Goal, is a refusal of In work.
func FindingFold(o FindingOpts) (FindingReport, error) {
	return findingAct(o, events.TypeFindingFold, func(t tables.Tables, ev *events.Event) error {
		g, err := t.Find(o.Goal)
		if err != nil {
			if e, ok := fault.As(err); ok && e.Code == fault.NoGoal {
				return fault.Refuse(fault.GuardInWork, fault.NoGoalToFold, o.Goal)
			}
			return err
		}
		if err := tables.GuardInWork(t, g.EventID); err != nil {
			return err
		}
		node := g.EventID
		p := events.FindingFold{Role: o.Role, Goal: node}
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
	for _, id := range ids {
		f, err := t.FindFinding(id)
		if err != nil {
			return FindingReport{}, err
		}
		rep.Findings = append(rep.Findings, f)
		if typ == events.TypeFindingFold && f.Into != nil {
			rep.Goal = f.Into.ID
		}
	}
	return rep, nil
}
