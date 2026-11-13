package store

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/dat9uy/hyp/internal/contain"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/replay"
)

// History is every store of a repository, read as one log: every event of
// every store, ordered by id. Stores stay in their folders and no event
// moves; the store an event belongs to is the folder of its events file.
type History struct {
	// Root is the repository root, the first directory at or above the
	// current one that holds .git. With no .git there, or when the store
	// found from the current directory adds nothing to the repository's
	// history, it is that store's own root and the history is that store.
	Root string
	// Stores are the folders of the history's stores relative to Root, in
	// slash form, "." for Root itself, sorted.
	Stores []string
	// Current is the store discovery finds from the current directory, as
	// one of Stores, or "" when it finds none.
	Current string
}

// Dir returns the folder of store s.
func (h *History) Dir(s string) string { return filepath.Join(h.Root, filepath.FromSlash(s)) }

// Single is the history of the one store at root.
func Single(root string) *History { return &History{Root: root, Stores: []string{"."}, Current: "."} }

// OpenHistory finds the history from start. Inside a repository it is
// every store the repository holds (repoStores), and the store discovery
// finds from start when git tracks none of its events files yet but it
// would add to the history once tracked: a store hyp init just made reads
// and cites the history from its first event. Outside a repository, or
// when the store found from start would add nothing to the history even
// if tracked, it is that store alone. With no store either way it returns
// ErrNoStore.
func OpenHistory(start string) (*History, error) {
	return openHistory(start, false)
}

// OpenRepoHistory is the history of the repository that holds start, from
// any folder in it: as OpenHistory, except that when the store found from
// start adds nothing to the history it is not read and Current is "".
// Outside a repository, or when git cannot list its files, it is
// OpenHistory.
func OpenRepoHistory(start string) (*History, error) {
	return openHistory(start, true)
}

func openHistory(start string, repoWide bool) (*History, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	cur, err := FindRoot(abs)
	if err != nil && !errors.Is(err, ErrNoStore) {
		return nil, err
	}
	repo, ok := recipe.RepoRoot(abs)
	if !ok {
		if cur == "" {
			return nil, ErrNoStore
		}
		return Single(cur), nil
	}
	stores, err := repoStores(repo)
	if err != nil {
		// Without a git that reads the repository, which stores git tracks
		// is not known: the history is the store found from start.
		if cur == "" {
			return nil, fmt.Errorf("%w (%v)", ErrNoStore, err)
		}
		return Single(cur), nil
	}
	h := &History{Root: repo, Stores: stores}
	if cur != "" {
		rel, err := filepath.Rel(repo, cur)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		switch {
		case slices.Contains(stores, rel):
			h.Current = rel
		case rel != "." && !strings.HasPrefix(rel, "../") && rel != ".." && storeOfHistory(repo, rel):
			// A store git does not track yet, read from its own folder. A
			// store at the repository root never joins this way: below it
			// lie the stores it would read, as in a clone a store cut wrote.
			h.Stores = append(h.Stores, rel)
			slices.Sort(h.Stores)
			h.Current = rel
		case !repoWide:
			return Single(cur), nil
		}
	}
	if len(h.Stores) == 0 {
		return nil, ErrNoStore
	}
	return h, nil
}

// RepoStores lists the stores of the history of the repository at repo,
// relative to it (repoStores).
func RepoStores(repo string) ([]string, error) { return repoStores(repo) }

// repoStores lists the stores of the repository at repo: each folder whose
// .tree/events holds an events file git tracks, except a store under
// another store's .tree/, under a testdata/ directory, under a bundle's
// known-answers/ (a directory beside a recipe.md), or under a directory
// below repo that holds its own .git.
func repoStores(repo string) ([]string, error) {
	cmd := exec.Command("git", "-C", repo, "ls-files", "-z", "--cached", "--", ":(glob)**/.tree/events/*.jsonl")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files in %s: %v", repo, err)
	}
	var stores []string
	for _, p := range strings.Split(string(out), "\x00") {
		if p == "" {
			continue
		}
		i := strings.LastIndex(p, ".tree/events/")
		if i < 0 || (i > 0 && p[i-1] != '/') {
			continue
		}
		dir := strings.TrimSuffix(p[:i], "/")
		if dir == "" {
			dir = "."
		}
		if slices.Contains(stores, dir) || !storeOfHistory(repo, dir) {
			continue
		}
		stores = append(stores, dir)
	}
	slices.Sort(stores)
	return stores, nil
}

// storeOfHistory reports whether the store at rel adds to the history of
// the repository at repo (repoStores).
func storeOfHistory(repo, rel string) bool {
	if fi, err := os.Stat(filepath.Join(repo, filepath.FromSlash(rel), ".tree", "events")); err != nil || !fi.IsDir() {
		return false
	}
	if rel == "." {
		return true
	}
	parts := strings.Split(rel, "/")
	for i, part := range parts {
		if part == ".tree" || part == "testdata" {
			return false
		}
		dir := filepath.Join(repo, filepath.FromSlash(strings.Join(parts[:i+1], "/")))
		if part == "known-answers" {
			if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "recipe.md")); err == nil {
				return false
			}
		}
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return false
		}
	}
	return true
}

