package acts

import (
	"encoding/json"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/tables"
)

// CloseOpts is the input of close.
type CloseOpts struct {
	Dir   string // where the command runs
	Actor string
	Role  string
	Why   string
	Goal  string // the Goal's id
}

// CloseReport is what close returns: the Goal, now ended.
type CloseReport struct {
	Goal string
}

// Close records the owner's close of a Goal that is not ended, at any stop
// or none, or refuses with the first failing guard, in the order its entry
// in hyp spec names: Not ended, Owner closes, then Goal's actor. Like a land,
// the close ends the Goal and names its snapshot. It checks no commit. Everything
// is checked under the lock.
func Close(o CloseOpts) (CloseReport, error) {
	w, err := begin(o.Dir, o.Actor, o.Role, o.Why)
	if err != nil {
		return CloseReport{}, err
	}
	var rep CloseReport
	_, err = w.append(eventsView, func(t tables.Tables, _ []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeClose, Why: o.Why, Actor: o.Actor}
		g, err := t.Find(o.Goal)
		if err != nil {
			return ev, err
		}
		payload, err := json.Marshal(events.Close{Role: o.Role, Seen: g.Seen()})
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeClose, err)
		}
		ev.Node, ev.Payload = g.EventID, payload
		if err := tables.GuardClose(t, ev); err != nil {
			return ev, err
		}
		rep = CloseReport{Goal: g.ID}
		return ev, nil
	})
	if err != nil {
		return CloseReport{}, err
	}
	return rep, nil
}
