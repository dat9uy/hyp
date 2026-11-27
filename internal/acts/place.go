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
	Draft        string // the name of the draft in the clone's trace that holds the example set; empty for goals/G
	ProbeTimeout time.Duration

	// snap is the snapshot of the draft that every read of the set uses.
	snap string

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
	dir             string // the example folder of a draft's snapshot; empty for a committed one
	hash            string // the example folder's tree hash
	probeHash       string // the tree hash of its probe/ folder
	ex              example.Example
	lineRow         tables.Line
	findings        []string // the finding.add ids its findings section names
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
// and names every cause. The guards its entry in hyp spec names, in
// order: Not landed, not Selected and Grill places for the Goal; then, for
// the set, Whole set, Committed, Adds only and Meetable, which need no
// Probe; then Kind, which runs each changed example on H0. A set that changes no example is
// refused by Adds only. The Probes run with the lock released, and every guard
// is checked again under the lock before the event is appended.
func Place(o PlaceOpts) (PlaceReport, error) {
	w, err := begin(o.Dir, o.Actor, o.Role, o.Why)
	if err != nil {
		return PlaceReport{}, err
	}
	root := w.root
	if o.Draft != "" {
		// One snapshot of the draft: the bytes that are checked, hashed and
		// run on H0 are the bytes the Goal's folder gets.
		snap, done, err := snapshotDraft(root, o.Draft)
		if err != nil {
			return PlaceReport{}, err
		}
		defer done()
		o.snap = snap
	}

	var first scan
	err = w.check(eventsView, func(t tables.Tables) error {
		var err error
		first, err = scanSet(root, o, t)
		if err != nil {
			return err
		}
		return first.refusal()
	})
	if err != nil {
		return PlaceReport{}, err
	}

	runs, err := measureOnH0(root, first.goal.Base, first.changes, o.ProbeTimeout)
	if err != nil {
		return PlaceReport{}, err
	}
	if o.afterProbe != nil {
		o.afterProbe()
	}

	var rep PlaceReport
	var written []change
	ev, err := w.append(eventsView, func(t tables.Tables, _ []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeExamplesPlace, Why: o.Why, Actor: o.Actor}
		again, err := scanSet(root, o, t)
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
		if causes := kindCauses(again.changes, runs); len(causes) > 0 {
			return ev, setRefusal(again.goal.ID, causes)
		}
		placed, outs, cells := placedExamples(again.goal.Base, again.goal.BaseTree, again.changes, runs)
		payload, err := json.Marshal(events.ExamplesPlace{Role: o.Role, Examples: placed})
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeExamplesPlace, err)
		}
		ev.Node, ev.Payload = again.goal.EventID, payload
		if err := tables.GuardPlaceSet(t, ev); err != nil {
			return ev, err
		}
		// The outputs and cells are stored only once every guard holds, just
		// before the event: a refused place leaves no file behind.
		if err := putOutputs(root, outs...); err != nil {
			return ev, err
		}
		if err := putCells(root, cells...); err != nil {
			return ev, err
		}
		rep.Goal = again.goal.ID
		written = again.changes
		for i, c := range again.changes {
			rep.Placed = append(rep.Placed, PlacedRow{Key: c.key, Hash: c.hash, Verdict: string(runs[i].verdict), Exit: runs[i].out.Exit})
		}
		return ev, nil
	})
	if err != nil {
		return PlaceReport{}, err
	}
	if o.snap != "" {
		// the Goal's own folder gets the examples the place changed, as later
		// places read them by default; hyp writes no git state, so the owner
		// commits them
		for _, c := range written {
			to := filepath.Join(root, "goals", rep.Goal, c.line, c.name)
			if err := os.RemoveAll(to); err != nil {
				return PlaceReport{}, fault.EnvFault(fault.CopyExamples, c.dir, to, err)
			}
			if err := copyFolder(c.dir, to); err != nil {
				return PlaceReport{}, fault.EnvFault(fault.CopyExamples, c.dir, to, err)
			}
		}
	}
	after, _, err := load(root, eventsView)
	if err != nil {
		return PlaceReport{}, err
	}
	if g := after.Goals[ev.Node]; g != nil {
		rep.State = g.State
		for i := range rep.Placed {
			if cur, ok := g.Current(rep.Placed[i].Key); ok {
				rep.Placed[i].Version = cur.Version
			}
		}
	}
	return rep, nil
}

