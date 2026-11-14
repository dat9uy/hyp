// Package contain finds Results whose verdict script hyp cannot vouch for
// (spec §7.6): a script cited outside recipes/, a pinned instrument without a
// clean command trace whose known answers passed, or a command trace that
// matches no pin.
package contain

import (
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/replay"
)

// Mark sets each Result record's Uncontained to its reasons, or nil when it
// is retracted or has none.
func Mark(root string, p *replay.Projection) {
	m := newMarker(root)
	for id, res := range p.Results {
		res.Uncontained = nil
		if res.Status == replay.ResRetracted {
			continue
		}
		if why := m.reasons(*p, id); len(why) > 0 {
			res.Uncontained = why
		}
	}
}

// Reasons lists, in rule order, why hyp cannot vouch for Result id's verdict
// script. root is the store root that evidence pointers are relative to.
func Reasons(root string, p replay.Projection, id string) []string {
	return newMarker(root).reasons(p, id)
}

// marker caches each ref's loaded bundle, nil when it does not load.
type marker struct {
	root    string
	bundles map[event.Ref]*recipe.Recipe
}

func newMarker(root string) *marker {
	return &marker{root: root, bundles: map[event.Ref]*recipe.Recipe{}}
}

// cited is one cited command.json: its pointer and its record, which is the
// zero Record with ok false when the file cannot be read or parsed.
type cited struct {
	ptr string
	rec cmdrecord.Record
	ok  bool
}

func (m *marker) reasons(p replay.Projection, id string) []string {
	r := p.Nodes[id]
	var why []string

	// A script cited outside a versioned bundle runs where hyp never sees it.
	for _, ptr := range r.Evidence {
		if m.script(ptr) {
			why = append(why, "script outside recipes/: "+path.Clean(ptr))
		}
	}

	pins := distinct(p.Results[id].Pins)
	records := m.records(r.Evidence)
	// A command trace among the Experiment's pinned records counts as
	// cited for reason 2: a reassessment is contained by the runs it read.
	// Reason 3 asks only the records this Result cites itself.
	var pinned []string
	if exp := p.Nodes[r.Parent]; exp != nil {
		for _, f := range exp.Records {
			pinned = append(pinned, f.Path)
		}
	}
	ran := append(slices.Clone(records), m.records(pinned)...)

	// A pinned bundle with commands needs a record of them run at that hash,
	// every command exiting 0 after its known answers passed; the
	// Experiment's own recipe included.
	for _, pin := range pins {
		if r := m.bundle(pin.Ref); r == nil || !r.IsInstrument() {
			continue
		}
		if !slices.ContainsFunc(ran, func(c cited) bool { return c.rec.CleanRun(pin.Ref, pin.Hash) }) {
			reason := "no command trace of " + pin.Ref.String() + " at " + h12(pin.Hash) + " with passing known answers"
			// Name why a cited record of this pin is not clean, so a failed
			// command is not read as failed known answers.
			for _, c := range ran {
				if u := c.rec.Unclean(pin.Ref, pin.Hash); u != "" {
					reason += "; in the cited record, " + u
					break
				}
			}
			why = append(why, reason)
		}
	}

	// A cited record must be a readable run of something this Result pins.
	for _, c := range records {
		if !c.ok || c.rec.Ref.IsZero() || c.rec.SHA256 == "" {
			why = append(why, "command trace "+c.ptr+" names no recipe run")
			continue
		}
		if !slices.ContainsFunc(pins, func(pin replay.Pin) bool { return pin.Ref == c.rec.Ref && pin.Hash == c.rec.SHA256 }) {
			why = append(why, "command trace of "+c.rec.Ref.String()+" at "+h12(c.rec.SHA256)+" matches no pin of this Result")
		}
	}

	return dedupe(why)
}

// dedupe keeps the first of each reason, so a pointer or record cited twice
// gives its reason once.
func dedupe(why []string) []string {
	var out []string
	for _, w := range why {
		if !slices.Contains(out, w) {
			out = append(out, w)
		}
	}
	return out
}

// sources are the source-file endings reason 1 counts besides recipe.IsScript.
var sources = []string{".go", ".js", ".ts", ".rb", ".pl", ".rs"}

// script reports whether ptr names a regular file, outside what a versioned
// bundle recipes/<name>@<version>/ hashes, that ends in .py or .sh, has an
// execute bit, starts with #!, or ends in one of sources. A file the bundle
// hash skips (a dot-entry or __pycache__) is outside: no pin covers it. URLs
// and missing files are not scripts.
func (m *marker) script(ptr string) bool {
	if isURL(ptr) || pinned(ptr) {
		return false
	}
	file := filepath.Join(m.root, ptr)
	if recipe.IsScript(file) {
		return true
	}
	fi, err := os.Stat(file)
	return err == nil && fi.Mode().IsRegular() && slices.Contains(sources, filepath.Ext(file))
}

// records parses each cited command.json, in citation order, skipping URLs.
// One that cannot be read or parsed is kept with ok false.
func (m *marker) records(evidence []string) []cited {
	var out []cited
	for _, ptr := range evidence {
		if isURL(ptr) || path.Base(ptr) != "command.json" {
			continue
		}
		rec, err := cmdrecord.Read(filepath.Join(m.root, ptr))
		out = append(out, cited{ptr: ptr, rec: rec, ok: err == nil})
	}
	return out
}

// bundle loads ref once, or returns nil when it does not load; questionable
// already reports such a ref.
func (m *marker) bundle(ref event.Ref) *recipe.Recipe {
	r, seen := m.bundles[ref]
	if !seen {
		var err error
		if r, err = recipe.Load(m.root, ref.String()); err != nil {
			r, _ = recipe.LoadRubric(m.root, ref.String())
		}
		m.bundles[ref] = r
	}
	return r
}

func isURL(ptr string) bool {
	return strings.HasPrefix(ptr, "http://") || strings.HasPrefix(ptr, "https://")
}

// pinned reports whether ptr names a file inside a versioned bundle that the
// bundle hash covers.
func pinned(ptr string) bool {
	ref, ok := event.EvidenceRef(ptr)
	if !ok {
		return false
	}
	rel := strings.TrimPrefix(path.Clean(ptr), "recipes/"+ref.String()+"/")
	return rel != path.Clean(ptr) && !recipe.Unhashed(rel)
}

// distinct keeps the first pin of each ref and hash, so citing several files
// of one bundle gives each reason once.
func distinct(pins []replay.Pin) []replay.Pin {
	var out []replay.Pin
	for _, pin := range pins {
		if !slices.ContainsFunc(out, func(o replay.Pin) bool { return o.Ref == pin.Ref && o.Hash == pin.Hash }) {
			out = append(out, pin)
		}
	}
	return out
}

// h12 is the first 12 hex digits of hash.
func h12(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}
