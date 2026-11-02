package sandbox

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/recipe"
)

const kaRecipe = "## Evidence\n\nRun: ./check.sh\n\n## Standalone\n\nRun echo: ./echo.sh\nRun rand: ./rand.sh\nRun slow: ./slow.sh\nRun bg: ./bg.sh\nRun log: ./log.sh\nRun count: ./count.sh\n"

// kaFiles is the known-answer fixture bundle: its scripts, plus cases.
func kaFiles(cases map[string]string) (map[string]string, []string) {
	files := map[string]string{
		"recipe.md": recipeMD("ka", kaRecipe),
		"check.sh":  "#!/bin/sh\nif [ -e \"$1/ok\" ]; then echo pass; else echo fail; exit 30; fi\n",
		"echo.sh":   "#!/bin/sh\necho \"$@\"\npwd\necho \"$HOME\"\ncase \"$1\" in ''|*[!0-9]*) exit 0;; *) exit \"$1\";; esac\n",
		"rand.sh":   "#!/bin/bash\necho $RANDOM$RANDOM$RANDOM\n",
		"slow.sh":   "#!/bin/sh\nsleep 5\n",
		"bg.sh":     "#!/bin/sh\nsleep 30 &\necho bg\n",
		"log.sh":    "#!/bin/sh\ngit -C \"$1/../repo\" log --format=%s\n",
		"count.sh":  "#!/bin/sh\necho run >> \"$1\"\nsleep 5\n",
	}
	for rel, body := range cases {
		files["known-answers/"+rel] = body
	}
	return files, []string{"check.sh", "echo.sh", "rand.sh", "slow.sh", "bg.sh", "log.sh", "count.sh"}
}

func expect(step, command string, args []string, exit int, stdout, repo string) string {
	m := map[string]any{"step": step, "command": command, "exit": exit, "stdout": stdout}
	if args != nil {
		m["args"] = args
	}
	if repo != "" {
		m["repo"] = repo
	}
	data, _ := json.Marshal(m)
	return string(data)
}

