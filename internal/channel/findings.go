package channel

import (
	"fmt"
	"strings"

	"github.com/dat9uy/hyp/internal/spec"
	"github.com/dat9uy/hyp/internal/tables"
)

// ListOpts selects what FindingList prints.
type ListOpts struct {
	All       bool // list the findings folded into an ended Goal too
	NotFolded bool // list only the findings that are not folded
}

// FindingList is what hyp finding list prints: every finding of this
// clone's trace in merge order, each with its state, fact, evidence, the
// add goal its agent drafted, and the finding acts it allows as ready lines;
// then the line that records a new finding. A finding folded into an ended
// Goal is left out unless o.All, and a last line counts those; with
// o.NotFolded only the findings that are not folded are listed. Each line of
// an agent's text is indented, so a finding ends where the next one starts.
func FindingList(t tables.Tables, o ListOpts) string {
	var b strings.Builder
	inWork := goalsInWork(t)
	shown, hidden := 0, 0
	for _, f := range tables.SortedFindings(t) {
		if o.NotFolded && f.State() != spec.FindingNotFolded {
			continue
		}
		if !o.All && f.IntoEnded() {
			hidden++
			continue
		}
		shown++
		fmt.Fprintf(&b, "%s  %s  by %s at %s\n", f.ID, findingStateText(f), f.Actor, f.Time)
		fmt.Fprintf(&b, "  fact: %s\n", indentText(f.Fact, "    "))
		fmt.Fprintf(&b, "  evidence: %s\n", indentText(f.Evidence, "    "))
		if f.Draft != "" {
			b.WriteString("  draft by the agent, which the owner may edit and run:\n")
			fmt.Fprintf(&b, "    %s\n", indentText(f.Draft, "    "))
		}
		if f.Refuted {
			continue
		}
		b.WriteString("  may:\n")
		for _, g := range inWork {
			if f.Into != g {
				fmt.Fprintf(&b, "    hyp finding fold %s --goal %s\n", f.ID, g.ID)
			}
		}
		fmt.Fprintf(&b, "    hyp finding refute %s\n", f.ID)
	}
	if shown == 0 {
		if o.NotFolded {
			b.WriteString("no finding is not folded.\n")
		} else {
			b.WriteString("no finding.\n")
		}
	}
	if hidden > 0 {
		fmt.Fprintf(&b, "%s folded into an ended Goal not shown; hyp finding list --all lists them.\n", findingCount(hidden))
	}
	b.WriteString("record a new finding:\n")
	b.WriteString("  hyp finding add --fact TEXT --evidence TEXT [--draft \"hyp add goal ...\"]\n")
	return b.String()
}

// goalsInWork are the Goals a fold may name: written, open or selected.
func goalsInWork(t tables.Tables) []*tables.Goal {
	var out []*tables.Goal
	for _, g := range tables.SortedGoals(t) {
		if g.State != tables.StateEnded {
			out = append(out, g)
		}
	}
	return out
}

// findingCount is "1 finding" or "N findings".
func findingCount(n int) string {
	if n == 1 {
		return "1 finding"
	}
	return fmt.Sprintf("%d findings", n)
}

// notFolded counts the findings that are not folded.
func notFolded(t tables.Tables) int {
	n := 0
	for _, f := range tables.SortedFindings(t) {
		if f.State() == spec.FindingNotFolded {
			n++
		}
	}
	return n
}
