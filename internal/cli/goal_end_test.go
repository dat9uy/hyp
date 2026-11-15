package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
)

// The end of a Goal is an item of the owner's answer, keyed by the Goal's
// id. These tests drive it through the command line and raw events only, so
// that this file compiles on a commit that has no end.

// endStore is a store with Goal g, added by the owner o, with lines, and
// one question q; the runner is a.
func endStore(t *testing.T, lines ...string) (g, q string, keys []string) {
	t.Helper()
	t.Chdir(t.TempDir())
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	g, keys = addGoal(t, "G", "o", lines...)
	q = must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	return g, q, keys
}

// rawEnd appends an answer by actor that ends Goal g with text, as a merged
// log or another binary can hold it, and returns its id.
func rawEnd(t *testing.T, actor, g, text string, more map[string]any) string {
	t.Helper()
	p := map[string]any{"ends": []map[string]string{{"key": fullID(t, g), "text": text}}}
	for k, v := range more {
		p[k] = v
	}
	return rawEvent(t, actor, typeAnswer, "", p, "")
}

// Line 1: the actor that added G ends it with hyp answer --end G=TEXT, on G
// in progress, and hyp status and hyp show read G as ended and print TEXT.
func TestEndOfAGoalInProgress(t *testing.T) {
	g, _, _ := endStore(t)
	want(t, hyp(t, "status", g), 0, "— in_progress (")
	want(t, hyp(t, "answer", "--end", g+"=the design is outgrown", "--actor", "o"), 0, " answer\n")
	st := hyp(t, "status", g)
	want(t, st, 0, "GOAL "+g+" G — ended (")
	want(t, st, 0, " o: the design is outgrown\n")
	sh := hyp(t, "show", g)
	want(t, sh, 0, "status: ended\n")
	want(t, sh, 0, "the design is outgrown")
	if strings.Contains(st.stdout, "in_progress") || strings.Contains(sh.stdout, "in_progress") {
		t.Fatalf("an ended Goal still reads in_progress:\n%s\n%s", st.stdout, sh.stdout)
	}
}

// Line 1, on a complete Goal: the end reads ended in place of complete.
func TestEndOfACompleteGoal(t *testing.T) {
	g, q, _ := endStore(t)
	h := supported(t, q, "H")
	must(t, "select", q, h, "--why", "w")
	want(t, hyp(t, "status", g), 0, "— complete (")
	want(t, hyp(t, "answer", "--end", g+"=wrong question", "--actor", "o"), 0, " answer\n")
	want(t, hyp(t, "status", g), 0, "GOAL "+g+" G — ended (")
	want(t, hyp(t, "show", g), 0, "status: ended\n")
	if out := hyp(t, "status", g).stdout + hyp(t, "show", g).stdout; strings.Contains(out, "complete") {
		t.Fatalf("an ended Goal still reads complete:\n%s", out)
	}
}

// Line 2, the write: an end by any actor other than the one that added the
// Goal is refused with exit 1, naming the rule. A Goal a runner added is
// ended only under that runner's name.
func TestEndByAnotherActorRefusedWrite(t *testing.T) {
	g, _, _ := endStore(t)
	want(t, hyp(t, "answer", "--end", g+"=not mine", "--actor", "a"), 1, "only the actor that added a Goal ends it")
	want(t, hyp(t, "status", g), 0, "— in_progress (")
	gr := must(t, "add", "goal", "--title", "GR", "--why", "a runner's Goal")
	want(t, hyp(t, "answer", "--end", gr+"=not mine", "--actor", "o"), 1, "only the actor that added a Goal ends it")
	want(t, hyp(t, "answer", "--end", gr+"=outgrown", "--actor", "a"), 0, " answer\n")
}

// Line 2, the replay half: a log that holds an end by another actor
// rejects it with the rule's own reason, and the Goal does not end.
func TestEndByAnotherActorRejectedReplay(t *testing.T) {
	g, _, _ := endStore(t)
	id := rawEnd(t, "a", g, "not mine", nil)
	wantRejected(t, id, "only the actor that added a Goal ends it")
	want(t, hyp(t, "status", g), 0, "— in_progress (")
}

// Line 3, the write: a second end of G is refused with exit 1, and so is one
// answer that ends G and also moves a key of G.
func TestSecondOrMixedEndRefusedWrite(t *testing.T) {
	g, q, _ := endStore(t)
	k := openReading(t, q)
	want(t, hyp(t, "answer", "--end", g+"=outgrown", "--answer", k+"=code only", "--actor", "o"), 1, "bad end: one answer ends")
	want(t, hyp(t, "status", g), 0, "— in_progress (")
	want(t, hyp(t, "answer", "--end", g+"=outgrown", "--actor", "o"), 0, " answer\n")
	want(t, hyp(t, "answer", "--end", g+"=again", "--actor", "o"), 1, "")
	want(t, hyp(t, "status", g), 0, "o: outgrown\n")
}

