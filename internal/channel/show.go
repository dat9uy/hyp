package channel

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// shownRead is one read show prints, and the outputs folder that holds its
// bytes: the trace's, or for a read an earlier hyp wrote into the log, the
// store's.
type shownRead struct {
	tables.Read
	outputs string
}

// readsOf returns every read of actor, oldest first: the reads of the clone's
// trace, and the read events an earlier hyp wrote into the log, which the log
// keeps.
func readsOf(root, actor string) ([]shownRead, error) {
	logEvs, _, err := logstore.StoreEvents(root)
	if err != nil {
		return nil, err
	}
	var out []shownRead
	for _, r := range tables.Replay(logEvs).Reads[actor] {
		out = append(out, shownRead{r, logstore.OutputsDir(root)})
	}
	tr, err := logstore.TraceOf(root)
	if err != nil {
		return nil, err
	}
	evs, _, err := tr.Events()
	if err != nil {
		return nil, err
	}
	for _, r := range tables.Replay(evs).Reads[actor] {
		out = append(out, shownRead{r, tr.Outputs()})
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
		writeStoredIn(&b, root, r.outputs, "stderr", r.Stderr, 0)
		writeStoredIn(&b, root, r.outputs, "stdout", r.Stdout, 0)
	}
	return b.String()
}
