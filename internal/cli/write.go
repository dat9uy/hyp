package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/claim"
	"github.com/dat9uy/hyp/internal/contain"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/render"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/sandbox"
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
	noWhy  bool // the event has no why: an answer (ADR 0020)
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
	*w.why = strings.TrimSpace(*w.why)
	if w.noWhy && *w.why != "" {
		return nil, usageErr("an answer has no --why: the owner's reason is each item's text")
	}
	if !w.noWhy && *w.why == "" {
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
	if ev.Type == event.TypeAnswer {
		fmt.Fprintf(w.stdout, "%s answer\n", ev.ID)
		return nil
	}
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
		printClosed(w.stdout, w.st, ev)
	}

	return nil
}

// printClosed prints, in full, the text of each limit a Result closes, so
// the runner rereads every case the limit names at the moment it claims to
// have run them. A case it did not run is a new --limit, referred when it
// is void (ADR 0016).
func printClosed(out io.Writer, st *replay.State, ev event.Event) {
	var p event.ResultRecord
	_ = json.Unmarshal(ev.Payload, &p)
	for _, c := range p.Closes {
		k, ok := event.ParseKey(c)
		n := st.Nodes[k.Result]
		if !ok || n == nil {
			continue
		}
		texts := n.Limits
		if n.Kind == event.KindGoal {
			texts = st.Lines[n.ID]
		}
		if k.N > len(texts) {
			continue
		}
		fmt.Fprintf(out, "  closes %s: %s\n", event.DisplayKey(n.Kind, c), texts[k.N-1])
	}
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
	maxBody := event.MaxBodyOf(kind)
	body := w.fs.String("body", "", fmt.Sprintf("body (at most %d runes)", maxBody))
	alias := w.fs.String("alias", "", "optional unique alias")
	var parent, recipeRef, decidedBy, rubricRef *string
	var targets, aims, subject, records, lineTexts, cites multi
	if kind != event.KindGoal {
		parent = w.fs.String("parent", "", "parent node (required)")
	} else {
		w.fs.Var(&lineTexts, "line", fmt.Sprintf("a line of the Goal: a case every candidate must meet, keyed G-id#n; a Goal's lines come with its add and no later event adds one (repeatable, at most %d, each 1–%d runes)", event.MaxLimits, event.MaxLine))
		w.fs.Var(&cites, "cite", fmt.Sprintf("N=KEY: line N cites KEY, a line, an accepted or closed limit, an answered reading, or a selected or contradicted Hypothesis, settled now (repeatable, at most %d per line)", event.MaxCites))
	}
	if kind == event.KindHypothesis {
		w.fs.Var(&subject, "subject", "a file or directory that embodies the claim; hyp ties the Hypothesis to each file's sha256 (repeatable)")
	}
	if kind == event.KindExperiment {
		recipeRef = w.fs.String("recipe", "", "NAME@VERSION or none (required)")
		w.fs.Var(&targets, "target", "sibling Hypothesis this Experiment also tests (repeatable), or none")
		rubricRef = w.fs.String("rubric", "", "NAME@VERSION of a rubric bundle, or reading: how measures become verdicts (required under a Hypothesis, refused under a DesignQuestion)")
		decidedBy = w.fs.String("decided-by", "", "reading: the older spelling of --rubric reading")
		w.fs.Var(&records, "records", "file or directory under .tree/evidence/ that an earlier run kept and this Experiment reads; hyp pins each file's sha256 (repeatable)")
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
	citeOf, err := lineCites(cites, len(lineTexts))
	if err != nil {
		return err
	}
	for i, t := range lineTexts {
		if err := truncatedHex(i+1, t); err != nil {
			return err
		}
		p.Lines = append(p.Lines, event.GoalLine{Text: t})
	}
	noTarget := len(targets) == 1 && targets[0] == "none"
	var warns []string
	var facts, rubricFacts replay.RecipeFacts
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
		if *decidedBy != "" && *rubricRef != "" {
			return usageErr("--decided-by reading is the older spelling of --rubric reading; pass only --rubric")
		}
		p.Recipe = *recipeRef
		p.Rubric = *rubricRef
		if *decidedBy != "" {
			p.Rubric = event.RubricReading
		}
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
			facts = replay.RecipeFacts{Instrument: r.IsInstrument(), HasCases: recipe.HasKnownAnswers(r.Dir), Verdicts: sandbox.PrintsVerdicts(r),
				Hash: r.Hash, GoWarnings: recipe.GoFileWarnings(r.Dir), Placeholder: placeholder}
		}
		if p.Rubric != "" && p.Rubric != event.RubricReading {
			r, err := recipe.LoadRubric(w.root, p.Rubric)
			if err != nil {
				return err
			}
			p.RubricHash = r.Hash
			placeholder, err := recipe.HasPlaceholder(r.Dir)
			if err != nil {
				return herr.Env("rubric %s: %v", p.Rubric, err)
			}
			rubricFacts = replay.RecipeFacts{Hash: r.Hash, Placeholder: placeholder, NotRubric: sandbox.RubricProblems(r)}
		}
		if len(records) > 0 {
			var err error
			if p.Records, err = recordFiles(w.root, records); err != nil {
				return err
			}
		}
		// The rules that read no tree refuse before the store is locked.
		if _, err := admitted((*replay.State)(nil).AdmitExp(p, facts, rubricFacts, noTarget)); err != nil {
			return err
		}
	}
	var files []event.SubjectFile
	if len(subject) > 0 {
		var err error
		if files, err = subjectFiles(w.root, subject); err != nil {
			return err
		}
	}
	err = w.run(func() (draft, error) {
		var err error
		for _, c := range citeOf {
			id, err := w.citeKey(c.key)
			if err != nil {
				return draft{}, err
			}
			p.Lines[c.line-1].Cites = append(p.Lines[c.line-1].Cites, id)
		}
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
			if warns, err = admitted(w.st.AdmitExp(p, facts, rubricFacts, noTarget)); err != nil {
				return draft{}, err
			}
		}
		if kind == event.KindHypothesis {
			if _, err = admitted(w.p.AdmitAddHyp(p.Parent)); err != nil {
				return draft{}, err
			}
		}
		return draft{event.TypeNodeAdd, "", p}, nil
	})
	if err == nil && kind == event.KindGoal {
		if w := blockWarning(p.Body); w != "" {
			fmt.Fprintln(stderr, w)
		}
		for i, l := range p.Lines {
			fmt.Fprintf(stdout, "  line %s: %s\n", event.DisplayKey(event.KindGoal, event.LimitKey(w.last.ID, i+1)), l.Text)
		}
	}
	if err == nil && len(files) > 0 {
		id := w.last.ID
		err = w.run(func() (draft, error) {
			return draft{event.TypeHypothesisSubject, id, event.HypothesisSubject{Files: files}}, nil
		})
	}
	if err == nil {
		for _, l := range warns {
			fmt.Fprintln(stderr, l)
		}
	}
	return err
}

