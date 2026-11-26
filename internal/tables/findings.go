package tables

import (
	"slices"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/spec"
)

// Finding is a row of the Findings table: a fact an agent found, with a
// pointer to its evidence, and its state. The finding names no Goal: a Goal
// names it, at add goal or in an example it places, and an old fold counts as
// a name by its Goal. Its events are in the clone's trace: a finding changes
// no Goal's state.
type Finding struct {
	ID       string // F- and the last six characters of EventID
	EventID  string // the finding.add event that made it
	Fact     string
	Evidence string
	Draft    string // an add goal command line the agent drafted, or empty
	Actor    string
	Time     string
	Names    []*Goal // the Goals that name it, in the order replay met them
	Issue    int     // the issue of its latest fold in the old form, or 0
	Refuted  bool    // a finding.refute names it; refuted is final
	Declined bool    // the owner's finding.decline names it; declined is final
	Why      string  // the why of the decline, when Declined
}

// State is the finding's state, as spec.FindingStates names it: folded while
// a Goal in work names it; once no Goal in work does, settled when a Goal
// that names it landed with a select, else ended when any Goal named it.
func (f *Finding) State() string {
	switch {
	case f.Refuted:
		return spec.FindingRefuted
	case f.Declined:
		return spec.FindingDeclined
	case f.InWork() != nil || f.Issue > 0:
		return spec.FindingFolded
	case f.SettledBy() != nil:
		return spec.FindingSettled
	case len(f.Names) > 0:
		return spec.FindingEnded
	}
	return spec.FindingNotFolded
}

// InWork is the latest Goal in work that names the finding, or nil.
func (f *Finding) InWork() *Goal {
	for i := len(f.Names) - 1; i >= 0; i-- {
		if f.Names[i].State != StateEnded {
			return f.Names[i]
		}
	}
	return nil
}

// SettledBy is the latest Goal that names the finding and landed with a
// select, or nil.
func (f *Finding) SettledBy() *Goal {
	for i := len(f.Names) - 1; i >= 0; i-- {
		if g := f.Names[i]; g.Land != nil && g.Land.Stop == StopSelected {
			return g
		}
	}
	return nil
}

// Ended reports whether every Goal that names the finding has ended: it
// reads settled or ended.
func (f *Finding) Ended() bool {
	s := f.State()
	return s == spec.FindingSettled || s == spec.FindingEnded
}

// NamedBy reports whether g names the finding.
func (f *Finding) NamedBy(g *Goal) bool { return slices.Contains(f.Names, g) }

// name adds g to the Goals that name the finding, once.
func (f *Finding) name(g *Goal) {
	if g != nil && !f.NamedBy(g) {
		f.Names = append(f.Names, g)
	}
}