// Line 3, the replay half: in a merged log the first end by id order
// applies; the later end and a mixed answer are rejected.
func TestSecondOrMixedEndRejectedReplay(t *testing.T) {
	g, q, _ := endStore(t)
	k := fullKey(t, openReading(t, q))
	mixed := rawEnd(t, "o", g, "mixed", map[string]any{"answers": []map[string]string{{"key": k, "text": "code only"}}})
	wantRejected(t, mixed, "bad end: one answer ends")
	first := rawEnd(t, "o", g, "first words", nil)
	second := rawEnd(t, "o", g, "second words", nil)
	wantRejected(t, second, "is already ended")
	if out := hyp(t, "check").stdout; strings.Contains(out, "rejected "+first) {
		t.Fatalf("the first end was rejected:\n%s", out)
	}
	st := hyp(t, "status", g).stdout
	if !strings.Contains(st, "o: first words") || strings.Contains(st, "second words") {
		t.Fatalf("the first end by id order does not apply:\n%s", st)
	}
	if out := hyp(t, "show", strings.Split(k, "#")[0]).stdout; !strings.Contains(out, " open — read the line as code-only") {
		t.Fatalf("the mixed answer answered the reading:\n%s", out)
	}
}

// Line 4: an ended Goal shows nothing that waits. Before the end, its runner
// owes an item, the owner a reading and a referred limit, and a question
// waits on both; after it, none of them shows, its GOALS row reads ended,
// and hyp show still prints its lines, readings, limits and Results.
func TestEndedGoalShowsNothingThatWaits(t *testing.T) {
	g, q, keys := endStore(t, "one tenant per row")
	lim := limitOf(t, q)
	k := openReading(t, q)
	must(t, "checkpoint", "--goal", g, "--refers", lim+"=no test bed", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "ask"), "--title", "H2", "--why", "w")
	// A sub-question is owed no line, so its turn waits only on its reading.
	openReading(t, must(t, "add", "dq", "--parent", h2, "--title", "QS", "--why", "ask"))
	g2, _ := addGoal(t, "G2", "o")
	must(t, "add", "dq", "--parent", g2, "--title", "Q3", "--why", "ask")

	before := hyp(t, "status", g).stdout
	for _, s := range []string{"runner: ", "reading:", "limit referred", "waits on "} {
		if !strings.Contains(before, s) {
			t.Fatalf("the store does not set up %q before the end:\n%s", s, before)
		}
	}
	must(t, "answer", "--end", g+"=outgrown", "--actor", "o")

	st := hyp(t, "status", g).stdout
	owed := st[strings.Index(st, "\nOWED\n"):strings.Index(st, "\nAWAITING OWNER\n")]
	if strings.Count(owed, "\n") != 2 || !strings.Contains(owed, "runner owes nothing") {
		t.Fatalf("OWED of an ended Goal lists an item:\n%s", st)
	}
	awaiting := st[strings.Index(st, "\nAWAITING OWNER\n")+len("\nAWAITING OWNER\n") : strings.Index(st, "\nRECENT\n")+1]
	if awaiting != "" {
		t.Fatalf("AWAITING OWNER of an ended Goal lists an item:\n%s", st)
	}
	if strings.Contains(st, "waits on ") {
		t.Fatalf("a question of an ended Goal waits:\n%s", st)
	}
	all := hyp(t, "status").stdout
	want(t, result{0, all, ""}, 0, "  "+g+" G — ended (")
	if i := strings.Index(all, "AWAITING OWNER"); strings.Contains(all[i:], g+" G") {
		t.Fatalf("the store's AWAITING OWNER lists the ended Goal:\n%s", all)
	}
	sh := hyp(t, "show", g).stdout
	for _, s := range []string{"line " + keys[0] + " ", "reading " + strings.Split(k, "#")[0]} {
		if !strings.Contains(sh, s) {
			t.Fatalf("hyp show of an ended Goal does not print %q:\n%s", s, sh)
		}
	}
	shq := hyp(t, "show", q).stdout
	if !strings.Contains(shq, "case C") || !strings.Contains(shq, "R-") {
		t.Fatalf("hyp show of an ended Goal's question does not print its limits and Results:\n%s", shq)
	}
}