// truncatedHexRe finds a whole run of 8 to 63 hex digits followed by an
// ellipsis, such as "c47a38de…": a shortened hash.
var truncatedHexRe = regexp.MustCompile(`(?:^|[^0-9A-Fa-f])([0-9A-Fa-f]{8,63})(?:…|\.\.\.)`)

// truncatedHex refuses line n of a new Goal when it names a shortened hash.
// A line never changes, so a hash it names must be whole; a log that
// already holds such a line still replays, because only the write refuses.
func truncatedHex(n int, text string) error {
	if m := truncatedHexRe.FindStringSubmatch(text); m != nil {
		return refusedErr("line %d names a hash cut to %d hex digits (%s…); a line never changes, so name the file by its whole sha256 (64 hex digits)", n, len(m[1]), m[1])
	}
	return nil
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
	w.fs.Var(&verdicts, "verdict", "REF=supports|contradicts|inconclusive, one per target of the Experiment and none under a measuring one (repeatable)")
	w.fs.Var(&evidence, "evidence", fmt.Sprintf("URL, or path under .tree/evidence or recipes; recipes/NAME@VERSION/FILE is looked up like --recipe (repeatable, at most %d, each at most %d runes)", event.MaxEvidence, event.MaxEvidenceItem))
	w.fs.Var(&limits, "limit", fmt.Sprintf("an open limit of the tested design (repeatable, at most %d, each at most %d runes)", event.MaxLimits, event.MaxLimitItem))
	w.fs.Var(&closes, "closes", "R-id#n of a limit this Result closes (repeatable)")
	w.fs.Var(&readings, "reading", fmt.Sprintf("how you read a Goal line a verdict rests on (repeatable, at most %d, each at most %d runes)", event.MaxLimits, event.MaxReadingItem))
	w.fs.Var(&readingsOf, "reading-of", "R-id#rn=TEXT: how you read the owner's answer to that reading; it narrows the answer and never reopens it (repeatable; numbered after the --reading items)")
	w.fs.Var(&applies, "applies", "R-id#rn of an answered reading whose answer these verdicts apply; the Experiment must aim at it (repeatable)")
	if _, err := w.begin(args, 0); err != nil {
		return err
	}
	if *exp == "" {
		return usageErr("record needs --exp")
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
	runs := runSubjects(w.root, evidence)
	if recordHashed != nil {
		recordHashed()
	}
	var warns []string
	err := w.run(func() (draft, error) {
		warns = nil
		expID, err := w.ref(*exp)
		if err != nil {
			return draft{}, err
		}
		p := event.ResultRecord{Parent: expID, Verdicts: map[string]string{}, Observation: *observation, Evidence: evidence,
			EvidenceHashes: hashes, Limits: limits, Subjects: runs}
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
		admitWarns, err := admitted(w.st.AdmitResult(p, w.resultFacts(expID), refHash))
		if err != nil {
			return draft{}, err
		}
		warns = admitWarns
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
		for _, k := range w.p.UnclosedAims(expID, p.Closes, len(p.Limits)) {
			warns = append(warns, fmt.Sprintf("warn: %s aims at %s and this Result does not close it; a close records that the case ran, whatever the verdicts (add --closes %s), and a case the run did not reach is a new --limit",
				event.DisplayID(event.KindExperiment, expID), w.p.KeyLabel(k), w.p.KeyLabel(k)))
		}
		return draft{event.TypeResultRecord, "", p}, nil
	})
	if err == nil {
		for _, l := range warns {
			fmt.Fprintln(stderr, l)
		}
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

// resultFacts loads what AdmitResult needs about Experiment expID: its
// recipe and rubric bundles, each left empty when it pins none, and its
// pinned records that no longer hash as pinned.
func (w *writer) resultFacts(expID string) replay.ResultFacts {
	exp := w.st.Nodes[expID]
	if exp == nil || exp.Kind != event.KindExperiment {
		return replay.ResultFacts{}
	}
	pin := func(load func(root, ref string) (*recipe.Recipe, error), ref string) replay.RecipeFacts {
		r, err := load(w.root, ref)
		if err != nil {
			return replay.RecipeFacts{Err: err}
		}
		return replay.RecipeFacts{Hash: r.Hash}
	}
	var f replay.ResultFacts
	if exp.Recipe != event.RecipeNone {
		f.Recipe = pin(recipe.Load, exp.Recipe)
	}
	if exp.RubricHash != "" {
		f.Rubric = pin(recipe.LoadRubric, exp.Rubric)
	}
	f.Changed = changedRecords(w.root, exp.Records)
	return f
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
	form := map[event.KeyKind]string{event.KeyLimit: "<R-id>#<n> or <G-id>#<n>", event.KeyReading: "<R-id>#r<n>"}
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
			if err != nil {
				return draft{}, err
			}
			_, err = admitted(admitOnNode(w.p, typ, id))
			return draft{typ, id, event.Empty{}}, err
		})
	}
}

