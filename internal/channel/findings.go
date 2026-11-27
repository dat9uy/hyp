package channel

import (
	"fmt"
	"strings"

	"github.com/dat9uy/hyp/internal/acts"
	"github.com/dat9uy/hyp/internal/spec"
	"github.com/dat9uy/hyp/internal/tables"
)

// ListOpts selects what FindingList prints.
type ListOpts struct {
	All       bool // list the settled, ended and declined findings too
	NotFolded bool // list only the findings that are not folded
}

// FindingList is what hyp finding list prints: every finding of this
// clone's trace in merge order, each with its state, fact, evidence, the
// add goal its agent drafted, and the finding acts it allows as ready lines;
// then the line that records a new finding. A finding that is settled or
// ended, or declined, is left out unless o.All, and a last line counts each;
// a declined finding prints the owner's why and no act; with
// o.NotFolded only the findings that are not folded are listed. Each line of
// an agent's text is indented, so a finding ends where the next one starts.
// A finding whose evidence names an example prints the example and its walk
// on the commit it names; in a repository that keeps a model at HEAD, a
// finding that names no example reads no walk. repo is the git top level.
func FindingList(t tables.Tables, repo string, o ListOpts) string {
	var b strings.Builder
	keptHere, _ := acts.KeepsModel(repo, "HEAD")
	shown, hidden, declined := 0, 0, 0
	for _, f := range tables.SortedFindings(t) {
		if o.NotFolded && f.State() != spec.FindingNotFolded {
			continue
		}
		if !o.All && f.Ended() {
			hidden++
			continue
		}
		if !o.All && f.Declined {
			declined++
			continue
		}
		shown++
		fmt.Fprintf(&b, "%s  %s  by %s at %s\n", f.ID, findingStateText(f), f.Actor, f.Time)
		fmt.Fprintf(&b, "  fact: %s\n", indentText(f.Fact, "    "))
		fmt.Fprintf(&b, "  evidence: %s\n", indentText(f.Evidence, "    "))
		findingWalk(&b, f, repo, keptHere)
		if f.Draft != "" {
			b.WriteString("  draft by the agent, which the owner may edit and run:\n")
			fmt.Fprintf(&b, "    %s\n", indentText(f.Draft, "    "))
		}
		if f.Declined {
			fmt.Fprintf(&b, "  why: %s\n", indentText(f.Why, "    "))
		}
		if f.Refuted || f.Declined {
			continue
		}
		b.WriteString("  may:\n")
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
		fmt.Fprintf(&b, "%s settled or ended not shown; hyp finding list --all lists them.\n", findingCount(hidden))
	}
	if declined > 0 {
		fmt.Fprintf(&b, "%s declined not shown; hyp finding list --all lists them.\n", findingCount(declined))
	}
	b.WriteString("record a new finding:\n")
	b.WriteString("  hyp finding add --fact TEXT --evidence TEXT [--example G/line-N/name --commit C] [--draft \"hyp add goal ...\"]\n")
	return b.String()
}

// findingWalk prints the example a finding's evidence names and its walk:
// each command of the example's run section, on the commit the finding
// names, with the step of the kept model it takes there, or no step. A
// finding that names no example reads no walk when keptHere: the repository
// keeps a model at HEAD; else it prints nothing, as before a kept model.
func findingWalk(b *strings.Builder, f *tables.Finding, repo string, keptHere bool) {
	if f.Example == "" {
		if keptHere {
			b.WriteString("  no walk: the evidence names no example\n")
		}
		return
	}
	fmt.Fprintf(b, "  example: %s at %s\n", f.Example, f.Commit)
	w, err := findingExampleWalk(repo, f)
	if err != nil {
		fmt.Fprintf(b, "  walk: not readable here: %s\n", bodyText(err, repo))
		return
	}
	if !w.Kept {
		fmt.Fprintf(b, "  walk: %s keeps no model in %s\n", f.Commit, acts.KeptDir)
		return
	}
	b.WriteString("  walk, each command of its run section with the step of the kept model it takes:\n")
	walkSteps(b, w, "    ")
}

// findingExampleWalk is the walk of the example a finding names, on its
// commit.
func findingExampleWalk(repo string, f *tables.Finding) (acts.Walk, error) {
	_, hash, err := acts.ExampleAt(repo, f.Example, f.Commit)
	if err != nil {
		return acts.Walk{}, err
	}
	return acts.WalkExample(repo, f.Commit, hash)
}

// walkSteps prints each command of a walk with its step, or no step, one
// to a line after indent.
func walkSteps(b *strings.Builder, w acts.Walk, indent string) {
	if len(w.Steps) == 0 {
		fmt.Fprintf(b, "%sno command\n", indent)
	}
	for _, s := range w.Steps {
		step := s.Step
		if step == "" {
			step = "no step"
		}
		fmt.Fprintf(b, "%s%s: %s\n", indent, indentText(s.Command, indent+"  "), step)
	}
}

// findingCount is "1 finding" or "N findings".
func findingCount(n int) string {
	if n == 1 {
		return "1 finding"
	}
	return fmt.Sprintf("%d findings", n)
}

// notFolded counts the findings that are not folded: those that wait for a
// Goal. A settled or ended finding does not wait.
func notFolded(t tables.Tables) int {
	n := 0
	for _, f := range tables.SortedFindings(t) {
		if f.State() == spec.FindingNotFolded {
			n++
		}
	}
	return n
}
