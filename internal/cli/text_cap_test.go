package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/oklog/ulid/v2"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/store"
)

// casesSHA256 is the sha256 of testdata/text-cap/CASES that the text-cap
// Goal's lines name: the drafts are fixed inputs, and a changed file is a
// different test.
const casesSHA256 = "56ff5491ba208eb2760995f21c68300335fc70a497458dea08f8eee53da871fd"

type textCase struct {
	field, file, draft string
}

// textCases reads testdata/text-cap/CASES and its drafts, and checks the
// file's sha256 and each draft's sha256 and rune count against its row.
func textCases(t *testing.T) []textCase {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("testdata", "text-cap"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "CASES"))
	if err != nil {
		t.Fatal(err)
	}
	if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) != casesSHA256 {
		t.Fatalf("CASES sha256 %x, want %s", sum, casesSHA256)
	}
	var out []textCase
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Fields(line)
		if len(f) < 4 {
			t.Fatalf("bad CASES row %q", line)
		}
		draft, err := os.ReadFile(filepath.Join(dir, f[3]))
		if err != nil {
			t.Fatal(err)
		}
		if sum := sha256.Sum256(draft); hex.EncodeToString(sum[:]) != f[0] {
			t.Fatalf("%s sha256 %x, want %s", f[3], sum, f[0])
		}
		if n, _ := strconv.Atoi(f[2]); utf8.RuneCount(draft) != n {
			t.Fatalf("%s has %d runes, want %s", f[3], utf8.RuneCount(draft), f[2])
		}
		out = append(out, textCase{field: f[1], file: f[3], draft: string(draft)})
	}
	if len(out) != 7 {
		t.Fatalf("CASES has %d rows, want 7", len(out))
	}
	return out
}

