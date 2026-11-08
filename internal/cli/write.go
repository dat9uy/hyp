package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/claim"
	"github.com/dat9uy/hyp/internal/contain"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/render"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/store"
)

// helpText is the usage text a -h flag asks for: printed on stdout, exit 0.
type helpText string

func (h helpText) Error() string { return string(h) }

func usageErr(format string, a ...any) error { return herr.BadArgs(format, a...) }

func refusedErr(format string, a ...any) error { return herr.Refused(format, a...) }

// multi is a repeatable string flag.
type multi []string

func (m *multi) String() string     { return strings.Join(*m, ",") }
func (m *multi) Set(v string) error { *m = append(*m, v); return nil }

// parse parses flags that may appear before, between or after positional
// arguments, and returns the positional arguments.
func parse(fs *flag.FlagSet, args []string) ([]string, error) {
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				var b strings.Builder
				fs.SetOutput(&b)
				fmt.Fprintf(&b, "usage of hyp %s:\n", fs.Name())
				fs.PrintDefaults()
				return nil, helpText(strings.TrimSuffix(b.String(), "\n"))
			}
			return nil, usageErr("%v", err)
		}
		if fs.NArg() == 0 {
			return pos, nil
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

// writer is one write command in progress. st and p are set only inside
// the build callback of run, where they hold the state the event is
// validated against.
type writer struct {
	fs     *flag.FlagSet
	why    *string
	actor  *string
	stdout io.Writer
	root   string
	st     *replay.State
	p      replay.Projection
	last   event.Event // the event run appended
}

func newWriter(name string, stdout, stderr io.Writer) *writer {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Run prints the one error line
	return &writer{fs: fs, stdout: stdout,
		why:   fs.String("why", "", fmt.Sprintf("why this event is recorded (required, 1–%d runes)", event.MaxWhy)),
		actor: fs.String("actor", "", "actor name (default $HYP_ACTOR)")}
}

// begin parses args, checks --why and the actor, and finds the store.
func (w *writer) begin(args []string, npos ...int) ([]string, error) {
	pos, err := parse(w.fs, args)
	if err != nil {
		return nil, err
	}
	if !containsInt(npos, len(pos)) {
		return nil, usageErr("%s: wrong number of arguments", w.fs.Name())
	}
	if *w.actor == "" {
		*w.actor = os.Getenv("HYP_ACTOR")
	}
	if *w.actor == "" {
		return nil, usageErr("missing actor: pass --actor or set HYP_ACTOR")
	}
	if !store.ValidActor(*w.actor) {
		return nil, usageErr("invalid actor %q (want ^[a-z0-9][a-z0-9-]{0,39}$)", *w.actor)
	}
	if *w.why = strings.TrimSpace(*w.why); *w.why == "" {
		return nil, usageErr("missing --why")
	}
	return pos, w.findRoot()
}

func (w *writer) findRoot() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if w.root, err = store.FindRoot(cwd); err != nil {
		return err
	}
	return nil
}

func (w *writer) ref(r string) (string, error) {
	id, err := resolveRef(w.p, r)
	if err != nil {
		return "", err
	}
	return id, nil
}

// draft is the event a command wants to write.
type draft struct {
	typ, node string
	payload   any
}

// run appends the event that build returns, with refs resolved inside build
// against the locked state, and prints "<display id> <event type>". It
// refuses text flags that cite an id that is no node (checkCites).
func (w *writer) run(build func() (draft, error)) error {
	ev, err := store.AppendWith(w.root, *w.actor, func(st *replay.State, p replay.Projection) (event.Event, error) {
		w.st, w.p = st, p
		d, err := build()
		if err != nil {
			return event.Event{}, err
		}
		if err := w.checkCites(); err != nil {
			return event.Event{}, err
		}
		raw, err := json.Marshal(d.payload)
		if err != nil {
			return event.Event{}, err
		}
		return event.Event{Type: d.typ, Node: d.node, Payload: raw, Why: *w.why}, nil
	})
	if err != nil {
		return err
	}
	w.last = ev
	kind := event.KindResult
	if ev.Type == event.TypeNodeAdd {
		var p event.NodeAdd
		_ = json.Unmarshal(ev.Payload, &p)
		kind = p.Kind
	} else if n := w.p.Nodes[ev.Node]; n != nil {
		kind = n.Kind
	}
	fmt.Fprintf(w.stdout, "%s %s\n", event.DisplayID(kind, ev.Node), ev.Type)
	if ev.Type == event.TypeResultRecord {
		printKeys(w.stdout, ev)
	}
	return nil
}

