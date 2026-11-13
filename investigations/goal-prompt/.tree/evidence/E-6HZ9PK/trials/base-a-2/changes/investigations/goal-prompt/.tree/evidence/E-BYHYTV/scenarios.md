# Scenarios for Q-where, argued in the session of 2026-11-12 (not run)

1. A Goal body whose constraints block has CRLF line ends and blank separator lines. Command: a blank CRLF line is a lone carriage return, so the block is invalid and `hyp prompt` exits 1 and names the line. Skill: the writer copies the text by hand and may normalise it; nothing refuses.
2. A Goal body whose block has a line that starts with no heading. Command: exit 1 and names the line. Skill: no exit code; the prompt can carry the line or drop it.
3. A runner actor that already has a Result under the Goal. Command: exit 1 naming the actor, derived from the log. Skill: it must recount by hand from hyp show.

Owner decision, 2026-11-12: hyp prompt is a command that reads the log; no test is needed.
