// Package acts holds each hyp act: its guards in spec order, the one event it
// appends, and the report it returns.
package acts

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// Roles are the four roles an actor may declare.
var Roles = []string{"owner", "grill", "probe-writer", "runner"}

// DefaultBudget is the number of trees a Goal may run; three for every Goal.
const DefaultBudget = 3

// LineInput is one line of a Goal as the owner gives it. hyp gives the key.
type LineInput struct {
	Kind string
	Text string
}

// AddGoalOpts is the input of add goal. Goal is empty for a new Goal and the
// Goal's id for a Draft replace.
type AddGoalOpts struct {
	Dir   string // where the command runs
	Actor string
	Role  string
	Why   string
	Goal  string
	From  string
	Words string
	Base  string
	Lines []LineInput

	// Findings are the findings the Goal names, each by its id (F-1A2B3C) or
	// its finding.add id: the findings of this clone's trace it is about.
	Findings []string

	// Examples is the folder, relative to the top level, that holds the
	// Goal's example set as a place reads it, FOLDER/<line>/<example>: the
	// add runs each on H0 and carries each with its hash and its verdict
	// there, and the Goal is open at once. Empty adds a written Goal.
	Examples     string
	ProbeTimeout time.Duration
}

// Report is what add goal returns.
type Report struct {
	Goal     string // the Goal id
	Replaced bool
	Words    string
	Base     string // the full commit sha
	BaseTree string // H0's row tree
	Lines    []events.Line
	Findings []string    // the ids (F-…) of the findings it names
	Placed   []PlacedRow // the examples it carries, in key order
	Folder   string      // the Goal's folder the examples were copied to, or empty
	State    string      // the Goal's state after the add
}

// ValidRole reports whether role is one of the four.
func ValidRole(role string) bool { return slices.Contains(Roles, role) }

// CheckActor is the argument check of every act that writes: the actor
// matches ^[a-z0-9][a-z0-9-]{0,39}$ and the role is one of the four.
func CheckActor(actor, role string) error {
	switch {
	case actor == "":
		return fault.BadArg(fault.NoActor)
	case !logstore.ValidActor(actor):
		return fault.BadArg(fault.BadActor, actor)
	case role == "":
		return fault.BadArg(fault.NoRole, Roles)
	case !ValidRole(role):
		return fault.BadArg(fault.BadRole, role, Roles)
	}
	return nil
}

// AddGoal records a Goal, or replaces a written one, or refuses with the first
// failing guard: Draft, Whole lines, From ended, then for each finding it
// names Not refuted, Not declined and One in work; a finding this clone's
// trace does not hold is a refusal. It creates .tree/ at the git top level
// when it is absent, and only after the guards pass.
func AddGoal(o AddGoalOpts) (Report, error) {
	if err := CheckActor(o.Actor, o.Role); err != nil {
		return Report{}, err
	}
	if strings.TrimSpace(o.Words) == "" {
		return Report{}, fault.BadArg(fault.EmptyWords)
	}
	lines := make([]events.Line, len(o.Lines))
	for i, l := range o.Lines {
		if strings.TrimSpace(l.Text) == "" {
			return Report{}, fault.BadArg(fault.LineNoText, i+1)
		}
		lines[i] = events.Line{Key: "line-" + strconv.Itoa(i+1), Text: l.Text, Kind: l.Kind}
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return Report{}, err
	}
	base, baseTree, err := resolveBase(root, o.Base)
	if err != nil {
		return Report{}, err
	}
	if o.Examples != "" && o.Goal != "" {
		return Report{}, fault.Refuse(fault.GuardDraft, fault.ReplaceExamples, o.Goal)
	}
	var placed []events.PlacedExample
	var outs [][]byte
	var cells []logstore.Cell
	if o.Examples != "" {
		if placed, outs, cells, err = addedExamples(root, o, base, baseTree, lines); err != nil {
			return Report{}, err
		}
	}

	var rep Report
	build := func(evs []events.Event) (events.Event, error) {
		t := tables.Replay(evs)
		ev := events.Event{Type: events.TypeGoalAdd}
		goal := &tables.Goal{}
		if o.Goal != "" {
			g, err := t.Find(o.Goal)
			if err != nil {
				return ev, err
			}
			ev.Node, goal = g.EventID, g
		}
		from := ""
		if o.From != "" {
			g, err := t.Find(o.From)
			if err != nil {
				return ev, err
			}
			from = g.EventID
		}
		var self *tables.Goal
		if o.Goal != "" {
			self = goal
		}
		named, ids, err := nameFindings(t, o.Findings, self)
		if err != nil {
			return ev, err
		}
		payload, err := json.Marshal(events.GoalAdd{
			Role: o.Role, Words: o.Words, Base: base, BaseTree: baseTree,
			Budget: DefaultBudget, From: from, Lines: lines, Findings: named, Examples: placed,
		})
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeGoalAdd, err)
		}
		ev.Payload, ev.Why = payload, o.Why
		if err := tables.GuardGoalAdd(t, ev); err != nil {
			return ev, err
		}
		if o.From != "" {
			// From ended is the act's guard only: replay does not check
			// it, so a log written before land, whose Goal cites a Goal
			// that is not ended, replays as it was. A landed and a closed
			// Goal are both ended.
			cited, _ := t.Find(o.From)
			if cited.State != tables.StateEnded {
				return ev, fault.Refuse(fault.GuardFromEnded, fault.FromNotEnded, cited.ID, cited.State)
			}
		}
		rep = Report{Replaced: o.Goal != "", Words: o.Words, Base: base, BaseTree: baseTree, Lines: lines, Findings: ids}
		for _, x := range placed {
			rep.Placed = append(rep.Placed, PlacedRow{Key: x.Line + "/" + x.Example, Version: 1, Hash: x.Hash, Verdict: x.H0.Verdict, Exit: x.H0.Exit})
		}
		if rep.Replaced {
			rep.Goal = goal.ID
		}
		return ev, nil
	}

	if logstore.Find(root) != nil {
		// A refused first act must leave no store behind: run the guards
		// against an empty log before creating one.
		if _, err := build(nil); err != nil {
			return Report{}, err
		}
		if err := logstore.Create(root); err != nil {
			return Report{}, err
		}
	}
	ev, err := logstore.Append(root, o.Actor, build)
	if err != nil {
		return Report{}, err
	}
	// H0's outputs and cells of the examples it carries go to the clone's
	// trace, a cache a reader fills again with try
	if err := putOutputs(root, outs...); err != nil {
		return Report{}, err
	}
	if err := putCells(root, cells...); err != nil {
		return Report{}, err
	}
	if len(placed) > 0 {
		// the Goal's own folder gets its example set, as later places read
		// it by default; hyp writes no git state, so the owner commits it
		if err := copyExamples(root, o.Examples, path.Join("goals", tables.GoalID(ev.ID)), lines); err != nil {
			return Report{}, err
		}
		rep.Folder = path.Join("goals", tables.GoalID(ev.ID))
	}
	if !rep.Replaced {
		rep.Goal = tables.GoalID(ev.ID)
	}
	rep.State = tables.StateWritten
	if evs, _, err := logstore.Load(root); err == nil {
		if g := tables.Replay(evs).Goals[ev.Node]; g != nil {
			rep.State = g.State
		}
	}
	return rep, nil
}

