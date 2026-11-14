package record

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// A snapshot is the record's tree at one commit, written as files under the
// common git directory so every read of the store reads plain files. A
// commit never changes, so a snapshot of it is written once and kept while
// it is in use. Its files are hard links to a blob cache keyed by object
// id, so a new snapshot costs a link per file and the bytes of new blobs
// only. Nothing a worktree does reaches the common directory: a clean, a
// checkout or a removed worktree leaves every snapshot as it was.

// snapshotKeep is how long a snapshot no read made stays on disk.
const snapshotKeep = time.Hour

func (r *Repo) hypDir(parts ...string) string {
	return filepath.Join(append([]string{r.Common, "hyp"}, parts...)...)
}

// keepPath reports whether a record path belongs in the snapshot of store
// name: every path but those under another store's folder. A file of a
// bundle belongs to the bundle, whatever its path holds below it, so a
// known-answer input that holds a store folder stays in every snapshot.
func keepPath(name, p string) bool {
	owner, _, ok := storeOf(p)
	return !ok || owner == name || (name == "" && owner == ".")
}

// storeOf splits a record path under a store's folder into the store's name
// and the slash path below its .tree folder. ok is false for a path under
// no store's folder, and for a file of a bundle, a path that reaches a
// folder recipes/<name>@<version> or rubrics/<name>@<version> before any
// .tree folder.
func storeOf(p string) (owner, rest string, ok bool) {
	parts := strings.Split(p, "/")
	for i, c := range parts {
		if (c == "recipes" || c == "rubrics") && i+1 < len(parts) && strings.Contains(parts[i+1], "@") {
			return "", "", false
		}
		if c != TreeDir {
			continue
		}
		owner = "."
		if i > 0 {
			owner = path.Join(parts[:i]...)
		}
		return owner, strings.Join(parts[i+1:], "/"), true
	}
	return "", "", false
}

// snapshot writes, once, the files of commit tip that a read of store name
// needs, and returns the snapshot's root. The root holds a .git folder,
// so the repository root of every path in it is the snapshot's root.
func (r *Repo) snapshot(tip, name string, entries map[string]entry) (string, error) {
	nh := sha256.Sum256([]byte(name))
	key := tip + "-" + hex.EncodeToString(nh[:4])
	dir := r.hypDir("snap", key)
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		now := time.Now()
		_ = os.Chtimes(dir, now, now)
		return dir, nil
	}
	blobs := r.hypDir("blobs")
	if err := os.MkdirAll(blobs, 0o755); err != nil {
		return "", err
	}
	var missing []string
	need := map[string]bool{}
	for p, e := range entries {
		if e.typ != "blob" || !keepPath(name, p) {
			continue
		}
		if _, err := os.Stat(filepath.Join(blobs, e.sha)); err != nil && !need[e.sha] {
			need[e.sha] = true
			missing = append(missing, e.sha)
		}
	}
	err := r.readBlobs(missing, func(sha string, data []byte) error {
		tmp, err := os.CreateTemp(blobs, "tmp-*")
		if err != nil {
			return err
		}
		if _, err := tmp.Write(data); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return err
		}
		if err := tmp.Sync(); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return err
		}
		if err := tmp.Close(); err != nil {
			os.Remove(tmp.Name())
			return err
		}
		if err := os.Chmod(tmp.Name(), 0o644); err != nil {
			os.Remove(tmp.Name())
			return err
		}
		return os.Rename(tmp.Name(), filepath.Join(blobs, sha))
	})
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(r.hypDir("snap"), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(r.hypDir("snap"), "tmp-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	if err := os.Mkdir(filepath.Join(tmp, GitName), 0o755); err != nil {
		return "", err
	}
	for p, e := range entries {
		if e.typ != "blob" || !keepPath(name, p) {
			continue
		}
		dst := filepath.Join(tmp, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return "", err
		}
		src := filepath.Join(blobs, e.sha)
		if e.mode == "100755" {
			// A link shares its mode, so an executable copies its blob.
			data, err := os.ReadFile(src)
			if err != nil {
				return "", err
			}
			if err := os.WriteFile(dst, data, 0o755); err != nil {
				return "", err
			}
			continue
		}
		if err := os.Link(src, dst); err != nil {
			return "", err
		}
	}
	if err := os.Rename(tmp, dir); err != nil {
		if fi, serr := os.Stat(dir); serr == nil && fi.IsDir() {
			return dir, nil
		}
		return "", err
	}
	r.prune(key)
	return dir, nil
}

// prune removes every snapshot but keep that no read used for snapshotKeep.
func (r *Repo) prune(keep string) {
	snaps, err := os.ReadDir(r.hypDir("snap"))
	if err != nil {
		return
	}
	for _, d := range snaps {
		if d.Name() == keep || strings.HasPrefix(d.Name(), "tmp-") {
			continue
		}
		fi, err := d.Info()
		if err != nil || time.Since(fi.ModTime()) < snapshotKeep {
			continue
		}
		if err := os.RemoveAll(filepath.Join(r.hypDir("snap"), d.Name())); err != nil && !errors.Is(err, os.ErrNotExist) {
			continue
		}
	}
}
