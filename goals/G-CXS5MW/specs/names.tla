------------------------------- MODULE names -------------------------------
(***************************************************************************)
(* Goal G-CXS5MW: "A Goal names the findings it is about in its own        *)
(* record, so it reads whole on any clone, and each finding's state        *)
(* follows from the Goals that name it."                                   *)
(*                                                                         *)
(* A Goal is "none", "work" (written, open or selected), "landed" (landed  *)
(* with a select) or "endedno" (closed, or landed with no select). The     *)
(* owner names findings at add goal (own); the grill names them in an      *)
(* example at place (ex). An old fold event counts as a name (old). A      *)
(* finding is "none" (not in this clone's trace), "added", "refuted" or    *)
(* "declined". Reads run on this clone ("here") or another ("other"),      *)
(* whose trace holds no finding.                                           *)
(*                                                                         *)
(* Variants, each FALSE in names.cfg, are cheaper builds:                  *)
(*   PlaceNoCover     place does not check that each named finding has an  *)
(*                    example that names it                                *)
(*   NoIdCheck        a name of an id the trace lacks is recorded          *)
(*   NameFinal        a refuted or declined finding can be named           *)
(*   TwoInWork        two Goals in work may name one finding               *)
(*   NoSettle         a land with a select reads ended, not settled        *)
(*   EndedNotWaits    ended, not settled counts among findings that wait   *)
(*   OldFoldIgnored   an old fold is not a name                            *)
(*   StatusGFromTrace status G prints no id on a clone without the trace   *)
(*   FoldKept         finding fold still records a fold                    *)
(*   DeclineNamed     a finding named only by ended Goals can be declined  *)
(***************************************************************************)
EXTENDS Integers, FiniteSets

CONSTANTS Ids, Goals,
    PlaceNoCover, NoIdCheck, NameFinal, TwoInWork, NoSettle, EndedNotWaits,
    OldFoldIgnored, StatusGFromTrace, FoldKept, DeclineNamed

VARIABLES
    gs,     \* per Goal: "none", "work", "landed" or "endedno"
    placed, \* per Goal: TRUE once a set is placed
    own,    \* per Goal: the findings the owner named at add goal
    ex,     \* per Goal: the findings its examples name
    old,    \* per Goal: the findings an old fold event folded into it
    f,      \* per finding: "none", "added", "refuted" or "declined"
    out     \* what the last command printed

vars == <<gs, placed, own, ex, old, f, out>>

Blank == [cmd |-> "none", ok |-> TRUE, g |-> "none", id |-> "none", why |-> "none",
          clone |-> "here", shown |-> [i \in Ids |-> "absent"], ids |-> {}, count |-> -1]

TypeOK ==
    /\ gs \in [Goals -> {"none", "work", "landed", "endedno"}]
    /\ f \in [Ids -> {"none", "added", "refuted", "declined"}]

\* At the start, Goal ONE may have ended long ago with an old fold of finding ONE.
Init ==
    /\ \E start \in {"none", "landed", "endedno"} :
         /\ gs = [g \in Goals |-> IF g = "ONE" THEN start ELSE "none"]
         /\ old = [g \in Goals |-> IF g = "ONE" /\ start # "none" THEN Ids ELSE {}]
         /\ f = [i \in Ids |-> IF start # "none" THEN "added" ELSE "none"]
         /\ placed = [g \in Goals |-> g = "ONE" /\ start # "none"]
    /\ own = [g \in Goals |-> {}]
    /\ ex = [g \in Goals |-> {}]
    /\ out = Blank

Named(g) == own[g] \cup ex[g] \cup (IF OldFoldIgnored THEN {} ELSE old[g])
NamedBy(i) == {g \in Goals : i \in Named(g)}
InWork == {g \in Goals : gs[g] = "work"}

FState(i) ==
    IF f[i] \in {"refuted", "declined"} THEN f[i]
    ELSE IF \E g \in NamedBy(i) : gs[g] = "work" THEN "folded"
    ELSE IF \E g \in NamedBy(i) : gs[g] = "landed" /\ ~NoSettle THEN "settled"
    ELSE IF NamedBy(i) # {} THEN "endednot"
    ELSE "waiting"

\* Why a name of finding i by Goal g is refused, or "none".
NameFault(i, g) ==
    IF f[i] = "none" /\ ~NoIdCheck THEN "unknown"
    ELSE IF f[i] \in {"refuted", "declined"} /\ ~NameFinal THEN "final"
    ELSE IF (\E h \in InWork \ {g} : i \in Named(h)) /\ ~TwoInWork THEN "twowork"
    ELSE "none"

\* --- the owner and the grill act on Goals ---------------------------------

OwnerAddsGoal(g, S) ==
    /\ gs[g] = "none"
    /\ LET bad == {i \in S : NameFault(i, g) # "none"}
       IN IF bad = {}
            THEN /\ gs' = [gs EXCEPT ![g] = "work"] /\ own' = [own EXCEPT ![g] = S]
                 /\ out' = [Blank EXCEPT !.cmd = "add", !.g = g, !.ids = S]
                 /\ UNCHANGED <<placed, ex, old, f>>
            ELSE \E i \in bad :
                 /\ out' = [Blank EXCEPT !.cmd = "add", !.ok = FALSE, !.g = g, !.id = i,
                            !.why = NameFault(i, g)]
                 /\ UNCHANGED <<gs, placed, own, ex, old, f>>

GrillPlaces(g, E) ==
    /\ gs[g] = "work" /\ ~placed[g]
    /\ LET bad == {i \in E : NameFault(i, g) # "none"}
           uncovered == own[g] \ E
       IN IF bad # {}
            THEN \E i \in bad :
                 /\ out' = [Blank EXCEPT !.cmd = "place", !.ok = FALSE, !.g = g, !.id = i,
                            !.why = NameFault(i, g)]
                 /\ UNCHANGED <<gs, placed, own, ex, old, f>>
          ELSE IF uncovered # {} /\ ~PlaceNoCover
            THEN \E i \in uncovered :
                 /\ out' = [Blank EXCEPT !.cmd = "place", !.ok = FALSE, !.g = g, !.id = i,
                            !.why = "uncovered"]
                 /\ UNCHANGED <<gs, placed, own, ex, old, f>>
          ELSE /\ placed' = [placed EXCEPT ![g] = TRUE] /\ ex' = [ex EXCEPT ![g] = E]
               /\ out' = [Blank EXCEPT !.cmd = "place", !.g = g, !.ids = E]
               /\ UNCHANGED <<gs, own, old, f>>

OwnerLands(g, sel) ==
    /\ gs[g] = "work" /\ placed[g]
    /\ gs' = [gs EXCEPT ![g] = IF sel THEN "landed" ELSE "endedno"]
    /\ out' = [Blank EXCEPT !.cmd = "end", !.g = g]
    /\ UNCHANGED <<placed, own, ex, old, f>>

OwnerCloses(g) ==
    /\ gs[g] = "work"
    /\ gs' = [gs EXCEPT ![g] = "endedno"]
    /\ out' = [Blank EXCEPT !.cmd = "end", !.g = g]
    /\ UNCHANGED <<placed, own, ex, old, f>>

\* --- agents and the owner act on findings ---------------------------------

AgentAdds(i) ==
    /\ f[i] = "none" /\ f' = [f EXCEPT ![i] = "added"]
    /\ out' = [Blank EXCEPT !.cmd = "addf", !.id = i]
    /\ UNCHANGED <<gs, placed, own, ex, old>>

AgentRefutes(i) ==
    /\ f[i] = "added" /\ f' = [f EXCEPT ![i] = "refuted"]
    /\ out' = [Blank EXCEPT !.cmd = "refute", !.id = i]
    /\ UNCHANGED <<gs, placed, own, ex, old>>

OwnerDeclines(i) ==
    /\ f[i] = "added"
    /\ IF NamedBy(i) = {} \/ (DeclineNamed /\ FState(i) = "endednot")
         THEN /\ f' = [f EXCEPT ![i] = "declined"]
              /\ out' = [Blank EXCEPT !.cmd = "decline", !.id = i]
         ELSE /\ out' = [Blank EXCEPT !.cmd = "decline", !.ok = FALSE, !.id = i]
              /\ UNCHANGED f
    /\ UNCHANGED <<gs, placed, own, ex, old>>

\* The old act. Under FoldKept it still names the finding for Goal g.
AgentFolds(i, g) ==
    /\ f[i] = "added" /\ gs[g] = "work"
    /\ IF FoldKept
         THEN /\ own' = [own EXCEPT ![g] = own[g] \cup {i}]
              /\ out' = [Blank EXCEPT !.cmd = "fold", !.id = i, !.g = g]
         ELSE /\ out' = [Blank EXCEPT !.cmd = "fold", !.ok = FALSE, !.id = i, !.g = g]
              /\ UNCHANGED own
    /\ UNCHANGED <<gs, placed, ex, old, f>>

\* --- reads -----------------------------------------------------------------

AgentLists ==
    /\ out' = [Blank EXCEPT !.cmd = "list",
               !.shown = [i \in Ids |-> IF f[i] = "none" THEN "absent" ELSE FState(i)]]
    /\ UNCHANGED <<gs, placed, own, ex, old, f>>

Waits(i) == f[i] = "added" /\ (FState(i) = "waiting" \/ (EndedNotWaits /\ FState(i) = "endednot"))

AgentReadsStatus ==
    /\ out' = [Blank EXCEPT !.cmd = "status", !.count = Cardinality({i \in Ids : Waits(i)})]
    /\ UNCHANGED <<gs, placed, own, ex, old, f>>

\* hyp status G on clone c. The other clone's trace holds no finding.
AgentReadsGoal(g, c) ==
    /\ gs[g] # "none"
    /\ LET held(i) == c = "here" /\ f[i] # "none"
       IN out' = [Blank EXCEPT !.cmd = "statusg", !.g = g, !.clone = c,
                  !.ids = IF StatusGFromTrace THEN {i \in Named(g) : held(i)} ELSE Named(g),
                  !.shown = [i \in Ids |-> IF i \notin Named(g) THEN "absent"
                                           ELSE IF held(i) THEN FState(i) ELSE "nottrace"]]
    /\ UNCHANGED <<gs, placed, own, ex, old, f>>

Next ==
    \/ \E g \in Goals, S \in SUBSET Ids : OwnerAddsGoal(g, S)
    \/ \E g \in Goals, E \in SUBSET Ids : GrillPlaces(g, E)
    \/ \E g \in Goals, s \in BOOLEAN : OwnerLands(g, s)
    \/ \E g \in Goals : OwnerCloses(g)
    \/ \E i \in Ids : AgentAdds(i) \/ AgentRefutes(i) \/ OwnerDeclines(i)
    \/ \E i \in Ids, g \in Goals : AgentFolds(i, g)
    \/ AgentLists \/ AgentReadsStatus
    \/ \E g \in Goals, c \in {"here", "other"} : AgentReadsGoal(g, c)

Spec == Init /\ [][Next]_vars

\* source: line-1
StatusGPrintsEveryName == out.cmd = "statusg" => out.ids = Named(out.g)

\* source: line-2
PlaceCoversNames == \A g \in Goals : placed[g] /\ gs[g] = "work" => own[g] \subseteq ex[g]

\* source: line-3
NoNameOfUnknown == (out.cmd \in {"add", "place"} /\ out.ok) =>
    \A i \in out.ids : f[i] = "added"

\* source: line-3
OneWorkGoalPerFinding == \A i \in Ids : Cardinality({g \in InWork : i \in Named(g)}) <= 1

\* source: line-4
LandWithSelectSettles == \A i \in Ids :
    (f[i] = "added" /\ ~(\E g \in NamedBy(i) : gs[g] = "work")
      /\ (\E g \in Goals : gs[g] = "landed" /\ i \in own[g] \cup ex[g] \cup old[g]))
      => FState(i) = "settled"

\* source: line-4
EndedNotDoesNotWait == out.cmd = "status" =>
    out.count = Cardinality({i \in Ids : f[i] = "added" /\ FState(i) = "waiting"})

\* source: line-4
OldFoldIsName == \A i \in Ids, g \in Goals : (i \in old[g] /\ f[i] = "added") => FState(i) # "waiting"

\* source: line-5
NoFoldAct == ~(out.cmd = "fold" /\ out.ok)

\* source: line-7
NoDeclineOfNamed == (out.cmd = "decline" /\ out.ok) =>
    \A g \in Goals : out.id \notin own[g] \cup ex[g] \cup old[g]
=============================================================================