// printKeys prints the key of each limit and reading a Result records, so
// a later --closes, --aims, --applies or --reading-of can name it without
// counting flags.
func printKeys(out io.Writer, ev event.Event) {
	var p event.ResultRecord
	_ = json.Unmarshal(ev.Payload, &p)
	id := event.DisplayID(event.KindResult, ev.Node)
	for i, l := range p.Limits {
		fmt.Fprintf(out, "  limit %s#%d: %s\n", id, i+1, render.Cut(l, 80))
	}
	for i, l := range p.Readings {
		of := ""
		if l.Of != "" {
			of = " of " + event.DisplayID(event.KindResult, event.KeyResult(l.Of)) + l.Of[len(event.KeyResult(l.Of)):]
		}
		fmt.Fprintf(out, "  reading %s#r%d%s: %s\n", id, i+1, of, render.Cut(l.Text, 80))
	}
}

// currentSelect returns the current selection id of DQ dq ("" if none).
func (w *writer) currentSelect(dq string) string {
	if sel := w.st.Decisions[dq]; sel != nil {
		return sel.Current
	}
	return ""
}

var addKinds = map[string]string{
	"goal": event.KindGoal,
	"dq":   event.KindDesignQuestion,
	"hyp":  event.KindHypothesis,
	"exp":  event.KindExperiment,
}

func cmdAdd(args []string, stdout, stderr io.Writer) error {
	const use = "usage: hyp add goal|dq|hyp|exp [flags]"
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		return helpText(use)
	}
	if len(args) == 0 || addKinds[args[0]] == "" {
		return usageErr(use)
	}
	kind := addKinds[args[0]]
	w := newWriter("add "+args[0], stdout, stderr)
	title := w.fs.String("title", "", fmt.Sprintf("title (required, at most %d runes)", event.MaxTitle))
	maxBody := event.MaxText
	if kind == event.KindGoal {
		maxBody = event.MaxGoalBody
	}
	body := w.fs.String("body", "", fmt.Sprintf("body (at most %d runes)", maxBody))
	alias := w.fs.String("alias", "", "optional unique alias")
	var parent, recipeRef, decidedBy *string
	var targets, aims multi
	if kind != event.KindGoal {
		parent = w.fs.String("parent", "", "parent node (required)")
	}
	if kind == event.KindExperiment {
		recipeRef = w.fs.String("recipe", "", "NAME@VERSION or none (required)")
		w.fs.Var(&targets, "target", "sibling Hypothesis this Experiment also tests (repeatable), or none")
		decidedBy = w.fs.String("decided-by", "", "reading: the verdicts are decided by reading; required when the recipe is none or not an instrument")
		w.fs.Var(&aims, "aims", "R-id#n of a limit, or R-id#rn of an answered reading, this Experiment means to settle (repeatable)")
	}
	if _, err := w.begin(args[1:], 0); err != nil {
		return err
	}
	if err := required("--title", title); err != nil {
		return err
	}
	if parent != nil && *parent == "" {
		return usageErr("missing --parent")
	}
	p := event.NodeAdd{Kind: kind, Title: *title, Body: *body, Alias: *alias}
	noTarget := len(targets) == 1 && targets[0] == "none"
	var warns []string
	var facts replay.RecipeFacts
	if kind == event.KindExperiment {
		if !noTarget && slices.Contains(targets, "none") {
			return usageErr("--target none cannot be combined with other targets")
		}
		if *recipeRef == "" {
			return usageErr("missing --recipe (NAME@VERSION or none)")
		}
		if *decidedBy != "" && *decidedBy != event.DecidedReading {
			return usageErr("invalid --decided-by %q (want %s)", *decidedBy, event.DecidedReading)
		}
		p.Recipe = *recipeRef
		p.DecidedBy = *decidedBy
		if *recipeRef != event.RecipeNone {
			r, err := recipe.Load(w.root, *recipeRef)
			if err != nil {
				return err
			}
			p.RecipeHash = r.Hash
			placeholder, err := recipe.HasPlaceholder(r.Dir)
			if err != nil {
				return herr.Env("recipe %s: %v", *recipeRef, err)
			}
			facts = replay.RecipeFacts{Instrument: r.IsInstrument(), HasCases: recipe.HasKnownAnswers(r.Dir),
				Hash: r.Hash, GoWarnings: recipe.GoFileWarnings(r.Dir), Placeholder: placeholder}
		}
		// The rules that read no tree refuse before the store is locked.
		if _, err := admitted((*replay.State)(nil).AdmitExp(p, facts, noTarget)); err != nil {
			return err
		}
	}
	err := w.run(func() (draft, error) {
		var err error
		if parent != nil {
			if p.Parent, err = w.ref(*parent); err != nil {
				return draft{}, err
			}
		}
		p.Targets = nil
		if noTarget {
			targets = nil
		}
		for _, t := range targets {
			id, err := w.ref(t)
			if err != nil {
				return draft{}, err
			}
			p.Targets = append(p.Targets, id)
		}
		if p.Aims, err = w.keys("--aims", aims, event.KeyLimit, event.KeyReading); err != nil {
			return draft{}, err
		}
		if kind == event.KindExperiment {
			if warns, err = admitted(w.st.AdmitExp(p, facts, noTarget)); err != nil {
				return draft{}, err
			}
		}
		return draft{event.TypeNodeAdd, "", p}, nil
	})
	if err == nil {
		for _, l := range warns {
			fmt.Fprintln(stderr, l)
		}
	}
	return err
}

