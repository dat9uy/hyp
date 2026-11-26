----------------------------- MODULE saywhy -----------------------------
(***************************************************************************)
(* hyp says why it holds back a select or refuses a show argument,         *)
(* Goal G-W8MR53: "When hyp holds back a select or refuses an argument of *)
(* show, it says why and what to run instead."                             *)
(*                                                                         *)
(* An open Goal has up to two candidate rows, c1 and c2. The select of a  *)
(* candidate is held back when its commit fails a select guard that reads *)
(* git: Spec (no-spec, rule-no-variant, no-chart). Own path is left out: *)
(* hyp run refuses the same commit, so no candidate is held back by it.   *)
(* A runner run makes a row a candidate, with any hold;                  *)
(* a fix commit clears a hold. Any role reads hyp status, with or without *)
(* G. Anyone runs hyp show with an actor name, a Goal id or another word. *)
(*                                                                         *)
(* Variants, each FALSE in saywhy.cfg, are cheaper builds:                 *)
(*   LoneOnly       names a hold only when the Goal has one candidate     *)
(*   RunnerOnly     names a hold only to the runner role                  *)
(*   GOnly          names a hold in hyp status G, not in hyp status       *)
(*   CodeNoFix      names the guard and code, not the fix                 *)
(*   NoteUnheld     also prints a note for a candidate that is not held   *)
(*   GoalIdOnly     the new show message only for a Goal id               *)
(*   NoStatusHint   show of a Goal id does not name hyp status G          *)
(*   ShowExit1      show of a non-actor argument exits 1, not 2           *)
(***************************************************************************)
EXTENDS FiniteSets

CONSTANTS LoneOnly, RunnerOnly, GOnly, CodeNoFix, NoteUnheld,
          GoalIdOnly, NoStatusHint, ShowExit1

Cands == {"c1", "c2"}
Holds == {"no-spec", "rule-no-variant", "no-chart"}
Roles == {"owner", "grill", "runner"}
Args == {"actor", "goal-id", "other"}

VARIABLES
    cand,   \* each row: "absent", "free" (selectable) or a hold
    out     \* what the last command printed

vars == <<cand, out>>

Present == {c \in Cands : cand[c] # "absent"}
Held == {c \in Cands : cand[c] \in Holds}
Free == {c \in Cands : cand[c] = "free"}

\* The selects hyp status lists, as on H0: the lone free candidate to the
\* runner; at Two candidates every free one to the owner.
Selects(role) ==
    IF Cardinality(Present) = 1 /\ role = "runner" THEN Free
    ELSE IF Cardinality(Present) = 2 /\ role = "owner" THEN Free
    ELSE {}

Named(role, withG) ==
    IF (LoneOnly /\ Cardinality(Present) # 1)
       \/ (RunnerOnly /\ role # "runner")
       \/ (GOnly /\ ~withG)
    THEN {}
    ELSE Held

Notes(role, withG) ==
    { [c |-> c, hold |-> cand[c], guard |-> "Spec", fix |-> ~CodeNoFix]
        : c \in Named(role, withG) }
    \cup (IF NoteUnheld
          THEN { [c |-> c, hold |-> "free", guard |-> "none", fix |-> FALSE] : c \in Free }
          ELSE {})

None == [cmd |-> "none", role |-> "none", withG |-> FALSE, arg |-> "none",
         notes |-> {}, selects |-> {}, actorword |-> FALSE, statushint |-> FALSE,
         exit |-> 0]

Init ==
    /\ cand = [c \in Cands |-> "absent"]
    /\ out = None

\* The runner runs a row on a commit: the row is a candidate, held or not.
RunnerRunsRow ==
    \E c \in Cands, h \in Holds \cup {"free"} :
        /\ cand[c] = "absent"
        /\ cand' = [cand EXCEPT ![c] = h]
        /\ out' = None

\* The runner commits the fix of a hold, and runs one line on that commit.
RunnerFixesHold ==
    \E c \in Held :
        /\ cand' = [cand EXCEPT ![c] = "free"]
        /\ out' = None

\* A role reads hyp status, with G or without.
ReadsStatus ==
    \E role \in Roles, withG \in BOOLEAN :
        /\ cand' = cand
        /\ out' = [None EXCEPT !.cmd = "status", !.role = role, !.withG = withG,
                    !.notes = Notes(role, withG), !.selects = Selects(role)]

\* Anyone runs hyp show with one argument.
RunsShow ==
    \E a \in Args :
        /\ cand' = cand
        /\ out' = IF a = "actor"
                  THEN [None EXCEPT !.cmd = "show", !.arg = a]
                  ELSE LET new == ~(GoalIdOnly /\ a # "goal-id")
                       IN [None EXCEPT !.cmd = "show", !.arg = a,
                            !.actorword = new,
                            !.statushint = new /\ a = "goal-id" /\ ~NoStatusHint,
                            !.exit = IF ShowExit1 THEN 1 ELSE 2]

Next ==
    \/ RunnerRunsRow
    \/ RunnerFixesHold
    \/ ReadsStatus
    \/ RunsShow

Spec == Init /\ [][Next]_vars

TypeOK ==
    /\ cand \in [Cands -> Holds \cup {"absent", "free"}]

\* source: line-1
EachHoldNamed ==
    out.cmd = "status" =>
        {n.c : n \in {m \in out.notes : m.hold \in Holds}} = Held

\* source: line-1
NoteHasFix ==
    out.cmd = "status" => \A n \in out.notes : n.hold \in Holds => n.fix

\* source: owner decision 2
HeldAtTwoNamed ==
    (out.cmd = "status" /\ Cardinality(Present) = 2) =>
        Held \subseteq {n.c : n \in out.notes}

\* source: line-3
NoNoteUnheld ==
    out.cmd = "status" => \A n \in out.notes : n.hold \in Holds

\* source: line-2
ShowNamesActor ==
    (out.cmd = "show" /\ out.arg # "actor") => out.actorword /\ out.exit = 2

\* source: line-2
ShowGoalIdHint ==
    (out.cmd = "show" /\ out.arg = "goal-id") => out.statushint

\* source: owner decision 3
ShowOtherArg ==
    (out.cmd = "show" /\ out.arg = "other") => out.actorword
=============================================================================
