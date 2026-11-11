package replay

import (
	"fmt"
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// Keyed items are limits, lines and readings: each is created as one entry
// inside its parent's event, a Result or the Goal's lines checkpoint, and is
// addressed as <parent>#<n> or <parent>#r<n>. Its text is fixed, it takes no
// edge, and it is gone when its Result is retracted. Only other events move
// its state. This file is the one place that says which move is admitted
// from which state and by whom (spec §8.2): Validate asks it before a move
// applies, apply records the move in Facts, and the derived limits, lines and
// readings read only what it recorded.

// Hand is what the latest refer, return or accept of a limit key stored:
// Move is HandRefer, HandReturn or HandAccept, Event and Actor the event
// that made it, and Referrer the actor of the latest refer, which an accept
// or a return must differ from.
type Hand struct {
	Move                   HandMove
	Event, Actor, Referrer string
}

// HandMove is a limit key's latest move: HandRefer, HandReturn or
// HandAccept, or "" when none.
type HandMove string

// HandRefer, HandReturn and HandAccept are the values of Hand.Move.
const (
	HandRefer  HandMove = "refer"  // the runner hands the limit to the owner
	HandReturn HandMove = "return" // the owner hands it back: the Goal covers its case
	HandAccept HandMove = "accept" // the owner accepts it: this tree does not run its case
)

// limitMove is one row of the limit move table: the payload field that
// makes the move, the moves the key's latest hand may be for it to apply,
// and whether only an owner checkpoint makes it. A move that answers a
// refer is made by an actor other than the referrer, and every move is
// refused once a Result closed a limit of a Result: its case ran.
type limitMove struct {
	field    string
	from     []HandMove // the latest hand's Move; "" stands for no hand yet
	owner    bool
	answers  bool // the actor differs from the latest referrer
	notRun   bool // refused once a Result closed the limit
	refusal  string
	ownerErr string
	ranErr   string // what follows "closed by R-…" in the notRun refusal
}

var limitMoves = map[HandMove]limitMove{
	HandRefer: {field: "refers", from: []HandMove{"", HandRefer, HandReturn}, notRun: true,
		refusal: "%s is accepted; only the owner reopens it, with returns",
		ranErr:  "; its case ran, so there is nothing to refer"},
	HandAccept: {field: "accepts", from: []HandMove{HandRefer}, owner: true, answers: true, notRun: true,
		refusal:  "%s is not referred; only a referred limit is accepted, so the runner refers it first",
		ownerErr: "bad checkpoint: only an owner accepts a limit", ranErr: "; its case ran, so there is nothing to accept"},
	HandReturn: {field: "returns", from: []HandMove{HandRefer, HandAccept}, owner: true, answers: true, notRun: true,
		refusal:  "%s is not referred or accepted; only a referred or accepted limit is returned",
		ownerErr: "bad checkpoint: only an owner returns a limit"},
}

// The table's rows are read by checkLimitMove and applyLimitMove only.

// checkLimitMove reports why move may not apply to limit key k now, made by
// actor on a checkpoint whose role is owner or not.
func (f *Facts) checkLimitMove(move HandMove, k, actor string, owner bool) error {
	m := limitMoves[move]
	if m.owner && !owner {
		return fmt.Errorf("%s", m.ownerErr)
	}
	if err := f.checkKey(k, event.KeyLimit, ""); err != nil {
		return fmt.Errorf("bad %s: %w", m.field, err)
	}
	if m.notRun && f.Nodes[event.KeyResult(k)].Kind == event.KindResult {
		if by := f.limitCloser(k); by != "" {
			return fmt.Errorf("bad %s: %s is closed by %s"+m.ranErr, m.field, f.keyDisplay(k), event.DisplayID(event.KindResult, by))
		}
	}
	h := f.LimitHand[k]
	if !slices.Contains(m.from, h.Move) {
		return fmt.Errorf("bad %s: "+m.refusal, m.field, f.keyDisplay(k))
	}
	if m.answers && h.Referrer != "" && h.Referrer == actor {
		return fmt.Errorf("separation of duties: %s referred the limit it %s", actor, map[HandMove]string{HandAccept: "accepts", HandReturn: "returns"}[move])
	}
	return nil
}

// applyLimitMove records move on limit key k, made by ev.
func (f *Facts) applyLimitMove(move HandMove, k string, ev event.Event) {
	h := f.LimitHand[k]
	h.Move, h.Event, h.Actor = move, ev.ID, ev.Actor
	if move == HandRefer {
		h.Referrer = ev.Actor
	}
	f.LimitHand[k] = h
}

// checkLimitMoves checks every refer, return and accept one event makes, in
// the order replay applies them, and that no key is named by two of them.
// Each move is checked against the hand before the event. owner is set on an
// answer and on an old owner checkpoint.
func (f *Facts) checkLimitMoves(actor string, owner bool, refers, returns, accepts []string) error {
	moves := []struct {
		move HandMove
		keys []string
	}{{HandRefer, refers}, {HandReturn, returns}, {HandAccept, accepts}}
	for _, mv := range moves {
		if len(mv.keys) > 0 && limitMoves[mv.move].owner && !owner {
			return fmt.Errorf("%s", limitMoves[mv.move].ownerErr)
		}
		if len(mv.keys) > event.MaxLimits {
			return fmt.Errorf("over limit: %s has %d items (max %d)", limitMoves[mv.move].field, len(mv.keys), event.MaxLimits)
		}
	}
	for _, mv := range moves {
		for _, k := range mv.keys {
			for _, other := range moves {
				if mv.move == HandReturn && other.move != HandReturn && slices.Contains(other.keys, k) {
					return fmt.Errorf("bad returns: %s is also in %s", f.keyDisplay(k), limitMoves[other.move].field)
				}
			}
			if err := f.checkLimitMove(mv.move, k, actor, owner); err != nil {
				return err
			}
		}
	}
	return nil
}

// applyLimitMoves records the refers, returns and accepts of ev, in that
// order.
func (f *Facts) applyLimitMoves(ev event.Event, refers, returns, accepts []string) {
	for _, mv := range []struct {
		move HandMove
		keys []string
	}{{HandRefer, refers}, {HandReturn, returns}, {HandAccept, accepts}} {
		for _, k := range mv.keys {
			f.applyLimitMove(mv.move, k, ev)
		}
	}
}

// readingMove is one row of the reading move table. A close answers an open
// reading and a revise replaces the answer of a closed one; both are the
// owner's, under an actor other than the one that recorded the reading.
type readingMove struct {
	field, verb string
	closed      bool // the reading must already be answered
	refusal     string
}

var readingMoves = map[string]readingMove{
	"closes":  {field: "closes", verb: "closes", refusal: "%s is answered; an owner revises an answer with revises"},
	"revises": {field: "revises", verb: "revises", closed: true, refusal: "%s is open; an owner answers an open reading with closes"},
}

// checkReadingMove reports why the move in field may not apply to reading k.
func (f *Facts) checkReadingMove(field, k, actor string) error {
	m := readingMoves[field]
	if key, ok := event.ParseKey(k); ok && key.Kind == event.KeyLimit {
		if field == "closes" {
			return fmt.Errorf("bad closes: %s is a limit; an owner accepts a limit with accepts", k)
		}
		return fmt.Errorf("bad revises: %s is a limit; a limit is closed, accepted or referred, never revised", k)
	}
	if err := f.checkKey(k, event.KeyReading, ""); err != nil {
		return err
	}
	if f.ReadingClosed(k) != m.closed {
		return fmt.Errorf("bad %s: "+m.refusal, m.field, k)
	}
	if f.Nodes[event.KeyResult(k)].Actor == actor {
		return fmt.Errorf("separation of duties: %s recorded the reading it %s", actor, m.verb)
	}
	return nil
}

// applyReadingMoves records the answers an old owner checkpoint gives: a
// close gives a reading its first answer, a revise the answer in force from
// then on, each with the checkpoint's summary.
func (f *Facts) applyReadingMoves(ev event.Event, p event.Checkpoint) {
	for _, k := range p.Closes {
		if len(f.Answers[k]) == 0 {
			f.Answers[k] = []Answer{{By: ev.ID, Text: p.Summary}}
		}
	}
	for _, k := range p.Revises {
		if len(f.Answers[k]) > 0 {
			f.Answers[k] = append(f.Answers[k], Answer{By: ev.ID, Text: p.Summary})
		}
	}
}

// checkAnswers reports why an answer may not give its answers. Whether an
// item is a first answer or a revision is derived from the reading's state,
// so one item is a close of an open reading and a revise of an answered one.
func (f *Facts) checkAnswers(actor string, items []event.KeyText) error {
	if len(items) > event.MaxLimits {
		return fmt.Errorf("over limit: answers has %d items (max %d)", len(items), event.MaxLimits)
	}
	seen := map[string]bool{}
	for _, it := range items {
		if seen[it.Key] {
			return fmt.Errorf("bad answers: %s is named twice", it.Key)
		}
		seen[it.Key] = true
		if key, ok := event.ParseKey(it.Key); ok && key.Kind == event.KeyLimit {
			return fmt.Errorf("bad answers: %s is a limit; an owner accepts or returns a limit", it.Key)
		}
		field := "closes"
		if f.ReadingClosed(it.Key) {
			field = "revises"
		}
		if err := f.checkReadingMove(field, it.Key, actor); err != nil {
			return err
		}
	}
	return nil
}

// applyAnswers records an answer's answers: the first answer of an open
// reading, or the answer in force from now on of an answered one.
func (f *Facts) applyAnswers(ev event.Event, items []event.KeyText) {
	for _, it := range items {
		f.Answers[it.Key] = append(f.Answers[it.Key], Answer{By: ev.ID, Text: it.Text, ByAnswer: true})
	}
}

// ReadingClosed reports whether the owner answered reading k.
func (f *Facts) ReadingClosed(k string) bool { return len(f.Answers[k]) > 0 }

// checkLines reports why an old checkpoint on Goal g may not write its
// lines: only an owner writes them, once per Goal, one case each. No command
// writes lines on a checkpoint any more; a Goal's lines come with its add
// (ADR 0020), which counts as the one write.
func (f *Facts) checkLines(g string, p event.Checkpoint) error {
	if len(p.Lines) == 0 {
		return nil
	}
	if p.Role != event.RoleOwner {
		return fmt.Errorf("bad checkpoint: only an owner writes the Goal's lines")
	}
	if by := f.LinesBy[g]; by != "" {
		if by == g {
			return fmt.Errorf("bad lines: a Goal's lines come with its node.add, and %s's did; a new line is a new Goal", event.DisplayID(event.KindGoal, g))
		}
		return fmt.Errorf("bad lines: the Goal's lines are written once, by %s; a new line is a new Goal", by)
	}
	if len(p.Lines) > event.MaxLimits {
		return fmt.Errorf("over limit: lines has %d items (max %d)", len(p.Lines), event.MaxLimits)
	}
	for _, l := range p.Lines {
		if err := textLimit("line", l, 1, event.MaxLimitItem); err != nil {
			return err
		}
	}
	return nil
}

// checkGoalLines reports why a Goal's add may not carry its lines: at most
// MaxLimits lines of 1 to MaxLine runes, each with at most MaxCites cites,
// every cite a key settled now (citeMeaning).
func (f *Facts) checkGoalLines(lines []event.GoalLine) error {
	if len(lines) > event.MaxLimits {
		return fmt.Errorf("over limit: lines has %d items (max %d)", len(lines), event.MaxLimits)
	}
	for i, l := range lines {
		if err := textLimit("line", l.Text, 1, event.MaxLine); err != nil {
			return err
		}
		if len(l.Cites) > event.MaxCites {
			return fmt.Errorf("over limit: line %d has %d cites (max %d)", i+1, len(l.Cites), event.MaxCites)
		}
		for j, c := range l.Cites {
			if slices.Contains(l.Cites[:j], c) {
				return fmt.Errorf("bad cite: line %d names %s twice", i+1, c)
			}
			if m, why := f.citeMeaning(c); m == "" {
				return fmt.Errorf("bad cite: line %d cites %s, which is not settled: %s; a cite is a line, an accepted or closed limit, an answered reading, or a selected or contradicted Hypothesis", i+1, f.citeDisplay(c), why)
			}
		}
	}
	return nil
}

// What a cite means, derived from the cited key's state at the line's place
// in the log (ADR 0020). CiteUnsettled is a state a key may reach later; a
// line is never written citing it.
const (
	CiteCarried   = "carried"
	CiteForbidden = "forbidden"
	CiteApplied   = "applied"
	CiteRealised  = "realised"
	CiteUnsettled = "unsettled"
)

// citeMeaning returns what citing c means now, or "" and why c is not
// settled: a line or an accepted limit is carried, a closed limit or a
// contradicted Hypothesis is forbidden, an answered reading is applied, and a
// selected Hypothesis is realised.
func (f *Facts) citeMeaning(c string) (meaning, why string) {
	if n := f.Nodes[c]; n != nil {
		if n.Kind != event.KindHypothesis {
			return "", fmt.Sprintf("a %s is cited by a key or, for a Hypothesis, its id", n.Kind)
		}
		for _, r := range f.ResultsByHyp[c] {
			if !f.Retracted[r] && f.Nodes[r].Verdicts[c] == event.VerdictContradicts {
				return CiteForbidden, ""
			}
		}
		if d := f.Decisions[n.Parent]; d != nil && d.Current != "" && d.Selects[d.Current].Hyp == c {
			return CiteRealised, ""
		}
		return "", "it is neither selected nor contradicted"
	}
	k, ok := event.ParseKey(c)
	if !ok {
		return "", "it names no node and no key"
	}
	if k.Kind == event.KeyReading {
		if err := f.checkKey(c, event.KeyReading, ""); err != nil {
			return "", err.Error()
		}
		if !f.ReadingClosed(c) {
			return "", "the reading is open"
		}
		return CiteApplied, ""
	}
	if err := f.checkKey(c, event.KeyLimit, ""); err != nil {
		return "", err.Error()
	}
	if f.Nodes[k.Result].Kind == event.KindGoal {
		return CiteCarried, ""
	}
	if f.limitCloser(c) != "" {
		return CiteForbidden, ""
	}
	if f.LimitHand[c].Move == HandAccept {
		return CiteCarried, ""
	}
	return "", "the limit is neither closed nor accepted"
}

// citeDisplay shows cite c as a display id or key.
func (f *Facts) citeDisplay(c string) string {
	if n := f.Nodes[c]; n != nil {
		return event.DisplayID(n.Kind, c)
	}
	return f.keyDisplay(c)
}

// keyDisplay shows key k as R-W6V06D#1, R-W6V06D#r1 or G-WH6GK5#1.
func (f *Facts) keyDisplay(k string) string {
	kind := event.KindResult
	if n := f.Nodes[event.KeyResult(k)]; n != nil {
		kind = n.Kind
	}
	return event.DisplayKey(kind, k)
}
