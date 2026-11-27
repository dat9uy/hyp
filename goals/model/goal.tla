----------------------------- MODULE goal -----------------------------
\* A Goal's lifecycle when git holds only the owner's acts: add goal carries
\* the examples, an owner act adds an example, runs record Results in each
\* clone's trace only, "selected" is calculated from a clone's Results, land
\* carries the selected commit and each example's verdict, and unselect names
\* the commit it rejects. The store starts empty (new form) or holds an older
\* Goal G-ONE with a place of e1 and a select of c1 (old form).
\*
\* Examples: e1 and e2 agree with their line on H0; bad does not; e1x is a
\* new version of e1 that changes its run section, not only its floors.
\* Commits: c1 meets every example; c2 meets e1 and fails e2.
\* Each clone's trace holds its own Results. At most one example is added
\* after add goal.

EXTENDS Naturals, FiniteSets, Sequences

\* GrillPlaces: as today, add goal carries no example and the grill's place
\* writes it to git; the Goal is written until then.
\* AddGoalAcceptsDisagreeing: add goal records an example whose verdict on H0
\* does not agree with its line.
\* AddExampleSkipsH0: the owner's act adds an example without its H0 check.
\* AddExampleSkipsAddsOnly: the owner's act adds a new version that changes
\* more than a floor or a ceiling.
\* RunWritesGit: run records its Result in git, not only in the trace.
\* SelectIgnoresMissing: the selected view counts a missing Result as met.
\* LandTrustsMissing: land carries "met" for a Result the clone lacks, and
\* runs nothing.
\* UnselectNamesNothing: unselect writes an event that names no commit.
\* SelectIgnoresUnselect: the selected view does not read unselect events.
\* IgnoreOldSelect: the selected view reads no select event of the old form.
\* StatusReadsHistory: hyp status reads the leaf marks of the rows from git
\* history, so a clone whose runner commits were pruned shows the Goal open.
\* RowsFromBranches: hyp status reads the rows from the branches, so a clone
\* whose runner branches were deleted shows the Goal open.
CONSTANTS GrillPlaces, AddGoalAcceptsDisagreeing, AddExampleSkipsH0,
          AddExampleSkipsAddsOnly, RunWritesGit, SelectIgnoresMissing,
          LandTrustsMissing, UnselectNamesNothing, SelectIgnoresUnselect,
          IgnoreOldSelect, StatusReadsHistory, RowsFromBranches

\* The number of clones and of commits, 1 or 2 each. The check config uses
\* two of each; the chart config, which only draws the chart, one of each.
CONSTANTS NClones, NCommits

\* pruned and branchGone are git outside hyp: the runner's branches were
\* deleted, and git pruned the commits no ref reaches. A run reads git and
\* its Result carries what it read, so no other variable depends on them.
VARIABLES form, added, atAdd, later, results, unselected, anyUnselect,
          rejected, landed, landVerdict, closed, gitWriters, out,
          pruned, branchGone

vars == <<form, added, atAdd, later, results, unselected, anyUnselect,
          rejected, landed, landVerdict, closed, gitWriters, out,
          pruned, branchGone>>

Clones == IF NClones = 1 THEN {"a"} ELSE {"a", "b"}
C == IF NCommits = 1 THEN {"c1"} ELSE {"c1", "c2"}
Ex == {"e1", "e2", "bad", "e1x"}
Verdicts == {"none", "met", "failed"}

\* The verdict each example gives on H0 agrees with its line.
Agrees(e) == e # "bad"
\* A new version only adds a floor or a ceiling.
AddsOnly(e) == e # "e1x"
\* The verdict an example gives on a commit's bytes.
Truth(c, e) == IF c = "c2" /\ e = "e2" THEN "failed" ELSE "met"

placed == atAdd \cup later
ended == landed # "none" \/ closed

