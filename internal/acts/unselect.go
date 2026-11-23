package acts

import (
	"encoding/json"
	"unicode/utf8"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// UnselectOpts is the input of unselect.
type UnselectOpts struct {
	Dir   string // where the command runs
	Actor string
	Role  string
	Why   string
	Goal  string // the Goal's id
}

// UnselectReport is what unselect returns: the Goal, open again, and the
// select it undid.
type UnselectReport struct {
	Goal   string
	Select tables.Select
}

// Unselect records the owner's unselect of a selected Goal, or refuses with
// the first failing guard, in the order its entry in hyp spec names: Not ended, Owner
// unselects, Goal's actor, then Selected (the Goal has a select). The Goal is
// open again: its Results and cells stay, the budget starts again, and the
// runner selects again only once a run records a new cell (New cell). Every
// guard reads only the tables; everything is checked under the lock.
func Unselect(o UnselectOpts) (UnselectReport, error) {
	if err := CheckActor(o.Actor, o.Role); err != nil {
		return UnselectReport{}, err
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return UnselectReport{}, fault.BadArg(fault.WhyTooLong, n, logstore.MaxWhy)
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return UnselectReport{}, err
	}
	var rep UnselectReport
	_, err = logstore.Append(root, o.Actor, func(evs []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeUnselect, Why: o.Why, Actor: o.Actor}
		t := tables.Replay(evs)
		g, err := t.Find(o.Goal)
		if err != nil {
			return ev, err
		}
		payload, err := json.Marshal(events.Unselect{Role: o.Role})
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeUnselect, err)
		}
		ev.Node, ev.Payload = g.EventID, payload
		if err := tables.GuardUnselect(t, ev); err != nil {
			return ev, err
		}
		rep = UnselectReport{Goal: g.ID, Select: *g.Select}
		return ev, nil
	})
	if err != nil {
		return UnselectReport{}, err
	}
	return rep, nil
}
