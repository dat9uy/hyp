package replay

import (
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
)

// RecipeFacts is what the CLI loads about a bundle an Experiment pins, its
// recipe or its rubric, for admission. Replay reads no files, so the caller
// fills it in.
type RecipeFacts struct {
	Err        error    // why the ref does not load; nil when it loads
	Instrument bool     // it has an Evidence or Standalone command
	HasCases   bool     // it has known answers
	Verdicts   bool     // a known answer expects a verdict: line (sandbox.PrintsVerdicts)
	Hash       string   // its current bundle hash
	GoWarnings []string // recipe.GoFileWarnings of its directory
	// Placeholder lists its files that still hold the template's
	// placeholder marker (recipe.HasPlaceholder).
	Placeholder []string
	// NotRubric lists why a rubric bundle is not one (sandbox.RubricProblems).
	NotRubric []string
}

// ResultFacts is what the CLI loads about a Result's Experiment for
// admission: its recipe and rubric bundles, and the paths of its pinned
// records that no longer hash as pinned, missing ones included.
type ResultFacts struct {
	Recipe, Rubric RecipeFacts
	Changed        []string
}

// Finding kinds. A refusal is a broken rule (exit 1), a bad-args finding a
// command line to fix (exit 2), and a warning changes no exit code.
const (
	FindRefuse  = "refuse"
	FindBadArgs = "bad-args"
	FindWarn    = "warn"
)

// Finding is one result of an admission rule. Text is the message without a
// "warn: " prefix.
type Finding struct {
	Kind, Text string
}

// AdmitExp lists what admitting Experiment p would break or warn about, in
// rule order: a recipe still holding the template placeholder, an
// instrument with no known answers, a recipe whose known answers print a
// verdict, a rubric bundle that is not a rubric, a rubric bundle with no
// measures to read; then, when st is not nil, the parent's rules: under a
// DesignQuestion no rubric and no --target, under a Hypothesis a rubric,
// one-rival, and the D1 and Go-file warnings (ADR 0018). f describes the
// recipe and is ignored when it is none; rb describes the rubric and is
// ignored unless p.Rubric names a bundle. noTarget is --target none, which
// waives one-rival. With st nil only the rules that read no tree run, so the
// caller can refuse before it loads the store; p.Parent and p.Targets must
// be resolved ids otherwise.
func (st *State) AdmitExp(p event.NodeAdd, f, rb RecipeFacts, noTarget bool) []Finding {
	var out []Finding
	placeholder := func(kind, ref string, files []string) {
		if len(files) > 0 {
			out = append(out, Finding{FindRefuse, fmt.Sprintf("%s %s still holds the template placeholder (%s) in %s; replace it before pinning the bundle",
				kind, ref, recipe.Placeholder, strings.Join(files, ", "))})
		}
	}
	if p.Recipe != event.RecipeNone {
		placeholder("recipe", p.Recipe, f.Placeholder)
		if f.Instrument && !f.HasCases {
			out = append(out, Finding{FindRefuse, fmt.Sprintf("recipe %s is an instrument with no known answers; add known-answers/NAME/ cases before an Experiment uses it", p.Recipe)})
		}
		if f.Verdicts {
			out = append(out, Finding{FindRefuse, fmt.Sprintf("recipe %s: a recipe measures and never prints a verdict; put the rule in a rubric", p.Recipe)})
		}
	}
	bundle := p.Rubric != "" && p.Rubric != event.RubricReading
	if bundle {
		placeholder("rubric", p.Rubric, rb.Placeholder)
		if len(rb.NotRubric) > 0 {
			out = append(out, Finding{FindRefuse, fmt.Sprintf("rubric %s is not a rubric: %s", p.Rubric, strings.Join(rb.NotRubric, "; "))})
		}
		if p.Recipe == event.RecipeNone && len(p.Records) == 0 {
			out = append(out, Finding{FindRefuse, fmt.Sprintf("rubric %s has no measures to read: pin a recipe or --records", p.Rubric)})
		}
	}
	if st == nil {
		return out
	}
	parent := st.Nodes[p.Parent]
	if parent != nil && parent.Kind == event.KindDesignQuestion {
		if p.Rubric != "" || len(p.Targets) > 0 || noTarget {
			out = append(out, Finding{FindBadArgs, "an Experiment under a DesignQuestion measures: it has no targets and no rubric; drop --rubric and --target"})
		}
		return out
	}
	// The rival rules read a Hypothesis's siblings. Any other parent is
	// refused when the event is validated.
	if parent == nil || parent.Kind != event.KindHypothesis {
		return out
	}
	if p.Rubric == "" {
		out = append(out, Finding{FindRefuse, NoRubric})
	}
	if !noTarget {
		if text := st.oneRivalRefusal(p.Parent, p.Targets); text != "" {
			out = append(out, Finding{FindRefuse, text})
		}
	}
	for _, w := range st.notRivalWarnings(p.Parent, p.Targets) {
		out = append(out, Finding{FindWarn, w})
	}
	if p.Recipe != event.RecipeNone {
		for _, g := range f.GoWarnings {
			out = append(out, Finding{FindWarn, "recipe " + p.Recipe + ": " + g})
		}
	}
	return out
}