GoalTypeOK ==
    /\ form \in {"old", "new"}
    /\ added \in BOOLEAN
    /\ atAdd \subseteq Ex /\ later \subseteq Ex
    /\ results \in [Clones -> [C -> [Ex -> Verdicts]]]
    /\ unselected \subseteq C /\ rejected \subseteq C
    /\ anyUnselect \in BOOLEAN
    /\ landed \in C \cup {"none"}
    /\ landVerdict \in [Ex -> Verdicts]
    /\ closed \in BOOLEAN
    /\ gitWriters \subseteq {"owner", "grill", "runner"}
    /\ pruned \in BOOLEAN /\ branchGone \in BOOLEAN

\* The old form's select of c1 holds while no unselect event exists.
OldSel == IF form = "old" /\ ~IgnoreOldSelect /\ ~anyUnselect
          THEN "c1" ELSE "none"

Unsel == IF SelectIgnoresUnselect THEN {} ELSE unselected

Met(k, c, e) == IF SelectIgnoresMissing
                THEN results[k][c][e] # "failed"
                ELSE results[k][c][e] = "met"

\* The commits clone k's Results show met on every line.
Candidates(k) ==
    {c \in C : /\ placed # {}
               /\ c \notin Unsel
               /\ \A e \in placed : Met(k, c, e)
               /\ \E e \in placed : results[k][c][e] = "met"}

\* What hyp status G shows selected in clone k, calculated, never stored.
SelectedIn(k) ==
    IF ended THEN "none"
    ELSE IF OldSel # "none" THEN OldSel
    ELSE IF Cardinality(Candidates(k)) = 1
         THEN CHOOSE c \in Candidates(k) : TRUE
         ELSE "none"

Init ==
    /\ form \in {"old", "new"}
    /\ added = (form = "old")
    /\ atAdd = IF form = "old" THEN {"e1"} ELSE {}
    /\ later = {}
    /\ results = [k \in Clones |-> [c \in C |-> [e \in Ex |-> "none"]]]
    /\ unselected = {}
    /\ anyUnselect = FALSE
    /\ rejected = {}
    /\ landed = "none"
    /\ landVerdict = [e \in Ex |-> "none"]
    /\ closed = FALSE
    /\ gitWriters = {}
    /\ out = "none"
    /\ pruned = FALSE
    /\ branchGone = FALSE

\* The owner adds the Goal with its examples, each with its hash and its
\* verdict on H0.
\* command: hyp add goal
AddGoal ==
    \E S \in {{"e1"}, {"e1", "bad"}} :
        /\ ~added
        /\ IF ~GrillPlaces /\ ~AddGoalAcceptsDisagreeing
                 /\ \E e \in S : ~Agrees(e)
           THEN /\ out' = "add goal exit 1: wrong-verdict-on-h0"
                /\ UNCHANGED <<added, atAdd, gitWriters, pruned, branchGone>>
           ELSE /\ added' = TRUE
                /\ atAdd' = IF GrillPlaces THEN {} ELSE S
                /\ gitWriters' = gitWriters \cup {"owner"}
                /\ out' = IF GrillPlaces THEN "add goal exit 0: written"
                          ELSE "add goal exit 0: open"
        /\ UNCHANGED <<form, later, results, unselected, anyUnselect,
                       rejected, landed, landVerdict, closed, pruned, branchGone>>

\* An act adds a new example to the Goal: the owner's, or with GrillPlaces
\* the grill's place of today.
\* command: hyp place
Place ==
    \E e \in Ex \ placed :
        /\ added /\ ~ended /\ later = {}
        /\ GrillPlaces \/ atAdd # {}
        /\ IF ~AddExampleSkipsH0 /\ ~Agrees(e)
           THEN /\ out' = "place exit 1: wrong-verdict-on-h0"
                /\ UNCHANGED <<later, gitWriters, pruned, branchGone>>
           ELSE IF ~AddExampleSkipsAddsOnly /\ ~AddsOnly(e)
           THEN /\ out' = "place exit 1: section-differs"
                /\ UNCHANGED <<later, gitWriters, pruned, branchGone>>
           ELSE /\ later' = {e}
                /\ gitWriters' = gitWriters \cup
                       {IF GrillPlaces THEN "grill" ELSE "owner"}
                /\ out' = "place exit 0: " \o e
        /\ UNCHANGED <<form, added, atAdd, results, unselected, anyUnselect,
                       rejected, landed, landVerdict, closed, pruned, branchGone>>

