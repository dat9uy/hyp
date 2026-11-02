---
status: accepted
---

# hyp runs and records every recipe command

In the session-replay recipe's early versions, a recipe drove the sandbox lifecycle and hyp was one step inside it. The operator ran the recipe's own scripts to pin skills, check the operator hyp, drive the agent and compute verdicts, so no record hyp wrote showed that any of that code had run. That is how unchecked instruments decided verdicts. We decided that hyp runs every executable script in a bundle from a `Run:` line, and records each run in the sandbox record or, when no sandbox exists, in a command record. The step a command sits under (Launch, Doctor, Drive, Cleanup, Evidence or Standalone) decides whether a sandbox exists when it runs, so a recipe never labels its commands. hyp's Go still learns no replay concepts: it runs a script with arguments, then records the exit code and output. Owner decision on 2026-10-28.

## Considered Options

- **Record only the output of code that can't run in a sandbox.** The operator would keep running the snapshot and re-test scripts by hand, and hyp would hash what they produce. Rejected because hyp could then show which inputs were used, but not that they were made by pinned code.
- **Move only the lifecycle under hyp.** Create, drive, remove and check would move, and the authoring scripts would stay outside any record. Rejected because it leaves the same gap for the standalone scripts.
- **Keep the recipe as a wrapper around hyp** (the shape up to @9). Rejected because every new harness recipe would rebuild the same wrapper, and the code it runs would stay invisible to hyp.

## Consequences

- When scripted, Evidence runs after Cleanup, on the kept files, and can be re-run from a later recipe version. Evidence gathered by hand while the sandbox is up stays as before.
- A command is whatever a `Run:` line names, and every other file in a bundle is hashed content ([ADR 0002](0002-known-answers-gate-every-recipe-run.md)). A bundle whose `recipe.md` lacks `format: 1` still hashes, but hyp refuses every use of it as a bundle missing a required field ([ADR 0007](0007-records-start-at-230.md)); `hyp recipe new` writes the line into every copy. "Frozen" means pinned by a tree. This replaces the unreached-script rule, which froze a bundle holding a script no command reached.
- hyp is not a job runner. Retries, background jobs and stdin streaming are out of scope. A future recipe that needs interactive or long-lived processes would reopen this decision for its Drive step.
