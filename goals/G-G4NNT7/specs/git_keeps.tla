----------------------------- MODULE git_keeps -----------------------------
(***************************************************************************)
(* Goal G-G4NNT7: "Git keeps only the owner's acts, the place of each      *)
(* example set and the select, while every cell, Result and output stays   *)
(* in the clone's trace as a cache that a reader fills by running an       *)
(* example again, and hyp warns when a selected commit is on no branch or  *)
(* tag."                                                                   *)
(*                                                                         *)
(* Two clones share one git history: clone A, where the grill places and   *)
(* the runner runs and selects, and clone B, a fresh clone. Git here is    *)
(* the merged default branch, so B sees every line A commits. Each clone   *)
(* has its own trace. The Goal has one example, line-1/a.                  *)
(*                                                                         *)
(* Variants, each FALSE in git_keeps.cfg, are cheaper builds:              *)
(*   PlaceCellsInGit    place still writes H0's cell and output to git     *)
(*   RunInGit           run still writes its Result to git                 *)
(*   SelectNoHashes     select names the commit but no example hash        *)
(*   MissingNoTryLine   a missing cell says so, with no try line           *)
(*   MissingAsNotRun    a missing cell prints as `not run`                 *)
(*   TryNoSave          try saves nothing in the trace                     *)
(*   RerunUnmarked      a rerun cell prints as the cell the owner saw      *)
(*   SameSubjectQuiet   no warning when a commit with the same subject     *)
(*                      hash is on a branch or tag                         *)
(*   WarnOnlyStatusG    the warning prints only in `status G`              *)
(*   BudgetFromGit      every clone counts the runs of every clone         *)
(*   RunNeedsH0Cell     run refuses in a clone whose trace lacks H0's cell *)
(* Left out: line-7 (landed Goals print as on the base) is about the old   *)
(* lines of the log, not a step.                                           *)
(***************************************************************************)
EXTENDS Integers

CONSTANTS
    PlaceCellsInGit, RunInGit, SelectNoHashes, MissingNoTryLine,
    MissingAsNotRun, TryNoSave, RerunUnmarked, SameSubjectQuiet,
    WarnOnlyStatusG, BudgetFromGit, RunNeedsH0Cell

Clones == {"A", "B"}

VARIABLES
    gitPlace,    \* git holds the place line of the set
    gitCells,    \* git holds a cell or an output (from place or run)
    gitSelect,   \* "none", "hashes" or "nohashes": the select line in git
    landed,      \* git holds the owner's land
    trH0,        \* per clone: H0's cell in the trace: "none", "cell", "rerun"
    trCand,      \* per clone: the candidate's cell: "none", "cell", "rerun"
    runs,        \* per clone: Hypotheses run in that clone's trace (0 or 1)
    reach,       \* the selected commit is on a branch or tag
    sameSubj,    \* another commit with the same subject hash is on one
    tried,       \* clone B ran a try of a missing cell
    refusedB,    \* a run in clone B was refused
    out          \* what the last command printed

vars == <<gitPlace, gitCells, gitSelect, landed, trH0, trCand, runs, reach,
          sameSubj, tried, refusedB, out>>

IsOpen     == gitPlace /\ gitSelect = "none"
IsSelected == gitSelect # "none"

Cell(v) ==
    IF v = "cell" THEN "met"
    ELSE IF v = "rerun" THEN (IF RerunUnmarked THEN "met" ELSE "met, run again")
    ELSE IF MissingAsNotRun THEN "not run"
    ELSE IF MissingNoTryLine THEN "missing"
    ELSE "missing, try line"

Runs(c) == IF BudgetFromGit THEN runs["A"] + runs["B"] ELSE runs[c]

\* The state hyp prints after the step, read from the primed variables.
GoalState(l, s, p) ==
    IF l THEN "landed" ELSE IF s # "none" THEN "selected"
    ELSE IF p THEN "open" ELSE "written"

\* What a command prints: its clone, the Goal's state, the runs (-1 when the
\* block prints none), the warning, and the record's cell with the trace
\* value it was read from ("" when no record prints).
Print(cmd, c, l, s, p, rch, same, h0, cand, withRecord) ==
    LET st == GoalState(l, s, p)
        sel == st \in {"selected", "landed"}
        tv == IF sel THEN cand[c] ELSE h0[c]
    IN [cmd |-> cmd, clone |-> c, state |-> st,
        runs |-> IF st = "open" THEN Runs(c) ELSE -1,
        warn |-> sel /\ ~rch /\ ~(SameSubjectQuiet /\ same)
                 /\ ~(WarnOnlyStatusG /\ cmd # "status G"),
        cell |-> IF withRecord /\ st # "written" THEN Cell(tv) ELSE "",
        trace |-> IF withRecord /\ st # "written" THEN tv ELSE ""]

Quiet(cmd) == [cmd |-> cmd, clone |-> "A", state |-> "", runs |-> -1,
               warn |-> FALSE, cell |-> "", trace |-> ""]

Init ==
    /\ gitPlace = FALSE /\ gitCells = FALSE /\ gitSelect = "none"
    /\ landed = FALSE
    /\ trH0 = [c \in Clones |-> "none"]
    /\ trCand = [c \in Clones |-> "none"]
    /\ runs = [c \in Clones |-> 0]
    /\ reach = TRUE /\ sameSubj = FALSE
    /\ tried = FALSE /\ refusedB = FALSE
    /\ out = Quiet("none")

\* The grill runs `hyp place G` in clone A; the record PR merges.
GrillPlaces ==
    /\ ~gitPlace
    /\ gitPlace' = TRUE
    /\ gitCells' = (gitCells \/ PlaceCellsInGit)
    /\ trH0' = [trH0 EXCEPT !["A"] = "cell"]
    /\ out' = Quiet("place")
    /\ UNCHANGED <<gitSelect, landed, trCand, runs, reach, sameSubj, tried, refusedB>>

\* The runner runs a line in a clone c.
RunnerRuns(c) ==
    /\ IsOpen /\ runs[c] = 0 /\ ~refusedB
    /\ IF c = "B" /\ RunNeedsH0Cell /\ trH0["B"] = "none"
       THEN /\ refusedB' = TRUE
            /\ out' = Quiet("run refused")
            /\ UNCHANGED <<gitCells, trCand, runs>>
       ELSE /\ trCand' = [trCand EXCEPT ![c] = "cell"]
            /\ runs' = [runs EXCEPT ![c] = 1]
            /\ gitCells' = (gitCells \/ RunInGit)
            /\ out' = Quiet("run")
            /\ UNCHANGED refusedB
    /\ UNCHANGED <<gitPlace, gitSelect, landed, trH0, reach, sameSubj, tried>>

\* The runner selects the candidate in clone A; the PR merges.
RunnerSelects ==
    /\ IsOpen /\ trCand["A"] = "cell"
    /\ gitSelect' = IF SelectNoHashes THEN "nohashes" ELSE "hashes"
    /\ out' = Print("select", "A", landed, gitSelect', gitPlace, reach, sameSubj,
                    trH0, trCand, TRUE)
    /\ UNCHANGED <<gitPlace, gitCells, landed, trH0, trCand, runs, reach,
                   sameSubj, tried, refusedB>>

\* The owner lands the selected Goal in a clone c.
OwnerLands(c) ==
    /\ IsSelected /\ ~landed
    /\ landed' = TRUE
    /\ out' = Print("land", c, TRUE, gitSelect, gitPlace, reach, sameSubj,
                    trH0, trCand, TRUE)
    /\ UNCHANGED <<gitPlace, gitCells, gitSelect, trH0, trCand, runs, reach,
                   sameSubj, tried, refusedB>>

\* A squash merge or a deleted branch leaves the selected commit on no ref.
BranchGoes ==
    /\ IsSelected /\ reach
    /\ reach' = FALSE
    /\ \E s \in BOOLEAN : sameSubj' = s
    /\ out' = Quiet("branch goes")
    /\ UNCHANGED <<gitPlace, gitCells, gitSelect, landed, trH0, trCand, runs,
                   tried, refusedB>>

\* Anyone runs `hyp status` or `hyp status G` in a clone c.
Status(c) ==
    /\ \E cmd \in {"status", "status G"} :
        out' = Print(cmd, c, landed, gitSelect, gitPlace, reach, sameSubj,
                     trH0, trCand, cmd = "status G")
    /\ UNCHANGED <<gitPlace, gitCells, gitSelect, landed, trH0, trCand, runs,
                   reach, sameSubj, tried, refusedB>>

\* A reader in clone B runs the try line of the cell its trace lacks.
ReaderTries ==
    /\ ~tried
    /\ \/ IsOpen /\ trH0["B"] = "none"
       \/ IsSelected /\ trCand["B"] = "none" /\ reach
    /\ tried' = TRUE
    /\ IF TryNoSave THEN UNCHANGED <<trH0, trCand>>
       ELSE IF IsSelected
            THEN /\ trCand' = [trCand EXCEPT !["B"] = "rerun"] /\ UNCHANGED trH0
            ELSE /\ trH0' = [trH0 EXCEPT !["B"] = "rerun"] /\ UNCHANGED trCand
    /\ out' = Quiet("try")
    /\ UNCHANGED <<gitPlace, gitCells, gitSelect, landed, runs, reach, sameSubj,
                   refusedB>>

Next ==
    \/ GrillPlaces
    \/ \E c \in Clones : RunnerRuns(c)
    \/ RunnerSelects
    \/ \E c \in Clones : OwnerLands(c)
    \/ BranchGoes
    \/ \E c \in Clones : Status(c)
    \/ ReaderTries

Spec == Init /\ [][Next]_vars

----------------------------------------------------------------------------
\* source: model only
TypeOK ==
    /\ gitPlace \in BOOLEAN /\ gitCells \in BOOLEAN
    /\ gitSelect \in {"none", "hashes", "nohashes"}
    /\ trH0 \in [Clones -> {"none", "cell", "rerun"}]
    /\ trCand \in [Clones -> {"none", "cell", "rerun"}]
    /\ runs \in [Clones -> {0, 1}]

\* source: line-1
NoCellInGit == ~gitCells

\* source: line-2
SelectNamesHashes == gitSelect # "nohashes"

\* source: line-3
MissingHasTryLine == out.trace = "none" => out.cell = "missing, try line"

\* source: line-4
TrySaves == tried => (trH0["B"] = "rerun" \/ trCand["B"] = "rerun")

\* source: line-4
RerunMarked == out.trace = "rerun" => out.cell = "met, run again"

\* source: line-5
WarnsUnreachable ==
    (out.state \in {"selected", "landed"} /\ ~reach) => out.warn

\* source: line-6
BudgetPerClone == (out.runs # -1 /\ runs[out.clone] = 0) => out.runs = 0

\* source: line-8
FreshCloneRuns == ~refusedB
=============================================================================
