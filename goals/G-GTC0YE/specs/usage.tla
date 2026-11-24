------------------------------- MODULE usage -------------------------------
(***************************************************************************)
(* Goal G-GTC0YE: "hyp spec states the whole usage of each command, as    *)
(* hyp help lists it."                                                     *)
(*                                                                         *)
(* A reader runs hyp help or hyp spec. The usage of add goal fills two     *)
(* lines in hyp help: its first line ("base") and a second line ("from":   *)
(* [--from G] [--why TEXT]). A usage is the set of its parts; hyp help     *)
(* prints it as a sequence of lines, hyp spec as one set in the act's      *)
(* entry, under the acts header.                                           *)
(*                                                                         *)
(* Variants, each FALSE in usage.cfg, are cheaper builds:                  *)
(*   FirstLineOnly   hyp spec prints only the first line of the usage (H0) *)
(*   HelpOneLine     hyp help prints add goal's usage on one line, so its  *)
(*                   first line is its whole usage                         *)
(*   DescInUsage     hyp spec joins every line of the help block,          *)
(*                   description included                                  *)
(*   HeaderReworded  hyp spec rewords the header of its acts section       *)
(***************************************************************************)
EXTENDS Sequences

CONSTANTS FirstLineOnly, HelpOneLine, DescInUsage, HeaderReworded

VARIABLES out  \* what the last command printed

vars == <<out>>

Whole == {"base", "from"}
H0Lines == <<{"base"}, {"from"}>>
HelpLines == IF HelpOneLine THEN <<Whole>> ELSE H0Lines

SpecUsage ==
    IF FirstLineOnly THEN HelpLines[1]
    ELSE IF DescInUsage THEN Whole \cup {"desc"}
    ELSE Whole

Blank == [cmd |-> "none", lines |-> <<>>, usage |-> {}, header |-> "none"]

TypeOK == out.cmd \in {"none", "help", "spec"}

Init == out = Blank

ReaderRunsHelp == out' = [Blank EXCEPT !.cmd = "help", !.lines = HelpLines]

ReaderRunsSpec ==
    out' = [Blank EXCEPT !.cmd = "spec", !.usage = SpecUsage,
            !.header = IF HeaderReworded THEN "reworded" ELSE "H0"]

Next == ReaderRunsHelp \/ ReaderRunsSpec

Spec == Init /\ [][Next]_vars

\* source: line-1
SpecUsageWhole == out.cmd = "spec" => out.usage = Whole

\* source: line-2
HelpUnchanged == out.cmd = "help" => out.lines = H0Lines

\* source: line-2
SpecOthersUnchanged == out.cmd = "spec" => out.header = "H0"
=============================================================================