// admitOnNode is the write-time rule of an event on one node: an open
// reading holds the reaffirm of its question or of a Hypothesis there, and
// the abandon of a Hypothesis its answer can move (spec §5.1).
func admitOnNode(p replay.Projection, typ, id string) []replay.Finding {
	switch typ {
	case event.TypeReaffirm:
		return p.AdmitReaffirm(id)
	case event.TypeHypothesisAbandon:
		if n := p.Nodes[id]; n != nil && n.Kind == event.KindHypothesis {
			return p.AdmitRetire("abandon", id)
		}
	}
	return nil
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
		if err != nil {
			return draft{}, err
		}
		if n := w.p.Nodes[to]; n != nil && n.Kind == event.KindHypothesis {
			if _, err := admitted(w.p.AdmitRetire("supersede", to)); err != nil {
				return draft{}, err
			}
		}
		return draft{event.TypeEdgeAdd, from, event.Edge{Type: event.EdgeSupersedes, To: to}}, nil
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
		// hyp supersede is the one command that writes a supersedes edge, so
		// its write-time refusals live in one place. hyp unlink still
		// removes one, which reopens the superseded node.
		if typ == event.TypeEdgeAdd && pos[1] == event.EdgeSupersedes {
			return usageErr("hyp link does not write supersedes: use hyp supersede %s --old %s", pos[0], pos[2])
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
	scope := w.fs.Bool("scope-change", false, "retired: "+retiredScope)
	pos, err := parse(w.fs, args)
	if err != nil {
		return err
	}
	if *scope {
		return usageErr("select --scope-change was retired: %s", retiredScope)
	}
	switch len(pos) {
	case 1:
		return suggest(w, pos[0])
	case 2:
	default:
		return usageErr("usage: hyp select DQ [HYP]")
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
			PrevSelect: w.currentSelect(dq)}}
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

// retiredScope is what replaces a scope-change select and its owner confirm
// (ADR 0019): a Goal line the answer will not meet is referred and accepted,
// and a Goal that needs more is a new Goal, so a new tree.
const retiredScope = "a scope change is no longer a select: the runner refers the Goal line the answer will not meet (hyp checkpoint --refers G-…#n) and the owner accepts it (hyp answer --accept G-…#n=TEXT); a Goal that needs more is a new Goal, which cites this one's settled keys"

// cmdConfirm is retired: the select is the gate (ADR 0015), and a scope
// change is a line the owner accepts or a new tree (ADR 0019).
func cmdConfirm(args []string, stdout, stderr io.Writer) error {
	const use = "hyp confirm was retired: the owner's choose, recorded as the select, ends a trade-off, and " + retiredScope
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		return helpText(use)
	}
	return usageErr(use)
}