// newEvents runs a write and returns the events it appended.
func newEvents(t *testing.T, dir string, write func()) []event.Event {
	t.Helper()
	before, _, err := store.Events(dir)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, ev := range before {
		seen[ev.ID] = true
	}
	write()
	after, _, err := store.Events(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []event.Event
	for _, ev := range after {
		if !seen[ev.ID] {
			out = append(out, ev)
		}
	}
	return out
}

// textCaseStore is a store with a Goal, a question and a Hypothesis under it.
func textCaseStore(t *testing.T) (dir, g, q, h string) {
	t.Helper()
	dir, g, q = setup(t)
	h = must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	return
}

var citedID = regexp.MustCompile(`\b([GQHER])-([0-9A-HJKMNP-TV-Z]{6})\b`)

// seedCited adds to the store a node for each id draft cites, of the kind
// its letter names, as the store the draft was written in held them: hyp
// refuses a text that cites an id the store does not hold. Each seed node
// is a raw event whose id ends in the cited id's six characters.
func seedCited(t *testing.T, g, q, h, draft string) {
	t.Helper()
	e := must(t, "add", "exp", "--parent", h, "--title", "seed E", "--recipe", "none", "--rubric", "reading", "--why", "seed")
	// A seed Hypothesis is no rival of h: it sits under a question of its own.
	sq := must(t, "add", "dq", "--parent", g, "--title", "seed Q", "--why", "seed")
	gid, qid, hid, eid := fullID(t, g), fullID(t, sq), fullID(t, h), fullID(t, e)
	cwd, _ := os.Getwd()
	root, err := store.FindRoot(cwd)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(root, ".tree", "events", "seed.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	seen := map[string]bool{}
	for _, m := range citedID.FindAllStringSubmatch(draft, -1) {
		if seen[m[0]] {
			continue
		}
		seen[m[0]] = true
		time.Sleep(2 * time.Millisecond)
		now := time.Now().UTC()
		id := ulid.MustNew(ulid.Timestamp(now), ulid.DefaultEntropy()).String()[:20] + m[2]
		typ, node := event.TypeNodeAdd, id
		var payload any
		switch m[1] {
		case "G":
			payload = event.NodeAdd{Kind: event.KindGoal, Title: "seed " + m[0]}
		case "Q":
			payload = event.NodeAdd{Kind: event.KindDesignQuestion, Parent: gid, Title: "seed " + m[0]}
		case "H":
			payload = event.NodeAdd{Kind: event.KindHypothesis, Parent: qid, Title: "seed " + m[0]}
		case "E":
			payload = event.NodeAdd{Kind: event.KindExperiment, Parent: hid, Title: "seed " + m[0], Recipe: "none", Rubric: event.RubricReading}
		case "R":
			typ = event.TypeResultRecord
			payload = event.ResultRecord{Parent: eid, Verdicts: map[string]string{hid: event.VerdictSupports}, Observation: "seed " + m[0]}
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		line, err := json.Marshal(event.Event{V: event.Version, ID: id, TS: now.Format(time.RFC3339Nano), Actor: "seed", Type: typ, Node: node, Payload: raw, Why: "seed"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			t.Fatal(err)
		}
	}
	if r := hyp(t, "check"); r.code != 0 {
		t.Fatalf("seeded store: %s%s", r.stdout, r.stderr)
	}
}

// writeCase writes c.draft into the field its row names. It returns the one
// event the write appended, the field's text decoded from that event, the
// node to show and the line of hyp show that must print the draft whole.
func writeCase(t *testing.T, dir, g, q, h string, c textCase) (ev event.Event, got, show, line string) {
	t.Helper()
	seedCited(t, g, q, h, c.draft)
	var args []string
	switch c.field {
	case "experiment-body":
		args = []string{"add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--body", c.draft, "--why", "w"}
		line = "body: " + c.draft
	case "checkpoint-summary":
		args = []string{"checkpoint", "--goal", g, "--summary", c.draft, "--why", "w"}
		line = "    summary: " + c.draft
		show = g
	case "goal-body":
		args = []string{"add", "goal", "--title", "G2", "--body", c.draft, "--why", "w"}
		line = "body: " + c.draft
	case "question-title":
		args = []string{"add", "dq", "--parent", g, "--title", c.draft, "--why", "w"}
		line = "title: " + c.draft
	default:
		t.Fatalf("unknown field %q in CASES", c.field)
	}
	var r result
	evs := newEvents(t, dir, func() { r = hyp(t, args...) })
	if r.code != 0 {
		t.Fatalf("%s: exit %d, stderr %q", c.file, r.code, r.stderr)
	}
	if len(evs) != 1 {
		t.Fatalf("%s: the write appended %d events, want 1", c.file, len(evs))
	}
	ev = evs[0]
	if c.field == "checkpoint-summary" {
		var p event.Checkpoint
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			t.Fatal(err)
		}
		got = p.Summary
	} else {
		var p event.NodeAdd
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			t.Fatal(err)
		}
		got = p.Body
		if c.field == "question-title" {
			got = p.Title
		}
		show = strings.Fields(r.stdout)[0]
	}
	return ev, got, show, line
}

// Each CASES draft written into its field exits 0, and the one event the
// write appends holds the draft whole; replay with .tree/evidence/ removed
// gives it back whole.
func TestTextCapDraftKeptWholeInItsEvent(t *testing.T) {
	for _, c := range textCases(t) {
		t.Run(c.file, func(t *testing.T) {
			dir, g, q, h := textCaseStore(t)
			ev, got, _, _ := writeCase(t, dir, g, q, h, c)
			if got != c.draft {
				t.Fatalf("the event's field is %d runes, want the draft's %d", utf8.RuneCountInString(got), utf8.RuneCountInString(c.draft))
			}
			if err := os.RemoveAll(filepath.Join(dir, ".tree", "evidence")); err != nil {
				t.Fatal(err)
			}
			p, _, err := store.Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(p.Rejections) != 0 {
				t.Fatalf("rejections: %v", p.Rejections)
			}
			var replayed string
			switch c.field {
			case "checkpoint-summary":
				var s event.Checkpoint
				_ = json.Unmarshal(p.LastCheckpoint[ev.Node].Payload, &s)
				replayed = s.Summary
			case "question-title":
				replayed = p.Nodes[ev.ID].Title
			default:
				replayed = p.Nodes[ev.ID].Body
			}
			if replayed != c.draft {
				t.Fatalf("replay gives %d runes, want the draft's %d", utf8.RuneCountInString(replayed), utf8.RuneCountInString(c.draft))
			}
		})
	}
}

// hyp show of the node a CASES row writes, or of the Goal for a checkpoint,
// prints the draft whole on the field's own line.
func TestTextCapShowPrintsTheDraftWhole(t *testing.T) {
	for _, c := range textCases(t) {
		t.Run(c.file, func(t *testing.T) {
			dir, g, q, h := textCaseStore(t)
			_, _, show, line := writeCase(t, dir, g, q, h, c)
			r := hyp(t, "show", show)
			if r.code != 0 {
				t.Fatalf("show %s: exit %d, stderr %q", show, r.code, r.stderr)
			}
			if !strings.Contains("\n"+r.stdout, "\n"+line+"\n") {
				t.Fatalf("show %s does not print the draft whole on its own line:\n%s", show, r.stdout)
			}
		})
	}
}

// After a checkpoint whose summary is the full G-ZBM2RE run 1 stop, hyp
// status G and hyp show G each print its keys and owner commands verbatim.
func TestTextCapStopReportInStatusAndShow(t *testing.T) {
	var draft string
	for _, c := range textCases(t) {
		if c.file == "gzbm2re-run1-stop-full.txt" {
			draft = c.draft
		}
	}
	_, g, q, h := textCaseStore(t)
	seedCited(t, g, q, h, draft)
	must(t, "checkpoint", "--goal", g, "--summary", draft, "--why", "stop")
	wants := []string{"R-KPWZZC#r1", "R-50GBXM#r1", "R-FRF929#r1", "R-9EGPYD#r1", "R-FRF929#r2", "R-9EGPYD#r2",
		"F4", "F8", "G4", "gh pr merge 140 --merge",
		"git checkout main && git pull && CGO_ENABLED=0 go build -o ~/.local/bin/hyp ./cmd/hyp",
		"Personal skills that name a changed command: none"}
	for _, args := range [][]string{{"status", g}, {"show", g}} {
		r := hyp(t, args...)
		if r.code != 0 {
			t.Fatalf("hyp %v: exit %d, stderr %q", args, r.code, r.stderr)
		}
		for _, w := range wants {
			if !strings.Contains(r.stdout, w) {
				t.Errorf("hyp %v does not print %q:\n%s", args, w, r.stdout)
			}
		}
	}
}

// hyp prompt of a Goal whose body is a CASES goal-body draft prints each
// constraints line of the draft whole, in the draft's order.
func TestTextCapPromptPrintsEachConstraintLine(t *testing.T) {
	n := 0
	for _, c := range textCases(t) {
		if c.field != "goal-body" {
			continue
		}
		n++
		t.Run(c.file, func(t *testing.T) {
			dir, g0, q, h := textCaseStore(t)
			if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
				t.Fatal(err)
			}
			seedCited(t, g0, q, h, c.draft)
			g := must(t, "add", "goal", "--title", "G2", "--body", c.draft, "--why", "w")
			r := hyp(t, "prompt", g, "--actor", "runner-1")
			if r.code != 0 {
				t.Fatalf("prompt: exit %d, stderr %q", r.code, r.stderr)
			}
			out := strings.Split(r.stdout, "\n")
			at := 0
			for _, l := range strings.Split(c.draft, "\n") {
				if headingOf(l) == "" {
					continue
				}
				for at < len(out) && out[at] != l {
					at++
				}
				if at == len(out) {
					t.Fatalf("prompt does not print %q whole after the lines before it:\n%s", l, r.stdout)
				}
				at++
			}
		})
	}
	if n == 0 {
		t.Fatal("CASES has no goal-body row")
	}
}

// capCase writes text of n runes into one field; it returns the write's
// result.
type capCase struct {
	name, field string
	max         int
	write       func(t *testing.T, g, q, h, text string) result
}

// recordWithReading records a Result under h that holds one reading, and
// returns the reading's key.
func recordWithReading(t *testing.T, h string) string {
	t.Helper()
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--reading", "line 1: two meanings", "--limit", "a case", "--why", "w")
	return r
}

var capCases = []capCase{
	{"goal title", "title", 200, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "add", "goal", "--title", s, "--why", "w")
	}},
	{"question title", "title", 200, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "add", "dq", "--parent", g, "--title", s, "--why", "w")
	}},
	{"hypothesis title", "title", 200, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "add", "hyp", "--parent", q, "--title", s, "--why", "w")
	}},
	{"experiment title", "title", 200, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "add", "exp", "--parent", h, "--title", s, "--recipe", "none", "--rubric", "reading", "--why", "w")
	}},
	{"goal body", "body", 3000, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "add", "goal", "--title", "G2", "--body", s, "--why", "w")
	}},
	{"question body", "body", 2000, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "add", "dq", "--parent", g, "--title", "Q2", "--body", s, "--why", "w")
	}},
	{"experiment body", "body", 2000, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--body", s, "--why", "w")
	}},
	{"hypothesis body", "body", 600, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "add", "hyp", "--parent", q, "--title", "H2", "--body", s, "--why", "w")
	}},
	{"hypothesis amend", "body", 600, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "amend", h, "--body", s, "--why", "w")
	}},
	{"observation", "observation", 600, func(t *testing.T, g, q, h, s string) result {
		e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
		return hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", s, "--why", "w")
	}},
	{"runner checkpoint summary", "summary", 2000, func(t *testing.T, g, q, h, s string) result {
		return hyp(t, "checkpoint", "--goal", g, "--summary", s, "--why", "w")
	}},
	{"refer text", "text of ", 600, func(t *testing.T, g, q, h, s string) result {
		r := recordWithReading(t, h)
		return hyp(t, "checkpoint", "--goal", g, "--refers", r+"#1="+s, "--why", "w")
	}},
	{"answer item text", "text of ", 600, func(t *testing.T, g, q, h, s string) result {
		r := recordWithReading(t, h)
		return hyp(t, "answer", "--answer", r+"#r1="+s, "--actor", "owner-dat")
	}},
}