// addedExamples reads the example set under o.Examples for a new Goal with
// lines, as a place reads a Goal's set: Whole set, Committed, Meetable, the
// names of each findings section and Covered, then Kind, which runs each
// example on H0. It returns the examples the add carries, and H0's outputs
// and cells for the clone's trace. A set with a cause is refused whole, and
// no Goal is recorded.
func addedExamples(root string, o AddGoalOpts, base, baseTree string, lines []events.Line) ([]events.PlacedExample, [][]byte, []logstore.Cell, error) {
	var t tables.Tables
	if logstore.Find(root) == nil {
		evs, _, err := logstore.Load(root)
		if err != nil {
			return nil, nil, nil, err
		}
		t = tables.Replay(evs)
	} else {
		t = tables.Replay(nil)
	}
	named, _, err := nameFindings(t, o.Findings, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	g := &tables.Goal{ID: "the new Goal", Base: base, BaseTree: baseTree, Lines: lines, Named: named}
	s, err := scanFolder(root, o.Examples, g, t)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := s.refusal(); err != nil {
		return nil, nil, nil, err
	}
	runs, err := measureOnH0(root, base, s.changes, o.ProbeTimeout)
	if err != nil {
		return nil, nil, nil, err
	}
	if causes := kindCauses(s.changes, runs); len(causes) > 0 {
		return nil, nil, nil, setRefusal(g.ID, causes)
	}
	placed, outs, cells := placedExamples(base, baseTree, s.changes, runs)
	return placed, outs, cells, nil
}

// nameFindings resolves each finding a Goal names in this clone's trace and
// checks the name: Not refuted, Not declined and One in work, with g the
// Goal that names them (nil for a new one). It returns their finding.add ids
// and their finding ids, each once, in the order given.
func nameFindings(t tables.Tables, names []string, g *tables.Goal) (eventIDs, ids []string, err error) {
	for _, name := range names {
		f, err := t.FindFinding(name)
		if err != nil {
			return nil, nil, err
		}
		if err := tables.GuardNameable(f, g); err != nil {
			return nil, nil, err
		}
		if !slices.Contains(eventIDs, f.EventID) {
			eventIDs = append(eventIDs, f.EventID)
			ids = append(ids, f.ID)
		}
	}
	return eventIDs, ids, nil
}

// resolveBase returns the full sha of the commit base names and its row tree.
// A base that is not a commit is a bad argument.
func resolveBase(root, base string) (commit, rowTree string, err error) {
	if base == "" || strings.HasPrefix(base, "-") {
		return "", "", fault.BadArg(fault.BaseNotCommit, base)
	}
	code, _, err := gitread.GitExit(root, "cat-file", "-e", base+"^{commit}")
	if err != nil {
		return "", "", err
	}
	if code != 0 {
		return "", "", fault.BadArg(fault.BaseNotHere, base)
	}
	out, err := gitread.Git(root, "rev-parse", "--verify", base+"^{commit}")
	if err != nil {
		return "", "", err
	}
	commit = strings.TrimSpace(string(out))
	rowTree, err = gitread.RowTree(root, commit)
	if err != nil {
		return "", "", err
	}
	return commit, rowTree, nil
}

// copyExamples copies each line folder from/<line> of the Goal's lines to
// to/<line>, file by file with each file's mode, so both folders hold the
// same tree hashes. A path that is already there is an environment problem:
// the copy overwrites nothing.
func copyExamples(root, from, to string, lines []events.Line) error {
	for _, l := range lines {
		src := filepath.Join(root, filepath.FromSlash(path.Join(from, l.Key)))
		dst := filepath.Join(root, filepath.FromSlash(path.Join(to, l.Key)))
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(src, p)
			if err != nil {
				return err
			}
			out := filepath.Join(dst, rel)
			if d.IsDir() {
				return os.MkdirAll(out, 0o755)
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			f, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
			if err != nil {
				return err
			}
			if _, err := f.Write(data); err != nil {
				f.Close()
				return err
			}
			return f.Close()
		})
		if err != nil {
			return fault.EnvFault(fault.CopyExamples, from, to, err)
		}
	}
	return nil
}
