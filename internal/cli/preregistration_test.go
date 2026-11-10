package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// preStore is a store with a measuring recipe tool@1, a recipe vote@1 whose
// known answers print a verdict, a rubric rule@1, and a bundle in rubrics/
// that is not a rubric.
func preStore(t *testing.T) (dir, q, h string) {
	t.Helper()
	dir, _, q = setup(t)
	bundleIn(t, dir, "tool", "## Standalone\n\nRun rows: ./rows.sh\n", map[string]string{
		"rows.sh":                        "#!/bin/sh\necho stored rows 4\n",
		"known-answers/ok/expect.json":   `{"step": "standalone", "command": "rows", "exit": 0, "stdout": "stored rows 4\n"}`,
		"known-answers/ok/input/a":       "",
		"known-answers/fail/expect.json": `{"step": "standalone", "command": "rows", "exit": 1, "stdout": ""}`,
		"known-answers/fail/input/a":     "",
	})
	bundleIn(t, dir, "vote", "## Standalone\n\nRun vote: ./vote.sh\n", map[string]string{
		"vote.sh":                      "#!/bin/sh\necho verdict: a=supports\n",
		"known-answers/ok/expect.json": `{"step": "standalone", "command": "vote", "exit": 0, "stdout": "verdict: a=supports\n"}`,
		"known-answers/ok/input/a":     "",
	})
	libIn(t, dir, "rubrics", "rule", "## Standalone\n\nRun rule: ./rule.sh\n", map[string]string{
		"rule.sh":                     "#!/bin/sh\n",
		"known-answers/s/expect.json": `{"step": "standalone", "command": "rule", "exit": 0, "stdout": "verdict: a=supports\n"}`,
		"known-answers/s/input/a":     "",
		"known-answers/c/expect.json": `{"step": "standalone", "command": "rule", "exit": 0, "stdout": "verdict: a=contradicts\n"}`,
		"known-answers/c/input/a":     "",
	})
	libIn(t, dir, "rubrics", "half", "## Launch\n\nRun: ./up.sh\n\n## Standalone\n\nRun rule: ./rule.sh\n", map[string]string{
		"up.sh":                       "#!/bin/sh\n",
		"rule.sh":                     "#!/bin/sh\n",
		"known-answers/s/expect.json": `{"step": "standalone", "command": "rule", "exit": 0, "stdout": "verdict: a=supports\n"}`,
		"known-answers/s/input/a":     "",
	})
	h = must(t, "add", "hyp", "--parent", q, "--title", "H", "--alias", "h", "--why", "w")
	return dir, q, h
}

// Under a Hypothesis an Experiment pins a rubric; a recipe measures and
// never prints a verdict; a rubric bundle is a rubric with measures to read.
func TestAddExpPinsARubric(t *testing.T) {
	dir, q, h := preStore(t)
	addExp := func(extra ...string) result {
		t.Helper()
		return hyp(t, append([]string{"add", "exp", "--parent", h, "--title", "E", "--target", "none", "--why", "w"}, extra...)...)
	}
	before := logLines(t, dir)
	want(t, addExp("--recipe", "tool@1"), 1, "refused: an Experiment under a Hypothesis pins a rubric: --rubric NAME@VERSION")
	want(t, addExp("--recipe", "vote@1", "--rubric", "reading"), 1, "refused: recipe vote@1: a recipe measures and never prints a verdict; put the rule in a rubric")
	want(t, addExp("--recipe", "tool@1", "--rubric", "half@1"), 1,
		"refused: rubric half@1 is not a rubric: it has a step section; a rubric has only its Standalone command; verdict a: no contradicts case")
	want(t, addExp("--recipe", "none", "--rubric", "rule@1"), 1, "refused: rubric rule@1 has no measures to read: pin a recipe or --records")
	want(t, addExp("--recipe", "tool@1", "--rubric", "tool@1"), 2, "rubric tool@1 not found")
	want(t, addExp("--recipe", "tool@1", "--rubric", "reading", "--decided-by", "reading"), 2, "pass only --rubric")
	want(t, addExp("--recipe", "none", "--decided-by", "measure"), 2, `invalid --decided-by "measure"`)
	want(t, addExp("--recipe", "none", "--rubric", "reading", "--records", "notes.md"), 2, "--records notes.md is not under .tree/evidence/")
	if n := logLines(t, dir); n != before {
		t.Fatalf("refused add exp wrote %d lines", n-before)
	}

	ruled := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "tool@1", "--rubric", "rule@1", "--target", "none", "--why", "w")
	want(t, hyp(t, "show", ruled), 0, "rubric: rule@1\n")
	want(t, hyp(t, "show", ruled), 0, "shape: run and rule\n")
	read := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	want(t, hyp(t, "show", read), 0, "rubric: reading\n")
	want(t, hyp(t, "show", read), 0, "shape: rule by reading\n")

	// Under a DesignQuestion an Experiment measures: no rubric, no target.
	want(t, hyp(t, "add", "exp", "--parent", q, "--title", "M", "--recipe", "tool@1", "--rubric", "reading", "--why", "w"), 2,
		"an Experiment under a DesignQuestion measures")
	want(t, hyp(t, "add", "exp", "--parent", q, "--title", "M", "--recipe", "tool@1", "--target", "h", "--why", "w"), 2,
		"an Experiment under a DesignQuestion measures")
	m := must(t, "add", "exp", "--parent", q, "--title", "M", "--recipe", "tool@1", "--why", "w")
	want(t, hyp(t, "show", m), 0, "shape: measure only\n")
	want(t, hyp(t, "record", "--exp", m, "--verdict", "h=supports", "--observation", "o", "--why", "w"), 1, "has no targets; it measures")
	must(t, "record", "--exp", m, "--observation", "stored rows 4", "--why", "w")
}

