package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// histRepo is a git repository with two stores, investigations/a and
// investigations/b, each with one Goal, both tracked by git.
type histRepo struct {
	root, a, b string
	ga, gb     string // display ids of the Goals
}

// newHistRepo makes a histRepo. Goal titles are "Goal A" and "Goal B"; Goal
// A has the line "a case".
func newHistRepo(t *testing.T) *histRepo {
	t.Helper()
	r := &histRepo{root: t.TempDir()}
	t.Setenv("HYP_ACTOR", "a")
	gitIn(t, r.root, "init", "-q", "-b", "main")
	r.a, r.b = filepath.Join(r.root, "investigations", "a"), filepath.Join(r.root, "investigations", "b")
	for _, d := range []string{r.a, r.b} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(r.a)
	must(t, "init")
	r.ga = must(t, "add", "goal", "--title", "Goal A", "--line", "a case", "--why", "start")
	gitIn(t, r.root, "add", "-A")
	t.Chdir(r.b)
	must(t, "init")
	r.gb = must(t, "add", "goal", "--title", "Goal B", "--why", "start")
	gitIn(t, r.root, "add", "-A")
	return r
}

// in runs hyp in dir.
func in(t *testing.T, dir string, args ...string) result {
	t.Helper()
	t.Chdir(dir)
	return hyp(t, args...)
}

// mustIn runs hyp in dir; it must succeed. It returns the display id printed.
func mustIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	t.Chdir(dir)
	return must(t, args...)
}

// fullID returns the full id of the node whose display id is disp, from the
// history read at root.
func histID(t *testing.T, root, disp string) string {
	t.Helper()
	r := in(t, root, "events", "--json")
	var out struct{ Events []struct{ ID string } }
	if err := json.Unmarshal([]byte(r.stdout), &out); err != nil {
		t.Fatalf("events --json: %v: %s", err, r.stdout)
	}
	for _, ev := range out.Events {
		if strings.HasSuffix(ev.ID, disp[2:]) {
			return ev.ID
		}
	}
	t.Fatalf("no event id ends with %s", disp)
	return ""
}

// appendLine appends one raw events line to actor's events file in store dir.
func appendLine(t *testing.T, dir, actor, line string) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(dir, ".tree", "events", actor+".jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(line + "\n"); err != nil {
		t.Fatal(err)
	}
}

// rawEvent is one events line with id id, by actor, of type typ on node.
func rawLine(id, actor, typ, node, payload string) string {
	return fmt.Sprintf(`{"v":1,"id":%q,"ts":"2026-11-13T00:00:00Z","actor":%q,"type":%q,"node":%q,"payload":%s,"why":"w"}`, id, actor, typ, node, payload)
}

// questionOf returns the display id of Hypothesis h's question, read in dir.
func questionOf(t *testing.T, dir, h string) string {
	t.Helper()
	out := in(t, dir, "show", h).stdout
	i := strings.Index(out, "path: ")
	path := strings.Fields(strings.SplitN(out[i:], "\n", 2)[0]) // path: G > Q > H
	return path[len(path)-3]
}