// admitted maps admission findings to the CLI: the first refusal (exit 1)
// or bad-args finding (exit 2) is the error, and each warning becomes a
// "warn: " line to print once the event is written.
func admitted(fs []replay.Finding) (warns []string, err error) {
	for _, f := range fs {
		switch f.Kind {
		case replay.FindRefuse:
			return nil, refusedErr("%s", f.Text)
		case replay.FindBadArgs:
			return nil, usageErr("%s", f.Text)
		default:
			warns = append(warns, "warn: "+f.Text)
		}
	}
	return warns, nil
}

// recordHashed, when set, runs after cmdRecord hashes its evidence and
// before the locked write; tests use it to edit a recipe in that window.
var recordHashed func()

func cmdRecord(args []string, stdout, stderr io.Writer) error {
	w := newWriter("record", stdout, stderr)
	exp := w.fs.String("exp", "", "Experiment the Result belongs to (required)")
	observation := w.fs.String("observation", "", fmt.Sprintf("what was observed (required, at most %d runes)", event.MaxText))
	var verdicts, evidence, limits, closes, readings, readingsOf, applies multi
	w.fs.Var(&verdicts, "verdict", "REF=supports|contradicts|inconclusive (repeatable, at least one)")
	w.fs.Var(&evidence, "evidence", fmt.Sprintf("URL, or path under .tree/evidence or recipes; recipes/NAME@VERSION/FILE is looked up like --recipe (repeatable, at most %d, each at most %d runes)", event.MaxEvidence, event.MaxEvidenceItem))
	w.fs.Var(&limits, "limit", fmt.Sprintf("an open limit of the tested design (repeatable, at most %d, each at most %d runes)", event.MaxLimits, event.MaxLimitItem))
	w.fs.Var(&closes, "closes", "R-id#n of a limit this Result closes (repeatable)")
	w.fs.Var(&readings, "reading", fmt.Sprintf("how you read a Goal line a verdict rests on (repeatable, at most %d, each at most %d runes)", event.MaxLimits, event.MaxReadingItem))
	w.fs.Var(&readingsOf, "reading-of", "R-id#rn=TEXT: how you read the owner's answer to that reading; it narrows the answer and never reopens it (repeatable; numbered after the --reading items)")
	w.fs.Var(&applies, "applies", "R-id#rn of an answered reading whose answer these verdicts apply; the Experiment must aim at it (repeatable)")
	if _, err := w.begin(args, 0); err != nil {
		return err
	}
	if *exp == "" || len(verdicts) == 0 {
		return usageErr("record needs --exp and at least one --verdict")
	}
	if err := required("--observation", observation); err != nil {
		return err
	}
	for _, v := range verdicts {
		if !strings.Contains(v, "=") {
			return usageErr("--verdict %q: want REF=supports|contradicts|inconclusive", v)
		}
	}
	// One hash per ref: a second pointer into a recipe reuses the first
	// pointer's hash, so an edit made meanwhile cannot split the pin.
	var hashes map[string]string
	refHash := map[string]string{}
	for _, e := range evidence {
		h, err := checkEvidence(w.root, e)
		if err != nil {
			return err
		}
		if h == "" {
			continue
		}
		r, _ := event.EvidenceRef(e)
		ref := r.String()
		if first, ok := refHash[ref]; ok {
			h = first
		}
		refHash[ref] = h
		if hashes == nil {
			hashes = map[string]string{}
		}
		hashes[e] = h
	}
	if recordHashed != nil {
		recordHashed()
	}
	err := w.run(func() (draft, error) {
		expID, err := w.ref(*exp)
		if err != nil {
			return draft{}, err
		}
		p := event.ResultRecord{Parent: expID, Verdicts: map[string]string{}, Observation: *observation, Evidence: evidence,
			EvidenceHashes: hashes, Limits: limits}
		for _, t := range readings {
			p.Readings = append(p.Readings, event.ReadingItem{Text: t})
		}
		for _, s := range readingsOf {
			ref, text, ok := strings.Cut(s, "=")
			if !ok {
				return draft{}, usageErr("--reading-of %q: want <R-id>#r<n>=TEXT", s)
			}
			keys, err := w.keys("--reading-of", []string{ref}, event.KeyReading)
			if err != nil {
				return draft{}, err
			}
			p.Readings = append(p.Readings, event.ReadingItem{Of: keys[0], Text: text})
		}
		if _, err := admitted(w.st.AdmitResult(p, w.recipeFacts(expID), refHash)); err != nil {
			return draft{}, err
		}
		for _, v := range verdicts {
			ref, verdict, _ := strings.Cut(v, "=")
			id, err := w.ref(ref)
			if err != nil {
				return draft{}, err
			}
			p.Verdicts[id] = verdict
		}
		if p.Closes, err = w.keys("--closes", closes, event.KeyLimit); err != nil {
			return draft{}, err
		}
		if p.Applies, err = w.keys("--applies", applies, event.KeyLimit, event.KeyReading); err != nil {
			return draft{}, err
		}
		return draft{event.TypeResultRecord, "", p}, nil
	})
	if err == nil {
		warnResult(w.root, w.last.Node, stderr)
	}
	return err
}

