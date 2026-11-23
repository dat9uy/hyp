package probe

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/example"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/testutil"
)

// fixtureProbe is the Probe of the Probe contract in hyp spec: it prints the
// stdout of run, a newline when that stdout lacks one, then "exit N".
const fixtureProbe = `#!/bin/sh
o=$(mktemp); sh -c "$HYP_RUN" >"$o"; c=$?; cat "$o"; [ -s "$o" ] && [ -n "$(tail -c1 "$o")" ] && echo; rm -f "$o"; echo "exit $c"
`

const folder = "goals/G-TEST00/line-1/a"

// fixture commits a subject and one example folder and returns the commit and
// the example folder's tree hash.
func fixture(t *testing.T, txtar, probeScript string, subject map[string]string) (r *testutil.Repo, commit, hash string) {
	t.Helper()
	r = testutil.NewRepo(t)
	if probeScript != "" {
		r.WriteExec(t, folder+"/probe/run", probeScript)
	}
	files := map[string]string{folder + "/example.txtar": txtar, ".tree/events/x.jsonl": "{}\n"}
	for k, v := range subject {
		files[k] = v
	}
	if len(subject) == 0 {
		files["main.go"] = "package main\n"
	}
	commit = r.Commit(t, files)
	hash = testutil.Git(t, r.Dir, "rev-parse", commit+":"+folder)
	return r, commit, hash
}

func txtar(run string, extra ...string) string {
	return "-- words --\nw\n-- run --\n" + run + "\n-- floor: contains --\nx\n" + strings.Join(extra, "")
}

func run(t *testing.T, run string, extra ...string) Out {
	t.Helper()
	r, c, h := fixture(t, txtar(run, extra...), fixtureProbe, nil)
	out, err := Run(Opts{Repo: r.Dir, Commit: c, ExampleHash: h, Timeout: 20 * time.Second})
	if err != nil {
		t.Fatalf("Run(%q): %v", run, err)
	}
	return out
}