// retiredOwnerFlags are the checkpoint flags that wrote an owner's act or a
// Goal's lines. No command writes those fields now (ADR 0020, owner decision
// of 2026-11-11): the owner's acts are hyp answer, and a Goal's lines come
// with hyp add goal --line.
var retiredOwnerFlags = map[string]string{
	"role":    "a checkpoint is the runner's; the owner's acts are their own event: hyp answer",
	"closes":  "the owner answers a reading with hyp answer --answer R-…#rN=TEXT",
	"revises": "the owner gives a new answer with hyp answer --answer R-…#rN=TEXT",
	"accept":  "the owner accepts a referred limit with hyp answer --accept K=TEXT",
	"returns": "the owner returns a limit with hyp answer --returns K=TEXT",
	"line":    "a Goal's lines come with its node.add: hyp add goal --line TEXT; a new line is a new Goal",
}

func cmdCheckpoint(args []string, stdout, stderr io.Writer) error {
	w := newWriter("checkpoint", stdout, stderr)
	summary := w.fs.String("summary", "", fmt.Sprintf("where the investigation stands (at most %d runes); a checkpoint holds a summary, refers, or both", event.MaxSummary))
	goal := w.fs.String("goal", "", "the Goal this checkpoint is on; required when the store holds more than one Goal")
	var refers multi
	w.fs.Var(&refers, "refers", "K or K=TEXT: R-id#n or G-id#n of a limit no test within the Goal's lines can settle, handed to the owner; TEXT, or else --summary, says why (repeatable)")
	retired := map[string]*multi{}
	for name, use := range retiredOwnerFlags {
		retired[name] = &multi{}
		w.fs.Var(retired[name], name, "retired: "+use)
	}
	if _, err := w.begin(args, 0); err != nil {
		return err
	}
	for _, name := range []string{"role", "closes", "revises", "accept", "returns", "line"} {
		if len(*retired[name]) > 0 {
			return usageErr("--%s was retired: %s", name, retiredOwnerFlags[name])
		}
	}
	*summary = strings.TrimSpace(*summary)
	err := w.run(func() (draft, error) {
		g, err := w.checkpointGoal(*goal)
		if err != nil {
			return draft{}, err
		}
		p := event.Checkpoint{Summary: *summary}
		if p.Refers, err = w.keyTexts("--refers", refers, false, event.KeyLimit); err != nil {
			return draft{}, err
		}
		return draft{event.TypeCheckpoint, g, p}, nil
	})
	return err
}

