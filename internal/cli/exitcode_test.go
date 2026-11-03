package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestExitCodeContract drives every command that finds a recipe over the
// same bad inputs. A recipe that breaks a rule is a refusal (1, refused:) in
// every command, a recipe that does not exist is a bad argument (2), and
// hyp check refuses nothing.
func TestExitCodeContract(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("HYP_ACTOR", "a")

	// repo is a git work tree, the sandbox host, and holds the library; the
	// store at repo/tree holds a different copy of twin@1.
	repo := t.TempDir()
	gitIn(t, repo, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(repo, "a.txt"), "a\n")
	gitIn(t, repo, "add", "a.txt")
	gitIn(t, repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "init")
	sha := gitIn(t, repo, "rev-parse", "HEAD")
	bundleIn(t, repo, "bad", "## Launch\n\nRun: launch.sh\n", map[string]string{"launch.sh": "#!/bin/sh\n"})
	bundleIn(t, repo, "twin", "", nil)
	bundleIn(t, repo, "chg", "## Launch\n\nRun: ./launch.sh\n", map[string]string{"launch.sh": "#!/bin/sh\n"})
	// old@1 has no format: 1 line, a required field.
	writeLib(t, repo, "old@1", "---\nname: old\nversion: 1\n---\n\n## Standalone\n\nRun x: ./x.sh\n")
	tree := filepath.Join(repo, "tree")
	bundleIn(t, tree, "twin", "A local edit.\n", nil)
	t.Chdir(tree)
	must(t, "init")
	g := must(t, "add", "goal", "--title", "G", "--why", "start")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")

	// chg@1 is pinned while it opens, then broken the way bad@1 is.
	must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "chg@1", "--decided-by", "reading", "--why", "w")
	writeFile(t, filepath.Join(repo, "recipes", "chg@1", "recipe.md"), "---\nname: chg\nversion: 1\nformat: 1\n---\n\n## Launch\n\nRun: launch.sh\n")

	out := t.TempDir()
	writeFile(t, filepath.Join(out, "sandbox.json"), "{}\n")
	recs := t.TempDir()
	stub := filepath.Join(t.TempDir(), "hyp")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho hyp 0.0.0-test\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// A sandbox whose record, as an older hyp wrote it, points a recipe's copy
	// outside the sandboxes root: drive refuses to run from it.
	must(t, "sandbox", "create", "esc", "--host", repo, "--base", sha, "--hyp", stub, "--recipe", "none")
	meta := filepath.Join(os.Getenv("HOME"), ".hyp-sandbox", "esc", "sandbox.json")
	data, err := os.ReadFile(meta)
	if err != nil {
		t.Fatal(err)
	}
	var rec map[string]any
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatal(err)
	}
	rec["recipes"] = []map[string]any{{"ref": "esc@1", "name": "esc", "version": "1", "path": t.TempDir()}}
	if data, err = json.Marshal(rec); err != nil {
		t.Fatal(err)
	}
	writeFile(t, meta, string(data))

	addExp := func(ref string) []string {
		return []string{"add", "exp", "--parent", h, "--title", "E", "--recipe", ref, "--why", "w"}
	}
	create := func(id, ref string) []string {
		return []string{"sandbox", "create", id, "--host", repo, "--base", sha, "--hyp", stub, "--recipe", ref, "--recipes", tree}
	}
	run := func(row, ref string) []string {
		return []string{"recipe", "run", "--record", filepath.Join(recs, row), ref, "x"}
	}
	evidence := func(row, ref string) []string {
		return []string{"recipe", "evidence", "--record", filepath.Join(recs, row), ref, out}
	}
	for _, c := range []struct {
		row  string
		args []string
		code int
	}{
		{"r1", addExp("bad@1"), 1},
		{"r2", addExp("nope@1"), 2},
		{"r4", addExp("twin@1"), 1},
		{"r5", []string{"recipe", "hash", filepath.Join(repo, "recipes", "bad@1")}, 1},
		{"r6", create("s6", "bad@1"), 1},
		{"r7", create("s7", "nope@1"), 2},
		{"r9", create("s9", "twin@1"), 1},
		{"r10", run("r10", "bad@1"), 1},
		{"r11", run("r11", "nope@1"), 2},
		{"r13", run("r13", "twin@1"), 1},
		{"r14", evidence("r14", "bad@1"), 1},
		{"r15", evidence("r15", "nope@1"), 2},
		{"r16", []string{"sandbox", "drive", "esc"}, 1},
		{"r17", []string{"check"}, 0},
		{"r18", addExp("old@1"), 1},
		{"r19", create("s-old", "old@1"), 1},
		{"r20", run("r20", "old@1"), 1},
	} {
		t.Run(c.row, func(t *testing.T) {
			r := hyp(t, c.args...)
			if r.code != c.code {
				t.Fatalf("hyp %v: exit %d, want %d; stderr %q", c.args, r.code, c.code, r.stderr)
			}
			if c.code == 1 && !strings.HasPrefix(r.stderr, "refused:") {
				t.Fatalf("hyp %v: stderr %q, want refused:", c.args, r.stderr)
			}
			if c.code == 0 && !strings.Contains(r.stdout+r.stderr, "warn:") {
				t.Fatalf("hyp %v: stdout %q, stderr %q, want warn:", c.args, r.stdout, r.stderr)
			}
		})
	}
}
