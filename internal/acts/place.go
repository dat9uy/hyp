package acts

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
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
	ProbeTimeout time.Duration

	// afterProbe, when set, runs after the Probes and before the second check;
	// a test uses it to land an event between the two checks.
	afterProbe func()
}

// PlaceReport is what place returns: every example the place changed, in key
// order, and the Goal's state after it.
type PlaceReport struct {
	Goal   string
	Placed []PlacedRow
	State  string
}

// PlacedRow is one example a place changed.
type PlacedRow struct {
	Key     string
	Version int
	Hash    string
	Verdict string // H0's verdict
	Exit    int    // the exit code of the run on H0
}

// SetCause is one reason a place places none of the set: the example key, or
// the line key for a line that holds no example, and the guard's refusal.
type SetCause struct {
	Key string
	Err error
}

// SetFacts is what a refusal that places none of the set shows: the Goal and
// every cause, in key order.
type SetFacts struct {
	Goal   string
	Causes []SetCause
}

// change is one example of the set that a place writes: a new example key,
// or a new version of a placed one.
type change struct {
	key, line, name string
	hash            string // the example folder's tree hash
	probeHash       string // the tree hash of its probe/ folder
	ex              example.Example
	lineRow         tables.Line
}

// scan is what the guards that need no Probe found for the whole set.
type scan struct {
	goal    *tables.Goal
	changes []change // in key order
	causes  []SetCause
}

// Place places the Goal's whole example set: every example folder under
// goals/G/<line>/, for each line of the Goal. It records one examples.place
// event that holds every example the set changed, or places none of the set
// and names every cause. The guards of the state machine, section 8, in
// order: Not landed, not Selected and Grill places for the Goal; then, for
// the set, Whole set, Committed and Adds only, which need no Probe; then Kind,
// which runs each changed example on H0. A set that changes no example is
// refused by Adds only. The Probes run with the lock released, and every guard
// is checked again under the lock before the event is appended.
func Place(o PlaceOpts) (PlaceReport, error) {
	if err := CheckActor(o.Actor, o.Role); err != nil {
		return PlaceReport{}, err
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return PlaceReport{}, fault.BadArg(fault.WhyTooLong, n, logstore.MaxWhy)
	}
	root, err := logstore.TopLevel(o.Dir)
	if err != nil {
		return PlaceReport{}, err
	}

	var first scan
	err = logstore.WithLock(root, func(evs []events.Event) error {
		var err error
		first, err = scanSet(root, o, evs)
		if err != nil {
			return err
		}
		return first.refusal()
	})
	if err != nil {
		return PlaceReport{}, err
	}

	type measured struct {
		out     probe.Out
		verdict example.Verdict
	}
	runs := make([]measured, len(first.changes))
	for i, c := range first.changes {
		out, err := probe.Run(probe.Opts{Repo: root, Commit: first.goal.Base, ExampleHash: c.hash, Timeout: o.ProbeTimeout})
		if err != nil {
			return PlaceReport{}, err
		}
		runs[i] = measured{out: out, verdict: example.Compare(out.Stdout, c.ex)}
	}
	if o.afterProbe != nil {
		o.afterProbe()
	}

	var rep PlaceReport
	ev, err := logstore.Append(root, o.Actor, func(evs []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeExamplesPlace, Why: o.Why}
		again, err := scanSet(root, o, evs)
		if err != nil {
			return ev, err
		}
		if err := again.refusal(); err != nil {
			return ev, err
		}
		if again.goal.Base != first.goal.Base {
			return ev, fault.Refuse(fault.GuardKind, fault.H0MovedWhilePlacing, again.goal.ID, first.goal.Base, again.goal.Base)
		}
		if !sameChanges(first.changes, again.changes) {
			return ev, fault.Refuse(fault.GuardCommitted, fault.SetChangedWhileProbe, again.goal.ID)
		}
		var causes []SetCause
		for i, c := range again.changes {
			if string(runs[i].verdict) != tables.Needs(c.lineRow.Kind) {
				causes = append(causes, SetCause{Key: c.key, Err: fault.Refuse(fault.GuardKind, fault.WrongVerdictOnH0, KindFacts{
					Line: c.lineRow, Verdict: string(runs[i].verdict), Out: runs[i].out,
					Floors: c.ex.Floors, Ceilings: c.ex.Ceilings,
				})})
			}
		}
		if len(causes) > 0 {
			return ev, setRefusal(again.goal.ID, causes)
		}
		p := events.ExamplesPlace{Role: o.Role}
		var outs [][]byte
		for i, c := range again.changes {
			out := runs[i].out
			p.Examples = append(p.Examples, events.PlacedExample{
				Line: c.line, Example: c.name, Hash: c.hash,
				H0: events.Cell{
					Commit: again.goal.Base, RowTree: again.goal.BaseTree, Probe: c.probeHash,
					Output: logstore.Hash(out.Stdout), Stderr: logstore.Hash(out.Stderr),
					Exit: out.Exit, Verdict: string(runs[i].verdict),
				},
			})
			outs = append(outs, out.Stdout, out.Stderr)
		}
		payload, err := json.Marshal(p)
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeExamplesPlace, err)
		}
		ev.Node, ev.Payload = again.goal.EventID, payload
		if err := tables.GuardPlaceSet(tables.Replay(evs), ev); err != nil {
			return ev, err
		}
		// The outputs are stored only once every guard holds, just before the
		// event that names them: a refused place leaves no file behind.
		if err := putOutputs(root, outs...); err != nil {
			return ev, err
		}
		rep.Goal = again.goal.ID
		for i, c := range again.changes {
			rep.Placed = append(rep.Placed, PlacedRow{Key: c.key, Hash: c.hash, Verdict: string(runs[i].verdict), Exit: runs[i].out.Exit})
		}
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
		for i := range rep.Placed {
			if cur, ok := g.Current(rep.Placed[i].Key); ok {
				rep.Placed[i].Version = cur.Version
			}
		}
	}
	return rep, nil
}

