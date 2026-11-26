------------------------------- MODULE report -------------------------------
(***************************************************************************)
(* Goal G-0P2A61: "The select's report holds the Goal's words, each line,  *)
(* each placed example's words, each chart and the selected commit and     *)
(* row, and no result of any example."                                     *)
(*                                                                         *)
(* A Goal is "open" (its line ran on a build that meets it), "selected" or *)
(* "landed". Reads run in this clone ("here") or in a fresh clone whose    *)
(* trace holds no Result. The selected commit may hold a chart; the        *)
(* example may write to stderr. A report is the set of parts it prints.    *)
(*                                                                         *)
(* Variants, each FALSE in report.cfg, are cheaper builds:                 *)
(*   KeepResults   the report of a selected Goal still prints each         *)
(*                 example's verdict and output, or a try line in a fresh  *)
(*                 clone (the old report)                                  *)
(*   KeepStderr    the report drops verdict and output but keeps stderr    *)
(*   EndedVerdict  the report of a landed Goal still prints "verdict: met, *)
(*                 by select" under each example                           *)
(*   NoCharts      the report drops the charts                             *)
(***************************************************************************)
CONSTANTS KeepResults, KeepStderr, EndedVerdict, NoCharts

VARIABLES
    gs,     \* "open", "selected" or "landed"
    clone,  \* "here" or "fresh"
    chart,  \* TRUE when the selected commit's spec folder holds a chart
    err,    \* TRUE when the example writes to stderr
    out     \* what the last command printed

vars == <<gs, clone, chart, err, out>>

Whole == {"words", "line", "example", "commit", "row"}
Results == {"verdict", "output", "stderr", "try"}

TypeOK ==
    /\ gs \in {"open", "selected", "landed"}
    /\ clone \in {"here", "fresh"}
    /\ chart \in BOOLEAN
    /\ err \in BOOLEAN

\* The parts of the report of a Goal in state s, read in clone c.
Report(s, c) ==
    LET base == Whole \cup (IF chart /\ ~NoCharts THEN {"chart"} ELSE {})
        res ==
            IF s = "selected" THEN
                IF KeepResults THEN
                    IF c = "here" THEN {"verdict", "output"} \cup (IF err THEN {"stderr"} ELSE {})
                    ELSE {"verdict", "try"}
                ELSE IF KeepStderr /\ c = "here" /\ err THEN {"stderr"}
                ELSE {}
            ELSE IF EndedVerdict THEN {"verdict"}
            ELSE {}
    IN base \cup res

Init ==
    /\ gs = "open"
    /\ clone = "here"
    /\ chart \in BOOLEAN
    /\ err \in BOOLEAN
    /\ out = [cmd |-> "none", gs |-> "open", parts |-> {}]

RunnerSelects ==
    /\ gs = "open"
    /\ gs' = "selected"
    /\ out' = [cmd |-> "select", gs |-> "selected", parts |-> Report("selected", clone)]
    /\ UNCHANGED <<clone, chart, err>>

OwnerReadsStatus ==
    /\ gs \in {"selected", "landed"}
    /\ out' = [cmd |-> "status", gs |-> gs, parts |-> Report(gs, clone)]
    /\ UNCHANGED <<gs, clone, chart, err>>

OwnerLands ==
    /\ gs = "selected"
    /\ gs' = "landed"
    /\ out' = [cmd |-> "land", gs |-> "landed", parts |-> {}]
    /\ UNCHANGED <<clone, chart, err>>

OwnerClones ==
    /\ clone = "here"
    /\ clone' = "fresh"
    /\ out' = [cmd |-> "clone", gs |-> gs, parts |-> {}]
    /\ UNCHANGED <<gs, chart, err>>

Next == RunnerSelects \/ OwnerReadsStatus \/ OwnerLands \/ OwnerClones

Spec == Init /\ [][Next]_vars

Reads == out.cmd \in {"select", "status"}

\* source: line-1
SelectedNoResult == Reads /\ out.gs = "selected" => out.parts \cap Results = {}

\* source: line-2
ReportWhole == Reads => Whole \subseteq out.parts /\ (chart => "chart" \in out.parts)

\* source: line-4
EndedNoVerdict == Reads /\ out.gs = "landed" => "verdict" \notin out.parts
=============================================================================
