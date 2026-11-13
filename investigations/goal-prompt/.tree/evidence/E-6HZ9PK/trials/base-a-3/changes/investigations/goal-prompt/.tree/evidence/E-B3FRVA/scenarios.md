# Scenarios for Q-PZ1M37, argued in the session of 2026-11-12 (owner: owner-standin)

verdict: H-BM50HC supports
verdict: H-QZMAYS contradicts

(a) The same log read twice. A command that reads the log prints the same bytes. Prose written by a skill each session can differ in wording, so the run can not compare two prompts.
(b) A Goal whose body has no valid constraints block (for example checkpoint-store:G-D1CH1R). A command exits 1 and prints no prompt. A skill has no exit code, so a run could start without stated limits.
(c) A store with two Goals. A command reads one Goal's OWED and body only. A skill that reads `hyp status` could copy the other Goal's items.

Decision: hyp prints the prompt from a command that reads the log. The prompt carries no owner answer, because the runner reads OWED. Source: issue #113 and the owner's answers in this session.
