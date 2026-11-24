----------------------------- MODULE foldlines -----------------------------
(***************************************************************************)
(* Goal G-G5F1JD: "The finding acts that hyp prints as ready lines are     *)
(* exactly the acts each finding allows, and a fold line leaves the Goal   *)
(* for the agent to choose."                                               *)
(*                                                                         *)
(* Each Goal is "work" (written, open or selected: a fold may name it) or *)
(* "ended". The finding is "none" (not added), "notfolded", "folded" (into *)
(* into) or "refuted". hyp finding list prints, for each finding it shows, *)
(* a may: block of ready lines: the fold lines, by the Goal each names, or *)
(* "G" for a Goal the agent fills in, and the refute line.                 *)
(*                                                                         *)
(* Variants, each FALSE in foldlines.cfg, are cheaper builds:              *)
(*   NoRefoldLine   a folded finding prints no fold line (H0)              *)
(*   NamesGoals     a fold line for each Goal in work but its own (H0)     *)
(*   EndedNoRefold  a finding of an ended Goal prints no fold line         *)
(*   FoldLineNoWork the fold line prints while no Goal is in work          *)
(*   RefutedActs    a refuted finding prints a may: block                  *)
(*   StatusGActs    hyp status G prints the may: block of its findings     *)
(*   FoldIntoEnded  a fold into an ended Goal is recorded                  *)
(***************************************************************************)
EXTENDS Integers, FiniteSets

CONSTANTS Ids, Goals,
    NoRefoldLine, NamesGoals, EndedNoRefold, FoldLineNoWork, RefutedActs,
    StatusGActs, FoldIntoEnded

VARIABLES
    gs,    \* per Goal: "work" or "ended"
    f,     \* per finding: "none", "notfolded", "folded" or "refuted"
    into,  \* per finding: the Goal it is folded into, "none" for none
    out    \* what the last command printed

vars == <<gs, f, into, out>>

NoFold == [i \in Ids |-> {}]
NoShow == [i \in Ids |-> "absent"]
Blank == [cmd |-> "none", ok |-> TRUE, why |-> "none", shown |-> NoShow,
          fold |-> NoFold, refute |-> [i \in Ids |-> FALSE], count |-> -1]
Out(cmd, ok, why) == [Blank EXCEPT !.cmd = cmd, !.ok = ok, !.why = why]

TypeOK ==
    /\ gs \in [Goals -> {"work", "ended"}]
    /\ f \in [Ids -> {"none", "notfolded", "folded", "refuted"}]
    /\ into \in [Ids -> Goals \cup {"none"}]

Init ==
    /\ gs = [g \in Goals |-> "work"]
    /\ f = [i \in Ids |-> "none"]
    /\ into = [i \in Ids |-> "none"]
    /\ out = Blank

InWork == {g \in Goals : gs[g] = "work"}
OfEnded(i) == f[i] = "folded" /\ into[i] \in Goals /\ gs[into[i]] = "ended"

\* --- the owner ends a Goal ------------------------------------------------

GoalEnds(g) ==
    /\ gs[g] = "work" /\ gs' = [gs EXCEPT ![g] = "ended"]
    /\ out' = Out("end", TRUE, "none")
    /\ UNCHANGED <<f, into>>

\* --- agents act on the finding --------------------------------------------

AgentAdds(i) ==
    /\ f[i] = "none"
    /\ f' = [f EXCEPT ![i] = "notfolded"]
    /\ out' = Out("add", TRUE, "none")
    /\ UNCHANGED <<gs, into>>

\* An agent folds finding i into Goal t, in any state, ended ones included.
AgentFolds(i, t) ==
    /\ f[i] \in {"notfolded", "folded"}
    /\ IF gs[t] = "ended" /\ ~FoldIntoEnded
         THEN out' = Out("fold", FALSE, t) /\ UNCHANGED <<f, into>>
         ELSE /\ f' = [f EXCEPT ![i] = "folded"]
              /\ into' = [into EXCEPT ![i] = t]
              /\ out' = [Out("fold", TRUE, t) EXCEPT !.count = IF gs[t] = "ended" THEN 1 ELSE 0]
    /\ UNCHANGED gs

AgentRefutes(i) ==
    /\ f[i] \in {"notfolded", "folded"}
    /\ f' = [f EXCEPT ![i] = "refuted"]
    /\ out' = Out("refute", TRUE, "none")
    /\ UNCHANGED <<gs, into>>

\* --- reads -----------------------------------------------------------------

\* The fold lines of finding i: the Goals they name, "G" for none named.
FoldLines(i) ==
    IF f[i] = "refuted" /\ ~RefutedActs THEN {}
    ELSE IF InWork = {} /\ ~FoldLineNoWork THEN {}
    ELSE IF f[i] = "folded" /\ NoRefoldLine THEN {}
    ELSE IF OfEnded(i) /\ EndedNoRefold THEN {}
    ELSE IF NamesGoals THEN {g \in InWork : g # into[i]}
    ELSE {"G"}

Mays(i) == f[i] # "refuted" \/ RefutedActs

\* hyp finding list, with mode "plain" or "all".
Listed(i, mode) == f[i] # "none" /\ (mode = "all" \/ ~OfEnded(i))

AgentLists(mode) ==
    /\ out' = [Blank EXCEPT !.cmd = "list", !.why = mode,
               !.shown = [i \in Ids |-> IF Listed(i, mode) THEN f[i] ELSE "absent"],
               !.fold = [i \in Ids |-> IF Listed(i, mode) THEN FoldLines(i) ELSE {}],
               !.refute = [i \in Ids |-> Listed(i, mode) /\ Mays(i)]]
    /\ UNCHANGED <<gs, f, into>>

\* hyp status, with no Goal named: the count of the findings not folded.
AgentReadsStatus ==
    /\ out' = [Blank EXCEPT !.cmd = "status",
               !.count = Cardinality({i \in Ids : f[i] = "notfolded"})]
    /\ UNCHANGED <<gs, f, into>>

\* hyp status G: the findings folded into G.
AgentReadsGoal(g) ==
    /\ out' = [Blank EXCEPT !.cmd = "status-g", !.why = g,
               !.shown = [i \in Ids |-> IF f[i] = "folded" /\ into[i] = g THEN "folded" ELSE "absent"],
               !.fold = [i \in Ids |-> IF StatusGActs /\ f[i] = "folded" /\ into[i] = g
                                         THEN FoldLines(i) ELSE {}],
               !.refute = [i \in Ids |-> StatusGActs /\ f[i] = "folded" /\ into[i] = g]]
    /\ UNCHANGED <<gs, f, into>>

