package channel

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/acts"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/spec"
	"github.com/dat9uy/hyp/internal/tables"
)

// The reports of the write acts. Each prints the values the act returned as
// hyp holds them: a Goal's words, a line's text and an example's words are
// never quoted or escaped.

func addGoalText(rep acts.Report) string {
	var b strings.Builder
	verb := "recorded"
	if rep.Replaced {
		verb = "replaced"
	}
	fmt.Fprintf(&b, "%s %s\n", verb, rep.Goal)
	fmt.Fprintf(&b, "words: %s\n", rep.Words)
	fmt.Fprintf(&b, "H0: %s\n", rep.Base)
	for _, l := range rep.Lines {
		fmt.Fprintf(&b, "%s %s: %s\n", l.Key, l.Kind, l.Text)
	}
	if len(rep.Findings) > 0 {
		fmt.Fprintf(&b, "findings: %s\n", strings.Join(rep.Findings, " "))
	}
	for _, p := range rep.Placed {
		fmt.Fprintf(&b, "%s @%d  hash: %s  H0: %s (exit %d)\n", p.Key, p.Version, p.Hash, p.Verdict, p.Exit)
	}
	if len(rep.Placed) > 0 {
		fmt.Fprintf(&b, "%s is %s\n", rep.Goal, rep.State)
	}
	if rep.Folder != "" {
		fmt.Fprintf(&b, "copied the examples to %s/; commit them with the store\n", rep.Folder)
	}
	return b.String()
}

// placeText is the report of a place: each example the place changed, with
// its version, hash and verdict on H0, then the Goal's state.
func placeText(rep acts.PlaceReport) string {
	var b strings.Builder
	n := len(rep.Placed)
	plural := "s"
	if n == 1 {
		plural = ""
	}
	fmt.Fprintf(&b, "placed %s: %d example%s\n", rep.Goal, n, plural)
	for _, p := range rep.Placed {
		fmt.Fprintf(&b, "%s @%d  hash: %s  H0: %s (exit %d)\n", p.Key, p.Version, p.Hash, p.Verdict, p.Exit)
	}
	fmt.Fprintf(&b, "%s is %s\n", rep.Goal, rep.State)
	return b.String()
}

// runText is the report of a run: the commit and row, and for each example
// its verdict, exit code and output hash, then its words as the example
// holds them.
func runText(rep acts.RunReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ran %s %s at %s\n", rep.Goal, rep.Line, rep.Commit)
	fmt.Fprintf(&b, "row: %s\n", rep.RowTree)
	for _, c := range rep.Cells {
		fmt.Fprintf(&b, "%s %s (exit %d, output %s)\n", c.Key, c.Verdict, c.Exit, c.Output)
		fmt.Fprintf(&b, "  words: %s\n", c.Words)
	}
	return b.String()
}

// tryText is the report of try: where the example ran, what it saved, the
// example folder as it is on disk, then its verdict, exit code, output and
// stderr as the cell place or run would record holds them, each output under
// the name it has in the trace's outputs.
func tryText(rep acts.TryReport) string {
	var b strings.Builder
	on := "H0 " + rep.Commit
	switch {
	case rep.Worktree:
		on = "this checkout's files: HEAD " + rep.Commit + " and every change not committed"
	case !rep.H0:
		on = "commit " + rep.Commit + ", row " + rep.RowTree
	}
	fmt.Fprintf(&b, "tried %s %s on %s\n", rep.ID, rep.Key, on)
	if rep.Saved {
		b.WriteString("saved: this clone's trace lacked this cell; it holds it now, marked as run again in this clone; no event and no run of the budget is recorded\n")
	} else {
		b.WriteString("nothing is recorded: no event, no output file, no run of the budget\n")
	}
	fmt.Fprintf(&b, "example: %s as it is on disk, hash %s, probe %s\n", rep.Folder, rep.Hash, or(rep.ProbeHash, "(none)"))
	fmt.Fprintf(&b, "words: %s\n", rep.Words)
	fmt.Fprintf(&b, "verdict: %s, exit %d\n", rep.Verdict, rep.Out.Exit)
	writeBlock(&b, "output "+rep.OutHash, rep.Out.Stdout, int64(len(rep.Out.Stdout)))
	writeBlock(&b, "stderr "+rep.ErrHash, rep.Out.Stderr, int64(len(rep.Out.Stderr)))
	return b.String()
}

