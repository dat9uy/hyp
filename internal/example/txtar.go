// Package example parses an example file and compares a Probe's output with
// it by the fixed rule of the state machine, section 9.
package example

import (
	"regexp"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
)

// Floor is a literal the output must contain (Op contains) or equal (Op
// equals).
type Floor struct{ Op, Text string }

// Input is a file the run uses, written at Path under the example folder's
// input/.
type Input struct{ Path, Body string }

// The two floor operations.
const (
	OpContains = "contains"
	OpEquals   = "equals"
)

// Example is a parsed example.txtar.
type Example struct {
	Words       string
	Run         string
	Inputs      []Input
	Floors      []Floor
	Ceilings    []string
	HasExternal bool
	External    string // the body of the external section, one name per line
}

var envNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ExternalNames lists the environment variable names of the external section.
func (e Example) ExternalNames() []string {
	var names []string
	for _, n := range strings.Split(e.External, "\n") {
		if n != "" {
			names = append(names, n)
		}
	}
	return names
}

// marker reports whether line is a section marker, "-- name --", and returns
// the name.
func marker(line string) (string, bool) {
	if len(line) < 6 || !strings.HasPrefix(line, "-- ") || !strings.HasSuffix(line, " --") {
		return "", false
	}
	return line[3 : len(line)-3], true
}

// validInputPath reports whether p is a relative path that stays under the
// example folder's input/.
func validInputPath(p string) bool {
	if p == "" || strings.HasPrefix(p, "/") || strings.ContainsRune(p, '\\') || strings.ContainsRune(p, 0) {
		return false
	}
	for _, el := range strings.Split(p, "/") {
		if el == "" || el == "." || el == ".." {
			return false
		}
	}
	return true
}

// Parse reads the txtar of an example file. Each section starts with a line
// "-- name --"; a body drops exactly one final newline and nothing is
// escaped. Text before the first marker, a name outside the allowed ones, a
// repeated words, run or external section, a missing words, run or floor, an
// input path that leaves the folder and an external line that is not an
// environment variable name are refusals that name the line.
func Parse(data []byte) (Example, error) {
	var ex Example
	type section struct {
		name string
		line int
		body strings.Builder
	}
	var secs []*section
	for i, raw := range strings.SplitAfter(string(data), "\n") {
		if raw == "" {
			continue
		}
		text := strings.TrimSuffix(raw, "\n")
		if name, ok := marker(text); ok {
			secs = append(secs, &section{name: name, line: i + 1})
			continue
		}
		if len(secs) == 0 {
			return ex, herr.Refused("example.txtar line %d: text before the first section", i+1)
		}
		secs[len(secs)-1].body.WriteString(raw)
	}

	var haveWords, haveRun bool
	seenInput := map[string]bool{}
	for _, sec := range secs {
		body := strings.TrimSuffix(sec.body.String(), "\n")
		switch {
		case sec.name == "words":
			if haveWords {
				return ex, herr.Refused("example.txtar line %d: a second words section", sec.line)
			}
			haveWords, ex.Words = true, body
		case sec.name == "run":
			if haveRun {
				return ex, herr.Refused("example.txtar line %d: a second run section", sec.line)
			}
			haveRun, ex.Run = true, body
		case sec.name == "floor: contains":
			ex.Floors = append(ex.Floors, Floor{Op: OpContains, Text: body})
		case sec.name == "floor: equals":
			ex.Floors = append(ex.Floors, Floor{Op: OpEquals, Text: body})
		case sec.name == "ceiling":
			ex.Ceilings = append(ex.Ceilings, body)
		case sec.name == "external":
			if ex.HasExternal {
				return ex, herr.Refused("example.txtar line %d: a second external section", sec.line)
			}
			for _, n := range strings.Split(body, "\n") {
				if body != "" && !envNameRe.MatchString(n) {
					return ex, herr.Refused("example.txtar line %d: external line %q is not an environment variable name", sec.line, n)
				}
			}
			ex.HasExternal, ex.External = true, body
		case strings.HasPrefix(sec.name, "input/"):
			path := strings.TrimPrefix(sec.name, "input/")
			if !validInputPath(path) {
				return ex, herr.Refused("example.txtar line %d: input path %q must be a relative path inside the folder", sec.line, path)
			}
			if seenInput[path] {
				return ex, herr.Refused("example.txtar line %d: input/%s appears twice", sec.line, path)
			}
			seenInput[path] = true
			ex.Inputs = append(ex.Inputs, Input{Path: path, Body: body})
		default:
			return ex, herr.Refused("example.txtar line %d: unknown section %q", sec.line, sec.name)
		}
	}
	switch {
	case !haveWords:
		return ex, herr.Refused("example.txtar has no words section")
	case !haveRun:
		return ex, herr.Refused("example.txtar has no run section")
	case len(ex.Floors) == 0:
		return ex, herr.Refused("example.txtar has no floor section")
	}
	return ex, nil
}
