package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/dat9uy/hyp/internal/check"
	"github.com/dat9uy/hyp/internal/contain"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/record"
	"github.com/dat9uy/hyp/internal/render"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/signals"
	"github.com/dat9uy/hyp/internal/store"
)

// reader is one read command: it replays the whole log and can replay again
// up to an --as-of point. hyp keeps no index; every read replays (ADR 0006).
type reader struct {
	store    *record.Store
	dir      string // the store's work-tree folder
	root     string // the directory every read reads below (record.Store.Root)
	evs      []event.Event
	lineErrs []store.LineError
	hashes   map[string]string
	full     replay.Projection
}

func openReader() (*reader, error) {
	r, err := openEvents()
	if err != nil {
		return nil, err
	}
	r.hashes = recipe.Hashes(r.root)
	r.full = r.project(r.evs, "")
	return r, nil
}

// openEvents reads the log without replaying it, for a command that
// replays the events itself.
func openEvents() (*reader, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	r := &reader{}
	if r.store, err = store.Open(cwd); err != nil {
		return nil, err
	}
	r.dir = r.store.Dir
	if r.root, err = r.store.Root(); err != nil {
		return nil, err
	}
	if r.evs, r.lineErrs, err = store.Events(r.root); err != nil {
		return nil, err
	}
	return r, nil
}

// project replays evs up to asOf and marks the uncontained Results, so no
// read path renders a projection without the marks.
func (r *reader) project(evs []event.Event, asOf string) replay.Projection {
	p := replay.Replay(evs, r.hashes, asOf)
	contain.Mark(r.root, &p)
	return p
}

// ancestors lists id's ancestors, nearest first.
func ancestors(p replay.Projection, id string) []string {
	var out []string
	for n := p.Nodes[p.Nodes[id].Parent]; n != nil; n = p.Nodes[n.Parent] {
		out = append(out, n.ID)
	}
	return out
}

// at replays up to point: "" for the whole log, an event id, the display id
// or alias of a node-creating event, or an RFC3339 time (the last applied
// event, in merge order, whose ts is at or before it).
func (r *reader) at(point string) (replay.Projection, error) {
	if point == "" {
		return r.full, nil
	}
	for _, ev := range r.evs {
		if ev.ID == point {
			return r.project(r.evs, point), nil
		}
	}
	if t, err := time.Parse(time.RFC3339Nano, point); err == nil {
		last := ""
		for _, ev := range r.full.Applied {
			if ts, err := time.Parse(time.RFC3339Nano, ev.TS); err == nil && !ts.After(t) {
				last = ev.ID
			}
		}
		if last == "" {
			return r.project(nil, ""), nil
		}
		return r.project(r.evs, last), nil
	}
	if dateRe.MatchString(point) {
		return replay.Projection{}, usageErr("%q is not an RFC3339 time (want e.g. 2026-10-24T09:00:00Z)", point)
	}
	id, err := resolveRef(r.full, point)
	if err != nil {
		return replay.Projection{}, err
	}
	return r.project(r.evs, id), nil
}

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

func readFlags(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Run prints the one error line
	return fs
}

func cmdStatus(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("status", stderr)
	asOf := fs.String("as-of", "", "event id, node ref or RFC3339 time")
	maxLines := fs.Int("max-lines", render.DefaultMaxLines, "line budget (0 for unlimited)")
	live := fs.Bool("live", false, "leave out of FRONTIER each contradicted Hypothesis with no marked Experiment, counted under its question; for the owner, since a runner needs the contradicted rivals")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	} else if len(pos) > 1 {
		return usageErr("usage: hyp status [G-id] [--as-of REF|TIME] [--max-lines N] [--live]")
	}
	r, err := openReader()
	if err != nil {
		return err
	}
	p, err := r.at(*asOf)
	if err != nil {
		return err
	}
	// hyp status G-… prints one Goal; with no Goal, a store with many Goals
	// prints the GOALS table (ADR 0020).
	goal := ""
	if len(pos) == 1 {
		id, err := resolveRef(p, pos[0])
		if err != nil {
			return usageErr("%v", err)
		}
		if p.Nodes[id].Kind != event.KindGoal {
			return usageErr("hyp status takes a Goal: %s is a %s", pos[0], p.Nodes[id].Kind)
		}
		goal = id
	}
	// A line that does not parse is not replayed. Say so on stderr, so
	// stdout stays the view a harness parses; hyp check reports it as an
	// error.
	for _, le := range r.lineErrs {
		fmt.Fprintf(stderr, "warn: skipped %s:%d: %v; run hyp check\n", le.File, le.Line, le.Err)
	}
	fmt.Fprint(stdout, render.StatusOf(p, goal, *maxLines, *live))
	return nil
}

