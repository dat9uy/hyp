# Page shape

Write the sections in this order. Each one feeds the next.

| # | Section | Holds |
|---|---|---|
| 0 | Banner | "Proposal, not decided", and the key to the source tags. |
| 1 | The event | What happened, with ids, in three to six sentences. The owner's concern in their words. |
| 2 | Where it breaks | A table: break, evidence. Three to six rows. |
| 3 | What gets cleared | The current rules built on the concept, struck through, each with its source. Then the rules that stay as given. |
| 4 | What any answer must keep | Four to six checkable requirements, each in a short card or row. |
| 5 | A worked case | One case the approaches can all run. Use the real event; if the event cannot show the concept, add one example tagged "Hypothetical" and say the event shows nothing. |
| 6 | Three approaches | One block each, with the approach table below and a diagram on the shared canvas (see "The diagrams"). |
| 7 | Side by side | The three diagrams in one row, small, then one table with approaches as columns. Mark each row of the event as caught or missed. |
| 8 | The lean | Tagged "a view, not a decision", with the condition under which it stops being right. |
| 9 | Open questions | One to three, each one the owner can answer. |

## The data file

The page is drawn from one JSON file by `scripts/draw_page.py`, so the page and its data
never disagree. `scripts/check_items.py` checks the file first. The file holds one
object with these fields, and no others:

| Field | Holds |
|---|---|
| `id` | The page id, such as `finding-key`. Every item id starts with it and a dash. |
| `title` | Optional. The page title. |
| `concept` | Optional. The concept's name, as the diagrams label it. |
| `items` | Every claim of the page, in reading order. |
| `parties` | The parties of the shared canvas, each `{"id", "label"}`, declared once. |
| `approaches` | The three approaches, each with where the concept sits and its steps. |

An item is `{"id", "tag", "date", "text", "section"}`. The tag is one of `Owner said`,
`Proposed`, `Open`, `Fact` or `Hypothetical`. The date is the day its words were
written. The section is the number of the section it sits in, 1 to 9; an item with no
section is drawn under "Other items".

An approach is `{"name", "concept_at", "steps", "caption", "table"}`. `concept_at` is
the id of the party where the concept lives. `steps` are the numbered steps of the
event: a text, or `{"text", "from", "to", "mark"}`, where `from` and `to` are party ids
and `mark` is `caught` or `missed`; such a step is drawn as a numbered arrow.
`caption` and `table` are optional: `table` maps each row of the approach table to its
text. An approach gives no positions and no colors: the script draws each diagram on
the one canvas, so only the concept moves.

```json
{"id": "demo", "concept": "the key",
 "items": [{"id": "demo-I1", "tag": "Owner said", "date": "2026-11-27",
            "text": "Keep the words of the owner.", "section": 1}],
 "parties": [{"id": "owner", "label": "Owner"}, {"id": "tool", "label": "Tool"}],
 "approaches": [{"name": "A", "concept_at": "owner",
                 "steps": [{"text": "The owner writes the key.", "from": "owner",
                            "to": "tool", "mark": "caught"}]}]}
```

## The approach table

Give each approach these rows. Leave none empty; write "nothing" when nothing applies.

| Row | Says |
|---|---|
| What it is | The concept in one sentence, as this approach shapes it. |
| Stores | What the system records that it does not record today. |
| Derives | What the system calculates from what it stores. |
| Who writes | Which party writes each stored thing (owner, runner, reviewer, tool). |
| Owner cost | What the owner does per run or per use. |
| Under the event | What the real event would have produced. |
| Strength | The one thing this approach does best. |
| Risk | What breaks, or what drifts back toward the old concept. |

## The diagrams

The owner tells the options apart by eye before reading the tables. So the three diagrams
are one picture with one part moved, not three pictures.

- **One canvas.** Draw the same parties and stores (for example owner, agents, the tool,
  the repository) in all three, at the same places, with the same colors. A party an
  approach does not use stays on the canvas, muted.
- **The concept is the one thing that moves.** Draw it as the only highlighted shape, at
  the place where this approach puts it.
- **The event runs on it.** Draw the real event as numbered arrows. Mark where the
  approach catches it (✓) or misses it (✗), at the point where that happens.
- **One caption.** A sentence under each diagram: "Here, <concept> lives in <place>."
- **Section 7 repeats them small.** The three diagrams sit in one row above the table, so
  the owner compares them in one look. The table marks each event row caught or missed.
- **Mechanics.** `scripts/draw_page.py` draws each diagram as inline SVG on the page's
  theme tokens, from the parties and steps of the data file. In Markdown, use a Mermaid
  flowchart per approach with the same node ids in all three.
- **Look at it rendered.** No label, mark or arrow may cross a shape or another label. A
  highlighted concept drawn on top of an existing box hides that box's name. Render the
  page when a browser is available and fix overlaps; otherwise say you did not look.

A diagram that only restates the approach table, such as boxes named after its rows, does
not count.

## Three approaches that differ in kind

Pick three different answers to "what is this concept, and where does it live?". Patterns
that worked:

- **A node, an edge pattern, a part of an existing thing.** The concept is stored as its
  own item; or it is never stored and is calculated from links between items; or it
  merges into an item that already exists.
- **Meaning held in cases, in words, or in a track record.** Concrete examples the owner
  placed; definitions the owner wrote; or the owner's rulings, against which an agent
  earns trust.
- **Centred on the question, on the candidates, or on the requirement.** The structure
  hangs from where the choice was found, from the things being chosen between, or from
  the rule that decides.

Three variants of one answer (three field names, three triggers) do not count.

## Worked example, compressed

Concept: DesignQuestion in a hypothesis tree. Event: a run's top question copied the Goal
title, and its parts had moved to other nodes.

- Breaks: it is asked, not found; the top question copies the Goal; its scope is a topic;
  it owns nothing of its own; it belongs to neither problem nor solution side.
- Cleared: adding a question on judgment; Goal completion through top-level questions;
  five states with ten reasons; sub-questions that block a select.
- Kept as given: the select is the gate; the owner chooses trade-offs.
- Must keep: found from a Result; an exact scope; grows in place; the finish line is the
  Goal's lines; works for build and discovery Goals.
- Approaches: a small branch node between Hypotheses; a branch calculated from two
  children that cite the same Result; the branch as the Experiment that separates its
  sides.
- Lean: the node, with one hard rule (no branch without the Result that showed it), and
  the measure and decider left to the Experiment.
