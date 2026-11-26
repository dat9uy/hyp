------------------------------ MODULE predict ------------------------------
(* The statechart tool predicts the cell of one fixture example under the  *)
(* fixture's variant NoFoldLine. The fixture is set up before the tool     *)
(* runs, so the case is chosen at the start: the case its Probe and run    *)
(* section set up. Each variant is a cheap build of the tool.              *)
CONSTANTS
    ShortestWalk,   \* the old tool: the shortest walk, every print matched
    AllPrints,      \* the Probe's walk, but setup prints matched too
    OutsideAsFail,  \* a floor no print gives is predicted fail
    SkipUnknown,    \* a command with no step is skipped
    ProbeOnly       \* the walk takes the Probe's commands, not the run section's

VARIABLES case, ran, shown

vars == <<case, ran, shown>>

Cases == {"setup", "setupPrint", "probeFloor", "noStep", "plain"}

MetCell == "| line-1/a | met |"
FailCell == "| line-1/a | fail |"
NoWalkCell == "| line-1/a | no walk |"
Outside == "line-1/a is outside the model's prints\n| line-1/a | not predicted |"
NoStep == "the model has no step for toy refute\n| line-1/a | not predicted |"

\* What the tool prints for the fixture example in each case.
Cell(c) ==
    IF ShortestWalk THEN
        CASE c = "setup" -> MetCell
          [] c = "setupPrint" -> NoWalkCell
          [] c = "probeFloor" -> NoWalkCell
          [] c = "noStep" -> MetCell
          [] c = "plain" -> MetCell
    ELSE IF ProbeOnly /\ c # "noStep" THEN Outside
    ELSE
        CASE c = "setup" -> FailCell
          [] c = "setupPrint" -> IF AllPrints THEN FailCell ELSE MetCell
          [] c = "probeFloor" -> IF OutsideAsFail THEN FailCell ELSE Outside
          [] c = "noStep" -> IF SkipUnknown THEN MetCell ELSE NoStep
          [] c = "plain" -> MetCell

Init == case \in Cases /\ ran = FALSE /\ shown = ""

\* command: skills/hyp-statechart/scripts/statechart
RunTool == ~ran /\ ran' = TRUE /\ shown' = Cell(case) /\ UNCHANGED case

Next == RunTool

Spec == Init /\ [][Next]_vars

TypeOK == case \in Cases /\ ran \in BOOLEAN

\* source: line-1
SetupCaseFails == (ran /\ case = "setup") => shown = FailCell

\* source: line-1
SetupPrintsUnmatched == (ran /\ case = "setupPrint") => shown = MetCell

\* source: line-2
OutsideNamed == (ran /\ case = "probeFloor") => shown = Outside

\* source: line-3
NoStepNamed == (ran /\ case = "noStep") => shown = NoStep

\* source: line-4
PlainAsOnH0 == (ran /\ case = "plain") => shown = MetCell
=============================================================================
