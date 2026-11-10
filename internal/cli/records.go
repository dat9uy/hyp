package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// recordFiles reads the --records arguments of hyp add exp: each path,
// relative to the store root like an --evidence pointer, names a file or a
// directory under .tree/evidence/, and a directory stands for each regular
// file under it. It returns each file once, by its slash path under the root
// and its sha256, in the order given (ADR 0018). The new Experiment has no
// run yet, so every file found is an earlier record.
func recordFiles(root string, paths []string) ([]event.SubjectFile, error) {
	var out []event.SubjectFile
	add := func(rel string) error {
		rel = filepath.ToSlash(rel)
		if slices.ContainsFunc(out, func(f event.SubjectFile) bool { return f.Path == rel }) {
			return nil
		}
		sum, err := fileSHA256(filepath.Join(root, rel))
		if err != nil {
			return usageErr("--records %s: %v", rel, err)
		}
		out = append(out, event.SubjectFile{Path: rel, SHA256: sum})
		return nil
	}
	for _, p := range paths {
		rel := p
		if filepath.IsAbs(p) {
			r, err := filepath.Rel(root, p)
			if err != nil {
				return nil, usageErr("--records %s is outside the store root %s", p, root)
			}
			rel = r
		}
		rel = filepath.ToSlash(filepath.Clean(rel))
		if !strings.HasPrefix(rel+"/", replay.EvidenceDir) || rel+"/" == replay.EvidenceDir {
			return nil, usageErr("--records %s is not under %s; a record is a file an earlier run kept there", p, replay.EvidenceDir)
		}
		abs := filepath.Join(root, filepath.FromSlash(rel))
		fi, err := os.Lstat(abs)
		if err != nil {
			return nil, usageErr("--records %s not found", p)
		}
		if !fi.IsDir() {
			if !fi.Mode().IsRegular() {
				return nil, usageErr("--records %s is not a regular file", p)
			}
			if err := add(rel); err != nil {
				return nil, err
			}
			continue
		}
		err = filepath.WalkDir(abs, func(q string, d fs.DirEntry, err error) error {
			if err != nil || !d.Type().IsRegular() {
				return err
			}
			r, err := filepath.Rel(root, q)
			if err != nil {
				return err
			}
			return add(r)
		})
		if err != nil {
			return nil, err
		}
	}
	if len(out) > event.MaxRecords {
		return nil, usageErr("--records names %d files, over the limit of %d", len(out), event.MaxRecords)
	}
	return out, nil
}

// changedRecords lists the pinned records of an Experiment that no longer
// hash as pinned, a missing file included.
func changedRecords(root string, records []event.SubjectFile) []string {
	var out []string
	for _, f := range records {
		sum, err := fileSHA256(filepath.Join(root, filepath.FromSlash(f.Path)))
		if err != nil || sum != f.SHA256 {
			out = append(out, f.Path)
		}
	}
	return out
}