func cmdShow(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("show", stderr)
	asOf := fs.String("as-of", "", "event id, node ref or RFC3339 time")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("usage: hyp show REF|R-id#n|G-id#n|FILE[@SHA256] [--as-of REF|TIME]")
	}
	r, err := openReader()
	if err != nil {
		return err
	}
	if strings.Contains(pos[0], "/") {
		return showFile(r, pos[0], stdout)
	}
	p, err := r.at(*asOf)
	if err != nil {
		return err
	}
	if k, ok := event.ParseKey(pos[0]); ok && k.Kind == event.KeyLimit {
		res, err := resolveRef(p, k.Result)
		if err != nil {
			return err
		}
		k.Result = res
		l, ok := p.Limit(k.String())
		if !ok {
			return usageErr("%s names no limit of a live Result and no line of the Goal", pos[0])
		}
		fmt.Fprint(stdout, render.ShowLimit(p, l))
		return nil
	}
	id, err := resolveRef(p, pos[0])
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, render.Show(p, id, ancestors(p, id)))
	return nil
}

// showFile prints the bytes of a file an applied event names, a path
// relative to the store's folder as the event holds it: in a record
// repository the record's copy, which no later change to the work tree
// reaches. When events name the path with other bytes, which hyp sync of
// two clones can join, PATH@SHA256 picks one by a prefix of its sha256.
func showFile(r *reader, arg string, stdout io.Writer) error {
	clean := path.Clean(filepath.ToSlash(arg))
	pick := ""
	if at := strings.LastIndex(clean, "@"); at > 0 && isHexPrefix(clean[at+1:]) {
		clean, pick = clean[:at], clean[at+1:]
	}
	named := false
	var sums []string
	for _, ev := range r.full.Applied {
		n := store.NamesOf(ev)
		if slices.Contains(n.Files, clean) {
			named = true
			if sum := n.Sums[clean]; sum != "" && !slices.Contains(sums, sum) {
				sums = append(sums, sum)
			}
			continue
		}
		if ev.Type == event.TypeResultRecord {
			if r.full.Nodes[ev.ID] != nil && slices.ContainsFunc(r.full.Nodes[ev.ID].Evidence, func(e string) bool { return path.Clean(e) == clean }) {
				named = true
			}
		}
	}
	if !named {
		return usageErr("%s is no file an event of this store names", arg)
	}
	// A file of a bundle is named by the bundle hash its Results pin: each
	// pinned copy the record keeps gives its bytes, so one pointer can name
	// several contents, as a named file can.
	copyOf := map[string]string{}
	ref, inBundle := event.EvidenceRef(clean)
	if inBundle {
		_, rest, _ := strings.Cut(strings.TrimPrefix(clean, "recipes/"), "/")
		var hashes []string
		for _, ev := range r.full.Applied {
			res := r.full.Results[ev.ID]
			if ev.Type != event.TypeResultRecord || res == nil {
				continue
			}
			for _, pin := range res.Pins {
				if path.Clean(pin.Pointer) == clean && !slices.Contains(hashes, pin.Hash) {
					hashes = append(hashes, pin.Hash)
				}
			}
		}
		for _, h := range hashes {
			b, err := recipe.LoadPinned(r.root, ref.String(), h)
			if err != nil || b == nil || b.Hash != h {
				continue
			}
			f := filepath.Join(b.Path, filepath.FromSlash(rest))
			s, err := fileSHA256(f)
			if err != nil {
				continue
			}
			if _, ok := copyOf[s]; !ok {
				copyOf[s] = f
				sums = append(sums, s)
			}
		}
	}
	sum := ""
	switch {
	case pick != "":
		for _, s := range sums {
			if strings.HasPrefix(s, pick) {
				if sum != "" {
					return usageErr("%s: the prefix %s fits more than one sha256 the events name", arg, pick)
				}
				sum = s
			}
		}
		if sum == "" {
			return usageErr("%s: no event names %s with a sha256 that starts %s", arg, clean, pick)
		}
	case len(sums) > 1:
		var picks []string
		for _, s := range sums {
			picks = append(picks, clean+"@"+s[:12])
		}
		return usageErr("events name %s with %d different contents; show one: %s", clean, len(sums), strings.Join(picks, ", "))
	case len(sums) == 1:
		sum = sums[0]
	}
	file := record.NamedPath(r.root, clean, sum)
	if f, ok := copyOf[sum]; ok {
		file = f
	} else if inBundle {
		b, err := recipe.Load(r.root, ref.String())
		if err != nil {
			return err
		}
		_, rest, _ := strings.Cut(strings.TrimPrefix(clean, "recipes/"), "/")
		file = filepath.Join(b.Path, filepath.FromSlash(rest))
	}
	data, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return herr.Env("%s: an event names it, but this store holds no copy of its bytes; hyp import from a commit or folder that holds them adds them", arg)
	}
	if err != nil {
		return herr.Env("%s: %v", arg, err)
	}
	if pick != "" {
		if got := sha256.Sum256(data); hex.EncodeToString(got[:]) != sum {
			return herr.Refused("%s: the bytes kept at %s have sha256 %s, not %s as an event names it; no copy with that sha256 is kept", arg, clean, hex.EncodeToString(got[:])[:12], sum[:12])
		}
	}
	_, err = stdout.Write(data)
	return err
}