\* The runner runs one example on one commit in clone k. The Result goes to
\* that clone's trace.
\* command: hyp run
Run ==
    \E k \in Clones, c \in C, e \in Ex :
        /\ added /\ ~ended /\ placed # {}
        /\ e \in placed
        /\ results[k][c][e] = "none"
        /\ results' = [results EXCEPT ![k][c][e] = Truth(c, e)]
        /\ gitWriters' = IF RunWritesGit THEN gitWriters \cup {"runner"}
                         ELSE gitWriters
        /\ out' = "run exit 0: " \o c \o " " \o e \o " " \o Truth(c, e)
        /\ UNCHANGED <<form, added, atAdd, later, unselected, anyUnselect,
                       rejected, landed, landVerdict, closed, pruned, branchGone>>

\* What hyp status G prints in clone k, calculated from events alone.
StatusFromEvents(k) ==
    IF ~added THEN "status: no goal"
    ELSE IF ended THEN "status: ended"
    ELSE IF placed = {} THEN "status: written"
    ELSE IF SelectedIn(k) # "none" THEN "status: selected on " \o SelectedIn(k)
    ELSE "status: open"

\* What hyp status G prints in clone k. A variant that reads git history
\* or branches loses the select its Results show once git lost them.
StatusText(k) ==
    IF /\ \/ StatusReadsHistory /\ pruned
          \/ RowsFromBranches /\ branchGone
       /\ SelectedIn(k) \notin {"none", OldSel}
    THEN "status: open"
    ELSE StatusFromEvents(k)

\* hyp status G in clone k: a read; it changes only what is printed.
\* command: hyp status
Status ==
    \E k \in Clones :
        /\ out' = StatusText(k)
        /\ UNCHANGED <<form, added, atAdd, later, results, unselected,
                       anyUnselect, rejected, landed, landVerdict, closed,
                       gitWriters, pruned, branchGone>>

\* The commits clone k lets the owner land: the selected one, or a row with
\* no failed Result and at least one met, whose missing Results land runs.
LandRow(k, c) ==
    /\ c \notin unselected
    /\ \/ c = SelectedIn(k)
       \/ /\ \A e \in placed : results[k][c][e] # "failed"
          /\ \E e \in placed : results[k][c][e] = "met"

\* The owner lands on commit c in clone k. Land runs each missing Result
\* first, carries c's tree and each example's verdict, and refuses when one
\* is not met.
\* command: hyp land
Land ==
    \E k \in Clones, c \in C :
        /\ added /\ ~ended /\ placed # {}
        /\ LandRow(k, c)
        /\ LET v(e) == IF results[k][c][e] # "none" THEN results[k][c][e]
                       ELSE IF LandTrustsMissing THEN "met"
                       ELSE Truth(c, e)
           IN /\ results' = IF LandTrustsMissing THEN results
                            ELSE [results EXCEPT ![k][c] =
                                    [e \in Ex |-> IF e \in placed THEN v(e)
                                                  ELSE results[k][c][e]]]
              /\ IF \A e \in placed : v(e) = "met"
                 THEN /\ landed' = c
                      /\ landVerdict' = [e \in Ex |->
                                           IF e \in placed THEN v(e) ELSE "none"]
                      /\ gitWriters' = gitWriters \cup {"owner"}
                      /\ out' = "land exit 0: " \o c
                 ELSE /\ out' = "land exit 1: not met on " \o c
                      /\ UNCHANGED <<landed, landVerdict, gitWriters, pruned, branchGone>>
        /\ UNCHANGED <<form, added, atAdd, later, unselected, anyUnselect,
                       rejected, closed, pruned, branchGone>>

