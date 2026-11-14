package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/record"
)

// Source is one work-tree store that hyp import reads: its name in the
// record, the folder that holds its store folder, and the root that its
// repository paths are relative to (the work tree, or a copy of a
// commit's tree).
type Source struct {
	Name, Src, Top string
}

// Import adds to the record of repo, in one write, every event of each
// source whose id the record lacks, with the bytes of each file it names,
// the files of each bundle it pins and the commits its subject names. It
// adds each event once, with all it names or not at all: when an event
// it would add names a file the source lacks or holds with other bytes
// than the sha256 the event stores, pins a bundle the source lacks or
// holds with another hash, names a subject commit the repository lacks,
// or has an id that another store holds, it refuses with exit 1 and adds
// nothing. It changes no event the record holds. It returns how many
// events it added from each source; run again, it adds none.
func Import(repo *record.Repo, srcs []Source, warn io.Writer) ([]int, error) {
	type read struct {
		files map[string][]event.Event
		raw   map[string][][]byte
	}
	reads := make([]read, len(srcs))
	for i, src := range srcs {
		if fi, err := os.Stat(record.EventsDir(src.Src)); err != nil || !fi.IsDir() {
			return nil, herr.BadArgs("%s holds no store", src.Src)
		}
		files, raw, lineErrs, err := readLines(src.Src)
		if err != nil {
			return nil, err
		}
		for _, le := range lineErrs {
			fmt.Fprintf(warn, "hyp import: %s:%d does not parse and is not imported: %v\n", le.File, le.Line, le.Err)
		}
		reads[i] = read{files, raw}
	}
	if len(srcs) == 0 {
		return nil, nil
	}
	s, err := repo.StoreNamed(srcs[0].Name)
	if err != nil {
		return nil, err
	}
	release, err := s.Lock()
	if err != nil {
		return nil, err
	}
	defer release()
	ids, err := s.StoreIDs()
	if err != nil {
		return nil, err
	}
	// owner is the store that holds each id: the record's, then each
	// event this import adds.
	owner := map[string]string{}
	for name, l := range ids {
		for _, id := range l {
			if _, ok := owner[id]; !ok || name < owner[id] {
				owner[id] = name
			}
		}
	}
	lines := map[string]map[string][][]byte{}
	var add []record.File
	var commits []string
	adding := map[string]bool{}
	adoptedBundle := map[string]bool{}
	counts := make([]int, len(srcs))
	for i, src := range srcs {
		view := s.Sibling(src.Name)
		files, raw := reads[i].files, reads[i].raw
		for _, actor := range slices.Sorted(maps.Keys(files)) {
			for j, ev := range files[actor] {
				if o, ok := owner[ev.ID]; ok {
					if o == src.Name {
						continue
					}
					return nil, herr.Refused("event %s of the store %s is also an event of the store %s; an id is one event of one store, so hyp import adds nothing", ev.ID, src.Name, o)
				}
				owner[ev.ID] = src.Name
				n := NamesOf(ev)
				for _, rel := range n.Files {
					f, err := namedFile(view, src, ev, n, rel)
					if err != nil {
						return nil, err
					}
					if key := src.Name + "\x00" + rel; !adding[key] {
						adding[key] = true
						add = append(add, f)
					}
				}
				for _, b := range n.Bundles {
					bf, dir, err := pinnedBundle(src, ev, b)
					if err != nil {
						return nil, err
					}
					if !adoptedBundle[dir] {
						adoptedBundle[dir] = true
						add = append(add, bf...)
					}
				}
				for _, c := range n.Commits {
					if _, ok := repo.IsCommit(c); !ok {
						return nil, herr.Refused("%s, the subject commit that %s names, is not in %s; fetch it, then run hyp import again, which adds nothing until then", c, ev.ID, repo.Top)
					}
					commits = append(commits, c)
				}
				if lines[src.Name] == nil {
					lines[src.Name] = map[string][][]byte{}
				}
				lines[src.Name][actor] = append(lines[src.Name][actor], raw[actor][j])
				counts[i]++
			}
		}
	}
	if len(lines) == 0 {
		return counts, repo.EnsureRecord()
	}
	for _, byActor := range lines {
		for actor := range byActor {
			slices.SortStableFunc(byActor[actor], func(a, b []byte) int { return compareLineID(a, b) })
		}
	}
	return counts, s.AppendStores(lines, add, commits)
}

// namedFile is the File of rel, which ev names, read from src: refused when
// src lacks it or holds it with other bytes than the sha256 ev stores.
func namedFile(view *record.Store, src Source, ev event.Event, n Names, rel string) (record.File, error) {
	p := filepath.Join(src.Src, filepath.FromSlash(rel))
	fi, err := os.Stat(p)
	if err != nil || !fi.Mode().IsRegular() {
		return record.File{}, herr.Refused("%s, which %s names, is not in %s; hyp import adds an event with the bytes it names or not at all, so it adds nothing", rel, ev.ID, src.Src)
	}
	sum := n.Sums[rel]
	got, err := fileSHA256(p)
	if err != nil {
		return record.File{}, err
	}
	if sum != "" && got != sum {
		return record.File{}, herr.Refused("%s, which %s names, hashes %s in %s, not %s as the event stores; hyp import adds an event with the bytes it names or not at all, so it adds nothing", rel, ev.ID, got[:12], src.Src, sum[:min(12, len(sum))])
	}
	return view.StoreFile(rel, p, sum), nil
}

// pinnedBundle lists the files of bundle b, which ev pins, read from src,
// with the bundle's folder: refused when src lacks the bundle or holds it
// with another hash than the pin.
func pinnedBundle(src Source, ev event.Event, b Pin) ([]record.File, string, error) {
	load := recipe.Load
	if b.Rubric {
		load = recipe.LoadRubric
	}
	r, err := load(src.Src, b.Ref)
	if err != nil || (b.Hash != "" && r.Hash != b.Hash) {
		return nil, "", herr.Refused("bundle %s, which %s pins, is not in %s with its pinned hash; hyp import adds an event with what it pins or not at all, so it adds nothing", b.Ref, ev.ID, src.Src)
	}
	bf, err := dirFilesUnder(src.Top, r.Path)
	return bf, r.Path, err
}

// compareLineID orders two event lines by their id, then by their bytes.
func compareLineID(a, b []byte) int {
	ea, erra := ParseEventLine(a)
	eb, errb := ParseEventLine(b)
	if erra == nil && errb == nil && ea != eb {
		if ea < eb {
			return -1
		}
		return 1
	}
	return slices.Compare(a, b)
}

// NamedFiles lists every file the events of the store read at root name,
// by slash path relative to the store's folder.
func NamedFiles(root string) (map[string]bool, error) {
	evs, _, err := Events(root)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, ev := range evs {
		for _, f := range NamesOf(ev).Files {
			out[f] = true
		}
	}
	return out, nil
}

// fileSHA256 is the hex sha256 of the bytes of the file at p.
func fileSHA256(p string) (string, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}