// NoRubric is the refusal of an Experiment under a Hypothesis that does not
// say how its measures become verdicts (ADR 0018).
const NoRubric = "an Experiment under a Hypothesis pins a rubric: --rubric NAME@VERSION for a rubric bundle, or --rubric reading when its body decides the verdicts by reading"

// oneRival is the one-rival rule: rivals, the live rivals of an
// Experiment's parent at its add, are not empty and targets names none.
func oneRival(rivals, targets []string) bool {
	for _, t := range targets {
		if slices.Contains(rivals, t) {
			return false
		}
	}
	return len(rivals) > 0
}

// oneRivalRefusal is the refusal text of an Experiment under hyp that would
// be one-rival, or "".
func (st *State) oneRivalRefusal(hyp string, targets []string) string {
	q := st.Nodes[hyp].Parent
	rivals := st.LiveRivals(q, hyp)
	if !oneRival(rivals, targets) {
		return ""
	}
	var live []string
	for _, id := range rivals {
		live = append(live, fmt.Sprintf("  %s %s", event.DisplayID(event.KindHypothesis, id), st.hypWord(id)))
	}
	return fmt.Sprintf("live rivals of %s under %s are not targeted:\n%s\npass --target <H-id> for the rival this result would count against, or --target none to proceed without one",
		event.DisplayID(event.KindHypothesis, hyp), event.DisplayID(st.Nodes[q].Kind, q), strings.Join(live, "\n"))
}

// notRivalWarnings names each of an Experiment's parent and targets that is
// not a live rival (NotRivals, the rule D1 reads) and how to fix it.
func (st *State) notRivalWarnings(parent string, targets []string) []string {
	hid := func(id string) string { return event.DisplayID(event.KindHypothesis, id) }
	var out []string
	for _, h := range st.NotRivals(append([]string{parent}, targets...)) {
		role := "target"
		if h == parent {
			role = "parent"
		}
		line := fmt.Sprintf("D1: %s %s is %s", role, hid(h), st.hypWord(h))
		switch st.hypWord(h) {
		case HypContradicted:
			var rs []string
			for _, r := range st.Contradictions(h) {
				rs = append(rs, event.DisplayID(event.KindResult, r))
			}
			line += fmt.Sprintf(" by %s; if that was wrong, hyp retract <R-id> reopens it; otherwise supersede %s and test the new Hypothesis", strings.Join(rs, ", "), hid(h))
		case HypSuperseded:
			var hs []string
			for _, s := range st.Superseders(h) {
				hs = append(hs, hid(s))
			}
			line += fmt.Sprintf(" by %s; use %s instead", strings.Join(hs, ", "), strings.Join(hs, " or "))
		default:
			line += "; it is no longer a rival"
		}
		out = append(out, line)
	}
	return out
}

