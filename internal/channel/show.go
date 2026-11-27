package channel

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/tables"
)

// stepped is one Result of an actor that keeps a step log.
type stepped struct {
	tables.Result
	Goal string
}

// steppedOf returns every Result of actor whose cells keep a step log,
// oldest first.
func steppedOf(actor string, log tables.Tables) []stepped {
	var out []stepped
	for _, g := range log.Goals {
		for _, r := range g.Results {
			if r.Actor != actor {
				continue
			}
			for _, c := range r.Cells {
				if c.Steps != "" {
					out = append(out, stepped{*r, g.ID})
					break
				}
			}
		}
	}
	slices.SortStableFunc(out, func(a, b stepped) int { return strings.Compare(a.ID, b.ID) })
	return out
}

// stepsText is the part of hyp show ACTOR after the reads: each Result of
// the actor that keeps a step log, oldest first, with its id, Goal, line and
// commit, then the step log of each of its examples that has one, as a block
// of the bytes the run kept.
func stepsText(root string, results []stepped) string {
	var b strings.Builder
	for i, r := range results {
		fmt.Fprintf(&b, "\nresult %d of %d with steps: %s at %s, %s/%s on %s\n", i+1, len(results), r.ID, r.Time, r.Goal, r.Line, r.Commit)
		for _, c := range r.Cells {
			if c.Steps != "" {
				writeStored(&b, root, "steps of "+r.Line+"/"+c.Example, c.Steps, 0)
			}
		}
	}
	return b.String()
}

// shownRead is one read show prints.
type shownRead struct {
	tables.Read
}

// readsOf returns every read of actor, oldest first: the reads of the clone's
// trace, and the read events an earlier hyp wrote into the log, which the log
// keeps. log replays both.
func readsOf(root, actor string, log tables.Tables) ([]shownRead, error) {
	var out []shownRead
	for _, r := range log.Reads[actor] {
		out = append(out, shownRead{r})
	}
	slices.SortStableFunc(out, func(a, b shownRead) int { return strings.Compare(a.ID, b.ID) })
	return out, nil
}

// showText is the report of hyp show ACTOR: every read of the actor, oldest
// first, each with its id, time, role, exit code and command, then the bytes
// hyp printed to stderr and to stdout, in the order it printed them, as
// blocks of the bytes it recorded. It prints the trace, never a fresh read:
// what the actor got stays what it got, however the log changed after.
func showText(root, actor string, reads []shownRead) string {
	var b strings.Builder
	if len(reads) == 0 {
		fmt.Fprintf(&b, "no read of %s is recorded\n", actor)
		return b.String()
	}
	n := strconv.Itoa(len(reads)) + " reads"
	if len(reads) == 1 {
		n = "1 read"
	}
	fmt.Fprintf(&b, "%s: %s, oldest first\n", actor, n)
	for i, r := range reads {
		role := r.Role
		if role == "" {
			role = "(none)"
		}
		fmt.Fprintf(&b, "\nread %d of %d: %s at %s, role %s, exit %d\n", i+1, len(reads), r.ID, r.Time, role, r.Exit)
		fmt.Fprintf(&b, "command: hyp %s\n", strings.Join(r.Command, " "))
		writeStored(&b, root, "stderr", r.Stderr, 0)
		writeStored(&b, root, "stdout", r.Stdout, 0)
	}
	return b.String()
}
