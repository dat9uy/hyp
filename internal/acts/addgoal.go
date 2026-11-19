// Package acts holds each hyp act: its guards in spec order, the one event it
// appends, and the report it returns.
package acts

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/herr"
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
}

// Report is what add goal returns.
type Report struct {
	Goal     string // the Goal id
	Replaced bool
	Words    string
	Base     string // the full commit sha
	BaseTree string // H0's row tree
	Lines    []events.Line
}

// ValidRole reports whether role is one of the four.
func ValidRole(role string) bool { return slices.Contains(Roles, role) }

// AddGoal records a Goal, or replaces a written one, or refuses with the first
// failing guard: Draft, then Whole lines. It creates .tree/ at the git top
// level when it is absent, and only after the guards pass.
func AddGoal(o AddGoalOpts) (Report, error) {
	if !logstore.ValidActor(o.Actor) {
		return Report{}, herr.BadArgs("actor %q must match ^[a-z0-9][a-z0-9-]{0,39}$ (--actor or $HYP_ACTOR)", o.Actor)
	}
	if !ValidRole(o.Role) {
		return Report{}, herr.BadArgs("role %q must be one of %s (--role or $HYP_ROLE)", o.Role, strings.Join(Roles, ", "))
	}
	if strings.TrimSpace(o.Words) == "" {
		return Report{}, herr.BadArgs("the Goal's words are empty")
	}
	lines := make([]events.Line, len(o.Lines))
	for i, l := range o.Lines {
		if strings.TrimSpace(l.Text) == "" {
			return Report{}, herr.BadArgs("line %d has no text", i+1)
		}
		lines[i] = events.Line{Key: fmt.Sprintf("line-%d", i+1), Text: l.Text, Kind: l.Kind}
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return Report{}, err
	}
	base, baseTree, err := resolveBase(root, o.Base)
	if err != nil {
		return Report{}, err
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
		payload, err := json.Marshal(events.GoalAdd{
			Role: o.Role, Words: o.Words, Base: base, BaseTree: baseTree,
			Budget: DefaultBudget, From: from, Lines: lines,
		})
		if err != nil {
			return ev, herr.Env("encode goal.add: %v", err)
		}
		ev.Payload, ev.Why = payload, o.Why
		if err := tables.GuardGoalAdd(t, ev); err != nil {
			return ev, err
		}
		rep = Report{Replaced: o.Goal != "", Words: o.Words, Base: base, BaseTree: baseTree, Lines: lines}
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
	if !rep.Replaced {
		rep.Goal = tables.GoalID(ev.ID)
	}
	return rep, nil
}

// resolveBase returns the full sha of the commit base names and its row tree.
// A base that is not a commit is a bad argument.
func resolveBase(root, base string) (commit, rowTree string, err error) {
	if base == "" || strings.HasPrefix(base, "-") {
		return "", "", herr.BadArgs("--base %q is not a commit", base)
	}
	code, _, err := gitread.GitExit(root, "cat-file", "-e", base+"^{commit}")
	if err != nil {
		return "", "", err
	}
	if code != 0 {
		return "", "", herr.BadArgs("--base %q is not a commit of this repository", base)
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
