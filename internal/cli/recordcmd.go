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
// commit, with the files those events name and their subject commits. For
// each store read from the checkout it then makes git ignore the store's
// evidence folder and untracks each file there that no event names,
// leaving it on disk.
func cmdImport(args []string, stdout, stderr io.Writer) error {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		return helpText(importUse + "\n\nadd each store's events the record lacks, with the files they name; PATH is a store folder, read from the checkout or from commit REF; an argument that names a folder on disk is a PATH, also when it holds a colon")
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
	for _, arg := range args {
		ref, p, fromRef := strings.Cut(arg, ":")
		if !fromRef || isDir(arg, cwd) {
			p, ref, fromRef = arg, "", false
		}
		abs := p
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(cwd, p)
		}
		abs = filepath.Clean(abs)
		name, ok := repo.Name(abs)
		if !ok {
			return herr.BadArgs("hyp import %s: %s is outside the repository at %s", arg, abs, repo.Top)
		}
		src, top := abs, repo.Top
		if fromRef {
			tmp, err := os.MkdirTemp("", "hyp-import-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmp)
			if err := repo.ExtractAt(ref, name, tmp); err != nil {
				return err
			}
			src, top = filepath.Join(tmp, filepath.FromSlash(name)), tmp
		}
		n, files, err := store.Import(repo, name, src, top, stderr)
		if err != nil {
			return fmt.Errorf("hyp import %s: %w", arg, err)
		}
		switch {
		case n == 0 && files == 0:
			fmt.Fprintf(stdout, "%s: the record holds every event\n", arg)
		case n == 0:
			fmt.Fprintf(stdout, "%s: the record holds every event; added %d named files it lacked\n", arg, files)
		default:
			fmt.Fprintf(stdout, "%s: imported %d events\n", arg, n)
		}
		if fi, err := os.Stat(record.Dir(abs)); err != nil || !fi.IsDir() {
			continue
		}
		if err := untrackUnnamed(repo, name, abs); err != nil {
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
