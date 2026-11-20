---------------------------- MODULE select ----------------------------
(***************************************************************************)
(* The Spec guard of hyp select, Goal G-G2R3VG. A runner commits files in  *)
(* the Goal's spec folder, then asks hyp to select the commit. The folder  *)
(* can hold no file, a plain spec, or a model: a .tla module whose check   *)
(* config checks some rules, switch configs that name some of them, and a  *)
(* chart. hyp reads the files and runs no model checker.                   *)
(*                                                                         *)
(* The rules come from the Goal's lines:                                   *)
(*   line 1  select refuses a model with a rule no switch config names,    *)
(*           or with no chart, and names each such rule or the chart;      *)
(*   line 4  a spec folder with no model is selected as before;            *)
(*   and the Spec guard as before: a commit with no spec is refused.       *)
(* Line 2 (the chart in the report) and line 3 (the skill) are not states  *)
(* of the guard; the model leaves them out.                                *)
(*                                                                         *)
(* Each switch is a cheap wrong version of the guard; the check config     *)
(* sets every switch FALSE.                                                *)
(***************************************************************************)
EXTENDS FiniteSets

CONSTANTS
    AnyFilePasses,      \* the guard before this change: any file passes
    NameFirstRuleOnly,  \* the refusal names only one missing rule
    NoChartCheck,       \* a model with no chart passes
    ChartAlways,        \* a chart is needed with no model too
    EmptyFolderPasses   \* a folder with no file passes

Rules == {"r1", "r2"}

VARIABLES
    files,    \* "none", "plain" or "model": what the spec folder holds
    named,    \* the rules some switch config names
    chart,    \* the folder holds a chart
    outcome,  \* "editing", "selected" or "refused"
    names     \* what the refusal names: rules, or "chart"

vars == <<files, named, chart, outcome, names>>

Init ==
    /\ files = "none" /\ named = {} /\ chart = FALSE
    /\ outcome = "editing" /\ names = {}

Unswitched == IF files = "model" THEN Rules \ named ELSE {}

(* The runner commits files in the spec folder.                            *)
CommitPlain ==
    /\ outcome = "editing" /\ files = "none"
    /\ files' = "plain"
    /\ UNCHANGED <<named, chart, outcome, names>>

CommitModel ==
    /\ outcome = "editing" /\ files \in {"none", "plain"}
    /\ files' = "model"
    /\ UNCHANGED <<named, chart, outcome, names>>

CommitSwitch ==
    /\ outcome = "editing" /\ files = "model"
    /\ \E r \in Rules \ named : named' = named \cup {r}
    /\ UNCHANGED <<files, chart, outcome, names>>

CommitChart ==
    /\ outcome = "editing" /\ ~chart
    /\ chart' = TRUE
    /\ UNCHANGED <<files, named, outcome, names>>

(* What the guard refuses for, as each version reads the folder.           *)
NoSpec      == files = "none" /\ ~EmptyFolderPasses
RuleRefused == ~AnyFilePasses /\ Unswitched # {}
ChartNeeded == /\ ~AnyFilePasses /\ ~chart
               /\ \/ files = "model" /\ ~NoChartCheck
                  \/ ChartAlways /\ files # "none"

(* The runner asks hyp to select the commit.                               *)
Select ==
    /\ outcome = "editing"
    /\ IF NoSpec \/ RuleRefused \/ ChartNeeded
         THEN /\ outcome' = "refused"
              /\ names' = CASE NoSpec      -> {}
                            [] RuleRefused -> IF NameFirstRuleOnly
                                                THEN {CHOOSE r \in Unswitched : TRUE}
                                                ELSE Unswitched
                            [] OTHER       -> {"chart"}
         ELSE /\ outcome' = "selected" /\ names' = {}
    /\ UNCHANGED <<files, named, chart>>

(* After a refusal the runner commits again.                               *)
Again ==
    /\ outcome = "refused"
    /\ outcome' = "editing" /\ names' = {}
    /\ UNCHANGED <<files, named, chart>>

Done ==
    /\ outcome = "selected"
    /\ UNCHANGED vars

Next == CommitPlain \/ CommitModel \/ CommitSwitch \/ CommitChart \/ Select \/ Again \/ Done

Spec == Init /\ [][Next]_vars

-----------------------------------------------------------------------------
(* Rules. Each is an INVARIANT of select.cfg.                              *)

(* As before: a selected commit holds a spec.                              *)
ASelectedCommitHoldsASpec ==
    outcome = "selected" => files # "none"

(* Line 1: a selected model has a switch config for every rule.            *)
ASelectedModelHasASwitchForEveryRule ==
    outcome = "selected" /\ files = "model" => named = Rules

(* Line 1: a selected model has a chart.                                   *)
ASelectedModelHasAChart ==
    outcome = "selected" /\ files = "model" => chart

(* Line 1: a refusal for rules names every rule no switch config names.    *)
TheRefusalNamesEveryUnswitchedRule ==
    outcome = "refused" /\ RuleRefused => Unswitched \subseteq names

(* Line 4: a folder with a plain spec and no model is never refused.       *)
APlainSpecIsSelectedAsBefore ==
    outcome = "refused" => files # "plain"

=============================================================================
