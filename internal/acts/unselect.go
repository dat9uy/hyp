package acts

import (
	"encoding/json"
	"strings"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/tables"
)

// UnselectOpts is the input of unselect.
type UnselectOpts struct {
	Dir    string // where the command runs
	Actor  string
	Role   string
	Why    string
	Goal   string // the Goal's id
	Commit string // the commit the owner rejects; empty for the one the Goal is selected on
}

// UnselectReport is what unselect returns: the Goal, open again, and the
// select it undid.
type UnselectReport struct {
	Goal   string
	Select tables.Select
}

// Unselect records the owner's unselect of a selected Goal, or refuses with
// the first failing guard, in the order its entry in hyp spec names: Not ended, Owner
// unselects, Goal's actor, then Selected (this clone shows the Goal selected
// on the commit). The unselect names the commit it rejects and its row tree:
// no clone shows the Goal selected on that row again, and a land on it is
// refused. The Goal is open again: its Results and cells stay, the budget
// starts again, and a select is calculated again only once a run records a
// new cell (New cell). Everything is checked under the lock.
func Unselect(o UnselectOpts) (UnselectReport, error) {
	w, err := begin(o.Dir, o.Actor, o.Role, o.Why)
	if err != nil {
		return UnselectReport{}, err
	}
	var rep UnselectReport
	_, err = w.append(calculatedView, func(t tables.Tables, _ []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeUnselect, Why: o.Why, Actor: o.Actor}
		g, err := t.Find(o.Goal)
		if err != nil {
			return ev, err
		}
		if err := tables.GuardNotEnded(g); err != nil {
			return ev, err
		}
		if err := tables.GuardOwnerUnselects(o.Role); err != nil {
			return ev, err
		}
		if err := tables.GuardGoalActor(g, o.Actor, o.Role); err != nil {
			return ev, err
		}
		if err := tables.GuardHasSelect(g); err != nil {
			return ev, err
		}
		sel := *g.Select
		if o.Commit != "" {
			commit, rowTree, err := commitAndRow(w.root, o.Commit)
			if err != nil {
				return ev, err
			}
			if err := tables.GuardSelectedOn(g, commit, rowTree); err != nil {
				return ev, err
			}
		}
		payload, err := json.Marshal(events.Unselect{Role: o.Role, Commit: sel.Commit, RowTree: sel.RowTree})
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeUnselect, err)
		}
		ev.Node, ev.Payload = g.EventID, payload
		if err := tables.GuardUnselect(t, ev); err != nil {
			return ev, err
		}
		rep = UnselectReport{Goal: g.ID, Select: sel}
		return ev, nil
	})
	if err != nil {
		return UnselectReport{}, err
	}
	return rep, nil
}

// commitAndRow resolves a commit name this clone holds to its full sha and
// its row tree. A name that is no commit here is refused by Bytes.
func commitAndRow(root, name string) (commit, rowTree string, err error) {
	if strings.HasPrefix(name, "-") {
		return "", "", fault.Refuse(fault.GuardBytes, fault.NotACommit, name)
	}
	if err := tables.GuardBytes(gitread.Ancestry{Dir: root}, name); err != nil {
		return "", "", err
	}
	if commit, err = gitread.CommitOf(root, name); err != nil {
		return "", "", err
	}
	if rowTree, err = gitread.RowTree(root, commit); err != nil {
		return "", "", err
	}
	return commit, rowTree, nil
}
