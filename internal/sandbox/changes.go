package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var copyCap int64 = 10 << 20 // changes/ copies an added or modified file up to this size

// entry is one path of a wt/ manifest: its size, mode, and its sha256 for a
// regular file or its target for a symlink.
type entry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256,omitempty"`
	Mode   string `json:"mode"`
	Link   string `json:"link,omitempty"`
}

// Changes is changes.json: every path in wt/ that changed since create
// finished, against the manifest create wrote then.
type Changes struct {
	Format   int      `json:"format"`
	Baseline string   `json:"baseline"`
	Changes  []Change `json:"changes"`
}

// Change is one changed path. Copied says, for an added or modified regular
// file, whether changes/ holds its new bytes (only files up to copyCap).
type Change struct {
	Path   string `json:"path"`
	Change string `json:"change"` // added, modified or deleted
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256,omitempty"`
	Mode   string `json:"mode"`
	Link   string `json:"link,omitempty"`
	Copied *bool  `json:"copied,omitempty"`
}

// manifest lists every path under wt but wt/.git and index.sqlite* under
// wt/.tree, which older hyp binaries write (ADR 0006), sorted by path.
// Directories are not listed.
// A file that vanishes while it is listed (a process still writing in wt/)
// is left out rather than failing the whole list.
func manifest(wt string) ([]entry, error) {
	var out []entry
	err := filepath.WalkDir(wt, func(p string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) && p != wt {
			return nil
		}
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(wt, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == ".git" && d.IsDir() {
			return fs.SkipDir
		}
		if strings.HasPrefix(rel, ".tree/") && strings.HasPrefix(d.Name(), "index.sqlite") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := os.Lstat(p)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		e := entry{Path: rel, Size: info.Size(), Mode: info.Mode().String()}
		switch {
		case info.Mode()&fs.ModeSymlink != 0:
			if e.Link, err = os.Readlink(p); err != nil {
				return err
			}
		case info.Mode().IsRegular():
			if e.SHA256, err = fileSum(p); errors.Is(err, fs.ErrNotExist) {
				return nil
			} else if err != nil {
				return err
			}
		}
		out = append(out, e)
		return nil
	})
	return out, err
}

func fileSum(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// writeManifest writes <box>/manifest.json, the baseline remove diffs wt/
// against.
func writeManifest(box, wt string) error {
	m, err := manifest(wt)
	if err != nil {
		return err
	}
	return writeJSON(filepath.Join(box, "manifest.json"), m)
}

// diffManifests lists every path whose entry was added, removed or changed
// (size, sha256, mode or link) from before to after, sorted by path.
func diffManifests(before, after []entry) []Change {
	old, cur := map[string]entry{}, map[string]entry{}
	var paths []string
	for _, e := range before {
		old[e.Path] = e
		paths = append(paths, e.Path)
	}
	for _, e := range after {
		cur[e.Path] = e
		if _, ok := old[e.Path]; !ok {
			paths = append(paths, e.Path)
		}
	}
	slices.Sort(paths)
	changes := []Change{}
	for _, p := range paths {
		o, wasThere := old[p]
		n, isThere := cur[p]
		kind := ""
		switch {
		case !wasThere:
			kind = "added"
		case !isThere:
			kind, n = "deleted", o
		case o != n:
			kind = "modified"
		default:
			continue
		}
		changes = append(changes, Change{Path: p, Change: kind, Size: n.Size, SHA256: n.SHA256, Mode: n.Mode, Link: n.Link})
	}
	return changes
}

// writeChanges writes OUT_DIR/changes.json, diffing wt/ against the
// manifest create wrote, and copies each added or modified regular file up
// to copyCap to OUT_DIR/changes/<path>. A sandbox with no manifest (setup did not
// finish, or an older hyp made it) gets a warning and no changes.json.
func writeChanges(box, wt, outDir string, warn io.Writer) error {
	data, err := os.ReadFile(filepath.Join(box, "manifest.json"))
	if errors.Is(err, fs.ErrNotExist) {
		if warn != nil {
			fmt.Fprintln(warn, "sandbox: no manifest (setup did not finish, or an older hyp made the sandbox); changes.json not written")
		}
		return nil
	}
	if err != nil {
		return err
	}
	var before []entry
	if err := json.Unmarshal(data, &before); err != nil {
		return fmt.Errorf("manifest.json: %v", err)
	}
	after, err := manifest(wt)
	if err != nil {
		return err
	}
	ch := Changes{Format: ChangesFormat, Baseline: "manifest.json at create", Changes: diffManifests(before, after)}
	for i := range ch.Changes {
		c := &ch.Changes[i]
		if c.Change == "deleted" || c.SHA256 == "" {
			continue
		}
		copied := c.Size <= copyCap
		c.Copied = &copied
		if !copied {
			continue
		}
		dst := filepath.Join(outDir, keptChanges, filepath.FromSlash(c.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := copyFile(filepath.Join(wt, filepath.FromSlash(c.Path)), dst, 0o644); err != nil {
			return err
		}
	}
	return writeJSON(filepath.Join(outDir, keptChangesJSON), ch)
}
