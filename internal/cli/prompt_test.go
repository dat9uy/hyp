package cli

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const promptBlock = "May change: this repository, in commits on one branch.\n" +
	"Read only: plans/.\n" +
	"Never: merge a PR.\n" +
	"Never: write under an owner's actor."

// promptStore makes a repository at a new temporary directory with a store
// under tree/, enters the store and returns the repository's path.
func promptStore(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(repo, "tree")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	return repo
}

func goalWithBody(t *testing.T, actor, body string) string {
	t.Helper()
	return must(t, "add", "goal", "--title", "G", "--body", body, "--actor", actor, "--why", "a purpose")
}

// For a Goal whose body holds a valid constraints block, hyp prompt prints
// the store's path, the actor, the Goal, the skill, the stop clauses and each
// block line byte for byte in body order, the same bytes on every call;
// --branch and --pr only add clauses after the block.
func TestPromptPrintsTheGoalsRunPrompt(t *testing.T) {
	promptStore(t)
	g := goalWithBody(t, "o", "Builds the design.\n"+promptBlock+"\n")
	r := hyp(t, "prompt", g, "--actor", "claude-run-1")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
	}
	wantOut := "/goal Run the hyp tree in tree with the hyp-goal-run skill, as HYP_ACTOR=claude-run-1. Its Goal is " + g +
		": read it with hyp show " + g + ", read hyp status " + g + ", and name --goal " + g + " on every hyp checkpoint.\n" +
		"Done when the latest printed hyp status " + g + " shows \"runner owes nothing: stop and report to the owner\" under OWED, " +
		"or 5 Results have been recorded under claude-run-1, or you cannot take the next step honestly and say why.\n" +
		promptBlock + "\n"
	if r.stdout != wantOut {
		t.Fatalf("hyp prompt printed\n%s\nwant\n%s", r.stdout, wantOut)
	}
	if again := hyp(t, "prompt", g, "--actor", "claude-run-1"); again.stdout != r.stdout {
		t.Fatalf("two calls on one log printed different bytes:\n%s\n---\n%s", r.stdout, again.stdout)
	}
	want(t, hyp(t, "prompt", g, "--actor", "claude-run-1", "--experiments", "3"), 0, "or 3 Results have been recorded under claude-run-1,")

	full := hyp(t, "prompt", g, "--actor", "claude-run-1", "--branch", "feat/x", "--pr", "7")
	if full.code != 0 || full.stdout != r.stdout+"Commit on the branch feat/x.\nThe run's PR is #7.\n" {
		t.Fatalf("--branch and --pr do not only add clauses after the block:\n%s", full.stdout)
	}
	// A store at the repository root prints its path as ".".
	t.Chdir(t.TempDir())
	if err := os.Mkdir(".git", 0o755); err != nil {
		t.Fatal(err)
	}
	must(t, "init")
	g2 := goalWithBody(t, "o", promptBlock)
	want(t, hyp(t, "prompt", g2, "--actor", "r"), 0, "/goal Run the hyp tree in . with the hyp-goal-run skill")
}

// hyp prompt prints no text of an owner answer: its output is byte-identical
// before and after an answer event is added to the log.
func TestPromptIgnoresOwnerAnswers(t *testing.T) {
	promptStore(t)
	g := goalWithBody(t, "o", promptBlock)
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	k := openReading(t, q)
	lim := limitOf(t, q)
	must(t, "checkpoint", "--refers", lim+"=void for every candidate", "--why", "w")
	before := hyp(t, "prompt", g, "--actor", "r")
	must(t, "answer", "--answer", k+"=code only, never the docs", "--accept", lim+"=not this Goal's case", "--actor", "o")
	after := hyp(t, "prompt", g, "--actor", "r")
	if before.code != 0 || before.stdout != after.stdout {
		t.Fatalf("an answer event changed the prompt:\n%s\n---\n%s", before.stdout, after.stdout)
	}
	for _, s := range []string{"code only", "not this Goal's case"} {
		if strings.Contains(after.stdout, s) {
			t.Fatalf("the prompt prints the owner's answer %q:\n%s", s, after.stdout)
		}
	}
}

