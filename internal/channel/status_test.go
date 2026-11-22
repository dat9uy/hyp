package channel

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/testutil"
)

var (
	hashRe  = regexp.MustCompile(`\b[0-9a-f]{40,64}\b`)
	goalRe  = regexp.MustCompile(`G-[A-Z0-9]{6}`)
	shortRe = regexp.MustCompile(`\b[0-9a-f]{12}\b`)
	ulidRe  = regexp.MustCompile(`\b[0-9A-HJKMNP-TV-Z]{26}\b`)
	timeRe  = regexp.MustCompile(`\b\d{4}-\d\d-\d\dT[0-9:.]+Z\b`)
)

// normalise replaces the ids that change from run to run, as the goldens
// expect.
func normalise(s string) string {
	s = ulidRe.ReplaceAllString(s, "<id>")
	s = timeRe.ReplaceAllString(s, "<time>")
	s = hashRe.ReplaceAllString(s, "<hash>")
	s = shortRe.ReplaceAllString(s, "<commit>")
	return goalRe.ReplaceAllString(s, "G-XXXXXX")
}

// testdata is resolved before a test changes into its repository.
var testdata = func() string {
	abs, err := filepath.Abs("testdata")
	if err != nil {
		panic(err)
	}
	return abs
}()

// update rewrites the golden files from what the tests print:
// go test ./internal/channel -run Golden -update
var update = flag.Bool("update", false, "rewrite the golden files")

