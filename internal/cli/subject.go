package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/event"
)

// cmdSubject writes hypothesis.subject: the files that embody a
// Hypothesis's claim, each named by its sha256. The latest tie stands until
// a Result's fingerprints hold it; then replay refuses a different one.
func cmdSubject(args []string, stdout, stderr io.Writer) error {
	w := newWriter("subject", stdout, stderr)
	pos, err := parse(w.fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 2 {
		return usageErr("usage: hyp subject HYP PATH... --why W")
	}
	if _, err := w.begin(args, len(pos)); err != nil {
		return err
	}
	files, err := subjectFiles(w.root, pos[1:])
	if err != nil {
		return err
	}
	return w.run(func() (draft, error) {
		id, err := w.ref(pos[0])
		return draft{event.TypeHypothesisSubject, id, event.HypothesisSubject{Files: files}}, err
	})
}

// subjectFiles hashes each path, read against the working directory: a file
// is one entry, and a directory is each regular file under it. Each path is
// stored relative to the store root, as a pointer for readers.
func subjectFiles(root string, paths []string) ([]event.SubjectFile, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var out []event.SubjectFile
	add := func(abs string) error {
		rel, err := filepath.Rel(root, abs)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return usageErr("subject %s is outside the store root %s", abs, root)
		}
		sum, err := fileSHA256(abs)
		if err != nil {
			return usageErr("subject %s: %v", abs, err)
		}
		out = append(out, event.SubjectFile{Path: filepath.ToSlash(rel), SHA256: sum})
		return nil
	}
	for _, p := range paths {
		abs := p
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(cwd, p)
		}
		fi, err := os.Stat(abs)
		if err != nil {
			return nil, usageErr("subject %s not found", p)
		}
		if !fi.IsDir() {
			if err := add(abs); err != nil {
				return nil, err
			}
			continue
		}
		err = filepath.WalkDir(abs, func(q string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type().IsRegular() {
				return add(q)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if len(out) > event.MaxSubjectFiles {
		return nil, usageErr("subject has %d files (max %d)", len(out), event.MaxSubjectFiles)
	}
	slices.SortFunc(out, func(a, b event.SubjectFile) int { return strings.Compare(a.Path, b.Path) })
	return out, nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
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

// fingerprints copies, from each cited command.json and sandbox.json under
// the store root, the sha256 of every file the record says the run was given:
// a command record's inputs and directory entries (format 6), and a sandbox
// record's patches and inputs file. It returns them sorted and unique. A
// record that cannot be read gives none; hyp check reports it.
func fingerprints(root string, evidence []string) []string {
	var out []string
	for _, e := range evidence {
		path := filepath.Join(root, filepath.FromSlash(filepath.ToSlash(filepath.Clean(e))))
		switch filepath.Base(e) {
		case "command.json":
			if r, err := cmdrecord.Read(path); err == nil {
				out = append(out, r.Fingerprints()...)
			}
		case "sandbox.json":
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var r struct {
				Inputs  *struct{ SHA256 string }  `json:"inputs"`
				Patches []struct{ SHA256 string } `json:"patches"`
			}
			if json.Unmarshal(data, &r) != nil {
				continue
			}
			if r.Inputs != nil {
				out = append(out, r.Inputs.SHA256)
			}
			for _, p := range r.Patches {
				out = append(out, p.SHA256)
			}
		}
	}
	out = slices.DeleteFunc(out, func(h string) bool { return !event.IsSHA256(h) })
	slices.Sort(out)
	return slices.Compact(out)
}