// warnResult prints the reasons Result id is uncontained (spec §7.6) and
// where its text disagrees with the files it cites (spec §7.6). The event
// is already appended, so nothing here changes the exit code.
func warnResult(root, id string, stderr io.Writer) {
	evs, _, err := store.Events(root)
	if err != nil {
		fmt.Fprintf(stderr, "warn: uncontained check skipped: %v\n", err)
		return
	}
	p := replay.Replay(evs, recipe.Hashes(root), "")
	for _, why := range contain.Reasons(root, p, id) {
		fmt.Fprintf(stderr, "warn: uncontained %s (%s)\n", event.DisplayID(event.KindResult, id), why)
	}
	for _, w := range claim.Warnings(root, p, id) {
		fmt.Fprintln(stderr, w)
	}
}

// recipeFacts loads what AdmitResult needs about Experiment expID's
// recipe; it is empty when expID is no Experiment or its recipe is none.
func (w *writer) recipeFacts(expID string) replay.RecipeFacts {
	exp := w.st.Nodes[expID]
	if exp == nil || exp.Kind != event.KindExperiment || exp.Recipe == event.RecipeNone {
		return replay.RecipeFacts{}
	}
	r, err := recipe.Load(w.root, exp.Recipe)
	if err != nil {
		return replay.RecipeFacts{Err: err}
	}
	return replay.RecipeFacts{Hash: r.Hash}
}

