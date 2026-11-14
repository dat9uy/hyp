package record

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
)

// KeptDir is the folder, inside a library folder recipes/ or rubrics/, where
// the record keeps each other copy of a bundle whose path it holds with
// other files: <lib>/.kept/<key>/<name>@<version>/, where key is the
// sha256 of the copy's file list. Two clones can each pin one bundle ref
// with their own bytes; the record keeps both, so each pin finds its copy.
// The bundle hash skips dot entries, and a lookup by ref reads only
// <lib>/<name>@<version>/, so a kept copy changes no other bundle.
const KeptDir = ".kept"

// bundleOf splits record path p into the folder of the bundle it is a file
// of and its path below that folder. ok is false for a path under a store's
// folder and for a file of a kept copy.
func bundleOf(p string) (dir, rel string, ok bool) {
	parts := strings.Split(p, "/")
	for i, c := range parts {
		if c == TreeDir {
			return "", "", false
		}
		if (c == "recipes" || c == "rubrics") && i+2 < len(parts) && strings.Contains(parts[i+1], "@") {
			return strings.Join(parts[:i+2], "/"), strings.Join(parts[i+2:], "/"), true
		}
	}
	return "", "", false
}

// bundles groups the files of each bundle the entries hold at the bundle's
// own path: bundle folder → path below it → entry.
func bundles(entries map[string]entry) map[string]map[string]entry {
	out := map[string]map[string]entry{}
	for p, e := range entries {
		if e.typ != "blob" {
			continue
		}
		dir, rel, ok := bundleOf(p)
		if !ok {
			continue
		}
		if out[dir] == nil {
			out[dir] = map[string]entry{}
		}
		out[dir][rel] = e
	}
	return out
}

// listingKey is the sha256 of a bundle's file list, each file's mode,
// object id and path, in path order: one key for each content of a bundle.
func listingKey(files map[string]entry) string {
	h := sha256.New()
	for _, rel := range slices.Sorted(maps.Keys(files)) {
		e := files[rel]
		fmt.Fprintf(h, "%s %s\t%s\n", e.mode, e.sha, rel)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// sameListing reports whether two file lists of a bundle are equal.
func sameListing(a, b map[string]entry) bool {
	return maps.EqualFunc(a, b, func(x, y entry) bool { return x.mode == y.mode && x.sha == y.sha })
}

// keptBundleDir is the record path of the kept copy, with the file list
// key, of the bundle at record path dir.
func keptBundleDir(dir, key string) string {
	return path.Join(path.Dir(dir), KeptDir, key, path.Base(dir))
}