// filler is a text of n runes made of one rune that no other field of the
// store holds, so one rune more can be put in its event in the log.
func filler(n int) string { return strings.Repeat("ế", n) }

// overInLog adds one rune to text in the log, in place, so the event that
// held text at its cap now holds one rune more.
func overInLog(t *testing.T, dir, text string) {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(dir, ".tree", "events", "*.jsonl"))
	n := 0
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if c := strings.Count(string(data), text); c > 0 {
			n += c
			if err := os.WriteFile(f, []byte(strings.Replace(string(data), text, text+"ế", 1)), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if n != 1 {
		t.Fatalf("the log holds the text %d times, want 1", n)
	}
}

// Each field's cap holds at its edge: exactly the cap is accepted, one rune
// more is refused at write with exit 1 naming the field, and a log holding
// one rune more is rejected on replay.
func TestTextCapEdges(t *testing.T) {
	for _, c := range capCases {
		t.Run(c.name, func(t *testing.T) {
			dir, g, q, h := textCaseStore(t)
			r := c.write(t, g, q, h, filler(c.max+1))
			if r.code != 1 || !strings.Contains(r.stderr, "over limit: "+c.field) {
				t.Fatalf("%d+1 runes: exit %d, stderr %q; want exit 1 and over limit: %s", c.max, r.code, r.stderr, c.field)
			}
			at := filler(c.max)
			var w result
			evs := newEvents(t, dir, func() { w = c.write(t, g, q, h, at) })
			if w.code != 0 {
				t.Fatalf("%d runes: exit %d, stderr %q", c.max, w.code, w.stderr)
			}
			overInLog(t, dir, at)
			wantRejected(t, evs[len(evs)-1].ID, "over limit: "+c.field)
		})
	}
}

// An old owner checkpoint's summary stays 600 runes, checked on replay only:
// no command writes one.
func TestTextCapOldOwnerSummaryEdge(t *testing.T) {
	_, g, _ := setup(t)
	text := filler(600)
	rawEvent(t, "owner-dat", event.TypeCheckpoint, fullID(t, g), map[string]any{"summary": text, "role": event.RoleOwner}, "old owner checkpoint")
	want(t, hyp(t, "check"), 0, "check: 0 errors")
	dir, _ := os.Getwd()
	overInLog(t, dir, text)
	want(t, hyp(t, "check"), 1, "over limit: summary has 601 runes (max 600)")
}

// A records path holds at most 300 runes.
func TestTextCapRecordsPathEdge(t *testing.T) {
	dir, _, _, h := textCaseStore(t)
	path := func(n int) string {
		p := ".tree/evidence/"
		for utf8.RuneCountInString(p)+100 < n {
			p += strings.Repeat("d", 99) + "/"
		}
		return p + strings.Repeat("f", n-utf8.RuneCountInString(p))
	}
	for _, n := range []int{300, 301} {
		p := filepath.Join(dir, filepath.FromSlash(path(n)))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	add := func(n int) result {
		return hyp(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--records", path(n), "--why", "w")
	}
	want(t, add(301), 1, "over limit: records path has 301 runes (max 300)")
	var w result
	evs := newEvents(t, dir, func() { w = add(300) })
	if w.code != 0 {
		t.Fatalf("300 runes: exit %d, stderr %q", w.code, w.stderr)
	}
	overInLog(t, dir, path(300))
	wantRejected(t, evs[0].ID, "over limit: records path")
}