// The constraints block runs from the first heading line to the end of the
// body. hyp prompt refuses a Goal with no block, a block missing a heading,
// or a block holding a line that starts with no heading, with exit 1, empty
// stdout and a message naming each problem; hyp add goal admits such a body
// and warns only when it has a heading but no valid block.
func TestPromptRefusesAnInvalidConstraintsBlock(t *testing.T) {
	promptStore(t)
	for _, c := range []struct {
		body, msg string
	}{
		{"Builds the design.", `the Goal body has no constraints block: no line starts with "May change:", "Read only:" or "Never:"`},
		{"May change: the code.\nNever: merge.", `the constraints block has no "Read only:" line`},
		{"Read only: plans/.", `the constraints block has no "May change:" line; the constraints block has no "Never:" line`},
		{"May change: the code.\nRead only: plans/.\nand the docs.\nNever: merge.", `the constraints block holds a line that starts with no heading: "and the docs."`},
		{"Builds it.\nMay change: the code.\nRead only: plans/.\nNever: merge.\n\nThe end.", `starts with no heading: "The end."`},
	} {
		r := hyp(t, "add", "goal", "--title", "G", "--body", c.body, "--actor", "o", "--why", "w")
		if r.code != 0 {
			t.Fatalf("hyp add goal refused body %q: exit %d, %q", c.body, r.code, r.stderr)
		}
		hasHeading := !strings.HasPrefix(c.msg, "the Goal body has no constraints block")
		if warned := strings.Contains(r.stderr, "warn: hyp prompt will refuse this Goal: "); warned != hasHeading {
			t.Fatalf("hyp add goal on body %q: warned %v, want %v; stderr %q", c.body, warned, hasHeading, r.stderr)
		}
		g := strings.Fields(r.stdout)[0]
		p := hyp(t, "prompt", g, "--actor", "r")
		if p.code != 1 || p.stdout != "" || !strings.Contains(p.stderr, c.msg) {
			t.Fatalf("hyp prompt on body %q: exit %d (want 1), stdout %q (want empty), stderr %q (want %q)", c.body, p.code, p.stdout, p.stderr, c.msg)
		}
	}
	// A valid block, with an empty line inside it, draws no warning and prints.
	r := hyp(t, "add", "goal", "--title", "G", "--body", "May change: the code.\n\nRead only: plans/.\nNever: merge.\n", "--actor", "o", "--why", "w")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("a valid block: exit %d, stderr %q", r.code, r.stderr)
	}
	want(t, hyp(t, "prompt", strings.Fields(r.stdout)[0], "--actor", "r"), 0, "\nMay change: the code.\n\nRead only: plans/.\nNever: merge.\n")
}

