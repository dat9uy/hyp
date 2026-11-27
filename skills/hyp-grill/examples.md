# The examples

The examples part of hyp-grill. Use it when the owner agreed to the Goal's
words and lines, before the owner adds the Goal: the add carries the examples,
and the Goal is open at once. You propose the examples; the owner chooses their direction:
which reading of each line an example checks. A reading is one part or one
meaning of a line's words. Ask in two rounds: first the readings of every line,
then, after you write the floors, the split question. Ask no question about
ceilings: a ceiling is text a build really printed, from H0 here (step 7) or
from candidates at the re-grill.

While a question is open, each message that ends your turn repeats it in
full: the owner may read only the last message, not the status before it.

1. **Show the readings of each line, and let the owner choose.** Start the
   part of the message for each Goal line with one line that gives its key and
   words, and put its readings under it. Each reading shows a case, what the
   owner would do, and the output it expects, so that the owner chooses between
   outputs, not between descriptions:
   - **Where a model exists**, each variant that no example separates is one
     reading of the line its rules come from. Show the case as two outputs,
     the model's print and the variant's print, each marked "model print, not
     built", and ask which output is wanted.
   - **A line no variant reaches** gets its readings from its words: name the
     parts or meanings the line can check, show beside each the output you
     expect, marked "not built", and ask which readings the line means. A
     reading in words alone is not enough to write a floor from.

   A model print is a direction, never a verdict: the example written from the
   choice is still tried on H0 before the owner adds it (step 7), and a model result
   never counts at select. Use plain English and no model word other than the
   "model print, not built" mark.
2. **Write one example per chosen reading.** Its `words` section is the
   reading in the exact words the owner picked, so the select report says what
   each example checks. Its `run` is the command the owner would type (for
   example `sh greet.sh`). Write floors from the part where the wanted output
   differs from the other, as short literals: a model print is a guess at the
   real text, so a long literal copied from it fails on bytes the owner never
   chose. Never copy the rejected output into a ceiling: a wrong build prints
   the same fault in other words. A rejected case the owner wants checked is
   its own example (for example "on a branch, no warning"), whose ceiling
   reuses a literal the owner wants elsewhere, such as a sibling's floor.
   When the owner answers "either output is fine", write no example for that
   reading and list it under "Not covered". The choice is the decision: do not
   ask the owner to approve the file again.
3. **Two things to see is two rules.** When the readings the owner chose for
   one line are separate things to see (for example a list and a count), warn
   that the line combines two rules, and propose to split it into two lines.
   Ask the owner whether to split; the owner decides. Raise a split only after
   the owner's choice names two things, never in advance. A split changes the
   lines, so the owner records it while the Goal is `written`. When the owner
   keeps one line, it holds one example per reading.
4. **One folder per example**, in a draft folder on the record branch, at
   `goals/draft-<name>/<line>/<example>/`, with `<line>` the key the line will
   get (`line-1` for the first), and with an `example.txtar` and a `probe/`
   folder. The owner's add copies the set to `goals/<G>/`. A new reading is a new example key,
   never a new version of an existing one. The `example.txtar` sections are
   `words` (the reading the owner chose), `run` (the command run on the subject),
   `input/<path>` files, and at least one `floor: contains` or `floor: equals`,
   with any `ceiling`. Floors and ceilings are literals, byte for byte.
5. **Give the generic Probe to each example it can run.** The generic Probe,
   [probe/run](probe/run), runs `$HYP_RUN` in a copy of the subject, prints its
   stdout, then a last line `exit N`. It can run an example when the example's
   command, run in a plain copy of the subject, shows what the owner sees.
   Copy it with `cp` into the example's `probe/run`, byte for byte, and keep it
   executable.
6. **Hand every other example to a Probe writer.** When the reading needs a
   state set up before the command (for example "in a git repository with
   three commits"), the generic Probe cannot run it. Follow
   [probe-writers.md](probe-writers.md): write no Probe code yourself, start a
   Probe writer from the block `hyp status G --role probe-writer` prints, and
   bring its commit onto the record branch.
7. **Run `hyp try` before the owner adds the Goal.** Do these in order:
   - **Try each example on H0.** A gap example must fail on H0 and a guard
     example must be met on H0. Run
     `hyp try --examples goals/draft-<name> --base COMMIT LINE/EXAMPLE`, with
     COMMIT the H0 the add will name, or `hyp try G LINE/EXAMPLE` for a new
     example of an added Goal, and read its verdict, exit code, output and
     stderr. It records nothing and uses no
     run. When the verdict is not the one the line needs, fix the example and
     try again.
   - **Add the H0 ceilings.** A gap example's H0 output shows the behaviour
     the line wants gone: text of it that a half-fixed build could still print
     becomes a ceiling, copied byte for byte. Name each such ceiling in the
     hand-off with its H0 text; the owner may drop it.
   - **Read the findings again.** Before the owner adds the Goal, run
     `hyp status` and `hyp finding list` again: each finding the Goal will
     name, and each finding recorded during the grill whose fact bears on it,
     is new evidence. It becomes an
     example on the line its fact bears on, which names the finding in a
     `findings` section, one finding id per line, or the hand-off says why no
     line needs it. hyp refuses the set while a finding that the Goal names at
     `add goal` has no example that names it. Check that the words of each
     example that names a finding state that finding's fact, in words the
     owner reads.
   - **The owner adds the Goal with the whole set.** When every line holds
     its examples, commit the draft folder and draft the owner's add with
     `--examples goals/draft-<name>`. hyp runs each example on H0 and records
     the Goal, open at once, with each example's hash and verdict, only when
     every example gives its line's verdict there. Otherwise it records no
     Goal and names each example that disagrees, each line that holds no
     example and each folder it cannot read, with the guard that refused it.
     Fix every cause and draft the add again. A new example of an open Goal
     is the owner's act too: the place line `hyp status G --role owner`
     prints, with `--examples` naming the folder that holds the set, under
     the same rules. A placed example stays in the set: never remove its
     folder.
8. **Stop when the Goal has enough.** The Goal has enough when every line
   holds an example that gives its line's verdict on H0 (`hyp status G` shows
   the Goal `open`), and every variant of the model is separated by an example
   or the owner said either output is fine (listed under "Not covered"). Then
   stop asking: say "This is enough.", and give the hand-off.
   Ask no further question and offer no more examples; more examples come only
   from real outputs.
9. **Report the surfaces no example covers.** List every surface each line
   reaches: each file, output or command where the line's words can be true
   or false. Each surface that no example covers goes under "Not covered" in
   the hand-off. Listing a surface is not a question to the owner.
