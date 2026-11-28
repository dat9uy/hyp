-------------------------------- MODULE rest --------------------------------
\* Two Goals: A, and B, which the owner adds with --from A. A's runner selects
\* A on a commit; the owner unselects, closes or lands A. B rests on A's select:
\* it opens on that commit when its examples give their verdicts there, and it
\* lands only after A. "own" is B's own H0, used when A had ended at B's add.
\* A commit in Bad is one where an example of B gives the wrong verdict.

EXTENDS Naturals

CONSTANTS FromNeedsEnded, OpenOnOwnH0, WrongVerdictSilent, LandAnyOrder,
          UnselectKeepsB, CutRowsUseBudget, CloseSendsBack, FromEndedWaits

VARIABLES a, aSel, aLanded, rejected, b, bH0, bWait, bRows, bUsed, bFromEnded, addRefused

vars == <<a, aSel, aLanded, rejected, b, bH0, bWait, bRows, bUsed, bFromEnded, addRefused>>

C == {"c1", "c2", "c3"}
Bad == {"c2"}
None == "none"
Budget == 3

Init == /\ a = "open" /\ aSel = None /\ aLanded = None /\ rejected = {}
        /\ b = "absent" /\ bH0 = None /\ bWait = "none" /\ bRows = {}
        /\ bUsed = 0 /\ bFromEnded = FALSE /\ addRefused = FALSE

AEnded == a \in {"landed", "closed"}
BInWork == b \in {"open", "selected"}

\* B opens on commit c of A, or waits for the owner on a wrong verdict.
OpenOn(c) ==
    IF c \in Bad /\ ~OpenOnOwnH0
      THEN /\ b' = "written" /\ bH0' = None
           /\ bWait' = IF WrongVerdictSilent THEN "A select" ELSE "wrong verdict"
           /\ UNCHANGED bUsed
      ELSE /\ b' = "open"
           /\ bH0' = IF OpenOnOwnH0 THEN "own" ELSE c
           /\ bWait' = "none"
           /\ bUsed' = IF CutRowsUseBudget THEN bUsed ELSE 0

\* command: hyp add goal
AddB ==
    /\ b = "absent"
    /\ IF ~AEnded /\ FromNeedsEnded
         THEN /\ addRefused' = TRUE
              /\ UNCHANGED <<b, bH0, bWait, bFromEnded, bUsed>>
       ELSE IF AEnded
         THEN /\ bFromEnded' = TRUE
              /\ IF FromEndedWaits
                   THEN b' = "written" /\ bH0' = None /\ bWait' = "A select"
                   ELSE b' = "open" /\ bH0' = "own" /\ bWait' = "none"
              /\ UNCHANGED <<addRefused, bUsed>>
       ELSE IF a = "selected"
         THEN /\ OpenOn(aSel)
              /\ UNCHANGED <<addRefused, bFromEnded>>
       ELSE /\ b' = "written" /\ bH0' = None /\ bWait' = "A select"
            /\ UNCHANGED <<addRefused, bFromEnded, bUsed>>
    /\ UNCHANGED <<a, aSel, aLanded, rejected, bRows>>

\* command: hyp run
ASelect(c) ==
    /\ a = "open" /\ c \notin rejected
    /\ a' = "selected" /\ aSel' = c
    /\ IF b = "written" /\ ~bFromEnded /\ bWait = "A select"
         THEN OpenOn(c)
         ELSE UNCHANGED <<b, bH0, bWait, bUsed>>
    /\ UNCHANGED <<aLanded, rejected, bRows, bFromEnded, addRefused>>

\* command: hyp unselect
AUnselect ==
    /\ a = "selected"
    /\ a' = "open" /\ aSel' = None /\ rejected' = rejected \cup {aSel}
    /\ IF ~bFromEnded /\ b \in {"open", "selected", "written"} /\ ~UnselectKeepsB
         THEN b' = "written" /\ bH0' = None /\ bWait' = "A select"
         ELSE UNCHANGED <<b, bH0, bWait>>
    /\ UNCHANGED <<aLanded, bRows, bUsed, bFromEnded, addRefused>>

\* command: hyp close
AClose ==
    /\ a \in {"open", "selected"}
    /\ a' = "closed" /\ aSel' = None
    /\ rejected' = IF aSel = None THEN rejected ELSE rejected \cup {aSel}
    /\ IF ~bFromEnded /\ b \in {"open", "selected", "written"}
         THEN /\ b' = "written" /\ bH0' = None
              /\ bWait' = IF CloseSendsBack THEN "A select" ELSE "A closed"
         ELSE UNCHANGED <<b, bH0, bWait>>
    /\ UNCHANGED <<aLanded, bRows, bUsed, bFromEnded, addRefused>>

\* command: hyp land
ALand ==
    /\ a = "selected"
    /\ a' = "landed" /\ aLanded' = aSel
    /\ UNCHANGED <<aSel, rejected, b, bH0, bWait, bRows, bUsed, bFromEnded, addRefused>>

\* command: hyp run
BRun ==
    /\ b = "open" /\ bUsed < Budget
    /\ b' = "selected"
    /\ bRows' = bRows \cup {bH0}
    /\ bUsed' = bUsed + 1
    /\ UNCHANGED <<a, aSel, aLanded, rejected, bH0, bWait, bFromEnded, addRefused>>

\* command: hyp land
BLand ==
    /\ b = "selected"
    /\ bFromEnded \/ LandAnyOrder \/ (a = "landed" /\ aLanded = bH0)
    /\ b' = "landed"
    /\ UNCHANGED <<a, aSel, aLanded, rejected, bH0, bWait, bRows, bUsed, bFromEnded, addRefused>>

Next == AddB \/ (\E c \in C : ASelect(c)) \/ AUnselect \/ AClose \/ ALand \/ BRun \/ BLand

Spec == Init /\ [][Next]_vars

\* source: G-FPFKYT:line-1
AddWhileInWork == ~addRefused

\* source: G-FPFKYT:line-2
OpensOnSelect ==
    BInWork /\ ~bFromEnded => bH0 \in C /\ bH0 \notin Bad

\* source: G-FPFKYT:line-3
WrongVerdictListed ==
    b = "written" /\ ~bFromEnded /\ a = "selected" /\ aSel \in Bad => bWait = "wrong verdict"

\* source: G-FPFKYT:line-4
LandsAfterPremise ==
    b = "landed" /\ ~bFromEnded => a = "landed" /\ aLanded = bH0

\* source: G-FPFKYT:line-5
NoWorkOnCut ==
    BInWork /\ ~bFromEnded => bH0 \notin rejected

\* source: G-FPFKYT:line-6
CutRowsCostNothing ==
    BInWork /\ ~bFromEnded => bUsed = IF bH0 \in bRows THEN 1 ELSE 0

\* source: G-FPFKYT:line-7
CloseWaitsForOwner ==
    a = "closed" /\ ~bFromEnded /\ b = "written" => bWait = "A closed"

\* source: G-FPFKYT:line-8
FromEndedOpensOnOwn ==
    bFromEnded => b # "written" /\ (BInWork => bH0 = "own")

=============================================================================
