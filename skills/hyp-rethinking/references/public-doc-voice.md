
# Public document voice

A copy of the owner's public-doc-voice rules, shipped with hyp-rethinking so that the
skill needs no other skill. It covers the prose of a document that a reader outside the
project reads first, such as a README, a project introduction, an overview page or a
rethink page. The result is text the owner accepts without a rewrite for tone, cliché
or scope.

A rethink page applies "The voice" and "Hard no" to every sentence. It takes its
structure from `page-shape.md`, not from "Structure that works".

## Before you write

1. Read the whole target document, the project glossary if one exists, and every source
   the text will describe. A claim about another author's work is checked against that
   work, at a pinned commit, before it is written.
2. Find the one concept the section is about. Name it, and build the section around it.
3. Decide the reader's level for each section: the introduction speaks in the reader's
   plain words; project terms appear only after the text introduces them.

## The voice

Write in ASD-STE100 Simplified Technical English. It is technical reading, not a story.

- One idea per sentence. A descriptive sentence has 25 words or fewer. A paragraph has
  6 sentences or fewer.
- Active voice. The subject is the party that acts: "You write the lines one time",
  "hyp calculates each status", not "lines are written".
- Simple words with one meaning: "examine" not "investigate", "calculate" not "derive",
  "a minimum of two" not "at least two", "one time" not "once". No -ing forms as nouns
  and no phrasal verbs when a single verb exists.
- Keep the project's technical names, CLI words and glossary terms exactly as they are.
  STE permits technical names; a synonym breaks the match with the glossary and the CLI.
- Bold-lead bullets for rules: **The rule.** Then one to three short sentences.

## Hard no

The owner rejected each of these in a real review:

- **Clichés and sales patterns.** "This problem is not new, and good answers already
  exist", "With X, your Y becomes Z", "cannot be bent", "the magic of", "seamless".
  Replace with a plain statement whose subject is a real party.
- **Narrative tone.** "hyp began with one investigation…", "Once deciding worked…".
  State the problem and the rule; do not tell how it felt.
- **Jargon leaked into the problem statement.** The problem is stated at the reader's
  level ("the agent takes the first design that seems to work"), not in the project's
  words ("does not show which hypothesis started the work").
- **A scope that is too narrow.** When the project serves two uses (for example design
  and build), the opening covers both at one level of abstraction; neither sits as an
  afterthought under the other.
- **Display names treated as concepts.** A CLI heading such as `AWAITING OWNER` is
  output. Center the text on the concept behind it (for example "owed"), and mention
  the output only as where the concept prints.
- **Internals in the first lines.** A tagline does not open with storage formats or
  supported harnesses. It names the thing, what the reader does, and what the tool does.
- **A flourish sentence the owner did not ask for.** If a sentence restates the next
  section, cut it.

## Structure that works

For an introduction or a "why" section, use this order. Each part leads into the next.

1. **The problem**, in plain words, for every use the project serves.
2. **Credit to earlier work.** One plain bridge ("Other authors already wrote skills for
   each part of this problem. X starts from their skills:"), then one bullet per source.
   Each bullet names which part of the problem that source answers.
3. **What this project adds**, framed by its own extra purpose, never as a flaw in the
   sources: "These skills work well, and X keeps their rules. But each skill solves one
   part alone…"
4. **Layers or sections**, each with the same shape where a source exists: the idea taken,
   what was kept, the gap that stays for this purpose, then a two-column table
   `| In <source> | The rule in <project> |`. A layer without a source opens with the gap
   it closes.

Honor sources with care: link each one at a pinned commit, describe it accurately and
generously, and say what the project kept before saying what it added.

Add a Mermaid diagram when the text explains a structure or a flow. Give it a one-sentence
lead in the same voice. Validate that it renders, or say that it was not rendered.

## How to work with the owner

- When a choice is about taste (a tagline, a bridge sentence, a heading), offer two to
  four short options, mark one as the recommendation with one reason, and wait.
- When the owner names a sentence they dislike, change that sentence and what depends on
  it. Keep everything they accepted.
- Show the result as the exact new text, then list facts you dropped or narrowed.

## Done when

- Every sentence passes the voice rules and no item from "Hard no" remains. Read the
  changed text once only for this check.
- Each claim about the project matches its spec or code, and each claim about a source
  matches that source at the linked commit. Overclaims found in the check are narrowed
  and reported.
- Every link resolves (file exists, or remote URL checked).
- The report names what was not verified, such as a diagram that was not rendered.

