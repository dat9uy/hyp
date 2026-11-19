------------------------------- MODULE try -------------------------------
(***************************************************************************)
(* See a verdict before it is recorded, Goal G-CJWT9T.                     *)
(*                                                                         *)
(* One Goal with H0. The grill writes an example folder and may commit it; *)
(* an agent may edit a tracked file and add a new one in the checkout, and *)
(* may commit that change. hyp try runs the example as it is on disk, on   *)
(* H0, or with --worktree on the checkout's files. hyp place records H0's  *)
(* cell of the committed example; hyp run records a cell on HEAD.          *)
(*                                                                         *)
(* A Probe's output, stderr, exit code and verdict are one function of the *)
(* bytes it runs on: Out(example, subject).                                *)
(*                                                                         *)
(* The rules are the Goal's line and the owner's decisions (PR #241):      *)
(*   line 1        an agent sees an example's output, stderr, exit code    *)
(*                 and verdict, as hyp would record them, on H0 or on an   *)
(*                 uncommitted change, without a commit, without recording *)
(*                 anything in the log, and without using a Hypothesis     *)
(*                 (TryShowsBeforeCommit, ShowsWhatPlaceRecords,           *)
(*                 ShowsWhatRunRecords, TryLeavesLogAlone, TryUsesNoRun)   *)
(*   decision 3    the command is hyp try G LINE/EXAMPLE: the folder as    *)
(*                 it is on disk, on H0; with --worktree, on this          *)
(*                 checkout's files; one new read; place and run get no    *)
(*                 --dry-run (ShowsWhatPlaceRecords, ShowsWhatRunRecords,  *)
(*                 PlaceAlwaysRecords, RunAlwaysRecords)                   *)
(*   spec §13      hyp writes no git state in the host (TryWritesNoGitState)*)
(* Decisions 1 (the words) and 2 (one gap line) name the Goal and its line *)
(* and have no behaviour to model.                                         *)
(*                                                                         *)
(* Switches. NoTry is H0: hyp try is an unknown command, exit 2, and shows *)
(* nothing. The others are cheaper builds the Goal's examples rule out.    *)
(* try.cfg sets all FALSE.                                                 *)
(* Left out: exit 2 and exit 3 of try, a selected Goal, a new version of a *)
(* placed example, a deleted or ignored file, and Stable: a Probe here is  *)
(* deterministic.                                                          *)
(***************************************************************************)
EXTENDS Naturals, Sequences

CONSTANTS
    MaxCalls,                  \* a bound on the acts a switch can repeat
    NoTry,                     \* H0: no hyp try
    TryOnWorktreeFiles,        \* try without --worktree runs on the checkout, not H0
    TryReadsCommittedExample,  \* try reads the example from git, not from disk
    WorktreeReadsHead,         \* try --worktree runs on HEAD's commit
    WorktreeSkipsUntracked,    \* try --worktree leaves out new files
    TryRecords,                \* try appends an event to the log
    TryUsesRun,                \* try records a Result, as a run does
    TryWritesGit,              \* try --worktree builds on git stash create or git add
    PlaceDryRun,               \* place gets a --dry-run that records nothing
    RunDryRun                  \* run gets a --dry-run that records nothing

None == "none"
Subjects == [edit : BOOLEAN, new : BOOLEAN]   \* an edited tracked file, a new file
H0 == [edit |-> FALSE, new |-> FALSE]
Out(e, s) == <<e, s>>

VARIABLES
    diskEx,      \* the example folder on disk: None or "e1"
    comEx,       \* the example folder committed: None or "e1"
    diskSubj,    \* the checkout's files
    headSubj,    \* HEAD's subject
    log,         \* the events of the log, in order
    shows,       \* what each try showed
    gitWrites,   \* git objects, refs or index writes hyp made
    placeCalls,  \* place commands run
    runCalls     \* run commands run

vars == <<diskEx, comEx, diskSubj, headSubj, log, shows, gitWrites, placeCalls, runCalls>>

Init ==
    /\ diskEx = None /\ comEx = None
    /\ diskSubj = H0 /\ headSubj = H0
    /\ log = <<>> /\ shows = {}
    /\ gitWrites = 0 /\ placeCalls = 0 /\ runCalls = 0

Events(t) == SelectSeq(log, LAMBDA x : x.type = t)
By(a) == SelectSeq(log, LAMBDA x : x.by = a)
Placed == Len(Events("place")) > 0
RanOn(s) == \E i \in 1..Len(log) : log[i].type = "run" /\ log[i].subj = s

GrillWritesExample ==
    /\ diskEx = None /\ diskEx' = "e1"
    /\ UNCHANGED <<comEx, diskSubj, headSubj, log, shows, gitWrites, placeCalls, runCalls>>

GrillCommitsExample ==
    /\ diskEx # None /\ comEx = None /\ comEx' = diskEx
    /\ UNCHANGED <<diskEx, diskSubj, headSubj, log, shows, gitWrites, placeCalls, runCalls>>

AgentEditsFile ==
    /\ ~diskSubj.edit /\ diskSubj' = [diskSubj EXCEPT !.edit = TRUE]
    /\ UNCHANGED <<diskEx, comEx, headSubj, log, shows, gitWrites, placeCalls, runCalls>>

AgentAddsFile ==
    /\ ~diskSubj.new /\ diskSubj' = [diskSubj EXCEPT !.new = TRUE]
    /\ UNCHANGED <<diskEx, comEx, headSubj, log, shows, gitWrites, placeCalls, runCalls>>

AgentCommitsChange ==
    /\ headSubj # diskSubj /\ headSubj' = diskSubj
    /\ UNCHANGED <<diskEx, comEx, diskSubj, log, shows, gitWrites, placeCalls, runCalls>>

\* What try writes besides what it prints: nothing, unless a switch says so.
TryWrites(e, s) ==
    /\ log' = IF TryRecords THEN Append(log, [type |-> "try", by |-> "try", ex |-> e, subj |-> s, out |-> Out(e, s)])
              ELSE IF TryUsesRun THEN Append(log, [type |-> "run", by |-> "try", ex |-> e, subj |-> s, out |-> Out(e, s)])
              ELSE log

\* The example try runs: the folder on disk, or the committed one.
TryEx == IF TryReadsCommittedExample THEN comEx ELSE diskEx
Shown(on, s) ==
    [on |-> on, ex |-> diskEx, subj |-> IF on = "H0" THEN H0 ELSE diskSubj,
     out |-> IF NoTry \/ TryEx = None THEN None ELSE Out(TryEx, s)]

AgentTriesOnH0 ==
    /\ diskEx # None /\ Len(log) < MaxCalls
    /\ LET s == IF TryOnWorktreeFiles THEN diskSubj ELSE H0 IN
       /\ shows' = shows \cup {Shown("H0", s)}
       /\ TryWrites(diskEx, s)
    /\ gitWrites' = gitWrites
    /\ UNCHANGED <<diskEx, comEx, diskSubj, headSubj, placeCalls, runCalls>>

AgentTriesOnWorktree ==
    /\ diskEx # None /\ Len(log) < MaxCalls /\ gitWrites < MaxCalls
    /\ LET s == IF WorktreeReadsHead THEN headSubj
                ELSE IF WorktreeSkipsUntracked THEN [diskSubj EXCEPT !.new = FALSE]
                ELSE diskSubj IN
       /\ shows' = shows \cup {Shown("worktree", s)}
       /\ TryWrites(diskEx, s)
    /\ gitWrites' = IF TryWritesGit THEN gitWrites + 1 ELSE gitWrites
    /\ UNCHANGED <<diskEx, comEx, diskSubj, headSubj, placeCalls, runCalls>>

GrillPlaces ==
    /\ comEx # None /\ ~Placed
    /\ log' = Append(log, [type |-> "place", by |-> "place", ex |-> comEx, subj |-> H0, out |-> Out(comEx, H0)])
    /\ placeCalls' = placeCalls + 1
    /\ UNCHANGED <<diskEx, comEx, diskSubj, headSubj, shows, gitWrites, runCalls>>

GrillPlacesDry ==
    /\ PlaceDryRun /\ comEx # None /\ placeCalls < MaxCalls
    /\ placeCalls' = placeCalls + 1
    /\ UNCHANGED <<diskEx, comEx, diskSubj, headSubj, log, shows, gitWrites, runCalls>>

RunnerRuns ==
    /\ Placed /\ ~RanOn(headSubj) /\ headSubj # H0
    /\ log' = Append(log, [type |-> "run", by |-> "run", ex |-> comEx, subj |-> headSubj, out |-> Out(comEx, headSubj)])
    /\ runCalls' = runCalls + 1
    /\ UNCHANGED <<diskEx, comEx, diskSubj, headSubj, shows, gitWrites, placeCalls>>

RunnerRunsDry ==
    /\ RunDryRun /\ Placed /\ runCalls < MaxCalls
    /\ runCalls' = runCalls + 1
    /\ UNCHANGED <<diskEx, comEx, diskSubj, headSubj, log, shows, gitWrites, placeCalls>>

\* The bound of a switch's repeated acts is reached; nothing more happens.
Done == Len(log) >= MaxCalls /\ UNCHANGED vars

Next ==
    \/ Done
    \/ GrillWritesExample \/ GrillCommitsExample
    \/ AgentEditsFile \/ AgentAddsFile \/ AgentCommitsChange
    \/ AgentTriesOnH0 \/ AgentTriesOnWorktree
    \/ GrillPlaces \/ GrillPlacesDry
    \/ RunnerRuns \/ RunnerRunsDry

Spec == Init /\ [][Next]_vars

TypeOK ==
    /\ diskEx \in {None, "e1"} /\ comEx \in {None, "e1"}
    /\ diskSubj \in Subjects /\ headSubj \in Subjects
    /\ gitWrites \in Nat /\ placeCalls \in Nat /\ runCalls \in Nat

\* line 1: an agent sees the output without a commit; with the folder not
\* committed, every try shows an output.
TryShowsBeforeCommit == \A s \in shows : s.out # None

\* line 1, decision 3: try on H0 shows what place records for that folder.
ShowsWhatPlaceRecords ==
    \A s \in shows, i \in 1..Len(log) :
        (s.on = "H0" /\ log[i].type = "place" /\ log[i].ex = s.ex) => s.out = log[i].out

\* line 1, decision 3: try --worktree shows what run records once those
\* files are committed.
ShowsWhatRunRecords ==
    \A s \in shows, i \in 1..Len(log) :
        (s.on = "worktree" /\ log[i].type = "run" /\ log[i].ex = s.ex /\ log[i].subj = s.subj)
            => s.out = log[i].out

\* line 1: try records nothing in the log; the log holds only acts.
TryLeavesLogAlone == \A i \in 1..Len(log) : log[i].type \in {"place", "run"}

\* line 1: try uses no Hypothesis; every Result is a run's.
TryUsesNoRun == \A i \in 1..Len(log) : log[i].type = "run" => log[i].by = "run"

\* spec §13: hyp writes no git state.
TryWritesNoGitState == gitWrites = 0

\* decision 3: place and run get no --dry-run; each one records.
PlaceAlwaysRecords == Len(By("place")) = placeCalls
RunAlwaysRecords == Len(By("run")) = runCalls
=============================================================================
