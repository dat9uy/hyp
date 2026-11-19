package acts

import (
	"encoding/json"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
)

// ReadOpts is what hyp printed to one actor for one command: the trace a
// read event keeps.
type ReadOpts struct {
	Dir     string // where the command ran
	Actor   string // a valid actor
	Role    string // the role the command declared, or empty
	Command []string
	Exit    int
	Stdout  []byte
	Stderr  []byte
}

// RecordRead keeps the trace of one command, before hyp prints it: it stores
// both outputs by hash in the clone's trace folder, then appends a read event
// to the actor's file there. The trace is never in the log: it changes no
// Goal, and no file git tracks or sees. It records nothing, and returns
// false, when the command ran outside a git repository or the repository has
// no store: a refused first add goal leaves no store behind. A read has no
// guard.
func RecordRead(o ReadOpts) (bool, error) {
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return false, nil
	}
	if logstore.Find(root) != nil {
		return false, nil
	}
	tr, err := logstore.TraceOf(root)
	if err != nil {
		return false, err
	}
	outHash, err := tr.PutOutput(o.Stdout)
	if err != nil {
		return false, err
	}
	errHash, err := tr.PutOutput(o.Stderr)
	if err != nil {
		return false, err
	}
	role := o.Role
	if !ValidRole(role) {
		role = ""
	}
	payload, err := json.Marshal(events.Read{Role: role, Command: o.Command, Exit: o.Exit, Stdout: outHash, Stderr: errHash})
	if err != nil {
		return false, fault.EnvFault(fault.EncodeEvent, events.TypeRead, err)
	}
	_, err = tr.Append(o.Actor, events.Event{Type: events.TypeRead, Payload: payload})
	return err == nil, err
}
