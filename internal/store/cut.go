package store

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// ErrNoEvent means a cut names an event the store does not hold.
var ErrNoEvent = errors.New("no such event")

// Cut is a store as it was at one event: the events replay applies up to and
// including it, in merge order, and the evidence of the Experiments among
// them. `hyp status` of the written store equals `hyp status --as-of AT` of
// the source.
type Cut struct {
	Src, At string
	Events  int
	lines   map[string][][]byte // kept lines per actor (events file without .jsonl), in merge order, so by id
	exps    map[string]bool     // evidence folders (E-<last 6 of the id>) of kept Experiments
}

// CutAt reads the store at src and keeps every event the merge puts at or
// before the event at. A line that does not parse is refused, since it has no
// place in the merge order.
func CutAt(src, at string) (*Cut, error) {
	if fi, err := os.Stat(eventsDir(src)); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("%s: %w", src, ErrNoStore)
	}
	files, err := filepath.Glob(filepath.Join(eventsDir(src), "*.jsonl"))
	if err != nil {
		return nil, err
	}
	evs := map[string][]event.Event{}
	raw := map[string][][]byte{}
	for _, f := range files {
		name := filepath.Base(f)
		// Keyed by actor, as ReadAll keys them: the merge breaks id ties by
		// this key, and "a" sorts before "a-b" where "a.jsonl" does not.
		actor := strings.TrimSuffix(name, ".jsonl")
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		sc := bufio.NewScanner(bytes.NewReader(data))
		sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
		for n := 1; sc.Scan(); n++ {
			line := sc.Bytes()
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}
			ev, err := event.ParseLine(bytes.TrimSpace(line))
			if err != nil {
				return nil, fmt.Errorf("%s:%d: %v; run hyp check in %s", name, n, err, src)
			}
			// Merge never reads Actor, so it carries the file's actor here.
			ev.Actor = actor
			evs[actor] = append(evs[actor], ev)
			raw[actor] = append(raw[actor], bytes.Clone(line))
		}
	}
	c := &Cut{Src: src, At: at, lines: map[string][][]byte{}, exps: map[string]bool{}}
	for _, pos := range replay.MergeOrder(evs) {
		ev := evs[pos.Actor][pos.Index]
		c.lines[ev.Actor] = append(c.lines[ev.Actor], raw[pos.Actor][pos.Index])
		c.Events++
		if ev.Type == event.TypeNodeAdd {
			var p struct{ Kind string }
			if json.Unmarshal(ev.Payload, &p) == nil && p.Kind == event.KindExperiment && len(ev.Node) >= 6 {
				c.exps["E-"+ev.Node[len(ev.Node)-6:]] = true
			}
		}
		// Replay stops after the first event with this id; so does the cut.
		if ev.ID == at {
			return c, nil
		}
	}
	return nil, fmt.Errorf("%w %s in %s", ErrNoEvent, at, src)
}

// Write makes dest/.tree/events with one file per actor that has a kept
// event, each kept line as read and the lines in merge order, so a store
// whose files hold the same lines in another order writes the same bytes; and dest/.tree/evidence without the
// E-* folders of Experiments added after the cut, which would leak the
// future into a replay. A symlink under the evidence is refused. dest/.tree must not exist. The index is not
// written; hyp rebuilds it.
func (c *Cut) Write(dest string) error {
	if _, err := os.Lstat(Dir(dest)); err == nil {
		return fmt.Errorf("%s exists", Dir(dest))
	}
	if err := os.MkdirAll(eventsDir(dest), 0o755); err != nil {
		return err
	}
	for actor, lines := range c.lines {
		var b bytes.Buffer
		for _, l := range lines {
			b.Write(l)
			b.WriteByte('\n')
		}
		if err := os.WriteFile(filepath.Join(eventsDir(dest), actor+".jsonl"), b.Bytes(), 0o644); err != nil {
			return err
		}
	}
	ev := filepath.Join(Dir(c.Src), "evidence")
	if fi, err := os.Stat(ev); err != nil || !fi.IsDir() {
		return nil
	}
	out := filepath.Join(Dir(dest), "evidence")
	return filepath.WalkDir(ev, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(ev, p)
		if err != nil {
			return err
		}
		if filepath.Dir(rel) == "." && strings.HasPrefix(rel, "E-") && !c.exps[rel] {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		target := filepath.Join(out, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("%s: not a regular file", p)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}
