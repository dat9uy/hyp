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

// Import adds to the record of repo every event of the store read at src
// that the record's store name lacks, by id, in one write, with the bytes
// of each file those events name and the commits their subjects name. top
// is the root src's repository paths are relative to: the repository's
// work tree, or a copy of a commit's tree. A named file src lacks or holds
// with other bytes than the sha256 its event stores, or a bundle src holds
// with another hash than its pin, is left out and printed to warn; a later
// import from a src that holds the file's bytes, or the bundle with its
// pinned hash, adds them to the event the record holds, as it adds a
// subject commit the repository holds now and the record does not reach.
// It returns how many events, and how many files and commits for events
// the record held, it added; run again, it adds none.
func Import(repo *record.Repo, name, src, top string, warn io.Writer) (int, int, error) {
	if fi, err := os.Stat(record.EventsDir(src)); err != nil || !fi.IsDir() {
		return 0, 0, herr.BadArgs("%s holds no store", src)
	}
	files, raw, lineErrs, err := readLines(src)
	if err != nil {
		return 0, 0, err
	}
	for _, le := range lineErrs {
		fmt.Fprintf(warn, "hyp import: %s:%d does not parse and is not imported: %v\n", le.File, le.Line, le.Err)
	}
	s, err := repo.StoreNamed(name)
	if err != nil {
		return 0, 0, err
	}
	release, err := s.Lock()
	if err != nil {
		return 0, 0, err
	}
	defer release()
	have := map[string]bool{}
	if s.InRecord {
		root, err := s.Root()
		if err != nil {
			return 0, 0, err
		}
		evs, _, err := Events(root)
		if err != nil {
			return 0, 0, err
		}
		for _, ev := range evs {
			have[ev.ID] = true
		}
	}
	lines := map[string][][]byte{}
	var add []record.File
	var commits []string
	added := 0
	actors := make([]string, 0, len(files))
	for a := range files {
		actors = append(actors, a)
	}
	slices.Sort(actors)
	held := maps.Clone(have)
	adding := map[string]bool{}
	// namedFile adds the file rel that ev names from src when src holds it
	// with the sha256 the event stores, and else prints that the record
	// lacks it. A file the record holds already is left as it is.
	namedFile := func(ev event.Event, n Names, rel string) error {
		if s.HoldsFile(rel) || adding[rel] {
			return nil
		}
		p := filepath.Join(src, filepath.FromSlash(rel))
		fi, err := os.Stat(p)
		if err != nil || !fi.Mode().IsRegular() {
			fmt.Fprintf(warn, "hyp import: %s, which %s names, is not in %s; the record lacks it\n", rel, ev.ID, src)
			return nil
		}
		sum := n.Sums[rel]
		got, err := fileSHA256(p)
		if err != nil {
			return err
		}
		if sum != "" && got != sum {
			fmt.Fprintf(warn, "hyp import: %s, which %s names, hashes %s in %s, not %s as the event stores; the record lacks those bytes\n", rel, ev.ID, got[:12], src, sum[:min(12, len(sum))])
			return nil
		}
		adding[rel] = true
		add = append(add, s.StoreFile(rel, p, sum))
		return nil
	}
	// recordHolds reports whether the record keeps a copy of bundle b with
	// its pinned hash, read once from the record's store.
	var recordHashes map[string]string
	recordHolds := func(b Pin) (bool, error) {
		if !s.InRecord {
			return false, nil
		}
		if recordHashes == nil {
			root, err := s.Root()
			if err != nil {
				return false, err
			}
			recordHashes = recipe.Hashes(root)
		}
		if b.Hash == "" {
			_, ok := recordHashes[b.Ref]
			return ok, nil
		}
		return recipe.Holds(recordHashes, b.Ref, b.Hash), nil
	}
	adoptedBundle := map[Pin]bool{}
	var lacked []string
	// pinnedBundle adds the files of bundle b that ev pins from src when src
	// holds it with the pinned hash, and else prints that the record lacks
	// it. For an event the record holds, a bundle the record keeps already
	// is left as it is.
	pinnedBundle := func(ev event.Event, b Pin, isHeld bool) error {
		if adoptedBundle[b] {
			return nil
		}
		if isHeld {
			ok, err := recordHolds(b)
			if err != nil || ok {
				return err
			}
		}
		load := recipe.Load
		if b.Rubric {
			load = recipe.LoadRubric
		}
		r, err := load(src, b.Ref)
		if err != nil || (b.Hash != "" && r.Hash != b.Hash) {
			fmt.Fprintf(warn, "hyp import: bundle %s that %s pins is not in %s with its hash; the record lacks it\n", b.Ref, ev.ID, src)
			return nil
		}
		bf, err := dirFilesUnder(top, r.Path)
		if err != nil {
			return err
		}
		adoptedBundle[b] = true
		add = append(add, bf...)
		return nil
	}
	for _, actor := range actors {
		for i, ev := range files[actor] {
			n := NamesOf(ev)
			isHeld := held[ev.ID]
			if !isHeld {
				if have[ev.ID] {
					continue
				}
				have[ev.ID] = true
				lines[actor] = append(lines[actor], raw[actor][i])
				added++
				commits = append(commits, n.Commits...)
			} else {
				for _, c := range n.Commits {
					if !slices.Contains(lacked, c) && s.LacksCommit(c) {
						lacked = append(lacked, c)
					}
				}
			}
			// An event the record holds gets each named file, pinned bundle
			// and subject commit an earlier import could not add.
			for _, rel := range n.Files {
				if err := namedFile(ev, n, rel); err != nil {
					return 0, 0, err
				}
			}
			for _, b := range n.Bundles {
				if err := pinnedBundle(ev, b, isHeld); err != nil {
					return 0, 0, err
				}
			}
		}
	}
	if added == 0 && len(add) == 0 && len(lacked) == 0 {
		return 0, 0, nil
	}
	for actor := range lines {
		slices.SortStableFunc(lines[actor], func(a, b []byte) int { return compareLineID(a, b) })
	}
	return added, len(add) + len(lacked), s.Append(lines, add, append(commits, lacked...))
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