Next ==
    \/ \E g \in Goals : GoalEnds(g)
    \/ \E i \in Ids : AgentAdds(i) \/ AgentRefutes(i)
    \/ \E i \in Ids, t \in Goals : AgentFolds(i, t)
    \/ \E m \in {"plain", "all"} : AgentLists(m)
    \/ AgentReadsStatus
    \/ \E g \in Goals : AgentReadsGoal(g)

Spec == Init /\ [][Next]_vars

\* source: model only
TypeOKRule == TypeOK

\* source: line-1
FoldedPrintsFoldLine ==
    out.cmd = "list" =>
        \A i \in Ids : (out.shown[i] = "folded" /\ InWork # {}) => out.fold[i] # {}

\* source: line-2
FoldLineNamesNoGoal == out.cmd = "list" => \A i \in Ids : out.fold[i] \subseteq {"G"}

\* source: line-3
NoFoldLineWithoutWork == (out.cmd = "list" /\ InWork = {}) => \A i \in Ids : out.fold[i] = {}

\* source: line-3
RefutedPrintsNoMay ==
    out.cmd = "list" =>
        \A i \in Ids : out.shown[i] = "refuted" => (out.fold[i] = {} /\ ~out.refute[i])

\* source: line-3
StatusGPrintsNoMay ==
    out.cmd = "status-g" => \A i \in Ids : out.fold[i] = {} /\ ~out.refute[i]

\* source: line-4
FoldOnlyIntoWork == ~(out.cmd = "fold" /\ out.ok /\ out.count = 1)
=============================================================================
