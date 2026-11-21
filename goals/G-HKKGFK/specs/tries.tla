------------------------------ MODULE tries ------------------------------
(***************************************************************************)
(* How the grill tries one variant of a Goal's model, and what the         *)
(* statechart report says separates it, Goal G-HKKGFK: "The grill stops    *)
(* trying a variant at its first failed example, and the statechart report *)
(* counts that failed try as separating the variant."                      *)
(*                                                                         *)
(* The Goal has two examples, a and b. Example b sits on the source line   *)
(* of a rule the variant breaks; a does not. At the start the model picks  *)
(* what the variant's patch gives on each example (truth) and what the     *)
(* tool predicts for each (pred: "fail" or "none"; a predicted met orders  *)
(* and counts like no prediction). The grill tries examples one at a time  *)
(* with hyp try --worktree, in the order seq, and the tool then writes the *)
(* report: sep is the "Separated by" list, and gap holds when the report   *)
(* writes the variant's part "no example separates it".                    *)
(*                                                                         *)
(* Variants. KeepTrying is how H0 tries: every example, after a fail too.  *)
(* SpotCheckAfterStop tries a predicted-fail example once more after the   *)
(* stop. LineOrderFirst tries in line order. PredictionOnly is how H0      *)
(* reports: only predicted fails separate. TriedOnly counts only tried     *)
(* fails. AnyTried counts every tried example. PredictedOverTried lets a   *)
(* predicted fail separate though its try met. tries.cfg sets all FALSE.   *)
(* Left out: more than one variant (each is tried alone), the walks and    *)
(* the predictions themselves, and the owner's question, which decision 15 *)
(* on #283 holds and this Goal does not change.                            *)
(***************************************************************************)
EXTENDS Naturals, Sequences, FiniteSets

CONSTANTS
    KeepTrying, SpotCheckAfterStop, LineOrderFirst,
    PredictionOnly, TriedOnly, AnyTried, PredictedOverTried

E == {"a", "b"}
Src == {"b"}
LineRank(e) == IF e = "a" THEN 1 ELSE 2

VARIABLES truth, pred, tried, seq, phase, sep, gap
vars == <<truth, pred, tried, seq, phase, sep, gap>>

Init ==
    /\ truth \in [E -> {"met", "fail"}]
    /\ pred \in [E -> {"fail", "none"}]
    /\ tried = [e \in E |-> "no"]
    /\ seq = << >>
    /\ phase = "trying"
    /\ sep = {}
    /\ gap = FALSE

HasFail == \E e \in E : tried[e] = "fail"
Untried == {e \in E : tried[e] = "no"}
Rank(e) == IF pred[e] = "fail" THEN 0 ELSE IF e \in Src THEN 1 ELSE 2
Next1(e) ==
    IF LineOrderFirst
    THEN \A d \in Untried : LineRank(e) <= LineRank(d)
    ELSE \A d \in Untried : Rank(e) <= Rank(d)

Record(e) ==
    /\ tried' = [tried EXCEPT ![e] = truth[e]]
    /\ seq' = Append(seq, e)
    /\ UNCHANGED <<truth, pred, phase, sep, gap>>

GrillTries(e) ==
    /\ phase = "trying"
    /\ e \in Untried
    /\ Next1(e)
    /\ (~HasFail \/ KeepTrying)
    /\ Record(e)

GrillSpotChecks(e) ==
    /\ SpotCheckAfterStop
    /\ phase = "trying"
    /\ HasFail
    /\ e \in Untried
    /\ pred[e] = "fail"
    /\ Record(e)

Counted ==
    IF PredictionOnly THEN {e \in E : pred[e] = "fail"}
    ELSE IF TriedOnly THEN {e \in E : tried[e] = "fail"}
    ELSE IF AnyTried THEN {e \in E : tried[e] # "no" \/ pred[e] = "fail"}
    ELSE IF PredictedOverTried THEN {e \in E : tried[e] = "fail" \/ pred[e] = "fail"}
    ELSE {e \in E : tried[e] = "fail" \/ (tried[e] = "no" /\ pred[e] = "fail")}

ToolReports ==
    /\ phase = "trying"
    /\ HasFail \/ Untried = {}
    /\ phase' = "reported"
    /\ sep' = Counted
    /\ gap' = (Counted = {})
    /\ UNCHANGED <<truth, pred, tried, seq>>

Done == phase = "reported" /\ UNCHANGED vars

Next == (\E e \in E : GrillTries(e) \/ GrillSpotChecks(e)) \/ ToolReports \/ Done

Spec == Init /\ [][Next]_vars

-----------------------------------------------------------------------------
\* source: model only. No variant breaks it, so tries.cfg leaves it out.
TypeOK ==
    /\ truth \in [E -> {"met", "fail"}]
    /\ pred \in [E -> {"fail", "none"}]
    /\ tried \in [E -> {"no", "met", "fail"}]
    /\ phase \in {"trying", "reported"}
    /\ sep \subseteq E
    /\ gap \in BOOLEAN

\* source: line-1
FailedTrySeparates ==
    phase = "reported" => \A e \in E : tried[e] = "fail" => e \in sep

\* source: line-1
UntriedPredictedFailSeparates ==
    phase = "reported" => \A e \in E : tried[e] = "no" /\ pred[e] = "fail" => e \in sep

\* source: line-1
TriedMetStands ==
    phase = "reported" => \A e \in E : tried[e] = "met" => e \notin sep

\* source: line-2
StopAtFirstFail ==
    \A i \in 1..Len(seq) : i < Len(seq) => tried[seq[i]] # "fail"

\* source: line-2
FirstTryOrder ==
    Len(seq) >= 1 =>
        IF \E e \in E : pred[e] = "fail" THEN pred[seq[1]] = "fail"
        ELSE seq[1] \in Src

\* source: line-3
AllMetIsAGap ==
    phase = "reported"
    /\ (\A e \in E : tried[e] # "fail")
    /\ ~(\E e \in E : tried[e] = "no" /\ pred[e] = "fail")
    => sep = {} /\ gap

=============================================================================
