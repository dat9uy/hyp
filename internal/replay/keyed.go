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
	Move, Event, Actor, Referrer string
}

// HandRefer, HandReturn and HandAccept are the values of Hand.Move.
const (
	HandRefer  = "refer"  // the runner hands the limit to the owner
	HandReturn = "return" // the owner hands it back: the Goal covers its case
	HandAccept = "accept" // the owner accepts it: this tree does not run its case
)

// limitMove is one row of the limit move table: the payload field that
// makes the move, the moves the key's latest hand may be for it to apply,
// and whether only an owner checkpoint makes it. A move that answers a
// refer is made by an actor other than the referrer, and every move is
// refused once a Result closed a limit of a Result: its case ran.
type limitMove struct {
	field    string
	from     []string // the latest hand's Move; "" stands for no hand yet
	owner    bool
	answers  bool // the actor differs from the latest referrer
	notRun   bool // refused once a Result closed the limit
	refusal  string
	ownerErr string
	ranErr   string // what follows "closed by R-…" in the notRun refusal
}

var limitMoves = map[string]limitMove{
	HandRefer: {field: "refers", from: []string{"", HandRefer, HandReturn}, notRun: true,
		refusal: "%s is accepted; only the owner reopens it, with returns",
		ranErr:  "; its case ran, so there is nothing to refer"},
	HandAccept: {field: "accepts", from: []string{HandRefer}, owner: true, answers: true, notRun: true,
		refusal:  "%s is not referred; only a referred limit is accepted, so the runner refers it first",
		ownerErr: "bad checkpoint: only an owner accepts a limit", ranErr: "; its case ran, so there is nothing to accept"},
	HandReturn: {field: "returns", from: []string{HandRefer, HandAccept}, owner: true, answers: true, notRun: true,
		refusal:  "%s is not referred or accepted; only a referred or accepted limit is returned",
		ownerErr: "bad checkpoint: only an owner returns a limit"},
}

// The table's rows are read by checkLimitMove and applyLimitMove only.

// checkLimitMove reports why move may not apply to limit key k now, made by
// actor on a checkpoint whose role is owner or not.
func (f *Facts) checkLimitMove(move, k, actor string, owner bool) error {
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
		return fmt.Errorf("separation of duties: %s referred the limit it %s", actor, map[string]string{HandAccept: "accepts", HandReturn: "returns"}[move])
	}
	return nil
}

// applyLimitMove records move on limit key k, made by ev.
func (f *Facts) applyLimitMove(move, k string, ev event.Event) {
	h := f.LimitHand[k]
	h.Move, h.Event, h.Actor = move, ev.ID, ev.Actor
	if move == HandRefer {
		h.Referrer = ev.Actor
	}
	f.LimitHand[k] = h
}

// checkLimitMoves checks every refer, return and accept a checkpoint makes,
// in the order replay applies them, and that no key is named by two of
// them. Each move is checked against the hand before the checkpoint.
func (f *Facts) checkLimitMoves(ev event.Event, p event.Checkpoint) error {
	owner := p.Role == event.RoleOwner
	for _, mv := range []struct {
		move string
		keys []string
	}{{HandRefer, p.Refers}, {HandReturn, p.Returns}, {HandAccept, p.Accepts}} {
		if len(mv.keys) > 0 && limitMoves[mv.move].owner && !owner {
			return fmt.Errorf("%s", limitMoves[mv.move].ownerErr)
		}
		if len(mv.keys) > event.MaxLimits {
			return fmt.Errorf("over limit: %s has %d items (max %d)", limitMoves[mv.move].field, len(mv.keys), event.MaxLimits)
		}
	}
	for _, mv := range []struct {
		move string
		keys []string
	}{{HandRefer, p.Refers}, {HandReturn, p.Returns}, {HandAccept, p.Accepts}} {
		for _, k := range mv.keys {
			for _, other := range []struct {
				move string
				keys []string
			}{{HandRefer, p.Refers}, {HandAccept, p.Accepts}} {
				if mv.move != other.move && mv.move == HandReturn && slices.Contains(other.keys, k) {
					return fmt.Errorf("bad returns: %s is also in %s", f.keyDisplay(k), limitMoves[other.move].field)
				}
			}
			if err := f.checkLimitMove(mv.move, k, ev.Actor, owner); err != nil {
				return err
			}
		}
	}
	return nil
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

// applyReadingMoves records the answers a checkpoint gives: a close gives
// a reading its first answer, a revise the answer in force from then on.
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

// ReadingClosed reports whether an owner checkpoint answered reading k.
func (f *Facts) ReadingClosed(k string) bool { return len(f.Answers[k]) > 0 }

// checkLines reports why a checkpoint may not write its lines: only an
// owner writes them, once per Goal, one case each.
func (f *Facts) checkLines(p event.Checkpoint) error {
	if len(p.Lines) == 0 {
		return nil
	}
	if p.Role != event.RoleOwner {
		return fmt.Errorf("bad checkpoint: only an owner writes the Goal's lines")
	}
	if f.LinesBy != "" {
		return fmt.Errorf("bad lines: the Goal's lines are written once, by %s; a new line is a new Goal, so a new tree", f.LinesBy)
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

// keyDisplay shows key k as R-W6V06D#1, R-W6V06D#r1 or G-WH6GK5#1.
func (f *Facts) keyDisplay(k string) string {
	kind := event.KindResult
	if n := f.Nodes[event.KeyResult(k)]; n != nil {
		kind = n.Kind
	}
	return event.DisplayKey(kind, k)
}
