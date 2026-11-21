------------------------------ MODULE report ------------------------------
(***************************************************************************)
(* The statechart tool's report and the grill's use of it, Goal G-SCAXYW:  *)
(* "The statechart tool, not the grill, says which rules each variant      *)
(* breaks and which examples each variant would meet; the grill tries what *)
(* the tool predicts, and only a tried variant that meets every example    *)
(* reaches the owner, as one plain question."                              *)
(*                                                                         *)
(* One small Goal model goes through the tool and the grill: rules A, B    *)
(* and C, variants V1 and V2, examples e1 and e2. V1 breaks A, V2 breaks A *)
(* and B, and no variant breaks C. Rule A names line-1 as its source, B    *)
(* line-2, and C names none. Each example's own walk prints its floor.     *)
(* Under V1 the model predicts e1 failed and e2 met; under V2 both met.    *)
(* hyp try on V1's patch may agree with that prediction on e1 or not       *)
(* (disagree); on every other cell it prints met. When it disagrees, the   *)
(* grill fixes the model, which then predicts V1 met on e1 too.            *)
(*                                                                         *)
(* The tool's steps: it writes the variant configs, matches each floor to  *)
(* a walk, predicts each cell, and, after the grill's tries, writes the    *)
(* report. The grill's steps: it tries each variant predicted to meet      *)
(* every example and one predicted-fail example of each other variant,     *)
(* fixes the model where a try differs, and asks the owner about each      *)
(* tried variant that met every example. Beside them, a select reads the   *)
(* spec folder in one of three layouts: configs under variants/, under the *)
(* old switches/, or no model.                                             *)
(*                                                                         *)
(* Each variant is a cheap build that a runner could ship instead. The     *)
(* check config report.cfg sets all FALSE.                                 *)
(* Left out: the exact print of a view, TLC itself, the Markdown form of   *)
(* the report, and the owner's answer to a question.                      *)
(***************************************************************************)
EXTENDS Naturals, FiniteSets

CONSTANTS
    ConfigListsAll,     \* a variant config lists every rule the check config checks
    DropUnbroken,       \* a rule no variant breaks is left out without a word
    LastStateOnly,      \* a floor is matched only to the last state's print
    AnyWalk,            \* a cell is met when any walk prints the floor
    NoSeparatedRow,     \* the separated-by row is left empty
    TriedReplaces,      \* a tried verdict replaces the predicted one in the grid
    NoSourceList,       \* rules with no source are not listed
    SkipSpotCheck,      \* the grill tries only variants predicted to meet every example
    FixExample,         \* where a try differs, the grill drops the example
    AskFromPrediction,  \* the grill asks the owner from the prediction, untried
    OnlySwitchesDir,    \* select reads variant configs only under switches/
    OnlyVariantsDir,    \* select reads variant configs only under variants/
    AskEveryVariant,    \* the grill asks the owner about every variant
    RequireModel        \* select refuses a commit whose spec folder holds no model

E      == {"e1", "e2"}
V      == {"V1", "V2"}
Rules  == {"A", "B", "C"}
None   == "-"
Phases == {"start", "configs", "walks", "cells", "tried", "report", "asked"}
Layouts == {"variants", "switches", "none"}

Source == [r \in Rules |-> IF r = "A" THEN "line-1" ELSE IF r = "B" THEN "line-2" ELSE "none"]
Breaks == [v \in V |-> IF v = "V1" THEN {"A"} ELSE {"A", "B"}]

\* What the model predicts on an example's own walk, before and after a fix.
OwnPred(e, v, fx) == IF e = "e1" /\ v = "V1" /\ ~fx THEN "fail" ELSE "met"
\* What hyp try prints on a variant's patch.
Code(e, v, dis) == IF e = "e1" /\ v = "V1" /\ ~dis THEN "fail" ELSE "met"

EmptyGrid == [e \in E |-> [v \in V |-> None]]

VARIABLES
    phase,     \* where the tool and the grill are
    disagree,  \* hyp try on V1's patch prints met on e1, unlike the model
    fixed,     \* the grill fixed the model
    exs,       \* the examples of the Goal
    cfg,       \* cfg[v]: the rules variant v's config lists
    notShown,  \* the rules printed as not shown to fail
    walk,      \* walk[e]: "found", "none", or None before the match
    pred,      \* pred[e][v]: the predicted cell
    sep,       \* sep[v]: the separated-by row
    tried,     \* tried[e][v]: what hyp try printed, or None
    shown,     \* shown[e][v]: the predicted cell as the grid shows it
    diff,      \* the cells named as differing
    noSource,  \* the rules listed with no source
    handoff,   \* the hand-off says the model was fixed
    asked,     \* the variants the owner is asked about
    layout,    \* the spec folder's layout at select
    sel        \* the select's result, or None

vars == <<phase, disagree, fixed, exs, cfg, notShown, walk, pred, sep, tried,
          shown, diff, noSource, handoff, asked, layout, sel>>
tool == <<phase, disagree, fixed, exs, cfg, notShown, walk, pred, sep, tried,
          shown, diff, noSource, handoff, asked>>

Init ==
    /\ phase = "start"
    /\ disagree \in BOOLEAN
    /\ fixed = FALSE
    /\ exs = E
    /\ cfg = [v \in V |-> {}]
    /\ notShown = {}
    /\ walk = [e \in E |-> None]
    /\ pred = EmptyGrid
    /\ sep = [v \in V |-> {}]
    /\ tried = EmptyGrid
    /\ shown = EmptyGrid
    /\ diff = {}
    /\ noSource = {}
    /\ handoff = FALSE
    /\ asked = {}
    /\ layout \in Layouts
    /\ sel = None

\* The tool runs each rule alone under each variant and writes the configs.
ToolWritesConfigs ==
    /\ phase = "start"
    /\ cfg' = IF ConfigListsAll THEN [v \in V |-> Rules] ELSE Breaks
    /\ notShown' = IF DropUnbroken THEN {} ELSE Rules \ UNION {Breaks[v] : v \in V}
    /\ phase' = "configs"
    /\ UNCHANGED <<disagree, fixed, exs, walk, pred, sep, tried, shown, diff, noSource, handoff, asked, layout, sel>>

\* The tool matches each example's floor to a walk. e2's floor holds the
\* prints of several steps, so only the whole walk gives it.
ToolMatchesWalks ==
    /\ phase = "configs"
    /\ walk' = [e \in E |-> IF LastStateOnly /\ e = "e2" THEN "none" ELSE "found"]
    /\ phase' = "walks"
    /\ UNCHANGED <<disagree, fixed, exs, cfg, notShown, pred, sep, tried, shown, diff, noSource, handoff, asked, layout, sel>>

\* The tool predicts each cell from the model alone.
ToolPredicts ==
    /\ phase = "walks"
    /\ pred' = [e \in E |-> [v \in V |->
           IF e \notin exs \/ walk[e] # "found" THEN None
           ELSE IF AnyWalk THEN "met" ELSE OwnPred(e, v, fixed)]]
    /\ sep' = IF NoSeparatedRow THEN [v \in V |-> {}]
              ELSE [v \in V |-> {e \in exs : pred'[e][v] = "fail"}]
    /\ asked' = IF AskFromPrediction THEN {v \in V : \A e \in exs : pred'[e][v] = "met"} ELSE asked
    /\ phase' = "cells"
    /\ UNCHANGED <<disagree, fixed, exs, cfg, notShown, walk, tried, shown, diff, noSource, handoff, layout, sel>>

\* The grill tries each variant predicted to meet every example on every
\* example, and one predicted-fail example of each other variant.
MetAll(v) == \A e \in exs : pred[e][v] = "met"
Spot(v)   == CHOOSE e \in exs : pred[e][v] = "fail"
ToTry ==
    {<<e, v>> \in exs \X V : MetAll(v)}
    \cup (IF SkipSpotCheck THEN {}
          ELSE {<<Spot(v), v>> : v \in {w \in V : \E e \in exs : pred[e][w] = "fail"}})

GrillTries ==
    /\ phase = "cells"
    /\ tried' = [e \in E |-> [v \in V |-> IF <<e, v>> \in ToTry THEN Code(e, v, disagree) ELSE None]]
    /\ phase' = "tried"
    /\ UNCHANGED <<disagree, fixed, exs, cfg, notShown, walk, pred, sep, shown, diff, noSource, handoff, asked, layout, sel>>

\* The tool writes the report with the tried verdicts the grill gave it.
ToolReports ==
    /\ phase = "tried"
    /\ shown' = [e \in E |-> [v \in V |->
           IF TriedReplaces /\ tried[e][v] # None THEN tried[e][v] ELSE pred[e][v]]]
    /\ diff' = {<<e, v>> \in E \X V : tried[e][v] # None /\ tried[e][v] # shown'[e][v]}
    /\ noSource' = IF NoSourceList THEN {} ELSE {r \in Rules : Source[r] = "none"}
    /\ phase' = "report"
    /\ UNCHANGED <<disagree, fixed, exs, cfg, notShown, walk, pred, sep, tried, handoff, asked, layout, sel>>

\* A try differs from its prediction: the grill fixes the model and says so
\* in the hand-off, and the tool predicts again.
GrillFixes ==
    /\ phase = "report"
    /\ diff # {}
    /\ IF FixExample
         THEN /\ exs' = exs \ {d[1] : d \in diff}
              /\ UNCHANGED fixed
         ELSE /\ fixed' = TRUE
              /\ UNCHANGED exs
    /\ handoff' = TRUE
    /\ tried' = EmptyGrid
    /\ diff' = {}
    /\ phase' = "walks"
    /\ UNCHANGED <<disagree, cfg, notShown, walk, pred, sep, shown, noSource, asked, layout, sel>>

\* Every try agrees: the grill asks the owner about each tried variant that
\* met every example.
GrillAsks ==
    /\ phase = "report"
    /\ diff = {}
    /\ asked' = IF AskFromPrediction THEN asked
                ELSE IF AskEveryVariant THEN V
                ELSE {v \in V : \A e \in exs : tried[e][v] = "met"}
    /\ phase' = "asked"
    /\ UNCHANGED <<disagree, fixed, exs, cfg, notShown, walk, pred, sep, tried, shown, diff, noSource, handoff, layout, sel>>

\* A select reads the spec folder. Every rule is named by a config in the
\* folder its layout uses.
Selects ==
    /\ sel = None
    /\ sel' = CASE layout = "none"     -> IF RequireModel THEN "refused" ELSE "recorded"
                [] layout = "variants" -> IF OnlySwitchesDir THEN "refused" ELSE "recorded"
                [] layout = "switches" -> IF OnlyVariantsDir THEN "refused" ELSE "recorded"
    /\ UNCHANGED <<tool, layout>>

Done ==
    /\ phase = "asked" /\ sel # None
    /\ UNCHANGED vars

Next ==
    \/ ToolWritesConfigs
    \/ ToolMatchesWalks
    \/ ToolPredicts
    \/ GrillTries
    \/ ToolReports
    \/ GrillFixes
    \/ GrillAsks
    \/ Selects
    \/ Done

Spec == Init /\ [][Next]_vars

-----------------------------------------------------------------------------
Predicted == {"cells", "tried", "report", "asked"}
Reported  == {"report", "asked"}

\* source: model only. No variant breaks it, so report.cfg leaves it out.
TypeOK ==
    /\ phase \in Phases
    /\ exs \subseteq E
    /\ cfg \in [V -> SUBSET Rules]
    /\ layout \in Layouts
    /\ sel \in {None, "recorded", "refused"}

\* source: line-1
ConfigsListExactlyWhatBreaks == phase # "start" => cfg = Breaks

\* source: line-1
UnbrokenRulePrinted == phase # "start" => notShown = {"C"}

\* source: line-2
EveryFloorHasAWalk == phase \notin {"start", "configs"} => walk = [e \in E |-> "found"]

\* source: line-3
CellsFollowTheOwnWalk ==
    phase \in Predicted => \A e \in exs, v \in V : pred[e][v] = OwnPred(e, v, fixed)

\* source: line-3
SeparatedRowShown ==
    phase \in Predicted => sep = [v \in V |-> {e \in exs : pred[e][v] = "fail"}]

\* source: line-4
TriedSitsBesidePrediction == phase \in Reported => shown = pred

\* source: line-4
DifferingCellNamed ==
    phase \in Reported =>
        diff = {<<e, v>> \in E \X V : tried[e][v] # None /\ tried[e][v] # pred[e][v]}

\* source: line-5
RuleWithNoSourceListed == phase \in Reported => noSource = {"C"}

\* source: line-6
SpotCheckTried ==
    phase \in {"tried", "report", "asked"} =>
        \A v \in V : (\E e \in exs : pred[e][v] = "fail") =>
            \E e \in exs : pred[e][v] = "fail" /\ tried[e][v] # None

\* source: line-6
ModelFixedNotExample == exs = E /\ (fixed => handoff)

\* source: line-7
OnlyTriedVariantAsked == \A v \in asked : \A e \in exs : tried[e][v] = "met"

\* source: line-7
EveryUnseparatedVariantAsked ==
    phase = "asked" => asked = {v \in V : \A e \in exs : tried[e][v] = "met"}

\* source: line-8
SelectReadsVariants == layout = "variants" /\ sel # None => sel = "recorded"

\* source: line-9
SelectReadsSwitches == layout = "switches" /\ sel # None => sel = "recorded"

\* source: line-10
NoModelSelected == layout = "none" /\ sel # None => sel = "recorded"

=============================================================================
