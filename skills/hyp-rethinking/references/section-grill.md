# Section grill

The owner asks to grill one section of a living page, such as "grill §3 of
<page>", or to walk a real event through the page and grill where it sticks. The result is the same page, with that section's open questions
answered in the owner's words, and nothing else changed. The owner reads only
the section and the answers; they never approve a table you wrote.

Read `living-page.md` first. Its rules on tags, backups, sweeps and the checker
all apply here.

## Before the first question

- Read the whole page, then the named section closely. List its Open items and
  the open questions of section 10 that name the section.
- Back up the page's data file, named for the grill:
  `<page>.json.bak-before-grill-<section>`.
- Do not read code or skills from an older version of the system to answer a
  question about the model on the page. The page and the pages it links are the
  source. Where the page is silent, say so.

## How to ask

- **One question at a time.** Ask the one open item that blocks the most others
  in the section. Send one question per prompt, and wait for the answer before
  the next. A follow-up to an answer is its own prompt.
- **First ask whether the concept is needed.** Before you ask how an item should
  work, ask whether the concept it rests on earns its place: name what on the
  page reads it. An item such as "does a line keep a mark?" hides that
  question. A warning or a limit rests on the concept it guards, so ask about
  that concept, not only the warning: in a real grill, "does the warning past
  three gap lines stay?" hid "do we need a coarse Goal at all?". The owner
  reframed both items that way: "the question should be asked is if we needed
  this split at all". The options must be at that level too. A question that
  opens "is X needed at all?" and then offers four ways to build X is still a
  question about mechanisms.
- **Start from the problem, in the owner's words.** Open each prompt with the
  problem the item serves, quoted from the owner's earlier answers or the page.
  Then ask. Options are choices about that problem, not designs for it. In a
  real walk the owner rescoped six of nine questions ("wrong scope", "Need to
  reframe the problem") because each one asked who writes a thing or what a
  thing holds before the problem was stated.
- **Keep the prompt short.** The problem in one or two sentences, then two to
  four options of one line each, with the consequence. Mark one as your lean
  with one reason. The owner may answer outside them.
- **Re-read before each lean.** Read the owner's answers so far and the page's
  removed words (its "words that leave" table) before you offer a lean. A lean
  must not stretch an answer past what the owner said, and must not bring back
  a removed concept under a new name. In a real walk, a lean turned "no coding
  session to run an example" into "no code at all", and an option offered the
  removed reviewer as "an agent adds examples".
- **No approval tables and no read-backs.** Do not show a rewritten section and
  ask "is this right?". The owner decides better from consequences than from
  drafts.
- **"I don't know" is an answer.** Record it as still Open, with the owner's
  words, and go to the next question. Do not push for a decision the owner says
  is premature.
- **Two rescopes in a row: stop asking.** When the owner answers two questions
  in a row by reframing the problem, you are asking at the wrong level. Stop
  the grill, record what the owner said, and offer a rethinking page on the
  reframed concept, where you do the framing and the owner chooses. Continuing
  spends the owner's attention on your framing work.
- **A thought is not a decision.** When the owner says they are "just telling a
  thought", record nothing as Owner said. Ask whether to note it, and tag it
  Open if they agree.

## Lines before examples

When the grill or a walk reaches a Goal's lines, check each line before you ask
for its example:

- **A want, not a mechanism.** The owner would see something when the line is
  met. "The channel is one block of code, defined with Go" is a way to build,
  not something to see.
- **One thing to see.** A line with two things to see is two lines.
- **No overlap.** For each pair of lines, an example exists that fails one and
  passes the other. If none exists, the lines hold one want.

Ask for the example with the open question "What would you do, and what would
you see?", and let the owner answer in their own words. Do not offer ready-made
examples as options: an answer with two things to see is how a combined line
shows, and a ready-made example always has one.

## Walking a real event

- Use an event whose Goal the owner still stands behind. If the owner ended it
  as a wrong design, start by grilling its words and lines, not by walking them.
- Walk only to find where the page has no rule. Each stuck step names what the
  real run did there, then follows "How to ask".

## Stay in the section

- A question that the answer opens in another section goes to the open-questions
  list under that section's row. Do not grill it now.
- An edit in another section is allowed when the option the owner chose names
  it. Make it and list it in the final reply. Any other conflict with another
  section becomes an Open row naming both, and you tell the owner at the end.
- Invent no mechanism to fill a gap the owner left. Write it as Open.

## Recording each answer

- Under **Owner said** goes only the owner's words and the text of the option
  they chose, quoted exactly where a quote matters. Your reasons for an option
  are not the owner's, even when the owner picked it: "nothing reads the split"
  is your argument. Keep it out of the tagged item, or tag it **Proposed**.
- An answer that replaces an Owner said item writes the new words and date into
  that item, which keeps its id, as `living-page.md` says. No "replaced" marker
  is added.
- Remove the answered row from the open-questions list in the same edit.
- After each answer that removes or renames something, sweep the page as
  `living-page.md` says. Include the tables: the words that leave and the open
  questions. A removed word left in a
  table points at nothing.

## Ending the grill

The grill ends when the section has no Open item the owner wants to answer
now, or when the owner stops it. Then:

1. Run the three checks of `living-page.md`, "Keep the page whole", and fix what
   they report.
2. Reply with: the questions answered and their tags, the items still Open, and
   any row added to another section. Name the backup file.
