---
name: hyp-owner-grill-eval
version: 2
from: hyp-owner-grill-eval@1 63adbc91c4d3f2d89cf474864cdccf286857d5637d5f3bd525be46317638a5df
format: 1
subject: One hyp-owner-grill session replayed by claude -p at a fixture (repository and hyp at its commit, the issue as of its time), with an arm's skills/hyp-owner-grill loaded and a stand-in owner answering from a paraphrase of the owner's positions; measures where the Goals it writes sit, their lines and constraints, what it asks the owner, and what it cost
---

# hyp-owner-grill-eval

One sandbox per trial. Launch makes the clone look like the repository the
grill saw at its fixture: the fixture's commit on one branch, the fixture's
skills with the arm's `hyp-owner-grill` over them, no later commit
reachable and no way to a remote. Drive lets the grill and a stand-in owner
take turns and keeps both sessions and every store at the end. `score`
measures one trial and writes the files a blind reviewer grades; `table`
sets the trials side by side. Neither prints a verdict.

Fixtures (the inputs file's `fixture`):

| fixture | commit | issue | stand-in |
|---|---|---|---|
| `a` | `a51f1408eb396f4b2aab2f2418d5f5bc4be9396c` | #113, `issue-a.md` (body and comments before 2026-11-11T16:20Z) | `standin-a.md` |
| `b` | `9c8e498b87683e50a94a410b6a3cfecdd7a0d441` | #141, `issue-b.md` (body and comments before 2026-11-12T03:45Z) | `standin-b.md` |

The stand-in files paraphrase the positions the owner took in the grill
sessions of each fixture; they quote none of them.

## Launch

Run: ./launch.sh

`hyp sandbox create` is given `--base` (the fixture's commit), `--hyp` (hyp
built from the fixture's commit) and `--inputs FILE`, whose `KEY=VALUE`
lines are:

- `arm`: the arm's label; `base` for the arm the others are compared with.
- `fixture` (`a` or `b`) and `trial`.
- `model`: the model of every call and every subagent; Doctor refuses one
  that is not a Sonnet model.
- `skills_commit`: the commit whose `skills/hyp-owner-grill` Launch loads;
  `skill_sha`: the sha256 of its `SKILL.md`.
- `skill_creator`: the absolute path of the ak-skill-creator skill
  directory, copied to `.claude/skills/`.
- `limit_min` and `max_exchanges`: Drive stops the trial after this many
  minutes or this many grill calls.

Launch lists every ref tip that is not in the fixture's history, copies
the fixture's `skills/` into `wt/.claude/skills/`, replaces
`hyp-owner-grill` there with the arm's, adds ak-skill-creator, keeps the
arm's skill as `skill.tar`, checks out branch `work` at the fixture,
deletes every other ref, every remote, the stash and the reflogs, prunes
unreachable objects, adds a `pre-push` hook that refuses, writes a git
config with no credential helper, `core.sshCommand=false` and dead push
URLs and an empty gh config directory, both in `$HYP_SANDBOX_TMP/trial-git/`, writes Claude Code settings that
enable its sandbox (Bash writes only inside `wt/` and reaches no network)
and deny reads, by Bash or Read, of every sandbox's `recipes/`, `inputs/`,
`patches/`, `tmp/hyp-owner-grill-eval/`, `sandbox.json`, `recipe.log` and
`manifest.json` and of `~/.claude/projects/`, so the grill cannot see the
stand-in's positions, the eval or a transcript,
copies the fixture's issue to `$HYP_SANDBOX_TMP/issue-<N>.md`, writes the
stand-in's system prompt (`standin-system.md` then the fixture's stand-in
file) and the grill's first message to `prompt.txt`:
`/hyp-owner-grill #<N> (the text of issue #<N> as of this session is in
<path>; this machine has no network, so gh cannot fetch it)`, and notes the
sha256 of `~/.local/bin/hyp`.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3`, `git` and the pinned `hyp` run; HEAD is
the fixture's commit; the model is a Sonnet model; the loaded
`hyp-owner-grill/SKILL.md` has `skill_sha`; ak-skill-creator's
`quick_validate.py` and `eval_skill.py` have the pinned sha256 values;
`.claude/` holds only `skills/`; no store is at the clone's top directory;
the only ref is `work`, there is no remote, no listed later commit is in
the clone and every commit is in HEAD's history; `prompt.txt` starts with
`/hyp-owner-grill `; the stand-in's prompt and the sandbox settings, with their read
denials, are in place; the trial's git config has no credential helper; and gh is not
logged in under the trial's config.

## Drive

Run: ./drive.py

The grill is `claude -p --model MODEL --setting-sources project --settings
<sandbox settings> --output-format json --permission-mode dontAsk
--allowedTools Bash Read Edit(./**) Write(./**) Glob Grep Skill Agent Task
TodoWrite` in `wt/`, started with `--session-id` on `prompt.txt` and
resumed with `--resume` on each stand-in reply. The stand-in is `claude -p
--model MODEL --setting-sources project --tools "" --system-prompt <the
stand-in's prompt> --output-format json` in an empty directory of its own,
resumed the same way on each grill message, prefixed with `The grill's
message:`. Every call runs without `HYP_*`, `CLAUDE_CODE_*`, `GH_TOKEN` or
`GITHUB_TOKEN`, with `CLAUDE_CODE_SUBAGENT_MODEL` and every
`ANTHROPIC_DEFAULT_*_MODEL` set to MODEL and `DISABLE_AUTOUPDATER=1`; the
grill also gets the trial's git and gh config, `GIT_TERMINAL_PROMPT=0`,
`GIT_SSH_COMMAND=false` and `GOCACHE` inside `wt/.git`. The loop ends when
the stand-in replies exactly `<<END>>` (`end`), a call fails or says
nothing (`grill_error`, `standin_error`), after `max_exchanges` grill calls
(`max_exchanges`), or at `limit_min` minutes (`limit_min`; the call in
flight is stopped).

Drive writes to this recipe's tmp: `calls.json` (one row per call: who,
exit, seconds, session, `total_cost_usd`, `num_turns`, models),
`drive.json` (arm, fixture, trial, model, `claude --version`, start time,
how the loop ended, exchanges, subagent count, `cost_usd`,
`cost_usd_literal`, the models the result records name, and whether
`~/.local/bin/hyp` is unchanged), `transcript.jsonl` (the grill's
session), `subagents.tar`, `standin.jsonl` and `stores.tar` (the `events/` of
every `.tree` under `wt/` at the end, outside `.git` and `.claude`). A resumed session's
`total_cost_usd` is its whole spend so far, so `cost_usd` is the sum over
the two sessions of each one's last `total_cost_usd`; `cost_usd_literal` is
the sum of `total_cost_usd` over every call. A trial stopped at
`limit_min` has neither. It exits 1 when the grill wrote no transcript.

## Evidence

Kept by hand at remove, for each of `state.txt`, `drive.json`,
`calls.json`, `transcript.jsonl`, `subagents.tar`, `standin.jsonl`,
`stores.tar`, `skill.tar`, `prompt.txt`, `claude.err` and `later.txt`:
`--keep tmp/hyp-owner-grill-eval/<file>`. Then `score` runs on that
OUT_DIR, and `table` on the score records, as Standalone commands.

## Standalone

Run score: ./score.py
External: OUT_DIR of hyp sandbox remove (the trial) and ak-skill-creator's scripts/ directory (runs its eval_skill.py, sha256 c47a38dee2456cc6172dff28472052c0f7b28d7baf310af6e83b5ca655d10a24, with PYTHONDONTWRITEBYTECODE=1); writes only $HYP_OUT

Run table: ./table.py
External: a reviews file and a label map (the blind reviewer's verdicts) and score records of this recipe (the trials); writes nothing

- `score OUT_DIR SCRIPTS_DIR` reads one trial. Trial nodes are the
  `node.add` events of every store at the end whose time is at or after
  Drive's start, read with `hyp events --json`. A trial Goal is a trial
  node of kind goal with lines. It writes `outputs/` under `$HYP_OUT`:
  `conversation.md` (the grill's session in order: each owner message,
  each grill text, each tool call with its working directory relative to
  the clone, and each tool result, clipped; a call that names
  `.claude/skills/hyp-owner-grill` has its result redacted, and the
  skill's own expansion is left out), `goals.md` (the stores at the end,
  each trial Goal with its store, actor, title, body and numbered lines,
  then each other trial node), and three files `{"pass": true|false}`:
  `has_goal.json` (a trial Goal exists), `l1.json` (a trial Goal exists,
  each sits in `investigations/goal-prompt` at fixture b and in
  `investigations/<one directory>` at fixture a, and no store is at the
  clone's top directory) and `l4.json` (a trial Goal exists and no line of
  one holds a run of 7 to 63 hex digits followed by `…` or `...`, or
  `sha256` followed within 12 characters by a run of 7 to 63 hex digits
  that ends there). It runs `eval_skill.py grade cases.json --case-id
  fixture-<f> --outputs outputs/ --run-id <arm>-<f>-<trial> --variant
  <arm> --skill <the arm's skill>` and keeps its stdout as `record.json`;
  exit 1 (rubric checks ungraded) is expected. It writes `facts.json` and
  prints `key<TAB>value` lines: `arm`, `fixture`, `trial`, `end`,
  `exchanges`, `timed_out`, `cost_usd`, `cost_usd_literal` (`none` when
  absent), `models` (every model an assistant message of the grill, its
  subagents or the stand-in names), `subagents`, `goals`, `stores`,
  `has_goal`, `l1`, `l4`, `eval_exit`. It exits 2 when OUT_DIR lacks a
  kept file, `eval_skill.py` is not the pinned one, or the grade exits 2.
- `table REVIEWS MAP SCORE...` reads score records (their `stdout` and
  `out/record.json`), a label map from each blind label to a run id, and
  the reviews, from each label to a `pass` or `fail` for each rubric
  assertion. It prints one row per arm and fixture: `trials`, the passing
  count of lines 1 to 6 (line 1: `l1-store` and `l1-init`; line 2: `l2`
  and line 3: `l3`, fixture a only; line 4: `l4-hash` and `l4-grader`;
  line 5: `l5`; line 6: `l6`, fixture b only; each also needs
  `has-goal`), the sorted `costs`, `median_cost` (the middle of the sorted
  costs, a trial with no cost counting highest), `ratio` (over the `base`
  arm's median at the same fixture), and the same two for
  `cost_usd_literal`. It exits 2 when a score lacks a field, a run id is
  missing from the map or scored twice, a review lacks a rubric
  assertion, or a fixture has trials and no `base` row.

`cases.json` is the eval in ak-skill-creator's format: one case per
fixture, whose assertions are the Goal's lines 1 to 6 as checks; the
`json-equals` ones read `score`'s pass files, and the rubric ones are
graded by a reviewer who did not run the trial and sees only `outputs/`
under a blind label.

`known-answers/` holds, for `score`, a fixture-b trial whose Goal sits in
`investigations/goal-prompt` with whole hashes (every pass file true), a
fixture-a trial with a store at the top directory and a line that names a
cut hash (`l1` and `l4` false), an OUT_DIR with no `stores.tar`, and a
scripts directory whose `eval_skill.py` is not the pinned one; for
`table`, two arms on both fixtures with one trial that has no cost, and a
set with no base arm. Every input was made by `fixture.py` with hyp; each
expected output was made by running the command once and checked by hand
against the rules above.

## Cleanup

Run: ./cleanup.sh

Stops the claude process group a Drive left running (SIGTERM, then
SIGKILL after 60 seconds). It succeeds when Drive never ran and when run
twice.
