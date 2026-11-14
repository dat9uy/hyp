package recipe

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
)

// Placeholder marks every line of the first-version template that the
// author must replace. hyp add exp refuses a bundle that still holds it
// (HasPlaceholder), so a generated bundle is never pinned unedited.
const Placeholder = "hyp:placeholder"

// nameToken is the recipe name in the template's files.
const nameToken = "__NAME__"

// template is the bundle hyp bundle new NAME@1 writes: a layout with a
// failing stub check and a must-pass and a must-fail case, not a method
// (docs/adr/0009-a-first-instrument-starts-from-the-template.md).
//
//go:embed template
var template embed.FS

// rubricTemplate is the bundle hyp bundle new NAME@1 --rubric writes into
// rubrics/: one stub Standalone rule and a supports and a contradicts case
// (ADR 0018).
//
//go:embed rubric-template
var rubricTemplate embed.FS

// NewFirst writes the first version target (NAME@1) of a new recipe from
// the embedded template into root/recipes/, root being the store's
// directory (the one holding its store folder), and returns its directory. A version
// other than 1 is a bad argument, since a next version is copied with
// --from; any NAME@N in a library the lookup searches from root is refused.
func NewFirst(root, target string) (string, error) { return newFirstIn(root, Recipes, target) }

// NewFirstRubric is NewFirst for a rubric: it writes the rubric template
// into root/rubrics/.
func NewFirstRubric(root, target string) (string, error) { return newFirstIn(root, Rubrics, target) }

// newFirstIn writes the template of library libName. A name in use in
// either library is refused, so a name lives in one library only.
func newFirstIn(root, libName, target string) (string, error) {
	tref, err := FirstRef(target)
	if err != nil {
		return "", err
	}
	for _, l := range []string{Recipes, Rubrics} {
		for _, d := range DirsOf(root, l) {
			matches, err := filepath.Glob(filepath.Join(d, tref.Name+"@*"))
			if err != nil {
				return "", err
			}
			for _, m := range matches {
				if r, err := event.ParseRef(filepath.Base(m)); err == nil && r.Name == tref.Name {
					return "", herr.Refused("%s %s exists at %s; make the next version with hyp bundle new %s@<N+1> --from %s", kindWord(l), tref.Name, m, tref.Name, filepath.Base(m))
				}
			}
		}
	}
	src, srcDir := template, "template"
	if libName == Rubrics {
		src, srcDir = rubricTemplate, "rubric-template"
	}
	lib := filepath.Join(root, libName)
	if err := os.MkdirAll(lib, 0o755); err != nil {
		return "", herr.Env("%v", err)
	}
	tmp, err := os.MkdirTemp(lib, "."+target+".new-")
	if err != nil {
		return "", herr.Env("%v", err)
	}
	defer os.RemoveAll(tmp)
	if err := writeTemplate(src, srcDir, tmp, tref.Name); err != nil {
		return "", err
	}
	dir := filepath.Join(lib, target)
	if err := os.Rename(tmp, dir); err != nil {
		return "", herr.Env("%v", err)
	}
	return dir, nil
}

// FirstRef parses target as the ref of a first version, NAME@1. A bad ref,
// or any other version, which is copied with --from, is a bad argument.
func FirstRef(target string) (event.Ref, error) {
	tref, err := parseRef(target)
	if err != nil {
		return event.Ref{}, herr.BadArgs("bad recipe ref: %s", herr.Text(err))
	}
	if tref.Version != "1" {
		return event.Ref{}, herr.BadArgs("missing --from: %s is not a first version; a first version is %s@1, and a next one needs --from %s@N", target, tref.Name, tref.Name)
	}
	return tref, nil
}

// writeTemplate writes the embedded template src, rooted at srcDir, into
// dst with name in place of nameToken. Embedding drops file modes, so a .sh
// file gets 0755 and every other file 0644.
func writeTemplate(src embed.FS, srcDir, dst, name string) error {
	return fs.WalkDir(src, srcDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, srcDir), "/")
		out := filepath.Join(dst, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		data, err := src.ReadFile(p)
		if err != nil {
			return err
		}
		data = bytes.ReplaceAll(data, []byte(nameToken), []byte(name))
		mode := os.FileMode(0o644)
		if path.Ext(p) == ".sh" {
			mode = 0o755
		}
		if err := os.WriteFile(out, data, mode); err != nil {
			return err
		}
		// WriteFile is subject to the umask.
		return os.Chmod(out, mode)
	})
}

// HasPlaceholder lists the bundle-relative files of the bundle in dir that
// hold Placeholder, sorted. The inputs of known answers are skipped, since
// an input may be a store that holds the marker on purpose, and so are the
// entries the hash skips.
func HasPlaceholder(dir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if Unhashed(rel) || inInput(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if bytes.Contains(data, []byte(Placeholder)) {
			out = append(out, rel)
		}
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return out, nil
}
