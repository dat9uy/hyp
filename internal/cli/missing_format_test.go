package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// oldMD is a recipe.md with no format: 1 line whose Run: line does not
// parse, so a refusal that names the format proves hyp read the frontmatter
// before anything else.
const oldMD = "---\nname: old\nversion: 1\n---\n\n## Standalone\n\nRun x: x.sh\n"

// TestUseRefusesMissingFormat runs every command that uses a recipe on a
// bundle whose recipe.md has no format: 1 line: each refuses it, exit 1, by
// that missing field, whatever its body.
func TestUseRefusesMissingFormat(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("HYP_ACTOR", "a")
	repo := t.TempDir()
	gitIn(t, repo, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(repo, "a.txt"), "a\n")
	gitIn(t, repo, "add", "a.txt")
	gitIn(t, repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "init")
	sha := gitIn(t, repo, "rev-parse", "HEAD")
	writeLib(t, repo, "old@1", oldMD)
	writeFile(t, filepath.Join(repo, "recipes", "old@1", "x.sh"), "#!/bin/sh\n")
	tree := filepath.Join(repo, "tree")
	if err := os.MkdirAll(tree, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tree)
	must(t, "init")
	g := must(t, "add", "goal", "--title", "G", "--why", "start")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "by hand", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	stub := filepath.Join(t.TempDir(), "hyp")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho hyp 0.0.0-test\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	writeFile(t, filepath.Join(out, "sandbox.json"), "{}\n")
	recs := t.TempDir()

	for name, args := range map[string][]string{
		"add exp":               {"add", "exp", "--parent", h, "--title", "E", "--recipe", "old@1", "--rubric", "reading", "--why", "w"},
		"sandbox create":        {"sandbox", "create", "s-old", "--host", repo, "--base", sha, "--hyp", stub, "--recipe", "old@1", "--recipes", repo},
		"recipe run":            {"recipe", "run", "--record", filepath.Join(recs, "run"), "old@1", "x"},
		"recipe evidence":       {"recipe", "evidence", "--record", filepath.Join(recs, "ev"), "old@1", out},
		"recipe known-answers":  {"recipe", "known-answers", "old@1"},
		"record foreign bundle": {"record", "--exp", e, "--verdict", h + "=supports", "--observation", "o", "--evidence", "recipes/old@1/x.sh", "--why", "w"},
	} {
		t.Run(name, func(t *testing.T) {
			r := hyp(t, args...)
			want(t, r, 1, "recipe old@1: recipe.md has no format: 1 line")
			if !strings.HasPrefix(r.stderr, "refused: ") {
				t.Errorf("stderr %q, want refused:", r.stderr)
			}
		})
	}
	if _, err := os.Lstat(filepath.Join(os.Getenv("HOME"), ".hyp-sandbox", "s-old")); !os.IsNotExist(err) {
		t.Errorf("sandbox s-old was made: %v", err)
	}
	for _, d := range []string{"run", "ev"} {
		if _, err := os.Lstat(filepath.Join(recs, d)); !os.IsNotExist(err) {
			t.Errorf("record %s was made: %v", d, err)
		}
	}
}

// hyp recipe hash hashes a bundle with no format: 1 line, exit 0 and no
// warning: hashing never parses, so its pins still verify.
func TestRecipeHashIgnoresMissingFormat(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "old@1", "---\nname: old\nversion: 1\n---\n\nRun it by hand.\n")
	bundle := filepath.Join(dir, "recipes", "old@1")
	r := hyp(t, "recipe", "hash", bundle)
	if r.code != 0 || r.stdout != bundleHash(t, bundle)+"\n" {
		t.Fatalf("recipe hash: %+v", r)
	}
	if r.stderr != "" {
		t.Fatalf("stderr %q, want none", r.stderr)
	}
}