// measured is one changed example run on H0: what it printed and its verdict.
type measured struct {
	out     probe.Out
	verdict example.Verdict
}

// measureOnH0 runs each changed example on the commit base, H0.
func measureOnH0(root, base string, changes []change, timeout time.Duration) ([]measured, error) {
	runs := make([]measured, len(changes))
	for i, c := range changes {
		out, err := probe.Run(probe.Opts{Repo: root, Commit: base, ExampleHash: c.hash, ExampleDir: c.dir, Timeout: timeout})
		if err != nil {
			return nil, err
		}
		runs[i] = measured{out: out, verdict: example.Compare(out.Stdout, c.ex)}
	}
	return runs, nil
}

// kindCauses is the Kind guard on the measured changes: each example whose
// verdict on H0 is not the one its line's kind needs.
func kindCauses(changes []change, runs []measured) []SetCause {
	var causes []SetCause
	for i, c := range changes {
		if string(runs[i].verdict) != tables.Needs(c.lineRow.Kind) {
			causes = append(causes, SetCause{Key: c.key, Err: fault.Refuse(fault.GuardKind, fault.WrongVerdictOnH0, KindFacts{
				Line: c.lineRow, Verdict: string(runs[i].verdict), Out: runs[i].out,
				Floors: c.ex.Floors, Ceilings: c.ex.Ceilings,
			})})
		}
	}
	return causes
}

// placedExamples are the examples an owner act records: git holds each
// example's hash and H0's verdict; H0's outputs and cell go to the clone's
// trace, a cache a reader fills again with try.
func placedExamples(base, baseTree string, changes []change, runs []measured) ([]events.PlacedExample, [][]byte, []logstore.Cell) {
	var placed []events.PlacedExample
	var outs [][]byte
	var cells []logstore.Cell
	for i, c := range changes {
		out := runs[i].out
		h0 := events.Cell{
			Commit: base, RowTree: baseTree, Probe: c.probeHash,
			Exit: out.Exit, Verdict: string(runs[i].verdict),
		}
		placed = append(placed, events.PlacedExample{Line: c.line, Example: c.name, Hash: c.hash, H0: h0, Findings: c.findings})
		outs = append(outs, out.Stdout, out.Stderr)
		cells = append(cells, logstore.Cell{
			Hash: c.hash, RowTree: h0.RowTree, Commit: h0.Commit, Probe: h0.Probe,
			Output: logstore.Hash(out.Stdout), Stderr: logstore.Hash(out.Stderr),
			Exit: h0.Exit, Verdict: h0.Verdict, By: logstore.CellByPlace,
		})
	}
	return placed, outs, cells
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
// Goal, Whole set, Committed, Adds only and Meetable on each example
// folder, then the names of each changed example's findings section (Not
// refuted, Not declined, One in work) and Covered for the set, which it collects as causes
// so a refusal names every one. Selected reads t in eventsView, only a select
// event: a select this clone's Results show holds no event, and the owner's
// place of a new example then moves the Goal back to open. An unknown Goal is
// a bad argument.
func scanSet(root string, o PlaceOpts, t tables.Tables) (scan, error) {
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
	if err := tables.GuardOwnerPlaces(o.Role); err != nil {
		return scan{}, err
	}
	if err := tables.GuardGoalActor(g, o.Actor, o.Role); err != nil {
		return scan{}, err
	}
	if o.snap != "" {
		return scanFolder(root, path.Join("drafts", o.Draft), o.snap, g, t)
	}
	return scanFolder(root, path.Join("goals", g.ID), "", g, t)
}

// exampleFolder is the clean slash form of a folder of examples given
// relative to the top level; a folder outside it is a bad argument.
func exampleFolder(dir string) (string, error) {
	clean := path.Clean(filepath.ToSlash(dir))
	if clean == "." || path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fault.BadArg(fault.BadExamplesFolder, dir)
	}
	return clean, nil
}

