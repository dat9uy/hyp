package record

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// A snapshot is the record's tree at one commit, written as files under the
// common git directory so every read of the store reads plain files. A
// commit never changes, so a snapshot of it is written once and kept while
// it is in use. Its files are hard links to a blob cache keyed by object
// id, so a new snapshot costs a link per file and the bytes of new blobs
// only. Nothing a worktree does reaches the common directory: a clean, a
// checkout or a removed worktree leaves every snapshot as it was.

// snapshotMark is the file in a snapshot's git folder that marks the
// folder as a snapshot of the record.
const snapshotMark = "hyp-snapshot"

// IsSnapshot reports whether dir lies in a snapshot of the record: the
// nearest folder at or above dir that holds a git folder is a snapshot's
// root. A read of a snapshot reads every copy of a bundle ref the record
// keeps, each at its own library level.
func IsSnapshot(dir string) bool {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	for d := abs; ; d = filepath.Dir(d) {
		if fi, err := os.Stat(filepath.Join(d, GitName)); err == nil {
			if !fi.IsDir() {
				return false
			}
			_, err := os.Stat(filepath.Join(d, GitName, snapshotMark))
			return err == nil
		}
		if filepath.Dir(d) == d {
			return false
		}
	}
}

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
// folder recipes/<name>@<version> or rubrics/<name>@<version>, or a kept
// copy of a bundle under recipes/.kept or rubrics/.kept, before any .tree
// folder.
func storeOf(p string) (owner, rest string, ok bool) {
	parts := strings.Split(p, "/")
	for i, c := range parts {
		if (c == "recipes" || c == "rubrics") && i+1 < len(parts) && (strings.Contains(parts[i+1], "@") || parts[i+1] == KeptDir) {
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
	key := tip + "-" + hex.EncodeToString(nh[:])
	dir := r.hypDir("snap", key)
	// The read holds its snapshot's lock until the process exits, so a
	// prune by any command never removes a snapshot a read still uses.
	if err := r.holdSnapshot(key); err != nil {
		return "", err
	}
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
	if err := os.WriteFile(filepath.Join(tmp, GitName, snapshotMark), nil, 0o644); err != nil {
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

// prune removes every snapshot but keep that no read used for snapshotKeep
// and that no running read holds: it takes each snapshot's lock without
// waiting, and leaves the snapshot when a read holds it.
func (r *Repo) prune(keep string) {
	snaps, err := os.ReadDir(r.hypDir("snap"))
	if err != nil {
		return
	}
	for _, d := range snaps {
		if d.Name() == keep || strings.HasPrefix(d.Name(), "tmp-") || !d.IsDir() {
			continue
		}
		fi, err := d.Info()
		if err != nil || time.Since(fi.ModTime()) < snapshotKeep {
			continue
		}
		f, err := os.OpenFile(r.snapLock(d.Name()), os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			continue
		}
		r.pruneOpen(f, d.Name())
		f.Close()
	}
}

// pruneOpen removes the snapshot key when it takes, at once, an exclusive
// lock on f, the lock file it opened for key. Another prune may have removed
// that file after f was opened, and a read may then hold a new one: a lock on
// the removed file excludes no read, so pruneOpen skips the snapshot.
func (r *Repo) pruneOpen(f *os.File, key string) {
	if syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil {
		return
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	if !lockedAt(f, r.snapLock(key)) {
		return
	}
	// The lock file goes with its snapshot. A read waiting on it finds the
	// file gone once it holds the lock, and takes a new one.
	if os.RemoveAll(filepath.Join(r.hypDir("snap"), key)) == nil {
		_ = os.Remove(r.snapLock(key))
	}
}

// lockedAt reports whether the open file f is still the file at path p, so a
// lock on f guards what p names.
func lockedAt(f *os.File, p string) bool {
	fi, err1 := f.Stat()
	pi, err2 := os.Stat(p)
	return err1 == nil && err2 == nil && os.SameFile(fi, pi)
}

// snapLock is the lock file of the snapshot key. It lives beside the
// snapshots, and a prune removes it with its snapshot.
func (r *Repo) snapLock(key string) string { return r.hypDir("snaplock", key) }

var (
	heldMu sync.Mutex
	held   = map[string]*os.File{}
)

// holdSnapshot takes a shared lock on the snapshot key's lock file and keeps
// it until the process exits.
func (r *Repo) holdSnapshot(key string) error {
	p := r.snapLock(key)
	heldMu.Lock()
	defer heldMu.Unlock()
	if _, ok := held[p]; ok {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	for {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return err
		}
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
			f.Close()
			return err
		}
		// A prune may have removed the file while this read waited for
		// its lock: the lock then guards nothing, so take it again.
		if lockedAt(f, p) {
			held[p] = f
			return nil
		}
		f.Close()
	}
}
