# Review: 3ec4979..0cda2f8 (a run subject keeps its patch apart from its inputs)

Judged against ADR 0020 "A subject has one shape", Goal line G-D1CH1R#3, and the owner's answer to R-0HABWN#7.
go vet and go test ./... -count=1 pass at 0cda2f8 (I ran them in a detached temp worktree).

## Findings

1. A Hypothesis tied with `hyp subject H arm.patch --base B` is covered by a run at bare B that has no Arm and whose `--inputs` file has the patch's bytes. The code ran was bare B, so the patch never changed it. `hyp show G#1` prints `covered` / `ran on H`. Confirmed with a temporary CLI test.
2. The cause of 1 also works the other way: a Hypothesis tied to B plus an input file I is held by a run at B whose patch is I (`HoldsSubject([{B, Patch:I}], {B, [I]})` is true), so a run whose code was changed counts as B given data. The cause is that `HypothesisSubject` stores no patch-vs-input role, though ADR 0020 says "one shape". Confirmed with a temporary unit test.
3. Rule 27 in docs/replay-spec.md §5 still documents `bad subjects: a subject with no base and no files`. The code now emits `…no base, no patch and no files` and adds the rejection `bad subjects: patch "<x>" is not a sha256`, which the rule list does not name. This breaks the AGENTS.md rule of one spec per state machine. Confirmed by grep.
4. ADR 0020 still defines the run subject as `{base, files}` and gives no patch field. The code, replay-spec §4 and the glossary now define `{base, patch, files}`. The ADR, which is the acceptance source of this Goal, therefore disagrees with the record shape that ships. Confirmed by grep.
5. A Result written by a 7257753 or 3ec4979 binary stores an Arm's patch inside `files`, for example `{base:B, files:[P]}`. This commit replays it as "run at B given input P", so that run now covers bare B, which line 3 forbids. No such log exists in the repository, so this matters only for dev-binary stores. Confirmed by reading HoldsSubject; not run.
6. No test checks rule 26 after the owner's case: once a run at B given inputs holds a bare-B Hypothesis, a re-tie of that Hypothesis must be refused with `subject ran in`. Only coverage (`show`/`status`) is asserted for the new case. Confirmed by reading the tests.

## Unresolved question

- A sandbox with Arms keeps only patched subjects. If its Launch also builds bare B as a control, that run gives no bare-B subject, and a bare-B Hypothesis stays uncovered. This behaviour predates this diff; whether a control build is "a cited run" is for the owner to decide.
