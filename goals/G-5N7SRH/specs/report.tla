------------------------------- MODULE report -------------------------------
(***************************************************************************)
(* Goal G-5N7SRH: "The select's report holds the Goal's words, each line,  *)
(* each placed example's words, each chart and the selected commit and    *)
(* row, and no result of any example."                                    *)
(*                                                                         *)
(* A Goal is "open", "selected" or "ended" (landed). The clone that reads *)
(* it is "here" (its trace holds the cells) or "fresh" (a new clone, its   *)
(* trace empty). out is what the last command printed: the parts of the    *)
(* report, or what a grep of the docs found.                               *)
(*                                                                         *)
(* Variants, each FALSE in report.cfg, are cheaper builds:                 *)
(*   KeepResult     the report still prints each verdict and output        *)
(*   KeepTry        a fresh clone's report still prints the try line       *)
(*   KeepStderr     the report still prints each example's stderr          *)
(*   KeepEndedMet   an ended Goal's report still prints "verdict: met"     *)
(*   DropExamples   the report drops the examples with their results       *)
(*   DropCharts     the report drops the charts                            *)
(*   OldGuide       the owner guide still says to read each output         *)
(*   OldGlossary    the glossary's PR body entry still lists the results   *)
(***************************************************************************)
CONSTANTS KeepResult, KeepTry, KeepStderr, KeepEndedMet, DropExamples,
    DropCharts, OldGuide, OldGlossary

VARIABLES gs, clone, out

vars == <<gs, clone, out>>

Whole == {"words", "lines", "examples", "charts", "commit", "row"}
Results == {"verdict", "output", "stderr", "try"}
Blank == [cmd |-> "none", parts |-> {}, doc |-> "none", says |-> FALSE]

TypeOK ==
    /\ gs \in {"open", "selected", "ended"}
    /\ clone \in {"here", "fresh"}

Init == gs = "open" /\ clone = "here" /\ out = Blank

\* The parts of the report of the Goal in state s, read in clone c.
Parts(s, c) ==
    ((Whole \ (IF DropExamples THEN {"examples"} ELSE {}))
            \ (IF DropCharts THEN {"charts"} ELSE {}))
    \cup (IF s = "ended" THEN (IF KeepEndedMet THEN {"verdict"} ELSE {})
          ELSE IF c = "fresh" THEN (IF KeepTry THEN {"verdict", "try"} ELSE {})
          ELSE (IF KeepResult THEN {"verdict", "output"} ELSE {})
               \cup (IF KeepStderr THEN {"stderr"} ELSE {}))

RunnerSelects ==
    /\ gs = "open"
    /\ gs' = "selected"
    /\ out' = [Blank EXCEPT !.cmd = "select", !.parts = Parts("selected", clone)]
    /\ UNCHANGED clone

OwnerReadsStatus ==
    /\ gs \in {"selected", "ended"}
    /\ out' = [Blank EXCEPT !.cmd = "status", !.parts = Parts(gs, clone)]
    /\ UNCHANGED <<gs, clone>>

OwnerLands ==
    /\ gs = "selected"
    /\ gs' = "ended"
    /\ out' = [Blank EXCEPT !.cmd = "land"]
    /\ UNCHANGED clone

OwnerClones ==
    /\ clone = "here"
    /\ clone' = "fresh"
    /\ out' = [Blank EXCEPT !.cmd = "clone"]
    /\ UNCHANGED gs

ReaderGreps(d) ==
    /\ out' = [Blank EXCEPT !.cmd = "grep", !.doc = d,
               !.says = IF d = "guide" THEN OldGuide ELSE OldGlossary]
    /\ UNCHANGED <<gs, clone>>

Next ==
    \/ RunnerSelects \/ OwnerReadsStatus \/ OwnerLands \/ OwnerClones
    \/ \E d \in {"guide", "glossary"} : ReaderGreps(d)

Spec == Init /\ [][Next]_vars

\* source: line-1
NoResultOfSelected ==
    (out.cmd \in {"select", "status"} /\ gs # "ended") => out.parts \cap Results = {}

\* source: line-2
ReportWhole ==
    out.cmd \in {"select", "status"} => Whole \subseteq out.parts

\* source: line-3
DocsHoldNoResult ==
    out.cmd = "grep" => ~out.says

\* source: line-4
NoVerdictWhenEnded ==
    (out.cmd = "status" /\ gs = "ended") => "verdict" \notin out.parts
=============================================================================