// cmdAnswer writes the owner's act, an answer (ADR 0020): answers to
// readings, accepts and returns of limits, each KEY=TEXT with the owner's
// words. It has no node, no summary and no why. One answer may settle keys
// of several Goals.
func cmdAnswer(args []string, stdout, stderr io.Writer) error {
	w := newWriter("answer", stdout, stderr)
	w.noWhy = true
	var answers, accepts, returns multi
	w.fs.Var(&answers, "answer", "R-id#rN=TEXT: the owner's answer to a reading, a first answer while it is open and a new one once answered (repeatable)")
	w.fs.Var(&accepts, "accept", "K=TEXT: R-id#n or G-id#n of a referred limit the owner accepts, so this Goal does not run its case (repeatable)")
	w.fs.Var(&returns, "returns", "K=TEXT: R-id#n or G-id#n of a referred or accepted limit whose case the Goal covers, handed back to the runner (repeatable)")
	if _, err := w.begin(args, 0); err != nil {
		return err
	}
	var warns []string
	err := w.run(func() (draft, error) {
		var p event.Answer
		var err error
		if p.Answers, err = w.keyTexts("--answer", answers, true, event.KeyReading); err != nil {
			return draft{}, err
		}
		if p.Accepts, err = w.keyTexts("--accept", accepts, true, event.KeyLimit); err != nil {
			return draft{}, err
		}
		if p.Returns, err = w.keyTexts("--returns", returns, true, event.KeyLimit); err != nil {
			return draft{}, err
		}
		var closes, revises []string
		for _, it := range p.Answers {
			if l, ok := w.p.Reading(it.Key); ok && !l.Open() {
				revises = append(revises, it.Key)
				warns = append(warns, citedElsewhere(w.p, it.Key)...)
			} else {
				closes = append(closes, it.Key)
			}
		}
		warns = append(warns, sameAskWarnings(w.p, closes, revises)...)
		return draft{event.TypeAnswer, "", p}, nil
	})
	if err == nil {
		for _, l := range warns {
			fmt.Fprintln(stderr, l)
		}
	}
	return err
}

