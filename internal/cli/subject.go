package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/record"
)

// cmdSubject writes hypothesis.subject: what embodies a Hypothesis's claim,
// a base commit, at most one patch over it, and files each named by its
// sha256. The latest tie stands until a Result's run subject holds it; then
// replay refuses a different one.
func cmdSubject(args []string, stdout, stderr io.Writer) error {
	w := newWriter("subject", stdout, stderr)
	base := w.fs.String("base", "", "the commit the subject is built on, resolved in the store root's git repository to its full id")
	patch := w.fs.String("patch", "", "the one patch file applied over --base; the PATHs are then the input files given to the code")
	pos, err := parse(w.fs, args)
	if err != nil {
		return err
	}
	if *patch != "" && *base == "" {
		return usageErr("--patch needs --base: a patch is applied over a base commit")
	}
	if len(pos) < 1 || (len(pos) < 2 && *base == "") {
		return usageErr("usage: hyp subject HYP [PATH...] [--base COMMIT [--patch FILE]] --why W")
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
	var patchFile *event.SubjectFile
	if *patch != "" {
		if info, err := os.Stat(*patch); err != nil || !info.Mode().IsRegular() {
			return usageErr("--patch %s is not a regular file", *patch)
		}
		pf, err := subjectFiles(w.root, []string{*patch})
		if err != nil {
			return err
		}
		patchFile = &pf[0]
	}
	commit := ""
	if *base != "" {
		if commit, err = resolveCommit(w.root, *base); err != nil {
			return err
		}
	}
	return w.run(func() (draft, error) {
		id, err := w.ref(pos[0])
		return draft{event.TypeHypothesisSubject, id, event.HypothesisSubject{Base: commit, Patch: patchFile, Files: files}}, err
	})
}

// resolveCommit names rev as a full commit id of the git repository that
// holds the store root.
func resolveCommit(root, rev string) (string, error) {
	id, err := record.ResolveCommit(root, rev)
	if errors.Is(err, record.ErrNotCommit) || (err == nil && !event.IsCommit(id)) {
		return "", usageErr("--base %s is not a commit in the repository at %s", rev, root)
	}
	if err != nil {
		return "", herr.Env("--base %s: %s", rev, herr.Text(err))
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
// the record names, the one patch applied over it, and the sha256 of every
// input file it says the run was given. A command trace is one run with no
// base and no patch, given its inputs and directory entries (format 6). A
// sandbox trace is one run per Arm, each at its base with that Arm's one
// patch and given the inputs file; with no Arm it is one run at its base
// given the inputs file. An Arm whose patch is not named by a sha256 gives
// no subject, since its code is unknown. A record that cannot be read, or
// names nothing, gives no subject; hyp check reports a record it cannot read. A subject
// equal to an earlier one is kept once (ADR 0020).
func runSubjects(rootOf func(rel string) string, evidence []string) []event.RunSubject {
	var out []event.RunSubject
	add := func(base, patch string, files []string) {
		files = slices.DeleteFunc(files, func(h string) bool { return !event.IsSHA256(h) })
		slices.Sort(files)
		files = slices.Compact(files)
		if len(files) == 0 {
			files = nil
		}
		run := event.RunSubject{Base: base, Patch: patch, Files: files}
		if run.Base == "" && run.Patch == "" && run.Files == nil {
			return
		}
		if slices.ContainsFunc(out, run.Same) {
			return
		}
		out = append(out, run)
	}
	for _, e := range evidence {
		path := filepath.Join(rootOf(e), filepath.FromSlash(filepath.ToSlash(filepath.Clean(e))))
		switch filepath.Base(e) {
		case "command.json":
			if r, err := cmdrecord.Read(path); err == nil {
				add("", "", r.Fingerprints())
			}
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
			base := ""
			if event.IsCommit(r.Base) {
				base = r.Base
			}
			var inputs []string
			if r.Inputs != nil {
				inputs = []string{r.Inputs.SHA256}
			}
			if len(r.Patches) == 0 {
				add(base, "", inputs)
			}
			for _, p := range r.Patches {
				if event.IsSHA256(p.SHA256) {
					add(base, p.SHA256, inputs)
				}
			}
		}
	}
	return out
}
