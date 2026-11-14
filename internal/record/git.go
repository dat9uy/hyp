package record

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
)

// Repo is one git repository.
type Repo struct {
	// Top is the root of the work tree discovery started in.
	Top string
	// Common is the repository's common git directory, which every
	// worktree of the clone shares.
	Common string
	// Record reports whether the repository is a record repository: it
	// holds Ref and carries no sandbox mark.
	Record bool
}

// sandboxMark is the file in a clone's common git directory that marks it a
// sandbox clone, which reads and writes the work-tree store folder.
const sandboxMark = "hyp-sandbox"

// MarkSandbox marks the clone at wt as a sandbox clone: hyp there reads and
// writes the work-tree store folder, and hyp init writes no record.
func MarkSandbox(wt string) error {
	common, err := gitOut(wt, nil, nil, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(common, sandboxMark), nil, 0o644)
}

// openRepo opens the repository at top. A .git that git cannot open as a
// repository, or no git on PATH, gives a repository with no record, which
// reads and writes the work-tree store folder as before the record.
func openRepo(top string) (*Repo, error) {
	r := &Repo{Top: top}
	if _, err := exec.LookPath("git"); err != nil {
		return r, nil
	}
	common, err := gitOut(top, nil, nil, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return r, nil
	}
	r.Common = common
	if _, err := os.Stat(filepath.Join(common, sandboxMark)); err == nil {
		return r, nil
	}
	tip, err := r.tip()
	if err != nil {
		return nil, err
	}
	r.Record = tip != ""
	return r, nil
}

// OpenRepo opens the repository whose work tree holds start, or returns
// ok false when no .git is at or above start.
func OpenRepo(start string) (repo *Repo, ok bool, err error) {
	top, ok := FindTop(start)
	if !ok {
		return nil, false, nil
	}
	repo, err = openRepo(top)
	return repo, err == nil, err
}

// tip is the commit Ref names, "" when the repository has no record.
func (r *Repo) tip() (string, error) {
	out, err := gitOut(r.Top, nil, nil, "for-each-ref", "--format=%(objectname)", Ref)
	if err != nil {
		return "", err
	}
	return out, nil
}

// ensureRecord writes the empty record when the repository has none and is
// not a sandbox clone.
func (r *Repo) ensureRecord() error {
	if r.Record || r.Common == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Join(r.Common, sandboxMark)); err == nil {
		return nil
	}
	tree, err := gitOut(r.Top, nil, nil, "mktree")
	if err != nil {
		return err
	}
	c, err := r.commitTree(tree, nil, "hyp: record")
	if err != nil {
		return err
	}
	if _, err := gitOut(r.Top, nil, nil, "update-ref", Ref, c, ""); err != nil {
		return err
	}
	r.Record = true
	return nil
}

// author is the identity on every record commit, so a commit needs no git
// configuration.
var author = []string{"GIT_AUTHOR_NAME=hyp", "GIT_AUTHOR_EMAIL=hyp@localhost", "GIT_COMMITTER_NAME=hyp", "GIT_COMMITTER_EMAIL=hyp@localhost"}

func (r *Repo) commitTree(tree string, parents []string, msg string) (string, error) {
	args := []string{"commit-tree", tree}
	for _, p := range parents {
		args = append(args, "-p", p)
	}
	args = append(args, "-m", msg)
	return gitOut(r.Top, author, nil, args...)
}

// entry is one entry of a record commit's tree.
type entry struct {
	mode, typ, sha string
}

// readTip reads the record's tip and its tree entries, trees included.
func (s *Store) readTip() error {
	tip, err := s.Repo.tip()
	if err != nil {
		return err
	}
	s.tip = tip
	s.entries = map[string]entry{}
	if tip == "" {
		return nil
	}
	s.entries, err = s.Repo.listTree(tip)
	return err
}

// listTree lists every entry of treeish, trees included, by slash path.
func (r *Repo) listTree(treeish string) (map[string]entry, error) {
	out, err := gitRaw(r.Top, nil, nil, "ls-tree", "-r", "-t", "-z", "--full-tree", treeish)
	if err != nil {
		return nil, err
	}
	m := map[string]entry{}
	for _, rec := range bytes.Split(out, []byte{0}) {
		if len(rec) == 0 {
			continue
		}
		meta, p, ok := bytes.Cut(rec, []byte{'\t'})
		if !ok {
			continue
		}
		f := strings.Fields(string(meta))
		if len(f) != 3 {
			continue
		}
		m[string(p)] = entry{mode: f[0], typ: f[1], sha: f[2]}
	}
	return m, nil
}

