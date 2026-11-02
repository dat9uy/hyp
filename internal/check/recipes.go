package check

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/sandbox"
	"github.com/dat9uy/hyp/internal/store"
)

// ErrNoRepo is returned by Recipes when start is not inside a git repository.
var ErrNoRepo = errors.New("not inside a git repository")

// findStores lists the store roots under top, sorted: each directory holding
// .tree. It skips .git, any directory below top that holds its own .git (a
// nested repository or a worktree checkout) and a bundle's known-answers/
// (case inputs are fixtures hashed into the bundle, not trees), and does not
// descend .tree.
func findStores(top string) ([]string, error) {
	var stores []string
	err := filepath.WalkDir(top, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk failed at %s: %w", p, err)
		}
		if !d.IsDir() {
			return nil
		}
		if p != top {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			if _, err := os.Lstat(filepath.Join(p, ".git")); err == nil {
				return filepath.SkipDir
			}
			if d.Name() == "known-answers" {
				if _, err := os.Stat(filepath.Join(filepath.Dir(p), "recipe.md")); err == nil {
					return filepath.SkipDir
				}
			}
		}
		if d.Name() == ".tree" {
			stores = append(stores, filepath.Dir(p))
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(stores)
	return stores, nil
}

// Recipes checks every store under the repository root above start: each pin
// to a versioned recipe bundle (an Experiment's recipe and a Result's
// evidence_hashes) must still hash as recorded. failures counts drifted pins
// and stores that do not load. It also
// names each versioned bundle the stores' libraries hold as frozen (a tree
// pins its current hash) or draft (none does), and warns about each current
// instrument whose known answers are weak (no cases, or a command without a
// must-pass or a must-fail case), about each Evidence or Standalone command
// whose External: line is missing or TODO, and about each .go file in a bundle when the
// repository root holds go.mod; those lines are never failures.
//
// A pin to name@N whose bundle is missing, while a folder recipes/name/ in one
// of the store's recipe directories hashes as the pin, is hinted: the first
// drift line that folder explains is followed by a hint to rename it to
// name@N, and hinted counts every such failure.
func Recipes(start string) (report []string, failures, hinted int, err error) {
	top, ok := recipe.RepoRoot(start)
	if !ok {
		return nil, 0, 0, ErrNoRepo
	}
	stores, err := findStores(top)
	if err != nil {
		return nil, 0, 0, err
	}
	pins, drift := 0, 0
	pinned := map[string]bool{} // "ref hash" of every pin
	bundles := map[string]bool{}
	hintedDirs := map[string]bool{} // leftover folders already given a hint
	for _, root := range stores {
		for _, d := range recipe.Dirs(root) {
			mds, _ := filepath.Glob(filepath.Join(d, "*@*", "recipe.md"))
			for _, md := range mds {
				bundles[filepath.Dir(md)] = true
			}
		}
		rel, _ := filepath.Rel(top, root)
		rel = filepath.ToSlash(rel)
		p, _, err := store.Load(root)
		if err != nil {
			report = append(report, fmt.Sprintf("error: %s: %v", rel, err))
			failures++
			continue
		}
		hashes := recipe.Hashes(root)
		check := func(id string, pin replay.Pin) {
			pins++
			pinned[pin.Ref.String()+" "+pin.Hash] = true
			now, ok := hashes[pin.Ref.String()]
			if ok && now == pin.Hash {
				return
			}
			if !ok {
				now = "missing"
			}
			file := pin.Pointer
			if file == "" {
				file = "recipes/" + pin.Ref.String() + "/"
			}
			report = append(report, fmt.Sprintf("drift: %s %s %s: recorded %s, now %s",
				rel, event.DisplayID(p.Nodes[id].Kind, id), file, pin.Hash, now))
			drift++
			failures++
			if ok {
				return
			}
			if dir, version := leftover(root, pin); dir != "" {
				hinted++
				if !hintedDirs[dir] {
					hintedDirs[dir] = true
					at, _ := filepath.Rel(top, dir)
					at = filepath.ToSlash(at)
					report = append(report, fmt.Sprintf("hint: %s/ holds %s (same hash) in a folder whose name has no version; rename it: git mv %s %s@%s",
						at, pin.Ref, at, at, version))
				}
			}
		}
		// A Result's pins include its Experiment's recipe, so an Experiment a
		// Result reached, even a retracted one, is not checked again. Every
		// other Experiment pins its recipe itself, planned or closed.
		reached := map[string]bool{}
		for _, id := range resultIDs(p) {
			reached[p.Nodes[id].Parent] = true
			for _, pin := range p.Results[id].Pins {
				check(id, pin)
			}
		}
		var exps []string
		for id := range p.Exps {
			if p.Nodes[id].Recipe != event.RecipeNone && !reached[id] {
				exps = append(exps, id)
			}
		}
		sort.Strings(exps)
		for _, id := range exps {
			n := p.Nodes[id]
			// validate admits only name@version or none, so the parse holds.
			ref, _ := event.ParseRef(n.Recipe)
			check(id, replay.Pin{Ref: ref, Hash: n.RecipeHash})
		}
	}
	var dirs []string
	for dir := range bundles {
		rel, _ := filepath.Rel(top, dir)
		dirs = append(dirs, filepath.ToSlash(rel))
	}
	sort.Strings(dirs)
	var warns []string
	for _, rel := range dirs {
		dir := filepath.Join(top, filepath.FromSlash(rel))
		name, version, hash, err := recipe.HashOnly(dir)
		if err == nil && pinned[event.Ref{Name: name, Version: version}.String()+" "+hash] {
			report = append(report, "frozen: "+rel)
		} else {
			report = append(report, "draft: "+rel)
		}
		for _, w := range recipe.GoFileWarnings(dir) {
			warns = append(warns, "warn: "+rel+"/"+w)
		}
		if r, err := recipe.Open(dir); err == nil {
			for _, w := range sandbox.WeakCases(r) {
				if w == "instrument with no known answers" {
					warns = append(warns, "warn: "+rel+": "+w)
				} else {
					warns = append(warns, "warn: "+rel+" "+w)
				}
			}
			for _, w := range sandbox.ExternalWarnings(r) {
				warns = append(warns, "warn: "+rel+" "+w)
			}
		}
	}
	report = append(report, warns...)
	report = append(report, fmt.Sprintf("recipe check: %d trees, %d pins, %d drift", len(stores), pins, drift))
	return report, failures, hinted, nil
}

// leftover finds the folder recipes/<name>/, without a version, that the first
// of root's recipe directories holding one has for pin name@N, and returns it
// with N when it hashes as the pin; "" otherwise. It reads no library lookup:
// Load and Scan never resolve such a folder.
func leftover(root string, pin replay.Pin) (dir, version string) {
	if !pin.Ref.Versioned() {
		return "", ""
	}
	name, version := pin.Ref.Name, pin.Ref.Version
	for _, d := range recipe.Dirs(root) {
		dir := filepath.Join(d, name)
		if _, err := os.Stat(filepath.Join(dir, "recipe.md")); err != nil {
			continue
		}
		if _, _, hash, err := recipe.HashOnly(dir); err == nil && hash == pin.Hash {
			return dir, version
		}
		return "", ""
	}
	return "", ""
}