// checkEvidence accepts a URL, or an existing file relative to the store root
// under .tree/evidence/ or recipes/. A path recipes/<name>@<version>/<file>
// names a file in that recipe wherever recipe.Load finds it: beside the store
// or in a parent library, every copy hashing the same; its bundle hash is
// returned so the Result can pin it. Replay never reads files, so this is
// checked here.
func checkEvidence(root, e string) (hash string, err error) {
	if strings.HasPrefix(e, "http://") || strings.HasPrefix(e, "https://") {
		return "", nil
	}
	bad := usageErr("evidence %s not found under .tree/evidence or recipes", e)
	if filepath.IsAbs(e) {
		return "", bad
	}
	clean := filepath.ToSlash(filepath.Clean(e))
	if !strings.HasPrefix(clean, ".tree/evidence/") && !strings.HasPrefix(clean, "recipes/") {
		return "", bad
	}
	if err := event.BadEvidenceRef(clean); err != nil {
		return "", usageErr("evidence %s: %v", e, err)
	}
	if ref, ok := event.EvidenceRef(clean); ok {
		_, file, _ := strings.Cut(strings.TrimPrefix(clean, "recipes/"), "/")
		r, err := recipe.Load(root, ref.String())
		if err != nil {
			return "", herr.New(herr.KindOf(err), "evidence %s: %s", e, herr.Text(err))
		}
		fi, err := os.Stat(filepath.Join(r.Path, filepath.FromSlash(file)))
		if err != nil {
			return "", usageErr("evidence %s not found in recipe %s at %s", e, ref, r.Path)
		}
		if err := notDir(e, fi); err != nil {
			return "", err
		}
		return r.Hash, nil
	}
	fi, err := os.Stat(filepath.Join(root, clean))
	if err != nil {
		return "", bad
	}
	return "", notDir(e, fi)
}

// notDir refuses a directory as evidence: a pointer names one file, so what
// it shows cannot grow or change with the files added beside it.
func notDir(e string, fi os.FileInfo) error {
	if fi.IsDir() {
		return usageErr("evidence %s is a directory; cite a file in it", e)
	}
	return nil
}

// keys resolves "<REF>#<n>" and "<REF>#r<n>" arguments to keys of full ids;
// each must be one of kinds.
func (w *writer) keys(flag string, refs []string, kinds ...event.KeyKind) ([]string, error) {
	form := map[event.KeyKind]string{event.KeyLimit: "<R-id>#<n>", event.KeyReading: "<R-id>#r<n>"}
	var want []string
	for _, k := range kinds {
		want = append(want, form[k])
	}
	var out []string
	for _, s := range refs {
		k, ok := event.ParseKey(s)
		if !ok || !slices.Contains(kinds, k.Kind) {
			if ok && flag == "--closes" && k.Kind == event.KeyLimit {
				return nil, usageErr("--closes %q is a limit: an owner accepts a limit with --accept", s)
			}
			if ok && flag == "--revises" && k.Kind == event.KeyLimit {
				return nil, usageErr("--revises %q is a limit: a limit is never revised", s)
			}
			return nil, usageErr("%s %q: want %s", flag, s, strings.Join(want, " or "))
		}
		id, err := w.ref(k.Result)
		if err != nil {
			return nil, err
		}
		k.Result = id
		out = append(out, k.String())
	}
	return out, nil
}

// cmdOnNode handles commands whose only argument is the node they act on.
func cmdOnNode(name, typ string) func([]string, io.Writer, io.Writer) error {
	return func(args []string, stdout, stderr io.Writer) error {
		w := newWriter(name, stdout, stderr)
		pos, err := w.begin(args, 1)
		if err != nil {
			return err
		}
		return w.run(func() (draft, error) {
			id, err := w.ref(pos[0])
			return draft{typ, id, event.Empty{}}, err
		})
	}
}

// cmdWithdraw writes experiment.withdraw for an Experiment and dq.withdraw
// for any other node, which Validate refuses unless it is a DesignQuestion.
func cmdWithdraw(args []string, stdout, stderr io.Writer) error {
	w := newWriter("withdraw", stdout, stderr)
	pos, err := w.begin(args, 1)
	if err != nil {
		return err
	}
	return w.run(func() (draft, error) {
		id, err := w.ref(pos[0])
		typ := event.TypeDQWithdraw
		if n := w.p.Nodes[id]; n != nil && n.Kind == event.KindExperiment {
			typ = event.TypeExperimentWithdraw
		}
		return draft{typ, id, event.Empty{}}, err
	})
}

func cmdSupersede(args []string, stdout, stderr io.Writer) error {
	w := newWriter("supersede", stdout, stderr)
	old := w.fs.String("old", "", "the Hypothesis or Experiment being superseded (required)")
	pos, err := w.begin(args, 1)
	if err != nil {
		return err
	}
	if *old == "" {
		return usageErr("missing --old")
	}
	return w.run(func() (draft, error) {
		from, err := w.ref(pos[0])
		if err != nil {
			return draft{}, err
		}
		to, err := w.ref(*old)
		return draft{event.TypeEdgeAdd, from, event.Edge{Type: event.EdgeSupersedes, To: to}}, err
	})
}