// IsCommit reports whether rev names a commit in the repository and
// returns its full id.
func (r *Repo) IsCommit(rev string) (string, bool) {
	out, err := gitOut(r.Top, nil, nil, "rev-parse", "--verify", "--quiet", rev+"^{commit}")
	if err != nil || out == "" {
		return "", false
	}
	return out, true
}

// ErrNotCommit means a revision names no commit of the repository.
var ErrNotCommit = errors.New("not a commit")

// ResolveCommit resolves rev to a full commit id in the repository whose
// work tree holds dir. It returns ErrNotCommit when no repository holds dir
// or rev names no commit there, and an environment error when git is not
// on PATH.
func ResolveCommit(dir, rev string) (string, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", herr.Env("git is not on PATH")
	}
	top, ok := FindTop(dir)
	if !ok {
		return "", ErrNotCommit
	}
	id, ok := (&Repo{Top: top}).IsCommit(rev)
	if !ok {
		return "", ErrNotCommit
	}
	return id, nil
}

// isAncestor reports whether commit a is an ancestor of commit b.
func (r *Repo) isAncestor(a, b string) bool {
	_, err := gitRaw(r.Top, nil, nil, "merge-base", "--is-ancestor", a, b)
	return err == nil
}

// readBlobs returns the content of each blob sha, by one git cat-file run.
func (r *Repo) readBlobs(shas []string, each func(sha string, data []byte) error) error {
	if len(shas) == 0 {
		return nil
	}
	cmd := exec.Command("git", "-C", r.Top, "cat-file", "--batch")
	cmd.Stdin = strings.NewReader(strings.Join(shas, "\n") + "\n")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return herr.Env("git cat-file: %v", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return herr.Env("git cat-file: %v", err)
	}
	br := bufio.NewReader(out)
	var readErr error
	for range shas {
		head, err := br.ReadString('\n')
		if err != nil {
			readErr = herr.Env("git cat-file: %v", err)
			break
		}
		f := strings.Fields(head)
		if len(f) != 3 {
			readErr = herr.Env("git cat-file: %s", strings.TrimSpace(head))
			break
		}
		n, err := strconv.Atoi(f[2])
		if err != nil {
			readErr = herr.Env("git cat-file: %s", strings.TrimSpace(head))
			break
		}
		data := make([]byte, n+1)
		if _, err := io.ReadFull(br, data); err != nil {
			readErr = herr.Env("git cat-file: %v", err)
			break
		}
		if err := each(f[0], data[:n]); err != nil {
			readErr = err
			break
		}
	}
	_, _ = io.Copy(io.Discard, br)
	if err := cmd.Wait(); err != nil && readErr == nil {
		readErr = herr.Env("git cat-file: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return readErr
}

// gitOut runs git in dir with extra environment and stdin, and returns its
// trimmed output.
func gitOut(dir string, env []string, stdin []byte, args ...string) (string, error) {
	out, err := gitRaw(dir, env, stdin, args...)
	return strings.TrimSpace(string(out)), err
}

// errGit is a git command that ran and exited non-zero.
type errGit struct {
	args   []string
	stderr string
	err    error
}

func (e *errGit) Error() string {
	return fmt.Sprintf("git %s failed: %v: %s", strings.Join(e.args, " "), e.err, e.stderr)
}

// gitRaw runs git in dir. Each run hardens the objects and refs it writes
// (core.fsync), so a record write that exits 0 survives an unclean
// shutdown, as the work-tree store's synced events file does.
func gitRaw(dir string, env []string, stdin []byte, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "core.fsync=objects,reference", "-c", "core.fsyncMethod=fsync"}, args...)...)
	cmd.Env = append(os.Environ(), env...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return out, herr.Env("%v", &errGit{args, strings.TrimSpace(stderr.String()), err})
		}
		return out, herr.Env("git: %v", err)
	}
	return out, nil
}
