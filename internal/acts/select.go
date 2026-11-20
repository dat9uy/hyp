package acts

import (
	"encoding/json"

	"unicode/utf8"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// SelectOpts is the input of select.
type SelectOpts struct {
	Dir   string // where the command runs
	Actor string
	Role  string
	Why   string
	Goal  string // the Goal's id
	Tree  string // the row tree to select, or a unique prefix of it
}

// SelectReport is what select returns: the Goal, now selected, and the
// commit and row tree it recorded. Its report is the PR body, which the
// channel prints from Selected.
type SelectReport struct {
	Goal     string
	Commit   string
	RowTree  string
	Selected *tables.Goal // the Goal as the log holds it after the select
}

// Select records a select, or refuses with the first failing guard of the
// state machine, section 8: Not ended (always holds), Selected, Candidate,
// New cell (for the runner), Goal's actor (for the owner), Spec, then Own
// path. Spec and Own path read the
// commit of the row's latest Result in merge order, which is the commit the
// select records. Everything is checked under the lock.
func Select(o SelectOpts) (SelectReport, error) {
	if err := CheckActor(o.Actor, o.Role); err != nil {
		return SelectReport{}, err
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return SelectReport{}, fault.BadArg(fault.WhyTooLong, n, logstore.MaxWhy)
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return SelectReport{}, err
	}

	var rep SelectReport
	ev, err := logstore.Append(root, o.Actor, func(evs []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeSelect, Why: o.Why, Actor: o.Actor}
		t := tables.Replay(evs)
		g, err := t.Find(o.Goal)
		if err != nil {
			return ev, err
		}
		if err := tables.GuardNotEnded(g); err != nil {
			return ev, err
		}
		if err := tables.GuardNotSelected(g); err != nil {
			return ev, err
		}
		row, err := tables.GuardCandidate(g, gitread.Ancestry{Dir: root}, o.Tree, o.Role)
		if err != nil {
			return ev, err
		}
		if err := tables.GuardNewCell(g, o.Role); err != nil {
			return ev, err
		}
		if err := tables.GuardGoalActor(g, o.Actor, o.Role); err != nil {
			return ev, err
		}
		commit := latestCommit(g, row.RowTree)
		if commit == "" {
			return ev, fault.Refuse(fault.GuardCandidate, fault.NoResultOnRow, row.RowTree)
		}
		if err := guardSpec(root, g, commit); err != nil {
			return ev, err
		}
		changes, err := gitread.OwnChange(root, commit, g.Base)
		if err != nil {
			return ev, err
		}
		if err := tables.GuardOwnPath(g, changes); err != nil {
			return ev, err
		}
		payload, err := json.Marshal(events.Select{Role: o.Role, Commit: commit, RowTree: row.RowTree})
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeSelect, err)
		}
		ev.Node, ev.Payload = g.EventID, payload
		if err := tables.GuardSelect(t, ev); err != nil {
			return ev, err
		}
		rep = SelectReport{Goal: g.ID, Commit: commit, RowTree: row.RowTree}
		return ev, nil
	})
	if err != nil {
		return SelectReport{}, err
	}
	evs, _, err := logstore.Events(root)
	if err != nil {
		return SelectReport{}, err
	}
	g := tables.Replay(evs).Goals[ev.Node]
	if g == nil || g.Select == nil {
		return SelectReport{}, fault.EnvFault(fault.SelectNotInLog, rep.Goal)
	}
	rep.Selected = g
	return rep, nil
}

// latestCommit is the commit of the latest Result, in merge order, that counts
// on the row with rowTree: not retracted and not refused by Stable.
func latestCommit(g *tables.Goal, rowTree string) string {
	commit := ""
	for _, r := range g.Results {
		if r.RowTree == rowTree && !r.Retracted && r.Refused == "" {
			commit = r.Commit
		}
	}
	return commit
}

// guardSpec is the Spec guard: the commit holds a spec under
// goals/<G>/specs/. When the spec is a model (a .tla file), every rule its
// check configs check is named by a switch config under specs/switches/, and
// the folder holds a chart (.mmd). hyp reads the files and runs no model
// checker.
func guardSpec(root string, g *tables.Goal, commit string) error {
	dir := SpecDir(g.ID)
	files, err := SpecFiles(root, commit, g.ID)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fault.Refuse(fault.GuardSpec, fault.NoSpec, commit, dir)
	}
	m, err := ReadModel(root, files)
	if err != nil {
		return err
	}
	if !m.Present {
		return nil
	}
	if missing := m.Unswitched(); len(missing) > 0 {
		return fault.Refuse(fault.GuardSpec, fault.RuleNoSwitch, commit, missing, dir+switchesDir)
	}
	if len(m.Charts) == 0 {
		return fault.Refuse(fault.GuardSpec, fault.NoChart, commit, dir)
	}
	return nil
}

// ExampleWords is the words section of the example folder with tree hash
// hash, as the example holds it.
func ExampleWords(root, hash string) (string, error) {
	ex, _, err := readExample(root, hash)
	if err != nil {
		return "", err
	}
	return ex.Words, nil
}
