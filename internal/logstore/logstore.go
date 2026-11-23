// Package logstore owns .tree/ at the git top level: creating it, the
// per-actor events files, the one append path, merge order and the outputs
// named by their sha256.
package logstore

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/oklog/ulid/v2"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
)

// GitAttributes is the content of .tree/.gitattributes: git merges an events
// file of two branches by keeping every line of both, and replay orders the
// lines by id whatever their order in the file.
const GitAttributes = "events/*.jsonl merge=union\n"

// MaxWhy is the longest why, in runes.
const MaxWhy = 600

var (
	actorRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)
	ulidRe  = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)
)

// LineError is a line of an events file that could not be used: the file's
// base name, the line number (0 when the file could not be read to its end)
// and why.
type LineError struct {
	File  string
	Line  int
	Err   error
	Trace bool // the file is in the clone's trace, not in .tree/events/
}

// A folder of events is a folder that holds events/<actor>.jsonl and
// outputs/<sha256>: the store's .tree/ in the repository, or the trace
// folder of a clone (trace.go). The functions that take a root read the store.
func treeDir(root string) string    { return filepath.Join(root, ".tree") }
func eventsDir(root string) string  { return eventsIn(treeDir(root)) }
func outputsDir(root string) string { return outputsIn(treeDir(root)) }
func eventsIn(base string) string   { return filepath.Join(base, "events") }
func outputsIn(base string) string  { return filepath.Join(base, "outputs") }

// ValidActor reports whether actor matches ^[a-z0-9][a-z0-9-]{0,39}$.
func ValidActor(actor string) bool { return actorRe.MatchString(actor) }

// TopLevel returns the git top level of the repository that holds start. Not
// a repository, or an object format other than sha1, is an environment error.
func TopLevel(start string) (string, error) {
	out, err := gitread.Git(start, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fault.EnvFault(fault.NotInRepo, start, err)
	}
	root := strings.TrimSpace(string(out))
	format, err := gitread.Git(root, "rev-parse", "--show-object-format")
	if err != nil {
		return "", err
	}
	if f := strings.TrimSpace(string(format)); f != "sha1" {
		return "", fault.EnvFault(fault.ObjectFormat, f)
	}
	return root, nil
}

// Create makes .tree/events, .tree/outputs and .tree/.gitattributes under
// root. It changes nothing that exists.
func Create(root string) error {
	for _, d := range []string{eventsDir(root), outputsDir(root)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fault.EnvFault(fault.CreateStore, err)
		}
	}
	attrs := filepath.Join(root, ".tree", ".gitattributes")
	if _, err := os.Stat(attrs); err == nil {
		return nil
	}
	if err := os.WriteFile(attrs, []byte(GitAttributes), 0o644); err != nil {
		return fault.EnvFault(fault.CreateStore, err)
	}
	return nil
}

// Find reports whether the store exists under root.
func Find(root string) error {
	if fi, err := os.Stat(eventsDir(root)); err != nil || !fi.IsDir() {
		return fault.Refuse(fault.NoGuard, fault.NoStore)
	}
	return nil
}

// ValidID reports whether id has the form of a ULID.
func ValidID(id string) bool { return ulidRe.MatchString(id) }

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
// id form above it. last is "" or has the id form (ValidID).
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

// At names one event of the files mergeOrder reads: its actor and its index
// in that actor's file.
type At struct {
	Actor string
	Index int
}

// mergeOrder orders the events of per-actor files by id, then by actor (the
// file name without .jsonl), whatever the line order inside a file: a file
// that git merged with the union driver holds both branches' lines in any
// order, and replays the same. tie, which must not be nil, orders two events
// with the same id and actor; the store passes one that compares their line
// bytes, so file order decides nothing. It returns positions in files.
func mergeOrder(files map[string][]events.Event, tie func(a, b At) int) []At {
	var out []At
	for name, evs := range files {
		for i := range evs {
			out = append(out, At{name, i})
		}
	}
	slices.SortFunc(out, func(a, b At) int {
		if c := strings.Compare(files[a.Actor][a.Index].ID, files[b.Actor][b.Index].ID); c != 0 {
			return c
		}
		if c := strings.Compare(a.Actor, b.Actor); c != 0 {
			return c
		}
		return tie(a, b)
	})
	return out
}

