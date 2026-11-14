# G-4S36M5 A write command never appends text that cites a node id this store does not hold, without the operator being told — complete

## Q-DVZZHY Which rule stops a write command from appending text that cites an id that is no node? — answered

Selected: H-88ZF3E
Design: Every write command scans --why, --body, --observation, --summary and each --limit for id-shaped tokens: at a word boundary, G-, Q-, H-, E- or R- (upper case) then six Crockford-base32 upper-case characters, or a placeholder mark (<, {, …, ..., or three or more X). Lower-case prose like E-mail is never id-shaped. A token not preceded by <name>: that does not resolve in this store refuses: nothing is appended and hyp exits 2, as for unknown node.

Limits:

- R-BV4CWV#1 accepted by owner-dat — The fixture is hand-made apart from K1, and small: one store, and request-verify for most cases.
- R-BV4CWV#2 accepted by owner-dat — K7 is unscored for the scan arms: a scan cannot see an id that is not written.
- R-BV4CWV#3 accepted by owner-dat — The arms are prototypes; the delivered rule must be reproduced in Go and give the selected arm's rows on every case.
- R-BV4CWV#4 accepted by owner-dat — h-cite's cases run request-verify only; select and confirm take --result in the arm but no case runs them. Under h-cite the why still holds R-<new> (C1a).
- R-BV4CWV#5 accepted by owner-dat — The arm binaries are read from absolute paths outside the bundle, which the known-answer gate cannot see.

- H-RAJAFK Scan text flags for id-shaped tokens; warn on an unresolved one and still append — supported
- H-88ZF3E Scan text flags for id-shaped tokens; refuse an unresolved one — supported
  - E-YKJ70D Each rule's arm on the frozen fixture: exit, append and warning per case — run (recipe id-citation-arms@1, run and rule)
    - R-BV4CWV recorded H-RAJAFK=supports H-88ZF3E=supports H-C5BA8J=supports H-BMN916=contradicts: id-citation-arms@1 cases on the frozen fixture: the today arm appended every slip K1-K6 (K1 is the rebuilt event) with exit 0 and no warning. text-warn exited 0, appended and warned on K1-K6; text-refuse exited 2 and appended nothing on K1-K6; both gave exit 0, appended, no warning on G1-G5. cite refused C1a and C1b (exit 2, nothing appended), accepted C1c, and matched today on G1-G5. K7 (prose) passed every arm unflagged, unscored. verdict printed supports for three rivals: an owner trade-off.
- H-C5BA8J request-verify, select and confirm take --result REF, resolved or refused — supported
- H-BMN916 No change: today's hyp is enough — contradicted