// A reassessment pins the records it reads at add; a Result cites only its
// own run and those records, and only while they hash as pinned.
func TestRecordsArePinnedAtAdd(t *testing.T) {
	dir, _, h := preStore(t)
	run := must(t, "add", "exp", "--parent", h, "--title", "run", "--recipe", "tool@1", "--rubric", "rule@1", "--target", "none", "--why", "w")
	writeFile(t, filepath.Join(dir, ".tree", "evidence", run, "stdout"), "stored rows 4\n")
	writeFile(t, filepath.Join(dir, ".tree", "evidence", run, "sub", "log"), "l\n")
	must(t, "record", "--exp", run, "--verdict", "h=supports", "--observation", "o", "--evidence", ".tree/evidence/"+run+"/stdout", "--why", "w")

	again := must(t, "add", "exp", "--parent", h, "--title", "again", "--recipe", "none", "--rubric", "rule@1", "--target", "none",
		"--records", ".tree/evidence/"+run, "--why", "w")
	r := hyp(t, "show", again)
	want(t, r, 0, "shape: reassess from kept records\n")
	want(t, r, 0, ".tree/evidence/"+run+"/stdout")
	want(t, r, 0, ".tree/evidence/"+run+"/sub/log")

	// A record of another run that is not pinned is refused.
	other := must(t, "add", "exp", "--parent", h, "--title", "other", "--recipe", "none", "--rubric", "reading", "--target", "none", "--why", "w")
	want(t, hyp(t, "record", "--exp", other, "--verdict", "h=supports", "--observation", "o", "--evidence", ".tree/evidence/"+run+"/stdout", "--why", "w"),
		1, "refused: bad evidence: .tree/evidence/"+run+"/stdout is a record outside "+other+"'s run and its pinned records")

	must(t, "record", "--exp", again, "--verdict", "h=supports", "--observation", "o", "--evidence", ".tree/evidence/"+run+"/stdout", "--why", "w")
	if err := os.WriteFile(filepath.Join(dir, ".tree", "evidence", run, "sub", "log"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	want(t, hyp(t, "record", "--exp", again, "--verdict", "h=supports", "--observation", "o", "--why", "w"),
		1, "refused: bad records: .tree/evidence/"+run+"/sub/log changed since "+again+" pinned it")
}

// A Result is refused when its Experiment's rubric changed since the add.
func TestRecordRefusesAChangedRubric(t *testing.T) {
	dir, _, h := preStore(t)
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "tool@1", "--rubric", "rule@1", "--target", "none", "--why", "w")
	writeFile(t, filepath.Join(dir, "rubrics", "rule@1", "notes.txt"), "n\n")
	want(t, hyp(t, "record", "--exp", e, "--verdict", "h=supports", "--observation", "o", "--why", "w"),
		1, "refused: rubric rule@1 changed since "+e+" pinned it")
}
