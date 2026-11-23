package channel

import (
	"bytes"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/dat9uy/hyp"
	"github.com/dat9uy/hyp/internal/acts"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// BlockOpts selects what StateBlock prints.
type BlockOpts struct {
	Goal   string // a Goal id or goal.add event id; empty for every Goal
	Role   string // filter the allowed acts to one role; empty for all four
	NoBody bool   // leave the PR body out: the report above already printed it
	Ask    bool   // a written or open Goal prints what it asks: hyp status G
}

// roles are the roles a block lists, in the order of the state machine,
// section 2.
var roles = []string{"owner", "grill", "probe-writer", "runner"}

// maxCell is how many runes of one output's first line the outputs table
// shows.
const maxCell = 60

// StateBlock is the state block of section 12: for each Goal its state, runs
// used, stop with the act that caused it and what it waits for, the allowed acts of every role as ready
// command lines, a marked split line, the Results Stable refused, and, while
// the Goal is selected, the PR body; at a stop with no select, the table of
// outputs. A selected or ended Goal prints its PR body only when o names it:
// the list of every Goal prints its state block alone. Events refused at replay close it,
// then the lines of the events files that replay could not read: such a line
// names no Goal, so every block lists all of them. repo is the git top level.
func StateBlock(t tables.Tables, unread []logstore.LineError, repo string, o BlockOpts) string {
	var b strings.Builder
	shown := map[string]bool{}
	for _, g := range tables.SortedGoals(t) {
		if o.Goal != "" && g.EventID != o.Goal && g.ID != o.Goal {
			continue
		}
		shown[g.EventID] = true
		goalBlock(&b, g, repo, o)
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
			if l.Line > 0 {
				where = fmt.Sprintf("%s line %d", where, l.Line)
			}
			fmt.Fprintf(&b, "  %s: %s\n", where, bodyText(l.Err, repo))
		}
	}
	return b.String()
}

func goalBlock(b *strings.Builder, g *tables.Goal, repo string, o BlockOpts) {
	st := acts.Stand(g, repo)
	fmt.Fprintf(b, "%s  %s  runs: %d of %d used\n", g.ID, g.State, st.Used, tables.MaxHypotheses)
	if st.Stop == acts.StopNone && st.Used == tables.MaxHypotheses-1 {
		fmt.Fprintf(b, "last run: after this, %s stops for the owner.\n", g.ID)
	}
	fmt.Fprintf(b, "stop: %s\n", stopText(st.Stop, st.Cause))
	if g.Land != nil {
		fmt.Fprintf(b, "landed: by %s, land %s at %s\n", g.Land.Actor, g.Land.ID, g.Land.Time)
	}
	if g.Close != nil {
		fmt.Fprintf(b, "closed: by %s, close %s at %s\n", g.Close.Actor, g.Close.ID, g.Close.Time)
	}
	fmt.Fprintf(b, "waits: %s\n", st.Waits)
	for _, row := range st.Rows {
		for _, line := range g.SplitLines(row.RowTree) {
			fmt.Fprintf(b, "split: %s on row %s\n", line, row.RowTree)
		}
	}
	ready := acts.AllowedFor(g, repo)
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
	unstable(b, g)
	if o.Ask && (g.State == tables.StateWritten || g.State == tables.StateOpen) {
		fmt.Fprintf(b, "\nThe Goal asks:\n\n%s", askText(repo, g))
	}
	switch {
	case o.Goal == "" && (g.State == tables.StateEnded || st.Stop == acts.StopSelected):
		// the list of every Goal prints no PR body: an output it quotes can
		// hold lines that read as Goals of this store
	case st.Stop == acts.StopSelected:
		if !o.NoBody {
			fmt.Fprintf(b, "\nPR body:\n\n%s", prBody(repo, g))
		}
	case st.Stop == acts.StopTwoCandidates || st.Stop == acts.StopBudgetUsed:
		outputsTable(b, g, repo, st.Rows)
	}
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
	fmt.Fprintf(b, "outputs: each cell shows at most %d characters of the output's first line; all its bytes are in .tree/outputs/<hash>\n", maxCell)
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

// askText is what a written or open Goal asks: its words, each line's key,
// kind and text, and under each line every placed example with its version,
// hash, verdict on H0 at place and its example.txtar byte for byte, which
// holds its words, run, floors and ceilings.
func askText(repo string, g *tables.Goal) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: %s\n", g.ID, g.Words)
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
