------------------------------ MODULE decline ------------------------------
(***************************************************************************)
(* Goal G-QC8DF5: "The owner can end a true finding that the owner does    *)
(* not want fixed, with a why, and a refute keeps its meaning: the fact is *)
(* false."                                                                 *)
(*                                                                         *)
(* A Goal is "work" (written, open or selected) or "ended". A finding is   *)
(* "none" (not added), "notfolded", "folded" (into into), "refuted" or     *)
(* "declined". A decline names an actor's role and has a why or not.       *)
(*                                                                         *)
(* Variants, each FALSE in decline.cfg, are cheaper builds:                *)
(*   AnyRole          a decline by a role other than owner is recorded     *)
(*   NoWhy            a decline with no why is recorded                    *)
(*   DeclineHeld      a finding a Goal in work holds is declined           *)
(*   DeclinedTakesActs a declined finding takes a fold or a refute         *)
(*   DeclinedListed   plain hyp finding list shows a declined finding      *)
(*   AllHidesDeclined hyp finding list --all leaves a declined one out     *)
(*   DeclinedCounted  hyp status counts a declined finding as not folded   *)
(*   DeclineRefutes   the decline records the finding as refuted           *)
(***************************************************************************)
EXTENDS Integers, FiniteSets

CONSTANTS Ids, Goals,
    AnyRole, NoWhy, DeclineHeld, DeclinedTakesActs, DeclinedListed,
    AllHidesDeclined, DeclinedCounted, DeclineRefutes

VARIABLES
    gs,    \* per Goal: "work" or "ended"
    f,     \* per finding: "none", "notfolded", "folded", "refuted" or "declined"
    into,  \* per finding: the Goal it is folded into, "none" for none
    out    \* what the last command printed

vars == <<gs, f, into, out>>

NoShow == [i \in Ids |-> "absent"]
Blank == [cmd |-> "none", ok |-> TRUE, id |-> "none", role |-> "none",
          hasWhy |-> FALSE, prior |-> "none", mode |-> "none",
          shown |-> NoShow, count |-> -1]

TypeOK ==
    /\ gs \in [Goals -> {"work", "ended"}]
    /\ f \in [Ids -> {"none", "notfolded", "folded", "refuted", "declined"}]
    /\ into \in [Ids -> Goals \cup {"none"}]

Init ==
    /\ gs = [g \in Goals |-> "work"]
    /\ f = [i \in Ids |-> "none"]
    /\ into = [i \in Ids |-> "none"]
    /\ out = Blank

Held(i) == f[i] = "folded" /\ gs[into[i]] = "work"
\* The state a refusal or a record names: a fold into a Goal in work is "held".
Prior(i) == IF Held(i) THEN "held" ELSE f[i]

\* --- the owner ends a Goal ------------------------------------------------

GoalEnds(g) ==
    /\ gs[g] = "work" /\ gs' = [gs EXCEPT ![g] = "ended"]
    /\ out' = [Blank EXCEPT !.cmd = "end"]
    /\ UNCHANGED <<f, into>>

\* --- agents act on the finding --------------------------------------------

AgentAdds(i) ==
    /\ f[i] = "none"
    /\ f' = [f EXCEPT ![i] = "notfolded"]
    /\ out' = [Blank EXCEPT !.cmd = "add", !.id = i]
    /\ UNCHANGED <<gs, into>>

Open(i) == f[i] \in {"notfolded", "folded"} \/ (f[i] = "declined" /\ DeclinedTakesActs)

AgentFolds(i, t) ==
    /\ f[i] # "none" /\ f[i] # "refuted" /\ gs[t] = "work"
    /\ IF Open(i)
         THEN /\ f' = [f EXCEPT ![i] = "folded"] /\ into' = [into EXCEPT ![i] = t]
              /\ out' = [Blank EXCEPT !.cmd = "fold", !.id = i, !.prior = Prior(i)]
         ELSE /\ out' = [Blank EXCEPT !.cmd = "fold", !.ok = FALSE, !.id = i, !.prior = Prior(i)]
              /\ UNCHANGED <<f, into>>
    /\ UNCHANGED gs

AgentRefutes(i) ==
    /\ f[i] # "none" /\ f[i] # "refuted"
    /\ IF Open(i)
         THEN /\ f' = [f EXCEPT ![i] = "refuted"]
              /\ out' = [Blank EXCEPT !.cmd = "refute", !.id = i, !.prior = Prior(i)]
         ELSE /\ out' = [Blank EXCEPT !.cmd = "refute", !.ok = FALSE, !.id = i, !.prior = Prior(i)]
              /\ UNCHANGED f
    /\ UNCHANGED <<gs, into>>

\* An actor with role r declines finding i, with a why (w) or not.
Declines(i, r, w) ==
    /\ f[i] # "none"
    /\ LET okRole == r = "owner" \/ AnyRole
           okWhy == w \/ NoWhy
           okState == f[i] = "notfolded" \/ (f[i] = "folded" /\ ~Held(i))
                      \/ (Held(i) /\ DeclineHeld)
                      \/ (f[i] = "declined" /\ DeclinedTakesActs)
       IN IF okRole /\ okWhy /\ okState
            THEN /\ f' = [f EXCEPT ![i] = IF DeclineRefutes THEN "refuted" ELSE "declined"]
                 /\ out' = [Blank EXCEPT !.cmd = "decline", !.id = i, !.role = r,
                            !.hasWhy = w, !.prior = Prior(i)]
            ELSE /\ out' = [Blank EXCEPT !.cmd = "decline", !.ok = FALSE, !.id = i,
                            !.role = r, !.hasWhy = w, !.prior = Prior(i)]
                 /\ UNCHANGED f
    /\ UNCHANGED <<gs, into>>

\* --- reads -----------------------------------------------------------------

OfEnded(i) == f[i] = "folded" /\ gs[into[i]] = "ended"
Listed(i, m) ==
    /\ f[i] # "none"
    /\ IF m = "all" THEN ~(f[i] = "declined" /\ AllHidesDeclined)
       ELSE ~OfEnded(i) /\ (f[i] # "declined" \/ DeclinedListed)

AgentLists(m) ==
    /\ out' = [Blank EXCEPT !.cmd = "list", !.mode = m,
               !.shown = [i \in Ids |-> IF Listed(i, m) THEN Prior(i) ELSE "absent"]]
    /\ UNCHANGED <<gs, f, into>>

AgentReadsStatus ==
    /\ out' = [Blank EXCEPT !.cmd = "status",
               !.count = Cardinality({i \in Ids : f[i] = "notfolded"
                                      \/ (f[i] = "declined" /\ DeclinedCounted)})]
    /\ UNCHANGED <<gs, f, into>>