// kaRecipeAt opens the fixture bundle with cases.
func kaRecipeAt(t *testing.T, cases map[string]string) *recipe.Recipe {
	t.Helper()
	files, ex := kaFiles(cases)
	root := recipeRoot(t, "ka", files, ex...)
	r, err := recipe.Open(filepath.Join(root, "recipes", "ka@1"))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// runKA runs r's known answers with a copy of the running hyp.
func runKA(t *testing.T, r *recipe.Recipe) (*cmdrecord.KnownAnswers, error) {
	t.Helper()
	hyp, err := NewSandboxHyp()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(hyp.Remove)
	return RunKnownAnswers(r, hyp, io.Discard)
}

// repoBundle makes a git bundle holding one commit with subject "first",
// and HEAD.
func repoBundle(t *testing.T) string {
	t.Helper()
	return gitBundle(t, "--all")
}

// gitBundle makes a git bundle of one commit with subject "first" on branch
// main, with git bundle create's rev arguments revs.
func gitBundle(t *testing.T, revs ...string) string {
	t.Helper()
	dir := t.TempDir()
	out := filepath.Join(t.TempDir(), "repo.bundle")
	for _, args := range [][]string{
		{"init", "--quiet", "--initial-branch=main", dir},
		{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "--quiet", "--allow-empty", "-m", "first"},
		append([]string{"-C", dir, "bundle", "create", "--quiet", out}, revs...),
	} {
		if b, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, b)
		}
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestKnownAnswers(t *testing.T) {
	old := caseTimeout
	caseTimeout = 2 * time.Second
	t.Cleanup(func() { caseTimeout = old })
	lines := "\n$CASE/cwd\n$CASE/home\n"
	r := kaRecipeAt(t, map[string]string{
		"ok-standalone/expect.json":            expect("standalone", "echo", []string{"hi", "$CASE/x"}, 0, "hi $CASE/x"+lines, ""),
		"ok-standalone/input/a":                "a\n",
		"ok-evidence/expect.json":              expect("evidence", "./check.sh", nil, 0, "pass\n", ""),
		"ok-evidence/input/ok":                 "",
		"fail-evidence/expect.json":            expect("evidence", "./check.sh", nil, 30, "fail\n", ""),
		"fail-evidence/input/other":            "",
		"wrong-exit/expect.json":               expect("standalone", "echo", nil, 3, lines, ""),
		"wrong-exit/input/a":                   "",
		"wrong-stdout/expect.json":             expect("standalone", "echo", nil, 0, "nope\n", ""),
		"wrong-stdout/input/a":                 "",
		"random/expect.json":                   expect("standalone", "rand", nil, 0, "1\n", ""),
		"random/input/a":                       "",
		"repo/expect.json":                     expect("standalone", "log", []string{"$CASE/input"}, 0, "first\n", "repo.bundle"),
		"repo/input/repo.bundle":               repoBundle(t),
		"slow/expect.json":                     expect("standalone", "slow", nil, 0, "", ""),
		"slow/input/a":                         "",
		"background/expect.json":               expect("standalone", "bg", nil, 0, "bg\n", ""),
		"background/input/a":                   "",
		"dot-input/expect.json":                expect("evidence", "./check.sh", nil, 0, "pass\n", ""),
		"dot-input/input/ok":                   "",
		"dot-input/input/.tree/events/x.jsonl": "{}\n",
		"at-input/expect.json":                 expect("evidence", "./check.sh", nil, 0, "pass\n", ""),
		"at-input/input/ok":                    "",
		"at-input/input/recipes/x@1/recipe.md": "pinned\n",
	})
	start := time.Now()
	ka, err := runKA(t, r)
	if err != nil {
		t.Fatal(err)
	}
	if ka.HypSHA256 == "" {
		t.Error("hyp_sha256 is empty")
	}
	got := map[string]string{}
	for _, c := range ka.Cases {
		got[c.Name] = c.Outcome
	}
	for name, want := range map[string]string{
		"ok-standalone": "ok", "ok-evidence": "ok", "fail-evidence": "ok", "dot-input": "ok", "at-input": "ok", "repo": "ok", "background": "ok",
		"wrong-exit": "exit 0, want 3", "wrong-stdout": "stdout differs", "random": "not stateless", "slow": "timed out after 2s",
	} {
		t.Run(name, func(t *testing.T) {
			if got[name] != want {
				t.Errorf("outcome = %q, want %q", got[name], want)
			}
		})
	}
	if len(ka.Cases) != 11 || ka.Passed() {
		t.Errorf("cases = %+v, passed = %v", ka.Cases, ka.Passed())
	}
	// The background case's sleep 30 must not hold the run.
	if d := time.Since(start); d > 30*time.Second {
		t.Errorf("run took %v", d)
	}
}

// A case whose first run timed out cannot pass, so it is not run again.
func TestKnownAnswersTimeoutRunsOnce(t *testing.T) {
	old := caseTimeout
	caseTimeout = time.Second
	t.Cleanup(func() { caseTimeout = old })
	runs := filepath.Join(t.TempDir(), "runs")
	r := kaRecipeAt(t, map[string]string{
		"hung/expect.json": expect("standalone", "count", []string{runs}, 0, "", ""),
		"hung/input/a":     "",
	})
	ka, err := runKA(t, r)
	if err != nil {
		t.Fatal(err)
	}
	if len(ka.Cases) != 1 || ka.Cases[0].Outcome != "timed out after 1s" {
		t.Fatalf("cases = %+v", ka.Cases)
	}
	data, err := os.ReadFile(runs)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(data), "\n"); n != 1 {
		t.Errorf("the case ran %d times, want 1", n)
	}
}

func TestKnownAnswersPassed(t *testing.T) {
	r := kaRecipeAt(t, map[string]string{
		"a/expect.json": expect("evidence", "./check.sh", nil, 0, "pass\n", ""),
		"a/input/ok":    "",
	})
	ka, err := runKA(t, r)
	if err != nil || !ka.Passed() {
		t.Fatalf("ka = %+v, err = %v", ka, err)
	}
	if (&cmdrecord.KnownAnswers{}).Passed() {
		t.Error("no cases passed")
	}
}

