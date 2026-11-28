# A rethink page explains the problem before it lists where the concept breaks

The spec of Goal G-HFSHFZ. It has no model, and the kept model in
`goals/model/` does not change: the change is to the hyp-rethinking skill's
data file format and its two scripts, which move no Goal between states.

## Line 1: section 2, "What it is for"

- The page shape has ten sections. Section 2 is "What it is for", between
  "1. The event" and "3. Where it breaks"; the old sections 2 to 9 are now 3
  to 10. `draw_page.py` draws an item with `"section": 2` there.
- An item's `section` is 1 to 10 (`section-unknown` otherwise).
- `check_page.py` asks for three diagrams in section 7 (the approaches) and
  one in section 8 (side by side).

## Line 2: diagrams of the event

- The data file may hold `diagrams`: each `{"id", "section", "caption",
  "nodes", "edges"}`, a node `{"id", "label"}`, an edge `{"from", "to"}`.
- `check_items.py` refuses an edge whose end names a node the diagram does not
  declare, with `node-unknown diagrams[i].edges[j].from` (or `.to`), exit 1;
  also a diagram with no id, a section not 1 to 10, a node declared twice, and
  an unknown field.
- `draw_page.py` draws each diagram in its section as an inline SVG with its
  caption, left to right: each node one column right of the nodes its edges
  come from. An edge that skips a column arcs above the boxes, so two shapes
  that differ only in such an edge look different.

## Line 3: the impact item

- An item may name `findings`, a list of finding ids. A section 1 item that
  names findings is an impact item.
- `check_items.py` refuses a page whose section 1 has items and no impact item,
  with `impact-missing items`, exit 1. A page with no section 1 item passes.
- `draw_page.py` draws the impact items of section 1 under "Who meets it",
  after the other items and diagrams of section 1, each with its findings.

## Line 4: section 2 holds facts with sources

- An item may name a `source`, a text. The page shows it after the item text.
- `check_items.py` refuses an item of section 2 whose tag is not `Fact`
  (`purpose-not-fact items[i]`) and one with no source
  (`source-missing items[i]`), exit 1. It names no other item.

## The skill text

- `references/page-shape.md` lists the ten sections, the new fields, the two
  section rules and the diagram format; `SKILL.md` and
  `references/living-page.md` name the new codes and the new section numbers.
