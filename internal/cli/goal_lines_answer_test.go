package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/store"
)

// setupLines is setup with a Goal whose add carries these lines, written
// by the owner o, and returns the lines' keys.
func setupLines(t *testing.T, texts ...string) (dir, g, q string, keys []string) {
	t.Helper()
	dir = t.TempDir()
	t.Chdir(dir)
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	args := []string{"add", "goal", "--title", "G", "--actor", "o", "--why", "start"}
	for _, l := range texts {
		args = append(args, "--line", l)
	}
	r := hyp(t, args...)
	if r.code != 0 {
		t.Fatalf("goal with lines: exit %d, stderr %q", r.code, r.stderr)
	}
	g = strings.Fields(r.stdout)[0]
	for i := range texts {
		keys = append(keys, g+"#"+string(rune('1'+i)))
		want(t, r, 0, "  line "+keys[i]+": "+texts[i]+"\n")
	}
	q = must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	return dir, g, q, keys
}

// addGoal adds a Goal whose add carries these lines, written by owner o,
// and returns it and its lines' keys: a Goal's lines come with its add.
func addGoal(t *testing.T, title, o string, texts ...string) (g string, keys []string) {
	t.Helper()
	args := []string{"add", "goal", "--title", title, "--actor", o, "--why", "a purpose"}
	for _, l := range texts {
		args = append(args, "--line", l)
	}
	r := hyp(t, args...)
	if r.code != 0 {
		t.Fatalf("one store holds many Goals, but Goal %s was refused: exit %d, stderr %q", title, r.code, r.stderr)
	}
	g = strings.Fields(r.stdout)[0]
	for i := range texts {
		keys = append(keys, g+"#"+string(rune('1'+i)))
	}
	return g, keys
}

