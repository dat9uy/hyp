# Q: what a Result keeps about its runs (session e6a35a24, 2026-11-11)

S1, a patched run. One Result cites a run at B with patch P and a run at C.
- One subject per run keeps {B, [P]} and {C, []}. No entry is bare B, so a Hypothesis whose subject is bare B is not covered. Holds.
- Merged lists keep bases {B, C} and files {P}. B is in the list and the Hypothesis names no files, so bare B reads as covered though it never ran alone. Breaks.
- One base per Result cannot cite both runs. S1 is refused, not wrong.

S2, red on base. The line "named tests red on B, green on C" is one claim about two runs.
- One subject per run: one Result cites both runs and closes the line; both halves are in its cited files. Holds.
- One base per Result: R-1 at B cannot cover the candidate at C; R-2 at C cannot cite B's sandbox.json. No single Result closes the line with both halves checked, unless a Result may cite a Result, a new concept. Breaks.

Old Results read as one entry with no base and all their fingerprints, so existing coverage is unchanged.
Owner's decision in the session: "A it is then" (one subject per cited run).