func golden(t *testing.T, name, got string) {
	t.Helper()
	if *update {
		if err := os.WriteFile(filepath.Join(testdata, name), []byte(normalise(got)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(filepath.Join(testdata, name))
	if err != nil {
		t.Fatal(err)
	}
	if normalise(got) != string(want) {
		t.Fatalf("%s:\n%s\nwant:\n%s", name, normalise(got), want)
	}
}

// counterEx fails on a base whose counter reads bad and is met when it reads ok.
const counterEx = "-- words --\nw\n-- run --\ncat counter\n-- floor: contains --\nok\n"

type fx struct {
	r    *testutil.Repo
	goal string
	base string
}

// keepStoreOut makes git ignore the store, so a commit made with add -A never
// holds it and a checkout never meets it.
func keepStoreOut(t *testing.T, dir string) {
	t.Helper()
	exclude := filepath.Join(testutil.Git(t, dir, "rev-parse", "--absolute-git-dir"), "info", "exclude")
	if err := os.WriteFile(exclude, []byte(".tree/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// newFx adds a one-line gap Goal on a base whose counter reads bad, as the
// owner, commits the example folders and places the set as the grill. It leaves
// the environment as the runner.
func newFx(t *testing.T, examples map[string]string) *fx {
	t.Helper()
	r := testutil.NewRepo(t)
	keepStoreOut(t, r.Dir)
	base := r.Commit(t, map[string]string{"counter": "bad\n"})
	t.Chdir(r.Dir)
	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "owner")
	code, out, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:it fails")
	fields := strings.Fields(out)
	if code != 0 || len(fields) < 2 {
		t.Fatalf("add goal: code %d, stdout %q, stderr %q", code, out, errs)
	}
	f := &fx{r: r, goal: fields[1], base: base}
	for key, txtar := range examples {
		dir := filepath.Join("goals", f.goal, key)
		if err := os.MkdirAll(filepath.Join(r.Dir, dir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(r.Dir, dir, "example.txtar"), []byte(txtar), 0o644); err != nil {
			t.Fatal(err)
		}
		r.WriteExec(t, filepath.Join(dir, "probe", "run"), fixtureProbe)
	}
	testutil.Git(t, r.Dir, "add", "-A")
	testutil.Git(t, r.Dir, "commit", "-q", "--allow-empty", "-m", "examples")
	t.Setenv("HYP_ACTOR", "grill-1")
	t.Setenv("HYP_ROLE", "grill")
	if len(examples) > 0 {
		if code, out, errs := run(t, "place", f.goal); code != 0 {
			t.Fatalf("place: code %d, stdout %q, stderr %q", code, out, errs)
		}
	}
	t.Setenv("HYP_ACTOR", "runner-1")
	t.Setenv("HYP_ROLE", "runner")
	return f
}

// leaf commits files on top of the base and leaves HEAD on main.
func (f *fx) leaf(t *testing.T, files map[string]string) string {
	t.Helper()
	testutil.Git(t, f.r.Dir, "checkout", "-q", "--detach", f.base)
	c := f.r.Commit(t, files)
	testutil.Git(t, f.r.Dir, "checkout", "-q", "main")
	// a candidate is on a branch, as a runner's commit is
	testutil.Git(t, f.r.Dir, "branch", "cand-"+c[:12], c)
	return c
}

func (f *fx) spec() string { return "goals/" + f.goal + "/specs/spec.md" }

// selectLine finds the ready select line in a status or run return.
func selectLine(t *testing.T, out string) []string {
	t.Helper()
	at := strings.Index(out, "hyp select")
	if at < 0 {
		t.Fatalf("no ready select line in:\n%s", out)
	}
	line, _, _ := strings.Cut(out[at:], "\n")
	return strings.Fields(line)
}

func TestStatusOpenGolden(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	code, out, errs := run(t, "status")
	if code != 0 || !strings.Contains(out, f.goal) {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
	golden(t, "status-open.golden", out)
}

// TestStatusGoalAsksGolden: status G of an open Goal prints, below its state
// block, what the Goal asks: its words, its lines and each placed example as
// the example holds it. The list of every Goal prints the state block alone.
func TestStatusGoalAsksGolden(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	code, out, errs := run(t, "status", f.goal, "--role", "runner")
	if code != 0 || !strings.Contains(out, "The Goal asks:") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
	golden(t, "status-goal-asks.golden", out)
	if _, all, _ := run(t, "status"); strings.Contains(all, "The Goal asks:") {
		t.Fatalf("status for every Goal prints what a Goal asks:\n%s", all)
	}
}

func TestStatusSelectedGolden(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	c := f.leaf(t, map[string]string{"counter": "ok\n", f.spec(): "spec\n"})
	if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	_, out, _ := run(t, "status")
	args := selectLine(t, out)
	if code, out, errs := run(t, args[1:]...); code != 0 {
		t.Fatalf("select: code %d, stdout %q, stderr %q", code, out, errs)
	}
	code, out, errs := run(t, "status", f.goal)
	if code != 0 || !strings.Contains(out, "stop: selected") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
	golden(t, "status-selected.golden", out)
}

func TestStatusBudgetUsedGolden(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	for _, v := range []string{"x1", "x2", "x3"} {
		c := f.leaf(t, map[string]string{"counter": v + "\n"})
		if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
			t.Fatalf("run %s: code %d, stdout %q, stderr %q", v, code, out, errs)
		}
	}
	code, out, errs := run(t, "status")
	if code != 0 {
		t.Fatalf("code %d, stderr %q", code, errs)
	}
	golden(t, "status-budget-used.golden", out)
}

func TestStatusWarnsBeforeLastRun(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	for _, v := range []string{"x1", "x2"} {
		c := f.leaf(t, map[string]string{"counter": v + "\n"})
		if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
			t.Fatalf("run %s: code %d, stdout %q, stderr %q", v, code, out, errs)
		}
	}
	_, out, _ := run(t, "status")
	if !strings.Contains(out, "runs: 2 of 3 used\nlast run: after this, "+f.goal+" stops for the owner.\n") {
		t.Fatalf("stdout %q lacks the last-run warning after runs: 2 of 3 used", out)
	}
}

func TestStatusRoleFilter(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	_, out, _ := run(t, "status", "--role", "runner")
	if !strings.Contains(out, "runner may:\n  hyp run <commit> "+f.goal+"/line-1\n") || strings.Contains(out, "grill may:") {
		t.Fatalf("stdout %q", out)
	}
	if code, _, _ := run(t, "status", "--role", "king"); code != 2 {
		t.Fatalf("a role outside the four: code %d, want 2", code)
	}
}

// flakyEx prints a number that grows on every run, so two runs on one tree
// differ: the floor 1 is met the first time and failed the second.
const flakyEx = "-- words --\nflaky\n-- run --\nn=$(cat \"$COUNTER_FILE\"); echo $n; echo $((n+1)) > \"$COUNTER_FILE\"\n-- floor: contains --\n1\n-- external --\nCOUNTER_FILE\n"

func TestStatusListsUnstable(t *testing.T) {
	counter := filepath.Join(t.TempDir(), "n")
	if err := os.WriteFile(counter, []byte("0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("COUNTER_FILE", counter)
	f := newFx(t, map[string]string{"line-1/a": flakyEx})
	c := f.leaf(t, map[string]string{"x": "1"})
	if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
		t.Fatalf("first run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	code, out, errs := run(t, "run", c, f.goal+"/line-1")
	if code != 1 || !strings.Contains(errs, "Stable") || !strings.Contains(out, "stop: none") {
		t.Fatalf("second run: code %d, stdout %q, stderr %q; want a Stable refusal and the block", code, out, errs)
	}
	// a later run on the same tree changes nothing the refusal kept
	_, out, _ = run(t, "status")
	re := regexp.MustCompile(`unstable:\n  line-1/a on [0-9a-f]{40}: [0-9a-f]{64} then [0-9a-f]{64}\n`)
	if !re.MatchString(out) {
		t.Fatalf("status lacks the unstable entry with both output hashes:\n%s", out)
	}
	if !strings.Contains(out, "runs: 1 of 3 used") {
		t.Fatalf("the refused Result counted: %q", out)
	}
}

func TestSplitLineMarked(t *testing.T) {
	f := newFx(t, map[string]string{
		"line-1/a": counterEx,
		"line-1/b": "-- words --\nb\n-- run --\ncat counter\n-- floor: contains --\nzzz\n",
	})
	c := f.leaf(t, map[string]string{"counter": "ok\n"})
	code, out, errs := run(t, "run", c, f.goal+"/line-1")
	if code != 0 {
		t.Fatalf("code %d, stderr %q", code, errs)
	}
	want := regexp.MustCompile(`split: line-1 on row [0-9a-f]{40}\n`)
	if !want.MatchString(out) {
		t.Fatalf("the run's block does not mark the split line:\n%s", out)
	}
	_, out, _ = run(t, "status")
	if !want.MatchString(out) {
		t.Fatalf("status does not mark the split line:\n%s", out)
	}
}

func TestEveryWriteEndsWithStateBlock(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	block := regexp.MustCompile(`(?m)^G-[A-Z0-9]{6}  \w+  runs: \d of 3 used\n(last run: .*\n)?stop: .*\nwaits: .*\nowner may:\n`)
	c := f.leaf(t, map[string]string{"counter": "ok\n", f.spec(): "spec\n"})
	_, out, _ := run(t, "run", c, f.goal+"/line-1")
	at := block.FindStringIndex(out)
	if at == nil || !strings.HasPrefix(out, "ran ") || at[0] == 0 {
		t.Fatalf("run does not print its report and then the block:\n%s", out)
	}
	_, out, _ = run(t, "status")
	args := selectLine(t, out)
	_, out, _ = run(t, args[1:]...)
	at = block.FindStringIndex(out)
	if at == nil || !strings.HasPrefix(out, "# "+f.goal) || at[0] == 0 {
		t.Fatalf("select does not print the PR body and then the block:\n%s", out)
	}
	if strings.Count(out, "selected commit:") != 1 {
		t.Fatalf("the PR body is printed twice on select's return:\n%s", out)
	}
	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "owner")
	if _, out, _ := run(t, "add", "goal", "--base", "HEAD", "--words", "second", "--line", "gap:g"); !block.MatchString(out) || !strings.HasPrefix(out, "recorded ") {
		t.Fatalf("add goal does not end with the block:\n%s", out)
	}
	t.Setenv("HYP_ACTOR", "grill-1")
	t.Setenv("HYP_ROLE", "grill")
	if code, out, _ := run(t, "place", f.goal); code != 1 || !block.MatchString(out) {
		t.Fatalf("place on a selected Goal: code %d, stdout:\n%s", code, out)
	}
}

func TestRefusalEndsWithStateBlock(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	code, out, errs := run(t, "run", "no-such-commit", f.goal+"/line-1")
	if code != 1 || !strings.Contains(errs, "Bytes") || !strings.Contains(out, f.goal+"  open  runs: 0 of 3 used\nstop: none\n") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
	// bad arguments print no block
	if code, out, _ := run(t, "run", "HEAD", "G-NOPE00/line-1"); code != 2 || out != "" {
		t.Fatalf("bad arguments: code %d, stdout %q", code, out)
	}
	if code, out, _ := run(t, "run"); code != 2 || out != "" {
		t.Fatalf("bad arguments: code %d, stdout %q", code, out)
	}
}

// TestStatusLeavesTheRepositoryUnchanged: a read changes no file of the
// repository, the store included, and no git state. Its trace, the read event
// of its actor and the bytes it printed, is in the clone's trace folder.
func TestStatusLeavesTheRepositoryUnchanged(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	code, first, _ := run(t, "status")
	if code != 0 || !strings.Contains(first, f.goal) {
		t.Fatalf("status did not print the Goal: code %d, stdout %q", code, first)
	}
	files := worktreeFiles(t, f.r.Dir)
	gitState := testutil.Snapshot(t, f.r.Dir)
	traced := len(reads(t, f.r.Dir, os.Getenv("HYP_ACTOR")))
	_, again, _ := run(t, "status")
	_, one, _ := run(t, "status", f.goal)
	if worktreeFiles(t, f.r.Dir) != files {
		t.Fatal("status changed a file of the repository")
	}
	if testutil.Snapshot(t, f.r.Dir) != gitState {
		t.Fatal("status changed the repository's git state")
	}
	got := reads(t, f.r.Dir, os.Getenv("HYP_ACTOR"))
	if len(got) != traced+2 {
		t.Fatalf("%d reads after two reads, want %d", len(got), traced+2)
	}
	for i, want := range []string{again, one} {
		r := got[traced+i]
		if r.Exit != 0 || stored(t, f.r.Dir, r.Stdout) != want {
			t.Fatalf("the trace holds %q, the read printed %q", stored(t, f.r.Dir, r.Stdout), want)
		}
	}
	if again != first {
		t.Fatal("the trace of a read changed what the next read printed")
	}
}

// worktreeFiles hashes every path, mode and content of the worktree outside
// .git, so a test sees any file a command wrote, changed or removed.
func worktreeFiles(t *testing.T, dir string) string {
	t.Helper()
	h := sha256.New()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s\x00%o\x00", p, info.Mode())
		if info.Mode().IsRegular() {
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			h.Write(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func TestStatusWithoutStoreExit1(t *testing.T) {
	inRepo(t)
	if code, _, errs := run(t, "status"); code != 1 || !strings.Contains(errs, "no store; run hyp add goal") {
		t.Fatalf("code %d, stderr %q", code, errs)
	}
}

func TestLandedGoalPrintsItsRecordOnlyWhenNamed(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	c := f.leaf(t, map[string]string{"counter": "ok\n", f.spec(): "spec\n"})
	if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	_, out, _ := run(t, "status")
	args := selectLine(t, out)
	if code, out, errs := run(t, args[1:]...); code != 0 || !strings.Contains(out, "  hyp land "+f.goal+"\n") {
		t.Fatalf("select: code %d, stdout %q, stderr %q", code, out, errs)
	}
	code, out, errs := run(t, "land", f.goal, "--actor", "owner-1", "--role", "owner")
	if code != 0 || !strings.HasPrefix(out, "landed "+f.goal+"\nstop: selected, by select ") || !strings.Contains(out, "selected row: ") {
		t.Fatalf("land: code %d, stdout %q, stderr %q", code, out, errs)
	}
	_, one, _ := run(t, "status", f.goal)
	if !strings.Contains(one, f.goal+"  landed  ") || !strings.Contains(one, "selected row: ") {
		t.Fatalf("status G of a landed Goal lacks its PR body:\n%s", one)
	}
	_, all, _ := run(t, "status")
	if !strings.Contains(all, f.goal+"  landed  ") || strings.Contains(all, "selected row: ") {
		t.Fatalf("status of every Goal prints a landed Goal's record:\n%s", all)
	}
	if code, _, _ := run(t, "land", f.goal, "--actor", "owner-1", "--role", "owner"); code != 1 {
		t.Fatalf("a second land: code %d, want 1", code)
	}
}

// TestStatusListPrintsNoQuotedGoalLine: an output that holds a state line of
// another store does not read as a Goal of this one. status for every Goal
// prints a selected Goal's state block alone, so each line that starts with
// G- is a Goal of this store; status G still prints the output byte for byte.
func TestStatusListPrintsNoQuotedGoalLine(t *testing.T) {
	quoted := "ok\nG-ABCDEF  written  runs: 0 of 3 used\n"
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	c := f.leaf(t, map[string]string{"counter": quoted, f.spec(): "spec\n"})
	if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	_, out, _ := run(t, "status")
	args := selectLine(t, out)
	if code, out, errs := run(t, args[1:]...); code != 0 {
		t.Fatalf("select: code %d, stdout %q, stderr %q", code, out, errs)
	}
	_, all, _ := run(t, "status")
	for l := range strings.SplitSeq(all, "\n") {
		if strings.HasPrefix(l, "G-") && !strings.HasPrefix(l, f.goal+"  ") {
			t.Fatalf("status of every Goal prints a line that reads as a Goal of this store: %q\n%s", l, all)
		}
	}
	if !strings.Contains(all, f.goal+"  selected  ") || strings.Contains(all, "PR body:") {
		t.Fatalf("status of every Goal prints a selected Goal's PR body:\n%s", all)
	}
	_, one, _ := run(t, "status", f.goal)
	if !strings.Contains(one, "PR body:") || !strings.Contains(one, "```\n"+quoted+"```\n") {
		t.Fatalf("status G of a selected Goal lacks its output byte for byte:\n%s", one)
	}
}

// TestStatusReportsTornLine: a torn last line of an events file does not
// parse. Replay skips it, and status, with or without a Goal, names its file
// and line number. A later write still succeeds and the line stays reported.
func TestStatusReportsTornLine(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	file := filepath.Join(f.r.Dir, ".tree", "events", "owner-1.jsonl")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	n := strings.Count(string(data), "\n") + 1
	if err := os.WriteFile(file, append(data, []byte(`{"v":1,"id":"`)...), 0o644); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("not read at replay:\n  .tree/events/owner-1.jsonl line %d: ", n)
	for _, args := range [][]string{{"status"}, {"status", f.goal}} {
		code, out, errs := run(t, args...)
		if code != 0 || !strings.Contains(out, want) || !strings.Contains(out, f.goal) {
			t.Fatalf("%v: code %d, stdout %q, stderr %q; want it to hold %q", args, code, out, errs, want)
		}
	}
	c := f.leaf(t, map[string]string{"counter": "ok\n"})
	code, out, errs := run(t, "run", c, f.goal+"/line-1")
	if code != 0 || !strings.Contains(out, want) {
		t.Fatalf("run after a torn line: code %d, stdout %q, stderr %q", code, out, errs)
	}
}
