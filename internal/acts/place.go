package acts

import (
	"encoding/json"
	"fmt"
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
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/herr"
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
// guard of the state machine, section 8: Not ended (always holds), Selected,
// Own case, Committed, Adds only, then Kind. Everything but Kind is checked
// before the Probe runs, so a refusal that needs no Probe costs no Probe run;
// the Probe runs with the lock released, and every guard is checked again
// under the lock before the event is appended.
func Place(o PlaceOpts) (PlaceReport, error) {
	if !logstore.ValidActor(o.Actor) {
		return PlaceReport{}, herr.BadArgs("actor %q must match ^[a-z0-9][a-z0-9-]{0,39}$ (--actor or $HYP_ACTOR)", o.Actor)
	}
	if !ValidRole(o.Role) {
		return PlaceReport{}, herr.BadArgs("role %q must be one of %s (--role or $HYP_ROLE)", o.Role, strings.Join(Roles, ", "))
	}
	if n := utf8.RuneCountInString(o.Why); n > logstore.MaxWhy {
		return PlaceReport{}, herr.BadArgs("why is %d runes; the limit is %d", n, logstore.MaxWhy)
	}
	lineKey, name, ok := strings.Cut(o.Key, "/")
	if !ok || !lineKeyRe.MatchString(lineKey) || !exampleNameRe.MatchString(name) {
		return PlaceReport{}, herr.BadArgs("%q is not LINE/EXAMPLE, for example line-1/a (an example name matches ^[a-z0-9][a-z0-9-]{0,39}$)", o.Key)
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
	outHash, err := logstore.PutOutput(root, out.Stdout)
	if err != nil {
		return PlaceReport{}, err
	}
	errHash, err := logstore.PutOutput(root, out.Stderr)
	if err != nil {
		return PlaceReport{}, err
	}
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
			return ev, herr.Refused("Committed: %s changed from %s to %s while the Probe ran; place it again", o.Key, first.hash, again.hash)
		case again.goal.Base != first.goal.Base:
			return ev, herr.Refused("Kind: H0 of %s moved from %s to %s while the Probe ran; place it again", again.goal.ID, first.goal.Base, again.goal.Base)
		}
		if string(verdict) != tables.Needs(again.line.Kind) {
			return ev, kindRefusal(again.line, again.ex, out, verdict)
		}
		payload, err := json.Marshal(events.ExamplePlace{
			Role: o.Role, Line: lineKey, Example: name, Hash: again.hash,
			H0: events.Cell{
				Commit: again.goal.Base, RowTree: again.goal.BaseTree, Probe: again.probeHash,
				Output: outHash, Stderr: errHash, Exit: out.Exit, Verdict: string(verdict),
			},
		})
		if err != nil {
			return ev, herr.Env("encode example.place: %v", err)
		}
		ev.Node, ev.Payload = again.goal.EventID, payload
		if err := tables.GuardPlace(tables.Replay(evs), ev); err != nil {
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
// machine: Selected, Own case, Committed, then Adds only. An unknown Goal or
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
		return checked{}, herr.BadArgs("%s has no line %s", g.ID, lineKey)
	}
	if err := tables.GuardNotSelected(g); err != nil {
		return checked{}, err
	}
	if err := tables.GuardOwnCase(o.Role); err != nil {
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
		return "", herr.Refused("Committed: %s is not a folder on disk", rel)
	}
	onDisk, err := gitread.DirTreeHash(disk)
	if err != nil {
		if herr.KindOf(err) == herr.Refusal {
			return "", herr.Refused("Committed: %s", herr.Text(err))
		}
		return "", err
	}
	code, out, err := gitread.GitExit(root, "rev-parse", "--verify", "HEAD:"+rel)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", herr.Refused("Committed: %s is not committed at HEAD", rel)
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
		return "", herr.Refused("Committed: %s on disk is not the folder committed at HEAD", rel)
	}
	return "", herr.Refused("Committed: %s differs from HEAD at %s", rel, path.Join(rel, diff))
}

// readExample reads and parses the example.txtar of the example folder with
// tree hash hash, and returns the tree hash of its probe/ folder.
func readExample(root, hash string) (example.Example, string, error) {
	if code, _, err := gitread.GitExit(root, "cat-file", "-e", hash); err != nil {
		return example.Example{}, "", err
	} else if code != 0 {
		return example.Example{}, "", herr.Env("object %s not in this clone", hash)
	}
	spec := hash + ":example.txtar"
	if code, _, err := gitread.GitExit(root, "cat-file", "-e", spec); err != nil {
		return example.Example{}, "", err
	} else if code != 0 {
		return example.Example{}, "", herr.Refused("example folder %s has no example.txtar", hash)
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

// kindRefusal is the Kind refusal: it prints H0's output beside the example's
// floors and ceilings.
func kindRefusal(line tables.Line, ex example.Example, out probe.Out, verdict example.Verdict) error {
	want := "fail"
	if line.Kind == events.KindGuard {
		want = "be met"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Kind: %s is a %s line, so its example must %s on H0; it is %s\n", line.Key, line.Kind, want, verdict)
	fmt.Fprintf(&b, "H0 output (exit %d):\n%s\n", out.Exit, indent(string(out.Stdout)))
	fmt.Fprintf(&b, "H0 stderr:\n%s\n", indent(string(out.Stderr)))
	b.WriteString("floors:")
	for _, f := range ex.Floors {
		fmt.Fprintf(&b, "\n  %s %q", f.Op, f.Text)
	}
	b.WriteString("\nceilings:")
	for _, c := range ex.Ceilings {
		fmt.Fprintf(&b, "\n  %q", c)
	}
	return herr.Refused("%s", b.String())
}

func indent(s string) string {
	if s == "" {
		return "  (empty)"
	}
	return "  " + strings.ReplaceAll(strings.TrimSuffix(s, "\n"), "\n", "\n  ")
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
			return herr.Refused("Adds only: %s differs from the current version; a change there is a new example key, not a new version", name)
		}
	}
	return nil
}
