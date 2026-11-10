package recipe

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
)

// NewVersion copies the versioned bundle from (name@N, found from start like
// Load) to target (name@N+1) in the same library directory, sets version:,
// writes from: <from> <sha256> and format: <Format> into the copy's recipe.md
// frontmatter, adds External: TODO under each Evidence and Standalone
// command that has no External: line (addExternalTODO), and returns the new directory and the from: value. The source
// is hashed, never parsed, so a bundle that no longer opens can still be
// copied and fixed; the caller opens the copy.
func NewVersion(start, target, from string) (dir, stamped string, err error) {
	tref, err := parseRef(target)
	if err != nil {
		return "", "", herr.BadArgs("bad recipe ref: %s", herr.Text(err))
	}
	fref, err := parseRef(from)
	if err != nil {
		return "", "", herr.BadArgs("bad recipe ref: %s", herr.Text(err))
	}
	name, tver, fver := tref.Name, tref.Version, fref.Version
	if name != fref.Name {
		return "", "", herr.BadArgs("bad recipe ref: target %s and --from %s name different recipes", target, from)
	}
	tv, terr := strconv.Atoi(tver)
	fv, ferr := strconv.Atoi(fver)
	if terr != nil || ferr != nil || tv < 1 || fv < 1 {
		return "", "", herr.BadArgs("bad recipe ref: versions in %s and %s must be whole numbers from 1", target, from)
	}
	srcPath, srcHash, err := findBundle(start, from)
	if err != nil {
		return "", "", err
	}
	if tv != fv+1 {
		return "", "", herr.Refused("%s is not the next version after %s; the next is %s@%d", target, from, name, fv+1)
	}
	lib := filepath.Dir(srcPath)
	// Every library Load searches counts, so a nearer copy of an old version
	// cannot branch past a newer one further up.
	highest := 0
	for _, d := range DirsOf(start, filepath.Base(lib)) {
		matches, err := filepath.Glob(filepath.Join(d, name+"@*"))
		if err != nil {
			return "", "", err
		}
		for _, m := range matches {
			r, err := event.ParseRef(filepath.Base(m))
			if err != nil || r.Name != name {
				continue
			}
			if k, err := strconv.Atoi(r.Version); err == nil && k > highest {
				highest = k
			}
		}
	}
	if highest > fv {
		return "", "", herr.Refused("%s@%d exists; the next version is %s@%d", name, highest, name, highest+1)
	}
	if r, err := LoadAny(start, target); err == nil {
		return "", "", herr.Refused("%s exists at %s", target, r.Path)
	}
	if _, err := os.Lstat(filepath.Join(lib, target)); err == nil {
		return "", "", herr.Refused("%s exists", target)
	}

	tmp, err := os.MkdirTemp(lib, "."+target+".new-")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tmp)
	if err := copyBundle(srcPath, tmp); err != nil {
		return "", "", err
	}
	md := filepath.Join(tmp, "recipe.md")
	info, err := os.Stat(md)
	if err != nil {
		return "", "", err
	}
	data, err := os.ReadFile(md)
	if err != nil {
		return "", "", err
	}
	stamped = from + " " + srcHash
	out, err := stamp(data, strconv.Itoa(tv), stamped)
	if err != nil {
		return "", "", err
	}
	if out, err = addExternalTODO(out); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(md, out, info.Mode().Perm()); err != nil {
		return "", "", err
	}
	dir = filepath.Join(lib, target)
	if err := os.Rename(tmp, dir); err != nil {
		return "", "", err
	}
	return dir, stamped, nil
}

// findBundle finds the versioned bundle ref from start, like Load but
// hashing each copy without parsing it, and returns the nearest copy and its
// hash. Every copy must hash the same.
func findBundle(start, ref string) (path, hash string, err error) {
	var dirs []string
	for _, l := range []string{Recipes, Rubrics} {
		dirs = append(dirs, DirsOf(start, l)...)
	}
	for _, d := range dirs {
		dir := filepath.Join(d, ref)
		if !isFile(filepath.Join(dir, "recipe.md")) {
			continue
		}
		_, _, h, err := HashOnly(dir)
		if err != nil {
			return "", "", herr.BadArgs("bad recipe ref: %s", herr.Text(err))
		}
		if path == "" {
			path, hash = dir, h
		} else if h != hash {
			return "", "", herr.BadArgs("bad recipe ref: %s", differ(ref, path, dir))
		}
	}
	if path != "" {
		return path, hash, nil
	}
	return "", "", herr.BadArgs("bad recipe ref: bundle %s not found; searched %s", ref, strings.Join(dirs, ", "))
}

// copyBundle copies every file of the bundle src that the hash covers into
// dst, keeping permission bits. Anything other than a directory or regular
// file is refused, as the hash refuses it.
func copyBundle(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if rel != "." && Unhashed(filepath.ToSlash(rel)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("%s is not a regular file", p)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.WriteFile(out, data, info.Mode().Perm()); err != nil {
			return err
		}
		// WriteFile is subject to the umask.
		return os.Chmod(out, info.Mode().Perm())
	})
}

// stamp sets version: in the frontmatter of a recipe.md and writes from: and
// format: right after it, dropping any from: or format: line already there. Everything outside
// the frontmatter is unchanged.
func stamp(data []byte, version, from string) ([]byte, error) {
	lines := strings.SplitAfter(string(data), "\n")
	if strings.TrimSpace(lines[0]) != "---" {
		return nil, errors.New("recipe.md has no frontmatter")
	}
	out := []string{lines[0]}
	stamped := false
	i := 1
	for ; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if t == "---" {
			break
		}
		switch {
		case strings.HasPrefix(t, "from:"), strings.HasPrefix(t, "format:"):
		case strings.HasPrefix(t, "version:"):
			out = append(out, "version: "+version+"\n", "from: "+from+"\n", "format: "+Format+"\n")
			stamped = true
		default:
			out = append(out, lines[i])
		}
	}
	if !stamped {
		return nil, errors.New("recipe.md frontmatter has no version: line")
	}
	out = append(out, lines[i:]...)
	return []byte(strings.Join(out, "")), nil
}

// addExternalTODO adds a line "External: TODO", indented as the command, under
// each Run line in the Evidence and Standalone sections of a recipe.md that
// has no External: line directly under it, so hyp bundle check names each
// command whose reach is still to be stated. The rest is unchanged.
func addExternalTODO(data []byte) ([]byte, error) {
	lines, err := splitLines(data)
	if err != nil {
		return nil, err
	}
	add := map[int]bool{}
	_ = eachLine(lines, func(i int, section Step) error {
		if (section == Evidence || section == Standalone) && isRun(strings.TrimSpace(lines[i])) && !externalAt(lines, i).Line {
			add[i] = true
		}
		return nil
	})
	if len(add) == 0 {
		return data, nil
	}
	var b strings.Builder
	raw := strings.SplitAfter(string(data), "\n")
	for i, l := range raw {
		b.WriteString(l)
		if !add[i] {
			continue
		}
		if !strings.HasSuffix(l, "\n") {
			b.WriteString("\n")
		}
		indent := l[:len(l)-len(strings.TrimLeft(l, " \t"))]
		b.WriteString(indent + "External: " + ExternalTODO + "\n")
	}
	return []byte(b.String()), nil
}