func TestKnownAnswersRefusals(t *testing.T) {
	ok := expect("evidence", "./check.sh", nil, 0, "pass\n", "")
	for name, c := range map[string]struct {
		cases map[string]string
		want  string
	}{
		"unknown-command": {map[string]string{"c/expect.json": expect("evidence", "./nope.sh", nil, 0, "", ""), "c/input/a": ""},
			`known answer c: command "./nope.sh" matches no Evidence line`},
		"unknown-label": {map[string]string{"c/expect.json": expect("standalone", "nope", nil, 0, "", ""), "c/input/a": ""},
			`command "nope" matches no Standalone label`},
		"args-evidence": {map[string]string{"c/expect.json": expect("evidence", "./check.sh", []string{"x"}, 0, "", ""), "c/input/a": ""},
			"args is not allowed with step evidence"},
		"missing-input": {map[string]string{"c/expect.json": ok}, "input/ is missing"},
		"unknown-field": {map[string]string{"c/expect.json": `{"step":"evidence","command":"./check.sh","exit":0,"stdout":"","x":1}`, "c/input/a": ""},
			`unknown field "x"`},
		"missing-key":  {map[string]string{"c/expect.json": `{"step":"evidence","command":"./check.sh","exit":0}`, "c/input/a": ""}, `no "stdout"`},
		"bad-step":     {map[string]string{"c/expect.json": expect("drive", "./check.sh", nil, 0, "", ""), "c/input/a": ""}, `step "drive"`},
		"bad-name":     {map[string]string{"Bad/expect.json": ok, "Bad/input/a": ""}, "the name must match"},
		"not-dir":      {map[string]string{"c": "x"}, "known answer c: is not a directory"},
		"repo-missing": {map[string]string{"c/expect.json": expect("evidence", "./check.sh", nil, 0, "", "r.bundle"), "c/input/a": ""}, `repo "r.bundle"`},
		"repo-no-head": {map[string]string{"c/expect.json": expect("evidence", "./check.sh", nil, 0, "", "r.bundle"), "c/input/r.bundle": gitBundle(t, "main")},
			`known answer c: repo "r.bundle" holds no HEAD, so its clone would have no checkout; make it with git bundle create FILE HEAD <branch>`},
		"old-sandbox": {map[string]string{"c/expect.json": ok, "c/input/sandbox.json": `{"format": 0}`}, "input/sandbox.json has format 0"},
		"changes":     {map[string]string{"c/expect.json": ok, "c/input/changes.json": `{}`}, "input/changes.json"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := runKA(t, kaRecipeAt(t, c.cases))
			if code(err) != 1 || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v (code %d), want 1 and %q", err, code(err), c.want)
			}
		})
	}
}

func TestKnownAnswersNone(t *testing.T) {
	ka, err := runKA(t, kaRecipeAt(t, nil))
	if err != nil || len(ka.Cases) != 0 || ka.Passed() {
		t.Fatalf("ka = %+v, err = %v", ka, err)
	}
}

// vcases makes known-answer cases from name -> "step command exit stdout",
// each with an input/ to read.
func vcases(spec map[string][4]any) map[string]string {
	files := map[string]string{}
	for name, s := range spec {
		files["known-answers/"+name+"/expect.json"] = expect(s[0].(string), s[1].(string), nil, s[2].(int), s[3].(string), "")
		files["known-answers/"+name+"/input/a"] = ""
	}
	return files
}

