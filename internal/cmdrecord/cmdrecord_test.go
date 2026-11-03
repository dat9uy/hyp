package cmdrecord

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"reflect"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
)

var tool = event.Ref{Name: "tool", Version: "1"}

func zero() *int { z := 0; return &z }

// clean is a record that is a clean run of tool at "h".
func clean() Record {
	return Record{Format: Format, Ref: tool, Name: "tool", Version: "1", SHA256: "h", Step: recipe.Standalone,
		Commands:     []Run{{Ref: "tool@1", Step: recipe.Standalone, Label: "x", Args: []string{"./x"}, Exit: zero()}},
		KnownAnswers: &KnownAnswers{HypSHA256: "s", Cases: []Case{{Name: "a", Outcome: "ok"}}}}
}

func TestCleanRun(t *testing.T) {
	if !clean().CleanRun(tool, "h") {
		t.Fatal("the clean record is not a clean run")
	}
	evidence := clean()
	evidence.Step = recipe.Evidence
	if !evidence.CleanRun(tool, "h") {
		t.Error("an Evidence record is not a clean run")
	}
	one := 1
	for name, c := range map[string]struct {
		rec  func(*Record)
		ref  event.Ref
		hash string
	}{
		"unreadable":       {func(r *Record) { *r = Record{} }, tool, "h"},
		"wrong ref":        {func(*Record) {}, event.Ref{Name: "other", Version: "1"}, "h"},
		"wrong hash":       {func(*Record) {}, tool, "other"},
		"drive step":       {func(r *Record) { r.Step = recipe.Drive }, tool, "h"},
		"no commands":      {func(r *Record) { r.Commands = nil }, tool, "h"},
		"command exited 1": {func(r *Record) { r.Commands[0].Exit = &one }, tool, "h"},
		"command no exit":  {func(r *Record) { r.Commands[0].Exit = nil; r.Commands[0].Error = "timed out" }, tool, "h"},
		"no known answers": {func(r *Record) { r.KnownAnswers = nil }, tool, "h"},
		"no cases":         {func(r *Record) { r.KnownAnswers.Cases = nil }, tool, "h"},
		"case not ok":      {func(r *Record) { r.KnownAnswers.Cases[0].Outcome = "stdout differs" }, tool, "h"},
	} {
		r := clean()
		c.rec(&r)
		if r.CleanRun(c.ref, c.hash) {
			t.Errorf("%s: CleanRun = true, want false", name)
		}
	}
}

// Read decodes the command.json the synthetic store cites.
func TestReadFixture(t *testing.T) {
	r, err := Read(filepath.Join("..", "testutil", "testdata", "synth", ".tree", "evidence", "E-8Y7V9P", "judge", "command.json"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Ref != (event.Ref{Name: "count-check", Version: "1"}) || r.Step != recipe.Standalone || len(r.Commands) != 1 ||
		r.Commands[0].Label != "judge" || r.Commands[0].Failed() || !r.KnownAnswers.Passed() || len(r.KnownAnswers.Cases) != 3 {
		t.Fatalf("Read = %+v", r)
	}
}

// A full record survives a write and Read unchanged.
func TestReadRoundTrip(t *testing.T) {
	want := clean()
	want.Path, want.Cwd, want.Created = "p", "c", "t"
	want.Commands[0].PGID, want.Commands[0].Started, want.Commands[0].Ended = 7, "s", "e"
	want.Inputs = []Input{{Arg: 1, Path: "inputs/1/a", SHA256: "b", Size: 2, Kept: true}}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "command.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Read = %+v, want %+v", got, want)
	}
}

func TestReadErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := Read(filepath.Join(dir, "missing.json")); err == nil {
		t.Error("Read of a missing file: no error")
	}
	bad := filepath.Join(dir, "bad.json")
	os.WriteFile(bad, []byte("{"), 0o644)
	if r, err := Read(bad); err == nil || !reflect.DeepEqual(r, Record{}) {
		t.Errorf("Read of bad JSON = %+v, %v; want the zero Record and an error", r, err)
	}
}

func TestRunOutcome(t *testing.T) {
	one := 1
	for want, c := range map[string]Run{
		"ok":          {Exit: zero(), Ended: "e"},
		"exited 1":    {Exit: &one, Ended: "e"},
		"timed out":   {Error: "timed out", Ended: "e"},
		"did not end": {},
	} {
		if got := c.Outcome(); got != want {
			t.Errorf("Outcome = %q, want %q", got, want)
		}
	}
}

// A format-3 record's arg_sha256 is read, and never written back.
func TestReadFormat3ArgSHA256(t *testing.T) {
	path := filepath.Join(t.TempDir(), "command.json")
	if err := os.WriteFile(path, []byte(`{"format": 3, "step": "standalone", "arg_sha256": {"/a": "b"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := Read(path)
	if err != nil || r.Format != 3 || r.ArgSHA256["/a"] != "b" {
		t.Fatalf("Read = %+v, %v", r, err)
	}
	if data, _ := json.Marshal(r); strings.Contains(string(data), "arg_sha256") {
		t.Errorf("written = %s", data)
	}
}
