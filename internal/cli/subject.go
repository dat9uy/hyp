package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/event"
)

// cmdSubject writes hypothesis.subject: what embodies a Hypothesis's claim,
// a base commit, files each named by its sha256, or both. The latest tie
// stands until a Result's run subject holds it; then replay refuses a
// different one.
func cmdSubject(args []string, stdout, stderr io.Writer) error {
	w := newWriter("subject", stdout, stderr)
	base := w.fs.String("base", "", "the commit the subject is built on, resolved in the store root's git repository to its full id")
	pos, err := parse(w.fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 || (len(pos) < 2 && *base == "") {
		return usageErr("usage: hyp subject HYP [PATH...] [--base COMMIT] --why W")
	}
	if _, err := w.begin(args, len(pos)); err != nil {
		return err
	}
	var files []event.SubjectFile
	if len(pos) > 1 {
		if files, err = subjectFiles(w.root, pos[1:]); err != nil {
			return err
		}
	}
	commit := ""
	if *base != "" {
		if commit, err = resolveCommit(w.root, *base); err != nil {
			return err
		}
	}
	return w.run(func() (draft, error) {
		id, err := w.ref(pos[0])
		return draft{event.TypeHypothesisSubject, id, event.HypothesisSubject{Base: commit, Files: files}}, err
	})
}

// resolveCommit names rev as a full commit id of the git repository that
// holds the store root.
func resolveCommit(root, rev string) (string, error) {
	out, err := exec.Command("git", "-C", root, "rev-parse", "--verify", "--quiet", rev+"^{commit}").Output()
	id := strings.TrimSpace(string(out))
	if err != nil || !event.IsCommit(id) {
		return "", usageErr("--base %s is not a commit in the repository at %s", rev, root)
	}
	return id, nil
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

// runSubjects keeps one subject per cited run, in evidence order: from each
// cited command.json and sandbox.json under the store root, the base commit
// the record names and the sha256 of every file it says the run was given
// over it. A command record names no base and gives its inputs and directory
// entries (format 6); a sandbox record gives its base, its patches and its
// inputs file. A record that cannot be read, or names nothing, gives no
// subject; hyp check reports a record it cannot read. A subject equal to an
// earlier one is kept once (ADR 0020).
func runSubjects(root string, evidence []string) []event.RunSubject {
	var out []event.RunSubject
	for _, e := range evidence {
		path := filepath.Join(root, filepath.FromSlash(filepath.ToSlash(filepath.Clean(e))))
		var run event.RunSubject
		switch filepath.Base(e) {
		case "command.json":
			r, err := cmdrecord.Read(path)
			if err != nil {
				continue
			}
			run.Files = r.Fingerprints()
		case "sandbox.json":
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var r struct {
				Base    string                    `json:"base"`
				Inputs  *struct{ SHA256 string }  `json:"inputs"`
				Patches []struct{ SHA256 string } `json:"patches"`
			}
			if json.Unmarshal(data, &r) != nil {
				continue
			}
			if event.IsCommit(r.Base) {
				run.Base = r.Base
			}
			if r.Inputs != nil {
				run.Files = append(run.Files, r.Inputs.SHA256)
			}
			for _, p := range r.Patches {
				run.Files = append(run.Files, p.SHA256)
			}
		default:
			continue
		}
		run.Files = slices.DeleteFunc(run.Files, func(h string) bool { return !event.IsSHA256(h) })
		slices.Sort(run.Files)
		run.Files = slices.Compact(run.Files)
		if len(run.Files) == 0 {
			run.Files = nil
		}
		if run.Base == "" && run.Files == nil {
			continue
		}
		if slices.ContainsFunc(out, func(o event.RunSubject) bool { return o.Base == run.Base && slices.Equal(o.Files, run.Files) }) {
			continue
		}
		out = append(out, run)
	}
	return out
}
