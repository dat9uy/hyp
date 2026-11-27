package channel

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/dat9uy/hyp"
	"github.com/dat9uy/hyp/internal/acts"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// BlockOpts selects what StateBlock prints.
type BlockOpts struct {
	Goal   string // a Goal id or goal.add event id; empty for every Goal
	Role   string // filter the allowed acts to one role; empty for all four
	NoBody bool   // leave the PR body out: the report above already printed it
	Ask    bool   // a written or open Goal prints what it asks: hyp status G
	All    bool   // with no Goal, list ended Goals too: hyp status --all
}

// roles are the roles a block lists, in the order of the roles in hyp help.
var roles = []string{"owner", "grill", "probe-writer", "runner"}

// maxCell is how many runes of one output's first line the outputs table
// shows.
const maxCell = 60

// StateBlock is the state block: for each Goal its state, runs
// used, stop with the act that caused it and what it waits for, the allowed acts of every role as ready
// command lines, a marked split line, the Results Stable refused, and, while
// the Goal is selected, the PR body; at a stop with no select, the table of
// outputs. A selected or ended Goal prints its PR body only when o names it:
// the list of every Goal prints its state block alone. Events refused at replay close it,
// then the lines of the events files that replay could not read: such a line
// names no Goal, so every block lists all of them. The list of every Goal
// leaves out the blocks of ended Goals unless o.All is set, and ends with a
// line that counts them; the events of an ended Goal that replay refused are
// still listed. repo is the git top level.
func StateBlock(t tables.Tables, unread []logstore.LineError, repo string, o BlockOpts) string {
	var b strings.Builder
	shown := map[string]bool{}
	inWork, hidden := 0, 0
	for _, g := range tables.SortedGoals(t) {
		if o.Goal != "" && g.EventID != o.Goal && g.ID != o.Goal {
			continue
		}
		if o.Goal == "" && !o.All && g.State == tables.StateEnded {
			hidden++
			continue
		}
		if g.State != tables.StateEnded {
			inWork++
		}
		shown[g.EventID] = true
		goalBlock(&b, t, g, repo, o)
	}
	var refused []tables.Refused
	for _, r := range t.RefusedAtReplay {
		if o.Goal == "" || shown[r.Node] {
			refused = append(refused, r)
		}
	}
	if len(refused) > 0 {
		b.WriteString("refused at replay:\n")
		for _, r := range refused {
			fmt.Fprintf(&b, "  %s: %s\n", r.EventID, bodyText(r.Err, repo))
		}
	}
	if len(unread) > 0 {
		b.WriteString("not read at replay:\n")
		for _, l := range unread {
			where := filepath.Join(".tree", "events", l.File)
			if l.Trace {
				where = filepath.Join("trace", "events", l.File)
			}
			if l.Line > 0 {
				where = fmt.Sprintf("%s line %d", where, l.Line)
			}
			fmt.Fprintf(&b, "  %s: %s\n", where, bodyText(l.Err, repo))
		}
	}
	if hidden > 0 {
		if inWork == 0 {
			b.WriteString("no Goal is in work.\n")
		}
		fmt.Fprintf(&b, "%s not shown; hyp status --all lists every Goal.\n", endedCount(hidden))
	}
	if o.Goal == "" {
		fmt.Fprintf(&b, "%d not folded: hyp finding list --not-folded\n", notFolded(t))
	}
	return b.String()
}

// endedCount is "1 ended Goal" or "N ended Goals".
func endedCount(n int) string {
	if n == 1 {
		return "1 ended Goal"
	}
	return fmt.Sprintf("%d ended Goals", n)
}