// citedElsewhere warns when the reading key k an answer revises is cited by
// a line of a Goal other than k's own: that line was written on the earlier
// answer (ADR 0020). The revision is admitted, at exit 0.
func citedElsewhere(p replay.Projection, k string) []string {
	own := p.GoalOf(event.KeyResult(k))
	var out []string
	for _, l := range p.Lines {
		if l.Result == own {
			continue
		}
		for _, c := range l.Cites {
			if c.Key == k {
				out = append(out, fmt.Sprintf("warn: %s revises %s, which %s cites; that line was written on the earlier answer", p.KeyLabel(k), p.KeyLabel(k), p.KeyLabel(l.Key)))
			}
		}
	}
	return out
}

// keyTexts resolves KEY=TEXT arguments, or KEY alone when needText is
// false, to keyed items of full ids; each key must be one of kinds.
func (w *writer) keyTexts(flag string, args []string, needText bool, kinds ...event.KeyKind) ([]event.KeyText, error) {
	var out []event.KeyText
	for _, a := range args {
		k, text, found := strings.Cut(a, "=")
		if text = strings.TrimSpace(text); needText && (!found || text == "") {
			return nil, usageErr("%s %q: want KEY=TEXT, the owner's words for that key", flag, a)
		}
		keys, err := w.keys(flag, []string{strings.TrimSpace(k)}, kinds...)
		if err != nil {
			return nil, err
		}
		out = append(out, event.KeyText{Key: keys[0], Text: text})
	}
	return out, nil
}

// lineCite is one --cite N=KEY of hyp add goal.
type lineCite struct {
	line int
	key  string
}

// lineCites parses --cite N=KEY arguments for a Goal with n lines.
func lineCites(args []string, n int) ([]lineCite, error) {
	var out []lineCite
	for _, a := range args {
		num, key, found := strings.Cut(a, "=")
		i, err := strconv.Atoi(strings.TrimSpace(num))
		if !found || err != nil || i < 1 || strings.TrimSpace(key) == "" {
			return nil, usageErr("--cite %q: want N=KEY, line N cites KEY", a)
		}
		if i > n {
			return nil, usageErr("--cite %q: the Goal has %d lines", a, n)
		}
		out = append(out, lineCite{i, strings.TrimSpace(key)})
	}
	return out, nil
}

// citeKey resolves a cite: a limit or reading key, or a node reference.
func (w *writer) citeKey(c string) (string, error) {
	if strings.Contains(c, "#") {
		keys, err := w.keys("--cite", []string{c}, event.KeyLimit, event.KeyReading)
		if err != nil {
			return "", err
		}
		return keys[0], nil
	}
	return w.ref(c)
}

// checkpointGoal is the Goal a checkpoint is on: the one named, or the
// store's only Goal. A store with many Goals needs it named.
func (w *writer) checkpointGoal(named string) (string, error) {
	if named != "" {
		return w.ref(named)
	}
	switch len(w.st.Goals) {
	case 0:
		return "", refusedErr("no Goal yet")
	case 1:
		return w.st.Goals[0], nil
	}
	return "", usageErr("missing --goal: the store holds %d Goals", len(w.st.Goals))
}

// sameAskWarnings names each reading that asks the same question as one the
// checkpoint closes or revises (spec §4 readings) and that the checkpoint
// leaves behind: an open copy it does not close, or an answered copy it does
// not revise. The owner may answer them apart on purpose, so it warns.
func sameAskWarnings(p replay.Projection, closes, revises []string) []string {
	var out []string
	seen := map[string]bool{}
	check := func(keys []string, open bool, flag, left string) {
		for _, k := range keys {
			for _, m := range p.SameAsk(k) {
				l, _ := p.Reading(m)
				if m == k || l.Open() != open || slices.Contains(keys, m) || seen[m] {
					continue
				}
				seen[m] = true
				out = append(out, fmt.Sprintf("warn: %s asks the same as %s and %s; add %s %s",
					p.KeyLabel(m), p.KeyLabel(k), left, flag, p.KeyLabel(m)))
			}
		}
	}
	check(closes, true, "--answer", "stays open")
	check(revises, false, "--answer", "keeps its earlier answer")
	return out
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
