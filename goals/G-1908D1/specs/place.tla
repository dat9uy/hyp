------------------------------ MODULE place ------------------------------
(***************************************************************************)
(* The grill's place lines in the state block, Goal G-1908D1: "While a     *)
(* line has no example, status offers a further example on a line that     *)
(* has one".                                                               *)
(*                                                                         *)
(* A Goal has up to three lines. The grill places examples; a Draft        *)
(* replace by the owner may change the lines and move H0, which leaves     *)
(* every placed example without a live H0 cell; the runner runs a line on  *)
(* H0 to give it one again. The Goal is open when every current line holds *)
(* an example with a live H0 cell. While open, the runner's runs may reach *)
(* the stops "two candidates" or "budget used", and a candidate may be     *)
(* selected. After every command (place, a refused place, status, run,     *)
(* add goal, select) hyp prints the state block; block is what the grill's *)
(* part of it offered: line keys, or Blank (0) for the blank             *)
(* `hyp place G <line>/<example>`.                                         *)
(*                                                                         *)
(* The rules are the Goal's line and the owner's decision 3:               *)
(*   line 1        while the Goal is not selected, every place the guards  *)
(*                 allow is offered: one on each current line              *)
(*   decision 3    one ready place line per Goal line, naming its key, in  *)
(*                 every state that allows a place: written and open; none *)
(*                 once selected                                           *)
(* Decisions 1 (the words) and 2 (one gap line) are facts of the Goal's    *)
(* record, not of what hyp does; this model holds no rule for them.        *)
(*                                                                         *)
(* Switches. OldH0 is how H0 worked: a place line only for lines with no   *)
(* example, and the blank form once every line held one. BlankBeside,      *)
(* PlaceWhenSelected, OfferGoneLines and OnlyWhenWritten are cheaper or    *)
(* wrong builds. place.cfg sets all FALSE.                                 *)
(* Left out: the guards Committed, Adds only and Kind, which refuse one    *)
(* example folder and never a line; the runner's and owner's blocks; and   *)
(* the line kinds, which no rule here reads.                               *)
(***************************************************************************)
EXTENDS Naturals, FiniteSets

CONSTANTS
    OldH0,              \* H0: offer only empty lines, then the blank form
    BlankBeside,        \* a build that keeps the blank line beside the ready lines
    PlaceWhenSelected,  \* a build that offers a place line in every state
    OfferGoneLines,     \* a build that offers every line the Goal ever had
    OnlyWhenWritten     \* a build that offers place lines only while written

AllLines == {1, 2, 3}
Blank    == 0   \* the blank `hyp place G <line>/<example>`

VARIABLES
    lines,     \* the current lines of the Goal
    ever,      \* every line the Goal has had
    cell,      \* cell[l]: "none" (no example), "live" (placed, live H0 cell),
               \* "stale" (placed, H0 moved by a Draft replace)
    stop,      \* "none", "two candidates", "budget used"
    selected,  \* TRUE once the Goal has a select
    block      \* what the grill's block offered after the last command

vars == <<lines, ever, cell, stop, selected, block>>

Open    == ~selected /\ \A l \in lines : cell[l] = "live"
Written == ~selected /\ ~Open
State   == IF selected THEN "selected" ELSE IF Open THEN "open" ELSE "written"

\* What this build offers the grill in a state.
Offer(ls, ev, c, sel) ==
    LET placedAll == \A l \in ls : c[l] # "none"
        open      == ~sel /\ \A l \in ls : c[l] = "live"
    IN
    IF sel /\ ~PlaceWhenSelected THEN {}
    ELSE IF OldH0
        THEN (IF placedAll THEN {Blank} ELSE {l \in ls : c[l] = "none"})
    ELSE IF OnlyWhenWritten /\ open THEN {}
    ELSE (IF OfferGoneLines THEN ev ELSE ls)
         \cup (IF BlankBeside /\ open THEN {Blank} ELSE {})

Print(ls, ev, c, sel) == block' = Offer(ls, ev, c, sel)

Init ==
    /\ lines \in {{1, 2, 3}, {1, 2}}
    /\ ever = lines
    /\ cell = [l \in AllLines |-> "none"]
    /\ stop = "none"
    /\ selected = FALSE
    /\ block = Offer(lines, ever, cell, FALSE)

\* The grill places an example on a current line; Kind runs it on H0, so it
\* has a live H0 cell. A further example on a placed line is the same act.
GrillPlaces ==
    /\ ~selected
    /\ \E l \in lines :
        /\ cell' = [cell EXCEPT ![l] = "live"]
        /\ Print(lines, ever, cell', selected)
    /\ UNCHANGED <<lines, ever, stop, selected>>

\* A place refused by a guard (Adds only, Committed, Kind, Selected) changes
\* nothing and prints the block again.
PlaceRefused ==
    /\ Print(lines, ever, cell, selected)
    /\ UNCHANGED <<lines, ever, cell, stop, selected>>

\* The owner replaces the written Goal: new lines, and H0 moved, so every
\* placed example has no live H0 cell. A placement on a gone line is history.
OwnerReplaces ==
    /\ Written
    /\ \E ls \in SUBSET AllLines \ {{}} :
        /\ lines' = ls
        /\ ever' = ever \cup ls
        /\ cell' = [l \in AllLines |-> IF cell[l] = "none" THEN "none" ELSE "stale"]
        /\ Print(ls, ever', cell', selected)
    /\ UNCHANGED <<stop, selected>>

\* The runner runs a line on H0, which gives its placed examples a live cell.
RunnerRunsH0 ==
    /\ Written
    /\ \E l \in lines :
        /\ cell[l] = "stale"
        /\ cell' = [cell EXCEPT ![l] = "live"]
        /\ Print(lines, ever, cell', selected)
    /\ UNCHANGED <<lines, ever, stop, selected>>

\* The runner runs a candidate on an open Goal; it may reach a stop.
RunnerRuns ==
    /\ Open
    /\ stop = "none"
    /\ stop' \in {"none", "two candidates", "budget used"}
    /\ Print(lines, ever, cell, selected)
    /\ UNCHANGED <<lines, ever, cell, selected>>

\* The runner (one candidate) or the owner (two) selects.
Selects ==
    /\ Open
    /\ stop \in {"none", "two candidates"}
    /\ selected' = TRUE
    /\ Print(lines, ever, cell, TRUE)
    /\ UNCHANGED <<lines, ever, cell, stop>>

\* Anyone reads status.
Status ==
    /\ Print(lines, ever, cell, selected)
    /\ UNCHANGED <<lines, ever, cell, stop, selected>>

Next ==
    \/ GrillPlaces
    \/ PlaceRefused
    \/ OwnerReplaces
    \/ RunnerRunsH0
    \/ RunnerRuns
    \/ Selects
    \/ Status

Spec == Init /\ [][Next]_vars

TypeOK ==
    /\ lines \subseteq AllLines /\ lines # {}
    /\ ever \subseteq AllLines /\ lines \subseteq ever
    /\ cell \in [AllLines -> {"none", "live", "stale"}]
    /\ stop \in {"none", "two candidates", "budget used"}
    /\ selected \in BOOLEAN
    /\ block \subseteq AllLines \cup {Blank}

\* line 1: every place the guards allow is offered, a further example on a
\* line that holds one included.
EveryAllowedPlaceOffered == ~selected => lines \subseteq block

\* decision 3: one ready place line per Goal line while written ...
OneLinePerLineWhileWritten == Written => block = lines

\* ... and while open, at every stop.
OneLinePerLineWhileOpen == Open => block = lines

\* decision 3: each line names its key; the blank form is never printed.
EveryLineNamesItsKey == Blank \notin block

\* decision 3: no state but written and open allows a place.
NothingOnceSelected == selected => block = {}

=============================================================================
