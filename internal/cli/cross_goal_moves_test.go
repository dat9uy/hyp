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

// twoGoals is a store with Goals A and B, each with one line, one top-level
// question and one supported candidate.
type twoGoals struct {
	gA, gB, qA, qB, hA, hB, kA, kB string
}

func setupTwoGoals(t *testing.T) twoGoals {
	t.Helper()
	t.Chdir(t.TempDir())
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	var s twoGoals
	var ks []string
	s.gA, ks = addGoal(t, "A", "o", "one tenant per row")
	s.kA = ks[0]
	s.qA = must(t, "add", "dq", "--parent", s.gA, "--title", "QA", "--why", "ask")
	s.gB, ks = addGoal(t, "B", "o", "survives a region outage")
	s.kB = ks[0]
	s.qB = must(t, "add", "dq", "--parent", s.gB, "--title", "QB", "--why", "ask")
	s.hA = supported(t, s.qA, "HA")
	s.hB = supported(t, s.qB, "HB")
	return s
}

// expOf adds a planned Experiment under h, with these aims.
func expOf(t *testing.T, h string, aims ...string) string {
	t.Helper()
	args := []string{"add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w"}
	for _, k := range aims {
		args = append(args, "--aims", k)
	}
	return must(t, args...)
}

// answeredReading records a reading under q and has the owner o answer it.
func answeredReading(t *testing.T, q string) string {
	t.Helper()
	k := openReading(t, q)
	must(t, "answer", "--answer", k+"=code only", "--actor", "o")
	return k
}

// rawResult appends a result.record to actor's events file as written, with
// no check, and returns its id: a Result's node is its own id.
func rawResult(t *testing.T, actor string, p map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	id := ulid.MustNew(ulid.Timestamp(now), ulid.DefaultEntropy()).String()
	line, err := json.Marshal(event.Event{V: event.Version, ID: id, TS: now.Format(time.RFC3339Nano), Actor: actor, Type: event.TypeResultRecord, Node: id, Payload: raw, Why: "w"})
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
	time.Sleep(2 * time.Millisecond)
	return id
}

const crossGoals = "a move never crosses Goals"

// A refer leads to its checkpoint's Goal: a checkpoint of Goal A that refers
// Goal B's line is refused with exit 1.
func TestCrossGoalReferRefusedWrite(t *testing.T) {
	s := setupTwoGoals(t)
	want(t, hyp(t, "checkpoint", "--goal", s.gA, "--refers", s.kB+"=no region in the test bed", "--why", "w"), 1, "bad refers: "+s.kB+" leads to "+s.gB)
	want(t, hyp(t, "checkpoint", "--goal", s.gA, "--refers", s.kB+"=no region in the test bed", "--why", "w"), 1, crossGoals)
	want(t, hyp(t, "checkpoint", "--goal", s.gB, "--refers", s.kB+"=no region in the test bed", "--why", "w"), 0, "")
}

// The replay half: a log whose checkpoint of Goal A refers Goal B's line
// rejects it, and the line is not referred.
func TestCrossGoalReferRejectedReplay(t *testing.T) {
	s := setupTwoGoals(t)
	c := rawEvent(t, "a", event.TypeCheckpoint, fullID(t, s.gA), map[string]any{"refers": []map[string]string{{"key": fullKey(t, s.kB), "text": "t"}}}, "w")
	wantRejected(t, c, "bad refers: ")
	wantRejected(t, c, crossGoals)
	if out := hyp(t, "show", s.kB).stdout; strings.Contains(out, "referred") || strings.Contains(out, "hand:") {
		t.Fatalf("a rejected refer moved %s:\n%s", s.kB, out)
	}
}

