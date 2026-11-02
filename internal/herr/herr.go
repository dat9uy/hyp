// Package herr holds the kinds of failure a hyp command can end in. Packages
// return an error of one kind; only cli.Run maps a kind to an exit code.
package herr

import (
	"errors"
	"fmt"
	"strings"
)

// Kind says whose fault a failure is.
type Kind int

const (
	// Refusal: the command line is right, but a rule of hyp refuses the
	// request or a file hyp reads breaks a rule.
	Refusal Kind = iota + 1
	// BadArguments: the command line is wrong.
	BadArguments
	// Environment: I/O, a busy lock, a missing tool, a leftover process.
	Environment
)

const refusedPrefix = "refused: "

// Error is a failure of one Kind. The message of a Refusal starts with
// "refused: ".
type Error struct {
	Kind Kind
	msg  string
}

func (e *Error) Error() string {
	if e.Kind == Refusal {
		return refusedPrefix + e.msg
	}
	return e.msg
}

func New(k Kind, format string, a ...any) error {
	return &Error{Kind: k, msg: fmt.Sprintf(format, a...)}
}

func Refused(format string, a ...any) error { return New(Refusal, format, a...) }
func BadArgs(format string, a ...any) error { return New(BadArguments, format, a...) }
func Env(format string, a ...any) error     { return New(Environment, format, a...) }

// KindOf returns the kind of err; an error of no kind is an Environment one.
func KindOf(err error) Kind {
	if e, ok := errors.AsType[*Error](err); ok {
		return e.Kind
	}
	return Environment
}

// Text is err's message without a leading "refused: ", for a caller that
// words the failure itself.
func Text(err error) string {
	return strings.TrimPrefix(err.Error(), refusedPrefix)
}
