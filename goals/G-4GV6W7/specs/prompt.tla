------------------------------- MODULE prompt -------------------------------
(***************************************************************************)
(* Goal G-4GV6W7: hyp prints the runner's Start prompt, so the grill no    *)
(* longer copies and fills it by hand.                                     *)
(*                                                                         *)
(* One Goal moves through its states: written, open, selected, landed and  *)
(* closed. In any state the owner runs `hyp status G`. A reader can also   *)
(* look in the hyp-grill skill, its run rules and the owner guide for a    *)
(* copy of the prompt, and for the sentence that says where hyp prints it. *)
(*                                                                         *)
(* Variants, each FALSE in prompt.cfg, are cheaper builds:                 *)
(*   NoPrompt        hyp status G prints no Start prompt (the base)        *)
(*   Unfilled        the prompt keeps the placeholders <G> and <actor>     *)
(*   KeepsCloses     the prompt still asks for a Closes line               *)
(*   PromptInWork    a written Goal prints the prompt too                  *)
(*   PromptAnyState  every state prints the prompt                         *)
(*   DocsKeepCopy    the skill keeps its copy of the prompt                *)
(*   DocsSilent      the docs do not say that hyp status prints it         *)
(***************************************************************************)
CONSTANTS NoPrompt, Unfilled, KeepsCloses, PromptInWork, PromptAnyState,
          DocsKeepCopy, DocsSilent

VARIABLES
    state,  \* the Goal's state
    out     \* what the last command printed

vars == <<state, out>>

States == {"written", "open", "selected", "landed", "closed"}

HasPrompt(s) ==
    /\ ~NoPrompt
    /\ \/ s = "open"
       \/ PromptInWork /\ s = "written"
       \/ PromptAnyState

Status(s) ==
    [cmd |-> "status", state |-> s, prompt |-> HasPrompt(s),
     filled |-> ~Unfilled, closes |-> KeepsCloses,
     copy |-> FALSE, says |-> FALSE]

Quiet(c) ==
    [cmd |-> c, state |-> "", prompt |-> FALSE, filled |-> TRUE,
     closes |-> FALSE, copy |-> FALSE, says |-> FALSE]

Init ==
    /\ state = "written"
    /\ out = Quiet("none")

GrillPlaces ==
    /\ state = "written"
    /\ state' = "open"
    /\ out' = Quiet("act")

RunnerSelects ==
    /\ state = "open"
    /\ state' = "selected"
    /\ out' = Quiet("act")

OwnerUnselects ==
    /\ state = "selected"
    /\ state' = "open"
    /\ out' = Quiet("act")

OwnerLands ==
    /\ state = "selected"
    /\ state' = "landed"
    /\ out' = Quiet("act")

OwnerCloses ==
    /\ state \in {"written", "open"}
    /\ state' = "closed"
    /\ out' = Quiet("act")

OwnerReadsStatus ==
    /\ out' = Status(state)
    /\ UNCHANGED state

ReaderSearchesCopy ==
    /\ out' = [Quiet("copy") EXCEPT !.copy = DocsKeepCopy]
    /\ UNCHANGED state

ReaderReadsDocs ==
    /\ out' = [Quiet("docs") EXCEPT !.says = ~DocsSilent]
    /\ UNCHANGED state

Next ==
    \/ GrillPlaces
    \/ RunnerSelects
    \/ OwnerUnselects
    \/ OwnerLands
    \/ OwnerCloses
    \/ OwnerReadsStatus
    \/ ReaderSearchesCopy
    \/ ReaderReadsDocs

Spec == Init /\ [][Next]_vars

-----------------------------------------------------------------------------
(* Rules. Each is an INVARIANT of prompt.cfg.                              *)

\* source: line-1
OpenGoalPrintsPrompt ==
    out.cmd = "status" /\ out.state = "open" => out.prompt

\* source: line-1
PromptIsFilled ==
    out.prompt => out.filled

\* source: line-1
PromptAsksNoCloses ==
    out.prompt => ~out.closes

\* source: line-3
OtherStatesPrintNoPrompt ==
    out.cmd = "status" /\ out.state # "open" => ~out.prompt

\* source: line-2
DocsHoldNoCopy ==
    out.cmd = "copy" => ~out.copy

\* source: line-2
DocsSayWhereHypPrintsIt ==
    out.cmd = "docs" => out.says

TypeOK ==
    /\ state \in States
    /\ out.cmd \in {"none", "act", "status", "copy", "docs"}

=============================================================================