func cmdAmend(args []string, stdout, stderr io.Writer) error {
	w := newWriter("amend", stdout, stderr)
	body := w.fs.String("body", "", fmt.Sprintf("the Hypothesis's revised body (required, at most %d runes)", event.MaxText))
	pos, err := w.begin(args, 1)
	if err != nil {
		return err
	}
	if err := required("--body", body); err != nil {
		return err
	}
	return w.run(func() (draft, error) {
		id, err := w.ref(pos[0])
		return draft{event.TypeHypothesisAmend, id, event.HypothesisAmend{Body: *body}}, err
	})
}

func cmdLink(name, typ string) func([]string, io.Writer, io.Writer) error {
	return func(args []string, stdout, stderr io.Writer) error {
		w := newWriter(name, stdout, stderr)
		pos, err := w.begin(args, 3)
		if err != nil {
			return err
		}
		return w.run(func() (draft, error) {
			from, err := w.ref(pos[0])
			if err != nil {
				return draft{}, err
			}
			to := pos[2]
			if pos[1] != event.EdgeExternalRef {
				if to, err = w.ref(pos[2]); err != nil {
					return draft{}, err
				}
			}
			return draft{typ, from, event.Edge{Type: pos[1], To: to}}, nil
		})
	}
}

func cmdSelect(args []string, stdout, stderr io.Writer) error {
	w := newWriter("select", stdout, stderr)
	scope := w.fs.Bool("scope-change", false, "the choice changes the Goal's scope")
	pos, err := parse(w.fs, args)
	if err != nil {
		return err
	}
	switch len(pos) {
	case 1:
		return suggest(w, pos[0])
	case 2:
	default:
		return usageErr("usage: hyp select DQ [HYP] [--scope-change]")
	}
	if pos, err = w.begin(args, 2); err != nil {
		return err
	}
	var warns []string
	err = w.run(func() (draft, error) {
		dq, err := w.ref(pos[0])
		if err != nil {
			return draft{}, err
		}
		h, err := w.ref(pos[1])
		if err != nil {
			return draft{}, err
		}
		d := draft{event.TypeDQSelect, dq, event.DQSelect{Hypothesis: h,
			PrevSelect: w.currentSelect(dq), ScopeChange: *scope}}
		// A select replay would refuse keeps replay's own refusal.
		if n := w.p.Nodes[h]; w.p.Questions[dq] == nil || n.Kind != event.KindHypothesis || n.Parent != dq || !replay.LiveHypStatus(w.p.Hyps[h].Status) {
			return d, nil
		}
		warns, err = admitted(w.p.AdmitSelect(dq, h))
		return d, err
	})
	if err == nil {
		for _, l := range warns {
			fmt.Fprintln(stderr, l)
		}
	}
	return err
}

// suggest prints the tool's suggestion for a DQ and writes nothing, so it
// needs neither --why nor an actor.
func suggest(w *writer, ref string) error {
	if err := w.findRoot(); err != nil {
		return err
	}
	p, _, err := store.Load(w.root)
	if err != nil {
		return err
	}
	w.p = p
	dq, err := w.ref(ref)
	if err != nil {
		return err
	}
	if p.Nodes[dq].Kind != event.KindDesignQuestion {
		return refusedErr("%s is not a DesignQuestion", ref)
	}
	// The suggestion reads the ledger, as admission does, so a suggested
	// select is one hyp admits.
	owed, choose := p.SelectOwed(dq)
	var ids []string
	for _, h := range owed {
		ids = append(ids, display(p, h))
	}
	switch {
	case choose:
		fmt.Fprintf(w.stdout, "trade-off: owner decides between %s\n", strings.Join(ids, ", "))
	case len(ids) == 1:
		fmt.Fprintf(w.stdout, "suggested: %s\n", ids[0])
	default:
		fmt.Fprintf(w.stdout, "suggested: none (%s)\n", p.SelectBlock(dq))
	}
	return nil
}

func cmdUnselect(args []string, stdout, stderr io.Writer) error {
	w := newWriter("unselect", stdout, stderr)
	pos, err := w.begin(args, 1)
	if err != nil {
		return err
	}
	return w.run(func() (draft, error) {
		dq, err := w.ref(pos[0])
		return draft{event.TypeDQUnselect, dq, event.DQUnselect{PrevSelect: w.currentSelect(dq)}}, err
	})
}

