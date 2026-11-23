package tables

import (
	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/spec"
)

// Finding is a row of the Findings table: a fact an agent found on a Goal,
// with a pointer to its evidence, and its state. Its events are in the
// clone's trace: a finding changes no Goal's state.
type Finding struct {
	ID       string // F- and the last six characters of EventID
	EventID  string // the finding.add event that made it
	Goal     string // the goal.add event id of its Goal
	Fact     string
	Evidence string
	Actor    string
	Time     string
	Issue    int  // the issue of its latest fold, or 0
	Refuted  bool // a finding.refute names it; refuted is final
}

// State is the finding's state, as spec.FindingStates names it.
func (f *Finding) State() string {
	switch {
	case f.Refuted:
		return spec.FindingRefuted
	case f.Issue > 0:
		return spec.FindingFolded
	}
	return spec.FindingNotFolded
}

// FindingID is the finding id of a finding.add event id.
func FindingID(eventID string) string {
	tail := eventID
	if len(tail) > 6 {
		tail = tail[len(tail)-6:]
	}
	return "F-" + tail
}

// FindFinding resolves a finding named by its id (F-1A2B3C) or by its
// finding.add event id. A finding the tables do not hold is a refusal; a
// short id that matches two findings is a bad argument.
func (t Tables) FindFinding(name string) (*Finding, error) {
	if f, ok := t.findings[name]; ok {
		return f, nil
	}
	var found *Finding
	n := 0
	for _, f := range t.findings {
		if f.ID == name {
			found = f
			n++
		}
	}
	switch n {
	case 0:
		return nil, fault.Refuse(fault.NoGuard, fault.NoFinding, name)
	case 1:
		return found, nil
	}
	return nil, fault.BadArg(fault.FindingTwice, name, n)
}

// GuardIssue refuses a fold that names no issue: an issue is a whole number
// above zero.
func GuardIssue(issue int) error {
	if issue <= 0 {
		return fault.Refuse(fault.GuardIssue, fault.NoIssue)
	}
	return nil
}

// GuardNotRefuted refuses a fold or a refute of a refuted finding.
func GuardNotRefuted(f *Finding) error {
	if f.Refuted {
		return fault.Refuse(fault.GuardNotRefuted, fault.Refuted, f.ID)
	}
	return nil
}

// GuardFindingAdd checks a finding.add against the tables: its node names a
// Goal, in any state.
func GuardFindingAdd(t Tables, ev events.Event) error {
	var p events.FindingAdd
	if err := events.Decode(ev, &p); err != nil {
		return err
	}
	if _, ok := t.Goals[ev.Node]; !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalForFinding, GoalID(ev.Node))
	}
	return nil
}

// GuardFindingFold checks a finding.fold against the tables: it names an
// issue, and every finding it names exists and is not refuted.
func GuardFindingFold(t Tables, ev events.Event) error {
	var p events.FindingFold
	if err := events.Decode(ev, &p); err != nil {
		return err
	}
	if err := GuardIssue(p.Issue); err != nil {
		return err
	}
	if len(p.Findings) == 0 {
		return fault.Refuse(fault.NoGuard, fault.FoldNoFinding)
	}
	for _, id := range p.Findings {
		f, ok := t.findings[id]
		if !ok {
			return fault.Refuse(fault.NoGuard, fault.NoFinding, FindingID(id))
		}
		if err := GuardNotRefuted(f); err != nil {
			return err
		}
	}
	return nil
}

// GuardFindingRefute checks a finding.refute against the tables: its node
// names a finding that is not refuted.
func GuardFindingRefute(t Tables, ev events.Event) error {
	var p events.FindingRefute
	if err := events.Decode(ev, &p); err != nil {
		return err
	}
	f, ok := t.findings[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoFinding, FindingID(ev.Node))
	}
	return GuardNotRefuted(f)
}

// replayFinding applies a finding event at its place in the merge order,
// when its guards allow it. No finding event is in an end's snapshot: the
// snapshot names only events stored in git.
func replayFinding(t *Tables, ev events.Event, refuse func(events.Event, error)) {
	switch ev.Type {
	case events.TypeFindingAdd:
		if err := GuardFindingAdd(*t, ev); err != nil {
			refuse(ev, err)
			return
		}
		var p events.FindingAdd
		if events.Decode(ev, &p) != nil {
			return // GuardFindingAdd decoded it already
		}
		f := &Finding{ID: FindingID(ev.ID), EventID: ev.ID, Goal: ev.Node, Fact: p.Fact, Evidence: p.Evidence,
			Actor: ev.Actor, Time: ev.TS}
		g := t.Goals[ev.Node]
		g.Findings = append(g.Findings, f)
		t.findings[ev.ID] = f
	case events.TypeFindingFold:
		if err := GuardFindingFold(*t, ev); err != nil {
			refuse(ev, err)
			return
		}
		var p events.FindingFold
		if events.Decode(ev, &p) != nil {
			return
		}
		for _, id := range p.Findings {
			t.findings[id].Issue = p.Issue
		}
	case events.TypeFindingRefute:
		if err := GuardFindingRefute(*t, ev); err != nil {
			refuse(ev, err)
			return
		}
		t.findings[ev.Node].Refuted = true
	}
}
