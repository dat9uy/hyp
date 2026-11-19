package logstore

import (
	"errors"
	"github.com/dat9uy/hyp/internal/fault"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Gone is the trace folder of a clone that no longer exists: the git common
// directory its file clone names is not on the disk.
type Gone struct {
	Key   string // the folder's name under the traces folder
	Clone string // the git common directory the folder's file clone names
}

// GoneTraces lists the trace folders under the traces folder whose clone no
// longer exists, sorted by clone. A folder whose file clone is missing,
// empty or not an absolute path is never listed: hyp cannot tell whose it is.
// A clone whose path cannot be checked for another cause than that it does
// not exist counts as live. So a live clone's trace is never listed.
func GoneTraces() (string, []Gone, error) {
	base, err := TracesDir()
	if err != nil {
		return "", nil, err
	}
	entries, err := os.ReadDir(base)
	if errors.Is(err, fs.ErrNotExist) {
		return base, nil, nil
	}
	if err != nil {
		return "", nil, fault.EnvFault(fault.ReadTraces, base, err)
	}
	var gone []Gone
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if clone, ok := goneClone(filepath.Join(base, e.Name())); ok {
			gone = append(gone, Gone{Key: e.Name(), Clone: clone})
		}
	}
	slices.SortFunc(gone, func(a, b Gone) int { return strings.Compare(a.Clone, b.Clone) })
	return base, gone, nil
}

// goneClone reads the file clone of the trace folder dir, and reports whether
// the clone it names no longer exists.
func goneClone(dir string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "clone"))
	if err != nil {
		return "", false
	}
	clone := strings.TrimSpace(string(data))
	if !filepath.IsAbs(clone) {
		return "", false
	}
	_, err = os.Lstat(clone)
	return clone, errors.Is(err, fs.ErrNotExist)
}

// RemoveGone removes the trace folder of each clone in gone that still does
// not exist, under the folder's own lock, and returns those it removed. It
// checks each clone again under the lock, so a clone that came back since it
// was listed keeps its trace.
func RemoveGone(base string, gone []Gone) ([]Gone, error) {
	var removed []Gone
	for _, g := range gone {
		dir := filepath.Join(base, g.Key)
		ok, err := removeIfGone(dir, g.Clone)
		if err != nil {
			return removed, err
		}
		if ok {
			removed = append(removed, g)
		}
	}
	return removed, nil
}

func removeIfGone(dir, clone string) (bool, error) {
	if _, err := os.Stat(eventsIn(dir)); err == nil {
		unlock, err := lockDir(dir)
		if err != nil {
			return false, err
		}
		defer unlock()
	}
	if now, ok := goneClone(dir); !ok || now != clone {
		return false, nil
	}
	if err := os.RemoveAll(dir); err != nil {
		return false, fault.EnvFault(fault.RemoveTrace, dir, err)
	}
	return true, nil
}
