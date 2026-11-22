---------------------------- MODULE place_set ----------------------------
(***************************************************************************)
(* The grill places a Goal's whole example set, Goal G-SBFXAY: "The grill *)
(* places a Goal's whole example set with one `hyp place G`, and the set   *)
(* counts only when every example agrees with its line about H0."          *)
(*                                                                         *)
(* The grill edits the examples under goals/G/line-*/ or the model under   *)
(* goals/G/specs/, then runs `hyp place G`. The set it sees has a set of   *)
(* causes to refuse: an example disagrees with its line about H0, a line   *)
(* holds no example, an example placed before is missing, an example is    *)
(* not committed, a new version changes more than floors and ceilings.     *)
(* A set with no cause and a changed example is placed whole; any other    *)
(* set is refused, and the refusal names every cause, or "unchanged".      *)
(* The runner runs once the Goal is open; a select ends the grill's acts.  *)
(*                                                                         *)
(* Variants, each FALSE in place_set.cfg, are cheaper builds:              *)
(*   PartialPlace      places the agreeing examples and refuses the rest   *)
(*   FirstCauseOnly    names only the first cause                          *)
(*   AllowEmptyLine    places a set with an empty line; the Goal stays     *)
(*                     written                                             *)
(*   DropMissing       drops a missing example from the set                *)
(*   RePlaceUnchanged  places an unchanged set and restarts the budget     *)
(*   SpecsInSet        counts a model edit as a change of the set          *)
(*   PerLineOffer      keeps one place line per line in the grill's block  *)
(* Left out: line-6 (old Goals keep their status) is about the stored log, *)
(* not a step. Committed and Adds only are checked as before, and no cheap *)
(* build of a set place drops them: line-3's examples check them, and this *)
(* model holds them out of Causes to keep the chart small.                 *)
(***************************************************************************)
EXTENDS Naturals, FiniteSets

CONSTANTS
    PartialPlace,
    FirstCauseOnly,
    AllowEmptyLine,
    DropMissing,
    RePlaceUnchanged,
    SpecsInSet,
    PerLineOffer

Causes == {"disagrees", "empty line", "missing"}

VARIABLES
    open,        \* the Goal is open: a whole set is placed with every line full
    exChanged,   \* an example changed since the last place
    specEdited,  \* the model changed since the last place
    runs,        \* 0 or 1: a Hypothesis ran since the budget's boundary
    selected,    \* the Goal has a select
    last,        \* what the last place did: "none", "placed", "part", "refused"
    seen,        \* the causes of the set the last place saw
    changedAt,   \* an example had changed when the last place ran
    named,       \* what the last refusal named
    partial,     \* a part of a set was placed
    dropped,     \* a placed example left the set
    block        \* the grill's block: "one", "per line" or "none"

vars == <<open, exChanged, specEdited, runs, selected, last, seen, changedAt,
          named, partial, dropped, block>>

Offer(sel) == IF sel THEN "none" ELSE IF PerLineOffer THEN "per line" ELSE "one"

Init ==
    /\ open = FALSE
    /\ exChanged = TRUE
    /\ specEdited = FALSE
    /\ runs = 0
    /\ selected = FALSE
    /\ last = "none"
    /\ seen = {}
    /\ changedAt = FALSE
    /\ named = {}
    /\ partial = FALSE
    /\ dropped = FALSE
    /\ block = Offer(FALSE)

\* The grill writes, fixes or commits an example.
GrillEditsExamples ==
    /\ ~selected
    /\ exChanged' = TRUE
    /\ block' = Offer(selected)
    /\ UNCHANGED <<open, specEdited, runs, selected, last, seen, changedAt,
                   named, partial, dropped>>

\* The grill edits only the model.
GrillEditsModel ==
    /\ ~selected
    /\ specEdited' = TRUE
    /\ block' = Offer(selected)
    /\ UNCHANGED <<open, exChanged, runs, selected, last, seen, changedAt,
                   named, partial, dropped>>

\* The grill runs `hyp place G` on a set with the causes cs.
GrillPlaces ==
    /\ ~selected
    /\ \E cs \in SUBSET Causes :
        LET ignored   == (IF AllowEmptyLine THEN {"empty line"} ELSE {})
                         \cup (IF DropMissing THEN {"missing"} ELSE {})
            eff       == cs \ ignored
            changed   == exChanged \/ (SpecsInSet /\ specEdited)
            accept    == RePlaceUnchanged \/ changed
            whyNot    == IF cs = {} /\ ~exChanged THEN {"unchanged"} ELSE cs
            refuse(n) == /\ last' = "refused"
                         /\ named' = n
                         /\ UNCHANGED <<open, exChanged, specEdited, runs, partial, dropped>>
        IN
        /\ seen' = cs
        /\ changedAt' = exChanged
        /\ IF accept /\ eff = {}
           THEN /\ last' = "placed"
                /\ open' = ("empty line" \notin cs)
                /\ runs' = 0
                /\ exChanged' = FALSE
                /\ specEdited' = FALSE
                /\ dropped' = (dropped \/ "missing" \in cs)
                /\ named' = {}
                /\ UNCHANGED partial
           ELSE IF PartialPlace /\ accept /\ eff = {"disagrees"}
           THEN /\ last' = "part"
                /\ partial' = TRUE
                /\ runs' = 0
                /\ named' = {"disagrees"}
                /\ UNCHANGED <<open, exChanged, specEdited, dropped>>
           ELSE IF FirstCauseOnly /\ Cardinality(whyNot) > 1
           THEN \E c \in whyNot : refuse({c})
           ELSE refuse(whyNot)
    /\ block' = Offer(selected)
    /\ UNCHANGED selected

\* The runner runs a Hypothesis on an open Goal.
RunnerRuns ==
    /\ open /\ ~selected
    /\ runs' = 1
    /\ block' = Offer(selected)
    /\ UNCHANGED <<open, exChanged, specEdited, selected, last, seen, changedAt,
                   named, partial, dropped>>

\* A candidate is selected.
Selects ==
    /\ open /\ ~selected /\ runs = 1
    /\ selected' = TRUE
    /\ block' = Offer(TRUE)
    /\ UNCHANGED <<open, exChanged, specEdited, runs, last, seen, changedAt,
                   named, partial, dropped>>

\* Anyone reads status; it changes nothing.
Status ==
    /\ block' = Offer(selected)
    /\ UNCHANGED <<open, exChanged, specEdited, runs, selected, last, seen,
                   changedAt, named, partial, dropped>>

Next ==
    \/ Status
    \/ GrillEditsExamples
    \/ GrillEditsModel
    \/ GrillPlaces
    \/ RunnerRuns
    \/ Selects

Spec == Init /\ [][Next]_vars

\* source: model only
TypeOK ==
    /\ open \in BOOLEAN /\ exChanged \in BOOLEAN /\ specEdited \in BOOLEAN
    /\ runs \in {0, 1} /\ selected \in BOOLEAN
    /\ last \in {"none", "placed", "part", "refused"}
    /\ seen \subseteq Causes /\ changedAt \in BOOLEAN
    /\ named \subseteq Causes \cup {"unchanged"}
    /\ partial \in BOOLEAN /\ dropped \in BOOLEAN
    /\ block \in {"one", "per line", "none"}

\* source: line-1
WholeSetOrNone == ~partial

\* source: line-1
PlacedSetAgrees == last \in {"placed", "part"} => "disagrees" \notin seen

\* What a refusal must name: every cause of the set, or "unchanged".
Due == IF seen = {} /\ ~changedAt THEN {"unchanged"} ELSE seen

\* source: line-2
NamesEveryCause == last = "refused" => named = Due

\* source: line-3
NoEmptyLinePlaced == last = "placed" => "empty line" \notin seen

\* source: line-3
NoExampleLeaves == ~dropped

\* source: line-4
OnlyChangedSetPlaced == last \in {"placed", "part"} => changedAt

\* source: line-5
OnePlaceLine == block = IF selected THEN "none" ELSE "one"

=============================================================================
