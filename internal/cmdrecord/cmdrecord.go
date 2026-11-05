// Package cmdrecord owns the command record, command.json (sandbox spec §5.1): the
// recipe that ran, each command hyp ran, and the known answers that ran
// before them. sandbox writes it; containment reads it. It is a leaf
// package, so both can import it.
package cmdrecord

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
)

// Format is the shape number of command.json.
const Format = 4

// Run is one command hyp ran, a row of the sandbox record or of a command
// record (docs/sandbox-spec.md §5.1). Exit is set when the script exited,
// Error when it did not: it timed out, a signal killed it, or it did not
// start. Args is the command as hyp ran it, its script shown as ./PATH
// inside the bundle. Label names a Standalone command. PGID is the process
// group the command led (its process id). hyp writes the row before the
// command starts; one with no Ended (and so no Exit or Error) is a command
// hyp stopped during.
type Run struct {
	Ref     string      `json:"ref"`
	Step    recipe.Step `json:"step"`
	Label   string      `json:"label,omitempty"`
	Args    []string    `json:"args"`
	Exit    *int        `json:"exit,omitempty"`
	Error   string      `json:"error,omitempty"`
	PGID    int         `json:"pgid,omitempty"`
	Started string      `json:"started"`
	Ended   string      `json:"ended,omitempty"`
}

// Shown is the command as recipe.log shows it.
func (c Run) Shown() string { return strings.Join(c.Args, " ") }

// Failed reports whether the command did not exit 0.
func (c Run) Failed() bool { return c.Exit == nil || *c.Exit != 0 }

// Outcome is "ok", "exited N", the reason it did not exit, or "did not
// end" for a row hyp stopped during.
func (c Run) Outcome() string {
	switch {
	case c.Exit == nil && c.Ended == "":
		return "did not end"
	case c.Exit == nil:
		return c.Error
	case *c.Exit == 0:
		return "ok"
	}
	return fmt.Sprintf("exited %d", *c.Exit)
}

// KnownAnswers is one bundle's gate: the hyp that ran it and each case's outcome.
type KnownAnswers struct {
	HypSHA256 string `json:"hyp_sha256"`
	Cases     []Case `json:"cases"`
}

// Case is one known-answer case and its outcome: ok, or why it failed.
type Case struct {
	Name    string `json:"name"`
	Outcome string `json:"outcome"`
}

// Passed reports whether every case is ok and there is at least one.
func (k *KnownAnswers) Passed() bool {
	if k == nil || len(k.Cases) == 0 {
		return false
	}
	for _, c := range k.Cases {
		if c.Outcome != "ok" {
			return false
		}
	}
	return true
}

// Record is command.json (sandbox spec §5.1): the recipe that ran, from a copy with
// this hash, and each command with its arguments and exit code. Its output
// is command.log beside it.
type Record struct {
	Format   int         `json:"format"`
	Ref      event.Ref   `json:"ref"`
	Name     string      `json:"name"`
	Version  string      `json:"version"`
	SHA256   string      `json:"sha256"`
	Path     string      `json:"path"`
	Step     recipe.Step `json:"step"`
	Cwd      string      `json:"cwd"`
	Created  string      `json:"created"`
	Commands []Run       `json:"commands"`
	// KnownAnswers is the outcome of the recipe's known answers, run before
	// its commands; nil until the gate runs them.
	KnownAnswers *KnownAnswers `json:"known_answers,omitempty"`
	// Inputs is each operator argument that names a regular file, in ARGS
	// order; absent when none does.
	Inputs []Input `json:"inputs,omitempty"`
	// ArgSHA256 is a format-3 record's arg_sha256, each file argument's
	// absolute path to its sha256. It is read, and never written.
	ArgSHA256 map[string]string `json:"-"`
}

// Input is one kept input: ARGS word Arg (counted from 1) named a regular
// file, which hyp copied to Path (inputs/<arg>/<base name>, relative to the
// record's directory) when Kept, hashed, and passed to the script as a copy.
// A file over the size cap is not kept; its sha256 and size still are.
type Input struct {
	Arg    int    `json:"arg"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Kept   bool   `json:"kept"`
}

type plainRecord Record

// UnmarshalJSON also reads a format-3 record's arg_sha256.
func (r *Record) UnmarshalJSON(data []byte) error {
	var v struct {
		plainRecord
		ArgSHA256 map[string]string `json:"arg_sha256"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*r = Record(v.plainRecord)
	r.ArgSHA256 = v.ArgSHA256
	return nil
}

// Read reads and decodes the command record at path. It checks no rule: a
// record that decodes is returned whatever it says.
func Read(path string) (Record, error) {
	var r Record
	data, err := os.ReadFile(path)
	if err != nil {
		return Record{}, err
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return Record{}, err
	}
	return r, nil
}

// CleanRun reports whether r is a clean run of ref at hash: an Evidence or
// Standalone record with at least one command, every one of which exited 0,
// whose known answers ran before it with at least one case, every one ok.
// The zero Record, which stands for one that could not be read, is not.
func (r Record) CleanRun(ref event.Ref, hash string) bool {
	if r.Ref != ref || r.SHA256 != hash || (r.Step != recipe.Evidence && r.Step != recipe.Standalone) || len(r.Commands) == 0 {
		return false
	}
	for _, c := range r.Commands {
		if c.Failed() {
			return false
		}
	}
	return r.KnownAnswers.Passed()
}
