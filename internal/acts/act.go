package acts

import (
	"unicode/utf8"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// view is which select holds a Goal selected in the tables an act reads.
type view int

const (
	// eventsView is the state the events give: only a select event of the
	// older form holds a Goal selected. The owner's place reads it, so a
	// select this clone's Results show lets the owner's place through and
	// the place moves the Goal back to open; close and the finding acts
	// read it too.
	eventsView view = iota
	// calculatedView adds the select this clone's Results show
	// (tables.Calculate). Run reads it, so the runner stops at the first
	// selectable candidate; select, unselect and land read it, and so does
	// every read that prints a Goal.
	calculatedView
)

// snapshot is what the tables of an act are calculated from: the events of
// the store and of the clone's trace, and the cells the trace keeps as files.
// An act reads both together, from readSnapshot, so no caller can calculate
// tables that leave the cells out.
type snapshot struct {
	evs   []events.Event
	cells map[string]logstore.Cell
}

// readSnapshot holds evs with the cells the clone's trace at root keeps as
// files. Under the lock it reads the cells once, with the events, so a
// refusal names H0's outputs as the clone's trace holds them.
func readSnapshot(root string, evs []events.Event) (snapshot, error) {
	tr, err := logstore.TraceOf(root)
	if err != nil {
		return snapshot{}, err
	}
	cells, err := tr.Cells()
	if err != nil {
		return snapshot{}, err
	}
	return snapshot{evs: evs, cells: cells}, nil
}

// snapshotOf holds evs with no cells. It is for the tables of a log that is
// not the store as it stands: an empty log before a store exists, or the
// log after an event an act just appended. The question such tables answer
// reads no Goal's cells, so the trace's cells are not part of it.
func snapshotOf(evs []events.Event) snapshot {
	return snapshot{evs: evs}
}

// tablesOf is the one place the tables are calculated: it replays the
// events of s, gives every Goal the cells of s and, in calculatedView, the
// select this clone's Results show.
func tablesOf(s snapshot, v view) tables.Tables {
	t := tables.Replay(s.evs)
	t.UseCells(s.cells)
	if v == calculatedView {
		t = tables.Calculate(t)
	}
	return t
}

// load reads every event of the store and of the clone's trace, and the
// cells the trace keeps as files, and calculates the tables in view v. It
// also returns the lines replay could not read. It takes no lock: an act
// checks its guards only on what it reads under the lock (writer.check and
// writer.append).
func load(root string, v view) (tables.Tables, []logstore.LineError, error) {
	evs, unread, err := logstore.Load(root)
	if err != nil {
		return tables.Tables{}, nil, err
	}
	s, err := readSnapshot(root, evs)
	if err != nil {
		return tables.Tables{}, nil, err
	}
	return tablesOf(s, v), unread, nil
}

// ReadTables reads every event of the store and of the clone's trace and
// calculates the tables, with the cells the trace keeps as files and the
// select this clone's Results show. It also returns the lines replay could
// not read.
func ReadTables(root string) (tables.Tables, []logstore.LineError, error) {
	return load(root, calculatedView)
}

// checkWhy is the length check of a why, the one every write act runs.
func checkWhy(why string) error {
	if n := utf8.RuneCountInString(why); n > logstore.MaxWhy {
		return fault.BadArg(fault.WhyTooLong, n, logstore.MaxWhy)
	}
	return nil
}

// writer is one write act on the store at root, by actor.
type writer struct {
	root  string
	actor string
}

// begin runs the argument checks every write act of a role runs before it
// reads the store, in this order: the actor and the role (CheckActor), the
// why's length, then the git top level of dir.
func begin(dir, actor, role, why string) (writer, error) {
	if err := CheckActor(actor, role); err != nil {
		return writer{}, err
	}
	if err := checkWhy(why); err != nil {
		return writer{}, err
	}
	return writerAt(dir, actor)
}

// writerAt finds the git top level of dir for a write act whose arguments are
// checked.
func writerAt(dir, actor string) (writer, error) {
	root, err := logstore.TopLevel(dir)
	if err != nil {
		return writer{}, err
	}
	return writer{root: root, actor: actor}, nil
}

// check takes the store's lock and the trace's, calls f with the tables of
// every event in view v, and releases the locks. It writes nothing: an act
// checks its guards with it before it runs a Probe, which must not hold the
// locks, and checks them again in append.
func (w writer) check(v view, f func(tables.Tables) error) error {
	return logstore.WithLock(w.root, func(evs []events.Event) error {
		s, err := readSnapshot(w.root, evs)
		if err != nil {
			return err
		}
		return f(tablesOf(s, v))
	})
}

// append takes the store's lock and the trace's, calls build with the
// tables of every event in view v and those events, and appends the event
// build returns; an error from build writes nothing. The tables are read
// under the lock, never carried over from a read before it, from one
// snapshot of the events and the trace's cells.
func (w writer) append(v view, build func(tables.Tables, []events.Event) (events.Event, error)) (events.Event, error) {
	return w.appendBefore(v, build, nil)
}

// appendBefore is append with before called with the event and its id just
// before the event is written; an error from before leaves the event unwritten.
func (w writer) appendBefore(v view, build func(tables.Tables, []events.Event) (events.Event, error), before func(events.Event) error) (events.Event, error) {
	return logstore.AppendBefore(w.root, w.actor, func(evs []events.Event) (events.Event, error) {
		s, err := readSnapshot(w.root, evs)
		if err != nil {
			return events.Event{}, err
		}
		return build(tablesOf(s, v), evs)
	}, before)
}
