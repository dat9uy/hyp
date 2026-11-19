package gitread

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dat9uy/hyp/internal/fault"
)

// lsFilesForm is the one ls-files form hyp runs: every tracked file and every
// untracked file git does not ignore, NUL-separated.
var lsFilesForm = []string{"-z", "--cached", "--others", "--exclude-standard"}

// CopyWorktree writes the files of the checkout whose top level is dir under
// dest, as a commit of every change in the checkout would hold them: each
// tracked file and each untracked file git does not ignore, with its bytes
// on disk. A tracked file deleted on disk is left out. A path whose first
// component is in skip is left out. A file is written 0755 when its owner
// execute bit is set, as git reads it, else 0644; a symlink is written as a
// symlink. A folder in the list (a submodule or a nested repository) is a
// refusal naming it, as an extracted tree refuses a submodule. hyp writes
// no git state: the list is read with ls-files, which takes no optional
// lock, and the bytes are read from disk.
func CopyWorktree(dir, dest string, skip ...string) error {
	out, err := Git(dir, append([]string{"ls-files"}, lsFilesForm...)...)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return fault.EnvFault(fault.Extract, dir, err)
	}
	seen := map[string]bool{}
	for _, rec := range bytes.Split(out, []byte{0}) {
		name := string(rec)
		// a file both tracked and listed again (an unmerged path) is one file
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		first, _, _ := strings.Cut(name, "/")
		if contains(skip, first) {
			continue
		}
		src := filepath.Join(dir, filepath.FromSlash(name))
		info, err := os.Lstat(src)
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			return fault.EnvFault(fault.ReadPath, src, err)
		case info.IsDir():
			return fault.Refuse(fault.NoGuard, fault.Submodule, "worktree", strings.TrimSuffix(name, "/"))
		}
		path, err := entryPath(dest, name)
		if err != nil {
			return err
		}
		if err := ensureDirs(dest, filepath.Dir(path)); err != nil {
			return err
		}
		if err := copyEntry(src, path, info); err != nil {
			return err
		}
	}
	return nil
}

// CopyDir writes the files and symlinks of the folder src under dest, by the
// rules of CopyWorktree: an owner execute bit makes a file 0755, anything
// else 0644, and a symlink stays a symlink. It is the folder as the disk
// tree hash reads it.
func CopyDir(src, dest string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fault.EnvFault(fault.ReadPath, p, err)
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return fault.EnvFault(fault.ReadPath, p, err)
		}
		to := filepath.Join(dest, rel)
		if d.IsDir() {
			if err := ensureDirs(filepath.Dir(dest), to); err != nil {
				return err
			}
			return nil
		}
		info, err := os.Lstat(p)
		if err != nil {
			return fault.EnvFault(fault.ReadPath, p, err)
		}
		return copyEntry(p, to, info)
	})
}

// copyEntry writes the file or symlink src, whose Lstat is info, at the new
// path to.
func copyEntry(src, to string, info fs.FileInfo) error {
	switch m := info.Mode(); {
	case m&fs.ModeSymlink != 0:
		target, err := os.Readlink(src)
		if err != nil {
			return fault.EnvFault(fault.ReadLink, src, err)
		}
		if err := os.Symlink(target, to); err != nil {
			return fault.EnvFault(fault.Extract, src, err)
		}
		return nil
	case m.IsRegular():
		body, err := os.ReadFile(src)
		if err != nil {
			return fault.EnvFault(fault.ReadPath, src, err)
		}
		perm := os.FileMode(0o644)
		if m.Perm()&0o100 != 0 {
			perm = 0o755
		}
		if err := writeNew(to, body, perm); err != nil {
			return fault.EnvFault(fault.Extract, src, err)
		}
		// The umask must not change what git would read.
		if err := os.Chmod(to, perm); err != nil {
			return fault.EnvFault(fault.Extract, src, err)
		}
		return nil
	}
	return fault.Refuse(fault.NoGuard, fault.NotFileOrDir, src)
}