// An Experiment aims only at a limit of its own Goal: one of Goal A that
// aims at a limit of Goal B's Result is refused with exit 1.
func TestCrossGoalAimRefusedWrite(t *testing.T) {
	s := setupTwoGoals(t)
	limB := limitOf(t, s.qB)
	want(t, hyp(t, "add", "exp", "--parent", s.hA, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--aims", limB, "--why", "w"), 1, "bad aims: ")
	want(t, hyp(t, "add", "exp", "--parent", s.hA, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--aims", limB, "--why", "w"), 1, crossGoals)
	expOf(t, s.hB, limB)
}

// The replay half: a log whose Experiment of Goal A aims at Goal B's limit
// rejects it.
func TestCrossGoalAimRejectedReplay(t *testing.T) {
	s := setupTwoGoals(t)
	limB := limitOf(t, s.qB)
	e := rawEvent(t, "a", event.TypeNodeAdd, "", map[string]any{"kind": event.KindExperiment, "parent": fullID(t, s.hA), "title": "E",
		"recipe": "none", "rubric": "reading", "aims": []string{limB}}, "w")
	wantRejected(t, e, "bad aims: ")
	wantRejected(t, e, crossGoals)
}

// A Result closes only a key of its Experiment's Goal: a Result of Goal A
// that closes Goal B's line is refused with exit 1.
func TestCrossGoalCloseRefusedWrite(t *testing.T) {
	s := setupTwoGoals(t)
	e := expOf(t, s.hA)
	want(t, hyp(t, "record", "--exp", e, "--verdict", s.hA+"=supports", "--observation", "ran", "--closes", s.kB, "--why", "w"), 1, "bad closes: ")
	want(t, hyp(t, "record", "--exp", e, "--verdict", s.hA+"=supports", "--observation", "ran", "--closes", s.kB, "--why", "w"), 1, crossGoals)
	must(t, "record", "--exp", e, "--verdict", s.hA+"=supports", "--observation", "ran", "--closes", s.kA, "--why", "w")
}

// The replay half: a log whose Result of Goal A closes Goal B's line rejects
// it, and the line has no closer.
func TestCrossGoalCloseRejectedReplay(t *testing.T) {
	s := setupTwoGoals(t)
	e := expOf(t, s.hA)
	r := rawResult(t, "a", map[string]any{"parent": fullID(t, e), "verdicts": map[string]string{fullID(t, s.hA): "supports"},
		"observation": "ran", "closes": []string{fullKey(t, s.kB)}})
	wantRejected(t, r, "bad closes: ")
	wantRejected(t, r, crossGoals)
	if out := hyp(t, "show", s.kB).stdout; strings.Contains(out, "ran in ") {
		t.Fatalf("a rejected close ran %s:\n%s", s.kB, out)
	}
}

// An Experiment applies only a reading of its own Goal: it applies the
// readings it aims at, so one of Goal A that aims at Goal B's answered
// reading is refused with exit 1.
func TestCrossGoalApplyRefusedWrite(t *testing.T) {
	s := setupTwoGoals(t)
	rdB := answeredReading(t, s.qB)
	want(t, hyp(t, "add", "exp", "--parent", s.hA, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--aims", rdB, "--why", "w"), 1, "bad applies: ")
	want(t, hyp(t, "add", "exp", "--parent", s.hA, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--aims", rdB, "--why", "w"), 1, crossGoals)
	expOf(t, s.hB, rdB)
}

// The replay half: a log whose Experiment of Goal A aims at Goal B's reading
// to apply it rejects it.
func TestCrossGoalApplyRejectedReplay(t *testing.T) {
	s := setupTwoGoals(t)
	rdB := fullKey(t, answeredReading(t, s.qB))
	e := rawEvent(t, "a", event.TypeNodeAdd, "", map[string]any{"kind": event.KindExperiment, "parent": fullID(t, s.hA), "title": "E",
		"recipe": "none", "rubric": "reading", "aims": []string{rdB}}, "w")
	wantRejected(t, e, "bad applies: ")
	wantRejected(t, e, crossGoals)
}

// mergeStores copies the events files of dirB into dirA's store.
func mergeStores(t *testing.T, dirA, dirB string) {
	t.Helper()
	files, err := os.ReadDir(filepath.Join(dirB, ".tree", "events"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(dirB, ".tree", "events", f.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(dirA, ".tree", "events", f.Name())); err == nil {
			t.Fatalf("both stores have %s", f.Name())
		}
		if err := os.WriteFile(filepath.Join(dirA, ".tree", "events", f.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// hyp status G-… prints one Goal: in a two-Goal store, each Goal's status is
// byte-identical to what its own events give in a store of their own, last
// checkpoint and RECENT included. hyp status prints a GOALS table and one
// AWAITING OWNER list instead of each Goal's FRONTIER and OWED.
func TestStatusOfOneGoalInATwoGoalStore(t *testing.T) {
	root := t.TempDir()
	dirA, dirB := filepath.Join(root, "a"), filepath.Join(root, "b")
	for _, d := range []string{dirA, dirB} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	gA, kA := buildGoalStore(t, dirA, "ra", "oa", "GA", "one tenant per row")
	must(t, "checkpoint", "--refers", kA+"=no test bed holds two tenants", "--summary", "A refers its line", "--why", "w")
	ownA := hyp(t, "status", "--max-lines", "0").stdout
	gB, _ := buildGoalStore(t, dirB, "rb", "ob", "GB", "survives a region outage")
	must(t, "checkpoint", "--summary", "B stops", "--why", "w")
	ownB := hyp(t, "status", "--max-lines", "0").stdout

	mergeStores(t, dirA, dirB)
	t.Chdir(dirA)
	for _, c := range []struct{ g, own string }{{gA, ownA}, {gB, ownB}} {
		r := hyp(t, "status", c.g, "--max-lines", "0")
		if r.code != 0 || r.stdout != c.own {
			t.Fatalf("hyp status %s in the two-Goal store (exit %d):\n%s\nin a store of its own:\n%s%s", c.g, r.code, r.stdout, c.own, r.stderr)
		}
	}
	out := hyp(t, "status").stdout
	for _, s := range []string{"GOALS", "\n  " + gA + " GA — in_progress", "; runner owes ", "; checkpoint ", "\n  " + gB + " GB — in_progress", "AWAITING OWNER\n  " + gA + " GA\n    "} {
		if !strings.Contains(out, s) {
			t.Fatalf("hyp status in a store with many Goals does not print %q:\n%s", s, out)
		}
	}
	if strings.Count(out, "AWAITING OWNER") != 1 || strings.Contains(out, "FRONTIER") || strings.Contains(out, "\nOWED") {
		t.Fatalf("hyp status in a store with many Goals prints more than a GOALS table and one AWAITING OWNER list:\n%s", out)
	}
	if !strings.Contains(out, "AWAITING OWNER\n  "+gA+" GA\n    Q-") || !strings.Contains(out, " limit referred, accept or return "+kA) {
		t.Fatalf("AWAITING OWNER does not list %s's referred line under it:\n%s", gA, out)
	}
	want(t, hyp(t, "status", kA), 2, "")
}

// A Goal's last checkpoint is its own runner's latest: another Goal's
// checkpoint and an old owner checkpoint on the Goal are not it.
func TestLastCheckpointIsPerGoalAndTheRunners(t *testing.T) {
	s := setupTwoGoals(t)
	must(t, "checkpoint", "--goal", s.gA, "--summary", "A's runner stops", "--why", "w")
	rawEvent(t, "o", event.TypeCheckpoint, fullID(t, s.gA), map[string]any{"summary": "an owner's old checkpoint", "role": "owner"}, "w")
	want(t, hyp(t, "status", s.gA), 0, "CHECKPOINT ")
	if out := hyp(t, "status", s.gA).stdout; !strings.Contains(out, ": A's runner stops\n") {
		t.Fatalf("%s's last checkpoint is not its runner's latest:\n%s", s.gA, out)
	}
	want(t, hyp(t, "status", s.gB), 0, "CHECKPOINT none\n")
}
