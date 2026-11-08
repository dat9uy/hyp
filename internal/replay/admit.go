package replay

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
)

// RecipeFacts is what the CLI loads about an Experiment's recipe for
// admission. Replay reads no files, so the caller fills it in.
type RecipeFacts struct {
	Err        error    // why the ref does not load; nil when it loads
	Instrument bool     // it has an Evidence or Standalone command
	HasCases   bool     // it has known answers
	Hash       string   // its current bundle hash
	GoWarnings []string // recipe.GoFileWarnings of its directory
	// Placeholder lists its files that still hold the template's
	// placeholder marker (recipe.HasPlaceholder).
	Placeholder []string
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
// instrument with no known answers, a measured recipe with
// --decided-by, a recipe that holds no measure, then, when st is not nil,
// one-rival and the D1 and Go-file warnings. f is ignored when p.Recipe is
// none. noTarget is --target none, which waives one-rival. With st nil only
// the rules that read no tree run, so the caller can refuse before it loads
// the store; p.Parent and p.Targets must be resolved ids otherwise.
func (st *State) AdmitExp(p event.NodeAdd, f RecipeFacts, noTarget bool) []Finding {
	var out []Finding
	measured := false
	if p.Recipe != event.RecipeNone {
		if len(f.Placeholder) > 0 {
			out = append(out, Finding{FindRefuse, fmt.Sprintf("recipe %s still holds the template placeholder (%s) in %s; replace it before pinning the bundle",
				p.Recipe, recipe.Placeholder, strings.Join(f.Placeholder, ", "))})
		}
		if f.Instrument && !f.HasCases {
			out = append(out, Finding{FindRefuse, fmt.Sprintf("recipe %s is an instrument with no known answers; add known-answers/NAME/ cases before an Experiment uses it", p.Recipe)})
		}
		measured = f.Instrument
	}
	if measured && p.DecidedBy != "" {
		out = append(out, Finding{FindBadArgs, fmt.Sprintf("--decided-by %s contradicts recipe %s, an instrument whose checks decide the verdicts; drop --decided-by", p.DecidedBy, p.Recipe)})
	}
	if !measured && p.DecidedBy == "" {
		out = append(out, Finding{FindRefuse, NoMeasure(p.Recipe)})
	}
	if st == nil {
		return out
	}
	// The rival rules read a Hypothesis's siblings. Any other parent is
	// refused when the event is validated.
	if parent := st.Nodes[p.Parent]; parent == nil || parent.Kind != event.KindHypothesis {
		return out
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

// NoMeasure is the refusal of an Experiment whose recipe is none or not an
// instrument and that does not say its verdicts are decided by reading.
func NoMeasure(ref string) string {
	return fmt.Sprintf("recipe %s holds no measure; a verdict a script decides needs an instrument recipe (RECIPE-FORMAT, Verdict scripts), and a verdict decided by reading says so with --decided-by reading", ref)
}

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

// AdmitResult lists what recording Result p would break: its Experiment is
// superseded or withdrawn, or its recipe no longer loads, changed since the Experiment pinned it, or hashed
// differently in p's evidence (refHash: evidence ref → the hash its
// pointers recorded), so one Result never pins two hashes for one ref.
// f is ignored when the Experiment's recipe is none.
func (st *State) AdmitResult(p event.ResultRecord, f RecipeFacts, refHash map[string]string) []Finding {
	exp := st.Nodes[p.Parent]
	if exp == nil || exp.Kind != event.KindExperiment {
		return nil
	}
	switch display := event.DisplayID(exp.Kind, p.Parent); st.expStatus(p.Parent) {
	case ExpWithdrawn:
		return []Finding{{FindRefuse, display + " is withdrawn; add a new Experiment for this Result"}}
	case ExpSuperseded:
		var by []string
		for _, e := range st.Superseders(p.Parent) {
			by = append(by, event.DisplayID(event.KindExperiment, e))
		}
		return []Finding{{FindRefuse, fmt.Sprintf("%s is superseded by %s; record the Result there, or hyp unlink %s supersedes %s to reopen %s",
			display, strings.Join(by, ", "), by[0], display, display)}}
	}
	if exp.Recipe == event.RecipeNone {
		return nil
	}
	remedy := "restore it, or add a new Experiment or a new recipe version"
	display := event.DisplayID(exp.Kind, p.Parent)
	refuse := func(format string, a ...any) []Finding {
		return []Finding{{FindRefuse, fmt.Sprintf(format, a...)}}
	}
	if f.Err != nil {
		return refuse("recipe %s of %s no longer loads (%v); %s", exp.Recipe, display, herr.Text(f.Err), remedy)
	}
	if f.Hash != exp.RecipeHash {
		return refuse("recipe %s changed since %s pinned it (recorded %s, now %s); %s", exp.Recipe, display, exp.RecipeHash, f.Hash, remedy)
	}
	if h, ok := refHash[exp.Recipe]; ok && h != exp.RecipeHash {
		return refuse("recipe %s changed while recording: its evidence hashed %s, %s pinned %s; record again", exp.Recipe, h, display, exp.RecipeHash)
	}
	return nil
}

// AdmitConfirm lists what admitting confirm p on question q would break: q has
// no current selection, or it is a sufficient owner confirm that leaves a
// gating limit unaccepted. p.Accepts must be resolved limit keys. Both rules
// read derived state, so Validate does not repeat them.
func (p Projection) AdmitConfirm(q string, c event.Confirm) []Finding {
	if _, ok := p.CurrentSelect(q); !ok {
		return []Finding{{FindRefuse, "no current selection"}}
	}
	if c.Role != event.RoleOwner || c.Verdict != event.ConfirmSufficient {
		return nil
	}
	var open []string
	for _, l := range p.LimitsOf(q) {
		if l.Gating && !slices.Contains(c.Accepts, l.Key) {
			open = append(open, fmt.Sprintf("  %s#%s %s", event.DisplayID(event.KindResult, l.Result), strings.TrimPrefix(l.Key, l.Result+"#"), l.Text))
		}
	}
	if len(open) == 0 {
		return nil
	}
	return []Finding{{FindRefuse, fmt.Sprintf("open limits under %s are not accepted:\n%s\npass --accept <R-id>#<n> for each limit you accept, or close it first with hyp record --closes <R-id>#<n>",
		event.DisplayID(p.Nodes[q].Kind, q), strings.Join(open, "\n"))}}
}

// AdmitSelect lists what admitting a select of Hypothesis h on question q
// would break, then its warnings. The select is admitted only when the ledger
// owes it: a select of h, or a choice on q with h a candidate. A refusal
// names what is owed instead (SelectBlock). An admitted select warns of the
// open limits under q whose Result did not judge h, since a limit on a
// rival's Result may also apply to the chosen one. h must be a resolved id.
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
	var other []string
	for _, l := range p.LimitsOf(q) {
		if l.State == LimitOpen && !slices.Contains(l.Hyps, h) {
			other = append(other, p.keyLabel(l.Key))
		}
	}
	if len(other) == 0 {
		return nil
	}
	noun := "limits"
	if len(other) == 1 {
		noun = "limit"
	}
	return []Finding{{FindWarn, fmt.Sprintf("%d open %s judged other Hypotheses: %s", len(other), noun, strings.Join(other, ", "))}}
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

// SelectBlock names why question q owes no select or choice, in the order the
// ledger reads them: the first item the runner owes on q, a reading on q that
// waits for the owner, the premise hold, the lone candidate's open
// sub-question, or else q's state.
func (p Projection) SelectBlock(q string) string {
	for _, o := range p.Owed() {
		if o.Question == q && o.Who == OwedRunner {
			return fmt.Sprintf("%s owes the runner first: %s", p.label(q), p.owedText(o))
		}
	}
	for _, l := range p.Readings {
		if l.Question == q && l.By == "" {
			return fmt.Sprintf("open reading %s waits for the owner", p.keyLabel(l.Key))
		}
	}
	if pr := p.PremiseHold(q); pr != nil {
		t := p.Nodes[pr.To].Parent
		return fmt.Sprintf("%s assumes %s, in %s %s; decide %s first", p.label(pr.From), p.label(pr.To), p.label(t), p.Questions[t].State, p.label(t))
	}
	qs := p.Questions[q]
	if len(qs.Candidates) == 1 && p.HasOpenSubquestion(qs.Candidates[0]) {
		return fmt.Sprintf("%s: %s has a sub-question not answered", ReasonAwaitingSubquestions, p.label(qs.Candidates[0]))
	}
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
		return o.Kind + " " + p.keyLabel(o.Key)
	case OwedPremise:
		t := p.Nodes[o.To].Parent
		rec, _ := p.CurrentSelect(t)
		return fmt.Sprintf("premise %s assumes %s, but %s selected %s", p.label(o.Key), p.label(o.To), p.label(t), p.label(rec.Hyp))
	}
	return o.Kind + " " + p.label(o.Key)
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

// keyLabel is a limit or reading key as R-…#n or R-…#rn.
func (p Projection) keyLabel(k string) string {
	r := event.KeyResult(k)
	return event.DisplayID(event.KindResult, r) + k[len(r):]
}
