package gitread

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/testutil"
)

func TestGitAllowsRevParse(t *testing.T) {
	r := testutil.NewRepo(t)
	want := r.Commit(t, map[string]string{"a.txt": "a"})
	out, err := Git(r.Dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(out)); got != want {
		t.Fatalf("rev-parse HEAD = %q, want %q", got, want)
	}
}

func TestGitRefusesWriteCommand(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a.txt": "a"})
	before := testutil.Snapshot(t, r.Dir)
	for _, args := range [][]string{
		{"commit", "-m", "x"},
		{"config", "x", "y"},
		{"archive", "--output=f", "HEAD"},
		{"cat-file", "--remote"},
		{"rev-parse", "--git-dir"},
		{"ls-tree", "--output=x", "HEAD"},
		{"diff-tree", "-w", "HEAD", "HEAD"},
		{"merge-base", "A", "B"},
		{"patch-id"},
		{},
	} {
		_, err := Git(r.Dir, args...)
		if err == nil || herr.KindOf(err) != herr.Environment {
			t.Errorf("git %v: err = %v, want an environment refusal", args, err)
			continue
		}
		if !strings.Contains(err.Error(), "is not on hyp's read list") {
			t.Errorf("git %v: message %q does not name the read list", args, err)
		}
		if _, _, err := GitExit(r.Dir, args...); err == nil {
			t.Errorf("GitExit %v was allowed", args)
		}
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("a refused command changed the repository")
	}
}

func TestGitAllowsEveryReadForm(t *testing.T) {
	for _, args := range [][]string{
		{"rev-parse", "--show-toplevel"},
		{"rev-parse", "--show-object-format"},
		{"rev-parse", "--verify", "HEAD^{commit}"},
		{"rev-parse", "HEAD:goals/G/line-1"},
		{"cat-file", "-e", "abc^{commit}"},
		{"cat-file", "-t", "abc"},
		{"cat-file", "-p", "abc"},
		{"cat-file", "blob", "abc"},
		{"cat-file", "--batch"},
		{"ls-tree", "-z", "abc"},
		{"ls-tree", "-r", "-z", "abc", "goals/G/specs/"},
		{"merge-base", "--is-ancestor", "a", "b"},
		{"diff-tree", "-r", "--name-only", "a", "b"},
		{"diff-tree", "-r", "-c", "--name-only", "a"},
		{"diff-tree", "-p", "a", "b"},
		{"patch-id", "--stable"},
	} {
		if err := allowed(args); err != nil {
			t.Errorf("git %v refused: %v", args, err)
		}
	}
}

func TestGitCommandFailureIsEnvironment(t *testing.T) {
	r := testutil.NewRepo(t)
	_, err := Git(r.Dir, "rev-parse", "--verify", "nope")
	var he *herr.Error
	if !errors.As(err, &he) || he.Kind != herr.Environment {
		t.Fatalf("err = %v, want an environment error", err)
	}
}

func TestGitExitReturnsCode(t *testing.T) {
	r := testutil.NewRepo(t)
	c := r.Commit(t, map[string]string{"a.txt": "a"})
	code, _, err := GitExit(r.Dir, "cat-file", "-e", c+"^{commit}")
	if err != nil || code != 0 {
		t.Fatalf("existing commit: code %d err %v", code, err)
	}
	code, _, err = GitExit(r.Dir, "cat-file", "-e", strings.Repeat("0", 40)+"^{commit}")
	if err != nil || code == 0 {
		t.Fatalf("missing commit: code %d err %v, want a non-zero code and no error", code, err)
	}
}

func TestReadInPartialCloneFetchesNothing(t *testing.T) {
	origin := testutil.NewRepo(t)
	origin.Commit(t, map[string]string{"a.txt": "a"})
	origin.Commit(t, map[string]string{"b.txt": "b"})
	gitCmd(t, origin.Dir, "config", "uploadpack.allowFilter", "true")
	clone := t.TempDir() + "/clone"
	gitCmd(t, origin.Dir, "clone", "-q", "--filter=tree:0", "file://"+origin.Dir, clone)
	before := testutil.Snapshot(t, clone)
	if _, err := Git(clone, "ls-tree", "HEAD~1"); err == nil {
		t.Fatal("ls-tree of a tree the partial clone lacks succeeded; it must not fetch it")
	} else if herr.KindOf(err) != herr.Environment {
		t.Fatalf("error kind: %v, want environment", err)
	}
	if after := testutil.Snapshot(t, clone); after != before {
		t.Fatal("a read in a partial clone changed .git")
	}
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
