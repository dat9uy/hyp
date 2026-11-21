// Package hyp holds the files of this repository that the binary prints as
// they are, so the binary and the skills that ship them never differ.
package hyp

import _ "embed"

// GenericProbe is the generic Probe the hyp-grill skill ships as probe/run,
// byte for byte.
//
//go:embed skills/hyp-grill/probe/run
var GenericProbe string