func goalBlock(b *strings.Builder, t tables.Tables, g *tables.Goal, repo string, o BlockOpts) {
	st := tables.Stand(g)
	view := acts.ViewOf(repo, g)
	ended := g.State == tables.StateEnded
	// an ended Goal prints no runs count: the runs are in each clone's trace,
	// and an ended Goal reads the same in every clone
	if ended {
		fmt.Fprintf(b, "%s  %s\n", g.ID, g.State)
	} else {
		fmt.Fprintf(b, "%s  %s  runs: %d of %d used\n", g.ID, g.State, st.Used, tables.MaxHypotheses)
	}
	if !ended && st.Stop == tables.StopNone && st.Used == tables.MaxHypotheses-1 {
		fmt.Fprintf(b, "last run: after this, %s stops for the owner.\n", g.ID)
	}
	fmt.Fprintf(b, "stop: %s\n", stopText(st.Stop, st.Cause))
	if g.Land != nil {
		fmt.Fprintf(b, "landed: by %s, land %s at %s\n", g.Land.Actor, g.Land.ID, g.Land.Time)
		endWhy(b, g.Land.Why, o)
	}
	if g.Close != nil {
		fmt.Fprintf(b, "closed: by %s, close %s at %s\n", g.Close.Actor, g.Close.ID, g.Close.Time)
		endWhy(b, g.Close.Why, o)
	}
	fmt.Fprintf(b, "waits: %s\n", st.Waits)
	if view.OnNoRef() {
		fmt.Fprintf(b, "warning: the selected commit %s is on no branch or tag of this clone; push it or keep a branch on it, or git may lose it\n", g.Select.Commit)
	}
	for _, row := range st.Rows {
		if ended {
			break
		}
		for _, line := range g.SplitLines(row.RowTree) {
			fmt.Fprintf(b, "split: %s on row %s\n", line, row.RowTree)
		}
	}
	ready := acts.AllowedFor(g)
	for _, role := range roles {
		if o.Role != "" && o.Role != role {
			continue
		}
		fmt.Fprintf(b, "%s may:\n", role)
		if role == "probe-writer" && o.Ask && (g.State == tables.StateWritten || g.State == tables.StateOpen) {
			probeWriter(b, g)
			continue
		}
		n := 0
		for _, l := range ready {
			if l.Role == role {
				fmt.Fprintf(b, "  %s\n", readyText(l))
				n++
			}
		}
		if n == 0 {
			b.WriteString("  (none)\n")
		}
	}
	heldBack(b, g, repo)
	if !ended {
		unstable(b, g)
	}
	if o.Goal == "" && (g.State == tables.StateWritten || g.State == tables.StateOpen) {
		// the list of every Goal names what each Goal in work asks, so an
		// agent finds the Goal a finding bears on
		fmt.Fprintf(b, "words: %s\n", indentText(g.Words, "  "))
	}
	if o.Goal != "" {
		findings(b, t, g)
		walks(b, g, view, repo)
	}
	if o.Ask && g.State == tables.StateOpen && (o.Role == "" || o.Role == "owner") {
		fmt.Fprintf(b, "\nThe runner's Start prompt, ready to copy:\n\n%s", fence(startPrompt(g.ID)))
	}
	if o.Ask && (g.State == tables.StateWritten || g.State == tables.StateOpen) {
		fmt.Fprintf(b, "\nThe Goal asks:\n\n%s", askText(repo, g))
	}
	if o.Ask && ended {
		fmt.Fprintf(b, "\nThe Goal asked:\n\n%s", askedText(g))
	}
	switch {
	case o.Goal == "" && (g.State == tables.StateEnded || st.Stop == tables.StopSelected):
		// the list of every Goal prints no PR body: an output it quotes can
		// hold lines that read as Goals of this store
	case st.Stop == tables.StopSelected:
		keptChange(b, t, view, repo)
		if !o.NoBody {
			fmt.Fprintf(b, "\nPR body:\n\n%s", prBodyOf(repo, g, view))
		}
	case st.Stop == tables.StopTwoCandidates || st.Stop == tables.StopBudgetUsed:
		outputsTable(b, g, repo, st.Rows)
	}
}

// keptChange lists what the selected commit adds or changes in the kept
// model, goals/model/, against H0: each step by the command it takes, and
// each rule with the words of the Goal line its source names as G:line-N,
// or "no line" when it names none. It prints no TLA+ text, and nothing when
// the commit leaves the kept model as H0 holds it.
func keptChange(b *strings.Builder, t tables.Tables, view *acts.GoalView, repo string) {
	items, err := view.Kept()
	if err != nil {
		fmt.Fprintf(b, "\nkept model: not readable here: %s\n", bodyText(err, repo))
		return
	}
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "\nkept model: what the selected commit adds or changes in %s against H0:\n", acts.KeptDir)
	for _, it := range items {
		if !it.Rule {
			fmt.Fprintf(b, "  step: %s\n", it.Command)
			continue
		}
		fmt.Fprintf(b, "  rule %s: %s\n", it.Name, sourceText(t, it.Source))
	}
}

// sourceLine is a rule's source that names a Goal line: G-XXXXXX:line-N.
var sourceLine = regexp.MustCompile(`^(G-[0-9A-Z]+):(line-[0-9]+)$`)

