package store

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/record"
)

// Names is what one event names outside its own line: files by slash path
// relative to the store's folder, the bundles it pins by ref and kind, and
// the commits its subject names. Sums holds the sha256 the event stores for
// each of its files that has one.
type Names struct {
	Files   []string
	Sums    map[string]string
	Bundles []Pin
	Commits []string
}

func (n *Names) file(rel, sum string) {
	n.Files = append(n.Files, rel)
	if sum != "" {
		if n.Sums == nil {
			n.Sums = map[string]string{}
		}
		n.Sums[rel] = sum
	}
}

// Pin is one bundle an event pins.
type Pin struct {
	Ref, Hash string
	Rubric    bool
}

// NamesOf lists what ev names. An evidence URL names no file; a pointer
// into a versioned recipe names that recipe's bundle.
func NamesOf(ev event.Event) Names {
	var n Names
	switch ev.Type {
	case event.TypeNodeAdd:
		var p event.NodeAdd
		if json.Unmarshal(ev.Payload, &p) != nil {
			return n
		}
		if p.Recipe != "" && p.Recipe != event.RecipeNone {
			n.Bundles = append(n.Bundles, Pin{Ref: p.Recipe, Hash: p.RecipeHash})
		}
		if p.Rubric != "" && p.Rubric != event.RubricReading {
			n.Bundles = append(n.Bundles, Pin{Ref: p.Rubric, Hash: p.RubricHash, Rubric: true})
		}
		for _, f := range p.Records {
			n.file(f.Path, f.SHA256)
		}
	case event.TypeHypothesisSubject:
		var p event.HypothesisSubject
		if json.Unmarshal(ev.Payload, &p) != nil {
			return n
		}
		for _, f := range p.Files {
			n.file(f.Path, f.SHA256)
		}
		if p.Patch != nil {
			n.file(p.Patch.Path, p.Patch.SHA256)
		}
		if p.Base != "" {
			n.Commits = append(n.Commits, p.Base)
		}
	case event.TypeResultRecord:
		var p event.ResultRecord
		if json.Unmarshal(ev.Payload, &p) != nil {
			return n
		}
		for _, e := range p.Evidence {
			if strings.HasPrefix(e, "http://") || strings.HasPrefix(e, "https://") {
				continue
			}
			clean := path.Clean(e)
			if ref, ok := event.EvidenceRef(clean); ok {
				n.Bundles = append(n.Bundles, Pin{Ref: ref.String(), Hash: p.EvidenceHashes[e]})
				continue
			}
			n.file(clean, p.EvidenceSHA256[e])
		}
	}
	return n
}

// Named lists the files ev names, read from the store's work-tree folder
// work or, for a file only the record holds, left out since the record
// keeps it already, with the hashed files of each bundle it pins and the
// commits its subject names. A named file found in neither is refused.
func Named(s *record.Store, work string, ev event.Event) ([]record.File, []string, error) {
	n := NamesOf(ev)
	var files []record.File
	for _, rel := range n.Files {
		src := filepath.Join(work, filepath.FromSlash(rel))
		if fi, err := os.Stat(src); err == nil && fi.Mode().IsRegular() {
			files = append(files, s.StoreFile(rel, src, n.Sums[rel]))
			continue
		}
		if s.HoldsFile(rel) {
			continue
		}
		return nil, nil, herr.Refused("%s, which the event names, is in neither the work tree nor the record", rel)
	}
	for _, b := range n.Bundles {
		bf, err := bundleFiles(s, work, b)
		if err != nil {
			return nil, nil, err
		}
		files = append(files, bf...)
	}
	return files, n.Commits, nil
}

// bundleFiles lists the hashed files of the bundle b pins, found from the
// work tree. A bundle the work tree lacks, or holds with another hash, adds
// nothing: the record's copy, when it has one, stays, and a pin no copy
// matches reads questionable, as it would on the work tree.
func bundleFiles(s *record.Store, work string, b Pin) ([]record.File, error) {
	load := recipe.Load
	if b.Rubric {
		load = recipe.LoadRubric
	}
	if r, err := load(work, b.Ref); err == nil && (b.Hash == "" || r.Hash == b.Hash) {
		return dirFiles(s, r.Path)
	}
	return nil, nil
}

// dirFiles lists every hashed regular file under the bundle at dir, in the
// store's repository.
func dirFiles(s *record.Store, dir string) ([]record.File, error) {
	if s.Repo == nil {
		return nil, nil
	}
	return dirFilesUnder(s.Repo.Top, dir)
}

// dirFilesUnder lists every hashed regular file under the bundle at dir,
// by its path below top, each naming the bundle's folder, so the record
// keeps a copy whose path it holds with other files whole beside it.
func dirFilesUnder(top, dir string) ([]record.File, error) {
	var out []record.File
	bundle, ok := record.FileUnder(top, dir)
	if !ok {
		return nil, nil
	}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		if rel != "." && recipe.Unhashed(filepath.ToSlash(rel)) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		f, ok := record.FileUnder(top, p)
		if ok {
			f.Bundle = bundle.Path
			out = append(out, f)
		}
		return nil
	})
	return out, err
}
