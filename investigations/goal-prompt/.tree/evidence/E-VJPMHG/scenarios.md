# Q-NCY8HY: who writes the /goal prompt (grill session 285fa63e, 2026-11-11)

Owner decision, grill round 1 question 7: "agree" to H-MCDPNB (h-cmd), settled in the session.

- S1, determinism. The same log and flags must give the same prompt bytes, and a Goal in a two-Goal store must get the prompt its own events give. H-MCDPNB (h-cmd) is a read command over the log and meets S1 by construction. H-E5K8VG (h-skill) is agent-written text: two sessions on the same log give different wording (runs 1 to 6 of checkpoint-store:G-D1CH1R each differed in prose), so it cannot meet S1.
- S2, AGENTS.md "Improving hyp with hyp", step 5: a failure that can be detected from the tree goes into the binary as a refusal or a warning, not into skill text. A missing constraints heading is such a failure, and only a command can refuse it with exit 1.

Verdicts: H-MCDPNB supports; H-E5K8VG contradicts.
