----------------------------- MODULE state_rule -----------------------------
(***************************************************************************)
(* Goal G-QJABD8: "Git keeps only the events that change a Goal's state,   *)
(* and hyp stores each event kind from that one rule, so every clone       *)
(* replays the same Goal states and refusals from git. Every Result and    *)
(* output stays in the clone's trace as a cache that a reader fills by     *)
(* running an example again, and hyp warns when a selected commit is on no *)
(* branch or tag."                                                         *)
(*                                                                         *)
(* Two clones share one git history: clone A, where the grill places and   *)
(* the runner runs and selects, and clone B, a fresh clone. Git here is    *)
(* the merged default branch, so B sees every line A commits. Each clone   *)
(* has its own trace. The Goal has one example, line-1/a.                  *)
(*                                                                         *)
(* Variants, each FALSE in state_rule.cfg, are cheaper builds:             *)
(*   NoStateMark        hyp spec prints no "changes a Goal's state" field  *)
(*   ResultInGit        run still writes its Result and outputs to git     *)
(*   NoReplayMark       hyp spec prints no "checked at replay" field       *)
(*   ReplayReadsTrace   replay checks Candidate against the clone's cells, *)
(*                      so a clone with no cells refuses the select        *)
(*   PlaceOutputsInGit  place still writes H0's outputs to git             *)
(*   PlaceDropsVerdict  place keeps no H0 verdict in git, so a clone whose *)
(*                      trace lacks H0's cell cannot tell the Goal is open *)
(*   BudgetFromGit      every clone counts the runs of every clone         *)
(*   MissingNoTryLine   a missing cell says so, with no try line           *)
(*   MissingAsNotRun    a missing cell prints as `not run`                 *)
(*   TryNoSave          try saves nothing in the trace                     *)
(*   RerunUnmarked      a rerun cell prints as the cell the owner saw      *)
(*   SameSubjectQuiet   no warning when a commit with the same subject     *)
(*                      hash is on a branch or tag                         *)
(*   WarnOnlyStatusG    the warning prints only in `status G`              *)
(*   RunNeedsH0Cell     run refuses in a clone whose trace lacks H0's cell *)
(* Left out: line-10 (landed Goals print as on the base) is about the old  *)
(* lines of the log, not a step.                                           *)
(***************************************************************************)
EXTENDS Integers

CONSTANTS
    NoStateMark, ResultInGit, NoReplayMark, ReplayReadsTrace,
    PlaceOutputsInGit, PlaceDropsVerdict, BudgetFromGit, MissingNoTryLine, MissingAsNotRun,
    TryNoSave, RerunUnmarked, SameSubjectQuiet, WarnOnlyStatusG,
    RunNeedsH0Cell

Clones == {"A", "B"}

VARIABLES
    gitPlace,    \* git holds the place line of the set
    gitPlaceOut, \* git holds an output file of H0's cell
    gitResult,   \* git holds a Result or its output
    gitSelect,   \* git holds the select
    landed,      \* git holds the owner's land
    trH0,        \* per clone: H0's cell in the trace: "none", "cell", "rerun"
    trCand,      \* per clone: the candidate's cell: "none", "cell", "rerun"
    runs,        \* per clone: Hypotheses run in that clone's trace (0 or 1)
    reach,       \* the selected commit is on a branch or tag
    sameSubj,    \* another commit with the same subject hash is on one
    tried,       \* clone B ran a try of a missing cell
    refusedB,    \* a run in clone B was refused
    out          \* what the last command printed

vars == <<gitPlace, gitPlaceOut, gitResult, gitSelect, landed, trH0, trCand,
          runs, reach, sameSubj, tried, refusedB, out>>

IsOpen     == gitPlace /\ ~gitSelect
IsSelected == gitSelect

Cell(v) ==
    IF v = "cell" THEN "met"
    ELSE IF v = "rerun" THEN (IF RerunUnmarked THEN "met" ELSE "met, run again")
    ELSE IF MissingAsNotRun THEN "not run"
    ELSE IF MissingNoTryLine THEN "missing"
    ELSE "missing, try line"

Runs(c) == IF BudgetFromGit THEN runs["A"] + runs["B"] ELSE runs[c]

\* The state replay gives from git alone.
GitState(l, s, p) ==
    IF l THEN "landed" ELSE IF s THEN "selected"
    ELSE IF p THEN "open" ELSE "written"

\* Replay in clone c refuses the select when it checks Candidate against the
\* clone's cells and the clone holds none of the candidate.
RefusesSelect(c, s, cand) == ReplayReadsTrace /\ s /\ cand[c] = "none"

\* What a command prints: its clone, the Goal's state, the runs (-1 when the
\* block prints none), the warning, the refused-at-replay mark, and the
\* record's cell with the trace value it was read from.
Print(cmd, c, l, s, p, rch, same, h0, cand, withRecord) ==
    LET refused == RefusesSelect(c, s, cand)
        lost == PlaceDropsVerdict /\ p /\ ~s /\ ~l /\ h0[c] = "none"
        st == IF refused /\ ~l THEN "open"
              ELSE IF lost THEN "written" ELSE GitState(l, s, p)
        sel == st \in {"selected", "landed"}
        tv == IF sel THEN cand[c] ELSE h0[c]
    IN [cmd |-> cmd, clone |-> c, state |-> st, git |-> GitState(l, s, p),
        refused |-> refused, lost |-> lost,
        runs |-> Runs(c),
        warn |-> sel /\ ~rch /\ ~(SameSubjectQuiet /\ same)
                 /\ ~(WarnOnlyStatusG /\ cmd # "status G"),
        cell |-> IF withRecord /\ st # "written" THEN Cell(tv) ELSE "",
        trace |-> IF withRecord /\ st # "written" THEN tv ELSE "",
        stateMark |-> FALSE, replayMark |-> FALSE, resultHome |-> ""]

Quiet(cmd) == [cmd |-> cmd, clone |-> "A", state |-> "", git |-> "",
               refused |-> FALSE, lost |-> FALSE, runs |-> -1, warn |-> FALSE, cell |-> "",
               trace |-> "", stateMark |-> FALSE, replayMark |-> FALSE,
               resultHome |-> ""]

Init ==
    /\ gitPlace = FALSE /\ gitPlaceOut = FALSE /\ gitResult = FALSE
    /\ gitSelect = FALSE /\ landed = FALSE
    /\ trH0 = [c \in Clones |-> "none"]
    /\ trCand = [c \in Clones |-> "none"]
    /\ runs = [c \in Clones |-> 0]
    /\ reach = TRUE /\ sameSubj = FALSE
    /\ tried = FALSE /\ refusedB = FALSE
    /\ out = Quiet("none")

\* Anyone runs `hyp spec`.
ReadsSpec ==
    /\ out' = [Quiet("spec") EXCEPT !.stateMark = ~NoStateMark,
                                    !.replayMark = ~NoReplayMark,
                                    !.resultHome = IF ResultInGit THEN "git" ELSE "trace"]
    /\ UNCHANGED <<gitPlace, gitPlaceOut, gitResult, gitSelect, landed, trH0,
                   trCand, runs, reach, sameSubj, tried, refusedB>>

\* The grill runs `hyp place G` in clone A; the record PR merges.
GrillPlaces ==
    /\ ~gitPlace
    /\ gitPlace' = TRUE
    /\ gitPlaceOut' = PlaceOutputsInGit
    /\ trH0' = [trH0 EXCEPT !["A"] = "cell"]
    /\ out' = Quiet("place")
    /\ UNCHANGED <<gitResult, gitSelect, landed, trCand, runs, reach, sameSubj,
                   tried, refusedB>>

\* The runner runs a line in a clone c.
RunnerRuns(c) ==
    /\ IsOpen /\ runs[c] = 0 /\ ~refusedB
    /\ IF c = "B" /\ RunNeedsH0Cell /\ trH0["B"] = "none"
       THEN /\ refusedB' = TRUE
            /\ out' = Quiet("run refused")
            /\ UNCHANGED <<gitResult, trCand, runs>>
       ELSE /\ trCand' = [trCand EXCEPT ![c] = "cell"]
            /\ runs' = [runs EXCEPT ![c] = 1]
            /\ gitResult' = (gitResult \/ ResultInGit)
            /\ out' = Quiet("run")
            /\ UNCHANGED refusedB
    /\ UNCHANGED <<gitPlace, gitPlaceOut, gitSelect, landed, trH0, reach,
                   sameSubj, tried>>

\* The runner selects the candidate in clone A; the PR merges.
RunnerSelects ==
    /\ IsOpen /\ trCand["A"] = "cell"
    /\ gitSelect' = TRUE
    /\ out' = Print("select", "A", landed, TRUE, gitPlace, reach, sameSubj,
                    trH0, trCand, TRUE)
    /\ UNCHANGED <<gitPlace, gitPlaceOut, gitResult, landed, trH0, trCand, runs,
                   reach, sameSubj, tried, refusedB>>

\* The owner lands the selected Goal in clone A.
OwnerLands ==
    /\ IsSelected /\ ~landed
    /\ landed' = TRUE
    /\ out' = Print("land", "A", TRUE, gitSelect, gitPlace, reach, sameSubj,
                    trH0, trCand, TRUE)
    /\ UNCHANGED <<gitPlace, gitPlaceOut, gitResult, gitSelect, trH0, trCand,
                   runs, reach, sameSubj, tried, refusedB>>

\* A squash merge or a deleted branch leaves the selected commit on no ref.
BranchGoes ==
    /\ IsSelected /\ reach
    /\ reach' = FALSE
    /\ \E s \in BOOLEAN : sameSubj' = s
    /\ out' = Quiet("branch goes")
    /\ UNCHANGED <<gitPlace, gitPlaceOut, gitResult, gitSelect, landed, trH0,
                   trCand, runs, tried, refusedB>>

\* Anyone runs `hyp status` or `hyp status G` in a clone c.
Status(c) ==
    /\ \E cmd \in {"status", "status G"} :
        out' = Print(cmd, c, landed, gitSelect, gitPlace, reach, sameSubj,
                     trH0, trCand, cmd = "status G")
    /\ UNCHANGED <<gitPlace, gitPlaceOut, gitResult, gitSelect, landed, trH0,
                   trCand, runs, reach, sameSubj, tried, refusedB>>

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
    /\ UNCHANGED <<gitPlace, gitPlaceOut, gitResult, gitSelect, landed, runs,
                   reach, sameSubj, refusedB>>

Next ==
    \/ ReadsSpec
    \/ GrillPlaces
    \/ \E c \in Clones : RunnerRuns(c)
    \/ RunnerSelects
    \/ OwnerLands
    \/ BranchGoes
    \/ \E c \in Clones : Status(c)
    \/ ReaderTries

Spec == Init /\ [][Next]_vars

----------------------------------------------------------------------------
\* source: model only
TypeOK ==
    /\ gitPlace \in BOOLEAN /\ gitPlaceOut \in BOOLEAN /\ gitResult \in BOOLEAN
    /\ gitSelect \in BOOLEAN /\ landed \in BOOLEAN
    /\ trH0 \in [Clones -> {"none", "cell", "rerun"}]
    /\ trCand \in [Clones -> {"none", "cell", "rerun"}]
    /\ runs \in [Clones -> {0, 1}]

\* source: line-1
SpecMarksStateChange == out.cmd = "spec" => out.stateMark

\* source: line-1
ResultInTrace == out.cmd = "spec" => out.resultHome = "trace"

\* source: line-2
SpecMarksReplay == out.cmd = "spec" => out.replayMark

\* source: line-2
ReplayReadsOnlyGit == ~out.refused

\* source: line-3
NoPlaceOutputInGit == ~gitPlaceOut

\* source: line-3
PlaceNamesVerdict == ~out.lost

\* source: line-4
RunWritesNothingInGit == ~gitResult

\* source: line-4
BudgetPerClone == (out.state # "" /\ runs[out.clone] = 0) => out.runs = 0

\* source: line-5
MissingHasTryLine == out.trace = "none" => out.cell = "missing, try line"

\* source: line-6
TrySaves == tried => (trH0["B"] = "rerun" \/ trCand["B"] = "rerun")

\* source: line-6
RerunMarked == out.trace = "rerun" => out.cell = "met, run again"

\* source: line-7
WarnsUnreachable ==
    (out.state \in {"selected", "landed"} /\ ~reach) => out.warn

\* source: line-8
SameStateEveryClone == out.state # "" => out.state = out.git

\* source: line-9
FreshCloneRuns == ~refusedB
=============================================================================