// rawEvent appends ev to actor's events file as written, with no check, as a
// log merged from another worktree or written by another binary can hold
// it, and returns its id. For node.add the node is the new id.
func rawEvent(t *testing.T, actor, typ, node string, payload any, why string) string {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	id := ulid.MustNew(ulid.Timestamp(now), ulid.DefaultEntropy()).String()
	if typ == event.TypeNodeAdd {
		node = id
	}
	line, err := json.Marshal(event.Event{V: event.Version, ID: id, TS: now.Format(time.RFC3339Nano), Actor: actor, Type: typ, Node: node, Payload: raw, Why: why})
	if err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	root, err := store.FindRoot(cwd)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(root, ".tree", "events", actor+".jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond) // the next id sorts after this one
	return id
}

// wantRejected checks that replay rejects event id with a reason holding
// reason: hyp check reports it as an error.
func wantRejected(t *testing.T, id, reason string) {
	t.Helper()
	r := hyp(t, "check")
	if !strings.Contains(r.stdout, "error: rejected ") || !strings.Contains(r.stdout, reason) {
		t.Fatalf("replay does not reject %s with %q:\n%s%s", id, reason, r.stdout, r.stderr)
	}
}

// appendRefused writes ev through the store, the path every write command
// takes, and checks it is refused with a reason holding reason.
func appendRefused(t *testing.T, actor string, ev event.Event, reason string) {
	t.Helper()
	cwd, _ := os.Getwd()
	root, err := store.FindRoot(cwd)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Append(root, actor, ev); err == nil || !strings.Contains(err.Error(), reason) {
		t.Fatalf("write of %s: err %v, want a refusal holding %q", ev.Type, err, reason)
	}
}

// fullID resolves a node reference to its full id, as stored events hold it.
func fullID(t *testing.T, ref string) string {
	t.Helper()
	cwd, _ := os.Getwd()
	root, err := store.FindRoot(cwd)
	if err != nil {
		t.Fatal(err)
	}
	p, _, err := store.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	id, err := resolveRef(p, ref)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// fullKey resolves the node of key k to its full id.
func fullKey(t *testing.T, k string) string {
	t.Helper()
	ref, n, _ := strings.Cut(k, "#")
	return fullID(t, ref) + "#" + n
}

// typeAnswer is the owner's act (ADR 0020), spelled out so that this file
// compiles on a commit that has no such event type.
const typeAnswer = "answer"

func payload(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// A Goal's lines come with its node.add, and no later event adds one. The
// write: hyp add goal --line writes the lines and prints their keys, hyp
// checkpoint no longer takes --line, and a checkpoint holding lines on a
// Goal whose add carried them is refused.
func TestGoalLinesComeWithTheAddWrite(t *testing.T) {
	_, g, _, keys := setupLines(t, "one tenant per row")
	g = fullID(t, g)
	want(t, hyp(t, "show", keys[0]), 0, "text: one tenant per row\n")
	want(t, hyp(t, "checkpoint", "--line", "x", "--why", "w"), 2, "a Goal's lines come with its node.add")
	appendRefused(t, "o", event.Event{Type: event.TypeCheckpoint, Node: g, Why: "w",
		Payload: payload(t, event.Checkpoint{Summary: "more", Role: event.RoleOwner, Lines: []string{"x"}})},
		"bad lines: a Goal's lines come with its node.add")
}

// The replay half: a log whose checkpoint adds lines to a Goal whose add
// carried them rejects that checkpoint, and the Goal keeps its lines.
func TestGoalLinesComeWithTheAddReplay(t *testing.T) {
	_, g, _, keys := setupLines(t, "one tenant per row")
	g = fullID(t, g)
	late := rawEvent(t, "o", event.TypeCheckpoint, g, event.Checkpoint{Summary: "more", Role: event.RoleOwner, Lines: []string{"x"}}, "w")
	wantRejected(t, late, "bad lines: a Goal's lines come with its node.add")
	if out := hyp(t, "show", g).stdout; !strings.Contains(out, "line "+keys[0]+" ") || strings.Contains(out, "#2") {
		t.Fatalf("the Goal's lines changed:\n%s", out)
	}
}

// A cite must be settled when its line is written: a proposed Hypothesis is
// refused with exit 1, and a selected one is admitted.
func TestGoalLineCiteSettledWrite(t *testing.T) {
	_, g, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	want(t, hyp(t, "add", "goal", "--title", "G2", "--line", "builds on H", "--cite", "1="+h, "--why", "w"), 1,
		"bad cite: line 1 cites "+h+", which is not settled")
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	h2 := supported(t, q2, "H2")
	q = q2
	must(t, "select", q, h2, "--why", "w")
	r := hyp(t, "add", "goal", "--title", "G2", "--line", "builds on H2", "--cite", "1="+h2, "--why", "w")
	want(t, r, 0, "  line ")
	want(t, hyp(t, "show", strings.Fields(r.stdout)[0]+"#1"), 0, "cites "+h2+": realised\n")
}

// The replay half: a log whose Goal add cites a key that is not settled
// rejects the add.
func TestGoalLineCiteSettledReplay(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	g2 := rawEvent(t, "o", event.TypeNodeAdd, "", map[string]any{"kind": event.KindGoal, "parent": "", "title": "G2",
		"lines": []map[string]any{{"text": "builds on H", "cites": []string{fullID(t, h)}}}}, "w")
	r := hyp(t, "check")
	if !strings.Contains(r.stdout, "error: rejected "+g2+": bad cite: line 1 cites ") || !strings.Contains(r.stdout, "which is not settled") {
		t.Fatalf("replay does not reject the Goal add with an unsettled cite:\n%s%s", r.stdout, r.stderr)
	}
}

// A later change in the cited key's state is only a hyp check warning: the
// line keeps what its cite meant, show prints both states, and no status
// changes.
func TestGoalLineCiteChangeIsOnlyACheckWarning(t *testing.T) {
	_, _, q := setup(t)
	h := supported(t, q, "H")
	must(t, "select", q, h, "--why", "w")
	g2 := strings.Fields(hyp(t, "add", "goal", "--title", "G2", "--line", "builds on H", "--cite", "1="+h, "--why", "w").stdout)[0]
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
	before := goalBlock(t, hyp(t, "status").stdout, g2)
	must(t, "unselect", q, "--why", "w")
	want(t, hyp(t, "check"), 0, "warn: "+g2+"#1 cites "+h+", which was realised when the line was written and is unsettled now")
	want(t, hyp(t, "show", g2+"#1"), 0, "cites "+h+": realised, now unsettled\n")
	if after := goalBlock(t, hyp(t, "status").stdout, g2); after != before {
		t.Fatalf("a cite's change moved %s's status:\n%s\nbefore:\n%s", g2, after, before)
	}
}

// openReading records a Result under question q with one reading, by actor
// a, and returns the reading's key.
func openReading(t *testing.T, q string) string {
	t.Helper()
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--reading", "read the line as code-only", "--why", "w")
	return r + "#r1"
}

// Owner and runner fields never mix in one event. The write: an answer that
// holds a runner field, and a checkpoint that holds the runner's refers with
// an owner field, are refused; hyp checkpoint no longer takes owner flags.
func TestOwnerAndRunnerFieldsNeverMixWrite(t *testing.T) {
	_, g, q := setup(t)
	k := fullKey(t, openReading(t, q))
	g = fullID(t, g)
	appendRefused(t, "o", event.Event{Type: typeAnswer,
		Payload: json.RawMessage(`{"answers":[{"key":"` + k + `","text":"code only"}],"refers":[{"key":"` + g + `#1","text":"t"}]}`)},
		"owner and runner fields never mix in one event")
	appendRefused(t, "o", event.Event{Type: typeAnswer,
		Payload: json.RawMessage(`{"answers":[{"key":"` + k + `","text":"code only"}],"summary":"s"}`)},
		"owner and runner fields never mix in one event")
	appendRefused(t, "a", event.Event{Type: event.TypeCheckpoint, Node: g, Why: "w",
		Payload: json.RawMessage(`{"summary":"s","role":"owner","refers":["` + g + `#1"],"closes":["` + k + `"]}`)},
		"owner and runner fields never mix in one event")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--refers", k, "--accept", k, "--why", "w"), 2, "--accept was retired")
}

// The replay half: a log holding an answer with a runner field, or a
// checkpoint that mixes refers with an owner field, rejects each.
func TestOwnerAndRunnerFieldsNeverMixReplay(t *testing.T) {
	_, g, q := setup(t)
	k := fullKey(t, openReading(t, q))
	g = fullID(t, g)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	lim := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "case C", "--why", "w") + "#1"
	lim = fullKey(t, lim)
	a := rawEvent(t, "o", typeAnswer, "", map[string]any{
		"answers": []map[string]string{{"key": k, "text": "code only"}},
		"refers":  []map[string]string{{"key": lim, "text": "t"}}}, "")
	wantRejected(t, a, "bad answer: owner and runner fields never mix in one event: refers is the runner's")
	c := rawEvent(t, "o", event.TypeCheckpoint, g, map[string]any{"summary": "s", "role": "owner",
		"refers": []string{lim}, "closes": []string{k}}, "w")
	wantRejected(t, c, "bad checkpoint: owner and runner fields never mix in one event: refers is the runner's")
	if out := hyp(t, "show", strings.Split(k, "#")[0]).stdout; !strings.Contains(out, " open — read the line as code-only") {
		t.Fatalf("a rejected event answered the reading:\n%s", out)
	}
}

// An empty answer is refused: an answer settles at least one key.
func TestEmptyAnswerRefusedWrite(t *testing.T) {
	setup(t)
	want(t, hyp(t, "answer", "--actor", "o"), 1, "bad answer: empty; an answer settles at least one key")
}

// The replay half: a log holding an empty answer rejects it.
func TestEmptyAnswerRejectedReplay(t *testing.T) {
	setup(t)
	a := rawEvent(t, "o", typeAnswer, "", map[string]any{}, "")
	wantRejected(t, a, "bad answer: empty; an answer settles at least one key")
}

// An empty checkpoint is refused: it holds a summary, refers, or both.
func TestEmptyCheckpointRefusedWrite(t *testing.T) {
	setup(t)
	want(t, hyp(t, "checkpoint", "--summary", "  ", "--why", "w"), 1, "bad checkpoint: empty; a checkpoint holds a summary, refers, or both")
}

// The replay half: a log holding an empty checkpoint rejects it.
func TestEmptyCheckpointRejectedReplay(t *testing.T) {
	_, g, _ := setup(t)
	c := rawEvent(t, "a", event.TypeCheckpoint, fullID(t, g), map[string]any{}, "w")
	wantRejected(t, c, "bad checkpoint: empty; a checkpoint holds a summary, refers, or both")
}

// One answer may settle keys of two Goals: a reading of one Goal and a
// referred line of another, in one owner act. The runner's checkpoint
// refers a line with its own why.
func TestOneAnswerSettlesKeysOfTwoGoals(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	g1, _ := addGoal(t, "G", "o")
	q1 := must(t, "add", "dq", "--parent", g1, "--title", "Q", "--why", "ask")
	g2, k2 := addGoal(t, "G2", "o", "survives a region outage")
	must(t, "add", "dq", "--parent", g2, "--title", "Q2", "--why", "ask")
	k := openReading(t, q1)
	must(t, "checkpoint", "--goal", g2, "--refers", k2[0]+"=no region in the test bed", "--why", "w")
	want(t, hyp(t, "show", k2[0]), 0, "checkpoint — refers: no region in the test bed\n")

	r := hyp(t, "answer", "--answer", k+"=code only", "--accept", k2[0]+"=not this Goal", "--actor", "o")
	want(t, r, 0, " answer\n")
	want(t, hyp(t, "show", strings.Split(k, "#")[0]), 0, "reading "+k+" closed by owner answer ")
	want(t, hyp(t, "show", k2[0]), 0, "hand: accepted by o\n")
	want(t, hyp(t, "show", k2[0]), 0, "o answer — accepts: not this Goal\n")
	want(t, hyp(t, "show", g1), 0, "    answers: "+k+" — code only\n")
	want(t, hyp(t, "show", g2), 0, "    accepts: "+k2[0]+" — not this Goal\n")
	if out := hyp(t, "show", g1).stdout; strings.Contains(out, k2[0]) {
		t.Fatalf("%s's history holds %s's key:\n%s", g1, g2, out)
	}
	want(t, hyp(t, "status"), 0, "o answer answers "+k+"; accepts "+k2[0])
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

// limitOf records a Result under question q with one limit and returns the
// limit's full key.
func limitOf(t *testing.T, q string) string {
	t.Helper()
	h := must(t, "add", "hyp", "--parent", q, "--title", "HL", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	return fullKey(t, must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "case C", "--why", "w")+"#1")
}

// A refer says why: a checkpoint whose refer has no text and no summary is
// refused, and one with either is admitted.
func TestReferWithoutReasonRefusedWrite(t *testing.T) {
	_, _, q := setup(t)
	k := limitOf(t, q)
	want(t, hyp(t, "checkpoint", "--refers", k, "--why", "w"), 1, "bad refers: "+k+" has no reason")
	want(t, hyp(t, "checkpoint", "--refers", k+"=void for every candidate", "--why", "w"), 0, "")
}

// The replay half: a log holding a refer with no text and no summary
// rejects it, and the limit stays the runner's.
func TestReferWithoutReasonRejectedReplay(t *testing.T) {
	_, g, q := setup(t)
	k := limitOf(t, q)
	c := rawEvent(t, "a", event.TypeCheckpoint, fullID(t, g), map[string]any{"refers": []string{k}}, "w")
	wantRejected(t, c, "bad refers: "+k+" has no reason")
	notIn(t, "limit referred")
}
