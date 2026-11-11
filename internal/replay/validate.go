package replay

import (
	"errors"
	"fmt"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/dat9uy/hyp/internal/event"
)

var (
	ulidRe   = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)
	actorRe  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)
	aliasRe  = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
	recipeRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*@[A-Za-z0-9._-]+$`)
	hexRe    = regexp.MustCompile(`^[0-9a-f]+$`)
)

// Validate reports why ev may not be applied to st, or nil when it may. The
// CLI calls it before appending; Replay calls it on every event.
func Validate(st *State, ev event.Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid event: %v", r)
		}
	}()
	if err := validateEnvelope(&st.Facts, ev); err != nil {
		return err
	}
	return rules[ev.Type].validate(&st.Facts, ev)
}

func validateEnvelope(f *Facts, ev event.Event) error {
	if ev.V != event.Version {
		return fmt.Errorf("unsupported version: v=%d", ev.V)
	}
	if !event.EventTypes[ev.Type] {
		return fmt.Errorf("unknown event type: %q", ev.Type)
	}
	if !ulidRe.MatchString(ev.ID) {
		return fmt.Errorf("missing field: id must be a 26-character ULID, got %q", ev.ID)
	}
	if f.SeenIDs[ev.ID] {
		return errors.New("duplicate id: " + ev.ID)
	}
	if !actorRe.MatchString(ev.Actor) {
		return fmt.Errorf("invalid actor: %q", ev.Actor)
	}
	if !strings.HasSuffix(ev.TS, "Z") {
		return fmt.Errorf("ts not UTC: %q", ev.TS)
	}
	if _, err := time.Parse(time.RFC3339Nano, ev.TS); err != nil {
		return fmt.Errorf("ts not RFC3339Nano: %q", ev.TS)
	}
	if ev.Node == "" {
		return errors.New("missing field: node")
	}
	return textLimit("why", ev.Why, 1, event.MaxWhy)
}

func (f *Facts) validateNodeAdd(ev event.Event) error {
	var p event.NodeAdd
	if err := decode(ev, &p); err != nil {
		return err
	}
	if ev.Node != ev.ID {
		return errors.New("node must equal id on node.add")
	}
	if !event.Kinds[p.Kind] || p.Kind == event.KindResult {
		return fmt.Errorf("unknown kind: %q", p.Kind)
	}
	if err := textLimit("title", p.Title, 1, event.MaxTitle); err != nil {
		return err
	}
	maxBody := event.MaxText
	if p.Kind == event.KindGoal {
		maxBody = event.MaxGoalBody // a Goal's body holds the owner's met rule
	}
	if err := textLimit("body", p.Body, 0, maxBody); err != nil {
		return err
	}
	if p.Alias != "" {
		if !aliasRe.MatchString(p.Alias) {
			return fmt.Errorf("invalid alias: %q", p.Alias)
		}
		if _, used := f.Aliases[p.Alias]; used {
			return fmt.Errorf("alias already used: %q", p.Alias)
		}
	}
	if p.Kind != event.KindExperiment && (p.Recipe != "" || p.RecipeHash != "" || len(p.Targets) > 0 || p.DecidedBy != "" || len(p.Aims) > 0 ||
		p.Rubric != "" || p.RubricHash != "" || len(p.Records) > 0) {
		return errors.New("unknown field: recipe, recipe_hash, targets, decided_by, rubric, rubric_hash, records and aims are for Experiments only")
	}

	if p.Kind == event.KindGoal {
		if p.Parent != "" {
			return errors.New("wrong parent kind: Goal must have no parent")
		}
		return nil
	}
	parent, err := f.need(p.Parent, "")
	if err != nil {
		return err
	}
	allowed := map[string][]string{
		event.KindDesignQuestion: {event.KindGoal, event.KindHypothesis},
		event.KindHypothesis:     {event.KindDesignQuestion},
		event.KindExperiment:     {event.KindHypothesis, event.KindDesignQuestion},
	}[p.Kind]
	if !contains(allowed, parent.Kind) {
		return fmt.Errorf("wrong parent kind: %s under %s", p.Kind, parent.Kind)
	}
	if p.Kind != event.KindExperiment {
		return nil
	}

	switch {
	case p.Recipe == "":
		return errors.New("missing field: recipe")
	case p.Recipe == event.RecipeNone:
		if p.RecipeHash != "" {
			return errors.New("unknown field: recipe_hash with recipe none")
		}
	case !recipeRe.MatchString(p.Recipe):
		return fmt.Errorf("invalid recipe: %q (want name@version or none)", p.Recipe)
	case p.RecipeHash == "":
		return errors.New("missing field: recipe_hash")
	case !hexRe.MatchString(p.RecipeHash):
		return fmt.Errorf("invalid recipe_hash: %q", p.RecipeHash)
	}
	if p.DecidedBy != "" && p.DecidedBy != event.DecidedReading {
		return fmt.Errorf("invalid decided_by: %q (want %s)", p.DecidedBy, event.DecidedReading)
	}
	if err := checkRubric(p); err != nil {
		return err
	}
	if err := checkRecords(p.Records); err != nil {
		return err
	}
	if parent.Kind == event.KindDesignQuestion {
		// An Experiment under a question gives no verdict: it measures, so
		// it pins a recipe and has no targets and no rubric (ADR 0018).
		if len(p.Targets) > 0 || p.Rubric != "" || p.RubricHash != "" || p.DecidedBy != "" {
			return errors.New("bad experiment: an Experiment under a DesignQuestion has no targets and no rubric; it measures")
		}
		if p.Recipe == event.RecipeNone {
			return errors.New("bad experiment: an Experiment under a DesignQuestion measures, so it pins a recipe")
		}
		return f.checkAims(p.Aims)
	}
	seen := map[string]bool{}
	for _, t := range p.Targets {
		tn, err := f.need(t, event.KindHypothesis)
		if err != nil {
			return fmt.Errorf("bad target: %w", err)
		}
		if t == p.Parent {
			return fmt.Errorf("bad target: %s is the Experiment's own parent", t)
		}
		if tn.Parent != parent.Parent {
			return fmt.Errorf("bad target: %s is not a sibling of %s", t, p.Parent)
		}
		if seen[t] {
			return fmt.Errorf("bad target: %s listed twice", t)
		}
		seen[t] = true
	}
	return f.checkAims(p.Aims)
}

// checkRubric reports why an Experiment's rubric fields are malformed: the
// rubric is reading or a NAME@VERSION pinned by a hex rubric_hash, and the
// older decided_by never comes with it.
func checkRubric(p event.NodeAdd) error {
	switch {
	case p.Rubric == "":
		if p.RubricHash != "" {
			return errors.New("bad rubric: rubric_hash without a rubric bundle")
		}
		return nil
	case p.DecidedBy != "":
		return errors.New("bad rubric: rubric and decided_by both given; decided_by is the older spelling of rubric reading")
	case p.Rubric == event.RubricReading:
		if p.RubricHash != "" {
			return errors.New("bad rubric: rubric_hash with rubric reading")
		}
		return nil
	case !recipeRe.MatchString(p.Rubric):
		return fmt.Errorf("invalid rubric: %q (want name@version or reading)", p.Rubric)
	case p.RubricHash == "":
		return errors.New("bad rubric: missing rubric_hash")
	case !hexRe.MatchString(p.RubricHash):
		return fmt.Errorf("bad rubric: invalid rubric_hash %q", p.RubricHash)
	}
	return nil
}

// checkRecords reports why an Experiment's pinned records are malformed.
func checkRecords(files []event.SubjectFile) error {
	if len(files) > event.MaxRecords {
		return fmt.Errorf("over limit: records has %d files (max %d)", len(files), event.MaxRecords)
	}
	seen := map[string]bool{}
	for _, f := range files {
		if err := textLimit("records path", f.Path, 1, event.MaxText/2); err != nil {
			return fmt.Errorf("bad records: %w", err)
		}
		if seen[f.Path] {
			return fmt.Errorf("bad records: %s named twice", f.Path)
		}
		seen[f.Path] = true
		if !hexRe.MatchString(f.SHA256) {
			return fmt.Errorf("bad records: %q is not a sha256", f.SHA256)
		}
	}
	return nil
}

// checkAims reports why an Experiment may not aim at keys: each must name a
// limit, or a reading the owner has closed, of a non-retracted Result, once.
func (f *Facts) checkAims(keys []string) error {
	if len(keys) > event.MaxLimits {
		return fmt.Errorf("over limit: aims has %d items (max %d)", len(keys), event.MaxLimits)
	}
	seen := map[string]bool{}
	for _, k := range keys {
		if seen[k] {
			return fmt.Errorf("bad aim: %s listed twice", k)
		}
		seen[k] = true
		if key, _ := event.ParseKey(k); key.Kind == event.KeyReading {
			if err := f.checkKey(k, event.KeyReading, ""); err != nil {
				return fmt.Errorf("bad aim: %w", err)
			}
			if !f.ReadingClosed(k) {
				return fmt.Errorf("bad aim: reading %s is open; only an answered reading can be applied", k)
			}
			continue
		}
		if err := f.checkKey(k, event.KeyLimit, ""); err != nil {
			return fmt.Errorf("bad aim: %w", err)
		}
	}
	return nil
}

func (f *Facts) validateResult(ev event.Event) error {
	var p event.ResultRecord
	if err := decode(ev, &p); err != nil {
		return err
	}
	if ev.Node != ev.ID {
		return errors.New("node must equal id on result.record")
	}
	exp, err := f.need(p.Parent, "")
	if err != nil {
		return err
	}
	if exp.Kind != event.KindExperiment {
		return fmt.Errorf("wrong parent kind: Result under %s", exp.Kind)
	}
	// A Result gives a verdict on each of its Experiment's targets and on
	// nothing else (ADR 0018); one under an Experiment with no targets
	// measures and gives none.
	targets := targetsOf(f.Nodes, exp)
	if len(targets) == 0 && len(p.Verdicts) > 0 {
		return fmt.Errorf("bad verdicts: %s has no targets; it measures", event.DisplayID(event.KindExperiment, exp.ID))
	}
	for h, v := range p.Verdicts {
		if !contains(targets, h) {
			return fmt.Errorf("bad verdict: %s is not the Experiment's hypothesis or a target", h)
		}
		if !event.Verdicts[v] {
			return fmt.Errorf("bad verdict: %q", v)
		}
	}
	for _, h := range targets {
		if _, ok := p.Verdicts[h]; !ok {
			return fmt.Errorf("bad verdicts: %s targets %s; give it a verdict, inconclusive when the run could not decide",
				event.DisplayID(event.KindExperiment, exp.ID), event.DisplayID(event.KindHypothesis, h))
		}
	}
	if err := textLimit("observation", p.Observation, 1, event.MaxText); err != nil {
		return err
	}
	if len(p.Evidence) > event.MaxEvidence {
		return fmt.Errorf("over limit: evidence has %d items (max %d)", len(p.Evidence), event.MaxEvidence)
	}
	for _, e := range p.Evidence {
		if err := textLimit("evidence item", e, 1, event.MaxEvidenceItem); err != nil {
			return err
		}
	}
	for _, e := range slices.Sorted(maps.Keys(p.EvidenceHashes)) {
		h := p.EvidenceHashes[e]
		if !contains(p.Evidence, e) {
			return fmt.Errorf("bad evidence_hashes: %s is not in evidence", e)
		}
		if ref, ok := event.EvidenceRef(e); !ok || path.Clean(e) == "recipes/"+ref.String() {
			return fmt.Errorf("bad evidence_hashes: %s is not in a versioned recipe (recipes/NAME@VERSION/FILE)", e)
		}
		if !hexRe.MatchString(h) {
			return fmt.Errorf("invalid evidence_hashes value for %s: %q", e, h)
		}
	}
	if len(p.Limits) > event.MaxLimits {
		return fmt.Errorf("over limit: limits has %d items (max %d)", len(p.Limits), event.MaxLimits)
	}
	if len(p.Fingerprints) > event.MaxFingerprints {
		return fmt.Errorf("over limit: fingerprints has %d items (max %d)", len(p.Fingerprints), event.MaxFingerprints)
	}
	for i, h := range p.Fingerprints {
		if !event.IsSHA256(h) {
			return fmt.Errorf("bad fingerprints: %q is not a sha256", h)
		}
		if i > 0 && h <= p.Fingerprints[i-1] {
			return errors.New("bad fingerprints: not sorted and unique")
		}
	}
	for _, l := range p.Limits {
		if err := textLimit("limit", l, 1, event.MaxLimitItem); err != nil {
			return err
		}
	}
	if len(p.Readings) > event.MaxLimits {
		return fmt.Errorf("over limit: readings has %d items (max %d)", len(p.Readings), event.MaxLimits)
	}
	for _, l := range p.Readings {
		if err := textLimit("reading", l.Text, 1, event.MaxReadingItem); err != nil {
			return err
		}
		if l.Of == "" {
			continue
		}
		// A reading of an answer reads one the owner gave: it narrows that
		// answer and never reopens it, so the chain points back in merge order.
		if err := f.checkKey(l.Of, event.KeyReading, ""); err != nil {
			return fmt.Errorf("bad reading of: %w", err)
		}
		if !f.ReadingClosed(l.Of) {
			return fmt.Errorf("bad reading of: %s is open; the owner has not answered it", l.Of)
		}
	}
	q := questionOf(f.Nodes, exp)
	for _, k := range p.Closes {
		// A limit under another question may be closed by a Result whose
		// Experiment aims at it.
		scope := q
		if contains(exp.Aims, k) {
			scope = ""
		}
		if err := f.checkKey(k, event.KeyLimit, scope); err != nil {
			return err
		}
	}
	if len(p.Applies) > event.MaxLimits {
		return fmt.Errorf("over limit: applies has %d items (max %d)", len(p.Applies), event.MaxLimits)
	}
	for _, k := range p.Applies {
		if key, ok := event.ParseKey(k); !ok || key.Kind != event.KeyReading {
			return fmt.Errorf("bad applies: %q (want <result>#r<n>)", k)
		}
		if !contains(exp.Aims, k) {
			return fmt.Errorf("bad applies: %s is not among %s's aims", k, p.Parent)
		}
	}
	return nil
}

// validateSubject is the replay rule of hypothesis.subject. The latest tie
// stands until a non-retracted Result's fingerprints hold it; after that
// the bytes ran as this Hypothesis, and a different subject is a different
// Hypothesis, added and superseding this one.
func (f *Facts) validateSubject(ev event.Event) error {
	var p event.HypothesisSubject
	if err := decode(ev, &p); err != nil {
		return err
	}
	if _, err := f.need(ev.Node, event.KindHypothesis); err != nil {
		return err
	}
	if f.superseded(ev.Node) || f.Abandoned[ev.Node] {
		return fmt.Errorf("bad subject: %s is superseded or abandoned", ev.Node)
	}
	if len(p.Files) == 0 {
		return errors.New("bad subject: no files")
	}
	if len(p.Files) > event.MaxSubjectFiles {
		return fmt.Errorf("over limit: files has %d items (max %d)", len(p.Files), event.MaxSubjectFiles)
	}
	seen := map[string]bool{}
	for _, file := range p.Files {
		if err := textLimit("path", file.Path, 1, event.MaxEvidenceItem); err != nil {
			return err
		}
		if seen[file.Path] {
			return fmt.Errorf("bad subject: %s listed twice", file.Path)
		}
		seen[file.Path] = true
		if !event.IsSHA256(file.SHA256) {
			return fmt.Errorf("bad subject: %q is not a sha256", file.SHA256)
		}
	}
	cur := f.Subjects[ev.Node]
	if slices.Equal(cur, p.Files) {
		return errors.New("bad subject: unchanged")
	}
	if r := f.subjectRan(ev.Node); r != "" {
		return fmt.Errorf("bad subject: %s's subject ran in %s; a different subject is a new Hypothesis: add it and supersede %s", event.DisplayID(event.KindHypothesis, ev.Node), event.DisplayID(event.KindResult, r), event.DisplayID(event.KindHypothesis, ev.Node))
	}
	return nil
}

// subjectRan is the first non-retracted Result whose fingerprints hold every
// file of Hypothesis h's subject, or "".
func (f *Facts) subjectRan(h string) string {
	sub := f.Subjects[h]
	if len(sub) == 0 {
		return ""
	}
	for _, id := range f.Order {
		n := f.Nodes[id]
		if n.Kind == event.KindResult && !f.Retracted[id] && HoldsSubject(n.Fingerprints, sub) {
			return id
		}
	}
	return ""
}

// HoldsSubject reports whether the sorted fingerprints hold every file of sub.
func HoldsSubject(fingerprints []string, sub []event.SubjectFile) bool {
	if len(sub) == 0 {
		return false
	}
	for _, file := range sub {
		if _, ok := slices.BinarySearch(fingerprints, file.SHA256); !ok {
			return false
		}
	}
	return true
}

func (f *Facts) validateAmend(ev event.Event) error {
	var p event.HypothesisAmend
	if err := decode(ev, &p); err != nil {
		return err
	}
	n, err := f.need(ev.Node, event.KindHypothesis)
	if err != nil {
		return err
	}
	if f.superseded(ev.Node) || f.Abandoned[ev.Node] {
		return fmt.Errorf("bad amend: %s is superseded or abandoned", ev.Node)
	}
	if err := textLimit("body", p.Body, 1, event.MaxText); err != nil {
		return err
	}
	if p.Body == n.Body {
		return errors.New("bad amend: body is unchanged")
	}
	return nil
}

// resultQuestion returns the DesignQuestion that Result r belongs to: its
// Experiment's question (questionOf).
func (f *Facts) resultQuestion(r string) string {
	return questionOf(f.Nodes, f.Nodes[f.Nodes[r].Parent])
}

// checkKey reports why s does not name an item of kind want in a
// non-retracted Result: a limit under question q (any question when q is ""),
// or a reading.
func (f *Facts) checkKey(s string, want event.KeyKind, q string) error {
	k, ok := event.ParseKey(s)
	if !ok || k.Kind != want {
		form := map[event.KeyKind]string{event.KeyLimit: "<result>#<n>", event.KeyReading: "<result>#r<n>"}[want]
		return fmt.Errorf("bad %s ref: %q (want %s)", want, s, form)
	}
	node := f.Nodes[k.Result]
	if want == event.KeyLimit && node != nil && node.Kind == event.KindGoal {
		// A line of the Goal: its scope is the Goal, so it is under every
		// question.
		if n := len(f.Lines[k.Result]); k.N > n {
			return fmt.Errorf("bad limit ref: the Goal has %d lines", n)
		}
		return nil
	}
	if node == nil || node.Kind != event.KindResult {
		return fmt.Errorf("bad %s ref: %s is not a Result", want, k.Result)
	}
	if f.Retracted[k.Result] {
		return fmt.Errorf("bad %s ref: %s is retracted", want, k.Result)
	}
	if n := map[event.KeyKind]int{event.KeyLimit: len(node.Limits), event.KeyReading: len(node.Readings)}[want]; k.N > n {
		return fmt.Errorf("bad %s ref: %s has %d %ss", want, k.Result, n, want)
	}
	if want == event.KeyLimit && q != "" && f.resultQuestion(k.Result) != q {
		return fmt.Errorf("bad limit ref: %s is not under %s", k.Result, q)
	}
	return nil
}

// checkClosesReadings reports why a checkpoint may not make its moves on
// readings and limits: the move tables in keyed.go say which move applies
// from which state, and by whom.
func (f *Facts) checkClosesReadings(ev event.Event, p event.Checkpoint) error {
	if p.Role != "" && p.Role != event.RoleOwner {
		return fmt.Errorf("bad checkpoint: role %q (want owner or none)", p.Role)
	}
	if len(p.Closes) > 0 && p.Role != event.RoleOwner {
		return errors.New("bad checkpoint: only an owner closes a reading")
	}
	if len(p.Revises) > 0 && p.Role != event.RoleOwner {
		return errors.New("bad checkpoint: only an owner revises an answer")
	}
	for _, k := range p.Closes {
		if err := f.checkReadingMove("closes", k, ev.Actor); err != nil {
			return err
		}
	}
	for _, k := range p.Revises {
		if slices.Contains(p.Closes, k) {
			return fmt.Errorf("bad revises: %s is also in closes", k)
		}
		if err := f.checkReadingMove("revises", k, ev.Actor); err != nil {
			return err
		}
	}
	return f.checkLimitMoves(ev, p)
}

// limitCloser is the first non-retracted Result that closes limit k, or "".
func (f *Facts) limitCloser(k string) string {
	for _, id := range f.Order {
		n := f.Nodes[id]
		if n.Kind == event.KindResult && !f.Retracted[id] && slices.Contains(n.Closes, k) {
			return id
		}
	}
	return ""
}

func (f *Facts) validateSelect(ev event.Event) error {
	var p event.DQSelect
	if err := decode(ev, &p); err != nil {
		return err
	}
	if _, err := f.need(ev.Node, event.KindDesignQuestion); err != nil {
		return err
	}
	h, err := f.need(p.Hypothesis, event.KindHypothesis)
	if err != nil {
		return err
	}
	if h.Parent != ev.Node {
		return fmt.Errorf("bad select: %s is not a child of %s", p.Hypothesis, ev.Node)
	}
	if f.superseded(p.Hypothesis) || f.Abandoned[p.Hypothesis] {
		return fmt.Errorf("bad select: %s is superseded or abandoned", p.Hypothesis)
	}
	// prev_select is never refused: a mismatch is a conflict (spec §7.4).
	return nil
}

func (f *Facts) validateConfirm(ev event.Event) error {
	var p event.Confirm
	if err := decode(ev, &p); err != nil {
		return err
	}
	if _, err := f.need(ev.Node, event.KindDesignQuestion); err != nil {
		return err
	}
	if p.Role == event.RoleVerifier {
		// Read-only history: replay accepts an old verifier confirm and it
		// applies nothing (spec §4, ADR 0015).
		return nil
	}
	if !event.Roles[p.Role] {
		return fmt.Errorf("bad confirm: unknown role %q", p.Role)
	}
	if !event.ConfirmVerdicts[p.Verdict] {
		return fmt.Errorf("bad confirm: unknown verdict %q", p.Verdict)
	}
	sel := f.Decisions[ev.Node]
	if sel == nil {
		return fmt.Errorf("bad confirm: %s is not a dq.select on %s", p.Select, ev.Node)
	}
	if _, ok := sel.Selects[p.Select]; !ok {
		return fmt.Errorf("bad confirm: %s is not a dq.select on %s", p.Select, ev.Node)
	}
	if len(p.Accepts) > 0 && (p.Role != event.RoleOwner || p.Verdict != event.ConfirmSufficient) {
		return errors.New("bad confirm: only a sufficient owner confirm accepts limits")
	}
	for _, k := range p.Accepts {
		if err := f.checkKey(k, event.KeyLimit, ev.Node); err != nil {
			return err
		}
		if err := f.closedAccept(k); err != nil {
			return err
		}
	}
	return nil
}

// closedAccept refuses an accept of limit k once a Result closed it: the
// case ran, so there is nothing left untested to accept, and the accept
// would change nothing.
func (f *Facts) closedAccept(k string) error {
	if by := f.limitCloser(k); by != "" {
		return fmt.Errorf("bad accepts: %s is closed by %s; its case ran, so there is nothing to accept", f.keyDisplay(k), event.DisplayID(event.KindResult, by))
	}
	return nil
}

func (f *Facts) validateEdge(ev event.Event) error {
	var p event.Edge
	if err := decode(ev, &p); err != nil {
		return err
	}
	if !event.EdgeTypes[p.Type] {
		return fmt.Errorf("unknown edge type: %q", p.Type)
	}
	if p.To == "" {
		return errors.New("missing field: to")
	}
	from, err := f.need(ev.Node, "")
	if err != nil {
		return err
	}
	live := f.liveEdge(ev.Node, p.Type, p.To) >= 0
	if ev.Type == event.TypeEdgeRemove {
		if !live {
			return fmt.Errorf("edge not live: %s %s %s", ev.Node, p.Type, p.To)
		}
		return nil
	}
	if live {
		return fmt.Errorf("duplicate edge: %s %s %s", ev.Node, p.Type, p.To)
	}
	if p.Type == event.EdgeExternalRef {
		return textLimit("to", p.To, 1, event.MaxExternalRef)
	}
	to, err := f.need(p.To, "")
	if err != nil {
		return err
	}
	if p.To == ev.Node {
		return fmt.Errorf("self edge: %s %s itself", ev.Node, p.Type)
	}
	if p.Type == event.EdgeSupersedes {
		switch {
		case from.Kind == event.KindHypothesis && to.Kind == event.KindHypothesis:
			if from.Parent != to.Parent {
				return errors.New("cross-DQ supersedes: both Hypotheses must share a DesignQuestion")
			}
		case from.Kind == event.KindExperiment && to.Kind == event.KindExperiment:
			if f.Nodes[from.Parent].Parent != f.Nodes[to.Parent].Parent {
				return errors.New("cross-DQ supersedes: both Experiments must be under one DesignQuestion")
			}
			if r := f.liveResult(p.To); r != "" {
				return fmt.Errorf("bad supersedes: %s has the non-retracted Result %s", p.To, r)
			}
		default:
			return errors.New("wrong kind: supersedes joins two Hypotheses or two Experiments")
		}
		if f.reachable(p.To, ev.Node, f.supersedesNext) {
			return errors.New("supersedes cycle")
		}
		return nil
	}
	if from.Kind != event.KindHypothesis && from.Kind != event.KindDesignQuestion {
		return fmt.Errorf("wrong kind: depends_on from %s", from.Kind)
	}
	if to.Kind != event.KindHypothesis {
		return fmt.Errorf("wrong kind: depends_on to %s", to.Kind)
	}
	if f.wouldCycle(ev.Node, p.To) {
		return errors.New("depends_on cycle in the combined status graph")
	}
	return nil
}

// wouldCycle reports whether adding depends_on from→to would make the combined
// status graph (spec §7.2) cyclic. A cycle can only run through depends_on, so
// that is whether from is reachable from to along those edges.
func (f *Facts) wouldCycle(from, to string) bool {
	return f.reachable(to, from, f.dependsOn)
}

func (f *Facts) supersedesNext(id string) []string {
	var next []string
	for _, i := range f.EdgesFrom[id] {
		if e := f.Edges[i]; e.Live && e.Type == event.EdgeSupersedes {
			next = append(next, e.To)
		}
	}
	return next
}

func (f *Facts) reachable(start, target string, next func(string) []string) bool {
	seen := map[string]bool{}
	stack := []string{start}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if id == target {
			return true
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		stack = append(stack, next(id)...)
	}
	return false
}

func (f *Facts) liveEdge(from, typ, to string) int {
	for _, i := range f.EdgesFrom[from] {
		if e := f.Edges[i]; e.Live && e.Type == typ && e.To == to {
			return i
		}
	}
	return -1
}

// liveResult is the first non-retracted Result of Experiment exp, or "".
func (f *Facts) liveResult(exp string) string {
	for _, r := range f.Nodes[exp].Children {
		if !f.Retracted[r] {
			return r
		}
	}
	return ""
}

func (f *Facts) superseded(id string) bool {
	return f.SupersededN[id] > 0
}

// need returns node id, refusing a missing node or (when kind != "") a node
// of another kind.
func (f *Facts) need(id, kind string) (*Node, error) {
	n := f.Nodes[id]
	if n == nil {
		return nil, fmt.Errorf("unknown node %s", id)
	}
	if kind != "" && n.Kind != kind {
		return nil, fmt.Errorf("wrong kind: %s is a %s, want %s", id, n.Kind, kind)
	}
	return n, nil
}

func decode(ev event.Event, dst any) error {
	if err := event.DecodePayload(ev, dst); err != nil {
		return fmt.Errorf("bad payload: %v", err)
	}
	return nil
}

func decodeEmpty(ev event.Event) error {
	var p event.Empty
	return decode(ev, &p)
}

func textLimit(field, s string, min, max int) error {
	n := event.Runes(s)
	if n < min {
		return fmt.Errorf("missing field: %s", field)
	}
	if n > max {
		return fmt.Errorf("over limit: %s has %d runes (max %d)", field, n, max)
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