// parseFile reads the events of one file, name being the file's base name.
// A line that does not parse, whose id is not a ULID or whose actor is not
// the file's actor is returned as a LineError and skipped.
func parseFile(name string, data []byte) ([]events.Event, [][]byte, []LineError) {
	actor := strings.TrimSuffix(name, ".jsonl")
	var evs []events.Event
	var lines [][]byte
	var errs []LineError
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for n := 1; sc.Scan(); n++ {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		ev, err := events.ParseLine(line)
		switch {
		case err != nil:
		case !ValidID(ev.ID):
			err = fault.Refuse(fault.NoGuard, fault.IDNotULID, ev.ID)
		case ev.Actor != actor:
			err = fault.Refuse(fault.NoGuard, fault.ActorNotFile, ev.Actor, actor)
		}
		if err != nil {
			errs = append(errs, LineError{File: name, Line: n, Err: err})
			continue
		}
		evs = append(evs, ev)
		lines = append(lines, bytes.Clone(line))
	}
	if err := sc.Err(); err != nil {
		errs = append(errs, LineError{File: name, Err: err})
	}
	return evs, lines, errs
}

// Events returns every event of the store and of the clone's trace in merge
// order, and the lines that could not be used. A torn last line is one of
// them, never an error. The kinds that change a Goal's state are in git; the
// rest are in the trace (events.Kinds), so a fresh clone replays every Goal
// state from git alone.
func Events(root string) ([]events.Event, []LineError, error) {
	tr, err := TraceOf(root)
	if err != nil {
		return nil, nil, err
	}
	return eventsOf(treeDir(root), tr.Dir)
}

// StoreEvents returns the events of the store's .tree/events/ alone, in merge
// order.
func StoreEvents(root string) ([]events.Event, []LineError, error) {
	return eventsOf(treeDir(root))
}

// eventsOf reads the events of the folders of events bases, in merge order.
// A folder with no events folder holds no event. The first base is the
// store's; a line error of any other is marked as the trace's.
func eventsOf(bases ...string) ([]events.Event, []LineError, error) {
	type line struct {
		ev    events.Event
		actor string
		raw   []byte
	}
	var all []line
	var lineErrs []LineError
	for i, base := range bases {
		paths, err := filepath.Glob(filepath.Join(eventsIn(base), "*.jsonl"))
		if err != nil {
			return nil, nil, fault.EnvFault(fault.ReadEvents, err)
		}
		for _, p := range paths {
			data, err := os.ReadFile(p)
			if err != nil {
				return nil, nil, fault.EnvFault(fault.ReadEvents, err)
			}
			name := filepath.Base(p)
			evs, lines, errs := parseFile(name, data)
			for j := range errs {
				errs[j].Trace = i > 0
			}
			lineErrs = append(lineErrs, errs...)
			actor := strings.TrimSuffix(name, ".jsonl")
			for j := range evs {
				all = append(all, line{evs[j], actor, lines[j]})
			}
		}
	}
	// Merge order: by id, then by actor, then by line bytes, so file order
	// decides nothing.
	slices.SortFunc(all, func(a, b line) int {
		if c := strings.Compare(a.ev.ID, b.ev.ID); c != 0 {
			return c
		}
		if c := strings.Compare(a.actor, b.actor); c != 0 {
			return c
		}
		return bytes.Compare(a.raw, b.raw)
	})
	// A union merge can keep one line twice; the copy holds no second event.
	out := make([]events.Event, 0, len(all))
	var prev []byte
	for i, l := range all {
		if i > 0 && bytes.Equal(l.raw, prev) {
			continue
		}
		prev = l.raw
		out = append(out, l.ev)
	}
	return out, lineErrs, nil
}

