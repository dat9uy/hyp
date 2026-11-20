package acts

import (
	"encoding/json"
	"unicode/utf8"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// LandOpts is the input of land.
type LandOpts struct {
	Dir   string // where the command runs
	Actor string
	Role  string
	Why   string
	Goal  string // the Goal's id
}

// LandReport is what land returns: the Goal, now landed, the stop it ended
// and the act that caused that stop.
type LandReport struct {
	Goal  string
	Stop  string
	Cause Cause
}

// Land records the owner's land of a Goal at a stop, or refuses with the
// first failing guard of the state machine, section 8: Not ended, Owner
// lands, Goal's actor, then Stop. The land ends the Goal: its events up to
// the land, and the outputs they name, are the snapshot later Goals cite. It
// checks no commit and does not need a merge. Everything is checked under the
// lock.
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
	var rep LandReport
	_, err = logstore.Append(root, o.Actor, func(evs []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeLand, Why: o.Why, Actor: o.Actor}
		t := tables.Replay(evs)
		g, err := t.Find(o.Goal)
		if err != nil {
			return ev, err
		}
		if err := tables.GuardNotEnded(g); err != nil {
			return ev, err
		}
		if err := tables.GuardOwnerLands(o.Role); err != nil {
			return ev, err
		}
		if err := tables.GuardGoalActor(g, o.Actor, o.Role); err != nil {
			return ev, err
		}
		st := Stand(g, root)
		if st.Stop == StopNone {
			return ev, fault.Refuse(fault.GuardStop, fault.NoStop, g.ID)
		}
		payload, err := json.Marshal(events.Land{Role: o.Role, Stop: st.Stop, Cause: st.Cause.ID})
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeLand, err)
		}
		ev.Node, ev.Payload = g.EventID, payload
		if err := tables.GuardLand(t, ev); err != nil {
			return ev, err
		}
		rep = LandReport{Goal: g.ID, Stop: st.Stop, Cause: st.Cause}
		return ev, nil
	})
	if err != nil {
		return LandReport{}, err
	}
	return rep, nil
}