Next ==
    \/ \E g \in Goals : GoalEnds(g)
    \/ \E i \in Ids : AgentAdds(i) \/ AgentRefutes(i)
    \/ \E i \in Ids, t \in Goals : AgentFolds(i, t)
    \/ \E i \in Ids, r \in {"owner", "runner"}, w \in BOOLEAN : Declines(i, r, w)
    \/ \E m \in {"plain", "all"} : AgentLists(m)
    \/ AgentReadsStatus

Spec == Init /\ [][Next]_vars

\* source: line-1
DeclineReadsDeclined == (out.cmd = "decline" /\ out.ok) => f[out.id] = "declined"

\* source: line-1
DeclineOwnerWithWhy == (out.cmd = "decline" /\ out.ok) => (out.role = "owner" /\ out.hasWhy)

\* source: line-2
NoDeclineOfHeld == (out.cmd = "decline" /\ out.ok) => out.prior \in {"notfolded", "folded"}

\* source: line-2
DeclinedIsFinal == (out.cmd \in {"fold", "refute", "decline"} /\ out.prior = "declined") => ~out.ok

\* source: line-3
PlainListHidesDeclined == (out.cmd = "list" /\ out.mode = "plain") => \A i \in Ids : out.shown[i] # "declined"

\* source: line-3
AllListsDeclined == (out.cmd = "list" /\ out.mode = "all") => \A i \in Ids : f[i] = "declined" => out.shown[i] = "declined"

\* source: line-3
StatusCountsNotFolded == out.cmd = "status" => out.count = Cardinality({i \in Ids : f[i] = "notfolded"})
=============================================================================
