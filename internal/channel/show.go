package channel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/tables"
)

// showText is the report of hyp show ACTOR: every read the store holds for
// the actor, oldest first, each with its id, time, role, exit code and
// command, then the bytes hyp printed to stderr and to stdout, in the order
// it printed them, as blocks of the bytes it recorded. It prints the trace,
// never a fresh read: what the actor got stays what it got, however the log
// changed after.
func showText(root, actor string, reads []tables.Read) string {
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
