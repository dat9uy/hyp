package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/recipe"
)

// gateFiles is an instrument bundle ins whose Launch writes mark and whose
// one case expects stdout; check.sh prints "pass".
func gateFiles(mark, stdout string) map[string]string {
	return map[string]string{
		"recipe.md":                     recipeMD("ins", "## Launch\n\nRun: ./launch.sh\n\n## Evidence\n\nRun: ./check.sh\n"),
		"launch.sh":                     "#!/bin/sh\n: > " + mark + "\n",
		"check.sh":                      "#!/bin/sh\necho pass\necho checked >&2\n",
		"known-answers/one/expect.json": expect("evidence", "./check.sh", nil, 0, stdout, ""),
		"known-answers/one/input/a":     "",
	}
}

func TestCreateRunsGateBeforeAnything(t *testing.T) {
	setup(t)
	mark := filepath.Join(t.TempDir(), "launched")
	root := recipeRoot(t, "ins", gateFiles(mark, "nope\n"), "launch.sh", "check.sh")
	_, box, err := createWith(t, "gate", CreateOptions{Refs: []string{"ins@1"}, Recipes: root})
	if code(err) != 1 || !strings.Contains(err.Error(), "recipe ins@1 failed its known answers: one: stdout differs") ||
		!strings.Contains(err.Error(), "nothing created") || !strings.Contains(err.Error(), "checked") {
		t.Fatalf("err = %v, want exit 1 naming the case and its stderr", err)
	}
	if exists(box) || exists(mark) {
		t.Fatalf("sandbox %v, Launch ran %v; want neither", exists(box), exists(mark))
	}

	root = recipeRoot(t, "ins", gateFiles(mark, "pass\n"), "launch.sh", "check.sh")
	rec, box, err := createWith(t, "gate", CreateOptions{Refs: []string{"ins@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	ka := rec.Recipes[0].KnownAnswers
	if ka == nil || len(ka.HypSHA256) != 64 || len(ka.Cases) != 1 || ka.Cases[0] != (cmdrecord.Case{Name: "one", Outcome: "ok"}) || !ka.Passed() {
		t.Fatalf("known answers = %+v", ka)
	}
	if !exists(mark) {
		t.Error("Launch did not run after the gate passed")
	}
	if log, _ := os.ReadFile(filepath.Join(box, "known-answers.log")); !strings.Contains(string(log), "== ins@1 known answer one\nchecked\n") {
		t.Errorf("known-answers.log = %q", log)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "gate", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(out, "known-answers.log")) {
		t.Error("known-answers.log not in OUT_DIR")
	}
}

func TestCreateGatesEachStackedRecipe(t *testing.T) {
	setup(t)
	mark := filepath.Join(t.TempDir(), "launched")
	root := recipeRoot(t, "ins", gateFiles(mark, "pass\n"), "launch.sh", "check.sh")
	writeFile(t, filepath.Join(root, "recipes", "env@1", "recipe.md"), recipeMD("env", "Set nothing up.\n"))
	rec, _, err := createWith(t, "stack", CreateOptions{Refs: []string{"env@1", "ins@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Recipes[0].KnownAnswers != nil {
		t.Errorf("env@1 known answers = %+v, want none (not an instrument)", rec.Recipes[0].KnownAnswers)
	}
	if ka := rec.Recipes[1].KnownAnswers; ka == nil || !ka.Passed() {
		t.Errorf("ins@1 known answers = %+v", ka)
	}
}

func TestGateTimeoutIsEnvironment(t *testing.T) {
	old := caseTimeout
	caseTimeout = time.Second
	t.Cleanup(func() { caseTimeout = old })
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	mark := filepath.Join(t.TempDir(), "launched")
	files := gateFiles(mark, "pass\n")
	files["check.sh"] = "#!/bin/sh\nsleep 5\n"
	root := recipeRoot(t, "ins", files, "launch.sh", "check.sh")
	_, box, err := createWith(t, "slow", CreateOptions{Refs: []string{"ins@1"}, Recipes: root})
	if code(err) != 3 || !strings.Contains(err.Error(), "known answers timed out: one: timed out after 1s") {
		t.Fatalf("err = %v, want exit 3", err)
	}
	if exists(box) || exists(mark) {
		t.Fatal("the gate timed out but the sandbox was made")
	}
}

func TestRunRecordsKnownAnswers(t *testing.T) {
	root := evidenceRoot(t)
	out := filepath.Join(t.TempDir(), "rec")
	_, err := Run(RunOptions{Dir: root, Ref: "chk@1", Step: "evidence", Args: []string{keptDir(t)}, Record: out})
	failed(t, err) // check.sh exits 30: a failed command, recorded
	rec := readCommandRecord(t, out)
	if ka := rec.KnownAnswers; ka == nil || len(ka.Cases) != 1 || ka.Cases[0] != (cmdrecord.Case{Name: "pass", Outcome: "ok"}) {
		t.Fatalf("known_answers = %+v", rec.KnownAnswers)
	}
	if log, _ := os.ReadFile(filepath.Join(out, "known-answers.log")); !strings.HasPrefix(string(log), "== chk@1 known answer pass\n") {
		t.Errorf("known-answers.log = %q", log)
	}
}

func TestRunRefusesFailedKnownAnswers(t *testing.T) {
	ran := filepath.Join(t.TempDir(), "ran")
	root := recipeRoot(t, "chk", passCase(map[string]string{
		"recipe.md": recipeMD("chk", "## Evidence\n\nRun: ./check.sh\n"),
		"check.sh":  "#!/bin/sh\n[ -e \"$1/sandbox.json\" ] && touch " + ran + "\necho \"$1\"\n",
	}, "evidence", "./check.sh", nil, 30, "$CASE/input\n"), "check.sh")
	out := filepath.Join(t.TempDir(), "rec")
	_, err := Run(RunOptions{Dir: root, Ref: "chk@1", Step: "evidence", Args: []string{keptDir(t)}, Record: out})
	if code(err) != 1 || !strings.Contains(err.Error(), "recipe chk@1 failed its known answers: pass: exit 0, want 30; nothing ran") {
		t.Fatalf("err = %v", err)
	}
	if exists(out) || exists(ran) {
		t.Fatalf("record dir %v, command ran %v; want neither", exists(out), exists(ran))
	}
}

// A case finds as hyp a copy of the running hyp, first on PATH, whatever
// else PATH holds; its hash is the one the record names.
func TestGateHypIsTheRunningBinary(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	decoy := t.TempDir()
	writeFile(t, filepath.Join(decoy, "hyp"), "#!/bin/sh\necho decoy\n")
	if err := os.Chmod(filepath.Join(decoy, "hyp"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", decoy+string(os.PathListSeparator)+os.Getenv("PATH"))
	files, ex := kaFiles(map[string]string{
		"which/expect.json": expect("standalone", "which", []string{exe}, 0, "same\n", ""),
		"which/input/a":     "",
	})
	files["recipe.md"] = recipeMD("ka", kaRecipe+"Run which: ./which.sh\n")
	files["which.sh"] = "#!/bin/sh\nh=$(command -v hyp)\n[ \"$h\" != \"$1\" ] && cmp -s \"$h\" \"$1\" && echo same\n"
	root := recipeRoot(t, "ka", files, append(ex, "which.sh")...)
	r, err := recipe.Open(filepath.Join(root, "recipes", "ka@1"))
	if err != nil {
		t.Fatal(err)
	}
	ka, err := runKA(t, r)
	if err != nil {
		t.Fatal(err)
	}
	if ka.HypSHA256 != hex.EncodeToString(sum[:]) || len(ka.Cases) != 1 || ka.Cases[0].Outcome != "ok" {
		t.Fatalf("known answers = %+v", ka)
	}
}