// SortedFindings lists every finding of the tables in merge order.
func SortedFindings(t Tables) []*Finding {
	return t.order
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

// GuardIssue refuses an old-form fold that names no issue: an issue is a
// whole number above zero.
func GuardIssue(issue int) error {
	if issue <= 0 {
		return fault.Refuse(fault.GuardIssue, fault.NoIssue)
	}
	return nil
}

// GuardInWork refuses an old fold into a Goal the tables do not hold, or
// into an ended Goal: a fold names a Goal in work, written, open or selected.
func GuardInWork(t Tables, node string) error {
	g, ok := t.Goals[node]
	if !ok {
		return fault.Refuse(fault.GuardInWork, fault.NoGoalToFold, GoalID(node))
	}
	if g.State == StateEnded {
		return fault.Refuse(fault.GuardInWork, fault.FoldIntoEnded, g.ID)
	}
	return nil
}

// GuardNotRefuted refuses a name, a refute or a decline of a refuted finding.
func GuardNotRefuted(f *Finding) error {
	if f.Refuted {
		return fault.Refuse(fault.GuardNotRefuted, fault.Refuted, f.ID)
	}
	return nil
}

// GuardNotDeclined refuses a name, a refute or a decline of a declined
// finding.
func GuardNotDeclined(f *Finding) error {
	if f.Declined {
		return fault.Refuse(fault.GuardNotDeclined, fault.Declined, f.ID)
	}
	return nil
}

// GuardNameable checks a name of finding f by Goal g, at add goal or at a
// place: the finding is not refuted, not declined, and no Goal in work but g
// names it. g is nil for a new Goal.
func GuardNameable(f *Finding, g *Goal) error {
	if err := GuardNotRefuted(f); err != nil {
		return err
	}
	if err := GuardNotDeclined(f); err != nil {
		return err
	}
	for _, h := range f.Names {
		if h != g && h.State != StateEnded {
			return fault.Refuse(fault.GuardOneInWork, fault.NamedInWork, f.ID, h.ID)
		}
	}
	return nil
}

// linkNames adds each Goal of the tables to the findings its record names:
// at add goal and in its placed examples. A name of a finding this clone's
// trace does not hold names no row.
func linkNames(t *Tables) {
	for _, g := range SortedGoals(*t) {
		for _, id := range g.Findings() {
			if f, ok := t.findings[id]; ok {
				f.name(g)
			}
		}
	}
}

// GuardOwnerDeclines refuses a decline by any role but the owner, and a
// decline with no why.
func GuardOwnerDeclines(role, why string) error {
	if role != "owner" {
		return fault.Refuse(fault.GuardOwnerDecl, fault.NotOwnerDecline, role)
	}
	if why == "" {
		return fault.Refuse(fault.GuardOwnerDecl, fault.DeclineNoWhy)
	}
	return nil
}

// GuardNotFolded refuses a decline of a finding that any Goal named, whether
// that Goal is in work or ended.
func GuardNotFolded(f *Finding) error {
	switch {
	case len(f.Names) > 0:
		return fault.Refuse(fault.GuardNotFolded, fault.DeclineFolded, f.ID, f.Names[len(f.Names)-1].ID)
	case f.Issue > 0:
		return fault.Refuse(fault.GuardNotFolded, fault.DeclineFoldedIssue, f.ID, f.Issue)
	}
	return nil
}

// GuardFindingAdd checks a finding.add against the tables: a node other than
// its own id, which only the old form has, names a Goal, in any state.
func GuardFindingAdd(t Tables, ev events.Event) error {
	var p events.FindingAdd
	if err := events.Decode(ev, &p); err != nil {
		return err
	}
	if ev.Node == "" || ev.Node == ev.ID {
		return nil
	}
	if _, ok := t.Goals[ev.Node]; !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalForFinding, GoalID(ev.Node))
	}
	return nil
}

// GuardFindingFold checks a finding.fold against the tables: it names a
// Goal in work, or in the old form an issue, and every finding it names
// exists and is not refuted.
func GuardFindingFold(t Tables, ev events.Event) error {
	var p events.FindingFold
	if err := events.Decode(ev, &p); err != nil {
		return err
	}
	switch {
	case p.Goal != "" && p.Issue != 0:
		return fault.Refuse(fault.NoGuard, fault.FoldGoalAndIssue)
	case p.Goal != "":
		if err := GuardInWork(t, p.Goal); err != nil {
			return err
		}
	default:
		if err := GuardIssue(p.Issue); err != nil {
			return err
		}
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
		if err := GuardNotDeclined(f); err != nil {
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
	if err := GuardNotRefuted(f); err != nil {
		return err
	}
	return GuardNotDeclined(f)
}

// GuardFindingDecline checks a finding.decline against the tables: the
// owner declines, with a why, a finding that is not refuted, not declined
// and not folded.
func GuardFindingDecline(t Tables, ev events.Event) error {
	var p events.FindingDecline
	if err := events.Decode(ev, &p); err != nil {
		return err
	}
	if err := GuardOwnerDeclines(p.Role, ev.Why); err != nil {
		return err
	}
	f, ok := t.findings[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoFinding, FindingID(ev.Node))
	}
	if err := GuardNotRefuted(f); err != nil {
		return err
	}
	if err := GuardNotDeclined(f); err != nil {
		return err
	}
	return GuardNotFolded(f)
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
		f := &Finding{ID: FindingID(ev.ID), EventID: ev.ID, Fact: p.Fact, Evidence: p.Evidence, Draft: p.Draft,
			Actor: ev.Actor, Time: ev.TS}
		t.findings[ev.ID] = f
		t.order = append(t.order, f)
	case events.TypeFindingFold:
		if err := GuardFindingFold(*t, ev); err != nil {
			refuse(ev, err)
			return
		}
		var p events.FindingFold
		if events.Decode(ev, &p) != nil {
			return
		}
		into := t.Goals[p.Goal] // nil for the oldest form
		for _, id := range p.Findings {
			f := t.findings[id]
			f.name(into)
			if p.Issue > 0 {
				f.Issue = p.Issue
			}
		}
	case events.TypeFindingRefute:
		if err := GuardFindingRefute(*t, ev); err != nil {
			refuse(ev, err)
			return
		}
		t.findings[ev.Node].Refuted = true
	case events.TypeFindingDecline:
		if err := GuardFindingDecline(*t, ev); err != nil {
			refuse(ev, err)
			return
		}
		f := t.findings[ev.Node]
		f.Declined, f.Why = true, ev.Why
	}
}
