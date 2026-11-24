------------------------------ MODULE learned ------------------------------
(***************************************************************************)
(* Goal G-T30ZAY: "What was learned reaches later work through the Goal:  *)
(* a decided want is a line of a written Goal, a fact is a finding that   *)
(* any agent folds into the one Goal in work whose want it bears on, and  *)
(* hyp's reads show both without a guess at which issue or comment holds  *)
(* them."                                                                 *)
(*                                                                         *)
(* Goals A and B move written -> open -> selected, and end from any of    *)
(* those. A finding is "none" (not added), "notfolded", "folded" (into    *)
(* into[i]) or "refuted", and may carry the agent's drafted add goal.      *)
(* Whether a finding's Goal has ended is calculated, never written.       *)
(*                                                                         *)
(* Variants, each FALSE in learned.cfg, are cheaper builds:               *)
(*   AddNeedsGoal       finding add still needs a Goal (H0)               *)
(*   FoldIntoEnded      a fold into an ended Goal is recorded             *)
(*   FoldIntoMissing    a fold into a Goal that does not exist is recorded *)
(*   RefutedReopens     a fold of a refuted finding is recorded           *)
(*   NoRefold           a folded finding takes no second fold             *)
(*   EndFreezes         a finding whose Goal ended takes no fold again    *)
(*   NoEndedMark        a finding of an ended Goal reads as plain folded  *)
(*   ListShowsEnded     finding list shows findings of ended Goals        *)
(*   NotFoldedShowsAll  --not-folded lists folded findings too            *)
(*   ListNoActs         finding list prints no ready lines                *)
(*   DraftOnFolded      a draft attaches to a folded finding too          *)
(*   ListNoDraft        finding list does not print the draft             *)
(*   StatusNoWords      hyp status prints no words of a Goal (H0)         *)
(*   WordsOfSelected    hyp status prints the words of a selected Goal    *)
(*   StatusNoCount      hyp status prints no count of free findings       *)
(*   StatusGNoWhy       hyp status G does not print the why (H0)          *)
(*   StatusGNoFolded    hyp status G does not list folded findings (H0)   *)
(***************************************************************************)
EXTENDS Integers, FiniteSets

CONSTANTS Ids, Goals,
    AddNeedsGoal, FoldIntoEnded, FoldIntoMissing, RefutedReopens, NoRefold,
    EndFreezes, NoEndedMark, ListShowsEnded, NotFoldedShowsAll, ListNoActs,
    DraftOnFolded, ListNoDraft, StatusNoWords, WordsOfSelected, StatusNoCount,
    StatusGNoWhy, StatusGNoFolded

VARIABLES
    gs,        \* per Goal: "written", "open", "selected" or "ended"
    f,         \* per finding: "none", "notfolded", "folded" or "refuted"
    into,      \* per finding: the Goal it is folded into, "none" for none
    draft,     \* per finding: it carries the agent's drafted add goal
    reopened,  \* an act changed a refuted finding
    out        \* what the last command printed

vars == <<gs, f, into, draft, reopened, out>>

Missing == "missing"
NoShow == [i \in Ids |-> "absent"]
NoWords == [g \in Goals |-> FALSE]
Blank == [cmd |-> "none", ok |-> TRUE, why |-> "none", shown |-> NoShow,
          acts |-> FALSE, drafts |-> FALSE, words |-> NoWords, count |-> -1,
          goalWhy |-> FALSE]
Out(cmd, ok, why) == [Blank EXCEPT !.cmd = cmd, !.ok = ok, !.why = why]

TypeOK ==
    /\ gs \in [Goals -> {"written", "open", "selected", "ended"}]
    /\ f \in [Ids -> {"none", "notfolded", "folded", "refuted"}]
    /\ into \in [Ids -> Goals \cup {"none"}]
    /\ draft \in [Ids -> BOOLEAN]
    /\ reopened \in BOOLEAN

Init ==
    /\ gs = [g \in Goals |-> "written"]
    /\ f = [i \in Ids |-> "none"]
    /\ into = [i \in Ids |-> "none"]
    /\ draft = [i \in Ids |-> FALSE]
    /\ reopened = FALSE
    /\ out = Blank

Free(i) == f[i] = "notfolded"
OfEnded(i) == f[i] = "folded" /\ into[i] \in Goals /\ gs[into[i]] = "ended"

\* The state a read shows for finding i: calculated, never written.
Read(i) ==
    IF OfEnded(i) /\ ~NoEndedMark
      THEN "folded-ended"
      ELSE f[i]


\* --- the owner and the runner move Goals ---------------------------------

GoalPlaced(g) ==
    /\ gs[g] = "written" /\ gs' = [gs EXCEPT ![g] = "open"]
    /\ out' = Out("place", TRUE, "none")
    /\ UNCHANGED <<f, into, draft, reopened>>

GoalSelected(g) ==
    /\ gs[g] = "open" /\ gs' = [gs EXCEPT ![g] = "selected"]
    /\ out' = Out("select", TRUE, "none")
    /\ UNCHANGED <<f, into, draft, reopened>>

GoalEnds(g) ==
    /\ gs[g] # "ended" /\ gs' = [gs EXCEPT ![g] = "ended"]
    /\ out' = Out("end", TRUE, "none")
    /\ UNCHANGED <<f, into, draft, reopened>>

\* --- agents act on findings -----------------------------------------------

\* An agent records finding i with its fact and evidence and no Goal.
AgentAdds(i) ==
    /\ f[i] = "none"
    /\ IF AddNeedsGoal
         THEN out' = Out("add", FALSE, "needs-goal") /\ UNCHANGED f
         ELSE f' = [f EXCEPT ![i] = "notfolded"] /\ out' = Out("add", TRUE, "none")
    /\ UNCHANGED <<gs, into, draft, reopened>>

\* Why a fold of S into target t is refused, or "none".
FoldRefusal(S, t) ==
    IF t = Missing /\ ~FoldIntoMissing THEN "missing"
    ELSE IF t # Missing /\ gs[t] = "ended" /\ ~FoldIntoEnded THEN "ended"
    ELSE IF \E i \in S : f[i] = "refuted" /\ ~RefutedReopens THEN "refuted"
    ELSE IF NoRefold /\ \E i \in S : f[i] = "folded" THEN "folded"
    ELSE IF EndFreezes /\ \E i \in S : OfEnded(i) THEN "frozen"
    ELSE "none"

\* An agent folds the findings S into Goal t (or a Goal id that is missing).
AgentFolds(S, t) ==
    /\ S # {} /\ \A i \in S : f[i] # "none"
    /\ LET why == FoldRefusal(S, t) IN
       IF why = "none"
         THEN /\ f' = [i \in Ids |-> IF i \in S THEN "folded" ELSE f[i]]
              /\ into' = [i \in Ids |-> IF i \in S /\ t # Missing THEN t ELSE into[i]]
              /\ reopened' = (reopened \/ \E i \in S : f[i] = "refuted")
              /\ out' = [Out("fold", TRUE, "none") EXCEPT !.why = IF t = Missing THEN "missing" ELSE IF gs[t] = "ended" THEN "ended" ELSE "work"]
         ELSE /\ out' = Out("fold", FALSE, why)
              /\ UNCHANGED <<f, into, reopened>>
    /\ UNCHANGED <<gs, draft>>

\* An agent refutes finding i.
AgentRefutes(i) ==
    /\ f[i] # "none"
    /\ IF f[i] # "refuted"
         THEN f' = [f EXCEPT ![i] = "refuted"] /\ out' = Out("refute", TRUE, "none")
         ELSE out' = Out("refute", FALSE, "refuted") /\ UNCHANGED f
    /\ UNCHANGED <<gs, into, draft, reopened>>

\* An agent attaches a drafted add goal to finding i.
AgentDrafts(i) ==
    /\ f[i] \in {"notfolded", "folded"} /\ ~draft[i]
    /\ IF f[i] = "notfolded" \/ DraftOnFolded
         THEN draft' = [draft EXCEPT ![i] = TRUE] /\ out' = Out("draft", TRUE, f[i])
         ELSE out' = Out("draft", FALSE, "folded") /\ UNCHANGED draft
    /\ UNCHANGED <<gs, f, into, reopened>>

\* --- reads -----------------------------------------------------------------

\* hyp finding list, with mode "plain", "notfolded" or "all".
Listed(i, mode) ==
    /\ f[i] # "none"
    /\ CASE mode = "notfolded" -> Free(i) \/ NotFoldedShowsAll
         [] mode = "all"       -> TRUE
         [] OTHER              -> ~OfEnded(i) \/ ListShowsEnded

AgentLists(mode) ==
    /\ out' = [Blank EXCEPT !.cmd = "list", !.why = mode,
               !.shown = [i \in Ids |-> IF Listed(i, mode) THEN Read(i) ELSE "absent"],
               !.acts = ~ListNoActs,
               !.drafts = (~ListNoDraft /\ \E i \in Ids : Listed(i, mode) /\ draft[i])]
    /\ UNCHANGED <<gs, f, into, draft, reopened>>

\* hyp status: the words of each Goal it shows, and the count line.
WordsShown(g) ==
    /\ ~StatusNoWords
    /\ (gs[g] \in {"written", "open"} \/ (WordsOfSelected /\ gs[g] = "selected"))

AgentReadsStatus ==
    /\ out' = [Blank EXCEPT !.cmd = "status",
               !.words = [g \in Goals |-> WordsShown(g)],
               !.count = IF StatusNoCount THEN -1
                         ELSE Cardinality({i \in Ids : Free(i)})]
    /\ UNCHANGED <<gs, f, into, draft, reopened>>

\* hyp status G: its why, and the findings folded into G.
AgentReadsGoal(g) ==
    /\ out' = [Blank EXCEPT !.cmd = "status-g", !.why = g,
               !.goalWhy = ~StatusGNoWhy,
               !.shown = [i \in Ids |->
                          IF ~StatusGNoFolded /\ f[i] = "folded" /\ into[i] = g
                            THEN Read(i) ELSE "absent"]]
    /\ UNCHANGED <<gs, f, into, draft, reopened>>

Next ==
    \/ \E g \in Goals : GoalPlaced(g) \/ GoalSelected(g) \/ GoalEnds(g)
    \/ \E i \in Ids : AgentAdds(i) \/ AgentRefutes(i) \/ AgentDrafts(i)
    \/ \E S \in SUBSET Ids, t \in Goals \cup {Missing} : AgentFolds(S, t)
    \/ \E m \in {"plain", "notfolded", "all"} : AgentLists(m)
    \/ AgentReadsStatus
    \/ \E g \in Goals : AgentReadsGoal(g)

Spec == Init /\ [][Next]_vars

\* source: model only
TypeOKRule == TypeOK

\* source: line-1
AddsWithNoGoal == ~(out.cmd = "add" /\ ~out.ok)

\* source: line-2
FoldOnlyIntoWork == ~(out.cmd = "fold" /\ out.ok /\ out.why \in {"ended", "missing"})

\* source: line-2
FoldedFoldsAgain == ~(out.cmd = "fold" /\ ~out.ok /\ out.why = "folded")

\* source: line-3
EndedFoldsAgain == ~(out.cmd = "fold" /\ ~out.ok /\ out.why = "frozen")

\* source: line-3
RefutedIsFinal == ~reopened

\* source: line-3
ReadsShowEnded ==
    \A i \in Ids : (out.shown[i] # "absent" /\ OfEnded(i)) => out.shown[i] = "folded-ended"

\* source: line-4
ListHidesEnded ==
    (out.cmd = "list" /\ out.why = "plain") =>
        \A i \in Ids : f[i] # "none" => (out.shown[i] # "absent" <=> ~OfEnded(i))

\* source: line-4
NotFoldedOnlyFree ==
    (out.cmd = "list" /\ out.why = "notfolded") =>
        \A i \in Ids : f[i] # "none" => (out.shown[i] # "absent" <=> Free(i))

\* source: line-4
ListPrintsActs == out.cmd = "list" => out.acts

\* source: line-5
DraftOnlyOnFree == ~(out.cmd = "draft" /\ out.ok /\ out.why = "folded")

\* source: line-5
ListShowsDraft ==
    out.cmd = "list" => (out.drafts <=> \E i \in Ids : out.shown[i] # "absent" /\ draft[i])

\* source: line-6
StatusWordsInWork ==
    out.cmd = "status" => \A g \in Goals : out.words[g] <=> gs[g] \in {"written", "open"}

\* source: line-6
StatusCountsFree ==
    out.cmd = "status" => out.count = Cardinality({i \in Ids : Free(i)})

\* source: line-7
StatusGPrintsWhy == out.cmd = "status-g" => out.goalWhy

\* source: line-8
StatusGListsFolded ==
    out.cmd = "status-g" =>
        \A i \in Ids : (f[i] = "folded" /\ into[i] = out.why) => out.shown[i] # "absent"
=============================================================================