\* The owner unselects the commit clone k shows selected, naming it.
\* command: hyp unselect
Unselect ==
    \E k \in Clones :
        /\ added /\ ~ended
        /\ SelectedIn(k) # "none"
        /\ SelectedIn(k) \notin rejected
        /\ rejected' = rejected \cup {SelectedIn(k)}
        /\ unselected' = IF UnselectNamesNothing THEN unselected
                         ELSE unselected \cup {SelectedIn(k)}
        /\ anyUnselect' = TRUE
        /\ gitWriters' = gitWriters \cup {"owner"}
        /\ out' = "unselect exit 0: " \o SelectedIn(k)
        /\ UNCHANGED <<form, added, atAdd, later, results, landed,
                       landVerdict, closed, pruned, branchGone>>

\* The runner's branches are deleted; their commits stay until git prunes
\* them. Git outside hyp: no variable of hyp changes.
\* command: git branch -D
DeleteBranch ==
    /\ ~branchGone
    /\ branchGone' = TRUE
    /\ out' = "branches deleted"
    /\ UNCHANGED <<form, added, atAdd, later, results, unselected,
                   anyUnselect, rejected, landed, landVerdict, closed,
                   gitWriters, pruned>>

\* Git prunes the commits no ref reaches, such as after a rebase, a squash
\* or a new clone.
\* command: git gc
Prune ==
    /\ branchGone /\ ~pruned
    /\ pruned' = TRUE
    /\ out' = "commits pruned"
    /\ UNCHANGED <<form, added, atAdd, later, results, unselected,
                   anyUnselect, rejected, landed, landVerdict, closed,
                   gitWriters, branchGone>>

\* command: hyp close
Close ==
    /\ added /\ ~ended
    /\ closed' = TRUE
    /\ gitWriters' = gitWriters \cup {"owner"}
    /\ out' = "close exit 0"
    /\ UNCHANGED <<form, added, atAdd, later, results, unselected,
                   anyUnselect, rejected, landed, landVerdict, pruned, branchGone>>

Next == AddGoal \/ Place \/ Run \/ Status \/ Land \/ Unselect \/ Close
        \/ DeleteBranch \/ Prune

Spec == Init /\ [][Next]_vars

\* source: G-T5QHQF:line-1
OpenAtOnce == added => (ended \/ placed # {})

\* source: G-T5QHQF:line-1
AddGoalAgrees == \A e \in atAdd : Agrees(e)

\* source: G-T5QHQF:line-2
NewExampleAgrees == \A e \in later : Agrees(e)

\* source: G-T5QHQF:line-2
NewVersionAddsOnly == \A e \in later : AddsOnly(e)

\* source: G-T5QHQF:line-3
OnlyOwnerWritesGit == gitWriters \subseteq {"owner"}

\* source: G-T5QHQF:line-4
SelectedMeansAllMet ==
    \A k \in Clones :
        LET c == SelectedIn(k) IN
            (c # "none" /\ c # OldSel) =>
                \A e \in placed : results[k][c][e] = "met"

\* source: G-T5QHQF:line-5
LandCarriesTrueVerdicts ==
    landed # "none" =>
        \A e \in placed : /\ landVerdict[e] = "met"
                          /\ Truth(landed, e) = "met"

\* source: G-T5QHQF:line-6
UnselectedNeverShown ==
    \A k \in Clones, c \in rejected : SelectedIn(k) # c

\* source: G-T5QHQF:line-8
OldStoreSameStatus ==
    (form = "old" /\ ~anyUnselect /\ ~ended) =>
        \A k \in Clones : SelectedIn(k) = "c1"

\* source: G-YFB9QP:line-1
StatusSameAfterPrune ==
    pruned => \A k \in Clones : StatusText(k) = StatusFromEvents(k)

\* source: G-YFB9QP:line-2
StatusSameWithoutBranch ==
    branchGone => \A k \in Clones : StatusText(k) = StatusFromEvents(k)

=============================================================================