// lock takes the store's directory lock, then the lock of the clone's
// trace, and returns the function that releases both and the trace. It
// refuses a root with no store. The lock is on the events directory, not one
// actor file, so that every writer checks against a log that already holds
// all other writers' events, in git and in the trace.
func lock(root string) (unlock func(), tr Trace, err error) {
	if err := Find(root); err != nil {
		return nil, Trace{}, err
	}
	if tr, err = TraceOf(root); err != nil {
		return nil, Trace{}, err
	}
	if err := tr.create(); err != nil {
		return nil, Trace{}, err
	}
	unlockStore, err := lockDir(treeDir(root))
	if err != nil {
		return nil, Trace{}, err
	}
	unlockTrace, err := lockDir(tr.Dir)
	if err != nil {
		unlockStore()
		return nil, Trace{}, err
	}
	return func() { unlockTrace(); unlockStore() }, tr, nil
}

// lockDir takes the directory lock of the events folder of base.
func lockDir(base string) (unlock func(), err error) {
	dir, err := os.Open(eventsIn(base))
	if err != nil {
		return nil, fault.EnvFault(fault.OpenStore, err)
	}
	if err := syscall.Flock(int(dir.Fd()), syscall.LOCK_EX); err != nil {
		dir.Close()
		return nil, fault.EnvFault(fault.LockStore, err)
	}
	return func() {
		syscall.Flock(int(dir.Fd()), syscall.LOCK_UN)
		dir.Close()
	}, nil
}

// WithLock takes the same directory lock as Append, reads every event, calls
// f with them and releases the lock. It writes nothing: an act uses it to
// check its guards before it runs a Probe, which must not hold the lock.
func WithLock(root string, f func([]events.Event) error) error {
	unlock, tr, err := lock(root)
	if err != nil {
		return err
	}
	defer unlock()
	evs, _, err := eventsOf(treeDir(root), tr.Dir)
	if err != nil {
		return err
	}
	return f(evs)
}

// Append takes the store's lock and the trace's, reads every event of both,
// calls build with them, and writes the event build returns as one fsynced
// line to <actor>.jsonl where its kind's entry stores it: .tree/events/ for a
// kind that changes a Goal's state, else the clone's trace (events.Kinds).
// V, ID, TS and Actor are filled in here, and Node too when build left it
// empty: a new Goal's goal.add names itself. The id sorts after every event
// of git and of the trace. An error from build is returned unchanged and
// nothing is written, not even the file.
func Append(root, actor string, build func([]events.Event) (events.Event, error)) (events.Event, error) {
	if !ValidActor(actor) {
		return events.Event{}, fault.BadArg(fault.BadActor, actor)
	}
	unlock, tr, err := lock(root)
	if err != nil {
		return events.Event{}, err
	}
	defer unlock()
	evs, _, err := eventsOf(treeDir(root), tr.Dir)
	if err != nil {
		return events.Event{}, err
	}
	ev, err := build(evs)
	if err != nil {
		return ev, err
	}
	k, ok := events.KindOf(ev.Type)
	if !ok {
		return ev, fault.EnvFault(fault.EventHome, ev.Type, "")
	}
	base := treeDir(root)
	if k.Home() == events.HomeTrace {
		base = tr.Dir
	}
	return writeEvent(base, actor, ev, evs)
}

// appendLocked reads the events of the folder of events base, whose lock the
// caller holds, and writes the event build returns there. home is where base
// is: an event whose kind's entry names another home is not written
// (events.Kinds).
func appendLocked(base, actor, home string, build func([]events.Event) (events.Event, error)) (events.Event, error) {
	evs, _, err := eventsOf(base)
	if err != nil {
		return events.Event{}, err
	}
	ev, err := build(evs)
	if err != nil {
		return ev, err
	}
	if k, ok := events.KindOf(ev.Type); !ok || k.Home() != home {
		return ev, fault.EnvFault(fault.EventHome, ev.Type, home)
	}
	return writeEvent(base, actor, ev, evs)
}

