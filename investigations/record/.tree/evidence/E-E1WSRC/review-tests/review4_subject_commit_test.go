package cli

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// G-SZCBXZ#8 and #2: hyp import adds each event "with ... their subject
// commits", and every commit a Hypothesis subject names stays reachable
// from a ref. A clone that lacks the subject's commit imports the event
// first (the commit cannot be added yet). Once the commit is fetched, an
// import run again must add it, as it now adds named bytes and pinned
// bundles an earlier import lacked; otherwise the commit is lost when its
// branch goes and git gc runs, and hyp sandbox create --base fails.
func TestReviewImportAddsASubjectCommitAnEarlierImportLacked(t *testing.T) {
	gitConfig(t)
	a, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, a, "init", "-q")
	writeFile(t, filepath.Join(a, "s", "README.md"), "r\n")
	gitIn(t, a, "add", "-A")
	gitIn(t, a, "commit", "-qm", "base")
	gitIn(t, a, "checkout", "-q", "-b", "feat")
	writeFile(t, filepath.Join(a, "feat.txt"), "only on feat\n")
	gitIn(t, a, "add", "feat.txt")
	gitIn(t, a, "commit", "-qm", "feat")
	c := gitIn(t, a, "rev-parse", "HEAD")
	gitIn(t, a, "checkout", "-q", "main")
	// A work-tree store (no record ref) whose subject names c.
	t.Chdir(filepath.Join(a, "s"))
	must(t, "init")
	dropRecord(t, a)
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	must(t, "subject", h, "--base", c, "--why", "w")
	gitIn(t, a, "add", "-f", "-A")
	gitIn(t, a, "commit", "-qm", "store")

	// b clones main only, so it lacks c.
	b := filepath.Join(t.TempDir(), "b")
	if out, err := exec.Command("git", "clone", "-q", "--no-local", "--single-branch", "--branch", "main", a, b).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v %s", err, out)
	}
	if exec.Command("git", "-C", b, "cat-file", "-e", c+"^{commit}").Run() == nil {
		t.Fatalf("setup: the clone already holds %s", c)
	}
	t.Chdir(b)
	first := hyp(t, "import", "s")
	t.Logf("first import (commit absent): exit %d\nstdout: %sstderr: %s", first.code, first.stdout, first.stderr)
	if first.code != 0 {
		t.Fatalf("first import: exit %d", first.code)
	}
	// The commit arrives.
	gitIn(t, b, "fetch", "-q", "origin", "feat:refs/remotes/origin/feat")
	if exec.Command("git", "-C", b, "cat-file", "-e", c+"^{commit}").Run() != nil {
		t.Fatalf("setup: the fetch did not bring %s", c)
	}
	second := hyp(t, "import", "s")
	t.Logf("second import (commit present): exit %d\nstdout: %sstderr: %s", second.code, second.stdout, second.stderr)
	if exec.Command("git", "-C", b, "merge-base", "--is-ancestor", c, "refs/hyp/record").Run() != nil {
		// Show the consequence for line 2.
		gitIn(t, b, "update-ref", "-d", "refs/remotes/origin/feat")
		gitIn(t, b, "reflog", "expire", "--expire=now", "--all")
		gitIn(t, b, "gc", "-q", "--prune=now")
		gone := exec.Command("git", "-C", b, "cat-file", "-e", c+"^{commit}").Run() != nil
		t.Fatalf("after an import run again with %s present, the record does not reach it (second import printed %q); after its branch is deleted and gc, the commit is gone: %v",
			c[:12], strings.TrimSpace(second.stdout), gone)
	}
}
