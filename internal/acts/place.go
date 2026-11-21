package acts

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/example"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/probe"
	"github.com/dat9uy/hyp/internal/tables"
)

var (
	lineKeyRe     = regexp.MustCompile(`^line-[0-9]+$`)
	exampleNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)
)

// PlaceOpts is the input of place.
type PlaceOpts struct {
	Dir          string // where the command runs
	Actor        string
	Role         string
	Why          string
	Goal         string // the Goal's id
	Key          string // LINE/EXAMPLE, for example line-3/a
	ProbeTimeout time.Duration

	// afterProbe, when set, runs after the Probe and before the second check;
	// a test uses it to land an event between the two checks.
	afterProbe func()
}

// PlaceReport is what place returns.
type PlaceReport struct {
	Goal    string
	Key     string
	Version int
	Hash    string
	Verdict string // H0's verdict
	Exit    int    // the exit code of the run on H0
	State   string // the Goal's state after the placement
}

// checked is what the guards that need no Probe found.
type checked struct {
	goal      *tables.Goal
	line      tables.Line
	hash      string // the example folder's tree hash
	probeHash string // the tree hash of its probe/ folder
	ex        example.Example
}

// Place records an example placement, or refuses with the first failing
// guard of the state machine, section 8: Not landed (always holds), Selected,
// Grill places, Committed, Adds only, then Kind. Everything but Kind is checked
// before the Probe runs, so a refusal that needs no Probe costs no Probe run;
// the Probe runs with the lock released, and every guard is checked again
// under the lock before the event is appended.
func Place(o PlaceOpts) (PlaceReport, error) {
	if err := CheckActor(o.Actor, o.Role); err != nil {
		return PlaceReport{}, err
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return PlaceReport{}, fault.BadArg(fault.WhyTooLong, n, logstore.MaxWhy)
	}
	lineKey, name, ok := strings.Cut(o.Key, "/")
	if !ok || !lineKeyRe.MatchString(lineKey) || !exampleNameRe.MatchString(name) {
		return PlaceReport{}, fault.BadArg(fault.BadExampleKey, o.Key)
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return PlaceReport{}, err
	}

	var first checked
	err = logstore.WithLock(root, func(evs []events.Event) error {
		var err error
		first, err = checkPlace(root, o, evs, lineKey, name)
		return err
	})
	if err != nil {
		return PlaceReport{}, err
	}

	out, err := probe.Run(probe.Opts{Repo: root, Commit: first.goal.Base, ExampleHash: first.hash, Timeout: o.ProbeTimeout})
	if err != nil {
		return PlaceReport{}, err
	}
	verdict := example.Compare(out.Stdout, first.ex)
	// The outputs are stored only once every guard holds, just before the
	// event that names them: a refused place leaves no file behind.
	outHash, errHash := logstore.Hash(out.Stdout), logstore.Hash(out.Stderr)
	if o.afterProbe != nil {
		o.afterProbe()
	}

	var rep PlaceReport
	ev, err := logstore.Append(root, o.Actor, func(evs []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeExamplePlace, Why: o.Why}
		again, err := checkPlace(root, o, evs, lineKey, name)
		if err != nil {
			return ev, err
		}
		switch {
		case again.hash != first.hash:
			return ev, fault.Refuse(fault.GuardCommitted, fault.ChangedWhileProbe, o.Key, first.hash, again.hash)
		case again.goal.Base != first.goal.Base:
			return ev, fault.Refuse(fault.GuardKind, fault.H0MovedWhilePlacing, again.goal.ID, first.goal.Base, again.goal.Base)
		}
		if string(verdict) != tables.Needs(again.line.Kind) {
			return ev, fault.Refuse(fault.GuardKind, fault.WrongVerdictOnH0, KindFacts{
				Line: again.line, Verdict: string(verdict), Out: out,
				Floors: again.ex.Floors, Ceilings: again.ex.Ceilings,
			})
		}
		payload, err := json.Marshal(events.ExamplePlace{
			Role: o.Role, Line: lineKey, Example: name, Hash: again.hash,
			H0: events.Cell{
				Commit: again.goal.Base, RowTree: again.goal.BaseTree, Probe: again.probeHash,
				Output: outHash, Stderr: errHash, Exit: out.Exit, Verdict: string(verdict),
			},
		})
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeExamplePlace, err)
		}
		ev.Node, ev.Payload = again.goal.EventID, payload
		if err := tables.GuardPlace(tables.Replay(evs), ev); err != nil {
			return ev, err
		}
		if err := putOutputs(root, out.Stdout, out.Stderr); err != nil {
			return ev, err
		}
		rep = PlaceReport{Goal: again.goal.ID, Key: o.Key, Hash: again.hash, Verdict: string(verdict), Exit: out.Exit}
		return ev, nil
	})
	if err != nil {
		return PlaceReport{}, err
	}
	evs, _, err := logstore.Events(root)
	if err != nil {
		return PlaceReport{}, err
	}
	if g := tables.Replay(evs).Goals[ev.Node]; g != nil {
		rep.State = g.State
		if cur, ok := g.Current(o.Key); ok {
			rep.Version = cur.Version
		}
	}
	return rep, nil
}

