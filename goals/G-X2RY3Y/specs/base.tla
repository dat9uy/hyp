-------------------------------- MODULE base --------------------------------
\* One Goal in work whose rows may stand on code newer than its H0. A commit
\* is the set of builds it holds: "g" is the Goal's own change, the others are
\* builds another PR or Goal made. A commit's base is the commit without "g".
\* H0 is the empty set. Each example measures one of three things on a base:
\* the same floors and ceilings as on H0, the same verdict with a floor that
\* differs, or another verdict. The environment can change on its own.

EXTENDS Naturals, FiniteSets

CONSTANTS NoBaseCheck, VerdictOnly, RefuseNewer, RefusalStopsGoal,
          SelectedRefusesRun, PrOmitsBase, BaseCostsRun, NoEnvRecheck

VARIABLES rows, sel, bases, refused, used, inWork, env, prNamesBase, belowRefused

vars == <<rows, sel, bases, refused, used, inWork, env, prNamesBase, belowRefused>>

\* N changes nothing an example measures; P changes a floor of e1 and keeps
\* its verdict; V changes the verdict of e2.
Others == {"N", "P", "V"}
Commits == {c \in SUBSET (Others \cup {"g"}) : "g" \in c}
Base(c) == c \ {"g"}
E == {"e1", "e2"}
Budget == 3

\* What example e measures on base b, against what it measured on H0.
M(e, b) == IF e = "e1" /\ "P" \in b THEN "floor"
           ELSE IF e = "e2" /\ "V" \in b THEN "verdict"
           ELSE "same"

\* What the check sees: a changed environment changes every measure, on the
\* new base and on the old one.
Differs(e, b) == \/ env = "changed"
                 \/ M(e, b) = "verdict"
                 \/ (M(e, b) = "floor" /\ ~VerdictOnly)

CodeDiffers(b) == \E e \in E : M(e, b) # "same"

Init == /\ rows = {}
        /\ sel = {}
        /\ bases = {{}}
        /\ refused = {}
        /\ used = 0
        /\ inWork = TRUE
        /\ env = "same"
        /\ prNamesBase = TRUE
        /\ belowRefused = FALSE

Above(c) == sel # {} /\ sel \subseteq c /\ sel # c

Accept(c) ==
    LET b == Base(c) IN
    /\ rows' = rows \cup {c}
    /\ bases' = bases \cup {b}
    /\ used' = used + 1 + (IF BaseCostsRun /\ b \notin bases THEN 1 ELSE 0)
    /\ sel' = c
    /\ prNamesBase' = ~PrOmitsBase
    /\ UNCHANGED <<refused, inWork, env, belowRefused>>

Refuse(c, why) ==
    /\ refused' = refused \cup {[base |-> Base(c), why |-> why, envChanged |-> env = "changed"]}
    /\ inWork' = IF RefusalStopsGoal THEN FALSE ELSE inWork
    /\ UNCHANGED <<rows, sel, bases, used, env, prNamesBase, belowRefused>>

\* command: hyp run
Run(c) ==
    LET b == Base(c) IN
    /\ inWork
    /\ used < Budget
    /\ c \notin rows
    /\ sel = {} \/ Above(c)
    /\ IF Above(c) /\ SelectedRefusesRun
         THEN /\ belowRefused' = TRUE
              /\ UNCHANGED <<rows, sel, bases, refused, used, inWork, env, prNamesBase>>
       ELSE IF b \in bases \/ NoBaseCheck
         THEN Accept(c)
       ELSE IF RefuseNewer
         THEN Refuse(c, "code")
       ELSE IF \E e \in E : Differs(e, b)
         THEN IF env = "changed" /\ ~NoEnvRecheck
                THEN Refuse(c, "environment")
                ELSE Refuse(c, "code")
       ELSE Accept(c)

\* The environment changes: no command takes this step.
EnvChange == /\ env = "same"
             /\ env' = "changed"
             /\ UNCHANGED <<rows, sel, bases, refused, used, inWork, prNamesBase, belowRefused>>

Next == (\E c \in Commits : Run(c)) \/ EnvChange

Spec == Init /\ [][Next]_vars

\* For the chart only: few steps, so the chart stays readable.
Small == Cardinality(rows) + Cardinality(refused) <= 2 /\ Cardinality(bases) <= 2

\* source: G-X2RY3Y:line-1
NoRefusalWhenSame ==
    \A r \in refused : r.why = "code" => CodeDiffers(r.base)

\* source: G-X2RY3Y:line-2
NoRowOnChangedBase ==
    \A c \in rows : ~CodeDiffers(Base(c))

\* source: G-X2RY3Y:line-3
RefusedKeepsGoalInWork ==
    refused # {} => inWork

\* source: G-X2RY3Y:line-4
RunBelowSelectAccepted == ~belowRefused

\* source: G-X2RY3Y:line-5
PrNamesBase ==
    sel # {} => prNamesBase

\* source: G-X2RY3Y:line-6
BaseCostsNoRun == used = Cardinality(rows)

\* source: G-X2RY3Y:line-7
EnvironmentNamed ==
    \A r \in refused : r.envChanged /\ ~CodeDiffers(r.base) => r.why = "environment"

=============================================================================