// historyFiles is every parsed line of a history, keyed by file.
type historyFiles struct {
	evs  map[replay.File][]event.Event
	raw  map[replay.File][][]byte // each line as read
	trim map[replay.File][][]byte // each line trimmed
}

// read parses every events file of every store. A line that does not parse
// is a LineError, or, with strict, the error.
func (h *History) read(strict bool) (*historyFiles, []LineError, error) {
	hf := &historyFiles{evs: map[replay.File][]event.Event{}, raw: map[replay.File][][]byte{}, trim: map[replay.File][][]byte{}}
	var lineErrs []LineError
	for _, s := range h.Stores {
		files, err := filepath.Glob(filepath.Join(eventsDir(h.Dir(s)), "*.jsonl"))
		if err != nil {
			return nil, nil, err
		}
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				return nil, nil, err
			}
			name := filepath.Base(f)
			shown := name
			if s != "." {
				shown = s + "/.tree/events/" + name
			}
			key := replay.File{Store: s, Actor: strings.TrimSuffix(name, ".jsonl")}
			sc := bufio.NewScanner(bytes.NewReader(data))
			sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
			for n := 1; sc.Scan(); n++ {
				line := bytes.TrimSpace(sc.Bytes())
				if len(line) == 0 {
					continue
				}
				ev, err := event.ParseLine(line)
				if err != nil {
					if strict {
						return nil, nil, fmt.Errorf("%s:%d: %v; run hyp check in %s", shown, n, err, h.Dir(s))
					}
					lineErrs = append(lineErrs, LineError{File: shown, Line: n, Err: err})
					continue
				}
				ev.Store = s
				hf.evs[key] = append(hf.evs[key], ev)
				hf.raw[key] = append(hf.raw[key], bytes.Clone(sc.Bytes()))
				hf.trim[key] = append(hf.trim[key], bytes.Clone(line))
			}
		}
	}
	return hf, lineErrs, nil
}

// order is the history's merge order: by id, then actor, then the store's
// path, then the line's bytes, trimmed and then as read, so the order of
// the lines in a file decides nothing.
func (hf *historyFiles) order() []replay.At {
	return replay.MergeOrder(hf.evs, func(a, b replay.At) int {
		fa, fb := replay.File{Store: a.Store, Actor: a.Actor}, replay.File{Store: b.Store, Actor: b.Actor}
		if c := bytes.Compare(hf.trim[fa][a.Index], hf.trim[fb][b.Index]); c != 0 {
			return c
		}
		return bytes.Compare(hf.raw[fa][a.Index], hf.raw[fb][b.Index])
	})
}

// Events returns the history's events in merge order, each with its Store.
func (h *History) Events() ([]event.Event, []LineError, error) {
	hf, lineErrs, err := h.read(false)
	if err != nil {
		return nil, nil, err
	}
	order := hf.order()
	out := make([]event.Event, len(order))
	for i, at := range order {
		out[i] = hf.evs[replay.File{Store: at.Store, Actor: at.Actor}][at.Index]
	}
	return out, lineErrs, nil
}

// Hashes returns each store's current recipe hashes.
func (h *History) Hashes() replay.Hashes {
	by := map[string]map[string]string{}
	for _, s := range h.Stores {
		by[s] = recipe.Hashes(h.Dir(s))
	}
	return replay.Hashes{ByStore: by}
}

// Replay replays evs, stopped after asOf ("" for the whole log), with each
// store's recipe hashes and the uncontained Results marked from each
// store's files.
func (h *History) Replay(evs []event.Event, hashes replay.Hashes, asOf string) replay.Projection {
	p := replay.ReplayHistory(evs, hashes, asOf)
	h.Mark(&p)
	return p
}

// Mark marks the uncontained Results of p, each from its own store's files.
func (h *History) Mark(p *replay.Projection) {
	roots := map[string]string{}
	for _, s := range h.Stores {
		roots[s] = h.Dir(s)
	}
	contain.MarkStores(roots, p)
}

// Load replays the whole history.
func (h *History) Load() (replay.Projection, []LineError, error) {
	evs, lineErrs, err := h.Events()
	if err != nil {
		return replay.Projection{}, nil, err
	}
	return h.Replay(evs, h.Hashes(), ""), lineErrs, nil
}

// StoreDir returns the folder of the store that holds node id in p, or ""
// when p holds no such node.
func (h *History) StoreDir(p replay.Projection, id string) string {
	if n := p.Nodes[id]; n != nil {
		return h.Dir(n.Store)
	}
	return ""
}

