------------------------------- MODULE change -------------------------------
\* The model of G-XVME1C: what hyp status prints of the kept model at the
\* Selected stop, what the Spec guard compares the kept model with, and where
\* the statechart tool writes the variant configs. Each case is one example's
\* setup: the selected commit adds a rule, changes a rule or adds a step of
\* the kept model; a candidate takes in a side branch that adds a rule while
\* the clone has that branch checked out; or the tool runs on the kept model
\* or on a Goal's own model.

EXTENDS Sequences

CONSTANTS SourceNotWords, StepByName, PrintsText, ListsAll, MergeBase,
          BesideModel, OutOnly

VARIABLES case, st, out

vars == <<case, st, out>>

StatusCases == {"new rule", "changed rule", "new step", "side branch"}
ToolCases == {"kept tool", "own tool"}

Init == case \in StatusCases \cup ToolCases /\ st = "none"
        /\ out = [cmd |-> "none", items |-> {}]

\* The side branch adds NewRule after H0 and no variant config names it.
\* Measured against H0 it is new, so the Spec guard holds the select back;
\* measured against the merge base of HEAD and the candidate it is not.
Held == case = "side branch" /\ ~MergeBase

\* command: hyp run
Run ==
    /\ case \in StatusCases /\ st = "none" /\ st' = "ran"
    /\ out' = [cmd |-> "run", items |-> {"ran"}]
    /\ UNCHANGED case

\* command: hyp select
Select ==
    /\ st = "ran" /\ ~Held /\ st' = "selected"
    /\ out' = [cmd |-> "select", items |-> {"selected"}]
    /\ UNCHANGED case

RuleLine(name, src, words) ==
    "rule " \o name \o ": " \o (IF SourceNotWords THEN src ELSE words)

Listed ==
    (CASE case = "new rule" -> {RuleLine("NewRule", "G-ONE:line-1", "G-ONE:line-1: the subject says bye.")}
       [] case = "changed rule" -> {RuleLine("OldRule", "model only", "no line")}
       [] case = "new step" -> {IF StepByName THEN "step: Fold" ELSE "step: toy fold",
                                RuleLine("NewRule", "G-ONE:line-1", "G-ONE:line-1: the subject says bye.")}
       [] OTHER -> {"rule NewRule: model only"})
    \cup (IF PrintsText THEN {"TLA text"} ELSE {})
    \cup (IF ListsAll THEN {"rule KeepsCount"} ELSE {})

\* command: hyp status
Status ==
    /\ st # "none"
    /\ out' = [cmd |-> "status",
               items |-> IF st = "selected" THEN Listed
                         ELSE IF Held THEN {"held back: rule-no-variant NewRule"}
                         ELSE {"hyp select"}]
    /\ UNCHANGED <<case, st>>

\* command: statechart
Tool ==
    /\ case \in ToolCases /\ st = "none" /\ st' = "done"
    /\ out' = [cmd |-> "tool",
               items |-> IF case = "kept tool"
                         THEN IF BesideModel THEN {"goals/model/variants/Breaks.cfg"}
                              ELSE {"goals/G-ONE/specs/variants/Breaks.cfg"}
                         ELSE IF OutOnly THEN {"no Goal part"}
                              ELSE {"| SaysBye | line-1 |", "no line unnamed"}]
    /\ UNCHANGED case

Next == Run \/ Select \/ Status \/ Tool

Spec == Init /\ [][Next]_vars

AtSelected == out.cmd = "status" /\ st = "selected"

\* source: line-1
ShowsLineWords ==
    /\ (AtSelected /\ case \in {"new rule", "new step"})
         => "rule NewRule: G-ONE:line-1: the subject says bye." \in out.items
    /\ (AtSelected /\ case = "changed rule") => "rule OldRule: no line" \in out.items

\* source: line-1
StepByCommand == (AtSelected /\ case = "new step") => "step: toy fold" \in out.items

\* source: line-1
NoTLAText == "TLA text" \notin out.items

\* source: line-1
OnlyChanged == "rule KeepsCount" \notin out.items

\* source: line-2
ComparesWithH0 ==
    (out.cmd = "status" /\ case = "side branch" /\ st = "ran")
        => out.items = {"held back: rule-no-variant NewRule"}

\* source: line-3
VariantsInGoalFolder ==
    (out.cmd = "tool" /\ case = "kept tool")
        => out.items = {"goals/G-ONE/specs/variants/Breaks.cfg"}

\* source: line-4
OwnSpecAsBefore ==
    (out.cmd = "tool" /\ case = "own tool")
        => out.items = {"| SaysBye | line-1 |", "no line unnamed"}
=============================================================================
