# What a written or open Goal asks

The spec of Goal G-X0HY3A. It has no model: the change adds one output to one
read and moves no Goal between states.

## Line 1: `hyp status G` prints what the Goal asks

- `hyp status G`, with or without `--role`, prints the state block, then, while
  the Goal is `written` or `open`, a part under `The Goal asks:`.
- That part prints `# G: <words>`, then for each line in the Goal's order
  `## <key> (<kind>): <text>`.
- Under each line, every placed example in its current version prints
  `### <key> @<version>`, its tree hash, who placed it and when, its verdict
  and exit code on H0 at place, and its `example.txtar` byte for byte in a
  fence. The file holds the words, run, inputs, floors and ceilings.
- hyp reads the file from git by the placed hash, never from disk, so the part
  shows what was placed even when the checkout holds other files.
- A line with no placed example prints `no example placed`.

## Line 2: every other output stays as on H0

- `hyp status` for every Goal prints the state block alone, as before.
- The return of every act that writes prints the state block alone, as before.
  The return of `try` also does.
- `hyp status G` of a `selected` or `landed` Goal prints its record (the PR
  body, or the table of outputs), as before, and no `The Goal asks:` part.

## Line 3: the grill reads a cited Goal from hyp

- Step 1 of "How to grill" in `skills/hyp-grill/SKILL.md` tells the grill,
  when the Goal it drafts will cite a landed Goal with `--from`, to read
  `hyp status` of that Goal before it drafts the words and lines (step 8). A
  landed Goal's `hyp status G` prints its record: words, lines, examples and
  verdicts.
