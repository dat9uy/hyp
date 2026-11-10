package sandbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
)

// A check is an Evidence or Standalone command. It runs the same way in
// its known-answer cases and in hyp bundle run and evidence (ADR 0003): with
// checkEnv, in an empty working directory, with stdin /dev/null, and its
// product is its exit code, its stdout and the tree under $HYP_OUT. Sandbox
// steps (Launch, Doctor, Drive, Cleanup) run with sandboxEnv instead.

// SandboxHyp is the operator hyp, read once and written as Dir/hyp: the hyp
// a check finds first on PATH, and the bytes SHA256 names in known_answers.
type SandboxHyp struct {
	Dir, SHA256 string
}

// NewSandboxHyp returns the operator hyp (os.Executable), copied for a
// sandbox into a new temporary directory, which Remove deletes.
func NewSandboxHyp() (*SandboxHyp, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, herr.Env("%v", err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		return nil, herr.Env("%v", err)
	}
	dir, err := os.MkdirTemp("", "hyp-bin-")
	if err != nil {
		return nil, herr.Env("%v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hyp"), data, 0o755); err != nil {
		os.RemoveAll(dir)
		return nil, herr.Env("%v", err)
	}
	h := sha256.Sum256(data)
	return &SandboxHyp{Dir: dir, SHA256: hex.EncodeToString(h[:])}, nil
}

// Remove deletes the copy.
func (h *SandboxHyp) Remove() { os.RemoveAll(h.Dir) }

// checkEnv is the whole environment of a check run under root, and of a
// case's git clone: the sandbox hyp in bin first on PATH, root's own home and
// tmp, git without its system config, and root's out as HYP_OUT. It
// inherits nothing else.
func checkEnv(root, bin string) []string {
	return []string{
		"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + filepath.Join(root, "home"),
		"TMPDIR=" + filepath.Join(root, "tmp"),
		"GIT_CONFIG_NOSYSTEM=1",
		"HYP_OUT=" + filepath.Join(root, "out"),
	}
}

// newCheckRoot makes a new check root, with checkRoot's layout, and returns
// it and the function that removes it. A known-answer case and a run's check
// both get their root here, so a check cannot tell the gate from the run by
// its paths: they differ only in the random digits of the name.
func newCheckRoot() (root string, remove func(), err error) {
	tmp, err := os.MkdirTemp("", "hyp-check-")
	if err != nil {
		return "", nil, herr.Env("%v", err)
	}
	if root, err = checkRoot(tmp); err != nil {
		os.RemoveAll(tmp)
		return "", nil, err
	}
	return root, func() { os.RemoveAll(tmp) }, nil
}

// newBundleCopy copies the bundle r into a new temporary directory, for the
// gate and the run alike, and returns the copy and the function that removes
// it. Its path has the same shape in both, so a script that reads its own
// path cannot tell the gate from the run.
func newBundleCopy(ref string, r *recipe.Recipe) (cp *recipe.Recipe, dir string, remove func(), err error) {
	dir, err = os.MkdirTemp("", "hyp-bundle-")
	if err != nil {
		return nil, "", nil, herr.Env("%v", err)
	}
	if cp, err = copyRecipe(ref, r, dir); err != nil {
		os.RemoveAll(dir)
		return nil, "", nil, err
	}
	return cp, dir, func() { os.RemoveAll(dir) }, nil
}

// checkRoot makes root's empty cwd, home, tmp and out, and returns root with
// its symlinks resolved, the path a check sees as its working directory.
func checkRoot(root string) (string, error) {
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", herr.Env("%v", err)
	}
	for _, d := range []string{"cwd", "home", "tmp", "out"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o755); err != nil {
			return "", herr.Env("%v", err)
		}
	}
	return root, nil
}

// checkCmd is argv as a check under root: checkEnv, root's cwd, stdin
// /dev/null, stdout and stderr to the given files.
func checkCmd(ctx context.Context, root, bin string, argv []string, stdout, stderr *os.File) *exec.Cmd {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = filepath.Join(root, "cwd")
	cmd.Env = checkEnv(root, bin)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	return cmd
}

// outEntry is one entry of an out tree: a directory, or a regular file's
// bytes.
type outEntry struct {
	dir  bool
	data string
}

// readTree reads the tree a check left under dir, called name (out, or a
// case's expect): the slash-separated relative path of each directory and
// regular file, with each file's bytes, in which from, when not "", becomes
// $CASE. It never follows a symlink. problem is why the tree cannot be a
// check's product, or "": dir is not a directory (the check removed or
// replaced it), or the first path, sorted, of an entry that is neither a
// directory nor a regular file, or that cannot be read. The gate and the run
// both read a tree with it, so they agree on what a product is.
func readTree(dir, name, from string) (tree map[string]outEntry, problem string) {
	if fi, err := os.Lstat(dir); err != nil || !fi.IsDir() {
		return nil, name + " is not a directory"
	}
	tree = map[string]outEntry{}
	var odd, unreadable []string
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if p == dir {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		switch {
		case err != nil:
			unreadable = append(unreadable, rel)
		case d.IsDir():
			tree[rel] = outEntry{dir: true}
		case d.Type().IsRegular():
			data, err := os.ReadFile(p)
			if err != nil {
				unreadable = append(unreadable, rel)
				return nil
			}
			s := string(data)
			if from != "" {
				s = strings.ReplaceAll(s, from, "$CASE")
			}
			tree[rel] = outEntry{data: s}
		default:
			odd = append(odd, rel)
		}
		return nil
	})
	switch {
	case len(odd) > 0:
		return tree, name + " holds a non-regular file: " + slices.Min(odd)
	case len(unreadable) > 0:
		return tree, name + " holds an unreadable entry: " + slices.Min(unreadable)
	}
	return tree, ""
}

// writeTree writes tree under dst, which it makes: directories, and regular
// files with mode 0644.
func writeTree(dst string, tree map[string]outEntry) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, rel := range slices.Sorted(maps.Keys(tree)) {
		p := filepath.Join(dst, filepath.FromSlash(rel))
		e := tree[rel]
		var err error
		if e.dir {
			err = os.MkdirAll(p, 0o755)
		} else if err = os.MkdirAll(filepath.Dir(p), 0o755); err == nil {
			err = os.WriteFile(p, []byte(e.data), 0o644)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// firstDiff is the first path, sorted, that a and b do not hold alike, or "".
func firstDiff(a, b map[string]outEntry) string {
	all := maps.Clone(a)
	maps.Copy(all, b)
	for _, p := range slices.Sorted(maps.Keys(all)) {
		x, inA := a[p]
		y, inB := b[p]
		if inA != inB || x != y {
			return p
		}
	}
	return ""
}
