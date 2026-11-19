package testutil

import (
	"fmt"
	"strings"

	"github.com/dat9uy/hyp/internal/fault"
)

// Mentions reports whether a value err carries, or a value of a failure it
// carries as a cause, contains s when printed. A part of hyp outside the
// channel writes no text, so a test of it checks the values its failure
// names, not words.
func Mentions(err error, s string) bool {
	e, ok := fault.As(err)
	if !ok {
		return err != nil && strings.Contains(err.Error(), s)
	}
	if strings.Contains(string(e.Code), s) {
		return true
	}
	for _, a := range e.Args {
		if cause, ok := a.(error); ok {
			if Mentions(cause, s) {
				return true
			}
			continue
		}
		if strings.Contains(fmt.Sprint(a), s) {
			return true
		}
	}
	return false
}
