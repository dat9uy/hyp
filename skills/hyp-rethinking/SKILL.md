---
name: hyp-rethinking
description: "Rethink one concept of a system from scratch: show where the concept breaks, set aside the rules built on it, state what any replacement must keep, and give three different ways to build it, each tested against the real event. Use when the owner says a concept feels wrong, asks to go back to square one, to remove the existing behaviour and suggest ways from scratch, rejects advice that only repairs the current mechanism, or asks what a loose word in the model means. Also use to update an existing rethink or proposal page as the owner decides its items, or to grill the owner on one section of such a page. Not for a new feature that has no existing concept to rethink."
metadata:
  version: 2.1.0
---

# Rethinking

Produces a proposal that lets the owner choose a new shape for one concept of their
system. The owner reads it, pushes back, and often asks for the same exercise on a
smaller concept that the first one exposed. The bar: the owner can see why the concept
breaks, can compare three real alternatives, and finds no repair of the old mechanism
presented as a new idea.

## When to use

- The owner says a concept "feels wrong", "doesn't make sense", or "doesn't move the
  architecture forward".
- The owner asks to "go back to square one", to "skip the current X", or to "remove the
  existing behaviour, then suggest three ways from scratch".
- A definition in the current discussion uses a loose word ("grows", "shows a choice")
  and the owner asks what it means.
- The owner decides, rejects or changes items on a page this skill wrote.
- The owner asks to grill one section of a living page ("grill §3 of <page>").
- Not for: a new feature with no existing concept, or a decision record after the
  owner chooses.

## Authority

The owner decides. This skill proposes, and it records the owner's decisions only on its
own page, as "Owner said" tags. It never edits the system or writes to the project's own
records. The page stays marked "Proposal, not decided" until the owner records the
decision elsewhere. Write only the page and its files.

The page is scratch: no record cites it. When the owner wants its items built, hand
the data file to the hyp-grill skill, which draws an issue from it with every item and
a holder table, and checks that each item has a holder. Do not copy items into an
issue by hand: a copy made from memory drops items unseen.

## How to work

When the request is only what a loose word means, answer that first and stop: rewrite
the definition as a sentence with named slots, give each slot "must be" and "if it can't
be filled", and say which existing concept each slot already belongs to. Run the full
rethink below only when a slot has no home, or the owner asks for it.

1. **Name the one concept** under rethink, and the event that points at it: a real run,
   a refusal, a defect, or the owner's words. With no event, say so on the page; an
   approach then has nothing real to be tested against.
2. **Gather the facts first.** Read the project's glossary, specs and the code that hold
   the concept. Every "today it works like this" claim cites a file and line or an id
   the reader can check. Look up each word you plan to use in the glossary before you
   use it: a word with a fixed project meaning (for example "round") misleads when used
   loosely. Then read what the owner already decided about the concept: the record
   that points at it, every record it cites, with their comments, and any earlier
   rethink page on the concept. A later comment can replace a body's decision, and a
   closed or landed record still holds decisions nothing has built yet.
3. **Show where it breaks.** List each break with its evidence from the event. Derive
   the breaks from the owner's principles and from what each part of the system is
   for, never from the design that already assumes the concept. When the concept
   decides how the system checks, reruns, counts or improves its own work, load
   `hyp-bayes-thinking` and take breaks and requirements from its five slots.
