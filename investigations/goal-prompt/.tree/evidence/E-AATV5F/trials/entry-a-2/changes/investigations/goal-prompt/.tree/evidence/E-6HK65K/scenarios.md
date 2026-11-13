# Scenarios for the prompt-form question (argued in the grill session of 2026-11-12, not run)

Rival 1 (H-108P9D): hyp prompt prints the prompt from the log.
Rival 2 (H-8TPB9X): a skill writes the prompt from hyp show and hyp status.

1. Same log, two prompts. A command reads only the log and the flags, so the bytes match. A skill's text varies between sessions, so a clause-for-clause check against a fixed clause list cannot rely on it. verdict: H-108P9D=supports H-8TPB9X=contradicts
2. Invalid constraints block. A command exits 1 and names the missing heading. A skill has no exit status and can print a prompt without the block. verdict: H-108P9D=supports H-8TPB9X=contradicts
3. Two Goals in one store. A command takes the Goal id and counts only that Goal's Results. A skill reads hyp status, which prints only a GOALS table without the stop sentence (internal/render/status.go:361 against :644), so it can name the wrong command. verdict: H-108P9D=supports H-8TPB9X=contradicts

The owner chose rival 1 in the session. A fresh reviewer's walk of the settled rules found the status-table gap in scenario 3.
