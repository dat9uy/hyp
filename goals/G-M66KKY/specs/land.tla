------------------------------- MODULE land -------------------------------
(***************************************************************************)
(* The owner's land of a Goal, Goal G-M66KKY: "A Goal's stop and its land  *)
(* are two events: a stop ends the runs, and the owner's land ends the     *)
(* Goal as a fixed snapshot that later Goals build on."                    *)
(*                                                                         *)
(* One Goal G. The grill places, the runner runs and selects, the owner    *)
(* selects among two candidates and lands. A run may reach a stop of the   *)
(* runs: two candidates or budget used; a place that changes an example    *)
(* moves the budget's boundary, so the Goal has no stop again. A select    *)
(* is the stop selected. The stop is calculated, never stored: here it is  *)
(* the variable stop, which only the acts above change. A second Goal may  *)
(* cite G with add goal --from. A clone that has not merged the land may   *)
(* still write to G; when the logs merge, replay checks each event again.  *)
(*                                                                         *)
(* The rules are the Goal's lines:                                         *)
(*   line 1  the owner lands at a stop; one land event names that stop; a  *)
(*           land by another role, or with no stop, is refused             *)
(*   line 4  after a select the Goal waits for the owner, and the owner's  *)
(*           acts hold the land                                            *)
(*   line 5  a landed Goal refuses place, run, select and a second land,   *)
(*           also when a merged log brings one in                          *)
(*   line 6  add goal --from G is refused unless G is landed               *)
(*   line 8  a selected Goal that is not landed refuses place, run and a   *)
(*           second select                                                 *)
(* Lines 2, 3 and 7 are what hyp prints and what the docs say; no state    *)
(* rule here reads them.                                                   *)
(*                                                                         *)
(* Switches are cheaper builds or other potential Hypotheses;              *)
(* land.cfg sets all FALSE.                                                *)
(*   AnyRoleLands    a build with no Owner lands guard                     *)
(*   LandWithNoStop  a build with no Stop guard                            *)
(*   NoReplayCheck   a build whose replay does not check Not ended         *)
(*   FromAnyGoal     a build with no From landed guard                     *)
(* Left out: the rows, the examples and their verdicts, which decide when  *)
(* a stop holds and which this model draws at random; and line 9, an old   *)
(* log, which holds no land and so no rule of this model.                 *)
(***************************************************************************)
EXTENDS Naturals

CONSTANTS AnyRoleLands, LandWithNoStop, NoReplayCheck, FromAnyGoal

Roles == {"owner", "grill", "runner"}
Stops == {"none", "selected", "two candidates", "budget used"}

VARIABLES
    stop,       \* the stop G is at, as hyp calculates it
    landed,     \* G has a land event
    landStop,   \* the stop the land event names, or "none"
    landRole,   \* the role of the land event, or "none"
    lands,      \* the land events of G the tables hold
    afterLand,  \* the writes to G the tables hold that came after its land
    pending,    \* a write from a clone that has not merged the land
    cites,      \* a later Goal's add goal --from G was recorded
    citedState  \* G was landed when it was cited

vars == <<stop, landed, landStop, landRole, lands, afterLand, pending, cites, citedState>>

TypeOK ==
    /\ stop \in Stops
    /\ landed \in BOOLEAN
    /\ landStop \in Stops
    /\ landRole \in Roles \cup {"none"}
    /\ lands \in 0..3
    /\ afterLand \in 0..3
    /\ pending \in BOOLEAN
    /\ cites \in BOOLEAN
    /\ citedState \in BOOLEAN

Init ==
    /\ stop = "none"
    /\ landed = FALSE
    /\ landStop = "none"
    /\ landRole = "none"
    /\ lands = 0
    /\ afterLand = 0
    /\ pending = FALSE
    /\ cites = FALSE
    /\ citedState = FALSE

\* Not ended, then Selected: the guards of place, run and select.
Writable == ~landed /\ stop # "selected"

\* A write the tables accept; one after the land only another build accepts.
Wrote == afterLand' = IF landed THEN afterLand + 1 ELSE afterLand

\* The grill places an example that changes the examples: the boundary moves.
Place ==
    /\ Writable
    /\ stop' = "none"
    /\ UNCHANGED <<landed, landStop, landRole, lands, afterLand, pending, cites, citedState>>

\* The runner runs a line; the run may reach a stop of the runs.
Run ==
    /\ Writable
    /\ stop = "none"
    /\ stop' \in {"none", "two candidates", "budget used"}
    /\ UNCHANGED <<landed, landStop, landRole, lands, afterLand, pending, cites, citedState>>

\* The runner selects its one candidate, or the owner one of two.
Select ==
    /\ Writable
    /\ stop \in {"none", "two candidates"}
    /\ stop' = "selected"
    /\ UNCHANGED <<landed, landStop, landRole, lands, afterLand, pending, cites, citedState>>

\* Not ended, Owner lands, then Stop. Goal's actor is left out: one owner.
LandAllowed(role) ==
    /\ ~landed
    /\ AnyRoleLands \/ role = "owner"
    /\ LandWithNoStop \/ stop # "none"

Land(role) ==
    /\ LandAllowed(role)
    /\ landed' = TRUE
    /\ landStop' = stop
    /\ landRole' = role
    /\ lands' = lands + 1
    /\ UNCHANGED <<stop, afterLand, pending, cites, citedState>>

\* A clone that has not merged the land writes to G: its guards held there.
CloneWrites ==
    /\ ~pending
    /\ stop # "selected"
    /\ pending' = TRUE
    /\ UNCHANGED <<stop, landed, landStop, landRole, lands, afterLand, cites, citedState>>

\* The logs merge. Replay checks Not ended again at the write's place in the
\* merge order: after a land the write is refused at replay and changes no
\* table. A write before any land is an ordinary write.
Merge ==
    /\ pending
    /\ pending' = FALSE
    /\ IF landed /\ ~NoReplayCheck
       THEN UNCHANGED <<stop, afterLand>>
       ELSE /\ Wrote
            /\ stop' = IF landed THEN stop ELSE "none"
    /\ UNCHANGED <<landed, landStop, landRole, lands, cites, citedState>>

\* A later Goal cites G with add goal --from: From landed.
Cite ==
    /\ ~cites
    /\ FromAnyGoal \/ landed
    /\ cites' = TRUE
    /\ citedState' = landed
    /\ UNCHANGED <<stop, landed, landStop, landRole, lands, afterLand, pending>>

Next ==
    \/ Place \/ Run \/ Select \/ CloneWrites \/ Merge \/ Cite
    \/ \E r \in Roles : Land(r)

Spec == Init /\ [][Next]_vars

\* What the state block says: who G waits for, and the owner's acts.
Waits == IF ~landed /\ stop # "none" THEN "owner" ELSE "none"
OwnerActs == IF LandAllowed("owner") THEN {"land"} ELSE {}

\* line 1: only the owner lands, only at a stop, and the land names it
OnlyOwnerLands == landed => landRole = "owner"
LandAtAStop == landed => landStop # "none"
OneLand == lands <= 1

\* line 4: the select waits for the owner, who may land
SelectWaitsForOwner == (stop = "selected" /\ ~landed) => (Waits = "owner" /\ "land" \in OwnerActs)

\* line 5: no write after the land, also not from a merged log
NoWriteAfterLand == afterLand = 0

\* line 6: --from cites only a landed Goal
CitesOnlyLanded == cites => citedState

\* line 8: a selected Goal that is not landed takes no place, run or select
SelectedRefuses == (stop = "selected" /\ ~landed) => ~ENABLED (Place \/ Run \/ Select)

\* the landed Goal keeps the stop it ended
LandedKeepsItsStop == landed => landStop = stop
=============================================================================
