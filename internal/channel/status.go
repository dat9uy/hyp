package channel

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/acts"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// BlockOpts selects what StateBlock prints.
type BlockOpts struct {
	Goal   string // a Goal id or goal.add event id; empty for every Goal
	Role   string // filter the allowed acts to one role; empty for all four
	NoBody bool   // leave the PR body out: the report above already printed it
}

// roles are the roles a block lists, in the order of the state machine,
// section 2.
var roles = []string{"owner", "grill", "probe-writer", "runner"}

// maxCell is how many runes of one output the outputs table shows.
const maxCell = 60

// StateBlock is the state block of section 12: for each Goal its state, runs
// used, stop and what it waits for, the allowed acts of every role as ready
// command lines, a marked split line, the Results Stable refused, and, while
// the Goal is selected, the PR body. Events refused at replay close it. repo
// is the git top level.
func StateBlock(t tables.Tables, repo string, o BlockOpts) string {
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
			fmt.Fprintf(&b, "  %s: %s\n", r.EventID, r.Reason)
		}
	}
	return b.String()
}

func goalBlock(b *strings.Builder, g *tables.Goal, repo string, o BlockOpts) {
	st := acts.Stand(g, repo)
	fmt.Fprintf(b, "%s  %s  runs: %d of %d used\n", g.ID, g.State, st.Used, tables.MaxTrees)
	if st.Stop == acts.StopNone && st.Used == tables.MaxTrees-1 {
		fmt.Fprintf(b, "last run: after this, %s stops for the owner.\n", g.ID)
	}
	fmt.Fprintf(b, "stop: %s\n", st.Stop)
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
		n := 0
		for _, l := range ready {
			if l.Role == role {
				fmt.Fprintf(b, "  %s\n", l.Line)
				n++
			}
		}
		if n == 0 {
			b.WriteString("  (none)\n")
		}
	}
	unstable(b, g)
	switch {
	case st.Stop == acts.StopSelected:
		if !o.NoBody {
			fmt.Fprintf(b, "\nPR body:\n\n%s", acts.PRBody(repo, g))
		}
	case st.Stop == acts.StopTwoCandidates || st.Stop == acts.StopBudgetUsed:
		outputsTable(b, g, repo, st.Rows)
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
// of outputs of a stop, with no reading of it.
func outputsTable(b *strings.Builder, g *tables.Goal, repo string, rows []tables.Row) {
	b.WriteString("outputs:\n")
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
			fmt.Fprintf(b, "    %s  %s  exit %d  %s\n", c.Key, c.Verdict, c.Exit, shown(repo, c.Output))
		}
	}
}

// shown is a stored output on one line, quoted and cut.
func shown(repo, hash string) string {
	data, err := os.ReadFile(logstore.OutputPath(repo, hash))
	if err != nil {
		return "(unreadable: " + err.Error() + ")"
	}
	runes := []rune(string(data))
	cut := ""
	if len(runes) > maxCell {
		runes, cut = runes[:maxCell], "..."
	}
	return strconv.Quote(string(runes)) + cut
}
