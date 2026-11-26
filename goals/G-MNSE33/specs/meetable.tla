----------------------------- MODULE meetable -----------------------------
(***************************************************************************)
(* hyp place refuses an example that no build can meet, Goal G-MNSE33:    *)
(* "hyp place refuses an example that no build can meet: a floor that     *)
(* reads exit N, or floors and ceilings that no output can hold together." *)
(*                                                                         *)
(* A written Goal has one gap line with two example folders, a and b. Each *)
(* example has one kind of floors and ceilings. The grill runs hyp place   *)
(* once: the set is placed and the Goal opens, or the set is refused, the  *)
(* Goal stays written, the refusal names each example it refused, and the *)
(* grill fixes the set to place it again.                                  *)
(*                                                                         *)
(* The kinds, and whether some output can meet them:                       *)
(*   plain                   a floor bye, a ceiling error         yes      *)
(*   exit contains           a contains floor exit 1              no       *)
(*   exit equals             an equals floor exit 0               no       *)
(*   exit in text            a floor said exit 1 twice            yes      *)
(*   exit line in text       a floor of two lines, the last exit 1 yes     *)
(*   equals differ           equals floors hello and bye          no       *)
(*   equals same             equals floors bye and bye            yes      *)
(*   equals lacks contains   equals hello, contains bye           no       *)
(*   equals holds contains   equals hello bye, contains bye       yes      *)
(*   contains holds ceiling  contains hello bye, ceiling bye      no       *)
(*   equals holds ceiling    equals hello bye, ceiling bye        no       *)
(* "exit line in text" is meetable: hyp cuts only the Probe's last line,  *)
(* so a Probe can print a line exit 1 before its exit line.                *)
(*                                                                         *)
(* Variants, each FALSE in meetable.cfg, are cheaper builds:               *)
(*   ExitOnlyContains      checks the exit line on contains floors only   *)
(*   ExitSubstring         refuses every floor that holds the text exit N  *)
(*   ExitLineAnywhere      refuses a floor with any line exit N in it      *)
(*   NoEqualsCheck         checks no equals floor against another floor   *)
(*   NoCeilingCheck        checks no floor against a ceiling              *)
(*   ContainsCeilingOnly   checks only contains floors against a ceiling  *)
(*   RefuseRepeatedEquals  refuses two equals floors, even the same text  *)
(*   NameFirstKey          names the first example of the set, not the    *)
(*                         one it refused                                  *)
(***************************************************************************)
EXTENDS FiniteSets

CONSTANTS
    ExitOnlyContains,
    ExitSubstring,
    ExitLineAnywhere,
    NoEqualsCheck,
    NoCeilingCheck,
    ContainsCeilingOnly,
    RefuseRepeatedEquals,
    NameFirstKey

Keys == {"a", "b"}

Kinds == {"plain", "exit contains", "exit equals", "exit in text",
          "exit line in text", "equals differ", "equals same",
          "equals lacks contains", "equals holds contains",
          "contains holds ceiling", "equals holds ceiling"}

ExitKinds     == {"exit contains", "exit equals"}
ConflictKinds == {"equals differ", "equals lacks contains",
                  "contains holds ceiling", "equals holds ceiling"}
\* the kinds no output can meet
Unmeetable == ExitKinds \cup ConflictKinds

VARIABLES
    goal,   \* "written" or "open"
    set,    \* the kind of each example folder, once the grill placed
    last,   \* what the place did: "none", "placed" or "refused"
    named   \* the examples the refusal names

vars == <<goal, set, last, named>>

\* What the build refuses.
Refuses(k) ==
    \/ k = "exit contains"
    \/ k = "exit equals" /\ ~ExitOnlyContains
    \/ k = "exit in text" /\ ExitSubstring
    \/ k = "exit line in text" /\ (ExitSubstring \/ ExitLineAnywhere)
    \/ k \in {"equals differ", "equals lacks contains"} /\ ~NoEqualsCheck
    \/ k = "equals same" /\ RefuseRepeatedEquals
    \/ k = "contains holds ceiling" /\ ~NoCeilingCheck
    \/ k = "equals holds ceiling" /\ ~NoCeilingCheck /\ ~ContainsCeilingOnly

Init ==
    /\ goal = "written"
    /\ set = [k \in Keys |-> "plain"]
    /\ last = "none"
    /\ named = {}

\* The grill runs hyp place on a set of two examples.
GrillPlaces ==
    /\ last = "none"
    /\ \E s \in [Keys -> Kinds] :
        LET bad == {k \in Keys : Refuses(s[k])}
        IN
        /\ set' = s
        /\ IF bad = {}
           THEN /\ goal' = "open"
                /\ last' = "placed"
                /\ named' = {}
           ELSE /\ goal' = "written"
                /\ last' = "refused"
                /\ named' = IF NameFirstKey THEN {"a"} ELSE bad

\* The grill fixes the refused set, to place it again.
GrillFixes ==
    /\ last = "refused"
    /\ last' = "none"
    /\ named' = {}
    /\ UNCHANGED <<goal, set>>

\* Anyone reads status; it changes nothing.
Status == UNCHANGED vars

Next ==
    \/ GrillPlaces
    \/ GrillFixes
    \/ Status

Spec == Init /\ [][Next]_vars

\* source: model only
TypeOK ==
    /\ goal \in {"written", "open"}
    /\ set \in [Keys -> Kinds]
    /\ last \in {"none", "placed", "refused"}
    /\ named \subseteq Keys

\* source: line-1
NoExitFloorPlaced == last = "placed" => \A k \in Keys : set[k] \notin ExitKinds

\* source: line-1
NamesEachRefused ==
    last = "refused" => named = {k \in Keys : set[k] \in Unmeetable}

\* source: line-2
NoConflictPlaced == last = "placed" => \A k \in Keys : set[k] \notin ConflictKinds

\* source: line-3
MeetableSetPlaced ==
    last = "refused" => \E k \in Keys : set[k] \in Unmeetable

=============================================================================
