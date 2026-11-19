package testutil

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Repo is a temporary git repository for tests.
type Repo struct{ Dir string }

// gitEnv pins what git reads, so the user's config never leaks in and commit
// hashes are the same on every run.
var gitEnv = []string{
	"GIT_CONFIG_GLOBAL=/dev/null",
	"GIT_CONFIG_NOSYSTEM=1",
	"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z",
	"GIT_COMMITTER_DATE=2026-01-01T00:00:00Z",
}

// git runs git in dir with the pinned environment and returns trimmed stdout.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), gitEnv...)
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, stderr)
	}
	return strings.TrimSpace(string(out))
}

// NewRepo makes an empty repository on branch main. The environment of the
// test process is pinned too, so the git calls hyp makes see the same config.
func NewRepo(t *testing.T) *Repo {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_DATE", "2026-01-01T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-01-01T00:00:00Z")
	dir := t.TempDir()
	// A temp dir can sit behind a symlink (macOS /var); git reports real paths.
	if real, err := filepath.EvalSymlinks(dir); err == nil {
		dir = real
	}
	git(t, dir, "init", "-q", "-b", "main")
	git(t, dir, "config", "user.name", "t")
	git(t, dir, "config", "user.email", "t@example.com")
	return &Repo{Dir: dir}
}

// Commit writes files, commits everything and returns the commit's sha.
func (r *Repo) Commit(t *testing.T, files map[string]string) string {
	t.Helper()
	for p, body := range files {
		full := filepath.Join(r.Dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git(t, r.Dir, "add", "-A")
	git(t, r.Dir, "commit", "-q", "--allow-empty", "-m", "c")
	return git(t, r.Dir, "rev-parse", "HEAD")
}

// WriteExec writes a file with mode 0755.
func (r *Repo) WriteExec(t *testing.T, path, body string) {
	t.Helper()
	full := filepath.Join(r.Dir, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(full, 0o755); err != nil {
		t.Fatal(err)
	}
}

// Symlink makes a symlink at path pointing to target.
func (r *Repo) Symlink(t *testing.T, path, target string) {
	t.Helper()
	full := filepath.Join(r.Dir, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, full); err != nil {
		t.Fatal(err)
	}
}

// Snapshot hashes every path, mode and content under the repository's git
// directory in sorted path order. Equal snapshots mean nothing under it
// changed.
func Snapshot(t *testing.T, dir string) string {
	t.Helper()
	h := sha256.New()
	root := filepath.Join(dir, ".git")
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		info, err := d.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s\x00%o\x00", rel, info.Mode())
		switch {
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(h, "%s\x00", target)
		case info.Mode().IsRegular():
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(h, "%x\x00", sha256.Sum256(b))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Git runs a git command in dir, for tests that need a command Repo lacks.
func Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return git(t, dir, args...)
}