4. **Clear the tangle.** List every current rule built on the concept, struck through,
   with its source. Say which rules stay as given (for example "the owner decides
   trade-offs"). The approaches start without the struck rules.
5. **State what any answer must keep.** Four to six requirements, each one checkable.
   A requirement says what must be true, never how it is stored or computed, because
   a mechanism in a requirement rules out the approaches that do it another way. Test
   each one: could all three approaches meet it? If not, it is an approach in
   disguise; move it. Example: "a finding names its Goal line in a stored field" rules
   out an approach that derives the line, so write "a finding's Goal line can be found
   without reading prose". Each owner decision found in step 2 enters here as a
   requirement tagged "Owner said", with its source. One you would rather change goes
   to the open questions; it does not leave the list unseen.
6. **Give three approaches that differ in kind.** They differ in what the concept *is*
   or where it lives (a node, a derived pattern, a part of an existing thing), not in
   detail. For each, give the rows of the approach table in
   `references/page-shape.md`, including how the real event runs under it.
7. **Prefer what exists.** When an approach lands on an existing concept of the system,
   say so and reuse it. Two approaches that collapse into one existing concept become
   one approach; name why.
8. **Compare, then lean.** A side-by-side table, then a short lean marked "a view, not
   a decision". Name when it stops being right in both directions: it misses what it
   should catch, or it costs the owner more than it gives (such as a question the owner
   cannot answer). A lean that breaks a rule of a lens you loaded, such as
   hyp-bayes-thinking, says so.
9. **End with the open questions** the owner must answer to choose. One to three.

## Rules that came from real reviews

- **Fill an approach from recorded events only.** For each state the approach holds,
  name the events it is calculated from. A state read off another system's history (git
  log, a merge, a file's presence) is a break, not a detail of the approach. Owner, after
  such a fill: "I don't like derived state based on side-effect".
- **Clear before you propose.** An answer that patches the current mechanism ("add a
  field to X", "change the trigger of Y") is not a rethink. The owner rejected such
  advice because it "doesn't move the architecture forward".
- **Tag every claim's source.** "Owner said" for what the owner stated or accepted in
  this conversation, "Proposed" for yours, "Open" for an undecided point, "Fact" for a
  "today" claim with its source, and "Hypothetical" for an example that no event
  produced. These five are the only tags; the checker refuses any other. Never move the
  owner's words into "Proposed" or the reverse.
- **Mark foresight.** A requirement, approach or example that no event supports is
  marked as foresight. Run the event under each approach; if it shows nothing, say so.
- **Define loose words into slots.** A definition that rests on a vague verb ("grows",
  "shows a choice") is rewritten into slots, as the opening of "How to work" says, also
  when it turns up inside a full rethink.
- **Settle the concept before the name.** Use the owner's placeholder name if given.
  Do not reuse a word the discussion already used in another sense.
- **Draw each approach on one shared canvas.** Owner, after a page of three text-only
  approaches: "always have visualization for me to differentiate between options". Each
  approach gets a diagram, and the three use the same parties at the same places, so the
  only thing that moves is where the concept lives. Section 8 is visual too. The rules are
  in `references/page-shape.md`, "The diagrams".
- **Separate fact from interpretation.** A test result is a fact; deciding which rule
  or requirement it falls under is interpretation, and it belongs to whoever owns
  meaning.

## Output

Write every sentence in the voice of `references/public-doc-voice.md`: read it and apply
its "The voice" and "Hard no" rules. Its "Structure that works" section does not apply;
use the page structure in `references/page-shape.md`. Keep project terms, CLI words and
ids exactly as written in the project.

The page is data first. Write its items, parties and approaches in a data file, as
`references/page-shape.md`, "The data file", says, under the project's plans or visuals
folder (for example `plans/visuals/<concept-slug>.json`). Then:

1. Run `python3 scripts/check_items.py <page>.json` from the skill folder. Fix each
   defect it names, by its code and path, until it exits 0.
2. Run `python3 scripts/draw_page.py <page>.json <page>.html`. It draws the whole page,
   the diagrams included, and draws nothing while the data has a defect.
3. Run `python3 scripts/check_page.py <page>.html` and fix what it reports in the data
   file, then draw again. Never edit the drawn HTML by hand: the next draw replaces it.

Open the drawn page when the runtime can open one. Otherwise write the same structure as
Markdown in the reply, with a Mermaid diagram per approach. When the exercise runs again
on a smaller concept, make a new page and link the earlier one.

When the owner asks to update a page, or decides items on it, read
`references/living-page.md` first: it holds how tags move, the sweep after each decision,
and the checks that keep the page whole. When the owner asks to grill one section, also
read `references/section-grill.md`.

## Done when

- The page holds every section of `references/page-shape.md` in order.
- Every "today" claim has a source the reader can check, and was checked this session.
- The three approaches differ in kind, and each shows the event run under it, in words and
  in a diagram on the shared canvas. Section 8 holds a visual comparison.
- `scripts/check_items.py` exits 0 on the data file, and the page is drawn from it.
- Every claim carries a source tag, every sentence passes the "The voice"
  rules of `references/public-doc-voice.md` (a descriptive sentence has 25 words or
  fewer), and no item from its "Hard no" list remains.
- After an update: the checks in `references/living-page.md` pass.
- The reply to the owner gives the page path, two or three lines per approach, the lean,
  and the open questions. It names what was not verified.

## Resources

- `references/page-shape.md`: the page sections, the approach table, and a short worked
  example. Read it before you write the page.
- `references/living-page.md`: how to update a page across a long design talk. Read it
  before you change a page that already exists.
- `references/section-grill.md`: how to grill the owner on one section of a living page,
  or walk a real event through it, one question at a time, recording only the owner's
  words. Read it for a section grill or an event walk.
- `references/public-doc-voice.md`: the voice every sentence of the page is written in.
- The data file, its checker and its drawing take their ideas from ak-render (MIT). No
  code is copied from it, and no AgentKit tool is needed.
- `scripts/check_items.py <page>.json`: refuses a data file with exit 1 and names each
  defect by its code and path; the codes are listed in the script's header. Exit 2 when
  the file cannot be read.
- `scripts/draw_page.py <page>.json <page>.html`: draws the whole page from its data
  file. Exit 1 when the data has a defect, 2 when a file cannot be read or written.
- `scripts/check_page.py <page>`: reports headings out of order, references to a missing
  section, duplicate ids, sentences over 25 words (a quotation counts as one word), and
  sections 7 and 8 without the diagrams they need. Exit 1 means findings. Python 3
  standard library only; runs on macOS, Linux and Windows. Tests: `python3 -m unittest discover -s
  scripts/tests`.
