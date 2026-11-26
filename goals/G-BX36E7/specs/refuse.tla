------------------------------- MODULE refuse -------------------------------
(***************************************************************************)
(* Goal G-BX36E7: when the statechart tool checks a Goal's model whose     *)
(* check config hyp select will refuse, it says so and does not exit 0.    *)
(*                                                                         *)
(* A writer runs the statechart tool on one of three Goal models: one      *)
(* whose check config checks a rule no variant breaks (unnamed), one whose *)
(* every rule a variant breaks (named), and one whose own behaviour breaks *)
(* a rule (breaks). A runner also reads the run rules for what to do when  *)
(* hyp prints no select line.                                              *)
(*                                                                         *)
(* Variants, each FALSE in refuse.cfg, are cheaper builds:                 *)
(*   ExitZero      says hyp select refuses, but still exits 0              *)
(*   NoMessage     exits 3, but does not say that hyp select refuses       *)
(*   NoReport      refuses before it writes the report and variant configs *)
(*   SaysAlways    says hyp select refuses on every model                  *)
(*   SaysOnBreak   says hyp select refuses when the model breaks a rule    *)
(*   KeepsSilent   the run rules keep "hyp status does not say so"         *)
(*   NoPointer     the run rules drop it but do not point at hyp status    *)
(***************************************************************************)
CONSTANTS ExitZero, NoMessage, NoReport, SaysAlways, SaysOnBreak,
          KeepsSilent, NoPointer

VARIABLE out  \* what the last command printed

Tool(case, exit, says, wrote) ==
    [cmd |-> "tool", case |-> case, exit |-> exit, says |-> says,
     wrote |-> wrote, silent |-> FALSE, points |-> FALSE]

Init == out = [Tool("none", 0, FALSE, FALSE) EXCEPT !.cmd = "none"]

RunToolOnUnnamed ==
    out' = Tool("unnamed", IF ExitZero THEN 0 ELSE 3, ~NoMessage, ~NoReport)

RunToolOnNamed ==
    out' = Tool("named", 0, SaysAlways, TRUE)

RunToolOnBreaks ==
    out' = Tool("breaks", 1, SaysAlways \/ SaysOnBreak, TRUE)

ReadRunRules ==
    out' = [Tool("rules", 0, FALSE, FALSE) EXCEPT !.cmd = "rules",
               !.silent = KeepsSilent, !.points = ~KeepsSilent /\ ~NoPointer]

Next ==
    \/ RunToolOnUnnamed
    \/ RunToolOnNamed
    \/ RunToolOnBreaks
    \/ ReadRunRules

Spec == Init /\ [][Next]_out

-----------------------------------------------------------------------------
(* Rules. Each is an INVARIANT of refuse.cfg.                              *)

\* source: line-1
UnnamedSaysSelectRefuses ==
    out.case = "unnamed" => out.says

\* source: line-1
UnnamedExitsNotZero ==
    out.case = "unnamed" => out.exit # 0

\* source: line-1
UnnamedStillWrites ==
    out.case = "unnamed" => out.wrote

\* source: line-3
NamedAsOnH0 ==
    out.case = "named" => out.exit = 0 /\ ~out.says /\ out.wrote

\* source: line-3
BreaksAsOnH0 ==
    out.case = "breaks" => out.exit = 1 /\ ~out.says

\* source: line-2
RulesNotSilent ==
    out.cmd = "rules" => ~out.silent

\* source: line-2
RulesPointAtStatus ==
    out.cmd = "rules" => out.points

TypeOK ==
    /\ out.cmd \in {"none", "tool", "rules"}
    /\ out.exit \in {0, 1, 3}

=============================================================================