// AdmitResult lists what admitting Result p would break: a closed
// Experiment; evidence under .tree/evidence/ outside its Experiment's own
// run and pinned records; a pinned record that no longer hashes as pinned;
// then the recipe pin and the rubric pin, each refused when its bundle no
// longer loads or hashes differently (ADR 0018). f is what the CLI loaded;
// refHash maps each bundle ref an evidence pointer names to the hash it
// pinned.
func (st *State) AdmitResult(p event.ResultRecord, f ResultFacts, refHash map[string]string) []Finding {
	exp := st.Nodes[p.Parent]
	if exp == nil || exp.Kind != event.KindExperiment {
		return nil
	}
	display := event.DisplayID(exp.Kind, p.Parent)
	refuse := func(format string, a ...any) []Finding {
		return []Finding{{FindRefuse, fmt.Sprintf(format, a...)}}
	}
	switch st.expStatus(p.Parent) {
	case ExpWithdrawn:
		return refuse("%s is withdrawn; add a new Experiment for this Result", display)
	case ExpSuperseded:
		var by []string
		for _, e := range st.Superseders(p.Parent) {
			by = append(by, event.DisplayID(event.KindExperiment, e))
		}
		return refuse("%s is superseded by %s; record the Result there, or hyp unlink %s supersedes %s to reopen %s",
			display, strings.Join(by, ", "), by[0], display, display)
	}
	for _, ptr := range p.Evidence {
		if outsideRun(ptr, display, exp.Records) {
			return refuse("bad evidence: %s is a record outside %s's run and its pinned records; pin it with --records on a new Experiment", ptr, display)
		}
	}
	if len(f.Changed) > 0 {
		return refuse("bad records: %s changed since %s pinned it; add a new Experiment that pins the records as they are", f.Changed[0], display)
	}
	remedy := "restore it, or add a new Experiment or a new recipe version"
	if exp.Recipe != event.RecipeNone {
		if f.Recipe.Err != nil {
			return refuse("recipe %s of %s no longer loads (%v); %s", exp.Recipe, display, herr.Text(f.Recipe.Err), remedy)
		}
		if f.Recipe.Hash != exp.RecipeHash {
			return refuse("recipe %s changed since %s pinned it (recorded %s, now %s); %s", exp.Recipe, display, exp.RecipeHash, f.Recipe.Hash, remedy)
		}
		if h, ok := refHash[exp.Recipe]; ok && h != exp.RecipeHash {
			return refuse("recipe %s changed while recording: its evidence hashed %s, %s pinned %s; record again", exp.Recipe, h, display, exp.RecipeHash)
		}
	}
	if exp.RubricHash != "" {
		remedy := "restore it, or add a new Experiment or a new rubric version"
		if f.Rubric.Err != nil {
			return refuse("rubric %s of %s no longer loads (%v); %s", exp.Rubric, display, herr.Text(f.Rubric.Err), remedy)
		}
		if f.Rubric.Hash != exp.RubricHash {
			return refuse("rubric %s changed since %s pinned it (recorded %s, now %s); %s", exp.Rubric, display, exp.RubricHash, f.Rubric.Hash, remedy)
		}
	}
	return nil
}

// EvidenceDir is the directory, relative to the store root, that holds the
// files of every run.
const EvidenceDir = ".tree/evidence/"

// outsideRun reports whether evidence pointer ptr names a file under
// .tree/evidence/ that is neither in the run of the Experiment shown as
// display, .tree/evidence/<display>/, nor one of its pinned records.
func outsideRun(ptr, display string, records []event.SubjectFile) bool {
	clean := path.Clean(ptr)
	if !strings.HasPrefix(clean, EvidenceDir) || strings.HasPrefix(clean, EvidenceDir+display+"/") {
		return false
	}
	return !slices.ContainsFunc(records, func(r event.SubjectFile) bool { return r.Path == clean })
}

// AdmitSelect lists what admitting a select of Hypothesis h on question q
// would break, then its warnings. The select is admitted only when the ledger
// owes it: a select of h, or a choice on q with h a candidate. A refusal
// names what is owed instead (SelectBlock). h must be a resolved id.
// It reads derived state, so Validate does not repeat it: replay still
// applies a select in a merged log.
func (p Projection) AdmitSelect(q, h string) []Finding {
	owed, choose := p.SelectOwed(q)
	if !slices.Contains(owed, h) {
		text := p.SelectBlock(q)
		switch {
		case len(owed) == 1:
			text = fmt.Sprintf("%s owes a select of %s, the only candidate, not %s", p.label(q), p.label(owed[0]), p.label(h))
		case choose:
			text = fmt.Sprintf("%s is not a candidate of %s; the owner chooses between %s", p.label(h), p.label(q), p.labels(owed))
		}
		return []Finding{{FindRefuse, text}}
	}
	return nil
}

