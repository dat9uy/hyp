package cli

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "write missing golden files")

// Two actors' events, interleaved by id so merge order differs from file
// order; a.jsonl's last event names an unknown parent and is rejected.
var eventsFixture = map[string][]string{
	"a.jsonl": {
		`{"v":1,"id":"01M661QG600000000000000001","ts":"2026-10-30T09:00:01Z","actor":"a","type":"node.add","node":"01M661QG600000000000000001","payload":{"kind":"Goal","parent":"","title":"G","alias":"g"},"why":"start"}`,
		`{"v":1,"id":"01M661QG600000000000000003","ts":"2026-10-30T09:00:03Z","actor":"a","type":"node.add","node":"01M661QG600000000000000003","payload":{"kind":"Hypothesis","parent":"01M661QG600000000000000002","title":"H1","alias":"h1"},"why":"rival"}`,
		`{"v":1,"id":"01M661QG600000000000000005","ts":"2026-10-30T09:00:05Z","actor":"a","type":"node.add","node":"01M661QG600000000000000005","payload":{"kind":"Experiment","parent":"01M661QG60000000000000XXXX","title":"E","alias":"e","recipe":"none"},"why":"orphan"}`,
	},
	"b.jsonl": {
		`{"v":1,"id":"01M661QG600000000000000002","ts":"2026-10-30T09:00:02Z","actor":"b","type":"node.add","node":"01M661QG600000000000000002","payload":{"kind":"DesignQuestion","parent":"01M661QG600000000000000001","title":"Q","alias":"q"},"why":"ask"}`,
		`{"v":1,"id":"01M661QG600000000000000004","ts":"2026-10-30T09:00:04Z","actor":"b","type":"node.add","node":"01M661QG600000000000000004","payload":{"kind":"Hypothesis","parent":"01M661QG600000000000000002","title":"H2","alias":"h2"},"why":"rival"}`,
	},
}

func TestEventsFormat(t *testing.T) {
	golden, err := filepath.Abs(filepath.Join("testdata", "events-format-1.golden"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	t.Chdir(dir)
	must(t, "init")
	stored := map[string]string{}
	for name, lines := range eventsFixture {
		if err := os.WriteFile(filepath.Join(dir, ".tree", "events", name), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, l := range lines {
			var ev struct{ ID string }
			json.Unmarshal([]byte(l), &ev)
			stored[ev.ID] = l
		}
	}
	r := hyp(t, "events", "--json")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
	}

	want, err := os.ReadFile(golden)
	if os.IsNotExist(err) && *update {
		if err := os.WriteFile(golden, []byte(r.stdout), 0o644); err != nil {
			t.Fatal(err)
		}
		want = []byte(r.stdout)
	} else if err != nil {
		t.Fatal(err)
	}
	if r.stdout != string(want) {
		t.Fatalf("hyp events --json differs from %s; a shape change bumps format\ngot:\n%s", golden, r.stdout)
	}

	// Every applied event is the stored line, field for field, in merge order.
	var out struct {
		Format     int
		Events     []json.RawMessage
		Rejections []struct{ Event, Reason string }
	}
	if err := json.Unmarshal([]byte(r.stdout), &out); err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, raw := range out.Events {
		var got, line map[string]any
		var ev struct{ ID string }
		json.Unmarshal(raw, &got)
		json.Unmarshal(raw, &ev)
		json.Unmarshal([]byte(stored[ev.ID]), &line)
		if !reflect.DeepEqual(got, line) {
			t.Errorf("event %s:\n got %s\nstored %s", ev.ID, raw, stored[ev.ID])
		}
		ids = append(ids, ev.ID[len(ev.ID)-1:])
	}
	if strings.Join(ids, "") != "1234" {
		t.Errorf("applied ids end in %v, want merge order 1234", ids)
	}
	if out.Format != 1 || len(out.Rejections) != 1 || out.Rejections[0].Event != "01M661QG600000000000000005" || out.Rejections[0].Reason == "" {
		t.Errorf("format %d, rejections %+v", out.Format, out.Rejections)
	}
}

func TestEventsEmptyStore(t *testing.T) {
	t.Chdir(t.TempDir())
	must(t, "init")
	r := hyp(t, "events", "--json")
	want(t, r, 0, `"events":[]`)
	want(t, r, 0, `"rejections":[]`)
}

func TestEventsNeedsJSON(t *testing.T) {
	setup(t)
	want(t, hyp(t, "events"), 2, "usage: hyp events --json")
	want(t, hyp(t, "events", "--json", "extra"), 2, "usage: hyp events --json")
	want(t, hyp(t, "events", "--json=false"), 2, "usage: hyp events --json")
}