// lock takes the history's write lock on its root's directory: the
// repository root, or a store read alone, whose root is its own folder.
// Every writer of the history then
// validates against a log that holds every other writer's events, and
// takes an id after all of them. unlock releases it.
func (h *History) lock() (unlock func(), err error) {
	f, err := os.Open(h.Root)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return func() { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }, nil
}

// target is the store event ev is appended to: the current store when
// discovery finds one, and otherwise the store that holds its node (a node
// add's parent, a Result's Experiment, an answer's keys). An event that
// names a node of another store is then refused by replay's store rule.
func (h *History) target(st *replay.State, ev event.Event) (string, error) {
	storeOf := func(id string) string {
		if n := st.Nodes[id]; n != nil {
			return n.Store
		}
		return ""
	}
	if ev.Type == event.TypeAnswer {
		var p event.Answer
		_ = json.Unmarshal(ev.Payload, &p)
		var stores []string
		for _, items := range [][]event.KeyText{p.Answers, p.Accepts, p.Returns} {
			for _, it := range items {
				if s := storeOf(event.KeyResult(it.Key)); s != "" && !slices.Contains(stores, s) {
					stores = append(stores, s)
				}
			}
		}
		if len(stores) > 1 {
			return "", herr.Refused("an answer's keys lie in two stores, %s and %s; an event never crosses stores, so answer each store's keys in its own hyp answer",
				replay.StoreName(stores[0]), replay.StoreName(stores[1]))
		}
		if h.Current != "" {
			return h.Current, nil
		}
		if len(stores) == 1 {
			return stores[0], nil
		}
		return "", nil
	}
	if h.Current != "" {
		return h.Current, nil
	}
	switch ev.Type {
	case event.TypeNodeAdd:
		var p event.NodeAdd
		_ = json.Unmarshal(ev.Payload, &p)
		if p.Kind == event.KindGoal {
			return "", herr.Refused("hyp add goal needs a store: store discovery finds none from here, and a Goal has no parent whose store holds it; run it in the store's folder")
		}
		return storeOf(p.Parent), nil
	case event.TypeResultRecord:
		var p event.ResultRecord
		_ = json.Unmarshal(ev.Payload, &p)
		return storeOf(p.Parent), nil
	}
	return storeOf(ev.Node), nil
}

// AppendWith validates the event build returns against the whole history
// and, when it is allowed, appends it to the store that holds its node
// (target), as one line of that store's .tree/events/<actor>.jsonl, and
// fsyncs. build receives the replayed history under the write lock. Its id
// sorts after every id the history holds. V, ID, TS, Actor and Store are
// filled in here; for node.add and result.record Node is set to the new id.
// A refused event is never written.
func (h *History) AppendWith(actor string, build func(*replay.State, replay.Projection) (event.Event, error)) (event.Event, error) {
	if !ValidActor(actor) {
		return event.Event{}, fmt.Errorf("invalid actor %q (want %s)", actor, actorRe)
	}
	unlock, err := h.lock()
	if err != nil {
		return event.Event{}, err
	}
	defer unlock()

	evs, _, err := h.Events()
	if err != nil {
		return event.Event{}, err
	}
	st, p := replay.BuildHistory(evs, h.Hashes(), "")
	ev, err := build(st, p)
	if err != nil {
		// An event build drafts and then refuses is refused for crossing
		// stores first, as replay's store rule comes before every other.
		if ev.Type != "" {
			if t, terr := h.target(st, ev); terr == nil && t != "" {
				ev.Store = t
				if cerr := st.CheckStore(ev); cerr != nil {
					return ev, herr.Refused("%v", cerr)
				}
			}
		}
		return ev, err
	}
	target, err := h.target(st, ev)
	if err != nil {
		return ev, err
	}

	now := time.Now().UTC()
	// The newest id the history holds is the greatest id of the id form;
	// a line whose id replay rejects for its form holds no event.
	last := ""
	for i := len(evs) - 1; i >= 0; i-- {
		if replay.ValidID(evs[i].ID) {
			last = evs[i].ID
			break
		}
	}
	ev.V = event.Version
	if ev.ID = idAfter(now, last); ev.ID == "" {
		return ev, herr.Refused("no ULID sorts after %s, the newest id the history holds", last)
	}
	ev.TS = now.Format(time.RFC3339Nano)
	ev.Actor = actor
	ev.Store = target
	if ev.Type == event.TypeNodeAdd || ev.Type == event.TypeResultRecord {
		ev.Node = ev.ID
	}
	if err := replay.Validate(st, ev); err != nil {
		return ev, herr.Refused("%v", err)
	}
	if target == "" {
		return ev, herr.Refused("no store holds the event's node")
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return ev, err
	}
	// Open only after validation, so a refused event leaves no file behind.
	f, err := os.OpenFile(filepath.Join(eventsDir(h.Dir(target)), actor+".jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return ev, err
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return ev, err
	}
	return ev, f.Sync()
}
