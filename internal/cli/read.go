package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"time"

	"github.com/dat9uy/hyp/internal/check"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/render"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/signals"
	"github.com/dat9uy/hyp/internal/store"
)

// reader is one read command: it replays the whole log and can replay again
// up to an --as-of point. hyp keeps no index; every read replays (ADR 0006).
type reader struct {
	hist     *store.History
	evs      []event.Event
	lineErrs []store.LineError
	hashes   replay.Hashes
	full     replay.Projection
}

func openReader() (*reader, error) { return openReaderWith(store.OpenHistory) }

// openReaderWith reads and replays the history open finds from the
// current directory.
func openReaderWith(open func(string) (*store.History, error)) (*reader, error) {
	r, err := openEventsWith(open)
	if err != nil {
		return nil, err
	}
	r.hashes = r.hist.Hashes()
	r.full = r.project(r.evs, "")
	return r, nil
}

// openEvents reads the log without replaying it, for a command that
// replays the events itself.
func openEvents() (*reader, error) { return openEventsWith(store.OpenHistory) }

func openEventsWith(open func(string) (*store.History, error)) (*reader, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	r := &reader{}
	if r.hist, err = open(cwd); err != nil {
		return nil, err
	}
	if r.evs, r.lineErrs, err = r.hist.Events(); err != nil {
		return nil, err
	}
	return r, nil
}

// project replays evs up to asOf and marks the uncontained Results, so no
// read path renders a projection without the marks.
func (r *reader) project(evs []event.Event, asOf string) replay.Projection {
	return r.hist.Replay(evs, r.hashes, asOf)
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
	id, err := resolveRef(r.full, r.hist.Current, point)
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
	all := fs.Bool("all", false, "every Goal of the repository's history, and every item awaiting the owner, grouped by Goal; the default where store discovery finds no store")
	live := fs.Bool("live", false, "leave out of FRONTIER each contradicted Hypothesis with no marked Experiment, counted under its question; for the owner, since a runner needs the contradicted rivals")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	} else if len(pos) > 1 {
		return usageErr("usage: hyp status [G-id | --all] [--as-of REF|TIME] [--max-lines N] [--live]")
	} else if len(pos) == 1 && *all {
		return usageErr("hyp status takes a Goal or --all, not both")
	}
	// --all reads the repository's history from any folder, also from one
	// whose store adds nothing to it.
	open := store.OpenHistory
	if *all {
		open = store.OpenRepoHistory
	}
	r, err := openReaderWith(open)
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
		id, err := resolveRef(p, r.hist.Current, pos[0])
		if err != nil && !*all {
			// A Goal the history found from here does not hold may be held
			// by the repository's history, when this folder's store adds
			// nothing to it.
			if rr, rerr := openReaderWith(store.OpenRepoHistory); rerr == nil && !sameHistory(rr.hist, r.hist) {
				if rp, perr := rr.at(*asOf); perr == nil {
					if rid, rerr := resolveRef(rp, rr.hist.Current, pos[0]); rerr == nil {
						r, p, id, err = rr, rp, rid, nil
					}
				}
			}
		}
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
	if goal != "" {
		// What the Goal's store folder prints: that store's rejections.
		fmt.Fprint(stdout, render.StatusOf(onlyStoreRejections(p, p.Nodes[goal].Store), goal, *maxLines, *live))
		return nil
	}
	// In a store's folder, the Goals and rejections that store holds; with
	// --all, or where discovery finds no store, all of the history's.
	view := r.hist.Current
	if *all {
		view = ""
	}
	goals := storeGoals(&store.History{Current: view}, p)
	fmt.Fprint(stdout, render.StatusOfGoals(onlyStoreRejections(p, view), goals, *maxLines, *live))
	return nil
}

// onlyStoreRejections is p with only the rejections of store s, or p when
// s is "", the whole history.
func onlyStoreRejections(p replay.Projection, s string) replay.Projection {
	if s == "" {
		return p
	}
	q := p
	q.Rejections = nil
	for _, x := range p.Rejections {
		if x.Store == s {
			q.Rejections = append(q.Rejections, x)
		}
	}
	return q
}

// sameHistory reports whether a and b read the same stores.
func sameHistory(a, b *store.History) bool {
	return a.Root == b.Root && slices.Equal(a.Stores, b.Stores)
}

func cmdShow(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("show", stderr)
	asOf := fs.String("as-of", "", "event id, node ref or RFC3339 time")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("usage: hyp show REF|R-id#n|G-id#n [--as-of REF|TIME]")
	}
	r, err := openReader()
	if err != nil {
		return err
	}
	p, err := r.at(*asOf)
	if err != nil {
		return err
	}
	if k, ok := event.ParseKey(pos[0]); ok && k.Kind == event.KeyLimit {
		res, err := resolveRef(p, r.hist.Current, k.Result)
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
	id, err := resolveRef(p, r.hist.Current, pos[0])
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, render.Show(p, id, ancestors(p, id)))
	return nil
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
	fmt.Fprint(stdout, render.Markdown(render.OnlyGoals(p, storeGoals(r.hist, p))))
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
	h, err := store.OpenHistory(cwd)
	if err != nil {
		return err
	}
	report, errs := check.RunHistory(h)
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
// rejected ones apart, for scripts that would otherwise read .tree/events
// (ADR 0006). A shape change bumps format.
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
	}{Format: 1, Events: []event.Event{}, Rejections: []rejection{}}
	// In a store's folder, the events of that store; where discovery finds
	// no store, every event of the history.
	cur := r.hist.Current
	for _, ev := range r.full.Applied {
		if cur == "" || ev.Store == cur {
			out.Events = append(out.Events, ev)
		}
	}
	for _, x := range r.full.Rejections {
		if cur == "" || x.Store == cur {
			out.Rejections = append(out.Rejections, rejection{x.EventID, x.Reason})
		}
	}
	return json.NewEncoder(stdout).Encode(out)
}
