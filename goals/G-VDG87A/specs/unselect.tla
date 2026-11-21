--------------------------- MODULE unselect ---------------------------
(***************************************************************************)
(* The owner's unselect, Goal G-VDG87A (#246). One Goal: runs record      *)
(* rows, the grill places example versions, a select records a row, the   *)
(* owner's unselect puts a selected Goal back to open, and the owner's    *)
(* land ends it. The latest row stands for every row that matters: a run  *)
(* on a new commit makes a new latest row, a run on the latest row reruns *)
(* it.                                                                     *)
(*                                                                         *)
(* The rules come from the Goal's lines:                                   *)
(*   line 1  an unselect opens the Goal again;                             *)
(*   line 2  the budget starts again, the old row keeps its live cells;    *)
(*   line 3  unselect is refused on a landed Goal, on a Goal that is not   *)
(*           selected, and for anyone but the owner who added the Goal;    *)
(*   line 4  after an unselect, the runner selects only once a run records *)
(*           a cell whose version, or whose row, was not run before it, and *)
(*           status prints the runner no select line until then; the owner *)
(*           selects at any time;                                          *)
(*   line 5  a Goal never unselected selects as before.                    *)
(*                                                                         *)
(* Each switch is a cheap potential Hypothesis; the check config sets      *)
(* every switch FALSE.                                                     *)
(***************************************************************************)
EXTENDS Naturals

CONSTANTS
    KeepsSelected,         \* the unselect is recorded, the Goal still reads selected
    NoBudgetReset,         \* the unselect does not move the budget's boundary
    DropsCells,            \* the unselect retracts the old row's cells
    UnselectsLanded,       \* a landed Goal can be unselected
    UnselectsUnselected,   \* a Goal that is not selected can be unselected
    AnyoneUnselects,       \* any actor or role can unselect
    ReselectsAtOnce,       \* the runner may reselect with no new cell
    RerunCounts,           \* a rerun of the same version counts as a new cell
    OwnerOnlyAfterUnselect, \* after any unselect only the owner selects
    StatusOffersSelect,     \* status prints the runner's select line for any candidate
    PairCounts              \* an old version run on an old row that never had it counts

Who == {"owner", "other"}
Budget == 3

VARIABLES
    st,          \* "open", "selected" or "landed"
    cand,        \* the latest row is met on every current example version
    pending,     \* a placed version has no cell on the latest row yet
    used,        \* runs counted since the budget's boundary
    since,       \* runs on new rows since the latest unselect
    unselected,  \* the Goal has an unselect
    fresh,       \* the build's record that a new cell came since the latest unselect
    newCell,     \* a run really recorded a new cell since the latest unselect
    by,          \* who made the latest select: "none", "runner" or "owner"
    landedEver,  \* the Goal has a land
    last,        \* the latest act tried
    lastOK,      \* whether hyp recorded it
    lastWho,     \* who tried it
    prevSt       \* the Goal's state before it

vars == <<st, cand, pending, used, since, unselected, fresh, newCell, by, landedEver,
          last, lastOK, lastWho, prevSt>>

Init ==
    /\ st = "open" /\ cand = FALSE /\ pending = FALSE
    /\ used = 0 /\ since = 0 /\ unselected = FALSE /\ fresh = FALSE /\ newCell = FALSE
    /\ by = "none" /\ landedEver = FALSE
    /\ last = "none" /\ lastOK = TRUE /\ lastWho = "owner" /\ prevSt = "open"

Tried(a, ok, w) == last' = a /\ lastOK' = ok /\ lastWho' = w /\ prevSt' = st

(* The runner runs the line on a commit not run before: a new latest row. *)
RunNewRow ==
    /\ st = "open" /\ used < Budget
    /\ \E met \in BOOLEAN : cand' = met
    /\ pending' = FALSE
    /\ used' = used + 1
    /\ since' = IF unselected THEN since + 1 ELSE since
    /\ fresh' = TRUE /\ newCell' = TRUE
    /\ Tried("run", TRUE, "runner")
    /\ UNCHANGED <<st, unselected, by, landedEver>>

(* The runner reruns the line on the latest row. A new cell only when a   *)
(* version placed since has no cell on that row yet.                      *)
RunOldRow ==
    /\ st = "open" /\ (used > 0 \/ unselected)
    /\ IF pending
         THEN /\ \E met \in BOOLEAN : cand' = met
              /\ fresh' = TRUE /\ newCell' = TRUE
         ELSE /\ cand' = cand
              /\ fresh' = (fresh \/ RerunCounts) /\ newCell' = newCell
    /\ pending' = FALSE
    /\ Tried("rerun", TRUE, "runner")
    /\ UNCHANGED <<st, used, since, unselected, by, landedEver>>

(* The runner runs an old version on another old row that never had it,   *)
(* and that row fails: the latest row stays the one candidate. The cell   *)
(* is new as a pair, but its version and its row each ran before.        *)
RunOldVersionOnOtherRow ==
    /\ st = "open" /\ (used > 0 \/ unselected)
    /\ fresh' = (fresh \/ PairCounts) /\ newCell' = newCell
    /\ Tried("rerun", TRUE, "runner")
    /\ UNCHANGED <<st, cand, pending, used, since, unselected, by, landedEver>>

(* The grill places a new example version: no row has its cell yet.       *)
Place ==
    /\ st = "open" /\ (used > 0 \/ unselected) /\ ~pending
    /\ pending' = TRUE /\ cand' = FALSE
    /\ Tried("place", TRUE, "grill")
    /\ UNCHANGED <<st, used, since, unselected, fresh, newCell, by, landedEver>>

RunnerMay ==
    /\ st = "open" /\ cand
    /\ ~unselected \/ fresh \/ ReselectsAtOnce
    /\ ~(OwnerOnlyAfterUnselect /\ unselected)

(* hyp status prints the runner's select line.                            *)
Offered ==
    IF StatusOffersSelect THEN st = "open" /\ cand ELSE RunnerMay

(* The runner asks hyp to select the latest row.                          *)
RunnerSelect ==
    /\ st \in {"open", "selected"}
    /\ IF RunnerMay
         THEN /\ st' = "selected" /\ by' = "runner"
              /\ Tried("rselect", TRUE, "runner")
         ELSE /\ UNCHANGED <<st, by>>
              /\ Tried("rselect", FALSE, "runner")
    /\ UNCHANGED <<cand, pending, used, since, unselected, fresh, newCell, landedEver>>

(* The owner selects the latest row: allowed whenever it is a candidate.  *)
OwnerSelect ==
    /\ st = "open" /\ cand
    /\ st' = "selected" /\ by' = "owner"
    /\ Tried("oselect", TRUE, "owner")
    /\ UNCHANGED <<cand, pending, used, since, unselected, fresh, newCell, landedEver>>

UnselectMay(w) ==
    /\ \/ st = "selected"
       \/ st = "landed" /\ UnselectsLanded
       \/ st = "open" /\ UnselectsUnselected
    /\ w = "owner" \/ AnyoneUnselects

(* Someone runs the unselect line hyp printed to the owner.               *)
Unselect(w) ==
    /\ IF UnselectMay(w)
         THEN /\ st' = IF KeepsSelected /\ st = "selected" THEN "selected" ELSE "open"
              /\ cand' = IF DropsCells THEN FALSE ELSE cand
              /\ used' = IF NoBudgetReset THEN used ELSE 0
              /\ since' = 0
              /\ unselected' = TRUE
              /\ fresh' = FALSE /\ newCell' = FALSE
              /\ Tried("unselect", TRUE, w)
         ELSE /\ UNCHANGED <<st, cand, used, since, unselected, fresh, newCell>>
              /\ Tried("unselect", FALSE, w)
    /\ UNCHANGED <<pending, by, landedEver>>

(* The owner lands the selected Goal.                                     *)
Land ==
    /\ st = "selected"
    /\ st' = "landed" /\ landedEver' = TRUE
    /\ Tried("land", TRUE, "owner")
    /\ UNCHANGED <<cand, pending, used, since, unselected, fresh, newCell, by>>

Next ==
    \/ RunNewRow \/ RunOldRow \/ RunOldVersionOnOtherRow \/ Place
    \/ RunnerSelect \/ OwnerSelect
    \/ \E w \in Who : Unselect(w)
    \/ Land

Spec == Init /\ [][Next]_vars

-----------------------------------------------------------------------------
(* Rules. Each is an INVARIANT of unselect.cfg.                            *)

TypeOK ==
    /\ st \in {"open", "selected", "landed"}
    /\ used \in 0..Budget
    /\ since \in 0..Budget
    /\ by \in {"none", "runner", "owner"}

(* Line 1: an unselect hyp records puts the Goal back to open.            *)
TheUnselectOpensTheGoal ==
    last = "unselect" /\ lastOK => st = "open"

(* Line 2: after an unselect, the budget counts only the runs since it.   *)
TheBudgetRestartsAtTheUnselect ==
    unselected => used = since

(* Line 2: the old row keeps its live cells: it is still a candidate.     *)
TheOldRowKeepsItsCells ==
    last = "unselect" /\ lastOK => cand

(* Line 3: a landed Goal stays landed.                                    *)
ALandedGoalStaysLanded ==
    landedEver => st = "landed"

(* Line 3: only a selected Goal is unselected.                            *)
OnlyASelectedGoalIsUnselected ==
    last = "unselect" /\ lastOK => prevSt = "selected"

(* Line 3: only the owner who added the Goal unselects.                   *)
OnlyTheOwnerUnselects ==
    last = "unselect" /\ lastOK => lastWho = "owner"

(* Line 4: the runner's select after an unselect needs a new cell.        *)
TheRunnerReselectsOnlyAfterANewCell ==
    st = "selected" /\ by = "runner" /\ unselected => newCell

(* Line 4: hyp status prints the runner's select line only when the       *)
(* runner's select would be recorded.                                     *)
StatusOffersTheRunnerOnlyWhatHypRecords ==
    Offered => RunnerMay

(* Lines 4 and 5: hyp refuses the runner's select of a candidate on an    *)
(* open Goal only after an unselect with no new cell since.               *)
TheRunnerIsRefusedOnlyWithoutANewCell ==
    last = "rselect" /\ ~lastOK /\ prevSt = "open" /\ cand => unselected /\ ~newCell

=============================================================================
