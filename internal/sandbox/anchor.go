package sandbox

import (
	"path/filepath"
)

// anchor writes p relative to root, as name/REL, or as name for root itself,
// so a record names a path inside root without the host's prefix. It reports
// false when p is not root or under it. Both are compared as written and,
// when that fails, symlink-resolved, so a root reached through a symlink
// (macOS /var is /private/var) anchors a path written either way.
func anchor(root, name, p string) (string, bool) {
	r, err1 := filepath.Abs(root)
	q, err2 := filepath.Abs(p)
	if err1 != nil || err2 != nil {
		return "", false
	}
	if a, ok := anchorRel(r, name, q); ok {
		return a, true
	}
	rr, err1 := resolve(r)
	rq, err2 := resolve(q)
	if err1 != nil || err2 != nil {
		return "", false
	}
	return anchorRel(rr, name, rq)
}

func anchorRel(root, name, p string) (string, bool) {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return "", false
	}
	if rel == "." {
		return name, true
	}
	if !filepath.IsLocal(rel) {
		return "", false
	}
	return name + "/" + filepath.ToSlash(rel), true
}

// plainName reports whether s is one path element: not empty, not . or ..,
// and with no separator, so joining it to a directory stays in that directory.
func plainName(s string) bool {
	return s != "" && s != "." && s != ".." && filepath.Base(s) == s && filepath.IsLocal(s)
}