// isHexPrefix reports whether s is 4 to 64 lower-case hex digits.
func isHexPrefix(s string) bool {
	if len(s) < 4 || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !('0' <= c && c <= '9' || 'a' <= c && c <= 'f') {
			return false
		}
	}
	return true
}

func cmdDiff(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("diff", stderr)
	since := fs.String("since", "", "event id, node ref or RFC3339 time (required)")
	if pos, err := parse(fs, args); err != nil {
		return err
	} else if len(pos) != 0 || *since == "" {
		return usageErr("usage: hyp diff --since REF|TIME")
	}
	r, err := openReader()
	if err != nil {
		return err
	}
	before, err := r.at(*since)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, render.Diff(before, r.full))
	return nil
}

func cmdRender(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("render", stderr)
	md := fs.Bool("markdown", false, "render the full tree as Markdown (required)")
	asOf := fs.String("as-of", "", "event id, node ref or RFC3339 time")
	if pos, err := parse(fs, args); err != nil {
		return err
	} else if len(pos) != 0 || !*md {
		return usageErr("usage: hyp render --markdown [--as-of REF|TIME]")
	}
	r, err := openReader()
	if err != nil {
		return err
	}
	p, err := r.at(*asOf)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, render.Markdown(p))
	return nil
}

func cmdCheck(args []string, stdout, stderr io.Writer) error {
	const use = "usage: hyp check"
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		return helpText(use)
	}
	if len(args) != 0 {
		return usageErr(use)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	s, err := store.Open(cwd)
	if err != nil {
		return err
	}
	report, errs := check.RunStore(s)
	for _, l := range report {
		fmt.Fprintln(stdout, l)
	}
	if errs > 0 {
		return herr.Refused("check found %d errors", errs)
	}
	return nil
}

func cmdSignals(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("signals", stderr)
	asJSON := fs.Bool("json", false, "print rounds and signals as JSON")
	if pos, err := parse(fs, args); err != nil {
		return err
	} else if len(pos) != 0 {
		return usageErr("usage: hyp signals [--json]")
	}
	// Compute replays the events itself, once.
	r, err := openEvents()
	if err != nil {
		return err
	}
	rep := signals.Compute(r.evs)
	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}
	for _, s := range rep.Signals {
		fmt.Fprintln(stdout, s.Detail)
	}
	return nil
}

// cmdEvents prints the applied events in merge order, as stored, and the
// rejected ones apart, for scripts that would otherwise read the events
// files (ADR 0006). A shape change bumps format.
func cmdEvents(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("events", stderr)
	asJSON := fs.Bool("json", false, "print as JSON (required; the only form)")
	if pos, err := parse(fs, args); err != nil {
		return err
	} else if len(pos) != 0 || !*asJSON {
		return usageErr("usage: hyp events --json")
	}
	r, err := openReader()
	if err != nil {
		return err
	}
	type rejection struct {
		Event  string `json:"event"`
		Reason string `json:"reason"`
	}
	out := struct {
		Format     int           `json:"format"`
		Events     []event.Event `json:"events"`
		Rejections []rejection   `json:"rejections"`
	}{Format: 1, Events: r.full.Applied, Rejections: []rejection{}}
	if out.Events == nil {
		out.Events = []event.Event{}
	}
	for _, x := range r.full.Rejections {
		out.Rejections = append(out.Rejections, rejection{x.EventID, x.Reason})
	}
	return json.NewEncoder(stdout).Encode(out)
}
