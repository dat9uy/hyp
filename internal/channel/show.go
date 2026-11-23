package channel

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/tables"
)

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