// writeEvent writes ev to the actor's file of the folder of events base,
// with an id after the greatest id of evs.
func writeEvent(base, actor string, ev events.Event, evs []events.Event) (events.Event, error) {
	if n := utf8.RuneCountInString(ev.Why); n > MaxWhy {
		return ev, fault.BadArg(fault.WhyTooLong, n, MaxWhy)
	}

	now := time.Now().UTC()
	// The newest id the store holds is the greatest id of the id form.
	last := ""
	if len(evs) > 0 {
		last = evs[len(evs)-1].ID
	}
	ev.V = events.Version
	if ev.ID = idAfter(now, last); ev.ID == "" {
		return ev, fault.Refuse(fault.NoGuard, fault.NoIDAfter, last)
	}
	ev.TS = now.Format(time.RFC3339Nano)
	ev.Actor = actor
	if ev.Node == "" {
		ev.Node = ev.ID
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return ev, fault.EnvFault(fault.EncodeEvent, ev.Type, err)
	}
	// Open only after build returned, so a refused act leaves no file behind.
	path := filepath.Join(eventsIn(base), actor+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o644)
	if err != nil {
		return ev, fault.EnvFault(fault.FileOpen, path, err)
	}
	defer f.Close()
	// A torn last line must stay alone: start a new line when the file does
	// not end in one.
	prefix := ""
	if fi, err := f.Stat(); err != nil {
		return ev, fault.EnvFault(fault.FileStat, path, err)
	} else if fi.Size() > 0 {
		var last [1]byte
		if _, err := f.ReadAt(last[:], fi.Size()-1); err != nil {
			return ev, fault.EnvFault(fault.FileRead, path, err)
		}
		if last[0] != '\n' {
			prefix = "\n"
		}
	}
	if _, err := f.Write([]byte(prefix + string(line) + "\n")); err != nil {
		return ev, fault.EnvFault(fault.FileWrite, path, err)
	}
	if err := f.Sync(); err != nil {
		return ev, fault.EnvFault(fault.FileSync, path, err)
	}
	return ev, nil
}

// OutputsDir is the store's outputs folder, which holds the outputs that
// events written before the trace kept them name.
func OutputsDir(root string) string { return outputsDir(root) }

// Hash is the name an output with bytes b is stored under: its sha256.
func Hash(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// PutOutput stores b in the clone's trace by its sha256 and returns the hash.
// Equal bytes are stored once. It is called before the event that names the
// hash is appended, so a crash leaves a file no event names, which is
// harmless. hyp writes no output under .tree/outputs/.
func PutOutput(root string, b []byte) (string, error) {
	tr, err := TraceOf(root)
	if err != nil {
		return "", err
	}
	return tr.PutOutput(b)
}

// putOutputIn is PutOutput for the outputs folder dir.
func putOutputIn(dir string, b []byte) (string, error) {
	name := Hash(b)
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		return name, nil
	}
	// git keeps no empty folder, so a clone of a store can lack outputs/.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fault.EnvFault(fault.StoreOutput, err)
	}
	tmp, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return "", fault.EnvFault(fault.StoreOutput, err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return "", fault.EnvFault(fault.StoreOutput, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return "", fault.EnvFault(fault.StoreOutput, err)
	}
	if err := tmp.Close(); err != nil {
		return "", fault.EnvFault(fault.StoreOutput, err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return "", fault.EnvFault(fault.StoreOutput, err)
	}
	return name, nil
}

// ReadOutput reads the stored output named hex: at most max bytes when max
// is positive, and the size of the whole file. It reads .tree/outputs/, where
// an output named by an event written before the trace is, and else the
// clone's trace.
func ReadOutput(root, hex string, max int) (data []byte, size int64, err error) {
	data, size, err = ReadOutputIn(outputsDir(root), hex, max)
	if err == nil {
		return data, size, nil
	}
	tr, terr := TraceOf(root)
	if terr != nil {
		return nil, 0, err
	}
	if _, serr := os.Stat(filepath.Join(tr.Outputs(), hex)); serr != nil {
		return nil, 0, err
	}
	return ReadOutputIn(tr.Outputs(), hex, max)
}

// ReadOutputIn is ReadOutput for the outputs folder dir.
func ReadOutputIn(dir, hex string, max int) (data []byte, size int64, err error) {
	path := filepath.Join(dir, hex)
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fault.EnvFault(fault.FileOpen, path, err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, 0, fault.EnvFault(fault.FileStat, path, err)
	}
	var r io.Reader = f
	if max > 0 {
		r = io.LimitReader(f, int64(max))
	}
	data, err = io.ReadAll(r)
	if err != nil {
		return nil, 0, fault.EnvFault(fault.FileRead, path, err)
	}
	return data, fi.Size(), nil
}
