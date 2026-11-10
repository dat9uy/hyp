package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/recipe"
)

// decoyHyp puts a hyp that exits 1 first on PATH.
func decoyHyp(t *testing.T) {
	t.Helper()
	decoy := t.TempDir()
	writeFile(t, filepath.Join(decoy, "hyp"), "#!/bin/sh\necho decoy\nexit 1\n")
	if err := os.Chmod(filepath.Join(decoy, "hyp"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", decoy+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// A Standalone and an Evidence command run by hyp bundle run and evidence
// find as hyp the copy the gate ran, whatever the caller's PATH holds
// first; known_answers names its hash.
func TestRunUsesTheGatedHyp(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	decoyHyp(t)
	same := "h=$(command -v hyp)\n[ \"$h\" != \"$exe\" ] && cmp -s \"$h\" \"$exe\" && echo same\n"
	root := recipeRoot(t, "which", map[string]string{
		"recipe.md":                    recipeMD("which", "## Evidence\n\nRun: ./ev.sh\n\n## Standalone\n\nRun which: ./which.sh\n"),
		"which.sh":                     "#!/bin/sh\nexe=$1\n" + same,
		"ev.sh":                        "#!/bin/sh\nexe=$(cat \"$1/exe\")\n" + same,
		"known-answers/sa/expect.json": expect("standalone", "which", []string{exe}, 0, "same\n", ""),
		"known-answers/sa/input/a":     "",
		"known-answers/ev/expect.json": expect("evidence", "./ev.sh", nil, 0, "same\n", ""),
		"known-answers/ev/input/exe":   exe,
	}, "which.sh", "ev.sh")
	kept := keptDir(t)
	writeFile(t, filepath.Join(kept, "exe"), exe)
	for _, o := range []RunOptions{
		{Dir: root, Ref: "which@1", Step: "standalone", Label: "which", Args: []string{exe}},
		{Dir: root, Ref: "which@1", Step: "evidence", Args: []string{kept}},
	} {
		o.Record = filepath.Join(t.TempDir(), "rec")
		rec, err := Run(o)
		if err != nil {
			t.Fatalf("%s: %v", o.Step, err)
		}
		if log, _ := os.ReadFile(filepath.Join(o.Record, "command.log")); !strings.Contains(string(log), "same\n") {
			t.Errorf("%s: command.log = %q, want the gated hyp", o.Step, log)
		}
		if ka := rec.KnownAnswers; ka == nil || ka.HypSHA256 != hex.EncodeToString(sum[:]) || !ka.Passed() {
			t.Errorf("%s: known answers = %+v", o.Step, ka)
		}
	}
}

// digits matches the random part of a temporary directory's name.
var digits = regexp.MustCompile(`[0-9]+`)

// tokenized replaces in out the root a check ran under (the parent of the
// pwd it printed last) and the hyp directory first on its PATH.
func tokenized(t *testing.T, out string) string {
	t.Helper()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	i := slices.IndexFunc(lines, func(l string) bool { return strings.HasPrefix(l, "pwd=") })
	j := slices.IndexFunc(lines, func(l string) bool { return strings.HasPrefix(l, "PATH=") })
	if i < 0 || j < 0 {
		t.Fatalf("output has no pwd= or PATH= line: %q", out)
	}
	root := filepath.Dir(strings.TrimPrefix(lines[i], "pwd="))
	bin, _, _ := strings.Cut(strings.TrimPrefix(lines[j], "PATH="), string(os.PathListSeparator))
	return strings.ReplaceAll(strings.ReplaceAll(out, bin, "$BIN"), root, "$ROOT")
}

// A check gets the same environment, working directory, stdin and paths as a
// known-answer case and as a Standalone run, with its scratch root and its
// bundle copy of the same shape, and nothing of the caller's environment but
// PATH. Only the random digits of a temporary directory's name may differ.
func TestCheckEnvSameInCaseAndRun(t *testing.T) {
	t.Setenv("FOO", "leak")
	t.Setenv("GIT_DIR", "/nonexistent")
	t.Setenv("HYP_SANDBOX_ID", "leak")
	caseOut, runOut := filepath.Join(t.TempDir(), "case"), filepath.Join(t.TempDir(), "run")
	root := recipeRoot(t, "env", passCase(map[string]string{
		"recipe.md": recipeMD("env", "## Standalone\n\nRun env: ./env.sh\n"),
		"env.sh":    "#!/bin/sh\n{ env | sort; echo \"self=$0\"; echo \"pwd=$(pwd)\"; echo \"ls=$(ls -A)\"; echo \"stdin=$(cat)\"; } > \"$1\"\n",
	}, "standalone", "env", []string{caseOut}, 0, ""), "env.sh")
	if _, err := Run(RunOptions{Dir: root, Ref: "env@1", Step: "standalone", Label: "env", Args: []string{runOut}, Record: filepath.Join(t.TempDir(), "rec")}); err != nil {
		t.Fatal(err)
	}
	c, err1 := os.ReadFile(caseOut)
	r, err2 := os.ReadFile(runOut)
	if err1 != nil || err2 != nil {
		t.Fatal(err1, err2)
	}
	// The paths themselves are compared, with only their digits masked, so a
	// root or bundle copy of another prefix or depth fails here.
	if got, want := digits.ReplaceAllString(string(r), "N"), digits.ReplaceAllString(string(c), "N"); got != want {
		t.Errorf("run environment differs from the case's:\nrun:\n%s\ncase:\n%s", got, want)
	}
	got := tokenized(t, string(r))
	for _, k := range []string{"FOO=", "GIT_DIR=", "HYP_SANDBOX_ID="} {
		if strings.Contains(got, "\n"+k) || strings.HasPrefix(got, k) {
			t.Errorf("the check inherited %s:\n%s", k, got)
		}
	}
	for _, k := range []string{"HOME=$ROOT/home\n", "TMPDIR=$ROOT/tmp\n", "HYP_OUT=$ROOT/out\n", "GIT_CONFIG_NOSYSTEM=1\n", "PATH=$BIN:", "pwd=$ROOT/cwd\n", "ls=\n", "stdin=\n"} {
		if !strings.Contains(got, k) {
			t.Errorf("the check's output lacks %q:\n%s", k, got)
		}
	}
}

// A case's git clone runs in the check's environment, so host git config
// and GIT_ variables cannot break or change it.
func TestCaseCloneIgnoresHostGit(t *testing.T) {
	r := kaRecipeAt(t, map[string]string{
		"repo/expect.json":       expect("standalone", "log", []string{"$CASE/input"}, 0, "first\n", "repo.bundle"),
		"repo/input/repo.bundle": repoBundle(t),
	})
	bad := filepath.Join(t.TempDir(), "gitconfig")
	writeFile(t, bad, "[core\nnot a config\n")
	t.Setenv("GIT_CONFIG_GLOBAL", bad)
	t.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "elsewhere"))
	ka, err := runKA(t, r)
	if err != nil || !ka.Passed() {
		t.Fatalf("ka = %+v, err = %v", ka, err)
	}
}

// A Standalone ARG is passed as given; the check runs in its own empty
// directory, so a relative path no longer names a file in the caller's.
func TestRunStandaloneArgsAsGiven(t *testing.T) {
	root := recipeRoot(t, "cat", passCase(map[string]string{
		"recipe.md": recipeMD("cat", "## Standalone\n\nRun cat: ./cat.sh\n"),
		"cat.sh":    "#!/bin/sh\n[ \"$1\" = case ] && exit 0\ncat \"$1\"\n",
	}, "standalone", "cat", []string{"case"}, 0, ""), "cat.sh")
	writeFile(t, filepath.Join(root, "x"), "found\n")
	rel := filepath.Join(t.TempDir(), "rel")
	_, err := Run(RunOptions{Dir: root, Ref: "cat@1", Step: "standalone", Label: "cat", Args: []string{"x"}, Record: rel})
	failed(t, err)
	abs := filepath.Join(t.TempDir(), "abs")
	if _, err := Run(RunOptions{Dir: root, Ref: "cat@1", Step: "standalone", Label: "cat", Args: []string{filepath.Join(root, "x")}, Record: abs}); err != nil {
		t.Fatal(err)
	}
	if log, _ := os.ReadFile(filepath.Join(abs, "command.log")); !strings.Contains(string(log), "found\n") {
		t.Errorf("command.log = %q", log)
	}
}

// outFiles is a bundle out@1 whose Standalone command out writes into
// $HYP_OUT what its first word names, plus cases.
func outFiles(cases map[string]string) (map[string]string, []string) {
	files := map[string]string{
		"recipe.md": recipeMD("out", "## Standalone\n\nRun out: ./out.sh\n"),
		"out.sh": "#!/bin/bash\ncase \"$1\" in\n" +
			"x) printf x > \"$HYP_OUT/a.txt\";;\n" +
			"sub) mkdir \"$HYP_OUT/sub\"; printf c > \"$HYP_OUT/sub/c\";;\n" +
			"rand) echo $RANDOM$RANDOM$RANDOM > \"$HYP_OUT/t\";;\n" +
			"where) echo \"$HYP_OUT\" > \"$HYP_OUT/where\";;\n" +
			"link) ln -s /etc \"$HYP_OUT/l\";;\n" +
			"rm) rm -rf \"$HYP_OUT\";;\n" +
			"swap) rm -rf \"$HYP_OUT\"; mkdir \"$TMPDIR/d\"; ln -s \"$TMPDIR/d\" \"$HYP_OUT\";;\n" +
			"locked) printf x > \"$HYP_OUT/u\"; chmod 000 \"$HYP_OUT/u\";;\n" +
			"esac\necho done\n",
	}
	for rel, body := range cases {
		files["known-answers/"+rel] = body
	}
	return files, []string{"out.sh"}
}

// A case's product is its exit, stdout and $HYP_OUT tree: the tree must
// equal expect/, or be empty without one, and be the same in both runs.
func TestKnownAnswersCompareOut(t *testing.T) {
	c := func(name, arg, expectFile, body string) map[string]string {
		m := map[string]string{
			name + "/expect.json": expect("standalone", "out", []string{arg}, 0, "done\n", ""),
			name + "/input/a":     "",
		}
		if expectFile != "" {
			m[name+"/expect/"+expectFile] = body
		}
		return m
	}
	cases := map[string]string{}
	for _, m := range []map[string]string{
		c("same", "x", "a.txt", "x"),
		c("differs", "x", "a.txt", "y"),
		c("missing", "none", "b.txt", "b"),
		c("extra", "x", "", ""),
		c("extra-dir", "sub", "", ""),
		c("nested", "sub", "sub/c", "c"),
		c("random", "rand", "t", "1\n"),
		c("case-token", "where", "where", "$CASE/out\n"),
		c("link", "link", "", ""),
		c("empty", "none", "", ""),
		c("removed", "rm", "", ""),
		c("swapped", "swap", "", ""),
		c("locked", "locked", "", ""),
	} {
		maps.Copy(cases, m)
	}
	files, ex := outFiles(cases)
	r, err := recipe.Open(filepath.Join(recipeRoot(t, "out", files, ex...), "recipes", "out@1"))
	if err != nil {
		t.Fatal(err)
	}
	ka, err := runKA(t, r)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, c := range ka.Cases {
		got[c.Name] = c.Outcome
	}
	for name, want := range map[string]string{
		"same": "ok", "nested": "ok", "case-token": "ok", "empty": "ok",
		"differs": "out differs: a.txt", "missing": "out differs: b.txt",
		"extra": "out not empty: a.txt", "extra-dir": "out not empty: sub",
		"random": "not stateless", "link": "out holds a non-regular file: l",
		"removed": "out is not a directory", "swapped": "out is not a directory",
		"locked": "out holds an unreadable entry: u",
	} {
		if got[name] != want {
			t.Errorf("%s: outcome = %q, want %q", name, got[name], want)
		}
	}
}

func TestKnownAnswersRefusesBadExpect(t *testing.T) {
	files, ex := outFiles(map[string]string{
		"c/expect.json": expect("standalone", "out", []string{"none"}, 0, "done\n", ""),
		"c/input/a":     "",
		"c/expect":      "not a directory",
	})
	r, err := recipe.Open(filepath.Join(recipeRoot(t, "out", files, ex...), "recipes", "out@1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runKA(t, r); code(err) != 1 || !strings.Contains(err.Error(), "expect is not a directory") {
		t.Errorf("err = %v, want exit 1", err)
	}
}

// The run keeps each check's $HYP_OUT tree and stdout in the record,
// even when the check fails.
func TestRunKeepsOutAndStdout(t *testing.T) {
	read := func(p string) string {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Error(err)
		}
		return string(data)
	}
	root := recipeRoot(t, "two", map[string]string{
		"recipe.md":                    recipeMD("two", "## Evidence\n\nRun: ./a.sh\nRun: ./b.sh\n\n## Standalone\n\nRun one: ./a.sh\n"),
		"a.sh":                         "#!/bin/sh\n[ -e \"$1/sandbox.json\" ] || exit 0\nprintf a > \"$HYP_OUT/a.txt\"\necho from-a\n",
		"b.sh":                         "#!/bin/sh\n[ -e \"$1/sandbox.json\" ] || { echo b; exit 30; }\nmkdir \"$HYP_OUT/d\"\nprintf b > \"$HYP_OUT/d/b.txt\"\necho from-b\necho err-b >&2\nexit 30\n",
		"known-answers/a/expect.json":  expect("evidence", "./a.sh", nil, 0, "", ""),
		"known-answers/a/input/x":      "",
		"known-answers/b/expect.json":  expect("evidence", "./b.sh", nil, 30, "b\n", ""),
		"known-answers/b/input/x":      "",
		"known-answers/sa/expect.json": expect("standalone", "one", []string{"/nonexistent"}, 0, "", ""),
		"known-answers/sa/input/x":     "",
	}, "a.sh", "b.sh")
	kept := keptDir(t)
	ev := filepath.Join(t.TempDir(), "ev")
	_, err := Run(RunOptions{Dir: root, Ref: "two@1", Step: "evidence", Args: []string{kept}, Record: ev})
	failed(t, err)
	for p, want := range map[string]string{
		"out/0/a.txt": "a", "stdout/0": "from-a\n", "out/1/d/b.txt": "b", "stdout/1": "from-b\n",
	} {
		if got := read(filepath.Join(ev, p)); got != want {
			t.Errorf("%s = %q, want %q", p, got, want)
		}
	}
	if log := read(filepath.Join(ev, "command.log")); !strings.Contains(log, "from-a\n") || !strings.Contains(log, "from-b\n") || !strings.Contains(log, "err-b\n") {
		t.Errorf("command.log = %q, want both streams", log)
	}
	sa := filepath.Join(t.TempDir(), "sa")
	if _, err := Run(RunOptions{Dir: root, Ref: "two@1", Step: "standalone", Label: "one", Args: []string{kept}, Record: sa}); err != nil {
		t.Fatal(err)
	}
	if got := read(filepath.Join(sa, "out", "a.txt")); got != "a" {
		t.Errorf("out/a.txt = %q", got)
	}
	if got := read(filepath.Join(sa, "stdout")); got != "from-a\n" {
		t.Errorf("stdout = %q", got)
	}
}

// The run keeps a product by the gate's rules: a symlink in $HYP_OUT is
// left out and refuses the run (exit 1), unless the command failed, whose
// exit code wins.
func TestRunKeepsProductByTheGatesRules(t *testing.T) {
	root := recipeRoot(t, "lnk", passCase(map[string]string{
		"recipe.md": recipeMD("lnk", "## Standalone\n\nRun lnk: ./lnk.sh\n"),
		"lnk.sh":    "#!/bin/sh\n[ \"$1\" = case ] && exit 0\nprintf a > \"$HYP_OUT/a.txt\"\nln -s /etc \"$HYP_OUT/l\"\nexit \"$1\"\n",
	}, "standalone", "lnk", []string{"case"}, 0, ""), "lnk.sh")
	rec := filepath.Join(t.TempDir(), "rec")
	_, err := Run(RunOptions{Dir: root, Ref: "lnk@1", Step: "standalone", Label: "lnk", Args: []string{"0"}, Record: rec})
	if code(err) != 1 || !strings.Contains(err.Error(), "out holds a non-regular file: l") {
		t.Fatalf("err = %v, want exit 1 naming l", err)
	}
	if !exists(filepath.Join(rec, "out", "a.txt")) || exists(filepath.Join(rec, "out", "l")) {
		t.Error("want out/a.txt kept and out/l left out")
	}
	rec = filepath.Join(t.TempDir(), "rec")
	_, err = Run(RunOptions{Dir: root, Ref: "lnk@1", Step: "standalone", Label: "lnk", Args: []string{"7"}, Record: rec})
	if cf := failed(t, err); *cf.Row.Exit != 7 {
		t.Errorf("exit = %d, want 7", *cf.Row.Exit)
	}
}