// refusal is the refusal of a scan that found a cause, or that changes no
// example; nil when the set may go on to its Probes.
func (s scan) refusal() error {
	if len(s.causes) > 0 {
		return setRefusal(s.goal.ID, s.causes)
	}
	if len(s.changes) == 0 {
		return fault.Refuse(fault.GuardAddsOnly, fault.SetUnchanged, s.goal.ID)
	}
	return nil
}

// setRefusal is the refusal of a place that places none of the set. Its guard
// is the first cause's, so a caller that reads one guard reads the first one
// the state machine checks.
func setRefusal(goal string, causes []SetCause) error {
	return fault.Refuse(fault.GuardOf(causes[0].Err), fault.SetNotPlaced, SetFacts{Goal: goal, Causes: causes})
}

// sameChanges reports whether two scans found the same changes, by key and
// hash.
func sameChanges(a, b []change) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].key != b[i].key || a[i].hash != b[i].hash {
			return false
		}
	}
	return true
}

// scanSet runs the guards that need no Probe: Not landed, not Selected and
// Grill places for the Goal, which refuse at once; then, for every line of the
// Goal, Whole set, Committed and Adds only on each example folder, which it
// collects as causes so a refusal names every one. An unknown Goal is a bad
// argument.
func scanSet(root string, o PlaceOpts, evs []events.Event) (scan, error) {
	t := tables.Replay(evs)
	g, err := t.Find(o.Goal)
	if err != nil {
		return scan{}, err
	}
	if err := tables.GuardNotEnded(g); err != nil {
		return scan{}, err
	}
	if err := tables.GuardNotSelected(g); err != nil {
		return scan{}, err
	}
	if err := tables.GuardGrillPlaces(o.Role); err != nil {
		return scan{}, err
	}
	s := scan{goal: g}
	for _, l := range g.Lines {
		lineDir := path.Join("goals", g.ID, l.Key)
		names, err := exampleNames(root, lineDir)
		if err != nil {
			return scan{}, err
		}
		for _, p := range g.Placed() {
			if p.Line == l.Key && !slices.Contains(names, p.Example) {
				s.causes = append(s.causes, SetCause{Key: p.Key, Err: fault.Refuse(fault.GuardWholeSet, fault.PlacedFolderGone, p.Key, path.Join(lineDir, p.Example))})
			}
		}
		if len(names) == 0 {
			s.causes = append(s.causes, SetCause{Key: l.Key, Err: fault.Refuse(fault.GuardWholeSet, fault.LineHoldsNoExample, l.Key, lineDir)})
			continue
		}
		for _, name := range names {
			key := l.Key + "/" + name
			c, changed, err := checkExample(root, g, l, name)
			if err != nil {
				if fault.KindOf(err) != fault.Refusal {
					return scan{}, err
				}
				s.causes = append(s.causes, SetCause{Key: key, Err: err})
				continue
			}
			if changed {
				s.changes = append(s.changes, c)
			}
		}
	}
	sort.SliceStable(s.causes, func(i, j int) bool { return s.causes[i].Key < s.causes[j].Key })
	return s, nil
}

// exampleNames lists the example folders of a line folder, on disk and at
// HEAD, sorted: a folder on disk that is not committed and a folder committed
// but gone from disk are both examples of the set, and Committed refuses
// them.
func exampleNames(root, lineDir string) ([]string, error) {
	seen := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(lineDir)))
	if err != nil && !os.IsNotExist(err) {
		return nil, fault.EnvFault(fault.ReadFolder, lineDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			seen[e.Name()] = true
		}
	}
	code, out, err := gitread.GitExit(root, "rev-parse", "--verify", "HEAD:"+lineDir)
	if err != nil {
		return nil, err
	}
	if code == 0 {
		committed, err := gitread.RootEntries(root, strings.TrimSpace(string(out)))
		if err != nil {
			return nil, err
		}
		for _, e := range committed {
			if e.Type == "tree" {
				seen[e.Name] = true
			}
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	return names, nil
}

// checkExample runs the guards that need no Probe on one example folder of
// the set: its name, Committed, then Adds only. It reports whether the folder
// changes the set: a new example key or a new hash on a placed one.
func checkExample(root string, g *tables.Goal, l tables.Line, name string) (change, bool, error) {
	rel := path.Join("goals", g.ID, l.Key, name)
	if !exampleNameRe.MatchString(name) {
		return change{}, false, fault.Refuse(fault.GuardWholeSet, fault.NotAnExampleName, rel)
	}
	c := change{key: l.Key + "/" + name, line: l.Key, name: name, lineRow: l}
	var err error
	if c.hash, err = committedHash(root, rel); err != nil {
		return change{}, false, err
	}
	cur, placed := g.Current(c.key)
	if placed && cur.Hash == c.hash {
		return change{}, false, nil
	}
	if c.ex, c.probeHash, err = readExample(root, c.hash); err != nil {
		return change{}, false, err
	}
	if err := tables.GuardAddsOnlyTable(g, c.key, c.hash); err != nil {
		return change{}, false, err
	}
	if placed {
		old, oldProbe, err := readExample(root, cur.Hash)
		if err != nil {
			return change{}, false, err
		}
		if err := sameBesidesExample(root, cur.Hash, c.hash); err != nil {
			return change{}, false, err
		}
		if err := example.AddsOnly(old, c.ex, oldProbe, c.probeHash); err != nil {
			return change{}, false, err
		}
	}
	return c, true, nil
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