// readingIn adds, under Goal g in store dir, a question, a Hypothesis with
// alias alias, an Experiment and a Result that judges it supports and
// records one reading and one limit. It returns the Hypothesis, the
// Experiment and the Result.
func readingIn(t *testing.T, dir, g, alias string) (h, x, r string) {
	t.Helper()
	q := mustIn(t, dir, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	h = mustIn(t, dir, "add", "hyp", "--parent", q, "--title", "H", "--alias", alias, "--why", "try")
	x = mustIn(t, dir, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "run")
	r = mustIn(t, dir, "record", "--exp", x, "--verdict", h+"=supports", "--observation", "ran", "--limit", "a case not run", "--reading", "line 1 has two meanings", "--why", "ran")
	return h, x, r
}

// TestStatusShowsTheFolderStoresGoals: hyp status run in a folder where store
// discovery finds a store shows only the Goals of the store in that folder.
func TestStatusShowsTheFolderStoresGoals(t *testing.T) {
	r := newHistRepo(t)
	ga2 := mustIn(t, r.a, "add", "goal", "--title", "Goal A2", "--why", "second")
	gitIn(t, r.root, "add", "-A")
	inA := in(t, r.a, "status")
	want(t, inA, 0, "GOALS")
	for _, g := range []string{r.ga, ga2} {
		want(t, inA, 0, g)
	}
	if strings.Contains(inA.stdout, r.gb) || strings.Contains(inA.stdout, "Goal B") {
		t.Errorf("status in store a shows Goal B of store b:\n%s", inA.stdout)
	}
	// Goal C is held by store b, whose events file holds its add, though
	// its line cites a line of store a's Goal A.
	gc := "7ZZZZZZZZZ0000000000000009"
	appendLine(t, r.b, "z", rawLine(gc, "z", "node.add", gc, `{"kind":"Goal","parent":"","title":"Goal C","lines":[{"text":"carries a case","cites":["`+histID(t, r.a, r.ga)+`#1"]}]}`))
	gitIn(t, r.root, "add", "-A")
	inB := in(t, r.b, "status")
	want(t, inB, 0, "GOALS")
	for _, g := range []string{r.gb + " Goal B", "G-000009 Goal C"} {
		if !strings.Contains(inB.stdout, g) {
			t.Errorf("status in store b does not show %s, a Goal its events file holds:\n%s", g, inB.stdout)
		}
	}
	if strings.Contains(inB.stdout, r.ga) || strings.Contains(inB.stdout, "Goal A") {
		t.Errorf("status in store b shows store a's Goals:\n%s", inB.stdout)
	}
	// A subfolder of a store is the store's folder too.
	sub := filepath.Join(r.b, "notes")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := in(t, sub, "status"); got != inB {
		t.Errorf("status in a subfolder of store b:\n%+v\nwant\n%+v", got, inB)
	}
}

// TestStatusAllListsTheHistory: hyp status --all, run anywhere in the
// repository, and plain hyp status where store discovery finds no store,
// list every Goal of the history and every item awaiting the owner, grouped
// by Goal; hyp status G-id prints what it prints in its store's folder.
func TestStatusAllListsTheHistory(t *testing.T) {
	r := newHistRepo(t)
	hb, _, rb := readingIn(t, r.b, r.gb, "h-b")
	gitIn(t, r.root, "add", "-A")
	other := filepath.Join(r.root, "docs")
	if err := os.Mkdir(other, 0o755); err != nil {
		t.Fatal(err)
	}
	all := in(t, r.root, "status", "--all")
	for _, dir := range []string{r.root, other} {
		for _, args := range [][]string{{"status"}, {"status", "--all"}} {
			if got := in(t, dir, args...); got != all {
				t.Errorf("hyp %v in %s:\n%+v\nwant hyp status --all at the root:\n%+v", args, dir, got, all)
			}
		}
	}
	for _, dir := range []string{r.a, r.b} {
		if got := in(t, dir, "status", "--all"); got != all {
			t.Errorf("hyp status --all in %s:\n%+v\nwant\n%+v", dir, got, all)
		}
	}
	want(t, all, 0, "GOALS")
	for _, s := range []string{r.ga + " Goal A", r.gb + " Goal B"} {
		want(t, all, 0, s)
	}
	owner := all.stdout[strings.Index(all.stdout, "AWAITING OWNER"):]
	if !strings.Contains(owner, r.gb+" Goal B") || !strings.Contains(owner, rb+"#r1") || strings.Index(owner, r.gb) > strings.Index(owner, rb+"#r1") {
		t.Errorf("AWAITING OWNER does not list %s#r1 under Goal B:\n%s", rb, owner)
	}
	inB := in(t, r.b, "status", r.gb)
	want(t, inB, 0, "GOAL "+r.gb)
	// A folder whose store adds nothing to the history is in the
	// repository too: a testdata store and a trial store under a .tree.
	var outside []string
	for _, rel := range []string{"internal/x/testdata/synth", "investigations/a/.tree/evidence/E-000001/trial"} {
		dir := filepath.Join(r.root, rel)
		copyStore(t, goalStore(t, "Left out"), dir)
		outside = append(outside, dir)
	}
	gitIn(t, r.root, "add", "-A")
	// An --as-of node of the Goal's store resolves there too.
	qb := questionOf(t, r.b, hb)
	asOf := in(t, r.b, "status", r.gb, "--as-of", qb)
	want(t, asOf, 0, "GOAL "+r.gb)
	// An --as-of alias resolves in the Goal's store, from any folder.
	asOfAlias := in(t, r.b, "status", r.gb, "--as-of", "h-b")
	want(t, asOfAlias, 0, "GOAL "+r.gb)
	for _, dir := range []string{r.a, r.root, outside[0]} {
		if got := in(t, dir, "status", r.gb, "--as-of", "h-b"); got != asOfAlias {
			t.Errorf("hyp status %s --as-of h-b in %s:\n%+v\nwant\n%+v", r.gb, dir, got, asOfAlias)
		}
	}
	for _, dir := range outside {
		if got := in(t, dir, "status", r.gb, "--as-of", qb); got != asOf {
			t.Errorf("hyp status %s --as-of %s in %s:\n%+v\nwant\n%+v", r.gb, qb, dir, got, asOf)
		}
	}
	for _, dir := range outside {
		want(t, in(t, dir, "status"), 0, "Left out")
		if got := in(t, dir, "status", "--all"); got != all {
			t.Errorf("hyp status --all in %s, whose store adds nothing to the history:\n%+v\nwant\n%+v", dir, got, all)
		}
	}
	for _, dir := range append([]string{r.root, r.a, other}, outside...) {
		if got := in(t, dir, "status", r.gb); got != inB {
			t.Errorf("hyp status %s in %s:\n%+v\nwant it as in its store's folder:\n%+v", r.gb, dir, got, inB)
		}
	}
}

// goalStore makes a store with one Goal titled title in a new folder outside
// any repository, and returns the store's folder.
func goalStore(t *testing.T, title string) string {
	t.Helper()
	dir := t.TempDir()
	mustIn(t, dir, "init")
	mustIn(t, dir, "add", "goal", "--title", title, "--why", "start")
	return dir
}

// copyStore copies the .tree of store src to dst/.tree.
func copyStore(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cp", "-R", filepath.Join(src, ".tree"), dst).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v: %s", err, out)
	}
}

