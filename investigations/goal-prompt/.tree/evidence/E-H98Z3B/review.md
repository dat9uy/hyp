# Review of 5c5902d (hyp prompt reads the store's real path and states its line rules)

Method: read fix-diff.patch, prompt.go and prompt_test.go at 5c5902d; built 5c5902d and b4f23b2 binaries under /private/tmp/claude-501/ (worktree removed); ran probes (scripts in /private/tmp/claude-501/scr/); ran `go test ./internal/cli -run Prompt` on 5c5902d (ok) and with b4f23b2's prompt.go swapped in (only TestPromptThroughASymlinkedDirectory fails, so the new symlink test is red on base).

## Prior limits
- R-8JQPWG#1 (symlinked cwd): addressed. Probe: store at p/repo/tree reached via p/link and via p/other/l (a link inside another repo) prints "in tree" with exit 0, same bytes as the physical path.
- R-8JQPWG#2 (CRLF blank line): under the owner's answer ("CRLF bodies need not be accepted") refusal is allowed. Confirmed: body `Builds.\r\nMay change: a.\r\n\r\nRead only: b.\r\nNever: c.\r\n` gives `refused: ... starts with no heading: "\r"`, exit 1, nothing on stdout; `hyp add goal` exits 0 with the warning. Doc/skill/test now state the rule.
- R-8JQPWG#3 (Goal id in one of four places): addressed. wantNamesOnly now counts all G-ids (must be 5) and requires each to equal g.

## Findings

### K1 (line G-ZBM2RE#1, borderline; otherwise none): a store symlinked into a repository from outside now exits 3
Input: store directory /q/ext (outside any repo); /q/repo (has .git) holds `tree` -> /q/ext; cwd /q/repo/tree.
- b4f23b2: exit 0, "/goal Run the hyp tree in tree ..." 
- 5c5902d: exit 3, `hyp prompt G-...: no repository holds the store /private/tmp/claude-501/q/ext, so its path relative to the repository root is not known`
Expected: the run is started from inside the repository at `tree`, so a path relative to the repository root exists; the fix traded one symlink case (link outside the repo) for the reverse case (target outside the repo). Confirmed by running both binaries (script /private/tmp/claude-501/scr/p2.sh). No test covers this direction. Spec text says "read after resolving symbolic links" and so documents the new behavior, but it does not say exit 3 for this case.

### K2 (none): test skips silently when a repository sits above the temp directory
TestPromptOutsideARepository calls t.Skipf when any ancestor of t.TempDir() holds .git (e.g. a home-directory repo). The old test ran unconditionally; the exit-3 path then goes untested on such machines and `go test` still reports ok. Not run-confirmed (temp dir here has no ancestor .git). Low.

### K3 (none): CRLF body without a blank line is accepted but prints "\r" into the prompt
Input body `Builds.\r\nMay change: a.\r\nRead only: b.\r\nNever: c.\r\n` -> exit 0; `od -c` of stdout shows `a . \r \n`, `b . \r \n`, `c . \r \n`. This matches "byte for byte" in line 1 and the owner's answer, but the glossary/spec/skill text says only blank lines are the problem; a reader is not told that an accepted CRLF body leaks "\r" into the pasted /goal line. Confirmed. Low; text gap only.

### K4 (none): skills/hyp-goal-run hunk changes guidance beyond the fix
Old: "use a pinned recipe when the Goal names one". New: "Pin the gate as a recipe when the repository keeps one, so a script decides the verdict." This repository keeps a recipes/ directory (none gates `go test`), so a run here can read "the repository keeps one" as obliging it to pin one of them. The change is unrelated to the three limits the commit is meant to fix, and it changes the condition from Goal-named to repository-kept. Text defect; also not described in the commit subject.

### K5 (none): spec section 7.5 omits exit 3 for a failed symlink resolution
cmdPrompt returns herr.Env (exit 3) when filepath.EvalSymlinks(r.root) fails; the spec says exit 3 only "when no repository holds the store". Not run-confirmed (needs a store root removed mid-run). Low.

## Docs/skills hunks (text only)
- glossary and spec: consistent with the code (split on \n, space/tab/lone \r bad). No defect beyond K3, K5.
- hyp-owner-grill: "write the body with `\n` line ends" is accurate; "leave a blank line truly empty" is accurate. Exit 2 example "a --branch with a space" matches code.
- hyp-owner-answer: "Exit 2 is a bad argument" omits that exit 2 also covers an unknown or non-Goal id and a missing --actor; harmless. "Run it on the log ... with the answers merged or on the branch the prompt names" matches the grill text.
- hyp-goal-run: see K4.

## No defect found
--branch check (unicode.IsSpace covers U+00A0 and U+2028; probe: NBSP gives exit 2). A zero-width space (U+200B) or a leading "-" in --branch is accepted and printed on a single line; not a line break, not reported.
