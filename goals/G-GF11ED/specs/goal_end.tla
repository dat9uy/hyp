------------------------------ MODULE goal_end ------------------------------
(***************************************************************************)
(* Goal G-GF11ED: "The owner ends a Goal with land or close, and an ended  *)
(* Goal holds exactly what it held when the owner ended it, in every       *)
(* clone."                                                                 *)
(*                                                                         *)
(* Two clones share one log once they merge: clone A, where the owner      *)
(* acts and a runner runs and selects, and clone B, where a second runner  *)
(* runs a line that reaches A's log only when the logs merge. Replay       *)
(* orders events by id, which is a time, so a run B wrote before A's end   *)
(* sorts in ahead of the end.                                              *)
(*                                                                         *)
(* Variants, each FALSE in goal_end.cfg, are cheaper builds:               *)
(*   CloseAsLand        close records a land, and status names a land      *)
(*   CloseNeedsStop     close is refused until the runs stop               *)
(*   CloseAnyRole       a role other than the owner may close              *)
(*   LandedWordKept     a landed Goal still prints `landed`, not `ended`   *)
(*   NoSnapshot         an end names nothing: replay refuses only events   *)
(*                      after the end, as today                            *)
(*   SnapshotLandOnly   land names its snapshot, close does not            *)
(*   ClosedTakesWrites  a closed Goal still takes a run                    *)
(*   FromLandedOnly     --from still refuses a closed Goal                 *)
(*   LandAnyTime        land no longer waits for the runs to stop          *)
(*   OldLandsRefused    a land written before close existed, which names   *)
(*                      no snapshot, is refused at replay                  *)
(***************************************************************************)
EXTENDS Integers, Sequences

CONSTANTS
    CloseAsLand, CloseNeedsStop, CloseAnyRole, LandedWordKept, NoSnapshot,
    SnapshotLandOnly, ClosedTakesWrites, FromLandedOnly, LandAnyTime,
    OldLandsRefused

VARIABLES
    placed,     \* the grill placed the set: the Goal is open
    runA,       \* a run of A's runner is in the log
    selected,   \* A's runner selected: the runs stopped
    runB,       \* B's runner ran a line: "none", "unmerged" or "merged"
    bEarly,     \* B's run was written before the owner's end
    bSeen,      \* B's run was merged when the owner ended: the end names it
    endBy,      \* "none", "land" or "close": the end the log holds
    rowsAtEnd,  \* the rows the owner saw when ending (-1 before)
    oldLand,    \* a land written before close existed was replayed
    out         \* what the last command printed

vars == <<placed, runA, selected, runB, bEarly, bSeen, endBy, rowsAtEnd, oldLand, out>>

Ended == endBy # "none"

\* An end that names its snapshot, so that replay refuses what it does not name.
Snapshots(e) ==
    ~NoSnapshot /\ ~(SnapshotLandOnly /\ e = "close")

\* B's run counts at replay once merged: before any end, or written before an
\* end that names no snapshot.
BCounts ==
    runB = "merged" /\ (~Ended \/ bSeen \/ (bEarly /\ ~Snapshots(endBy)))

Rows == (IF runA THEN 1 ELSE 0) + (IF BCounts THEN 1 ELSE 0)

StateWord ==
    IF endBy = "land" THEN (IF LandedWordKept THEN "landed" ELSE "ended")
    ELSE IF endBy = "close" THEN "ended"
    ELSE IF selected THEN "selected"
    ELSE IF placed THEN "open" ELSE "written"

EndLine ==
    IF endBy = "land" THEN "landed: by owner"
    ELSE IF endBy = "close" THEN (IF CloseAsLand THEN "landed: by owner" ELSE "closed: by owner")
    ELSE ""

Status(cmd) == [cmd |-> cmd, by |-> "", state |-> StateWord, endLine |-> EndLine,
                rows |-> Rows, refused |-> ""]

Refused(cmd, guard) == [cmd |-> cmd, by |-> "", state |-> "", endLine |-> "", rows |-> -1,
                        refused |-> guard]

Init ==
    /\ placed = FALSE /\ runA = FALSE /\ selected = FALSE
    /\ runB = "none" /\ bEarly = FALSE /\ bSeen = FALSE
    /\ endBy = "none" /\ rowsAtEnd = -1 /\ oldLand = FALSE
    /\ out = [cmd |-> "none", by |-> "", state |-> "", endLine |-> "", rows |-> -1, refused |-> ""]

\* The grill places the example set.
GrillPlaces ==
    /\ ~placed /\ ~Ended
    /\ placed' = TRUE
    /\ out' = [Status("place") EXCEPT !.state = "open"]
    /\ UNCHANGED <<runA, selected, runB, bEarly, endBy, rowsAtEnd, oldLand, bSeen>>

\* A's runner runs a line, or is refused on an ended Goal.
RunnerRunsA ==
    /\ placed /\ ~runA
    /\ IF Ended /\ ~(ClosedTakesWrites /\ endBy = "close")
       THEN /\ out' = Refused("run", "Not ended")
            /\ UNCHANGED runA
       ELSE /\ runA' = TRUE
            /\ out' = [Status("run") EXCEPT !.rows = Rows + 1]
    /\ UNCHANGED <<placed, selected, runB, bEarly, endBy, rowsAtEnd, oldLand, bSeen>>

\* A's runner selects its candidate: the runs stop.
RunnerSelects ==
    /\ runA /\ ~selected /\ ~Ended
    /\ selected' = TRUE
    /\ out' = [Status("select") EXCEPT !.state = "selected"]
    /\ UNCHANGED <<placed, runA, runB, bEarly, endBy, rowsAtEnd, oldLand, bSeen>>

\* B's runner runs a line in its own clone, before or after A's end.
RunnerRunsB ==
    /\ placed /\ runB = "none" /\ ~selected
    /\ runB' = "unmerged"
    /\ bEarly' = ~Ended
    /\ out' = [cmd |-> "run in B", by |-> "", state |-> "", endLine |-> "", rows |-> -1, refused |-> ""]
    /\ UNCHANGED <<placed, runA, selected, endBy, rowsAtEnd, oldLand, bSeen>>

\* The logs of A and B merge.
LogsMerge ==
    /\ runB = "unmerged"
    /\ runB' = "merged"
    /\ out' = [cmd |-> "merge", by |-> "", state |-> "", endLine |-> "", rows |-> -1, refused |-> ""]
    /\ UNCHANGED <<placed, runA, selected, bEarly, endBy, rowsAtEnd, oldLand, bSeen>>

\* The owner, or another role, closes the Goal.
Closes(role) ==
    /\ ~Ended
    /\ IF role # "owner" /\ ~CloseAnyRole
       THEN /\ out' = Refused("close", "Owner closes")
            /\ UNCHANGED <<endBy, rowsAtEnd>>
       ELSE IF CloseNeedsStop /\ ~selected
       THEN /\ out' = Refused("close", "Stop")
            /\ UNCHANGED <<endBy, rowsAtEnd>>
       ELSE /\ endBy' = IF CloseAsLand THEN "land" ELSE "close"
            /\ rowsAtEnd' = Rows
            /\ out' = [cmd |-> "close", by |-> role, state |-> "ended",
                       endLine |-> (IF CloseAsLand THEN "landed: by " ELSE "closed: by ") \o role,
                       rows |-> Rows, refused |-> ""]
    /\ UNCHANGED <<placed, runA, selected, runB, bEarly, oldLand>>
    /\ bSeen' = IF Ended' THEN (runB = "merged") ELSE bSeen

\* The owner lands the Goal: allowed once the runs stop.
OwnerLands ==
    /\ ~Ended
    /\ IF ~selected /\ ~LandAnyTime
       THEN /\ out' = Refused("land", "Stop")
            /\ UNCHANGED <<endBy, rowsAtEnd>>
       ELSE /\ endBy' = "land"
            /\ rowsAtEnd' = Rows
            /\ out' = [cmd |-> "land", by |-> "owner", state |-> IF LandedWordKept THEN "landed" ELSE "ended",
                       endLine |-> "landed: by owner", rows |-> Rows, refused |-> ""]
    /\ UNCHANGED <<placed, runA, selected, runB, bEarly, oldLand>>
    /\ bSeen' = IF Ended' THEN (runB = "merged") ELSE bSeen

\* A later Goal cites this one with --from.
AddFrom ==
    /\ out' = IF endBy = "land" \/ (endBy = "close" /\ ~FromLandedOnly)
              THEN [cmd |-> "add goal --from", by |-> "", state |-> "written", endLine |-> "",
                    rows |-> -1, refused |-> ""]
              ELSE Refused("add goal --from", "From ended")
    /\ UNCHANGED <<placed, runA, selected, runB, bEarly, endBy, rowsAtEnd, oldLand, bSeen>>

\* A log written before close existed, holding a land that names no
\* snapshot, is replayed.
OldLogReplays ==
    /\ ~oldLand
    /\ oldLand' = TRUE
    /\ out' = IF OldLandsRefused
              THEN Refused("status, old log", "refused at replay")
              ELSE [cmd |-> "status, old log", by |-> "", state |-> "ended",
                    endLine |-> "landed: by owner", rows |-> 1, refused |-> ""]
    /\ UNCHANGED <<placed, runA, selected, runB, bEarly, endBy, rowsAtEnd, bSeen>>

\* Anyone runs `hyp status G` in clone A.
ShowStatus ==
    /\ out' = Status("status G")
    /\ UNCHANGED <<placed, runA, selected, runB, bEarly, endBy, rowsAtEnd, oldLand, bSeen>>

Next ==
    \/ GrillPlaces
    \/ RunnerRunsA
    \/ RunnerSelects
    \/ RunnerRunsB
    \/ LogsMerge
    \/ \E r \in {"owner", "runner"} : Closes(r)
    \/ OwnerLands
    \/ AddFrom
    \/ OldLogReplays
    \/ ShowStatus

Spec == Init /\ [][Next]_vars

----------------------------------------------------------------------------
\* source: model only
TypeOK ==
    /\ placed \in BOOLEAN /\ runA \in BOOLEAN /\ selected \in BOOLEAN
    /\ runB \in {"none", "unmerged", "merged"}
    /\ endBy \in {"none", "land", "close"}

\* source: line-1
CloseNamesClose ==
    (out.cmd = "close" /\ out.refused = "") => out.endLine = "closed: by " \o out.by

\* source: line-1
OnlyOwnerCloses == (out.cmd = "close" /\ out.refused = "") => out.by = "owner"

\* source: owner decision 9
CloseAnyTime == out.cmd = "close" => out.refused # "Stop"

\* source: line-2
LandShowsEnded ==
    (out.cmd \in {"land", "status G"} /\ endBy = "land") => out.state = "ended"

\* source: line-3
SnapshotFixed == (Ended /\ out.cmd = "status G") => out.rows = rowsAtEnd

\* source: line-4
ClosedRefusesRun == (out.cmd = "run" /\ endBy = "close") => out.refused # ""

\* source: line-5
FromAcceptsClosed ==
    (out.cmd = "add goal --from" /\ endBy = "close") => out.refused = ""

\* source: line-6
LandWaitsForStop == (out.cmd = "land" /\ out.refused = "") => selected

\* source: line-7
OldLogReplaysClean == out.cmd = "status, old log" => out.refused = ""
=============================================================================
