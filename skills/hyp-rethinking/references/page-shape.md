# Page shape

Write the sections in this order. Each one feeds the next.

| # | Section | Holds |
|---|---|---|
| 0 | Banner | "Proposal, not decided", and the key to the source tags. |
| 1 | The event | What happened, with ids, in three to six sentences. The owner's concern in their words. A diagram of the event when its shape matters, such as a commit graph of each branch shape that gives a different verdict. Then the impact: who meets the problem on the normal path, with the findings that show it, drawn under "Who meets it". |
| 2 | What it is for | What the concept is for, and each project word the page uses, such as H0 or the own change, introduced before the breaks use it. Each item is a Fact with a source the reader can check. |
| 3 | Where it breaks | A table: break, evidence. Three to six rows. |
| 4 | What gets cleared | The current rules built on the concept, struck through, each with its source. Then the rules that stay as given. |
| 5 | What any answer must keep | Four to six checkable requirements, each in a short card or row. |
| 6 | A worked case | One case the approaches can all run. Use the real event; if the event cannot show the concept, add one example tagged "Hypothetical" and say the event shows nothing. |
| 7 | Three approaches | One block each, with the approach table below and a diagram on the shared canvas (see "The diagrams"). |
| 8 | Side by side | The three diagrams in one row, small, then one table with approaches as columns. Mark each row of the event as caught or missed. |
| 9 | The lean | Tagged "a view, not a decision", with the condition under which it stops being right. |
| 10 | Open questions | One to three, each one the owner can answer. |

## The data file

The page is drawn from one JSON file by `scripts/draw_page.py`, so the page and its data
never disagree. `scripts/check_items.py` checks the file first. The file holds one
object with these fields, and no others:

| Field | Holds |
|---|---|
| `id` | The page id, such as `finding-key`. Every item id starts with it and a dash. The id stays the same when the page is rewritten, so an item id and the issue drawn from the page still name the same item. |
| `title` | Optional. The page title. |
| `concept` | Optional. The concept's name, as the diagrams label it. |
| `items` | Every claim of the page, in reading order. |
| `parties` | The parties of the shared canvas, each `{"id", "label"}`, declared once. |
| `approaches` | The three approaches, each with where the concept sits and its steps. |
| `diagrams` | Optional. Diagrams of the event that are not the shared canvas, such as a commit graph. |

An item is `{"id", "tag", "date", "text", "section", "source", "findings"}`. The tag is
one of `Owner said`, `Proposed`, `Open`, `Fact` or `Hypothetical`. The date is the day
its words were written. The section is the number of the section it sits in, 1 to 10;
an item with no section is drawn under "Other items". `source` is optional: where the
reader checks the item, such as a path with its line, a commit or a finding id; the page
shows it after the text. `findings` is optional: the ids of the findings the item rests
on.

Three rules `scripts/check_items.py` checks:

- **The page opens with a picture of the problem.** The data holds at least one diagram
  of section 0. The page draws it in a block above "1. The event". A page with none is
  refused (`picture-missing`); a diagram of section 1 does not stand in for it. When the
  problem is numbers against a limit, draw bars (below), not a graph.
- **Section 1 is short.** It holds at most six items besides the one impact item
  (`event-too-long`).
- **Section 1 names who meets the problem.** When section 1 holds items, one of them is
  the impact item: it names the `findings` that show who meets the problem on the normal
  path. The page draws it under "Who meets it". A section 1 with items and no impact item
  is refused (`impact-missing`).
- **Section 2 holds only facts.** Each item of section 2 is tagged `Fact` and has a
  `source` (`purpose-not-fact`, `source-missing`).

A diagram is `{"id", "section", "caption", "nodes", "edges"}`, with `section` 0 to 10. Each node is
`{"id", "label"}`, and each edge `{"from", "to"}` names two nodes of the same diagram. The
page draws it in its section, left to right, each node one column right of the nodes its
edges come from; an edge that skips a column arcs above the boxes. Use one diagram per
shape the event can take, each with a caption that names its verdict, so the owner sees
why the two shapes differ.

A diagram of numbers is `{"id", "section", "caption", "bars", "limit"}` in place of nodes
and edges. Each bar is `{"label", "value"}` and `limit` is `{"label", "value"}`. The page
draws each bar to scale with its number, marks "over" on each bar above the limit, and
draws the limit as a line. A bar or a limit with no number is refused
(`bar-number-missing`).

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
            "text": "Keep the words of the owner.", "section": 1},
           {"id": "demo-I2", "tag": "Fact", "date": "2026-11-27", "section": 1,
            "text": "A runner who branches from main meets the refusal.",
            "source": "F-1NG9VB", "findings": ["F-1NG9VB"]},
           {"id": "demo-I3", "tag": "Fact", "date": "2026-11-27", "section": 2,
            "text": "H0 is the commit a Goal starts from.", "source": "docs/glossary.md"}],
 "diagrams": [{"id": "problem", "section": 0, "caption": "The branch is refused",
               "nodes": [{"id": "main", "label": "main"}, {"id": "refused", "label": "refused"}],
               "edges": [{"from": "main", "to": "refused"}]},
              {"id": "commits", "section": 1, "caption": "Refused: branch from main",
               "nodes": [{"id": "h0", "label": "H0"}, {"id": "fix", "label": "fix"}],
               "edges": [{"from": "h0", "to": "fix"}]}],
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
- **Section 8 repeats them small.** The three diagrams sit in one row above the table, so
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
