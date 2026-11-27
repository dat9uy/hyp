---------------------------- MODULE store ----------------------------
\* hyp's store across two branches: the acts each branch records, the
\* files those acts write, a plain git merge with no merge driver, as on
\* GitHub, and what replay calculates on the merge. The base holds Goal
\* G-ONE, added by the owner, in a store of the older form (one file per
\* actor) or of the new form (one file per event).

EXTENDS Naturals

\* AppendPerActor: an act appends a line to its actor's file, as hyp did
\* before one file per event. ReadNewOnly: replay reads only the files of
\* one event and skips the older per-actor files. ReplayBoth: replay applies
\* both of two acts that end one Goal.
CONSTANTS AppendPerActor, ReadNewOnly, ReplayBoth

VARIABLES form, added, changed, ended, phase, conflict, reads, applied, out

vars == <<form, added, changed, ended, phase, conflict, reads, applied, out>>

Branches == {"main", "side"}

\* Each branch records at most this many acts.
MaxActs == 2

StoreTypeOK ==
    /\ form \in {"old", "new"}
    /\ added \in [Branches -> 0..MaxActs]
    /\ changed \in [Branches -> 0..MaxActs]
    /\ ended \in [Branches -> BOOLEAN]
    /\ phase \in {"apart", "merged"}
    /\ conflict \in BOOLEAN
    /\ reads \in BOOLEAN
    /\ applied \in 0..2
    /\ out \in {"none", "recorded", "close exit: 0", "merge exit: 0", "merge exit: 1"}

Init ==
    /\ form \in {"old", "new"}
    /\ added = [b \in Branches |-> 0]
    /\ changed = [b \in Branches |-> 0]
    /\ ended = [b \in Branches |-> FALSE]
    /\ phase = "apart"
    /\ conflict = FALSE
    \* Replay reads the base's Goal unless it skips the base's file.
    /\ reads = (form = "new" \/ ~ReadNewOnly)
    /\ applied = 0
    /\ out = "none"

\* The owner's act on branch b writes its event: a new file of its own, or
\* with AppendPerActor a line appended to the owner's file, which the base
\* already holds, so the act changes that file.
StoreWrite(b) ==
    /\ added[b] + changed[b] < MaxActs
    /\ IF AppendPerActor
         THEN changed' = [changed EXCEPT ![b] = @ + 1] /\ UNCHANGED added
         ELSE added' = [added EXCEPT ![b] = @ + 1] /\ UNCHANGED changed

\* command: hyp add goal
StoreAddGoal ==
    \E b \in Branches :
        /\ phase = "apart"
        /\ StoreWrite(b)
        /\ out' = "recorded"
        /\ UNCHANGED <<form, ended, phase, conflict, reads, applied>>

\* command: hyp close
StoreClose ==
    \E b \in Branches :
        /\ phase = "apart"
        /\ ~ended[b]
        /\ StoreWrite(b)
        /\ ended' = [ended EXCEPT ![b] = TRUE]
        /\ out' = "close exit: 0"
        /\ UNCHANGED <<form, phase, conflict, reads, applied>>

\* The ends of G-ONE the branches recorded.
StoreEnds == (IF ended["main"] THEN 1 ELSE 0) + (IF ended["side"] THEN 1 ELSE 0)

\* A plain merge with no merge driver stops on a conflict when both branches
\* changed one file. Replay on the merge applies the first end of G-ONE and
\* refuses the later one, or with ReplayBoth applies both.
\* command: git merge
GitMerge ==
    /\ phase = "apart"
    /\ phase' = "merged"
    /\ conflict' = (changed["main"] > 0 /\ changed["side"] > 0)
    /\ applied' = IF conflict' THEN 0
                  ELSE IF ReplayBoth THEN StoreEnds
                  ELSE IF StoreEnds > 0 THEN 1 ELSE 0
    /\ out' = IF conflict' THEN "merge exit: 1" ELSE "merge exit: 0"
    /\ UNCHANGED <<form, added, changed, ended, reads>>

Next == StoreAddGoal \/ StoreClose \/ GitMerge

Spec == Init /\ [][Next]_vars

\* source: G-TWJ4PF:line-1
OneNewFilePerAct == \A b \in Branches : changed[b] = 0

\* source: G-TWJ4PF:line-2
PlainMergeNoConflict == ~conflict

\* source: G-TWJ4PF:line-4
OldStoreReplays == reads

\* source: G-TWJ4PF:line-5
LaterEndRefused == applied <= 1

=============================================================================
