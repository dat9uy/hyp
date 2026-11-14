package record

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
)

// storeName is the name of the store whose folder is dir: its slash path
// from the repository root top, with symbolic links resolved, so a folder
// reached through a link inside the repository is the store it links to.
// Each name is spelled as its folder spells it, so on a filesystem that
// folds letter case or Unicode normalization a folder typed another way is
// the same store. ok is false when dir, links resolved, is outside top: no
// store of the repository is there.
func storeName(top, dir string) (string, bool) {
	rt, err1 := filepath.EvalSymlinks(top)
	rd, err2 := filepath.EvalSymlinks(dir)
	if err1 != nil || err2 != nil {
		// A folder that does not exist yet has no links to resolve.
		return relInside(top, dir)
	}
	return relInside(realCase(rt), realCase(rd))
}

// relInside is the slash path of dir from top, ok false when dir is not at
// or below top.
func relInside(top, dir string) (string, bool) {
	rel, err := filepath.Rel(top, dir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// outside is the refusal of a store folder that, links resolved, is outside
// the repository at top.
func outside(dir, top string) error {
	real, err := filepath.EvalSymlinks(dir)
	if err != nil {
		real = dir
	}
	return herr.Refused("the store folder %s is %s, outside the repository at %s; a repository's stores are below its root", dir, real, top)
}

// CheckPlain refuses a store folder dir of the repository at top whose
// name is not plain: a part with a character other than ASCII letters,
// digits, '.', '_', '-' and '@', a part that is git's folder name, "." or
// "..", or a part that differs only in letter case from another name its
// folder lists. It refuses a folder outside the repository too. The root's
// own store, ".", is plain. hyp init and hyp import call it before they
// write anything.
func CheckPlain(top, dir string) error {
	name, ok := storeName(top, dir)
	if !ok {
		return outside(dir, top)
	}
	if name == "." {
		return nil
	}
	rt, err := filepath.EvalSymlinks(top)
	if err != nil {
		rt = top
	}
	cur := realCase(rt)
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." || IsGitName(part) {
			return herr.Refused("the store folder %q has the part %q; a store folder's name has no part %s, . or ..", name, part, GitName)
		}
		for _, c := range part {
			if !plainRune(c) {
				return herr.Refused("the store folder %q holds %q; a store folder's name holds only ASCII letters, digits, '.', '_', '-', '@' and '/'", name, string(c))
			}
		}
		if ents, err := os.ReadDir(cur); err == nil {
			for _, e := range ents {
				if e.Name() != part && strings.EqualFold(e.Name(), part) {
					return herr.Refused("the store folder %q has the part %q, and its folder also holds %q; each part of a store folder's name differs from every other name in its folder by more than letter case", name, part, e.Name())
				}
			}
		}
		cur = filepath.Join(cur, part)
	}
	return nil
}

func plainRune(c rune) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '.' || c == '_' || c == '-' || c == '@'
}
