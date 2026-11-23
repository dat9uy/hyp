----------------------------- MODULE ended_goal -----------------------------
(***************************************************************************)
(* Goal G-KSCGXV: "An ended Goal reads the same in every clone: hyp status *)
(* G prints each example of its selected row as met by the select, from    *)
(* git alone, and hyp reads Results, reads and outputs only from the       *)
(* clone's trace, so git's .tree/ needs only the owner's acts and the      *)
(* places."                                                                *)
(*                                                                         *)
(* Clone A runs, selects and lands the Goal; clone B is a fresh clone with *)
(* an empty trace. Git may also hold a result.record line that an older    *)
(* hyp wrote, and an old place may name an output kept under              *)
(* .tree/outputs/.                                                         *)
(*                                                                         *)
(* Variants, each FALSE in ended_goal.cfg, are cheaper builds:             *)
(*   StatusShowsRuns  the state line of an ended Goal keeps its runs count *)
(*   MetNoSelectId    an ended Goal's example prints "met" and names no    *)
(*                    select                                               *)
(*   WarnGitResult    hyp skips a result.record in git but lists it under  *)
(*                    "not read at replay"                                 *)
(*   OutputFromGit    hyp still prints an old place's output from          *)
(*                    .tree/outputs/                                       *)
(*   SpecMentionsOld  hyp spec still says what an earlier hyp stored in    *)
(*                    git                                                  *)
(*   StrictEnded      status of an ended Goal whose trace lacks the        *)
(*                    select's cells exits 3                               *)
(***************************************************************************)
CONSTANTS StatusShowsRuns, MetNoSelectId, WarnGitResult, OutputFromGit,
          SpecMentionsOld, StrictEnded

VARIABLES
    ran,       \* clone A's trace holds the candidate's Result
    selected,  \* git holds the select
    landed,    \* git holds the owner's land
    gitResult, \* git holds a result.record line an older hyp wrote
    out        \* what the last command printed

vars == <<ran, selected, landed, gitResult, out>>

Blank == [cmd |-> "none", same |-> TRUE, names |-> TRUE, runsA |-> FALSE,
          runsB |-> FALSE, oldBytes |-> FALSE, earlier |-> FALSE,
          exitAll |-> TRUE]

TypeOK ==
    /\ ran \in BOOLEAN /\ selected \in BOOLEAN /\ landed \in BOOLEAN
    /\ gitResult \in BOOLEAN
    /\ out.cmd \in {"none", "run", "select", "land", "old", "compare",
                    "gitresult", "table", "spec", "replay"}

Init == ran = FALSE /\ selected = FALSE /\ landed = FALSE
        /\ gitResult = FALSE /\ out = Blank

RunnerRuns ==
    /\ ~ran /\ ~selected
    /\ ran' = TRUE /\ out' = [Blank EXCEPT !.cmd = "run"]
    /\ UNCHANGED <<selected, landed, gitResult>>

RunnerSelects ==
    /\ ran /\ ~selected
    /\ selected' = TRUE /\ out' = [Blank EXCEPT !.cmd = "select"]
    /\ UNCHANGED <<ran, landed, gitResult>>

OwnerLands ==
    /\ selected /\ ~landed
    /\ landed' = TRUE /\ out' = [Blank EXCEPT !.cmd = "land"]
    /\ UNCHANGED <<ran, selected, gitResult>>

OlderHypWrites ==
    /\ ran /\ ~gitResult
    /\ gitResult' = TRUE /\ out' = [Blank EXCEPT !.cmd = "old"]
    /\ UNCHANGED <<ran, selected, landed>>

\* status G of the ended Goal in clone A (its trace holds the Result) and in
\* clone B (its trace is empty). A runs count differs: A ran 1, B ran 0.
ReaderComparesClones ==
    /\ landed
    /\ out' = [Blank EXCEPT !.cmd = "compare",
                            !.same = ~StatusShowsRuns /\ ~StrictEnded,
                            !.names = ~MetNoSelectId,
                            !.runsA = StatusShowsRuns,
                            !.runsB = StatusShowsRuns /\ ~StrictEnded]
    /\ UNCHANGED <<ran, selected, landed, gitResult>>

\* status G in a fresh clone, with and without the older hyp's line in git.
ReaderComparesGitResult ==
    /\ gitResult
    /\ out' = [Blank EXCEPT !.cmd = "gitresult", !.same = ~WarnGitResult]
    /\ UNCHANGED <<ran, selected, landed, gitResult>>

\* status G at a stop that prints the table of outputs, H0's place naming an
\* output kept only under .tree/outputs/.
ReaderReadsTable ==
    /\ out' = [Blank EXCEPT !.cmd = "table", !.oldBytes = OutputFromGit]
    /\ UNCHANGED <<ran, selected, landed, gitResult>>

ReaderReadsSpec ==
    /\ out' = [Blank EXCEPT !.cmd = "spec", !.earlier = SpecMentionsOld]
    /\ UNCHANGED <<ran, selected, landed, gitResult>>

\* status G of every Goal ended on main, from git alone, with an empty trace.
ReaderReplaysMain ==
    /\ out' = [Blank EXCEPT !.cmd = "replay", !.exitAll = ~StrictEnded]
    /\ UNCHANGED <<ran, selected, landed, gitResult>>

Next == RunnerRuns \/ RunnerSelects \/ OwnerLands \/ OlderHypWrites
        \/ ReaderComparesClones \/ ReaderComparesGitResult
        \/ ReaderReadsTable \/ ReaderReadsSpec \/ ReaderReplaysMain

Spec == Init /\ [][Next]_vars

\* source: line-1
SameInEveryClone == out.cmd = "compare" => out.same

\* source: line-1
MetBySelect == out.cmd = "compare" => out.names

\* source: line-2
NoRunsWhenEnded == out.cmd = "compare" => (~out.runsA /\ ~out.runsB)

\* source: line-3
GitResultIgnored == out.cmd = "gitresult" => out.same

\* source: line-4
NoGitOutput == out.cmd = "table" => ~out.oldBytes

\* source: line-4
SpecSilentOnOld == out.cmd = "spec" => ~out.earlier

\* source: line-5
EndedExitsZero == out.cmd = "replay" => out.exitAll

\* source: model only
TypeInvariant == TypeOK
=============================================================================
