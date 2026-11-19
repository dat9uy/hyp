package acts

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/herr"
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

// SelectReport is what select returns. Body is the PR body.
type SelectReport struct {
	Goal    string
	Commit  string
	RowTree string
	Body    string
}

// String is the report as the command prints it: the PR body.
func (r SelectReport) String() string { return r.Body }

// Select records a select, or refuses with the first failing guard of the
// state machine, section 8: Not ended (always holds), Selected, Candidate,
// Goal's actor (for the owner), Spec, then Own path. Spec and Own path read the
// commit of the row's latest Result in merge order, which is the commit the
// select records. Everything is checked under the lock.
func Select(o SelectOpts) (SelectReport, error) {
	if !logstore.ValidActor(o.Actor) {
		return SelectReport{}, herr.BadArgs("actor %q must match ^[a-z0-9][a-z0-9-]{0,39}$ (--actor or $HYP_ACTOR)", o.Actor)
	}
	if !ValidRole(o.Role) {
		return SelectReport{}, herr.BadArgs("role %q must be one of %s (--role or $HYP_ROLE)", o.Role, strings.Join(Roles, ", "))
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return SelectReport{}, herr.BadArgs("why is %d runes; the limit is %d", n, logstore.MaxWhy)
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
		if err := tables.GuardNotSelected(g); err != nil {
			return ev, err
		}
		row, err := tables.GuardCandidate(g, gitread.Ancestry{Dir: root}, o.Tree, o.Role)
		if err != nil {
			return ev, err
		}
		if err := tables.GuardGoalActor(g, o.Actor, o.Role); err != nil {
			return ev, err
		}
		commit := latestCommit(g, row.RowTree)
		if commit == "" {
			return ev, herr.Refused("Candidate: row %s has no Result to select", row.RowTree)
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
			return ev, herr.Env("encode select: %v", err)
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
		return SelectReport{}, herr.Env("the select of %s is not in the log", rep.Goal)
	}
	rep.Body = PRBody(root, g)
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
// goals/<G>/specs/.
func guardSpec(root string, g *tables.Goal, commit string) error {
	dir := path.Join("goals", g.ID, "specs") + "/"
	out, err := gitread.Git(root, "ls-tree", "-r", "-z", commit, dir)
	if err != nil {
		return err
	}
	if len(out) == 0 {
		return herr.Refused("Spec: commit %s holds no file under %s; commit the spec, then run one line on that commit", commit, dir)
	}
	return nil
}

// prOutputMax is how much of one stored output the PR body prints.
const prOutputMax = 16384

var backtickRun = regexp.MustCompile("`+")

// fence wraps text in a code fence longer than any run of backticks in it.
func fence(text string) string {
	n := 3
	for _, m := range backtickRun.FindAllString(text, -1) {
		n = max(n, len(m)+1)
	}
	bar := strings.Repeat("`", n)
	return bar + "\n" + strings.TrimSuffix(text, "\n") + "\n" + bar + "\n"
}

// PRBody is the PR body of a selected Goal (section 12): its words, each line
// with its kind, each placed example with its words, output, stderr, exit code
// and verdict on the selected row, and the selected commit and row. A split
// line is marked. It reads only the tables and the stored outputs.
func PRBody(root string, g *tables.Goal) string {
	if g.Select == nil {
		return ""
	}
	tree := g.Select.RowTree
	split := g.SplitLines(tree)
	live := map[string]tables.LiveCell{}
	for _, c := range g.LiveCells(tree) {
		if _, ok := live[c.Key]; !ok {
			live[c.Key] = c
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: %s\n", g.ID, g.Words)
	for _, l := range g.Lines {
		mark := ""
		if contains(split, l.Key) {
			mark = " [split line: its examples give different verdicts on this row]"
		}
		fmt.Fprintf(&b, "\n## %s (%s): %s%s\n", l.Key, l.Kind, l.Text, mark)
		for _, p := range g.Placed() {
			if p.Line != l.Key {
				continue
			}
			words := exampleWords(root, p.Hash)
			fmt.Fprintf(&b, "\n### %s: %s\n\n", p.Key, words)
			c, ok := live[p.Key]
			if !ok {
				b.WriteString("verdict: not run\n")
				continue
			}
			fmt.Fprintf(&b, "verdict: %s, exit %d\n", c.Verdict, c.Exit)
			writeStored(&b, root, "output", c.Output)
			writeStored(&b, root, "stderr", c.Stderr)
		}
	}
	fmt.Fprintf(&b, "\nselected commit: %s\nselected row: %s\n", g.Select.Commit, tree)
	return b.String()
}

// writeStored prints a stored output under a label: inline when empty, in a
// fenced block otherwise.
func writeStored(b *strings.Builder, root, label, hash string) {
	text := readOutputMax(root, hash, prOutputMax)
	if text == "" {
		fmt.Fprintf(b, "\n%s: (empty)\n", label)
		return
	}
	fmt.Fprintf(b, "\n%s:\n\n%s", label, fence(text))
}

// exampleWords is the words section of the example folder with tree hash
// hash, or a note when this clone cannot read it.
func exampleWords(root, hash string) string {
	ex, _, err := readExample(root, hash)
	if err != nil {
		return "(example " + hash + " is not readable here)"
	}
	return ex.Words
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
