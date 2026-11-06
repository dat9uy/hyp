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
	if p.Kind != event.KindExperiment && (p.Recipe != "" || p.RecipeHash != "" || len(p.Targets) > 0 || p.DecidedBy != "") {
		return errors.New("unknown field: recipe, recipe_hash, targets and decided_by are for Experiments only")
	}

	if p.Kind == event.KindGoal {
		if p.Parent != "" {
			return errors.New("wrong parent kind: Goal must have no parent")
		}
		if f.GoalID != "" {
			return errors.New("wrong parent kind: only one Goal per log")
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
		event.KindExperiment:     {event.KindHypothesis},
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
	if len(p.Verdicts) == 0 {
		return errors.New("missing field: verdicts needs at least one entry")
	}
	for h, v := range p.Verdicts {
		if h != exp.Parent && !contains(exp.Targets, h) {
			return fmt.Errorf("bad verdict: %s is not the Experiment's hypothesis or a target", h)
		}
		if !event.Verdicts[v] {
			return fmt.Errorf("bad verdict: %q", v)
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
	for _, l := range p.Limits {
		if err := textLimit("limit", l, 1, event.MaxLimitItem); err != nil {
			return err
		}
	}
	q := f.Nodes[exp.Parent].Parent
	for _, k := range p.Closes {
		if err := f.checkLimitRef(k, q); err != nil {
			return err
		}
	}
	return nil
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

// resultQuestion returns the DesignQuestion that Result r belongs to: the
// parent of its Experiment's Hypothesis.
func (f *Facts) resultQuestion(r string) string {
	exp := f.Nodes[f.Nodes[r].Parent]
	return f.Nodes[exp.Parent].Parent
}

// checkLimitRef reports why key does not name a limit of a non-retracted
// Result under question q.
func (f *Facts) checkLimitRef(key, q string) error {
	r, n, ok := event.ParseLimitKey(key)
	if !ok {
		return fmt.Errorf("bad limit ref: %q (want <result>#<n>)", key)
	}
	node := f.Nodes[r]
	if node == nil || node.Kind != event.KindResult {
		return fmt.Errorf("bad limit ref: %s is not a Result", r)
	}
	if f.Retracted[r] {
		return fmt.Errorf("bad limit ref: %s is retracted", r)
	}
	if n > len(node.Limits) {
		return fmt.Errorf("bad limit ref: %s has %d limits", r, len(node.Limits))
	}
	if f.resultQuestion(r) != q {
		return fmt.Errorf("bad limit ref: %s is not under %s", r, q)
	}
	return nil
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
	rec, ok := sel.Selects[p.Select]
	if !ok {
		return fmt.Errorf("bad confirm: %s is not a dq.select on %s", p.Select, ev.Node)
	}
	if p.Role == event.RoleVerifier && rec.Actor == ev.Actor {
		return fmt.Errorf("separation of duties: verifier %s also made the select", ev.Actor)
	}
	if len(p.Accepts) > 0 && (p.Role != event.RoleOwner || p.Verdict != event.ConfirmSufficient) {
		return errors.New("bad confirm: only a sufficient owner confirm accepts limits")
	}
	for _, k := range p.Accepts {
		if err := f.checkLimitRef(k, ev.Node); err != nil {
			return err
		}
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