// In a store with two Goals, where one answer event settles a key of each,
// hyp prompt G prints the same bytes as for G's events in a store of their
// own.
func TestPromptInATwoGoalStoreEqualsItsOwnStore(t *testing.T) {
	repo := promptStore(t)
	g1 := goalWithBody(t, "a", "First.\n"+promptBlock)
	q1 := must(t, "add", "dq", "--parent", g1, "--title", "Q", "--why", "ask")
	k := openReading(t, q1)
	r := hyp(t, "add", "goal", "--title", "G2", "--body", "Second.\n"+promptBlock, "--line", "survives a region outage", "--actor", "b", "--why", "w")
	if r.code != 0 {
		t.Fatalf("second Goal: %q", r.stderr)
	}
	g2 := strings.Fields(r.stdout)[0]
	t.Setenv("HYP_ACTOR", "b")
	must(t, "add", "dq", "--parent", g2, "--title", "Q2", "--why", "ask")
	must(t, "checkpoint", "--goal", g2, "--refers", g2+"#1=no region in the test bed", "--why", "w")
	must(t, "answer", "--answer", k+"=code only", "--accept", g2+"#1=not this Goal", "--actor", "o")
	two := hyp(t, "prompt", g1, "--actor", "r")
	if two.code != 0 {
		t.Fatalf("exit %d, %q", two.code, two.stderr)
	}

	// G1's events in a store of their own: actor a wrote them all, and the
	// owner's answer keeps only G1's key.
	own := t.TempDir()
	events := filepath.Join(own, "tree", ".tree", "events")
	for _, d := range []string{filepath.Join(own, ".git"), events} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	src := filepath.Join(repo, "tree", ".tree", "events")
	b, err := os.ReadFile(filepath.Join(src, "a.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(events, "a.jsonl"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join(src, "o.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var kept []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var ev map[string]any
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			t.Fatal(err)
		}
		if ev["type"] == "answer" {
			delete(ev["payload"].(map[string]any), "accepts")
		}
		line, _ := json.Marshal(ev)
		kept = append(kept, string(line))
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(events, "o.jsonl"), []byte(strings.Join(kept, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(own, "tree"))
	want(t, hyp(t, "check"), 0, "check: 0 errors")
	if out := hyp(t, "status").stdout; strings.Contains(out, "GOALS") {
		t.Fatalf("the store of its own holds more than one Goal:\n%s", out)
	}
	one := hyp(t, "prompt", g1, "--actor", "r")
	if one.code != 0 || one.stdout != two.stdout {
		t.Fatalf("the two-Goal store and G's own store print different prompts:\n%s\n---\n%s", two.stdout, one.stdout)
	}
}

// hyp prompt with no --actor exits 2 and never falls back to $HYP_ACTOR; an
// unknown id or a non-Goal id exits 2.
func TestPromptBadArguments(t *testing.T) {
	promptStore(t)
	g := goalWithBody(t, "o", promptBlock)
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	t.Setenv("HYP_ACTOR", "claude-env")
	want(t, hyp(t, "prompt", g), 2, "missing --actor")
	want(t, hyp(t, "prompt", "G-NOSUCH", "--actor", "r"), 2, "unknown node G-NOSUCH")
	want(t, hyp(t, "prompt", q, "--actor", "r"), 2, "hyp prompt takes a Goal: "+q+" is a DesignQuestion")
	want(t, hyp(t, "prompt", "--actor", "r"), 2, "usage: hyp prompt")
	want(t, hyp(t, "prompt", g, "--actor", "r", "--experiments", "0"), 2, "--experiments must be at least 1")
	want(t, hyp(t, "prompt", g, "--actor", "Bad Actor"), 2, "invalid actor")
}

// The fixture is checkpoint-store:G-D1CH1R's events through the owner answer
// of 2026-11-11 14:31, its Goal body given a constraints block. Its prompt is
// the golden file an eval case compares, clause by clause, with run 6's
// hand-written prompt; the prompt is the same with or without that answer.
func TestPromptForTheCheckpointStoreFixture(t *testing.T) {
	repo := t.TempDir()
	if err := os.CopyFS(repo, os.DirFS("testdata/prompt-fixture")); err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile(filepath.Join(repo, "prompt.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(repo, "investigations", "checkpoint-store"))
	args := []string{"prompt", "G-D1CH1R", "--actor", "claude-g1b-6", "--branch", "feat/one-subject-per-run", "--pr", "132"}
	r := hyp(t, args...)
	if r.code != 0 || r.stdout != string(golden) {
		t.Fatalf("exit %d, stderr %q; printed\n%s\nwant\n%s", r.code, r.stderr, r.stdout, golden)
	}
	// Remove the owner answer, the last owner event: the prompt is unchanged.
	path := filepath.Join(".tree", "events", "owner-dat.jsonl")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	if !strings.Contains(lines[len(lines)-1], `"type":"answer"`) {
		t.Fatalf("the fixture's last owner event is not the answer: %s", lines[len(lines)-1])
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines[:len(lines)-1], "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if again := hyp(t, args...); again.stdout != string(golden) {
		t.Fatalf("the prompt changed without the owner answer:\n%s", again.stdout)
	}
}