// sourceText is the Goal line a kept rule's source names, as
// "G:line-N: words", or "no line" when the source names none.
func sourceText(t tables.Tables, source string) string {
	m := sourceLine.FindStringSubmatch(source)
	if m == nil {
		return "no line"
	}
	if lg, err := t.Find(m[1]); err == nil {
		for _, l := range lg.Lines {
			if l.Key == m[2] {
				return source + ": " + indentText(l.Text, "    ")
			}
		}
	}
	return source + ": not a line of this store"
}

// probeWriter is the probe-writer block of a written or open Goal in
// hyp status G: the Probe writer's contract, the generic Probe and one task
// for each line of the Goal. The Probe writer writes no event, so the block
// holds no write act. Every line is indented, so the block ends where the
// next role starts.
func probeWriter(b *strings.Builder, g *tables.Goal) {
	b.WriteString("  the Probe writer's contract, for the Probe of one example:\n")
	b.WriteString("    1. Read the line, its example and H0.\n")
	fmt.Fprintf(b, "    2. Write and commit only that example's probe folder, goals/%s/<line>/<example>/probe.\n", g.ID)
	b.WriteString("    3. Start from the generic Probe and write new code only when it cannot run the example.\n")
	b.WriteString("    4. Run the example on H0 with hyp try and return its output.\n")
	b.WriteString("    5. Write no event.\n")
	b.WriteString("  the generic Probe, probe/run:\n")
	for _, l := range strings.Split(strings.TrimSuffix(hyp.GenericProbe, "\n"), "\n") {
		if l == "" {
			b.WriteString("\n")
			continue
		}
		fmt.Fprintf(b, "    %s\n", l)
	}
	b.WriteString("  tasks, one for each line:\n")
	for _, l := range g.Lines {
		fmt.Fprintf(b, "    %s (%s): the Probe of each example of %s; check it with: hyp try %s %s/<example>\n", l.Key, l.Kind, l.Key, g.ID, l.Key)
	}
}

// findings lists every finding g names: the ids its record in git names, on
// any clone, then those an old fold in this clone's trace folded into it.
// Each prints its state, its evidence and its fact when this clone's trace
// holds it, and says so when it does not. Only a block that names its Goal
// prints them: a fact is an agent's text, which can hold lines that read as
// Goals of this store. Each line of a fact is indented, so the list ends
// where the next part of the block starts.
func findings(b *strings.Builder, t tables.Tables, g *tables.Goal) {
	ids := g.Findings()
	for _, f := range tables.SortedFindings(t) {
		if f.NamedBy(g) && !slices.Contains(ids, f.EventID) {
			ids = append(ids, f.EventID)
		}
	}
	if len(ids) == 0 {
		return
	}
	b.WriteString("findings:\n")
	for _, id := range ids {
		f, err := t.FindFinding(id)
		if err != nil {
			fmt.Fprintf(b, "  %s  not in this clone's trace\n", tables.FindingID(id))
			continue
		}
		fmt.Fprintf(b, "  %s  %s  by %s at %s\n", f.ID, findingStateText(f), f.Actor, f.Time)
		fmt.Fprintf(b, "    evidence: %s\n", indentText(f.Evidence, "      "))
		fmt.Fprintf(b, "    fact: %s\n", indentText(f.Fact, "      "))
	}
}

// walks prints, for each line of g, each command of its placed examples'
// run sections with the step of the kept model it takes, or no step: on H0,
// or on the selected commit while g has a select. It prints nothing when
// that commit keeps no model in goals/model/.
func walks(b *strings.Builder, g *tables.Goal, view *acts.GoalView, repo string) {
	commit, selected := view.WalkCommit()
	on := "H0"
	if selected {
		on = "the selected commit"
	}
	kept, err := view.Keeps()
	if err != nil {
		fmt.Fprintf(b, "walks: not readable here: %s\n", bodyText(err, repo))
		return
	}
	placed := g.Placed()
	if !kept || len(placed) == 0 {
		return
	}
	fmt.Fprintf(b, "walks on %s, %s: each command of each placed example's run section with the step of the kept model it takes:\n", on, commit)
	for _, l := range g.Lines {
		first := true
		for _, p := range placed {
			if p.Line != l.Key {
				continue
			}
			if first {
				fmt.Fprintf(b, "  %s\n", l.Key)
				first = false
			}
			fmt.Fprintf(b, "    %s\n", p.Key)
			w, err := view.Walk(p.Hash)
			if err != nil {
				fmt.Fprintf(b, "      not readable here: %s\n", bodyText(err, repo))
				continue
			}
			walkSteps(b, w, "      ")
		}
	}
}

