# A rethink page holds its items as data that a shipped checker refuses when they are wrong

The spec of Goal G-AZ5JBT. It has no model, and the kept model in
`goals/model/` does not change: the change adds two skills and their scripts,
and moves no Goal, finding or store between states. No command of hyp changes.

## Line 1: the two skills ship with hyp

- `skills/hyp-rethinking/SKILL.md` has `name: hyp-rethinking` and loads
  `hyp-bayes-thinking`; `skills/hyp-bayes-thinking/SKILL.md` has
  `name: hyp-bayes-thinking` and names `hyp-rethinking`. Neither names the
  old bare names.

## Line 2: the voice ships as a reference

- `skills/hyp-rethinking/references/public-doc-voice.md` is a copy of the
  public-doc-voice rules, with "The voice" and "Hard no", and names no other
  skill. The SKILL.md reads it from there.

## Line 3: the data checker

- `scripts/check_items.py PAGE.json` prints one line per defect as
  `CODE PATH: message` and exits 1. Codes: `id-missing`, `id-duplicate` (on
  the later item), `id-prefix`, `tag-unknown`, `date-missing`, `text-missing`,
  `section-unknown`, `party-unknown`, `field-unknown`, `shape`.
- The tags are `Owner said`, `Proposed`, `Open`, `Fact`, `Hypothetical`.
- The fields are: page `id, title, concept, items, parties, approaches`; item
  `id, tag, date, text, section`; party `id, label`; approach
  `name, concept_at, steps, caption, table`; step `text, from, to, mark`.
- Exit 2 when the file cannot be read or is not JSON; exit 0 on a whole file.

## Line 4: the page is drawn from its data

- `scripts/draw_page.py PAGE.json PAGE.html` checks the data as line 3, then
  writes the whole page: banner, sections 1 to 9 with each item in its
  section, a diagram per approach in section 6, and the same diagrams small in
  section 7 above the side-by-side table. Every diagram draws every declared
  party at the same place; only the highlighted concept moves.
- The drawn page passes `scripts/check_page.py`. Exit 1 on defective data
  (nothing is written), 2 when a file cannot be read or written.

## Line 5: a replaced item keeps its id

- An item shows its id, tag, date and text. No field holds old words or a
  replaced marker: such a field is `field-unknown`, and an item with no date
  is `date-missing`. `references/living-page.md` tells the writer to write the
  new words and date into the item.

## Line 6: the page checker

- `scripts/check_page.py` ships unchanged from the owner's rethinking skill:
  heading order, reference to a missing section, duplicate id, sentence over
  25 words, and sections 6 and 7 without their diagrams; exit 1 on findings.
