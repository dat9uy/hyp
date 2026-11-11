# Fresh review of git diff 10566d4 b4f23b2 (general-purpose subagent, default model), findings in brief

Given: the fix diff, the findings F1-F3, F5, F7, F9, F10 it claims to fix, design G-GZR6C6, G-ZBM2RE's lines verbatim, the settled selects with the owner's words, AGENTS.md rules. Not given: the runner's Results.

Checked correct: F1 (a tab-only or space-only line, also a trailing one, exits 1 and add goal warns); F2 (a .git file of a worktree is found; invalid block outside a repository exits 1 before the exit-3 check); F3 (newline and space branches exit 2); the F1-F3 tests each fail on 10566d4; --recipe none --rubric reading is a valid pair.

G1 (line 1, low-medium) prompt.go:86-97,148-151: repoRoot walks up from r.root, taken from os.Getwd() (logical $PWD), symlinks unresolved. A store reached through outside/store -> repo/inv/t exits 3; through cd -P it exits 0 with "in inv/t". Same log and flags, different exit and bytes. An ancestor of the symlink holding .git would give exit 0 and a wrong path (not probed).
G2 (line 3, low) prompt.go:42,47: lines split on "\n"; a CRLF body's blank line is "\r", a bad line. Body "Builds.\r\nMay change: a.\r\n\r\nRead only: b.\r\nNever: c.\r\n": add goal warns, prompt exits 1 naming "\r". Accepted before the fix. Consistent with the literal rule; no spec or skill says a CR-only line is bad.
G3 (none, low) prompt.go:128-130: IsSpace/IsControl refuse names git accepts (no-break space, U+0085, U+2028) with the message "a branch name has none", which is wrong for them; names git refuses (a..b~^:, -x) are accepted. A guard against injected lines, not a branch validator.
G4 (line 6, medium) skills/hyp-goal-run/SKILL.md:31: tells a build Goal's Experiment that runs the repository's own gate to use --recipe none --rubric reading; RECIPE-FORMAT.md:786-787 says --rubric reading is "for a verdict no script can decide", and replay-spec.md:248 says a script whose output a verdict counts lives in a bundle. A gate's exit status is script-decidable: the new rule conflicts with the skill it names and the spec.
G5 (line 3, low) skills/hyp-owner-grill/SKILL.md:104 and docs/glossary.md:38 say "every non-empty line" without saying a line of spaces, a tab or a CR is non-empty; the spec now says it.
G6 (none, low) skills/hyp-owner-answer/SKILL.md:143 runs hyp prompt with no rule on which log; neither skill says what to do on exit 3 (store outside a repository) or exit 2 (bad --branch).
G7 (line 1 test, low) prompt_test.go:80-86: the exit-3 test passes only if no directory above $TMPDIR holds .git.
G8 (lines 2 and 4 tests, low) prompt_test.go:91-101,118: wantNamesOnly is called with no other ids in the answer test, and checks one of the four places the Goal id appears; both new assertions pass on 10566d4.
