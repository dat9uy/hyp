// Package fault is how a part of hyp reports a failure without writing text:
// the failure's kind, the guard that refused, a code that names the failure,
// and the values it is about. The channel (internal/channel) holds the words
// of every code; it is the only place a failure becomes text or an exit code.
package fault

import "errors"

// Kind says whose fault a failure is.
type Kind int

const (
	// Refusal: the command line is right, but a rule of hyp refuses the
	// request or a file hyp reads breaks a rule.
	Refusal Kind = iota + 1
	// BadArguments: the command line is wrong.
	BadArguments
	// Environment: I/O, a missing tool, a Probe that failed.
	Environment
)

// Guard is the name of a guard, as its entry in internal/spec names it and
// hyp spec prints it. It is data: a Stable refusal stores it in the log.
type Guard string

// The guards.
const (
	GuardDraft       Guard = "Draft"
	GuardNotEnded    Guard = "Not ended"
	GuardOwnerLands  Guard = "Owner lands"
	GuardOwnerClose  Guard = "Owner closes"
	GuardOwnerUnsel  Guard = "Owner unselects"
	GuardNewCell     Guard = "New cell"
	GuardStop        Guard = "Stop"
	GuardLandedFirst Guard = "Landed first"
	GuardWholeLines  Guard = "Whole lines"
	GuardOpen        Guard = "Open"
	GuardGoalsActor  Guard = "Goal's actor"
	GuardOwnerPlaces Guard = "Owner places"
	GuardEvidence    Guard = "Evidence"
	GuardKind        Guard = "Kind"
	GuardCommitted   Guard = "Committed"
	GuardAddsOnly    Guard = "Adds only"
	GuardMeetable    Guard = "Meetable"
	GuardWholeSet    Guard = "Whole set"
	GuardLeafOnly    Guard = "Leaf only"
	GuardBytes       Guard = "Bytes"
	GuardStable      Guard = "Stable"
	GuardBudget      Guard = "Budget"
	GuardSameBase    Guard = "Same base"
	GuardCandidate   Guard = "Candidate"
	GuardSpec        Guard = "Spec"
	GuardOwnPath     Guard = "Own path"
	GuardSelected    Guard = "Selected"
	GuardIssue       Guard = "Issue"
	GuardInWork      Guard = "In work"
	GuardNotRefuted  Guard = "Not refuted"
	GuardOwnerDecl   Guard = "Owner declines"
	GuardNotDeclined Guard = "Not declined"
	GuardNotFolded   Guard = "Not folded"
	GuardOneInWork   Guard = "One in work"
	GuardCovered     Guard = "Covered"
	NoGuard          Guard = "" // a refusal that no guard names
)

// Code names one failure. Its value is an identifier, never text for the
// agent: the channel maps it to its words.
type Code string

// Error is one failure: its kind, the guard that refused (empty when no
// guard did), its code and the values its words name, in order. A value may
// be another error: the failure that caused this one.
type Error struct {
	Kind  Kind
	Guard Guard
	Code  Code
	Args  []any
}

// Error returns the code. It is an identifier for a log or a test, never
// the text hyp prints: only the channel words a failure.
func (e *Error) Error() string { return string(e.Code) }

// Refuse is a refusal by guard g.
func Refuse(g Guard, c Code, args ...any) error {
	return &Error{Kind: Refusal, Guard: g, Code: c, Args: args}
}

// BadArg is a bad argument.
func BadArg(c Code, args ...any) error { return &Error{Kind: BadArguments, Code: c, Args: args} }

// EnvFault is an environment failure.
func EnvFault(c Code, args ...any) error { return &Error{Kind: Environment, Code: c, Args: args} }

// As returns the failure err is, or holds in its chain.
func As(err error) (*Error, bool) { return errors.AsType[*Error](err) }

// KindOf returns the kind of err; an error that is no failure of hyp is an
// Environment one.
func KindOf(err error) Kind {
	if e, ok := As(err); ok {
		return e.Kind
	}
	return Environment
}

// GuardOf returns the guard that refused, or "".
func GuardOf(err error) Guard {
	if e, ok := As(err); ok {
		return e.Guard
	}
	return ""
}

// CodeOf returns the code of err, or "".
func CodeOf(err error) Code {
	if e, ok := As(err); ok {
		return e.Code
	}
	return ""
}