// Line 5, the write: every write whose node leads to an ended Goal is
// refused with exit 1, naming the end; hyp add goal and a cite from another
// Goal stay admitted.
func TestWriteUnderAnEndedGoalRefused(t *testing.T) {
	g, q, _ := endStore(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--target", "none", "--why", "w")
	hs := supported(t, q, "HS")
	k := openReading(t, q)
	qd := must(t, "add", "dq", "--parent", g, "--title", "QD", "--why", "ask")
	must(t, "defer", qd, "--why", "park")
	if err := os.WriteFile("subject.txt", []byte("s"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := strings.Split(k, "#")[0]
	must(t, "answer", "--end", g+"=outgrown", "--actor", "o")

	for _, args := range [][]string{
		{"add", "dq", "--parent", g, "--title", "Q9", "--why", "w"},
		{"add", "hyp", "--parent", q, "--title", "H9", "--why", "w"},
		{"add", "exp", "--parent", h, "--title", "E9", "--recipe", "none", "--rubric", "reading", "--target", "none", "--why", "w"},
		{"record", "--exp", e, "--verdict", h + "=supports", "--observation", "o", "--why", "w"},
		{"retract", res, "--why", "w"},
		{"select", q, hs, "--why", "w"},
		{"defer", q, "--why", "w"},
		{"withdraw", q, "--why", "w"},
		{"resume", qd, "--why", "w"},
		{"subject", h, "subject.txt", "--why", "w"},
		{"link", h, "depends_on", hs, "--why", "w"},
		{"checkpoint", "--goal", g, "--summary", "s", "--why", "w"},
		{"answer", "--answer", k + "=code only", "--actor", "o"},
	} {
		r := hyp(t, args...)
		if r.code != 1 || !strings.Contains(r.stderr, "is ended") || !strings.Contains(r.stderr, "outgrown") {
			t.Fatalf("hyp %v under an ended Goal: exit %d, stderr %q; want exit 1 naming the end", args, r.code, r.stderr)
		}
	}
	g2 := must(t, "add", "goal", "--title", "G2", "--line", "builds on G", "--cite", "1="+g, "--why", "w")
	q2 := must(t, "add", "dq", "--parent", g2, "--title", "Q2", "--why", "ask")
	h2 := must(t, "add", "hyp", "--parent", q2, "--title", "H2", "--why", "w")
	want(t, hyp(t, "link", h2, "depends_on", hs, "--why", "w"), 0, "")
}

// Line 5, the replay half: an event under an ended Goal that a merged log
// holds after the end applies, and hyp check warns about it.
func TestWriteUnderAnEndedGoalAppliesOnReplayWithAWarning(t *testing.T) {
	g, _, _ := endStore(t)
	rawEnd(t, "o", g, "outgrown", nil)
	late := rawEvent(t, "a", event.TypeNodeAdd, "", map[string]any{"kind": event.KindDesignQuestion, "parent": fullID(t, g), "title": "late"}, "merged")
	r := hyp(t, "check")
	if strings.Contains(r.stdout, "rejected "+late) || !strings.Contains(r.stdout, "warn: "+late) || !strings.Contains(r.stdout, "ended") {
		t.Fatalf("a merged event under an ended Goal is not applied with a warning:\n%s", r.stdout)
	}
	want(t, hyp(t, "show", g), 0, "late")
}

// Line 6: hyp prompt for an ended Goal exits 1, names the end and prints
// nothing on stdout.
func TestPromptForAnEndedGoalRefused(t *testing.T) {
	promptStore(t)
	g := goalWithBody(t, "o", "Builds the design.\n"+promptBlock+"\n")
	want(t, hyp(t, "prompt", g, "--actor", "claude-run-1"), 0, "/goal ")
	must(t, "answer", "--end", g+"=outgrown", "--actor", "o")
	r := hyp(t, "prompt", g, "--actor", "claude-run-1")
	if r.code != 1 || r.stdout != "" || !strings.Contains(r.stderr, "ended") {
		t.Fatalf("hyp prompt of an ended Goal: exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}

// Line 7: an end settles every key of its Goal. A later line may cite the
// ended Goal's id, which reads ended, and a key not settled before the end,
// which reads ended; a key settled before keeps its meaning. hyp show of the
// line prints the end's words beside the cite, and a line written before the
// end prints that its Goal has since ended.
func TestEndedGoalIsCited(t *testing.T) {
	g, q, keys := endStore(t, "one tenant per row")
	open := openReading(t, q)
	answered := openReading(t, must(t, "add", "dq", "--parent", g, "--title", "QA", "--why", "ask"))
	must(t, "answer", "--answer", answered+"=code only", "--actor", "o")
	early := must(t, "add", "goal", "--title", "EARLY", "--line", "builds on the answer", "--cite", "1="+answered, "--why", "w")
	want(t, hyp(t, "add", "goal", "--title", "NOT", "--line", "builds on G", "--cite", "1="+g, "--why", "w"), 1, "bad cite")
	must(t, "answer", "--end", g+"=outgrown", "--actor", "o")

	r := hyp(t, "add", "goal", "--title", "LATE", "--line", "after G", "--cite", "1="+g, "--cite", "1="+open, "--cite", "1="+answered, "--cite", "1="+keys[0], "--why", "w")
	want(t, r, 0, "  line ")
	late := strings.Fields(r.stdout)[0]
	sh := hyp(t, "show", late+"#1").stdout
	for _, s := range []string{"cites " + g + ": ended — " + g + " ended: outgrown\n", "cites " + open + ": ended", "cites " + answered + ": applied"} {
		if !strings.Contains(sh, s) {
			t.Fatalf("hyp show of the line does not print %q:\n%s", s, sh)
		}
	}
	want(t, hyp(t, "show", early+"#1"), 0, g+" has since ended: outgrown")
	want(t, hyp(t, "check"), 0, "check: 0 errors")
}

// Line 7's last sentence: the id of a Goal not ended is refused as a cite
// with exit 1.
func TestGoalNotEndedRefusedAsACite(t *testing.T) {
	g, _, _ := endStore(t)
	want(t, hyp(t, "add", "goal", "--title", "G2", "--line", "builds on G", "--cite", "1="+g, "--why", "w"), 1, "bad cite: line 1 cites "+g)
}
