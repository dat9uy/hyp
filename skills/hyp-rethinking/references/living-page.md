# Living page

A rethink page often becomes the record of a long design talk: the owner decides item
after item, and the page must say at every moment what is decided and what is proposed. Each rule below comes from a defect the owner found in such a page.

## Where the page lives

- Edit the page's data file in place, under the project's plans or visuals folder, and
  draw the page again from it. The data file in the project is the only source. A copy
  kept in a temporary folder is lost with the session. Never edit the drawn HTML: the
  next draw replaces it.
- Before each edit, copy the data file to a backup named for the change. A wrong edit is
  then one copy away from undone, and the backups are the page's history.
- Make each edit a replacement that must match exactly once, and stop when it does not.
  A pattern that matches zero or two places means the page is not what you think it is.

## Tags change only with the owner's words

- **Proposed → Owner said** only when the owner accepts that item in this conversation.
  "Agree all" accepts what the reply in front of it listed, nothing more.
- **Open → decided** only by the owner's answer to that question. When a decision on
  another item seems to imply an answer, keep the question Open and show the
  implication as Proposed, beside it.
- **A replaced item keeps its id.** When the owner replaces an item, a proposal or the
  owner's own words, write the new words and the date of the owner's answer into that
  item. The item keeps its id and shows only its new words and date. Add no "replaced"
  marker, no "later" tag and no field for the old words: the backups hold them.
- Quote the owner exactly inside quote marks. A paraphrase is not a quote.
- An item the owner left for later stays tagged "Open", not dropped.
- The tags are `Owner said`, `Proposed`, `Open`, `Fact` and `Hypothetical`. The checker
  refuses any other.

## After each decision, sweep the page

A decision removes a word, a rule or a step. Every other place that still uses it now
contradicts the decision.

- After a decision that removes or renames something, scan the whole page for it: the
  word, its synonyms, and rules that depend on it. For a long page, give the scan to a
  subagent with the page as plain text, the list of what was removed, and the instruction
  to quote each hit with its section.
- When the page defines actors and their acts, the scan also checks both directions:
  every action an actor takes maps to a defined act, and every state comes from an act.
  A sentence saying the owner does something they have no act for is a defect.
- Fix what needs no decision. Bring the rest to the owner as questions.

## One home per rule

- State each rule in one section, and link it from the others. When a rule moves into a
  new section, delete it from the old one in the same edit.
- When two sections disagree, the owner has one rule in mind: ask which, then fix both.

## Keep the page whole

After each set of edits, run from the skill folder:

1. `python3 scripts/check_items.py <page>.json`. It names each defect of the data by its
   code and path: `id-missing`, `id-duplicate`, `id-prefix`, `tag-unknown`,
   `date-missing`, `text-missing`, `section-unknown`, `impact-missing`,
   `purpose-not-fact`, `source-missing`, `party-unknown`, `node-unknown`,
   `field-unknown` and `shape`. Fix each one in the data file until it exits 0.
2. `python3 scripts/draw_page.py <page>.json <page>.html`, which draws the page again.
3. `python3 scripts/check_page.py <page>.html`. Fix what it reports in the data file,
   then draw again. It checks four things that went wrong by hand:

- Section numbers run in order, and no "section N" reference points at a missing section.
  It cannot tell whether a reference points at the *right* section; read the ones near
  what moved.
- Every id in the page is unique, figure ids and arrow markers included. A copied figure
  keeps its ids, and the second one then breaks the first.
- Every descriptive sentence has 25 words or fewer. A long table cell or a quote may be a
  false hit; split it, or leave a quote as the owner said it.
- Section 7 holds a diagram per approach, and section 8 holds the visual comparison. A page
  written before that rule fails here: add the diagrams on its next update.

Then look at each new or changed figure rendered, or say in the reply that you did not.

## The reply after an update

Say what changed on the page and which tags moved, in a few lines. End with the open
questions that block the next item.
