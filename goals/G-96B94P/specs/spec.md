# A rethink page becomes an issue that holds every item, and no item loses its holder without a check saying so

The spec of Goal G-96B94P. It has no model, and the kept model in
`goals/model/` does not change: the change is to the hyp-grill skill's text
and two scripts it ships, which move no Goal between states.

## Line 1: the issue body is drawn from the page's data file

- `skills/hyp-grill/scripts/draw_issue.py PAGE.json` checks the data with
  `skills/hyp-rethinking/scripts/check_items.py`, the skill beside it. On a
  defect it prints each defect to stderr as `CODE PATH: message`, prints no
  body and exits 1; exit 2 when the file or `check_items.py` cannot be read.
- Otherwise it prints a Markdown body: one line naming the page id and title,
  then each section of the page shape that holds items, in page order, as
  `## N. Name`, each item as `- **ID** (TAG, DATE): TEXT`, with its source and
  findings. Items with no section go under `## Other items`.

## Line 2: the holder table

- The body ends with `## Holders` and the table `| Item | Tag | Holder |`,
  one row for each item, in the data file's order, with the holder empty.

## Line 3: the holder check

- `skills/hyp-grill/scripts/check_holders.py BODY.md` reads that table and
  prints `CODE ITEM: message` per defect, exit 1; exit 0 and nothing printed
  when every item is held; exit 2 when the body cannot be read or `hyp`
  cannot run.
- Holder forms: `G-X line-N` (a Goal line), `G-X line-N/NAME` or
  `draft NAME line-N/NAME` (an example), `F-X` (a finding), `owner: "WORDS"`
  (the owner's answer), `close: "WORDS"` (the owner's close).
- Kinds per tag: Owner said and Proposed, a Goal line; Fact, a finding or a
  Goal line; Open, the owner's answer; Hypothetical, an example; any tag, a
  close.
- Codes: `table-missing`, `tag-unknown`, `id-duplicate`, `holder-missing`,
  `holder-form`, `holder-kind`, `holder-unknown` (no such Goal, or no such
  line in it), `holder-ended` (the Goal is ended and was closed, or landed
  with no `stop: selected`).
- A Goal holder is read with `hyp status G --actor A`, A being `$HYP_ACTOR`
  or `check-holders`. A finding holder is not read from the store.

## Lines 4 to 8: the skill text

- `SKILL.md`, "A rethink page becomes an issue": the grill opens the issue
  from the page itself with `draw_issue.py`; the kind of holder each tag
  takes; an item too vague for a sharp line is held by a line of a written
  Goal and sharpened before the Goal opens; `check_holders.py` runs at every
  split and before any report of done; at the end of each Goal the land and
  each remainder go to the tracking issue, and with no tracking issue each
  remainder is a finding only.
- `run-rules.md`, rule 10: a partial fix is recorded as a finding; a PR's
  notes alone do not hold it.
- `re-grill.md`, "A partial fix at Selected": the orchestrator records it as a
  finding at Selected, whatever the owner answers.
- `docs/glossary.md` adds Rethink page, Holder and Orchestrator.
