package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// goalLines writes Goal g's lines as owner o, naming the Goal, and returns
// their keys.
func goalLines(t *testing.T, g, o string, texts ...string) []string {
	t.Helper()
	args := []string{"checkpoint", "--goal", g, "--summary", "the lines", "--role", "owner", "--actor", o, "--why", "w"}
	for _, l := range texts {
		args = append(args, "--line", l)
	}
	r := hyp(t, args...)
	if r.code != 0 {
		t.Fatalf("lines of %s: exit %d, stderr %q", g, r.code, r.stderr)
	}
	keys := make([]string, len(texts))
	for i := range texts {
		keys[i] = g + "#" + string(rune('1'+i))
	}
	return keys
}

// goalBlock is the part of status output that belongs to Goal g: from its
// GOAL line up to the next GOAL line or AWAITING OWNER.
func goalBlock(t *testing.T, out, g string) string {
	t.Helper()
	start := strings.Index(out, "GOAL "+g+" ")
	if start < 0 {
		t.Fatalf("status prints no GOAL line for %s:\n%s", g, out)
	}
	rest := out[start:]
	end := len(rest)
	for _, mark := range []string{"\nGOAL ", "\nAWAITING OWNER"} {
		if i := strings.Index(rest, mark); i >= 0 && i+1 < end {
			end = i + 1
		}
	}
	return rest[:end]
}

// One store holds many Goals. A second Goal is a write hyp admits, and
// each Goal's OWED, lines and coverage hold only its own items: a line of
// one Goal is never owed on the other Goal's question, and a Result that
// covers one Goal's candidate leaves the other Goal's line owed.
func TestTwoGoalStoreAdmitsASecondGoalWrite(t *testing.T) {
	_, g1, q1 := setup(t)
	r := hyp(t, "add", "goal", "--title", "G2", "--why", "a second purpose")
	if r.code != 0 {
		t.Fatalf("one store holds many Goals, but a second Goal was refused: exit %d, stderr %q", r.code, r.stderr)
	}
	g2 := strings.Fields(r.stdout)[0]
	q2 := must(t, "add", "dq", "--parent", g2, "--title", "Q2", "--why", "ask")

	want(t, hyp(t, "checkpoint", "--summary", "s", "--why", "w"), 2, "missing --goal: the store holds 2 Goals")
	k1 := goalLines(t, g1, "o", "one tenant per row")
	k2 := goalLines(t, g2, "o", "survives a region outage")
	if k1[0] == k2[0] {
		t.Fatalf("both Goals' first lines have key %s", k1[0])
	}
	h1 := supported(t, q1, "H1")
	h2 := supported(t, q2, "H2")

	out := hyp(t, "status").stdout
	for _, s := range []string{
		"runner: " + q1 + " limit " + k1[0] + " [owed]",
		"runner: " + q2 + " limit " + k2[0] + " [owed]",
		"  " + k1[0] + " uncovered on " + q1 + ": " + h1 + " — one tenant per row\n",
		"  " + k2[0] + " uncovered on " + q2 + ": " + h2 + " — survives a region outage\n",
	} {
		if !strings.Contains(out, s) {
			t.Fatalf("status does not print %q:\n%s", s, out)
		}
	}
	for _, s := range []string{"runner: " + q1 + " limit " + k2[0], "runner: " + q2 + " limit " + k1[0], k1[0] + " uncovered on " + q2, k2[0] + " uncovered on " + q1} {
		if strings.Contains(out, s) {
			t.Fatalf("status mixes the Goals' items, printing %q:\n%s", s, out)
		}
	}
	if b := goalBlock(t, out, g1); strings.Contains(b, k2[0]) || strings.Contains(b, q2) || strings.Contains(b, h2) {
		t.Fatalf("%s's part of status holds %s's items:\n%s", g1, g2, b)
	}

	// A Result that runs g1's line on H1 covers g1's line on q1 only.
	e := must(t, "add", "exp", "--parent", h1, "--title", "runs the line", "--recipe", "none", "--decided-by", "reading", "--aims", k1[0], "--target", "none", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--observation", "ran", "--closes", k1[0], "--why", "w")
	notIn(t, "limit "+k1[0])
	want(t, hyp(t, "status"), 0, "runner: "+q2+" limit "+k2[0]+" [owed]")
	show := hyp(t, "show", g2).stdout
	if strings.Contains(show, k1[0]) || !strings.Contains(show, "line "+k2[0]+" uncovered on "+q2) {
		t.Fatalf("show %s does not hold only its own line:\n%s", g2, show)
	}
	want(t, hyp(t, "check"), 0, "check: 0 errors")
}

// buildGoalStore builds, in dir, a store with one Goal, one owner line, one
// question and one supported candidate, written by runner a and owner o.
func buildGoalStore(t *testing.T, dir, a, o, title, line string) (g, key string) {
	t.Helper()
	t.Chdir(dir)
	t.Setenv("HYP_ACTOR", a)
	must(t, "init")
	g = must(t, "add", "goal", "--title", title, "--why", "start")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q of "+title, "--why", "ask")
	key = goalLines(t, g, o, line)[0]
	supported(t, q, "H of "+title)
	return g, key
}

// A log whose events hold two Goals replays without a rejection, and each
// Goal's part of status, and each line's show, equals what the Goal's own
// events give in a store of their own.
func TestTwoGoalStoreReplaysEachGoalAsItsOwn(t *testing.T) {
	root := t.TempDir()
	dirA, dirB := filepath.Join(root, "a"), filepath.Join(root, "b")
	for _, d := range []string{dirA, dirB} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	gA, kA := buildGoalStore(t, dirA, "ra", "oa", "GA", "one tenant per row")
	statusA, showA := hyp(t, "status").stdout, hyp(t, "show", kA).stdout
	gB, kB := buildGoalStore(t, dirB, "rb", "ob", "GB", "survives a region outage")
	statusB, showB := hyp(t, "status").stdout, hyp(t, "show", kB).stdout

	// The merged store is A's store with B's events files beside A's.
	for _, f := range []string{"rb.jsonl", "ob.jsonl"} {
		b, err := os.ReadFile(filepath.Join(dirB, ".tree", "events", f))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dirA, ".tree", "events", f), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dirA)
	if r := hyp(t, "check"); r.code != 0 || !strings.Contains(r.stdout+r.stderr, "check: 0 errors") {
		t.Fatalf("one store holds many Goals, but replay rejected the merged log: exit %d\n%s%s", r.code, r.stdout, r.stderr)
	}
	merged := hyp(t, "status").stdout
	if strings.Contains(merged, "REJECTED") {
		t.Fatalf("replay rejected events of the merged log:\n%s", merged)
	}
	for _, c := range []struct{ g, own string }{{gA, statusA}, {gB, statusB}} {
		if got, want := goalBlock(t, merged, c.g), goalBlock(t, c.own, c.g); got != want {
			t.Fatalf("%s in the two-Goal store:\n%s\nin a store of its own:\n%s", c.g, got, want)
		}
	}
	for _, c := range []struct{ k, own string }{{kA, showA}, {kB, showB}} {
		if got := hyp(t, "show", c.k).stdout; got != c.own {
			t.Fatalf("show %s in the two-Goal store:\n%s\nin a store of its own:\n%s", c.k, got, c.own)
		}
	}
}