func cmdConfirm(args []string, stdout, stderr io.Writer) error {
	w := newWriter("confirm", stdout, stderr)
	role := w.fs.String("role", "", "owner (required); the verifier role was retired")
	verdict := w.fs.String("verdict", "", "sufficient or insufficient (required)")
	var accepts multi
	w.fs.Var(&accepts, "accept", "R-id#n of an open limit the owner accepts (repeatable)")
	pos, err := w.begin(args, 1)
	if err != nil {
		return err
	}
	if *role == "" || *verdict == "" {
		return usageErr("confirm needs --role and --verdict")
	}
	if *role == event.RoleVerifier {
		return refusedErr("the verifier role was retired; a review that finds a defect records a Result, a retract or a limit")
	}
	return w.run(func() (draft, error) {
		dq, err := w.ref(pos[0])
		if err != nil {
			return draft{}, err
		}
		keys, keyErr := w.keys("--accept", accepts, event.KeyLimit)
		// No current selection outranks a bad --accept, and AdmitConfirm refuses it.
		sel := w.currentSelect(dq)
		if keyErr != nil && sel != "" {
			return draft{}, keyErr
		}
		p := event.Confirm{Select: sel, Role: *role, Verdict: *verdict, Accepts: keys}
		if _, err := admitted(w.p.AdmitConfirm(dq, p)); err != nil {
			return draft{}, err
		}
		return draft{event.TypeConfirm, dq, p}, nil
	})
}

func cmdCheckpoint(args []string, stdout, stderr io.Writer) error {
	w := newWriter("checkpoint", stdout, stderr)
	summary := w.fs.String("summary", "", fmt.Sprintf("where the investigation stands (required, at most %d runes)", event.MaxText))
	role := w.fs.String("role", "", "owner, required with --closes, --revises and --accept")
	var closes, revises, accepts, refers multi
	w.fs.Var(&closes, "closes", "R-id#rn of an open reading this owner checkpoint answers; --summary holds the owner's words (repeatable)")
	w.fs.Var(&revises, "revises", "R-id#rn of an answered reading whose answer --summary replaces from now on (repeatable)")
	w.fs.Var(&accepts, "accept", "R-id#n of a limit the owner accepts untested (repeatable)")
	w.fs.Var(&refers, "refers", "R-id#n of a limit no test within the Goal's lines can settle, handed to the owner; --summary says why (repeatable)")
	if _, err := w.begin(args, 0); err != nil {
		return err
	}
	if err := required("--summary", summary); err != nil {
		return err
	}
	if *role != "" && *role != event.RoleOwner {
		return usageErr("--role must be owner")
	}
	if len(closes) > 0 && *role != event.RoleOwner {
		return usageErr("--closes needs --role owner")
	}
	if len(revises) > 0 && *role != event.RoleOwner {
		return usageErr("--revises needs --role owner")
	}
	if len(accepts) > 0 && *role != event.RoleOwner {
		return usageErr("--accept needs --role owner")
	}
	return w.run(func() (draft, error) {
		if w.st.GoalID == "" {
			return draft{}, refusedErr("no Goal yet")
		}
		p := event.Checkpoint{Summary: *summary, Role: *role}
		var err error
		if p.Closes, err = w.keys("--closes", closes, event.KeyReading); err != nil {
			return draft{}, err
		}
		if p.Revises, err = w.keys("--revises", revises, event.KeyReading); err != nil {
			return draft{}, err
		}
		if p.Accepts, err = w.keys("--accept", accepts, event.KeyLimit); err != nil {
			return draft{}, err
		}
		if p.Refers, err = w.keys("--refers", refers, event.KeyLimit); err != nil {
			return draft{}, err
		}
		return draft{event.TypeCheckpoint, w.st.GoalID, p}, nil
	})
}

// required trims a required text flag in place and reports a usage error
// when nothing is left.
func required(name string, v *string) error {
	if *v = strings.TrimSpace(*v); *v == "" {
		return usageErr("missing %s", name)
	}
	return nil
}

func containsInt(list []int, n int) bool {
	for _, x := range list {
		if x == n {
			return true
		}
	}
	return false
}