// or is v, or placeholder when v is empty.
func or(v, placeholder string) string {
	if v == "" {
		return placeholder
	}
	return v
}

// landText is the report of a land: the Goal, and the stop it ended with the
// act that caused it.
func landText(rep acts.LandReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "landed %s\nstop: %s\n", rep.Goal, stopText(rep.Stop, rep.Cause))
	if rep.Commit != "" {
		fmt.Fprintf(&b, "commit: %s  tree: %s  row: %s\n", rep.Commit, rep.Tree, rep.RowTree)
		for _, v := range rep.Verdicts {
			ran := ""
			if slices.Contains(rep.Ran, v.Line+"/"+v.Example) {
				ran = "  (run by the land: this clone held no Result)"
			}
			fmt.Fprintf(&b, "%s/%s  hash: %s  %s%s\n", v.Line, v.Example, v.Hash, v.Verdict, ran)
		}
	}
	return b.String()
}

// findingText is the report of an act on a finding: each finding it wrote
// about, with its state after the act.
func findingText(rep acts.FindingReport) string {
	var b strings.Builder
	for _, f := range rep.Findings {
		fmt.Fprintf(&b, "recorded %s  %s\n", f.ID, findingStateText(f))
	}
	return b.String()
}

// findingStateText is a finding's state as hyp prints it: not folded,
// folded into G, settled in G, ended in G, the old form's folded into #N,
// refuted or declined.
func findingStateText(f *tables.Finding) string {
	switch f.State() {
	case spec.FindingFolded:
		if g := f.InWork(); g != nil {
			return "folded into " + g.ID
		}
		return fmt.Sprintf("folded into #%d", f.Issue)
	case spec.FindingSettled:
		return "settled in " + f.SettledBy().ID
	case spec.FindingEnded:
		return "ended in " + f.Names[len(f.Names)-1].ID + ", not settled"
	}
	return f.State()
}

// closeText is the report of a close: the Goal, now ended.
func closeText(rep acts.CloseReport) string {
	return fmt.Sprintf("closed %s\n", rep.Goal)
}

// unselectText is the report of an unselect: the Goal, open again, and the
// select it undid.
func unselectText(rep acts.UnselectReport) string {
	return fmt.Sprintf("unselected %s\nundid: select %s at %s, commit %s, row %s\n",
		rep.Goal, rep.Select.ID, rep.Select.Time, rep.Select.Commit, rep.Select.RowTree)
}

// stopText is a stop as an event: which stop, the act that caused it, its
// event id and when. The stop none has no cause.
func stopText(stop string, c tables.Cause) string {
	if c.ID == "" {
		return stop
	}
	if c.Act == "" {
		return fmt.Sprintf("%s, by %s", stop, c.ID)
	}
	return fmt.Sprintf("%s, by %s %s at %s", stop, c.Act, c.ID, c.Time)
}

// readyText is an allowed act as the command line its role runs. A key hyp
// cannot know prints as a placeholder in angle brackets.
func readyText(r acts.Ready) string {
	switch r.Act {
	case acts.ActAddGoal:
		return "hyp add goal " + r.Goal + " --base <commit> --words <words> --line <kind>:<text>"
	case acts.ActPlace:
		return "hyp place " + r.Goal
	case acts.ActRun:
		return "hyp run " + or(shortCommit(r.Commit), "<commit>") + " " + r.Goal + "/" + r.Line
	case acts.ActSelect:
		return "hyp select " + r.Goal + " " + r.Tree
	case acts.ActLand:
		if r.Commit != "" {
			return "hyp land " + r.Goal + " " + shortCommit(r.Commit)
		}
		return "hyp land " + r.Goal
	case acts.ActUnselect:
		if r.Commit != "" {
			return "hyp unselect " + r.Goal + " " + shortCommit(r.Commit)
		}
		return "hyp unselect " + r.Goal
	case acts.ActClose:
		return "hyp close " + r.Goal
	}
	return "hyp " + r.Act + " " + r.Goal
}