// TestHistoryLeavesOutStores: a store adds nothing to the history when git
// tracks none of its events files, or when it sits under another store's
// .tree/, under a testdata/ directory, under a bundle's known-answers/, or
// under a directory below the repository root that holds its own .git.
// With no .git at or above the current directory, the history is the one
// store discovery finds.
func TestHistoryLeavesOutStores(t *testing.T) {
	r := newHistRepo(t)
	left := map[string]string{
		"untracked":     "investigations/untracked",
		"in a .tree":    "investigations/a/.tree/evidence/E-000001",
		"testdata":      "internal/x/testdata/synth",
		"known-answers": "recipes/boot@1/known-answers/case",
		"nested .git":   "vendor/sub/store",
	}
	for name, rel := range left {
		copyStore(t, goalStore(t, "Left out: "+name), filepath.Join(r.root, rel))
	}
	if err := os.WriteFile(filepath.Join(r.root, "recipes", "boot@1", "recipe.md"), []byte("---\nname: boot\nversion: 1\nformat: 1\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(r.root, "vendor", "sub", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Track every store but the untracked one; the nested repository's
	// files are added by force, past git's stop at its .git.
	gitIn(t, r.root, "add", "-A", "--", ".", ":!investigations/untracked")
	for _, f := range []string{"investigations/a/.tree/evidence/E-000001/.tree/events/a.jsonl", "vendor/sub/store/.tree/events/a.jsonl"} {
		gitIn(t, r.root, "update-index", "--add", "--cacheinfo", "100644,"+gitIn(t, r.root, "hash-object", "-w", f)+","+f)
	}
	tracked := gitIn(t, r.root, "ls-files")
	for _, name := range []string{"in a .tree", "testdata", "known-answers", "nested .git"} {
		if !strings.Contains(tracked, left[name]+"/.tree/events/a.jsonl") {
			t.Fatalf("setup: git does not track the store %s:\n%s", left[name], tracked)
		}
	}
	all := in(t, r.root, "status", "--all")
	want(t, all, 0, "Goal A")
	want(t, all, 0, "Goal B")
	if strings.Contains(all.stdout, "Left out") {
		t.Errorf("the history holds a store it leaves out:\n%s", all.stdout)
	}
	want(t, in(t, r.root, "check"), 0, "check: 0 errors")

	// No .git at or above: the history is the one store discovery finds,
	// and a store beside it adds nothing.
	plain := t.TempDir()
	copyStore(t, goalStore(t, "Here"), filepath.Join(plain, "here"))
	copyStore(t, goalStore(t, "Beside"), filepath.Join(plain, "beside"))
	here := in(t, filepath.Join(plain, "here"), "status", "--all")
	want(t, here, 0, "Here")
	if strings.Contains(here.stdout, "Beside") {
		t.Errorf("with no .git, the history holds a store beside the one discovery finds:\n%s", here.stdout)
	}
	want(t, in(t, plain, "status", "--all"), 3, "no .tree store found")
}

// TestCiteAcrossStores: a Goal line's --cite and a reading's of may name a
// key of a Goal in another store of the history; each is checked settled at
// its place in the history, and refused with exit 1 when it is not. hyp show
// given a Goal's id lists the Goals it cites and the Goals that cite it.
func TestCiteAcrossStores(t *testing.T) {
	r := newHistRepo(t)
	_, _, ra := readingIn(t, r.a, r.ga, "h-a")
	gitIn(t, r.root, "add", "-A")

	// Not settled: an open limit and an open reading of store a.
	want(t, in(t, r.b, "add", "goal", "--title", "G open", "--line", "x", "--cite", "1="+ra+"#1", "--why", "w"), 1, "#1, which is not settled")
	hb, xb, _ := readingIn(t, r.b, r.gb, "h-b")
	want(t, in(t, r.b, "record", "--exp", xb, "--verdict", hb+"=supports", "--observation", "o", "--reading-of", ra+"#r1=it means one", "--why", "w"), 1, "#r1 is open; the owner has not answered it")

	// Settled: a line of store a, and its reading once the owner answers it.
	gc := mustIn(t, r.b, "add", "goal", "--title", "Goal C", "--line", "carries a case", "--cite", "1="+r.ga+"#1", "--why", "carry")
	mustIn(t, r.root, "answer", "--answer", ra+"#r1=meaning one", "--actor", "owner")
	mustIn(t, r.b, "record", "--exp", xb, "--verdict", hb+"=supports", "--observation", "o", "--reading-of", ra+"#r1=it means one", "--why", "w")
	gitIn(t, r.root, "add", "-A")
	want(t, in(t, r.root, "check"), 0, "check: 0 errors")

	// A store hyp init just made, which git does not track yet, reads the
	// history from its own folder: its first Goal cites store a's line and
	// names store b's Goal in its body. From elsewhere it adds nothing.
	c := filepath.Join(r.root, "investigations", "c")
	if err := os.MkdirAll(c, 0o755); err != nil {
		t.Fatal(err)
	}
	mustIn(t, c, "init")
	fresh := mustIn(t, c, "add", "goal", "--title", "Fresh", "--line", "y", "--cite", "1="+r.ga+"#1", "--body", "see "+r.gb, "--why", "w")
	want(t, in(t, c, "show", fresh), 0, "cites: "+r.ga+" (store investigations/a) Goal A")
	if all := in(t, r.root, "status", "--all"); strings.Contains(all.stdout, "Fresh") {
		t.Errorf("an untracked store adds to the history read from the root:\n%s", all.stdout)
	}

	showA := in(t, r.b, "show", r.ga)
	want(t, showA, 0, "cited by: "+r.gb+" (store investigations/b) Goal B")
	want(t, showA, 0, "cited by: "+gc+" (store investigations/b) Goal C")
	want(t, in(t, r.a, "show", gc), 0, "cites: "+r.ga+" (store investigations/a) Goal A")
	want(t, in(t, r.a, "show", r.gb), 0, "cites: "+r.ga+" (store investigations/a) Goal A")

	// Replay checks a cite at its place in the history: a line that cites
	// store a's line, with an id before that line's Goal, is rejected.
	ga := histID(t, r.root, r.ga)
	early := "00000000000000000000000001"
	appendLine(t, r.b, "z", rawLine(early, "z", "node.add", early, `{"kind":"Goal","parent":"","title":"early","lines":[{"text":"x","cites":["`+ga+`#1"]}]}`))
	want(t, in(t, r.root, "check"), 1, "rejected "+early+" in store investigations/b: bad cite")
}

// TestAppendSortsAfterOtherStores: an appended event's id sorts after every
// event id the history holds, also an id another store holds, from a clock
// far ahead.
func TestAppendSortsAfterOtherStores(t *testing.T) {
	r := newHistRepo(t)
	future := "7ZZZZZZZZZ0000000000000000"
	appendLine(t, r.a, "z", rawLine(future, "z", "node.add", future, `{"kind":"DesignQuestion","parent":"`+histID(t, r.root, r.ga)+`","title":"ahead"}`))
	gitIn(t, r.root, "add", "-A")
	q := mustIn(t, r.b, "add", "dq", "--parent", r.gb, "--title", "after", "--why", "w")
	if id := histID(t, r.root, q); id <= future {
		t.Errorf("appended id %s in store b sorts at or before %s, an id of store a", id, future)
	}
	// A store git does not track yet appends after the history too.
	c := filepath.Join(r.root, "investigations", "c")
	if err := os.MkdirAll(c, 0o755); err != nil {
		t.Fatal(err)
	}
	mustIn(t, c, "init")
	gc := mustIn(t, c, "add", "goal", "--title", "fresh", "--why", "w")
	if id := histID(t, c, gc); id <= future {
		t.Errorf("appended id %s in an untracked store sorts at or before %s, an id of store a", id, future)
	}
}

// TestDuplicateIDAcrossStores: a history whose stores hold one event id twice
// is rejected on replay at the later line, ordered by id, then actor, then
// the store's path, then the line's bytes.
func TestDuplicateIDAcrossStores(t *testing.T) {
	r := newHistRepo(t)
	id := "7ZZZZZZZZZ0000000000000001"
	goal := func(title string) string { return `{"kind":"Goal","parent":"","title":"` + title + `"}` }
	// Same id and actor: the store's path orders them, so b's line is later.
	appendLine(t, r.a, "x", rawLine(id, "x", "node.add", id, goal("Dup in a")))
	appendLine(t, r.b, "x", rawLine(id, "x", "node.add", id, goal("Dup in b")))
	gitIn(t, r.root, "add", "-A")
	all := in(t, r.root, "status", "--all")
	want(t, all, 0, "Dup in a")
	if strings.Contains(all.stdout, "Dup in b") {
		t.Errorf("the later line of a twice-held id was applied:\n%s", all.stdout)
	}
	want(t, in(t, r.root, "check"), 1, "duplicate id: "+id)
	// The rejected line is store b's: only b's view counts it.
	want(t, in(t, r.b, "status"), 0, "REJECTED 1 events")
	if got := in(t, r.a, "status"); strings.Contains(got.stdout, "REJECTED") {
		t.Errorf("status in store a counts store b's rejected line:\n%s", got.stdout)
	}
	rejected := func(dir string) int {
		var out struct{ Rejections []struct{ Event string } }
		if err := json.Unmarshal([]byte(in(t, dir, "events", "--json").stdout), &out); err != nil {
			t.Fatal(err)
		}
		return len(out.Rejections)
	}
	if a, b := rejected(r.a), rejected(r.b); a != 0 || b != 1 {
		t.Errorf("events --json rejections: store a %d, store b %d; want 0 and 1", a, b)
	}

	// The actor orders before the store: under actor "w", b's line is first.
	id2 := "7ZZZZZZZZZ0000000000000002"
	appendLine(t, r.a, "x", rawLine(id2, "x", "node.add", id2, goal("Second in a")))
	appendLine(t, r.b, "w", rawLine(id2, "w", "node.add", id2, goal("Second in b")))
	gitIn(t, r.root, "add", "-A")
	all = in(t, r.root, "status", "--all")
	want(t, all, 0, "Second in b")
	if strings.Contains(all.stdout, "Second in a") {
		t.Errorf("actor w sorts before x, so a's line is the later one:\n%s", all.stdout)
	}
}

// TestAliasLocalToStore: two stores may hold the same alias. An alias
// resolves among the nodes of the store found from the current folder, and
// a cited verdict label among those of its Result's store; where it names
// nodes of two stores it is ambiguous node with exit 2. An alias used twice
// in one store is still refused at write and rejected on replay.
func TestAliasLocalToStore(t *testing.T) {
	r := newHistRepo(t)
	ha, _, _ := readingIn(t, r.a, r.ga, "h-null")
	hb, _, _ := readingIn(t, r.b, r.gb, "h-null")
	gitIn(t, r.root, "add", "-A")
	want(t, in(t, r.a, "show", "h-null"), 0, "Hypothesis "+ha)
	want(t, in(t, r.b, "show", "h-null"), 0, "Hypothesis "+hb)
	want(t, in(t, r.root, "show", "h-null"), 2, "ambiguous node h-null")
	// A write resolves the alias in its own store.
	want(t, in(t, r.b, "abandon", "h-null", "--why", "w"), 0, hb+" hypothesis.abandon")

	// A cited verdict label binds in the Result's store: b's Result cites
	// "verdict: h-null=contradicts" and typed supports for b's h-null.
	xb2 := mustIn(t, r.b, "add", "exp", "--parent", hb, "--title", "E2", "--recipe", "none", "--rubric", "reading", "--why", "w")
	ev := filepath.Join(r.b, ".tree", "evidence", xb2, "verdict.txt")
	if err := os.MkdirAll(filepath.Dir(ev), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ev, []byte("verdict: h-null=contradicts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rb := mustIn(t, r.b, "record", "--exp", xb2, "--verdict", hb+"=supports", "--observation", "o", "--evidence", ".tree/evidence/"+xb2+"/verdict.txt", "--why", "w")
	gitIn(t, r.root, "add", "-A")
	// Map order picks among same-named nodes at random; a label bound
	// outside b would miss the warning in about half the runs.
	for range 20 {
		want(t, in(t, r.root, "check"), 0, "warn: verdict mismatch "+rb+": "+hb+" (h-null) typed supports")
	}

	// Twice in one store: refused at write, rejected on replay.
	want(t, in(t, r.a, "add", "hyp", "--parent", questionOf(t, r.a, ha), "--title", "again", "--alias", "h-null", "--why", "w"), 1, "alias already used")
	qa := histID(t, r.root, questionOf(t, r.a, ha))
	id := "7ZZZZZZZZZ0000000000000003"
	appendLine(t, r.a, "z", rawLine(id, "z", "node.add", id, `{"kind":"Hypothesis","parent":"`+qa+`","title":"again","alias":"h-null"}`))
	want(t, in(t, r.root, "check"), 1, "alias already used")
}

// TestEventNeverCrossesStores: an event is appended to the store that holds
// its node. A node add whose parent, or a Result, select, edge, refer or
// checkpoint whose node or key, lies in another store is refused with exit
// 1 and rejected on replay; an answer whose keys lie in two stores, and hyp
// add goal where store discovery finds no store, are refused with exit 1,
// naming the rule.
func TestEventNeverCrossesStores(t *testing.T) {
	r := newHistRepo(t)
	ha, _, ra := readingIn(t, r.a, r.ga, "h-a")
	hb, xb, rb := readingIn(t, r.b, r.gb, "h-b")
	gitIn(t, r.root, "add", "-A")
	qb := questionOf(t, r.b, hb)

	before := treeFiles(t, filepath.Join(r.a, ".tree", "events"))
	for _, args := range [][]string{
		{"add", "dq", "--parent", r.gb, "--title", "q", "--why", "w"},
		{"add", "hyp", "--parent", qb, "--title", "h", "--why", "w"},
		{"record", "--exp", xb, "--verdict", hb + "=supports", "--observation", "o", "--why", "w"},
		{"link", ha, "depends_on", hb, "--why", "w"},
		{"supersede", hb, "--old", ha, "--why", "w"},
		{"checkpoint", "--goal", r.gb, "--summary", "s", "--why", "w"},
		{"checkpoint", "--goal", r.ga, "--refers", rb + "#1=void", "--why", "w"},
		{"select", qb, hb, "--why", "w"},
		{"abandon", hb, "--why", "w"},
	} {
		want(t, in(t, r.a, args...), 1, "event crosses stores")
	}
	want(t, in(t, r.root, "answer", "--answer", ra+"#r1=one", "--answer", rb+"#r1=two", "--actor", "owner"), 1, "an answer's keys lie in two stores")
	want(t, in(t, r.root, "add", "goal", "--title", "nowhere", "--why", "w"), 1, "hyp add goal needs a store")
	if after := treeFiles(t, filepath.Join(r.a, ".tree", "events")); !maps.EqualFunc(before, after, bytes.Equal) {
		t.Errorf("a refused write changed store a's events")
	}

	// Where discovery finds no store, an event goes to the store of its node.
	mustIn(t, r.root, "add", "dq", "--parent", r.gb, "--title", "from the root", "--why", "w")
	if data, _ := os.ReadFile(filepath.Join(r.b, ".tree", "events", "a.jsonl")); !bytes.Contains(data, []byte("from the root")) {
		t.Errorf("an add from the root did not go to store b, which holds its parent")
	}

	// Replay rejects each event a store's file holds on another store's node.
	full := func(d string) string { return histID(t, r.root, d) }
	gb, hbF, haF, xbF, rbF, qbF, gaF := full(r.gb), full(hb), full(ha), full(xb), full(rb), full(qb), full(r.ga)
	n := 0
	id := func() string { n++; return fmt.Sprintf("7ZZZZZZZZZ00000000000001%02d", n) }
	for _, ev := range []func(string) string{
		func(i string) string {
			return rawLine(i, "z", "node.add", i, `{"kind":"DesignQuestion","parent":"`+gb+`","title":"q"}`)
		},
		func(i string) string {
			return rawLine(i, "z", "result.record", i, `{"parent":"`+xbF+`","verdicts":{"`+hbF+`":"supports"},"observation":"o"}`)
		},
		func(i string) string { return rawLine(i, "z", "edge.add", haF, `{"type":"depends_on","to":"`+hbF+`"}`) },
		func(i string) string { return rawLine(i, "z", "checkpoint", gb, `{"summary":"s"}`) },
		func(i string) string {
			return rawLine(i, "z", "checkpoint", gaF, `{"summary":"s","refers":[{"key":"`+rbF+`#1","text":"void"}]}`)
		},
		func(i string) string {
			return rawLine(i, "z", "dq.select", qbF, `{"hypothesis":"`+hbF+`","prev_select":"","scope_change":false}`)
		},
		func(i string) string { return rawLine(i, "z", "hypothesis.abandon", hbF, `{}`) },
		func(i string) string {
			return fmt.Sprintf(`{"v":1,"id":%q,"ts":"2026-11-13T00:00:00Z","actor":"owner","type":"answer","node":"","payload":{"answers":[{"key":"%s#r1","text":"t"}]}}`, i, rbF)
		},
	} {
		i := id()
		appendLine(t, r.a, "z", ev(i))
		got := in(t, r.root, "check")
		if !strings.Contains(got.stdout, i[20:]+" in store investigations/a: event crosses stores") {
			t.Errorf("replay did not reject %s for crossing stores:\n%s", ev(i), got.stdout)
		}
	}
	// A legacy confirm in store b, on b's question, that accepts a line of
	// store a's Goal: its key's node lies in another store.
	ci := id()
	appendLine(t, r.b, "z", rawLine(ci, "z", "confirm", qbF, `{"select":"","role":"owner","verdict":"","accepts":["`+gaF+`#1"]}`))
	if got := in(t, r.root, "check"); !strings.Contains(got.stdout, ci[20:]+" in store investigations/b: event crosses stores") {
		t.Errorf("replay did not reject a confirm that accepts another store's line:\n%s", got.stdout)
	}
}

// TestSandboxHistoryCut: hyp sandbox create --cut EVENT without --store gives
// a clone where each store of the host's history holds, in its own folder,
// exactly the history's events at or before EVENT, and no evidence folder of
// an Experiment added after EVENT; a store of the base that holds no such
// event is not in the clone.
func TestSandboxHistoryCut(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	r := newHistRepo(t)
	// Store a: an Experiment before the cut and one after, each with an
	// evidence folder. Store b holds an event before the cut too.
	_, x1, _ := readingIn(t, r.a, r.ga, "h-a")
	mustIn(t, r.b, "add", "dq", "--parent", r.gb, "--title", "before", "--why", "w")
	// Store d: its Goal is the cut event; git tracks it, but the base does
	// not hold it.
	d := filepath.Join(r.root, "investigations", "d")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	mustIn(t, d, "init")
	gd := mustIn(t, d, "add", "goal", "--title", "Goal D", "--why", "start")
	gitIn(t, r.root, "add", "-A")
	cut := histID(t, r.root, gd)
	// After the cut: a second Experiment in a, more of b, and store c.
	x2 := mustIn(t, r.a, "add", "exp", "--parent", "h-a", "--title", "E2", "--recipe", "none", "--rubric", "reading", "--why", "after")
	for _, x := range []string{x1, x2} {
		p := filepath.Join(r.a, ".tree", "evidence", x, "run.log")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(x+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustIn(t, r.b, "add", "dq", "--parent", r.gb, "--title", "after", "--why", "w")
	c := filepath.Join(r.root, "investigations", "c")
	if err := os.MkdirAll(c, 0o755); err != nil {
		t.Fatal(err)
	}
	mustIn(t, c, "init")
	mustIn(t, c, "add", "goal", "--title", "Goal C after", "--why", "start")
	// The base holds a, b and c, and not d.
	gitIn(t, r.root, "add", "-A")
	gitIn(t, r.root, "reset", "-q", "--", "investigations/d")
	gitIn(t, r.root, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "base")
	sha := gitIn(t, r.root, "rev-parse", "HEAD")
	gitIn(t, r.root, "add", "-A")

	// The oracle: per store, each events line whose id sorts at or before
	// the cut, per actor file, in id order.
	cutLines := map[string]map[string][]byte{}
	for _, st := range []string{"a", "b", "c", "d"} {
		files, _ := filepath.Glob(filepath.Join(r.root, "investigations", st, ".tree", "events", "*.jsonl"))
		for _, f := range files {
			data, _ := os.ReadFile(f)
			var kept []string
			for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				var ev struct{ ID string }
				if json.Unmarshal([]byte(l), &ev) == nil && ev.ID <= cut {
					kept = append(kept, l)
				}
			}
			if len(kept) == 0 {
				continue
			}
			slices.Sort(kept)
			if cutLines[st] == nil {
				cutLines[st] = map[string][]byte{}
			}
			cutLines[st][filepath.Base(f)] = []byte(strings.Join(kept, "\n") + "\n")
		}
	}
	if got := slices.Sorted(maps.Keys(cutLines)); !slices.Equal(got, []string{"a", "b", "d"}) {
		t.Fatalf("setup: stores with events at or before the cut %v", got)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	stub := filepath.Join(t.TempDir(), "hyp")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho hyp 0.0.0-test\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	want(t, in(t, r.root, "sandbox", "create", "t1", "--host", r.root, "--base", sha, "--hyp", stub, "--recipe", "none", "--cut", cut), 0, "history cut at "+cut)
	wt := filepath.Join(home, ".hyp-sandbox", "t1", "wt")
	for st, wantFiles := range cutLines {
		got := treeFiles(t, filepath.Join(wt, "investigations", st, ".tree", "events"))
		if !maps.EqualFunc(got, wantFiles, bytes.Equal) {
			t.Errorf("clone store %s holds events %v, want the history's events at or before the cut %v", st, got, wantFiles)
		}
	}
	if _, err := os.Stat(filepath.Join(wt, "investigations", "a", ".tree", "evidence", x2)); !os.IsNotExist(err) {
		t.Errorf("clone holds the evidence of %s, added after the cut: %v", x2, err)
	}
	if _, err := os.Stat(filepath.Join(wt, "investigations", "a", ".tree", "evidence", x1, "run.log")); err != nil {
		t.Errorf("clone lacks the evidence of %s, added before the cut: %v", x1, err)
	}
	// Each store of the cut keeps its merge driver and ignore file.
	for _, f := range []string{".gitattributes", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(wt, "investigations", "a", ".tree", f)); err != nil {
			t.Errorf("clone store a lacks its .tree/%s: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(wt, "investigations", "c", ".tree")); !os.IsNotExist(err) {
		t.Errorf("store c of the base holds no event at or before the cut, yet is in the clone: %v", err)
	}
	// The clone's history is the cut: its status --all is the host's as of
	// the cut.
	inClone := in(t, wt, "status", "--all")
	asOf := in(t, r.root, "status", "--all", "--as-of", cut)
	if inClone.code != 0 || inClone != asOf {
		t.Errorf("status --all in the clone:\n%+v\nstatus --all --as-of %s in the host:\n%+v", inClone, cut, asOf)
	}
	want(t, in(t, r.root, "sandbox", "remove", "t1", "--discard"), 0, "removed t1; discarded")
}

// TestUntrackedStoreAllAndRoot: hyp status --all in a store git does not
// track yet lists the history's Goals and its own; a store at the
// repository root that git does not track, as in a clone a store cut
// wrote, never joins the history and is read alone, with no duplicate of
// a store below it.
func TestUntrackedStoreAllAndRoot(t *testing.T) {
	r := newHistRepo(t)
	c := filepath.Join(r.root, "investigations", "c")
	if err := os.MkdirAll(c, 0o755); err != nil {
		t.Fatal(err)
	}
	mustIn(t, c, "init")
	mustIn(t, c, "add", "goal", "--title", "Fresh", "--why", "w")
	all := in(t, c, "status", "--all")
	for _, s := range []string{"Goal A", "Goal B", "Fresh"} {
		want(t, all, 0, s)
	}

	// A copy of store a at the root, untracked, holds the same ids.
	if out, err := exec.Command("cp", "-R", filepath.Join(r.a, ".tree"), r.root).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v: %s", err, out)
	}
	want(t, in(t, r.root, "check"), 0, "check: 0 errors")
	root := in(t, r.root, "status")
	want(t, root, 0, "Goal A")
	if strings.Contains(root.stdout, "Goal B") {
		t.Errorf("an untracked root store joined the history:\n%s", root.stdout)
	}
}

// TestSameAskAcrossStores: a reading of another store that asks the same as
// the one answered is answered with its own hyp answer, since an answer's
// keys never lie in two stores.
func TestSameAskAcrossStores(t *testing.T) {
	r := newHistRepo(t)
	_, _, ra := readingIn(t, r.a, r.ga, "h-a")
	_, _, rb := readingIn(t, r.b, r.gb, "h-b")
	gitIn(t, r.root, "add", "-A")
	got := in(t, r.a, "answer", "--answer", ra+"#r1=one", "--actor", "owner")
	want(t, got, 0, "warn: "+rb+"#r1 asks the same as "+ra+"#r1 and stays open; it is in store investigations/b, so answer it with its own hyp answer --answer "+rb+"#r1")
}

// TestStoreByAnotherPath: a folder that reaches a store of the history by
// a symbolic link, or by another letter case where the file system ignores
// case, is that store: it joins the history once, not again as an
// untracked store.
func TestStoreByAnotherPath(t *testing.T) {
	r := newHistRepo(t)
	readingIn(t, r.a, r.ga, "h-a")
	gitIn(t, r.root, "add", "-A")
	link := filepath.Join(r.root, "alink")
	if err := os.Symlink(filepath.Join("investigations", "a"), link); err != nil {
		t.Fatal(err)
	}
	paths := []string{link}
	upper := filepath.Join(r.root, "INVESTIGATIONS", "a")
	if _, err := os.Stat(upper); err == nil {
		paths = append(paths, upper)
	}
	// A link from outside the repository reads the repository's history:
	// --all lists store b's Goal, and an append sorts after b's ids.
	outLink := filepath.Join(t.TempDir(), "alink")
	if err := os.Symlink(r.a, outLink); err != nil {
		t.Fatal(err)
	}
	want(t, in(t, outLink, "status", "--all"), 0, "Goal B")
	future := "7ZZZZZZZZZ0000000000000007"
	appendLine(t, r.b, "z", rawLine(future, "z", "node.add", future, `{"kind":"DesignQuestion","parent":"`+histID(t, r.root, r.gb)+`","title":"ahead"}`))
	gitIn(t, r.root, "add", "-A")
	q := mustIn(t, outLink, "add", "dq", "--parent", r.ga, "--title", "via link", "--why", "w")
	if id := histID(t, r.root, q); id <= future {
		t.Errorf("an append through an outside link has id %s, at or before %s, an id of store b", id, future)
	}
	inA := in(t, r.a, "status")
	for _, p := range paths {
		if got := in(t, p, "check"); !strings.Contains(got.stdout, "check: 0 errors") {
			t.Errorf("hyp check in %s, which reaches store a by another path:\n%+v", p, got)
		}
		if got := in(t, p, "status"); got != inA {
			t.Errorf("hyp status in %s:\n%+v\nwant it as in store a:\n%+v", p, got, inA)
		}
	}
}

// TestAppendSortsAfterUntracked: an append in a tracked store sorts after
// every id of a store git does not track yet, which the history read from
// the tracked store's folder does not hold, so the order of ids stays the
// order of appends once git tracks that store.
func TestAppendSortsAfterUntracked(t *testing.T) {
	r := newHistRepo(t)
	u := filepath.Join(r.root, "investigations", "u")
	if err := os.MkdirAll(u, 0o755); err != nil {
		t.Fatal(err)
	}
	mustIn(t, u, "init")
	ahead := "7ZZZZZZZZZ0000000000000005"
	appendLine(t, u, "z", rawLine(ahead, "z", "node.add", ahead, `{"kind":"Goal","parent":"","title":"ahead"}`))
	q := mustIn(t, r.b, "add", "dq", "--parent", r.gb, "--title", "after u", "--why", "w")
	if id := histID(t, r.root, q); id <= ahead {
		t.Errorf("appended id %s in store b sorts at or before %s, an id of an untracked store", id, ahead)
	}
}
