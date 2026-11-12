---
name: run-review-replay
version: 1
format: 1
subject: A build Goal's unattended /goal run replayed from a store cut by claude -p, with an arm's skills/ loaded and the arm's hyp printing the prompt; measures whether each candidate it builds or selects has a review a subagent wrote after reading the candidate's diff, and what the run cost
---

# run-review-replay

One sandbox per trial. Launch makes the clone look like the repository the
replayed run saw at the cut: the base commit on the run's branch, the store
cut in the tree's directory, the arm's `skills/` loaded from
`.claude/skills`, no later commit reachable and no way to a remote. Drive
runs the prompt the arm's hyp prints as a `claude -p` session and keeps its
transcript, its subagents' transcripts, the store at its end and every
commit of the clone. `score` measures one trial and `table` sets the trials
side by side. Neither prints a verdict.

## Launch

Run: ./launch.sh

`hyp sandbox create` is given `--base` (the commit the replayed run's code
was at), `--store` and `--cut` (the store and the event to cut it at),
`--hyp` (the arm's hyp; its default `--experiments` is the arm's N) and
`--inputs FILE`, whose `KEY=VALUE` lines are:

- `arm`: the arm's label; `base` for the arm the others are compared with.
- `cut` (`a` or `b`) and `cut_event`: the cut's label and its event id.
- `trial`, `model` (the runner's model), `actor` (the runner's actor).
- `goal` (the replayed Goal's id), `branch` (the branch its prompt names)
  and `store` (the tree's directory, relative to the repository root).
- `skills_commit`: the commit whose `skills/` Launch loads; `skill_sha`:
  the sha256 of that `skills/hyp-goal-run/SKILL.md`.
- `skill_creator`: the absolute path of the ak-skill-creator skill
  directory the replayed Goal names, copied to `.claude/skills/`.
- `limit_min`: minutes from the runner's start until Drive stops it.

Launch lists every ref tip that is not in the base's history, loads the
arm's skills and ak-skill-creator into `wt/.claude/skills/`, moves the
store cut from `wt/.tree` to `wt/<store>/.tree` (keeping the base's
tracked `.gitignore`), checks out `branch` at the base, deletes every
other ref, the stash and the reflogs, prunes unreachable objects, adds a
`pre-push` hook that refuses, writes a git config with no credential
helper, `core.sshCommand=false` and dead push URLs, makes an empty gh
config directory, prints `hyp prompt <goal> --actor <actor> --branch
<branch>` in `wt/<store>` to `prompt.txt`, and notes the sha256 of
`~/.local/bin/hyp`.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3`, `go`, `git` and the pinned `hyp` run;
the loaded `hyp-goal-run/SKILL.md` has `skill_sha`; ak-skill-creator's
`quick_validate.py` and `lint_cruft.py` have the sha256 values the
replayed Goal pins; `.claude/` holds only `skills/`; the only ref is
`branch`, there is no remote, no listed later commit is in the clone and
every commit is in HEAD's history; the store's last event is `cut_event`
and `hyp status <goal>` runs there; `prompt.txt` starts with `/goal ` and
names the actor; the trial's git config has no credential helper; and gh
is not logged in under the trial's config.

## Drive

Run: ./drive.py start

Run: ./drive.py finish

`start` runs `claude -p --model MODEL --setting-sources project
--output-format json --permission-mode dontAsk --allowedTools Bash Read
Edit Write Glob Grep Skill Agent Task` in `wt/`, with `prompt.txt` on
stdin, in its own session, without `HYP_*`, `CLAUDE_CODE_*`, `GH_TOKEN` or
`GITHUB_TOKEN`, with `GIT_CONFIG_GLOBAL` set to the trial's git config,
`GIT_CONFIG_NOSYSTEM=1`, `GH_CONFIG_DIR` set to the empty gh config,
`GIT_TERMINAL_PROMPT=0`, `GIT_SSH_COMMAND=false` and
`DISABLE_AUTOUPDATER=1`. It waits up to 55 minutes, or `limit_min` when
that is shorter. `finish` waits until `limit_min` minutes after the start,
then stops the runner, and writes to this recipe's tmp: `drive.json`
(exit, `claude --version` at the start, `timed_out`, `num_turns`,
`cost_usd` from `total_cost_usd`,
`duration_ms`, permission denials, subagent count, and whether
`~/.local/bin/hyp` is unchanged), `transcript.jsonl`, `subagents.tar` (the
session's `subagents/` transcripts), `tree.tar` (`wt/<store>/.tree` at the
end) and `commits.json` (every commit in the clone: parents, and each file
it changes against its first parent with status and blob id). It exits 1
when no transcript was written. A trial stopped at `limit_min` has no
result record, so no cost.

## Evidence

Kept by hand at remove, for each of `state.txt`, `prompt.txt`,
`drive.json`, `transcript.jsonl`, `subagents.tar`, `tree.tar`,
`commits.json`, `claude.err` and `later.txt`: `--keep
tmp/run-review-replay/<file>`. Then `score` runs on that OUT_DIR and
`table` on the score outputs, as Standalone commands.

## Standalone

Run score: ./score.py
External: OUT_DIR of hyp sandbox remove (the trial); writes nothing

Run table: ./table.py
External: score outputs of this recipe (the trials); writes nothing

- `score OUT_DIR` prints `key<TAB>value` lines for one trial: `arm`,
  `cut`, `trial`, `actor`, `n` (the N in `prompt.txt`), `cost_usd`
  (`none` with no result record), `duration_ms`, `num_turns`,
  `timed_out`, `subagents`, `built`, `reviewed`, `selects`,
  `unreviewed_unnamed`, and 0/1 flags `a1`, `a1_strict`, `b2`,
  `b2_strict`, `l4`, `l4_strict`. Trial events are the store's applied
  events after `cut_event`. A candidate is a Hypothesis under the Goal
  with a subject commit; built means its node.add is a trial event. A
  review Result of a candidate is a trial Result that gives it a verdict
  and cites a file that one subagent transcript of the session wrote (a
  Write to that path, or the file holds a 200-character window of the
  subagent's last text) after it read the candidate's diff (its prompt or
  tool results hold `diff --git a/P b/P` and `..` plus the first 7 hex
  digits of P's blob at the commit, for every file the commit changes
  against its first parent and keeps). `a1`: every built candidate has a
  review Result; `a1_strict`: also at least one was built. `b2`: a trial
  select of a candidate, not later replaced or undone, after its review
  Result, with at most N Results under the actor before it; `b2_strict`:
  fewer than N. `l4`: each live candidate at the end (proposed, supported
  or inconclusive) with no review Result is named in the actor's last
  trial checkpoint (summary or refers) by display id, alias or a 7+ hex
  prefix of its commit; `l4_strict`: the same over live and built
  candidates. It exits 2 when OUT_DIR lacks a kept file, the cut event or
  the Goal is not in the store, or `prompt.txt` states no N.
- `table SCORE...` prints one row per arm and cut: `trials`, the sorted
  `costs`, `median_cost` (the middle of the sorted costs, a trial with no
  cost counting highest), `ratio` (over the `base` arm's median at the
  same cut), the passing counts of `a1` and `a1_strict` (cut a) and `b2`
  and `b2_strict` (cut b), and the failing counts of `l4` and
  `l4_strict`. It exits 2 when a score lacks a field or names another cut,
  or a cut has trials and no `base` row.

`known-answers/` holds, for `score`, a cut-a trial whose built candidate
has a review (all flags pass), a cut-a trial whose review read another
commit's diff and that builds a second candidate it never reviews or names,
a cut-b trial that reviews and selects within N, the same with N 1 (so only
the strict flag fails) and no cost, a cut-b trial whose Result cites a gate
log and not a review but whose checkpoint names the candidate's commit, and
an OUT_DIR with no `tree.tar`; for `table`, three arms over two cuts with
missing costs, and a set with no base arm. Every input was made by
`fixture.py` with hyp and git, and every expected output was written by
hand from the rules above.

## Cleanup

Run: ./cleanup.sh

Stops the runner's process group when it still runs (SIGTERM, then
SIGKILL after 60 seconds). It succeeds when Drive never ran and when run
twice.
