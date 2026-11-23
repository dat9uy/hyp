------------------------------ MODULE findings ------------------------------
(***************************************************************************)
(* Goal G-SZ39PG: "An agent records a finding on the Goal where it found   *)
(* it, and any agent reads a Goal's findings with their state, folds them  *)
(* into an issue or refutes one, all in the clone's trace and never in     *)
(* git."                                                                   *)
(*                                                                         *)
(* One Goal G moves open -> selected -> ended. Findings Ids hang on G.     *)
(* A finding is "none" (not added), "notfolded", "folded" (into issue[i]) *)
(* or "refuted". Agents add, fold, refute and read; the owner lands.       *)
(*                                                                         *)
(* Variants, each FALSE in findings.cfg, are cheaper builds:              *)
(*   AddRefusedEnded   add refuses a Goal that has ended                  *)
(*   AddMissingGoal    add accepts a Goal that does not exist             *)
(*   StatusOnlyOpen    status G prints only findings not folded           *)
(*   StatusNoState     status G prints findings with no state             *)
(*   FoldNoIssue       a fold that names no issue is recorded             *)
(*   NoRefold          a folded finding cannot be folded again            *)
(*   RefutedReopens    a fold or refute of a refuted finding is recorded  *)
(*   SpecNoStates      hyp spec lists the kinds but no finding states     *)
(*   FindingInGit      a finding act writes its event under .tree/        *)
(*   LandWaitsFold     land refuses while a finding is not folded         *)
(***************************************************************************)
EXTENDS Integers, FiniteSets

CONSTANTS Ids, Issues,
    AddRefusedEnded, AddMissingGoal, StatusOnlyOpen, StatusNoState,
    FoldNoIssue, NoRefold, RefutedReopens, SpecNoStates, FindingInGit,
    LandWaitsFold

VARIABLES
    goal,      \* "open", "selected" or "ended"
    f,         \* per finding: "none", "notfolded", "folded", "refuted"
    issue,     \* per finding: the issue it is folded into, 0 for none
    gitTree,   \* a finding act changed a file under .tree/
    reopened,  \* an act changed a refuted finding
    out        \* what the last command printed

vars == <<goal, f, issue, gitTree, reopened, out>>

NoShow == [i \in Ids |-> "absent"]
Out(cmd, ok) == [cmd |-> cmd, ok |-> ok, why |-> "none", shown |-> NoShow, states |-> FALSE]
OutWhy(cmd, ok, why) == [cmd |-> cmd, ok |-> ok, why |-> why, shown |-> NoShow, states |-> FALSE]

TypeOK ==
    /\ goal \in {"open", "selected", "ended"}
    /\ f \in [Ids -> {"none", "notfolded", "folded", "refuted"}]
    /\ issue \in [Ids -> {0} \cup Issues]
    /\ gitTree \in BOOLEAN /\ reopened \in BOOLEAN

Init ==
    /\ goal = "open"
    /\ f = [i \in Ids |-> "none"]
    /\ issue = [i \in Ids |-> 0]
    /\ gitTree = FALSE /\ reopened = FALSE
    /\ out = Out("none", TRUE)

\* What status G shows for finding i: its state, or nothing.
Shown(i) ==
    IF f[i] = "none" THEN "absent"
    ELSE IF StatusOnlyOpen /\ f[i] # "notfolded" THEN "absent"
    ELSE IF StatusNoState THEN "no state"
    ELSE f[i]

\* An agent adds finding i on G, in any state of G.
AgentAdds(i) ==
    /\ f[i] = "none"
    /\ IF AddRefusedEnded /\ goal = "ended"
         THEN /\ out' = Out("add", FALSE)
              /\ UNCHANGED <<f, gitTree>>
         ELSE /\ f' = [f EXCEPT ![i] = "notfolded"]
              /\ out' = Out("add", TRUE)
              /\ gitTree' = (gitTree \/ FindingInGit)
    /\ UNCHANGED <<goal, issue, reopened>>

\* An agent adds a finding on a Goal id that does not exist.
AgentAddsNoGoal ==
    /\ out' = Out("add-no-goal", AddMissingGoal)
    /\ gitTree' = (gitTree \/ (AddMissingGoal /\ FindingInGit))
    /\ UNCHANGED <<goal, f, issue, reopened>>

Foldable(i) ==
    \/ f[i] = "notfolded"
    \/ f[i] = "folded" /\ ~NoRefold
    \/ f[i] = "refuted" /\ RefutedReopens

\* An agent folds the findings S into issue n.
AgentFolds(S, n) ==
    /\ S # {} /\ \A i \in S : f[i] # "none"
    /\ IF \A i \in S : Foldable(i)
         THEN /\ f' = [i \in Ids |-> IF i \in S THEN "folded" ELSE f[i]]
              /\ issue' = [i \in Ids |-> IF i \in S THEN n ELSE issue[i]]
              /\ reopened' = (reopened \/ \E i \in S : f[i] = "refuted")
              /\ gitTree' = (gitTree \/ FindingInGit)
              /\ out' = Out("fold", TRUE)
         ELSE /\ out' = OutWhy("fold", FALSE,
                     IF \E i \in S : f[i] = "refuted" THEN "refuted" ELSE "folded")
              /\ UNCHANGED <<f, issue, reopened, gitTree>>
    /\ UNCHANGED goal

\* An agent folds finding i and names no issue.
AgentFoldsNoIssue(i) ==
    /\ f[i] \in {"notfolded", "folded"}
    /\ out' = Out("fold-no-issue", FoldNoIssue)
    /\ f' = IF FoldNoIssue THEN [f EXCEPT ![i] = "folded"] ELSE f
    /\ issue' = IF FoldNoIssue THEN [issue EXCEPT ![i] = 0] ELSE issue
    /\ UNCHANGED <<goal, gitTree, reopened>>

\* An agent refutes finding i.
AgentRefutes(i) ==
    /\ f[i] # "none"
    /\ IF f[i] # "refuted" \/ RefutedReopens
         THEN /\ reopened' = (reopened \/ f[i] = "refuted")
              /\ f' = [f EXCEPT ![i] = "refuted"]
              /\ gitTree' = (gitTree \/ FindingInGit)
              /\ out' = Out("refute", TRUE)
         ELSE /\ out' = Out("refute", FALSE)
              /\ UNCHANGED <<f, reopened, gitTree>>
    /\ UNCHANGED <<goal, issue>>

\* An agent reads status G or hyp spec.
AgentReadsStatus ==
    /\ out' = [cmd |-> "status", ok |-> TRUE, why |-> "none",
               shown |-> [i \in Ids |-> Shown(i)], states |-> FALSE]
    /\ UNCHANGED <<goal, f, issue, gitTree, reopened>>
AgentReadsSpec ==
    /\ out' = [cmd |-> "spec", ok |-> TRUE, why |-> "none",
               shown |-> NoShow, states |-> ~SpecNoStates]
    /\ UNCHANGED <<goal, f, issue, gitTree, reopened>>

RunnerSelects ==
    /\ goal = "open" /\ goal' = "selected"
    /\ out' = Out("select", TRUE)
    /\ UNCHANGED <<f, issue, gitTree, reopened>>

OwnerLands ==
    /\ goal = "selected"
    /\ IF LandWaitsFold /\ \E i \in Ids : f[i] = "notfolded"
         THEN out' = Out("land", FALSE) /\ UNCHANGED goal
         ELSE out' = Out("land", TRUE) /\ goal' = "ended"
    /\ UNCHANGED <<f, issue, gitTree, reopened>>

Next ==
    \/ \E i \in Ids : AgentAdds(i)
    \/ AgentAddsNoGoal
    \/ \E S \in SUBSET Ids, n \in Issues : AgentFolds(S, n)
    \/ \E i \in Ids : AgentFoldsNoIssue(i)
    \/ \E i \in Ids : AgentRefutes(i)
    \/ AgentReadsStatus
    \/ AgentReadsSpec
    \/ RunnerSelects
    \/ OwnerLands

Spec == Init /\ [][Next]_vars

\* source: model only
TypeOKRule == TypeOK

\* source: line-1
AddsInAnyState == ~(out.cmd = "add" /\ ~out.ok)

\* source: line-1
RefusesMissingGoal == ~(out.cmd = "add-no-goal" /\ out.ok)

\* source: line-2
StatusShowsEveryState ==
    out.cmd = "status" => \A i \in Ids : f[i] # "none" => out.shown[i] = f[i]

\* source: line-3
FoldNeedsIssue == ~(out.cmd = "fold-no-issue" /\ out.ok)

\* source: line-3
FoldedFoldsAgain == ~(out.cmd = "fold" /\ ~out.ok /\ out.why = "folded")

\* source: line-4
RefutedIsFinal == ~reopened

\* source: line-5
SpecListsStates == out.cmd = "spec" => out.states

\* source: line-6
NothingInGit == ~gitTree

\* source: line-7
LandNeverWaits == ~(out.cmd = "land" /\ ~out.ok)
=============================================================================
