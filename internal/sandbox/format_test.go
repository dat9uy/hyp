package sandbox

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/event"
)

var update = flag.Bool("update", false, "write missing golden files")

// shape lists one "path type" line per leaf of v as JSON, sorted. Object keys
// join with ".", an array element is "[]".
func shape(t *testing.T, v any) []string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var tree any
	if err := json.Unmarshal(data, &tree); err != nil {
		t.Fatal(err)
	}
	var lines []string
	var walk func(path string, n any)
	walk = func(path string, n any) {
		join := func(k string) string {
			if path == "" {
				return k
			}
			return path + "." + k
		}
		switch n := n.(type) {
		case map[string]any:
			if len(n) == 0 {
				lines = append(lines, path+" object")
			}
			for k, c := range n {
				walk(join(k), c)
			}
		case []any:
			if len(n) == 0 {
				lines = append(lines, path+" array")
			}
			for _, c := range n {
				walk(path+"[]", c)
			}
		case string:
			lines = append(lines, path+" string")
		case float64:
			lines = append(lines, path+" number")
		case bool:
			lines = append(lines, path+" bool")
		case nil:
			lines = append(lines, path+" null")
		}
	}
	walk("", tree)
	slices.Sort(lines)
	return slices.Compact(lines)
}

// jsonNames lists every JSON key the struct type t (and the structs it holds)
// declares.
func jsonNames(t reflect.Type, into map[string]bool) {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	for f := range t.Fields() {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" || !f.IsExported() {
			continue
		}
		into[name] = true
		jsonNames(f.Type, into)
	}
}

func fullKnownAnswers() *cmdrecord.KnownAnswers {
	return &cmdrecord.KnownAnswers{HypSHA256: "h", Cases: []cmdrecord.Case{{Name: "a", Outcome: "ok"}}}
}

func fullCommand() cmdrecord.Run {
	exit := 0
	return cmdrecord.Run{Ref: "r@1", Step: "standalone", Label: "l", Args: []string{"./x"}, Exit: &exit, Error: "e", PGID: 1, Started: "s", Ended: "e"}
}

func fullRecord() Record {
	return Record{Format: SandboxFormat, ID: "i", Host: "h", Base: "b", HypSHA256: "h", HypVersion: "v",
		Recipes:  []Entry{{Ref: event.Ref{Name: "r", Version: "1"}, Name: "r", Version: "1", SHA256: "s", Setup: "passed", Cleanup: "ok", KnownAnswers: fullKnownAnswers()}},
		Inputs:   &Inputs{Name: "n", SHA256: "h"},
		Patches:  []PatchSum{{Arm: "a", SHA256: "h"}},
		Commands: []cmdrecord.Run{fullCommand()}, Store: "s", Cut: "c", Created: "c", Removed: "r"}
}

func fullCommandRecord() cmdrecord.Record {
	return cmdrecord.Record{Format: cmdrecord.Format, Ref: event.Ref{Name: "r", Version: "1"}, Name: "r", Version: "1", SHA256: "s", Path: "p", Step: "standalone", Cwd: "c", Created: "c",
		Commands: []cmdrecord.Run{fullCommand()}, KnownAnswers: fullKnownAnswers(),
		Inputs: []cmdrecord.Input{{Arg: 1, Path: "inputs/1/f", SHA256: "h", Size: 1, Kept: true}}}
}

func fullChanges() Changes {
	copied := true
	return Changes{Format: ChangesFormat, Baseline: "b", Changes: []Change{
		{Path: "a", Change: "added", Size: 1, SHA256: "h", Mode: "-rw-r--r--", Copied: &copied},
		{Path: "l", Change: "modified", Size: 1, Mode: "Lrwxr-xr-x", Link: "t"},
	}}
}

// TestRecordShapesHaveGoldens fails when a record's JSON shape changes
// without a new format number, so a reader can trust the number.
func TestRecordShapesHaveGoldens(t *testing.T) {
	for _, c := range []struct {
		name   string
		format int
		v      any
	}{
		{"sandbox", SandboxFormat, fullRecord()},
		{"command", cmdrecord.Format, fullCommandRecord()},
		{"changes", ChangesFormat, fullChanges()},
	} {
		lines := shape(t, c.v)
		// Every declared key must be set in the full value, or the golden
		// would miss an omitempty field.
		names := map[string]bool{}
		jsonNames(reflect.TypeOf(c.v), names)
		text := strings.Join(lines, "\n")
		for n := range names {
			if !strings.Contains(text, n+" ") && !strings.Contains(text, n+".") && !strings.Contains(text, n+"[]") {
				t.Errorf("%s: full value leaves %q unset; set it in full%s", c.name, n, c.name)
			}
		}
		golden := filepath.Join("testdata", fmt.Sprintf("%s-format-%d.golden", c.name, c.format))
		got := text + "\n"
		want, err := os.ReadFile(golden)
		if os.IsNotExist(err) && *update {
			if err := os.MkdirAll("testdata", 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: %v; run go test ./internal/sandbox -run TestRecordShapesHaveGoldens -update", c.name, err)
		}
		if string(want) != got {
			title := strings.ToUpper(c.name[:1]) + c.name[1:]
			t.Errorf("%s.json changed shape; bump %sFormat to %d, then run go test ./internal/sandbox -run TestRecordShapesHaveGoldens -update\ngot:\n%swant:\n%s",
				c.name, title, c.format+1, got, want)
		}
	}
}
