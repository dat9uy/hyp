// Package store owns the .tree directory: discovery, init, reading the
// per-actor JSONL logs and the single append path that validates first.
package store

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/replay"
)

// ErrNoStore means no .tree/events directory was found above the start dir.
var ErrNoStore = errors.New("no .tree store found; run 'hyp init'")

var actorRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)

// LineError is a line of an events file that could not be parsed.
type LineError struct {
	File string
	Line int
	Err  error
}

// Dir returns the .tree directory under root.
func Dir(root string) string { return filepath.Join(root, ".tree") }

func eventsDir(root string) string { return filepath.Join(root, ".tree", "events") }

// FindRoot walks up from start to the directory that contains .tree/events.
// The walk stops at the repository root, the first directory that holds .git
// (a directory, or a file in a worktree), so a store above the repository is
// never found. With no .git at or above start it walks to the filesystem
// root.
func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if fi, err := os.Stat(eventsDir(dir)); err == nil && fi.IsDir() {
			return dir, nil
		}
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return "", ErrNoStore
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNoStore
		}
		dir = parent
	}
}

// GitAttributes is the content of .tree/.gitattributes: git merges an events
// file of two branches by keeping every line of both, and replay orders the
// lines by id whatever their order in the file.
const GitAttributes = "events/*.jsonl merge=union\n"

// Init creates .tree/events, .tree/evidence, .tree/.gitignore and
// .tree/.gitattributes in dir. It refuses, and creates nothing, when store
// discovery from dir finds a store, at dir or above it, since a command run
// there would then read that store and never the new one. The .gitignore
// names index.sqlite because older hyp binaries write one (ADR 0006).
func Init(dir string) error {
	root, err := FindRoot(dir)
	if err == nil {
		return herr.Refused("store discovery from %s finds the store at %s; run hyp init where it finds none", dir, root)
	}
	if !errors.Is(err, ErrNoStore) {
		return err
	}
	for _, sub := range []string{"events", "evidence"} {
		if err := os.MkdirAll(filepath.Join(Dir(dir), sub), 0o755); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(Dir(dir), ".gitattributes"), []byte(GitAttributes), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(Dir(dir), ".gitignore"), []byte("index.sqlite\nindex.sqlite-*\n"), 0o644)
}

// ReadAll parses every .tree/events/*.jsonl file, keyed by actor (the base
// name without .jsonl). Unparseable lines are returned as LineErrors.
func ReadAll(root string) (map[string][]event.Event, []LineError, error) {
	out, _, lineErrs, err := readLines(root)
	return out, lineErrs, err
}

// readLines is ReadAll that also returns each parsed event's line, trimmed,
// at the same actor and index.
func readLines(root string) (map[string][]event.Event, map[string][][]byte, []LineError, error) {
	files, err := filepath.Glob(filepath.Join(eventsDir(root), "*.jsonl"))
	if err != nil {
		return nil, nil, nil, err
	}
	out := map[string][]event.Event{}
	raw := map[string][][]byte{}
	var lineErrs []LineError
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, nil, nil, err
		}
		actor := strings.TrimSuffix(filepath.Base(f), ".jsonl")
		evs, lines, errs := parseFile(filepath.Base(f), data)
		out[actor], raw[actor] = evs, lines
		lineErrs = append(lineErrs, errs...)
	}
	return out, raw, lineErrs, nil
}

func parseFile(name string, data []byte) ([]event.Event, [][]byte, []LineError) {
	var evs []event.Event
	var lines [][]byte
	var errs []LineError
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for n := 1; sc.Scan(); n++ {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		ev, err := event.ParseLine(line)
		if err != nil {
			errs = append(errs, LineError{File: name, Line: n, Err: err})
			continue
		}
		evs = append(evs, ev)
		lines = append(lines, bytes.Clone(line))
	}
	return evs, lines, errs
}

// Events returns the merged event list of the store at root.
func Events(root string) ([]event.Event, []LineError, error) { return Single(root).Events() }

// Load replays the store at root with its current recipe hashes.
func Load(root string) (replay.Projection, []LineError, error) {
	return LoadAsOf(root, "")
}

// LoadAsOf is Load stopped after the event asOf ("" for the whole log),
// with the uncontained Results marked from files.
func LoadAsOf(root, asOf string) (replay.Projection, []LineError, error) {
	h := Single(root)
	evs, lineErrs, err := h.Events()
	if err != nil {
		return replay.Projection{}, nil, err
	}
	return h.Replay(evs, h.Hashes(), asOf), lineErrs, nil
}

var (
	entropyMu sync.Mutex
	entropy   = ulid.Monotonic(rand.Reader, 0)
)

func newID(t time.Time) string {
	entropyMu.Lock()
	defer entropyMu.Unlock()
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// idAfter returns a new id for time t that sorts after last, the greatest id
// the store holds. When the clock reads earlier than last, or the same
// millisecond, the id takes the millisecond after last's, with fresh random
// bits: two clones of one store whose clocks are behind then append
// different ids, which a union merge keeps apart. In the greatest ULID
// millisecond it adds one plus a random 32-bit amount to last's random bits,
// or one when that overflows, which two clones may then share. It returns ""
// when no ULID sorts after last: last is the greatest ULID, or an id of the
// id form above it. last is "" or has the id form (replay.ValidID).
func idAfter(t time.Time, last string) string {
	id := newID(t)
	if id > last {
		return id
	}
	u, err := ulid.ParseStrict(last)
	if err != nil {
		return ""
	}
	if ms := u.Time(); ms < ulid.MaxTime() {
		return newID(ulid.Time(ms + 1))
	}
	var step [4]byte
	rand.Read(step[:]) // never fails since Go 1.24
	n := new(big.Int).SetBytes(u[6:])
	for _, add := range []int64{1 + int64(binary.BigEndian.Uint32(step[:])), 1} {
		next := new(big.Int).Add(n, big.NewInt(add))
		if next.BitLen() <= 80 {
			next.FillBytes(u[6:])
			return u.String()
		}
	}
	return ""
}

// Append validates ev against the log of the store at root and, when it is
// allowed, appends it (History.AppendWith).
func Append(root, actor string, ev event.Event) (event.Event, error) {
	return AppendWith(root, actor, func(*replay.State, replay.Projection) (event.Event, error) { return ev, nil })
}

// AppendWith is History.AppendWith on the history of the one store at root.
func AppendWith(root, actor string, build func(*replay.State, replay.Projection) (event.Event, error)) (event.Event, error) {
	return Single(root).AppendWith(actor, build)
}

// ValidActor reports whether actor matches ^[a-z0-9][a-z0-9-]{0,39}$.
func ValidActor(actor string) bool { return actorRe.MatchString(actor) }