// scanFolder runs, for every line of g, Whole set, Committed, Adds only and
// Meetable on each example folder under dir/<line>/, then the names of each
// changed example's findings section and Covered, and collects the causes.
// With disk set, the set is the snapshot of a draft at disk, which git does
// not hold: dir only names it in a refusal, and Committed does not apply.
func scanFolder(root, dir, disk string, g *tables.Goal, t tables.Tables) (scan, error) {
	dir, err := exampleFolder(dir)
	if err != nil {
		return scan{}, err
	}
	s := scan{goal: g}
	for _, l := range g.Lines {
		lineDir := path.Join(dir, l.Key)
		names, err := exampleNames(root, lineDir, diskLine(disk, l.Key))
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
			c, changed, err := checkExample(root, dir, disk, g, l, name)
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
	s.causes = append(s.causes, s.nameFindings(t)...)
	sort.SliceStable(s.causes, func(i, j int) bool { return s.causes[i].Key < s.causes[j].Key })
	return s, nil
}

// nameFindings resolves the findings section of each changed example in the
// tables t and checks each new name, then checks Covered: every finding the
// Goal names at add goal is named by an example of the set, a changed one or
// the current version of an unchanged one. A finding the Goal names already
// is not checked again, so a clone whose trace lacks it places the set. It
// returns the causes it found.
func (s *scan) nameFindings(t tables.Tables) []SetCause {
	g := s.goal
	named := g.Findings()
	var causes []SetCause
	covered := map[string]bool{}
	changed := map[string]bool{}
	for i := range s.changes {
		c := &s.changes[i]
		changed[c.key] = true
		for _, name := range c.ex.Findings {
			id := ""
			for _, n := range named {
				if n == name || tables.FindingID(n) == name {
					id = n
				}
			}
			if id == "" {
				f, err := t.FindFinding(name)
				if err == nil {
					err = tables.GuardNameable(f, g)
				}
				if err != nil {
					causes = append(causes, SetCause{Key: c.key, Err: err})
					continue
				}
				id = f.EventID
			}
			if !slices.Contains(c.findings, id) {
				c.findings = append(c.findings, id)
			}
			covered[id] = true
		}
	}
	for _, p := range g.Placed() {
		if !changed[p.Key] {
			for _, id := range p.Findings {
				covered[id] = true
			}
		}
	}
	for _, id := range g.Named {
		if !covered[id] {
			causes = append(causes, SetCause{Key: tables.FindingID(id), Err: fault.Refuse(fault.GuardCovered, fault.NotCovered, g.ID, tables.FindingID(id))})
		}
	}
	return causes
}

// exampleNames lists the example folders of a line folder, on disk and at
// HEAD, sorted: a folder on disk that is not committed and a folder committed
// but gone from disk are both examples of the set, and Committed refuses
// them.
func exampleNames(root, lineDir, disk string) ([]string, error) {
	seen := map[string]bool{}
	from := filepath.Join(root, filepath.FromSlash(lineDir))
	if disk != "" {
		from = disk
	}
	entries, err := os.ReadDir(from)
	if err != nil && !os.IsNotExist(err) {
		return nil, fault.EnvFault(fault.ReadFolder, lineDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			seen[e.Name()] = true
		}
	}
	if disk != "" {
		return sortedNames(seen), nil
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
	return sortedNames(seen), nil
}

// sortedNames is the keys of set, sorted.
func sortedNames(set map[string]bool) []string {
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// diskLine is the folder of a line in the snapshot at disk; empty for a set
// git holds.
func diskLine(disk, line string) string {
	if disk == "" {
		return ""
	}
	return filepath.Join(disk, line)
}

// checkExample runs the guards that need no Probe on one example folder of
// the set: its name, Committed, Adds only, then Meetable on a new hash. It
// reports whether the folder changes the set: a new example key or a new
// hash on a placed one.
func checkExample(root, dir, disk string, g *tables.Goal, l tables.Line, name string) (change, bool, error) {
	rel := path.Join(dir, l.Key, name)
	if !exampleNameRe.MatchString(name) {
		return change{}, false, fault.Refuse(fault.GuardWholeSet, fault.NotAnExampleName, rel)
	}
	c := change{key: l.Key + "/" + name, line: l.Key, name: name, lineRow: l}
	var err error
	if disk != "" {
		c.dir = filepath.Join(disk, l.Key, name)
		if c.hash, err = diskHash(c.dir, rel); err != nil {
			return change{}, false, err
		}
	} else if c.hash, err = committedHash(root, rel); err != nil {
		return change{}, false, err
	}
	read := func() (example.Example, string, error) {
		if c.dir != "" {
			return readExampleDir(c.dir)
		}
		return readExample(root, c.hash)
	}
	cur, placed := g.Current(c.key)
	if placed && cur.Hash == c.hash {
		if cur.H0.RowTree == g.BaseTree {
			return change{}, false, nil
		}
		// A replace moved H0 since the place: the same version is measured
		// again on the new H0, and git then holds its verdict there.
		if c.ex, c.probeHash, err = read(); err != nil {
			return change{}, false, err
		}
		return c, true, nil
	}
	if c.ex, c.probeHash, err = read(); err != nil {
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
		if err := sameBesidesExample(root, cur.Hash, c.hash, c.dir); err != nil {
			return change{}, false, err
		}
		if err := example.AddsOnly(old, c.ex, oldProbe, c.probeHash); err != nil {
			return change{}, false, err
		}
	}
	if err := example.Meetable(c.ex); err != nil {
		return change{}, false, err
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

// diskHash is the tree hash of the draft's example folder at disk, as git
// would hash it; rel names the folder in a refusal.
func diskHash(disk, rel string) (string, error) {
	if fi, err := os.Stat(disk); err != nil || !fi.IsDir() {
		return "", fault.Refuse(fault.GuardWholeSet, fault.NotAFolder, rel)
	}
	h, err := gitread.DirTreeHash(disk)
	if err != nil {
		if fault.KindOf(err) == fault.Refusal {
			return "", fault.Refuse(fault.GuardWholeSet, fault.FolderOnDisk, err)
		}
		return "", err
	}
	return h, nil
}

// readExampleDir reads and parses the example.txtar of the example folder at
// disk, and returns the tree hash of its probe/ folder.
func readExampleDir(disk string) (example.Example, string, error) {
	data, err := os.ReadFile(filepath.Join(disk, "example.txtar"))
	if err != nil {
		return example.Example{}, "", fault.Refuse(fault.NoGuard, fault.NoExampleTxtar, disk)
	}
	ex, err := example.Parse(data)
	if err != nil {
		return example.Example{}, "", err
	}
	probeHash := ""
	if fi, err := os.Stat(filepath.Join(disk, "probe")); err == nil && fi.IsDir() {
		if probeHash, err = gitread.DirTreeHash(filepath.Join(disk, "probe")); err != nil {
			return example.Example{}, "", err
		}
	}
	return ex, probeHash, nil
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
func sameBesidesExample(root, oldHash, newHash, newDisk string) error {
	oldEntries, err := gitread.RootEntries(root, oldHash)
	if err != nil {
		return err
	}
	var newEntries []gitread.Entry
	if newDisk != "" {
		newEntries, err = gitread.DirEntries(newDisk)
	} else {
		newEntries, err = gitread.RootEntries(root, newHash)
	}
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
		if inOld != inNew || strings.TrimLeft(a.Mode, "0") != strings.TrimLeft(b.Mode, "0") || a.Hash != b.Hash {
			if a.Type == "tree" || b.Type == "tree" {
				name += "/"
			}
			return fault.Refuse(fault.GuardAddsOnly, fault.PathDiffers, name)
		}
	}
	return nil
}

// putOutputs stores each output in the clone's trace by its sha256. An act
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

// putCells saves each cell in the clone's trace.
func putCells(root string, cells ...logstore.Cell) error {
	if len(cells) == 0 {
		return nil
	}
	tr, err := logstore.TraceOf(root)
	if err != nil {
		return err
	}
	for _, c := range cells {
		if err := tr.PutCell(c); err != nil {
			return err
		}
	}
	return nil
}