// shortCommit is the form of a commit in a ready line.
func shortCommit(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

var backtickRun = regexp.MustCompile("`+")

// fence wraps text in a code fence longer than any run of backticks in it.
// The fence holds the bytes as they are; the label above it gives their
// count, so a final newline the fence cannot show is not lost.
func fence(text string) string {
	n := 3
	for _, m := range backtickRun.FindAllString(text, -1) {
		n = max(n, len(m)+1)
	}
	bar := strings.Repeat("`", n)
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return bar + "\n" + text + bar + "\n"
}

// prBody is the PR body of a selected Goal: its words, each line with its
// kind, each placed example with its key and words, each chart of the
// selected commit's spec folder, and the selected commit and row. A split
// line is marked. It holds no output, stderr, exit code or verdict of any
// example: every placed example is met by the select, so a result teaches
// the owner nothing at the merge. It reads only the tables and git, so it
// reads the same in every clone.
func prBody(root string, g *tables.Goal) string {
	if g == nil || g.Select == nil {
		return ""
	}
	return prBodyOf(root, g, acts.ViewOf(root, g))
}

// prBodyOf is the PR body of g, a selected Goal, from view, the reads of
// its commit that the rest of its status block shares.
func prBodyOf(root string, g *tables.Goal, view *acts.GoalView) string {
	tree := g.Select.RowTree
	var split []string
	if g.State != tables.StateEnded {
		split = g.SplitLines(tree)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: %s\n", g.ID, g.Words)
	for _, l := range g.Lines {
		mark := ""
		for _, s := range split {
			if s == l.Key {
				mark = " [split line: its examples give different verdicts on this row]"
			}
		}
		fmt.Fprintf(&b, "\n## %s (%s): %s%s\n", l.Key, l.Kind, l.Text, mark)
		for _, p := range g.Placed() {
			if p.Line != l.Key {
				continue
			}
			words, err := view.ExampleWords(p.Hash)
			if err != nil {
				words = "(example " + p.Hash + " is not readable here: " + bodyText(err, root) + ")"
			}
			fmt.Fprintf(&b, "\n### %s: %s\n", p.Key, words)
		}
	}
	prCharts(&b, root, g, view)
	fmt.Fprintf(&b, "\nselected commit: %s\nselected row: %s\n", g.Select.Commit, tree)
	return b.String()
}

// tryLine is the ready hyp try line that runs the cell of key again in this
// clone: on H0 when commit is empty, else on commit.
func tryLine(g *tables.Goal, key, commit string) string {
	if commit == "" {
		return "hyp try " + g.ID + " " + key
	}
	return "hyp try " + g.ID + " " + key + " --commit " + commit
}

// prCharts prints each chart (.mmd) of the selected commit's spec folder as
// a mermaid block, so the owner reads the model's chart in the PR: the
// charts the run or the land of that commit read (acts.GoalView.Charts). A
// spec folder with no chart prints nothing.
func prCharts(b *strings.Builder, root string, g *tables.Goal, view *acts.GoalView) {
	charts, err := view.Charts()
	if err != nil {
		fmt.Fprintf(b, "\n## Chart\n\nnot readable here: %s\n", bodyText(err, root))
		return
	}
	for _, c := range charts {
		fmt.Fprintf(b, "\n## Chart: %s%s\n\n", tables.SpecDir(g.ID), c.Path)
		f := fence(c.Text)
		b.WriteString(strings.Replace(f, "\n", "mermaid\n", 1))
	}
}

// pruneText is the report of hyp prune: the traces folder, then one line per
// trace folder of a clone that no longer exists, its key and the clone its
// file clone names. With removed, the folders listed are those hyp removed.
func pruneText(base string, gone []logstore.Gone, removed bool) string {
	var b strings.Builder
	switch {
	case len(gone) == 0 && removed:
		fmt.Fprintf(&b, "removed nothing: no trace under %s is of a repository that no longer exists\n", base)
	case len(gone) == 0:
		fmt.Fprintf(&b, "no trace under %s is of a repository that no longer exists\n", base)
	case removed && len(gone) == 1:
		fmt.Fprintf(&b, "removed 1 trace under %s of a repository that no longer exists:\n", base)
	case removed:
		fmt.Fprintf(&b, "removed %d traces under %s of repositories that no longer exist:\n", len(gone), base)
	case len(gone) == 1:
		fmt.Fprintf(&b, "1 trace under %s is of a repository that no longer exists:\n", base)
	default:
		fmt.Fprintf(&b, "%d traces under %s are of repositories that no longer exist:\n", len(gone), base)
	}
	for _, g := range gone {
		fmt.Fprintf(&b, "  %s  %s\n", g.Key, g.Clone)
	}
	if len(gone) > 0 && !removed {
		b.WriteString("hyp prune --remove removes them; it removes nothing else\n")
	}
	return b.String()
}
