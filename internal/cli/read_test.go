package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadCommandsHappyPath(t *testing.T) {
	dir, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	mid := hyp(t, "status")
	want(t, mid, 0, "FRONTIER")
	must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--observation", "holds", "--why", "ran it")

	want(t, hyp(t, "status"), 0, h1+" supported")
	want(t, hyp(t, "show", h1), 0, "path: G-")
	want(t, hyp(t, "diff", "--since", e), 0, "proposed → supported")
	want(t, hyp(t, "render", "--markdown"), 0, "- "+h1+" H1 — supported")
	want(t, hyp(t, "signals"), 0, "")
	want(t, hyp(t, "events", "--json"), 0, `"format":1`)
	// hyp keeps no index (ADR 0006): no read writes one.
	if m, _ := filepath.Glob(filepath.Join(dir, ".tree", "index.sqlite*")); len(m) != 0 {
		t.Fatalf("a read wrote %v", m)
	}
	if asOf := hyp(t, "status", "--as-of", e); asOf.code != 0 || asOf.stdout != mid.stdout {
		t.Fatalf("status --as-of %s differs from the status taken then:\n%s\n%s", e, mid.stdout, asOf.stdout)
	}
	want(t, hyp(t, "status", "--as-of", "2000-01-01T00:00:00Z"), 0, "GOAL none")
	want(t, hyp(t, "check"), 0, "check: 0 errors")
	want(t, hyp(t, "status", "--max-lines", "x"), 2, "invalid value")
	if r := hyp(t, "show", "Q-NOPE00"); r.code != 2 || !strings.Contains(r.stderr, "unknown node") {
		t.Fatalf("show unknown: %+v", r)
	}
}

// A read writes nothing, so a read-only store reads without a warning.
func TestReadOnlyStoreReadsWithoutWarning(t *testing.T) {
	dir, _, q := setup(t)
	must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	tree := filepath.Join(dir, ".tree")
	if err := os.Chmod(tree, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(tree, 0o755) })
	for _, args := range [][]string{{"status"}, {"show", q}, {"render", "--markdown"}, {"signals"}, {"events", "--json"}} {
		if r := hyp(t, args...); r.code != 0 || r.stderr != "" {
			t.Fatalf("hyp %v: exit %d, stderr %q", args, r.code, r.stderr)
		}
	}
}

func TestAsOfBareDateIsUsage(t *testing.T) {
	setup(t)
	want(t, hyp(t, "status", "--as-of", "2026-10-24"), 2, "not an RFC3339 time")
}

// A line that does not parse is not replayed; status says so on stderr and
// keeps its exit code and its stdout.
func TestStatusWarnsOnUnparseableLine(t *testing.T) {
	dir, g, _ := setup(t)
	clean := hyp(t, "status")
	cleanAsOf := hyp(t, "status", "--as-of", g)
	if clean.code != 0 || clean.stderr != "" || cleanAsOf.code != 0 || cleanAsOf.stderr != "" {
		t.Fatalf("clean store: %+v %+v", clean, cleanAsOf)
	}
	f, err := os.OpenFile(filepath.Join(dir, ".tree", "events", "a.jsonl"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	// setup wrote lines 1 and 2; line 4 is blank and still counts.
	f.WriteString("{not json\n\nnope\n")
	f.Close()

	for _, c := range []struct {
		args   []string
		before result
	}{
		{[]string{"status"}, clean},
		{[]string{"status", "--as-of", g}, cleanAsOf},
	} {
		r := hyp(t, c.args...)
		lines := strings.Split(strings.TrimSuffix(r.stderr, "\n"), "\n")
		if r.code != 0 || r.stdout != c.before.stdout || len(lines) != 2 ||
			!strings.HasPrefix(lines[0], "warn: skipped a.jsonl:3: ") ||
			!strings.HasPrefix(lines[1], "warn: skipped a.jsonl:5: ") ||
			!strings.HasSuffix(lines[0], "; run hyp check") ||
			!strings.HasSuffix(lines[1], "; run hyp check") {
			t.Fatalf("hyp %v: exit %d, stderr %q, stdout changed: %v", c.args, r.code, r.stderr, r.stdout != c.before.stdout)
		}
	}
	want(t, hyp(t, "check"), 1, "error: a.jsonl:3 ")
}

// --target none skips the add-time refusal, not the one-rival signal.
func TestSignalsOneRivalAfterTargetNone(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	want(t, hyp(t, "signals"), 0, "one-rival "+e+" at add: "+h1+"=proposed (parent), "+h2+"=proposed; targets: none; could target: "+h2+"\n")
	want(t, hyp(t, "signals", "--json"), 0, `"key": "one-rival:`+e+`"`)
	want(t, hyp(t, "signals", "extra"), 2, "usage: hyp signals [--json]")
}

func TestSignalsOnAStoreWithNoRounds(t *testing.T) {
	setup(t)
	r := hyp(t, "signals", "--json")
	want(t, r, 0, `"rounds": []`)
	want(t, r, 0, `"signals": []`)
	if r := hyp(t, "signals"); r.stdout != "" {
		t.Fatalf("stdout %q, want nothing", r.stdout)
	}
}
