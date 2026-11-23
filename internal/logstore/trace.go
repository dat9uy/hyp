package logstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
)

// Trace is where hyp keeps what it printed to each actor of one clone: a
// folder of events, outside every worktree and outside the git directory, so
// no git command sees it and no commit can carry it. Every worktree of one
// clone shares one git common directory, and so one trace folder.
type Trace struct {
	Dir   string // the folder of events: events/<actor>.jsonl, outputs/<sha256>, cells/
	Clone string // the clone's git common directory, which names the folder
}

// TracesDir is the folder that holds the trace folder of every clone:
// $HOME/.hyp/traces. hyp reads no other variable for it: not
// $XDG_STATE_HOME, and nothing under $HOME/.local. With no absolute $HOME
// there is no such folder, which is an environment problem.
func TracesDir() (string, error) {
	home := os.Getenv("HOME")
	if !filepath.IsAbs(home) {
		return "", fault.EnvFault(fault.NoTraceFolder)
	}
	return filepath.Join(home, ".hyp", "traces"), nil
}

// TraceOf returns the trace folder of the clone that holds root:
// $HOME/.hyp/traces/<sha256 of the clone's git common directory>. It creates
// nothing.
func TraceOf(root string) (Trace, error) {
	out, err := gitread.Git(root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return Trace{}, err
	}
	clone := strings.TrimSpace(string(out))
	if real, err := filepath.EvalSymlinks(clone); err == nil {
		clone = real
	}
	base, err := TracesDir()
	if err != nil {
		return Trace{}, err
	}
	return Trace{Dir: filepath.Join(base, Hash([]byte(clone))), Clone: clone}, nil
}

// Outputs is the folder that holds the trace's outputs.
func (t Trace) Outputs() string { return outputsIn(t.Dir) }

// Events returns the trace's events in merge order, and the lines that could
// not be used. A trace folder not yet written holds no event.
func (t Trace) Events() ([]events.Event, []LineError, error) { return eventsOf("", t.Dir) }

// create makes the trace folder, and names the clone it belongs to in a file
// clone beside events/ and outputs/, for a person who finds the folder.
func (t Trace) create() error {
	for _, d := range []string{eventsIn(t.Dir), outputsIn(t.Dir)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fault.EnvFault(fault.CreateTrace, err)
		}
	}
	name := filepath.Join(t.Dir, "clone")
	if _, err := os.Stat(name); err == nil {
		return nil
	}
	if err := os.WriteFile(name, []byte(t.Clone+"\n"), 0o644); err != nil {
		return fault.EnvFault(fault.CreateTrace, err)
	}
	return nil
}

// PutOutput stores b in the trace by its sha256 and returns the hash, before
// the read event that names it.
func (t Trace) PutOutput(b []byte) (string, error) {
	if err := t.create(); err != nil {
		return "", err
	}
	return putOutputIn(t.Outputs(), b)
}

// Append writes ev as one fsynced line to the actor's file of the trace,
// under the trace folder's lock, with V, ID, TS and Actor filled in as Append
// fills them for the store.
func (t Trace) Append(actor string, ev events.Event) (events.Event, error) {
	if !ValidActor(actor) {
		return events.Event{}, fault.BadArg(fault.BadActor, actor)
	}
	if err := t.create(); err != nil {
		return events.Event{}, err
	}
	unlock, err := lockDir(t.Dir)
	if err != nil {
		return events.Event{}, err
	}
	defer unlock()
	return appendLocked(t.Dir, actor, events.HomeTrace, func([]events.Event) (events.Event, error) { return ev, nil })
}

// Cell is one cell of an example on one row that the clone's trace keeps as
// a file, not as an event: H0's cell of a place, whose event in git names no
// output, or a cell a try ran again because this clone's trace lacked it.
// Its outputs are in the trace's outputs/.
type Cell struct {
	Hash    string `json:"hash"`     // the example folder's tree hash
	RowTree string `json:"row_tree"` // the subject hash of the row
	Commit  string `json:"commit"`
	Probe   string `json:"probe"`
	Output  string `json:"output"`
	Stderr  string `json:"stderr"`
	Exit    int    `json:"exit"`
	Verdict string `json:"verdict"`
	By      string `json:"by"` // CellByPlace or CellByTry
}

// Who saved a cell of the trace.
const (
	CellByPlace = "place" // the place that measured H0
	CellByTry   = "try"   // a try that ran the cell again in this clone
)

// CellKey is the key of the cell of the example with hash on the row with
// rowTree.
func CellKey(hash, rowTree string) string { return hash + " " + rowTree }

func (t Trace) cellsDir() string { return filepath.Join(t.Dir, "cells") }

// PutCell saves c in the trace as cells/<hash>-<row tree>.json, in place of
// a cell saved there before.
func (t Trace) PutCell(c Cell) error {
	if err := t.create(); err != nil {
		return err
	}
	if err := os.MkdirAll(t.cellsDir(), 0o755); err != nil {
		return fault.EnvFault(fault.StoreOutput, err)
	}
	data, err := json.Marshal(c)
	if err != nil {
		return fault.EnvFault(fault.StoreOutput, err)
	}
	tmp, err := os.CreateTemp(t.cellsDir(), "tmp-*")
	if err != nil {
		return fault.EnvFault(fault.StoreOutput, err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return fault.EnvFault(fault.StoreOutput, err)
	}
	if err := tmp.Close(); err != nil {
		return fault.EnvFault(fault.StoreOutput, err)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(t.cellsDir(), c.Hash+"-"+c.RowTree+".json")); err != nil {
		return fault.EnvFault(fault.StoreOutput, err)
	}
	return nil
}

// Cells returns the cells the trace keeps, by CellKey. A file that does not
// parse is left out: the trace is a cache, which a try fills again.
func (t Trace) Cells() (map[string]Cell, error) {
	out := map[string]Cell{}
	paths, err := filepath.Glob(filepath.Join(t.cellsDir(), "*.json"))
	if err != nil {
		return nil, fault.EnvFault(fault.ReadEvents, err)
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fault.EnvFault(fault.ReadEvents, err)
		}
		var c Cell
		if json.Unmarshal(data, &c) != nil || c.Hash == "" {
			continue
		}
		out[CellKey(c.Hash, c.RowTree)] = c
	}
	return out, nil
}
