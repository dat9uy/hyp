-------------------------------- MODULE own ---------------------------------
\* Own path at one run of a Goal G. The run's commit holds "g", the runner's
\* change, and may hold more: "O" is an owner's act from the default branch,
\* such as another Goal's add, whose bytes that act recorded, examples and
\* specs; "we" is a runner's write to another Goal's example; "ws" is a
\* runner's write to another Goal's spec; "f" is a runner's new file in a
\* goals/ folder of no Goal; "N" is another fix newer than H0 that the
\* runner merged. The owner's act changes no code, so it is no base. A
\* write the owner's acts did not record is the runner's, whatever branch
\* brought it.

EXTENDS Naturals, FiniteSets

CONSTANTS AuthorWalk, NoForeignCheck, ExamplesOnly, NoFolderCheck,
          OwnerActIsBase

VARIABLES run, result

vars == <<run, result>>

Parts == {"O", "we", "ws", "f", "N"}

\* The bases of a commit: each build below it newer than H0 that is not the
\* runner's. OwnerActIsBase counts the owner's act as one.
Bases(c) == ({"N"} \cap c) \cup (IF OwnerActIsBase THEN {"O"} \cap c ELSE {})

Commits == {c \in SUBSET (Parts \cup {"g"}) : "g" \in c}

\* What the check reads as a write to another Goal's folder. AuthorWalk
\* reads the owner's act as the runner's, because it is on the branch.
Foreign(c) == \/ "we" \in c
              \/ ("ws" \in c /\ ~ExamplesOnly)
              \/ ("O" \in c /\ AuthorWalk)

Init == run = {} /\ result = "none"

\* command: hyp run
Run(c) ==
    /\ result = "none"
    /\ run' = c
    /\ result' = IF Foreign(c) /\ ~NoForeignCheck THEN "other goal"
                 ELSE IF "f" \in c /\ ~NoFolderCheck THEN "no goal"
                 ELSE IF Cardinality(Bases(c)) > 1 THEN "two bases"
                 ELSE "recorded"

Next == \E c \in Commits : Run(c)

Spec == Init /\ [][Next]_vars

\* source: G-B3HMNX:line-1
OwnerActNotRefused ==
    result = "other goal" => ("we" \in run \/ "ws" \in run)

\* source: G-B3HMNX:line-2
OwnerActIsNoBase ==
    result = "two bases" => Cardinality(run \ {"g", "O", "we", "ws", "f"}) > 1

\* source: G-B3HMNX:line-3
ForeignWriteRefused ==
    result = "recorded" => ("we" \notin run /\ "ws" \notin run)

\* source: G-B3HMNX:line-4
NoGoalFolderRefused ==
    result = "recorded" => "f" \notin run

=============================================================================