// SelectOwed returns the Hypotheses a select of question q may name now:
// the lone candidate when the ledger owes its select, or every candidate,
// with choose true, when it owes the owner's choice. It is empty when the
// ledger owes neither.
func (p Projection) SelectOwed(q string) (hyps []string, choose bool) {
	for _, o := range p.Owed() {
		switch {
		case o.Question != q:
		case o.Kind == OwedSelect:
			return []string{o.Key}, false
		case o.Kind == OwedChoose:
			return p.Questions[q].Candidates, true
		}
	}
	return nil, false
}

// SelectBlock names why question q owes no select or choice: the first thing
// its turn waits on, or else q's state.
func (p Projection) SelectBlock(q string) string {
	if t := p.Turn(q); t.Kind == TurnWaits {
		switch w := t.Waits[0]; w.Kind {
		case WaitRunner:
			return fmt.Sprintf("%s owes the runner first: %s", p.label(q), p.owedText(w.Item))
		case WaitReading:
			return fmt.Sprintf("open reading %s waits for the owner; its answer can move %s", p.keyLabel(w.Key), p.labels(w.Reach))
		case WaitReferred:
			return fmt.Sprintf("referred limit %s waits for the owner: an accept releases the select, a return sends its case to run", p.keyLabel(w.Key))
		case WaitPremise:
			t := p.Nodes[w.Premise.To].Parent
			return fmt.Sprintf("%s assumes %s, in %s %s; decide %s first", p.label(w.Premise.From), p.label(w.Premise.To), p.label(t), p.Questions[t].State, p.label(t))
		case WaitSubquestion:
			return fmt.Sprintf("%s: %s has a sub-question not answered", ReasonAwaitingSubquestions, p.label(w.Key))
		}
	}
	qs := p.Questions[q]
	state := string(qs.State)
	if qs.Reason != "" {
		state += "/" + qs.Reason
	}
	switch {
	case qs.State == DQAnswered:
		return fmt.Sprintf("%s is answered; hyp unselect it first to change the choice", p.label(q))
	case len(qs.Candidates) == 0:
		return fmt.Sprintf("%s has no candidate (%s)", p.label(q), state)
	}
	return fmt.Sprintf("%s owes no select (%s)", p.label(q), state)
}

// owedText names one owed item for a refusal.
func (p Projection) owedText(o Owed) string {
	switch o.Kind {
	case OwedLimit, OwedReading:
		return string(o.Kind) + " " + p.keyLabel(o.Key)
	case OwedPremise:
		t := p.Nodes[o.To].Parent
		rec, _ := p.CurrentSelect(t)
		return fmt.Sprintf("premise %s assumes %s, but %s selected %s", p.label(o.Key), p.label(o.To), p.label(t), p.label(rec.Hyp))
	}
	return string(o.Kind) + " " + p.label(o.Key)
}

// label is a node's display id, with its alias in parentheses when it has one.
func (p Projection) label(id string) string {
	n := p.Nodes[id]
	if n == nil {
		return id
	}
	s := event.DisplayID(n.Kind, id)
	if n.Alias != "" {
		s += " (" + n.Alias + ")"
	}
	return s
}

func (p Projection) labels(ids []string) string {
	var out []string
	for _, id := range ids {
		out = append(out, p.label(id))
	}
	return strings.Join(out, ", ")
}

// keyLabel is a limit or reading key as R-…#n or R-…#rn, or a line as G-…#n.
func (p Projection) keyLabel(k string) string {
	kind := event.KindResult
	if n := p.Nodes[event.KeyResult(k)]; n != nil {
		kind = n.Kind
	}
	return event.DisplayKey(kind, k)
}
