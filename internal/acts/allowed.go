package acts

import (
	"fmt"
	"slices"

	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/tables"
)

// The stops of a Goal.
const (
	StopNone          = "none"
	StopSelected      = "selected"
	StopTwoCandidates = "two candidates"
	StopBudgetUsed    = "budget used"
)

// ReadyLine is one allowed act as a command line a role may run. A key hyp
// cannot know is a placeholder in angle brackets.
type ReadyLine struct {
	Goal string
	Role string
	Line string
}

// Standing is what hyp calculates about a Goal from the tables and git.
type Standing struct {
	Stop       string
	Waits      string // who the Goal waits for: grill, owner or none
	Used       int    // N of "runs: N of 3 used"
	Rows       []tables.Row
	Candidates []tables.Row
}

// Stand calculates the standing of a Goal: its rows and candidates, and the
// stop of section 11. repo is the git top level.
func Stand(g *tables.Goal, repo string) Standing {
	anc := gitread.Ancestry{Dir: repo}
	st := Standing{Stop: StopNone, Waits: "none", Used: tables.RunsUsed(g), Rows: g.Rows(anc)}
	st.Candidates = tables.Candidates(g, anc)
	switch {
	case g.Select != nil:
		st.Stop = StopSelected
	case len(st.Candidates) >= 2:
		st.Stop, st.Waits = StopTwoCandidates, "owner"
	case len(st.Candidates) == 0 && budgetUsed(g):
		st.Stop, st.Waits = StopBudgetUsed, "owner"
	case g.State == tables.StateWritten:
		st.Waits = "grill"
	}
	return st
}

// budgetUsed reports the stop of section 11: three trees counted, the third
// with a live cell for every placed example. The caller checks no candidate.
func budgetUsed(g *tables.Goal) bool {
	counted := g.CountedTrees()
	if len(counted) < tables.MaxTrees {
		return false
	}
	third := counted[tables.MaxTrees-1]
	for _, p := range g.Placed() {
		if g.Verdict(p.Key, third) == tables.NotRun {
			return false
		}
	}
	return true
}

// short is the form of a commit in a ready line.
func short(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

// Allowed lists the allowed acts of every Goal, in the order of the Goals'
// events, as ready command lines.
func Allowed(t tables.Tables, repo string) []ReadyLine {
	var out []ReadyLine
	for _, g := range tables.SortedGoals(t) {
		out = append(out, AllowedFor(g, repo)...)
	}
	return out
}

// AllowedFor lists the allowed acts of one Goal. It checks every guard that
// reads only the tables and git; Kind and Stable need a Probe run and are
// checked when the act runs. Acts the binary lacks are never listed. Within a
// role the lines keep the order of the Goal's lines.
func AllowedFor(g *tables.Goal, repo string) []ReadyLine {
	if g.Select != nil {
		return nil
	}
	st := Stand(g, repo)
	var out []ReadyLine
	add := func(role, format string, a ...any) {
		out = append(out, ReadyLine{Goal: g.ID, Role: role, Line: fmt.Sprintf(format, a...)})
	}

	// owner
	if g.State == tables.StateWritten {
		add("owner", "hyp add goal %s --base <commit> --words <words> --line <kind>:<text>", g.ID)
	}
	if st.Stop == StopTwoCandidates || (st.Stop == StopNone && len(st.Candidates) == 1) {
		for _, c := range st.Candidates {
			if selectable(g, repo, c) {
				add("owner", "hyp select %s %s", g.ID, c.RowTree)
			}
		}
	}

	// grill
	placed := map[string]bool{}
	for _, p := range g.Placed() {
		placed[p.Line] = true
	}
	for _, l := range g.Lines {
		if !placed[l.Key] {
			add("grill", "hyp place %s %s/<example>", g.ID, l.Key)
		}
	}
	if !hasRole(out, "grill") {
		add("grill", "hyp place %s <line>/<example>", g.ID)
	}

	// runner
	switch {
	case st.Stop == StopNone && len(st.Candidates) == 1 && selectable(g, repo, st.Candidates[0]):
		add("runner", "hyp select %s %s", g.ID, st.Candidates[0].RowTree)
	}
	if st.Stop == StopNone {
		out = append(out, runLines(g, st)...)
	}
	return orderRoles(out)
}

func hasRole(ls []ReadyLine, role string) bool {
	return slices.ContainsFunc(ls, func(l ReadyLine) bool { return l.Role == role })
}

// roleOrder is the order roles print in: the order section 2 names them.
var roleOrder = []string{"owner", "grill", "probe-writer", "runner"}

func orderRoles(ls []ReadyLine) []ReadyLine {
	slices.SortStableFunc(ls, func(a, b ReadyLine) int {
		return slices.Index(roleOrder, a.Role) - slices.Index(roleOrder, b.Role)
	})
	return ls
}

// runLines are the runner's run lines. On a Goal that is written only a run on
// H0 is open, and only where an example's H0 cell is not run. On an open Goal
// a run on any leaf is open, so the commit is a placeholder; once three trees
// are counted, only the lines the counted leaf rows lack a cell for.
func runLines(g *tables.Goal, st Standing) []ReadyLine {
	var out []ReadyLine
	add := func(commit, line string) {
		out = append(out, ReadyLine{Goal: g.ID, Role: "runner", Line: fmt.Sprintf("hyp run %s %s/%s", commit, g.ID, line)})
	}
	if g.State == tables.StateWritten {
		for _, l := range g.Lines {
			for _, p := range g.Placed() {
				if p.Line == l.Key && g.H0Verdict(p.Key) == tables.NotRun {
					add(short(g.Base), l.Key)
					break
				}
			}
		}
		return out
	}
	if g.State != tables.StateOpen {
		return nil
	}
	counted := g.CountedTrees()
	for _, l := range g.Lines {
		if !slices.ContainsFunc(g.Placed(), func(p tables.Placement) bool { return p.Line == l.Key }) {
			continue
		}
		if len(counted) < tables.MaxTrees {
			add("<commit>", l.Key)
			continue
		}
		for _, row := range st.Rows {
			if !slices.Contains(counted, row.RowTree) || !row.Leaf || row.Unknown != "" {
				continue
			}
			for _, p := range g.Placed() {
				if p.Line == l.Key && row.Verdicts[p.Key] == tables.NotRun {
					add(short(row.Commits[len(row.Commits)-1]), l.Key)
					break
				}
			}
		}
	}
	return out
}

// selectable checks the select guards that read only git: Spec and Own path
// on the commit the select would record. A candidate that fails them is not
// listed; the select act names the guard.
func selectable(g *tables.Goal, repo string, c tables.Row) bool {
	commit := latestCommit(g, c.RowTree)
	if commit == "" || guardSpec(repo, g, commit) != nil {
		return false
	}
	changes, err := gitread.OwnChange(repo, commit, g.Base)
	return err == nil && tables.GuardOwnPath(g, changes) == nil
}