// WeakCases warns, per check, when a verdict check's known answers miss a
// supports or a contradicts case for a label, print a malformed verdict
// line, or print one label twice in one case; a check whose cases print no
// verdict line gets only the must-pass and must-fail warnings.
func TestWeakCasesVerdicts(t *testing.T) {
	const standalone = "## Standalone\n\nRun v: ./v.sh\n"
	sup, con, inc := "verdict: x=supports\n", "verdict: x=contradicts\n", "verdict: x=inconclusive\n"
	fail := [4]any{"standalone", "v", 2, ""}
	for _, tc := range []struct {
		name, body string
		cases      map[string][4]any
		want       []string
	}{
		{"covered", standalone, map[string][4]any{
			"s": {"standalone", "v", 0, sup}, "c": {"standalone", "v", 0, con}, "f": fail,
		}, nil},
		{"no contradicts", standalone, map[string][4]any{
			"s": {"standalone", "v", 0, sup},
		}, []string{"v: no must-fail case", "v: verdict x: no contradicts case"}},
		{"no supports", standalone, map[string][4]any{
			"c": {"standalone", "v", 0, con}, "i": {"standalone", "v", 0, inc}, "f": fail,
		}, []string{"v: verdict x: no supports case"}},
		{"malformed", standalone, map[string][4]any{
			"s": {"standalone", "v", 0, sup + "verdict: x supports\n"},
			"c": {"standalone", "v", 0, "verdict: X=yes\n" + con},
			"f": fail,
		}, []string{`v: case c: malformed verdict line "verdict: X=yes"`, `v: case s: malformed verdict line "verdict: x supports"`}},
		{"printed twice", standalone, map[string][4]any{
			"s": {"standalone", "v", 0, sup + inc + con}, "c": {"standalone", "v", 0, con}, "f": fail,
		}, []string{"v: case s: label x printed twice"}},
		{"no verdict line", standalone, map[string][4]any{
			"p": {"standalone", "v", 0, "pass\nthe verdict: x=supports\n"}, "f": fail,
		}, nil},
		{"two labels", standalone, map[string][4]any{
			"a": {"standalone", "v", 0, "verdict: x=supports\nverdict: y=contradicts\n"},
			"b": {"standalone", "v", 0, "verdict: y=supports\nverdict: x=contradicts\n"},
			"f": fail,
		}, nil},
		{"inconclusive only", standalone, map[string][4]any{
			"s": {"standalone", "v", 0, sup + "verdict: z=inconclusive\n"}, "c": {"standalone", "v", 0, con}, "f": fail,
		}, []string{"v: verdict z: no supports case", "v: verdict z: no contradicts case"}},
		// Full exit-code coverage does not stand in for a missing verdict.
		{"exit codes covered", standalone, map[string][4]any{
			"p": {"standalone", "v", 0, sup}, "f": {"standalone", "v", 2, sup},
		}, []string{"v: verdict x: no contradicts case"}},
		{"evidence", "## Evidence\n\nRun: ./ev.sh\n", map[string][4]any{
			"s": {"evidence", "./ev.sh", 0, sup}, "f": {"evidence", "./ev.sh", 1, ""},
		}, []string{"./ev.sh: verdict x: no contradicts case"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := vcases(tc.cases)
			files["recipe.md"] = recipeMD("vc", tc.body)
			files["v.sh"], files["ev.sh"] = "#!/bin/sh\n", "#!/bin/sh\n"
			r, err := recipe.Open(filepath.Join(recipeRoot(t, "vc", files, "v.sh", "ev.sh"), "recipes", "vc@1"))
			if err != nil {
				t.Fatal(err)
			}
			if got := WeakCases(r); !slices.Equal(got, tc.want) {
				t.Errorf("WeakCases = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBundleRefs(t *testing.T) {
	sha := strings.Repeat("a", 40)
	for name, c := range map[string]struct {
		data string
		want []string
		ok   bool
	}{
		"v2":         {"# v2 git bundle\n" + sha + " refs/heads/main\n" + sha + " HEAD\n\nPACK", []string{"refs/heads/main", "HEAD"}, true},
		"v3":         {"# v3 git bundle\n@object-format=sha1\n-" + sha + " parent\n" + sha + " refs/heads/main\n\nPACK", []string{"refs/heads/main"}, true},
		"not-bundle": {"hello\n", nil, false},
		"no-end":     {"# v2 git bundle\n" + sha + " HEAD\n", nil, false},
	} {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "b")
			if err := os.WriteFile(p, []byte(c.data), 0o644); err != nil {
				t.Fatal(err)
			}
			got, ok := bundleRefs(p)
			if ok != c.ok || !slices.Equal(got, c.want) {
				t.Errorf("bundleRefs = %q, %v; want %q, %v", got, ok, c.want, c.ok)
			}
		})
	}
}