func TestProbeRunsOwnerExampleLine1(t *testing.T) {
	data, err := os.ReadFile("../example/testdata/line1.txtar")
	if err != nil {
		t.Fatal(err)
	}
	r, c, h := fixture(t, string(data), fixtureProbe, nil)
	out, err := Run(Opts{Repo: r.Dir, Commit: c, ExampleHash: h, Timeout: 20 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Stdout) != 0 || out.Exit != 1 {
		t.Fatalf("stdout %q exit %d; want empty and 1", out.Stdout, out.Exit)
	}
	ex, err := example.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if v := example.Compare(out.Stdout, ex); v != example.Met {
		t.Fatalf("verdict %s, want met", v)
	}
}

func TestProbeSeesSubjectWithoutGoals(t *testing.T) {
	out := run(t, "ls -A")
	got := string(out.Stdout)
	if !strings.Contains(got, "main.go") {
		t.Fatalf("ls = %q, want the subject's main.go", got)
	}
	for _, bad := range []string{"goals", ".tree"} {
		if strings.Contains(got, bad) {
			t.Fatalf("ls = %q; the subject must not hold %s", got, bad)
		}
	}
}

func TestProbeInputsStayOutOfSubject(t *testing.T) {
	out := run(t, `ls; cat "$HYP_EXAMPLE/input/data/p.txt"`, "-- input/data/p.txt --\npayload\n")
	got := string(out.Stdout)
	if strings.Contains(got, "input") {
		t.Fatalf("ls = %q; inputs must stay out of the subject", got)
	}
	if !strings.HasSuffix(got, "payload") || out.Exit != 0 {
		t.Fatalf("stdout %q exit %d; want the input read from HYP_EXAMPLE", got, out.Exit)
	}
}

func TestProbeSplitsExitLine(t *testing.T) {
	out := run(t, "false")
	if len(out.Stdout) != 0 || out.Exit != 1 {
		t.Fatalf("stdout %q exit %d; want empty and 1", out.Stdout, out.Exit)
	}
	if out := run(t, "echo hi; exit 4"); string(out.Stdout) != "hi" || out.Exit != 4 {
		t.Fatalf("stdout %q exit %d; want hi and 4", out.Stdout, out.Exit)
	}
}

func TestProbeDropsOneFinalNewline(t *testing.T) {
	for _, tc := range []struct{ run, want string }{
		{"echo abc", "abc"},
		{"printf abc", "abc"},
		{`printf 'abc\n\n'`, "abc\n"},
		{`printf 'a\nb'`, "a\nb"},
		{`printf '\n'`, ""},
	} {
		if out := run(t, tc.run); string(out.Stdout) != tc.want {
			t.Errorf("run %q: stdout %q, want %q", tc.run, out.Stdout, tc.want)
		}
	}
}

func TestProbeKeepsStderrApart(t *testing.T) {
	out := run(t, "echo x >&2")
	if len(out.Stdout) != 0 || string(out.Stderr) != "x\n" {
		t.Fatalf("stdout %q stderr %q; want empty and x", out.Stdout, out.Stderr)
	}
}

func TestProbePassesOnlyExternalEnv(t *testing.T) {
	t.Setenv("FOO", "foo-value")
	t.Setenv("BAR", "bar-value")
	out := run(t, "env", "-- external --\nFOO\n")
	got := string(out.Stdout)
	for _, want := range []string{"FOO=foo-value", "PATH=", "HOME=", "HYP_SUBJECT=", "HYP_EXAMPLE=", "HYP_RUN=env"} {
		if !strings.Contains(got, want) {
			t.Errorf("env lacks %s:\n%s", want, got)
		}
	}
	if strings.Contains(got, "BAR=") {
		t.Errorf("env holds BAR, which no external section names:\n%s", got)
	}
}

func TestProbeTimeoutKillsGroup(t *testing.T) {
	r, c, h := fixture(t, txtar("sleep 30"), fixtureProbe, nil)
	start := time.Now()
	_, err := Run(Opts{Repo: r.Dir, Commit: c, ExampleHash: h, Timeout: time.Second})
	if err == nil || fault.KindOf(err) != fault.Environment || !testutil.Mentions(err, h) {
		t.Fatalf("err = %v, want an environment error naming the example hash", err)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Fatalf("took %s; the process group must be killed at the timeout", d)
	}
}

func TestProbeFailureIsEnv(t *testing.T) {
	for name, script := range map[string]string{
		"exits non-zero":  "#!/bin/sh\nexit 7\n",
		"no exit line":    "#!/bin/sh\necho hello\n",
		"bad exit line":   "#!/bin/sh\necho 'exit x'\n",
		"exit line first": "#!/bin/sh\necho 'exit 0'\necho tail\n",
		"missing":         "",
	} {
		r, c, h := fixture(t, txtar("true"), script, nil)
		_, err := Run(Opts{Repo: r.Dir, Commit: c, ExampleHash: h, Timeout: 10 * time.Second})
		if err == nil || fault.KindOf(err) != fault.Environment || !testutil.Mentions(err, h) {
			t.Errorf("%s: err = %v, want an environment error naming %s", name, err, h)
		}
	}
}

func TestProbeBadExampleFileIsRefusal(t *testing.T) {
	r, c, h := fixture(t, "no sections here\n", fixtureProbe, nil)
	_, err := Run(Opts{Repo: r.Dir, Commit: c, ExampleHash: h, Timeout: 10 * time.Second})
	if err == nil || fault.KindOf(err) != fault.Refusal {
		t.Fatalf("err = %v, want a refusal", err)
	}
}

func TestProbeLeavesNoTempFiles(t *testing.T) {
	r, c, h := fixture(t, txtar("echo hi"), fixtureProbe, nil)
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	if _, err := Run(Opts{Repo: r.Dir, Commit: c, ExampleHash: h, Timeout: 10 * time.Second}); err != nil {
		t.Fatal(err)
	}
	left, err := os.ReadDir(tmp)
	if err != nil || len(left) != 0 {
		t.Fatalf("temp dir holds %d entries after Run: %v %v", len(left), left, err)
	}
}

func TestProbeWritesNoGitState(t *testing.T) {
	r, c, h := fixture(t, txtar("echo hi"), fixtureProbe, nil)
	before := testutil.Snapshot(t, r.Dir)
	out, err := Run(Opts{Repo: r.Dir, Commit: c, ExampleHash: h, Timeout: 10 * time.Second})
	if err != nil || string(out.Stdout) != "hi" {
		t.Fatalf("Run = %q, %v; it must have run", out.Stdout, err)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("Run changed the repository")
	}
}
