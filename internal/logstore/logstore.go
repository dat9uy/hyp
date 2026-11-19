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
	File string
	Line int
	Err  error
}

func eventsDir(root string) string  { return filepath.Join(root, ".tree", "events") }
func outputsDir(root string) string { return filepath.Join(root, ".tree", "outputs") }

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

// Events returns every event of the store in merge order, and the lines that
// could not be used. A torn last line is one of them, never an error.
func Events(root string) ([]events.Event, []LineError, error) {
	paths, err := filepath.Glob(filepath.Join(eventsDir(root), "*.jsonl"))
	if err != nil {
		return nil, nil, fault.EnvFault(fault.ReadEvents, err)
	}
	files := map[string][]events.Event{}
	raw := map[string][][]byte{}
	var lineErrs []LineError
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, nil, fault.EnvFault(fault.ReadEvents, err)
		}
		name := filepath.Base(p)
		evs, lines, errs := parseFile(name, data)
		actor := strings.TrimSuffix(name, ".jsonl")
		files[actor], raw[actor] = evs, lines
		lineErrs = append(lineErrs, errs...)
	}
	order := mergeOrder(files, func(a, b At) int { return bytes.Compare(raw[a.Actor][a.Index], raw[b.Actor][b.Index]) })
	// A union merge can keep one line twice; the copy holds no second event.
	out := make([]events.Event, 0, len(order))
	var prev []byte
	for i, at := range order {
		line := raw[at.Actor][at.Index]
		if i > 0 && bytes.Equal(line, prev) {
			continue
		}
		prev = line
		out = append(out, files[at.Actor][at.Index])
	}
	return out, lineErrs, nil
}

// lock takes the store's directory lock and returns the function that
// releases it. It refuses a root with no store. The lock is on the events
// directory, not one actor file, so that every writer checks against a log
// that already holds all other writers' events.
func lock(root string) (unlock func(), err error) {
	if err := Find(root); err != nil {
		return nil, err
	}
	dir, err := os.Open(eventsDir(root))
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
	unlock, err := lock(root)
	if err != nil {
		return err
	}
	defer unlock()
	evs, _, err := Events(root)
	if err != nil {
		return err
	}
	return f(evs)
}

// Append takes the store's lock, reads every event, calls build with them,
// and writes the event build returns as one fsynced line to
// .tree/events/<actor>.jsonl. V, ID, TS and Actor are filled in here, and Node
// too when build left it empty: a new Goal's goal.add names itself. An error
// from build is returned unchanged and nothing is written, not even the file.
func Append(root, actor string, build func([]events.Event) (events.Event, error)) (events.Event, error) {
	if !ValidActor(actor) {
		return events.Event{}, fault.BadArg(fault.BadActor, actor)
	}
	unlock, err := lock(root)
	if err != nil {
		return events.Event{}, err
	}
	defer unlock()

	evs, _, err := Events(root)
	if err != nil {
		return events.Event{}, err
	}
	ev, err := build(evs)
	if err != nil {
		return ev, err
	}
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
	path := filepath.Join(eventsDir(root), actor+".jsonl")
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

// OutputPath is where the output named hex is stored.
func OutputPath(root, hex string) string { return filepath.Join(outputsDir(root), hex) }

// PutOutput stores b under .tree/outputs/<sha256> and returns the hash. Equal
// bytes are stored once. It is called before the event that names the hash is
// appended, so a crash leaves a file no event names, which is harmless.
func PutOutput(root string, b []byte) (string, error) {
	sum := sha256.Sum256(b)
	name := hex.EncodeToString(sum[:])
	path := OutputPath(root, name)
	if _, err := os.Stat(path); err == nil {
		return name, nil
	}
	// git keeps no empty folder, so a clone of a store can lack outputs/.
	if err := os.MkdirAll(outputsDir(root), 0o755); err != nil {
		return "", fault.EnvFault(fault.StoreOutput, err)
	}
	tmp, err := os.CreateTemp(outputsDir(root), "tmp-*")
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
// is positive, and the size of the whole file.
func ReadOutput(root, hex string, max int) (data []byte, size int64, err error) {
	path := OutputPath(root, hex)
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
