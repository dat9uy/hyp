package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/record"
	"github.com/dat9uy/hyp/internal/store"
)

const syncUse = "usage: hyp sync [REMOTE]"

// cmdSync joins the record with REMOTE's (origin by default) and pushes
// the join (record.Sync).
func cmdSync(args []string, stdout, stderr io.Writer) error {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		return helpText(syncUse + "\n\nfetch REMOTE's record, join the two logs as a set by event id, and push the join; REMOTE is origin by default")
	}
	if len(args) > 1 || (len(args) == 1 && strings.HasPrefix(args[0], "-")) {
		return usageErr(syncUse)
	}
	remote := "origin"
	if len(args) == 1 {
		remote = args[0]
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return record.Sync(cwd, remote, stdout, stderr)
}

const importUse = "usage: hyp import [REF:]PATH..."

// cmdImport adds to the record every event of each named work-tree store
// that the record lacks, read from the checkout or, with REF:, from that
// commit, with the files those events name and their subject commits, in
// one write that adds each event whole or refuses and adds nothing. For
// each store read from the checkout it then makes git ignore the store's
// evidence folder and untracks each file there that no event names,
// leaving it on disk.
func cmdImport(args []string, stdout, stderr io.Writer) error {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		return helpText(importUse + "\n\nadd each store's events the record lacks, with the files they name; PATH is a store folder, read from the checkout or from commit REF; an argument that names a folder on disk is a PATH; each event is added with all it names, or the import exits 1 and adds nothing")
	}
	if len(args) == 0 || slices.ContainsFunc(args, func(a string) bool { return strings.HasPrefix(a, "-") }) {
		return usageErr(importUse)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repo, err := record.OpenRecord(cwd)
	if err != nil {
		return err
	}
	// Every argument is read before anything is written, so a refusal of
	// one adds nothing of any.
	srcs := make([]store.Source, len(args))
	work := make([]string, len(args))
	for i, arg := range args {
		ref, p, fromRef := strings.Cut(arg, ":")
		if !fromRef || isDir(arg, cwd) {
			p, ref, fromRef = arg, "", false
		}
		abs := p
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(cwd, p)
		}
		abs = filepath.Clean(abs)
		name, err := repo.Name(abs)
		if err != nil {
			return fmt.Errorf("hyp import %s: %w", arg, err)
		}
		srcs[i] = store.Source{Name: name, Src: abs, Top: repo.Top}
		if fromRef {
			tmp, err := os.MkdirTemp("", "hyp-import-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmp)
			if err := repo.ExtractAt(ref, name, tmp); err != nil {
				return err
			}
			srcs[i].Src, srcs[i].Top = filepath.Join(tmp, filepath.FromSlash(name)), tmp
		} else {
			work[i] = abs
		}
	}
	for i, a := range args {
		for j := range i {
			if srcs[j].Name == srcs[i].Name {
				return herr.BadArgs("hyp import: %s and %s name one store, %s", args[j], a, srcs[i].Name)
			}
		}
	}
	counts, err := store.Import(repo, srcs, stderr)
	if err != nil {
		return fmt.Errorf("hyp import: %w", err)
	}
	for i, arg := range args {
		if counts[i] == 0 {
			fmt.Fprintf(stdout, "%s: the record holds every event\n", arg)
		} else {
			fmt.Fprintf(stdout, "%s: imported %d events\n", arg, counts[i])
		}
		if work[i] == "" {
			continue
		}
		if fi, err := os.Stat(record.Dir(work[i])); err != nil || !fi.IsDir() {
			continue
		}
		if err := untrackUnnamed(repo, srcs[i].Name, work[i]); err != nil {
			return fmt.Errorf("hyp import %s: %w", arg, err)
		}
	}
	return nil
}

// isDir reports whether arg, read against cwd when it is relative, names a
// folder on disk: such an argument is a PATH, even when it holds a colon.
func isDir(arg, cwd string) bool {
	if !filepath.IsAbs(arg) {
		arg = filepath.Join(cwd, arg)
	}
	fi, err := os.Stat(arg)
	return err == nil && fi.IsDir()
}

// untrackUnnamed makes git ignore the evidence folder of the store at dir
// and untracks each file there that no event of the record's store names.
func untrackUnnamed(repo *record.Repo, name, dir string) error {
	if err := record.IgnoreEvidence(dir); err != nil {
		return err
	}
	s, err := repo.StoreNamed(name)
	if err != nil {
		return err
	}
	if !s.InRecord {
		return nil
	}
	root, err := s.Root()
	if err != nil {
		return err
	}
	named, err := store.NamedFiles(root)
	if err != nil {
		return err
	}
	tracked, err := repo.TrackedEvidence(dir)
	if err != nil {
		return err
	}
	var drop []string
	for _, f := range tracked {
		if !named[f] {
			drop = append(drop, f)
		}
	}
	return repo.Untrack(dir, drop)
}
