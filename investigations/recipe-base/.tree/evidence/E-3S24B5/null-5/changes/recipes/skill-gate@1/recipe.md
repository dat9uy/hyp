---
name: skill-gate
version: 1
format: 1
subject: A change to a hyp skill's text (skills/*/SKILL.md), judged by fresh claude runners started from one recorded point of a hyp tree
---

# skill-gate

The instrument of the pre-registered method "a skill-text gate"
(`METHOD.md`). A change to a skill's text ships only when `base` shows a
failure in 3 or more of 5 trials and the changed arm in 1 or fewer of 5.
Three Standalone commands, all run with `hyp recipe run --record DIR
skill-gate@1 NAME ARGS` (every path in ARGS absolute):

- `trial` runs one trial and keeps its evidence;
- `score` reads one trial's evidence and prints `key<TAB>value` lines;
- `verdict` reads the scores of every trial and prints the verdict.

## Standalone

Run trial: ./trial.py
External: ARGS = REPO COMMIT STORE CUT ARM VARIANT TRIAL BRIEF HYP CLAUDE [TIMEOUT_SECONDS]. Reads the git repository REPO at COMMIT, the hyp store STORE up to event CUT, the skill variant directory VARIANT (`none` for the arm `base`), the BRIEF file, and runs the binaries HYP and CLAUDE; makes and removes one sandbox under $HOME/.hyp-sandbox; writes only $HYP_OUT

`trial` makes one hyp sandbox from REPO at COMMIT with STORE copied in at
CUT, strips agent files, lays the arm's variant over the two skills, runs a
fresh `CLAUDE -p --model sonnet --setting-sources project` with BRIEF on
stdin and the sandbox as its working directory, stops it after TIMEOUT
seconds (50 minutes, `3000`, when the argument is left out; a real trial
leaves it out), then removes the sandbox keeping its changes. It exits 0
whatever the runner did (a runner that failed or timed out is evidence);
it exits 1 when the sandbox could not be made or removed, and 2 on a bad
argument.

- `ARM` matches `[a-z][a-z0-9-]{0,31}` and is not `repro`. `base` takes
  `VARIANT` = `none`. Any other arm takes a directory holding
  `<skill>/SKILL.md` for each skill it changes, where `<skill>` is a
  directory of `skills/` in the sandbox. Any other entry is refused.
- `TRIAL` is a positive integer; the trial's number within its arm.
- `CLAUDE` is a wrapper or the `claude` binary; it gets the flags above plus
  `--permission-mode bypassPermissions --output-format stream-json
  --verbose`, so the stdout stream is the transcript. The caller's wrapper
  carries its credentials, because the check's `HOME` is a scratch one.
- Before the runner starts, the trial deletes `AGENTS.md`, `CLAUDE.md`,
  `.agentkit`, `.claude` except `.claude/skills`, and the `plans` and
  `docs/plans` directories; copies `skills/<skill>` to `.claude/skills/<skill>`;
  then lays the variant over both `skills/` and `.claude/skills/`. The
  runner's `PATH` starts with the sandbox's pinned HYP, the same in every
  arm; `HYP_ACTOR` is `runner` unless set.
- Kept in `$HYP_OUT`: `meta.json` (arm, trial, commit, cut), `setup.json`
  (the paths the trial itself deleted or laid, which `changes.json` also
  lists), `transcript.jsonl` (the runner's stream), `summary.json` (exit,
  timed_out, turns, cost), `new-events.json` (type and kind of each event
  the runner's hyp applied after the cut, from `hyp events --json` before
  and after), and from remove `changes.json` and `changes/` (every file
  the runner added or changed, its tree's events included).

Run score: ./score.py
External: reads the trial evidence directory given as the argument (the directory `trial` wrote); writes nothing

`score DIR` prints, in this order, `arm`, `trial`, `skill_loaded`,
`failing_hyp`, `f_reading_pin`, `f_body_limit` and `end_state`, as
`key<TAB>value` lines. It exits 2 when DIR lacks a file or holds one that
does not parse.

- `skill_loaded`: 1 when the transcript holds a `Skill` tool call or a
  `Read` of a file named `SKILL.md`, else 0.
- `failing_hyp`: the number of `Bash` calls whose tool result is an error and
  whose command's exit status is hyp's: the last command of the call (after
  the last `;`, `&&`, `||`, `|`, `&` or newline outside quotes, with heredoc
  bodies, leading `VAR=value`, `!`, `time`, `command`, `exec`, `env` and
  brackets skipped) has `hyp` as its program. A pipeline into `tail`, a
  `hyp ...; echo`, `hyp ... || true` and a trailing `&` return another
  status and do not count.
- `f_reading_pin`: 1 when an errored tool result holds
  `--decided-by reading contradicts recipe`.
- `f_body_limit`: 1 when an errored tool result holds
  `over limit: body has N runes (max 600)`.
- `end_state`: `recorded` when `new-events.json` holds a `result.record`
  event, else `registered` when it holds a `node.add` of kind `Experiment`,
  else `none`.

Run verdict: ./verdict.py
External: reads the *.tsv score files in the directory given as the first argument; writes nothing

`verdict SCOREDIR ROW...` reads every `*.tsv` file in SCOREDIR (one trial's
`score` output each) and the rows to judge (`reading_pin` or `f_reading_pin`,
`body_limit` or `f_body_limit`). It needs the arm `base`, exactly 5 trials
per arm and distinct trial numbers within an arm, else it exits 2 and prints
nothing on stdout. For each row and arm, k is the number of trials whose
fingerprint is 1, where a trial that never loaded a skill counts against its
arm: for `base` it is never a reproduction (not counted), for any other arm
it always counts as a failure (counted). A row reproduces when `base` has
k >= 3; a variant arm passes a row at k <= 1. It prints `k: ARM ROW=N`
lines, then `verdict: repro=supports` when any row given reproduces, else
`verdict: repro=contradicts`; and for each variant arm, sorted by name,
`verdict: ARM=supports` when it passes every reproducing row, `contradicts`
when it fails one, and `inconclusive` when no row reproduces.