// heldBack names, for each candidate whose select hyp holds back, the guard,
// its code and the words with the fix that hyp select would print for it. It
// prints for every role: hyp status lists no select line for such a
// candidate, and the line it lists instead does not say why.
func heldBack(b *strings.Builder, g *tables.Goal, repo string) {
	holds := tables.HeldBack(g)
	if len(holds) == 0 {
		return
	}
	b.WriteString("held back: hyp select refuses each candidate below, and says why:\n")
	for _, h := range holds {
		fmt.Fprintf(b, "  row %s: %s\n", h.Row.RowTree, holdText(h.Err, repo))
	}
}

// holdText is the refusal of a held-back select as hyp select prints it, with
// its code after the guard: "Spec (no-spec): commit ... ; commit the spec, ...".
func holdText(err error, repo string) string {
	text := bodyText(err, repo)
	e, ok := fault.As(err)
	if !ok {
		return text
	}
	if e.Guard == fault.NoGuard {
		return "(" + string(e.Code) + "): " + text
	}
	prefix := string(e.Guard) + ": "
	return string(e.Guard) + " (" + string(e.Code) + "): " + strings.TrimPrefix(text, prefix)
}

// indentText puts indent before every line of s after the first.
func indentText(s, indent string) string {
	return strings.ReplaceAll(strings.TrimRight(s, "\n"), "\n", "\n"+indent)
}

// unstable lists every Result Stable refused: the example, the tree and the
// two output hashes.
func unstable(b *strings.Builder, g *tables.Goal) {
	first := true
	for _, r := range g.Results {
		for _, u := range r.Unstable {
			if first {
				b.WriteString("unstable:\n")
				first = false
			}
			fmt.Fprintf(b, "  %s on %s: %s then %s\n", u.Key, u.Tree, orUnknown(u.First), u.Second)
		}
	}
}

func orUnknown(hash string) string {
	if hash == "" {
		return "(unknown)"
	}
	return hash
}

// outputsTable prints each row's output for each placed example: the table
// of outputs of a stop, with no reading of it. A cell shows the output's
// size, its hash and the start of its first line as the bytes are; the whole
// output is the file of that hash.
func outputsTable(b *strings.Builder, g *tables.Goal, repo string, rows []tables.Row) {
	fmt.Fprintf(b, "outputs: each cell shows at most %d characters of the output's first line; all its bytes are in this clone's trace, outputs/<hash>\n", maxCell)
	for _, row := range rows {
		fmt.Fprintf(b, "  row %s\n", row.RowTree)
		cells := g.LiveCells(row.RowTree)
		for _, p := range g.Placed() {
			i := slices.IndexFunc(cells, func(c tables.LiveCell) bool { return c.Key == p.Key })
			if i < 0 {
				fmt.Fprintf(b, "    %s  %s\n", p.Key, tables.NotRun)
				continue
			}
			c := cells[i]
			if !logstore.HasOutput(repo, c.Output) {
				fmt.Fprintf(b, "    %s  %s  exit %d  output not in this clone's trace\n", c.Key, c.Verdict, c.Exit)
				continue
			}
			fmt.Fprintf(b, "    %s  %s  exit %d  %s  %s\n", c.Key, c.Verdict, c.Exit, c.Output, cellText(repo, c.Output))
		}
	}
}

// cellText is one cell of the outputs table: the output's size, then the
// start of its first line, at most maxCell characters, as the bytes are. It
// never quotes or escapes them, and never adds a mark to them: the size says
// whether the cell shows all of it.
func cellText(repo, hash string) string {
	data, size, err := logstore.ReadOutput(repo, hash, 4*maxCell)
	if err != nil {
		return "not readable here: " + bodyText(err, repo)
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		data = data[:i]
	}
	for n, i := 0, 0; i < len(data); n++ {
		if n == maxCell {
			data = data[:i]
			break
		}
		_, w := utf8.DecodeRune(data[i:])
		i += w
	}
	return fmt.Sprintf("%d bytes: %s", size, data)
}

