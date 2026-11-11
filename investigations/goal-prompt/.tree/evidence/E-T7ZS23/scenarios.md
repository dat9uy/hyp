# Q-MAKJ8R: does the /goal prompt carry the owner's answers (grill session 285fa63e, 2026-11-11)

Owner decision, grill round 2 question 2: "settle H-none", in the session. The scenarios were traced in the code by an independent reviewer and checked against the source.

- S1. internal/replay/replay.go:57 counts every checkpoint whose role is not owner as the Goal's last runner checkpoint. skills/hyp-owner-answer "End each turn" records a runner checkpoint after the owner's answers. H-CHKW8Y (h-checkpoint) then prints no answers while a returned limit is still owed: contradicted.
- S2. A run that writes a refer-only checkpoint after an answer and stops before closing the returned limit also hides that answer under H-CHKW8Y.
- S3. One answer event may settle keys of two Goals. Filtering by event leaks the other Goal's key. H-9Y4F4H (h-state) avoids this only by filtering each key by its Goal.
- S4. At the 14:31 cut of checkpoint-store:G-D1CH1R, hyp status --as-of 2026-11-11T14:32:00Z lists "runner: Q-PC04ED limit R-9SMRB2#4" under OWED: the returned limit is already the runner's item with no prompt text, and an accept leaves the runner nothing. H-9Y4F4H prints a second copy of what OWED lists, a derived fact stored in the prompt that can disagree with the ledger. H-RSDX0H (h-none) prints nothing OWED holds, and keeps the prompt within the 4000-character hand-off limit, since a Goal body is at most 2000 runes.

Verdicts: H-RSDX0H supports; H-9Y4F4H contradicts; H-CHKW8Y contradicts.