// checkPlace runs the guards that need no Probe, in the order of the state
// machine: Selected, Grill places, Committed, then Adds only. An unknown Goal or
// line is a bad argument.
func checkPlace(root string, o PlaceOpts, evs []events.Event, lineKey, name string) (checked, error) {
	t := tables.Replay(evs)
	g, err := t.Find(o.Goal)
	if err != nil {
		return checked{}, err
	}
	var c checked
	c.goal = g
	for _, l := range g.Lines {
		if l.Key == lineKey {
			c.line = l
		}
	}
	if c.line.Key == "" {
		return checked{}, fault.BadArg(fault.NoLine, g.ID, lineKey)
	}
	if err := tables.GuardNotLanded(g); err != nil {
		return checked{}, err
	}
	if err := tables.GuardNotSelected(g); err != nil {
		return checked{}, err
	}
	if err := tables.GuardGrillPlaces(o.Role); err != nil {
		return checked{}, err
	}
	if c.hash, err = committedHash(root, path.Join("goals", g.ID, lineKey, name)); err != nil {
		return checked{}, err
	}
	if c.ex, c.probeHash, err = readExample(root, c.hash); err != nil {
		return checked{}, err
	}
	key := lineKey + "/" + name
	if err := tables.GuardAddsOnlyTable(g, key, c.hash); err != nil {
		return checked{}, err
	}
	if cur, ok := g.Current(key); ok {
		old, oldProbe, err := readExample(root, cur.Hash)
		if err != nil {
			return checked{}, err
		}
		if err := sameBesidesExample(root, cur.Hash, c.hash); err != nil {
			return checked{}, err
		}
		if err := example.AddsOnly(old, c.ex, oldProbe, c.probeHash); err != nil {
			return checked{}, err
		}
	}
	return c, nil
}

// committedHash is the Committed guard: the folder's tree hash computed from
// disk must equal the hash git holds for it at HEAD. It returns that hash.
func committedHash(root, rel string) (string, error) {
	disk := filepath.Join(root, filepath.FromSlash(rel))
	if fi, err := os.Stat(disk); err != nil || !fi.IsDir() {
		return "", fault.Refuse(fault.GuardCommitted, fault.NotAFolder, rel)
	}
	onDisk, err := gitread.DirTreeHash(disk)
	if err != nil {
		if fault.KindOf(err) == fault.Refusal {
			return "", fault.Refuse(fault.GuardCommitted, fault.FolderOnDisk, err)
		}
		return "", err
	}
	code, out, err := gitread.GitExit(root, "rev-parse", "--verify", "HEAD:"+rel)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fault.Refuse(fault.GuardCommitted, fault.NotCommitted, rel)
	}
	committed := strings.TrimSpace(string(out))
	if committed == onDisk {
		return committed, nil
	}
	diff, err := gitread.FirstDifference(root, disk, committed)
	if err != nil {
		return "", err
	}
	if diff == "" {
		return "", fault.Refuse(fault.GuardCommitted, fault.DiskNotCommitted, rel)
	}
	return "", fault.Refuse(fault.GuardCommitted, fault.DiffersAt, rel, path.Join(rel, diff))
}

// readExample reads and parses the example.txtar of the example folder with
// tree hash hash, and returns the tree hash of its probe/ folder.
func readExample(root, hash string) (example.Example, string, error) {
	if code, _, err := gitread.GitExit(root, "cat-file", "-e", hash); err != nil {
		return example.Example{}, "", err
	} else if code != 0 {
		return example.Example{}, "", fault.EnvFault(fault.ObjectMissing, hash)
	}
	spec := hash + ":example.txtar"
	if code, _, err := gitread.GitExit(root, "cat-file", "-e", spec); err != nil {
		return example.Example{}, "", err
	} else if code != 0 {
		return example.Example{}, "", fault.Refuse(fault.NoGuard, fault.NoExampleTxtar, hash)
	}
	data, err := gitread.Git(root, "cat-file", "blob", spec)
	if err != nil {
		return example.Example{}, "", err
	}
	ex, err := example.Parse(data)
	if err != nil {
		return example.Example{}, "", err
	}
	probeHash, err := gitread.SubtreeHash(root, hash, "probe")
	if err != nil {
		return example.Example{}, "", err
	}
	return ex, probeHash, nil
}

// KindFacts is what a Kind refusal shows: the line, the verdict H0 gave and
// what H0 printed, beside the example's floors and ceilings.
type KindFacts struct {
	Line     tables.Line
	Verdict  string
	Out      probe.Out
	Floors   []example.Floor
	Ceilings []string
}

// sameBesidesExample refuses a new version whose folder differs from the
// current one in any root entry but example.txtar, by mode or hash: the probe/
// folder and every other file must be byte-equal. It names the first
// differing entry.
func sameBesidesExample(root, oldHash, newHash string) error {
	oldEntries, err := gitread.RootEntries(root, oldHash)
	if err != nil {
		return err
	}
	newEntries, err := gitread.RootEntries(root, newHash)
	if err != nil {
		return err
	}
	byName := func(es []gitread.Entry) map[string]gitread.Entry {
		m := map[string]gitread.Entry{}
		for _, e := range es {
			m[e.Name] = e
		}
		return m
	}
	o, n := byName(oldEntries), byName(newEntries)
	names := make([]string, 0, len(o)+len(n))
	for name := range o {
		names = append(names, name)
	}
	for name := range n {
		if _, ok := o[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		if name == "example.txtar" {
			continue
		}
		a, inOld := o[name]
		b, inNew := n[name]
		if inOld != inNew || a.Mode != b.Mode || a.Hash != b.Hash {
			if a.Type == "tree" || b.Type == "tree" {
				name += "/"
			}
			return fault.Refuse(fault.GuardAddsOnly, fault.PathDiffers, name)
		}
	}
	return nil
}

// putOutputs stores each output under .tree/outputs by its sha256. An act
// calls it under the lock, after its guards held and before it appends the
// event that names the outputs.
func putOutputs(root string, outs ...[]byte) error {
	for _, b := range outs {
		if _, err := logstore.PutOutput(root, b); err != nil {
			return err
		}
	}
	return nil
}