// h0Cell says what this clone's trace holds of H0's cell of a placement:
// nothing to add when it holds the outputs the place measured; a cell a try ran again, with a warning when its verdict differs
// from the verdict git holds; or that the trace lacks it, with the ready try
// line that runs it again.
func h0Cell(b *strings.Builder, g *tables.Goal, p tables.Placement) {
	if p.H0.RowTree != g.BaseTree {
		return
	}
	tc, ok := g.TraceCell(p.Hash, g.BaseTree)
	switch {
	case !ok:
		fmt.Fprintf(b, "H0's cell is not in this clone's trace; run it again in this clone with:\n  %s\n", tryLine(g, p.Key, ""))
	case tc.By == logstore.CellByTry:
		fmt.Fprintf(b, "verdict on H0 in this clone: %s, exit %d, run again in this clone\n", tc.Verdict, tc.Exit)
		if tc.Verdict != p.H0.Verdict {
			fmt.Fprintf(b, "warning: %s gave two verdicts on H0: at place: %s, in this clone: %s\n", p.Key, p.H0.Verdict, tc.Verdict)
		}
	}
}

// askText is what a written or open Goal asks: its words, each line's key,
// kind and text, and under each line every placed example with its version,
// hash, verdict on H0 at place and its example.txtar byte for byte, which
// holds its words, run, floors and ceilings.
func askText(repo string, g *tables.Goal) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: %s\n", g.ID, g.Words)
	if g.Why != "" {
		fmt.Fprintf(&b, "\nwhy: %s\n", g.Why)
	}
	placed := g.Placed()
	for _, l := range g.Lines {
		fmt.Fprintf(&b, "\n## %s (%s): %s\n", l.Key, l.Kind, l.Text)
		n := 0
		for _, p := range placed {
			if p.Line != l.Key {
				continue
			}
			n++
			fmt.Fprintf(&b, "\n### %s @%d\n\n", p.Key, p.Version)
			fmt.Fprintf(&b, "hash: %s\nplaced: by %s at %s\n", p.Hash, p.Actor, p.Time)
			fmt.Fprintf(&b, "verdict on H0 at place: %s, exit %d\n", p.H0.Verdict, p.H0.Exit)
			h0Cell(&b, g, p)
			data, err := acts.ExampleFile(repo, p.Hash)
			if err != nil {
				fmt.Fprintf(&b, "\nexample.txtar: not readable here: %s\n", bodyText(err, repo))
				continue
			}
			fmt.Fprintf(&b, "\nexample.txtar, %d bytes:\n\n%s", len(data), fence(string(data)))
		}
		if n == 0 {
			b.WriteString("\nno example placed\n")
		}
	}
	return b.String()
}

// askedText is what an ended Goal asked: its words, the why of its latest
// goal.add, and each line with its kind. hyp status G prints it, so the want a
// close gave up stays readable; land and close do not print it.
func askedText(g *tables.Goal) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: %s\n", g.ID, g.Words)
	if g.Why != "" {
		fmt.Fprintf(&b, "\nwhy: %s\n", g.Why)
	}
	for _, l := range g.Lines {
		fmt.Fprintf(&b, "\n## %s (%s): %s\n", l.Key, l.Kind, l.Text)
	}
	return b.String()
}

// endWhy prints the why of a land or close under its line, in hyp status G
// only: the return of land and close prints the block without it.
func endWhy(b *strings.Builder, why string, o BlockOpts) {
	if o.Ask && why != "" {
		fmt.Fprintf(b, "  why: %s\n", indentText(why, "    "))
	}
}

// startPrompt is the prompt the owner gives a runner of the open Goal id:
// the Goal, and the actor runner- with the Goal id in lower case, filled in.
// It names no issue: the owner, not the runner, closes the issues a Goal
// settles.
func startPrompt(id string) string {
	actor := "runner-" + strings.ToLower(id)
	return fmt.Sprintf("Run the hyp Goal %[1]s as its runner.\n"+
		"- Your actor is %[2]s. Set HYP_ACTOR=%[2]s and HYP_ROLE=runner each\n"+
		"  time you call hyp.\n"+
		"- Work in your own worktree, on a new branch from the default branch.\n"+
		"- Follow the run rules of the hyp-grill skill (run-rules.md) and this\n"+
		"  repository's agent files, such as AGENTS.md.\n"+
		"- The Goal's placed examples are the whole task: start from what\n"+
		"  `hyp status %[1]s --role runner` prints. No issue, PR or hand-off is a\n"+
		"  source of work.\n"+
		"- Open a PR whose body is the select's report. Do not merge the PR: the\n"+
		"  owner merges.\n", id, actor)
}
