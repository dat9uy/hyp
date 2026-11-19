----------------------------- MODULE trace -----------------------------
(***************************************************************************)
(* Where hyp keeps what it printed to an agent, Goal G-3FDQC0: "Take reads *)
(* back out of the log".                                                   *)
(*                                                                         *)
(* One clone has two worktrees, the main checkout and the runner's        *)
(* worktree; a second clone has one. Each worktree has its own files: the  *)
(* log's events and the output files of .tree/. Agents run commands in a   *)
(* worktree: a read, a refusal, a bad argument, a recorded act, or a run   *)
(* that Stable refuses, which is exit 1 and still records its event. An    *)
(* act hashes its Probes' outputs, checks its guards again, and only then  *)
(* writes the outputs and its event. hyp then writes the trace, and only   *)
(* then prints. The owner merges the runner's worktree into the main       *)
(* checkout. Before the change, a read was already recorded in the log:    *)
(* the old read, id 0.                                                     *)
(*                                                                         *)
(* The rules are the Goal's two lines and the owner's three decisions:     *)
(*   line 1      an agent's or the owner's read makes no change in the     *)
(*               repository;                                               *)
(*   line 2      the owner can read exactly what the agent read;           *)
(*   decision 1  the log holds only acts that change a Goal;               *)
(*   decision 2  a read, a refusal or a bad argument leaves nothing; a     *)
(*               recorded act leaves its event and the outputs it names;   *)
(*   decision 3  the owner reads an agent's reads from any worktree of the *)
(*               clone the agent used.                                     *)
(* TraceInLog and WriteOutputsEarly are the two ways the build before this *)
(* change worked; the config sets both FALSE.                             *)
(***************************************************************************)
EXTENDS Naturals, FiniteSets

CONSTANTS
    Actors,             \* the agents and the owner
    MaxCommands,        \* how many commands run in all
    TraceInLog,         \* TRUE: a read event goes into the worktree's log
    WriteOutputsEarly   \* TRUE: an act writes its outputs before it checks again

Worktrees == {"main", "runner-wt", "other-clone"}
CloneOf   == [w \in Worktrees |-> IF w = "other-clone" THEN "c2" ELSE "c1"]
Clones    == {"c1", "c2"}

\* What a command ends as. "act" is recorded; "stable" is exit 1 and recorded.
Kinds     == {"read", "refusal", "bad-argument", "act", "stable"}
LeaveNothing == {"read", "refusal", "bad-argument"}
Recorded     == {"act", "stable"}
GoalActs     == {"goal.add", "example.place", "result.record", "select"}

\* The read an earlier hyp wrote into the log, with its output, id 0.
OldRead == [id |-> 0, type |-> "read", actor |-> "runner", names |-> {0}]

VARIABLES
    repo,     \* repo[w]: the files of worktree w: its events and output files
    trace,    \* trace[c]: the reads in the trace folder of clone c
    got,      \* got[a]: what actor a got: a command id and the clone it ran in
    phase,    \* "idle", "probing", "checked", "outputs written", "ended", "traced"
    cur,      \* the command in flight
    before,   \* the files of its worktree when it started
    pending,  \* the outputs it hashed and has not written
    last,     \* the last command that ended: its kind, its id, and its worktree's files before and after
    n,        \* commands run
    crashes,  \* acts that crashed after their outputs, before their event
    skipped   \* files hyp told git to skip or ignore

vars == <<repo, trace, got, phase, cur, before, pending, last, n, crashes, skipped>>

NoCommand == [w |-> "main", actor |-> "runner", kind |-> "none", id |-> 0]
NoLast    == [kind |-> "none", id |-> 0, before |-> [ev |-> {}, out |-> {}], after |-> [ev |-> {}, out |-> {}]]

Init ==
    /\ repo = [w \in Worktrees |-> [ev |-> {OldRead}, out |-> {0}]]
    /\ trace = [c \in Clones |-> {}]
    /\ got = [a \in Actors |-> IF a = "runner" THEN {[id |-> 0, clone |-> "c1"], [id |-> 0, clone |-> "c2"]} ELSE {}]
    /\ phase = "idle"
    /\ cur = NoCommand
    /\ before = [ev |-> {}, out |-> {}]
    /\ pending = {}
    /\ last = NoLast
    /\ n = 0
    /\ crashes = 0
    /\ skipped = {}

SetRepo(w, files) == repo' = [repo EXCEPT ![w] = files]

(* An actor runs a command in a worktree. A read, a refusal by a guard     *)
(* that needs no Probe, and a bad argument after the actor end at once.    *)
Start(w, a, k) ==
    /\ phase = "idle" /\ n < MaxCommands
    /\ n' = n + 1
    /\ cur' = [w |-> w, actor |-> a, kind |-> k, id |-> n + 1]
    /\ before' = repo[w]
    /\ phase' = IF k \in Recorded THEN "probing" ELSE "ended"
    /\ UNCHANGED <<repo, trace, got, pending, last, crashes, skipped>>

(* The Probes end: hyp hashes their outputs.                               *)
ProbesEnd ==
    /\ phase = "probing"
    /\ pending' = {cur.id}
    /\ IF WriteOutputsEarly
         THEN SetRepo(cur.w, [repo[cur.w] EXCEPT !.out = @ \cup {cur.id}])
         ELSE UNCHANGED repo
    /\ phase' = "checked"
    /\ UNCHANGED <<trace, got, cur, before, last, n, crashes, skipped>>

(* A guard checked again under the lock refuses: another act landed while *)
(* the Probes ran.                                                         *)
RecheckRefuses ==
    /\ phase = "checked"
    /\ cur' = [cur EXCEPT !.kind = "refusal"]
    /\ pending' = {} /\ phase' = "ended"
    /\ UNCHANGED <<repo, trace, got, before, last, n, crashes, skipped>>

(* Every guard holds again: hyp writes the outputs the event will name.   *)
OutputsWritten ==
    /\ phase = "checked"
    /\ SetRepo(cur.w, [repo[cur.w] EXCEPT !.out = @ \cup pending])
    /\ phase' = "outputs written"
    /\ UNCHANGED <<trace, got, cur, before, pending, last, n, crashes, skipped>>

EventWritten ==
    /\ phase = "outputs written"
    /\ SetRepo(cur.w, [repo[cur.w] EXCEPT !.ev = @ \cup
                {[id |-> cur.id, type |-> "result.record", actor |-> cur.actor, names |-> pending]}])
    /\ pending' = {} /\ phase' = "ended"
    /\ UNCHANGED <<trace, got, cur, before, last, n, crashes, skipped>>

(* hyp crashes between the outputs and the event: the outputs stay, named *)
(* by no event.                                                           *)
CrashBeforeEvent ==
    /\ phase = "outputs written"
    /\ crashes' = crashes + 1
    /\ pending' = {} /\ phase' = "idle" /\ cur' = NoCommand
    /\ UNCHANGED <<repo, trace, got, before, last, n, skipped>>

(* hyp records what it will print: in the clone's trace folder, or, the   *)
(* build before this change, as a read event in the worktree's log.       *)
TraceWritten ==
    /\ phase = "ended"
    /\ IF TraceInLog
         THEN /\ SetRepo(cur.w, [ev |-> repo[cur.w].ev \cup
                     {[id |-> cur.id, type |-> "read", actor |-> cur.actor, names |-> {cur.id}]},
                                 out |-> repo[cur.w].out \cup {cur.id}])
              /\ UNCHANGED trace
         ELSE /\ trace' = [trace EXCEPT ![CloneOf[cur.w]] = @ \cup {[id |-> cur.id, actor |-> cur.actor]}]
              /\ UNCHANGED repo
    /\ phase' = "traced"
    /\ UNCHANGED <<got, cur, before, pending, last, n, crashes, skipped>>

Printed ==
    /\ phase = "traced"
    /\ got' = [got EXCEPT ![cur.actor] = @ \cup {[id |-> cur.id, clone |-> CloneOf[cur.w]]}]
    /\ last' = [kind |-> cur.kind, id |-> cur.id, before |-> before, after |-> repo[cur.w]]
    /\ phase' = "idle" /\ cur' = NoCommand
    /\ UNCHANGED <<repo, trace, before, pending, n, crashes, skipped>>

(* The owner merges the runner's worktree into the main checkout.         *)
OwnerMerges ==
    /\ phase = "idle"
    /\ repo["runner-wt"] # repo["main"]
    /\ SetRepo("main", [ev |-> repo["main"].ev \cup repo["runner-wt"].ev,
                        out |-> repo["main"].out \cup repo["runner-wt"].out])
    /\ last' = NoLast
    /\ UNCHANGED <<trace, got, phase, cur, before, pending, n, crashes, skipped>>

Done ==
    /\ phase = "idle" /\ n = MaxCommands
    /\ UNCHANGED vars

Next ==
    \/ \E w \in Worktrees, a \in Actors, k \in Kinds : Start(w, a, k)
    \/ ProbesEnd
    \/ RecheckRefuses
    \/ OutputsWritten
    \/ EventWritten
    \/ CrashBeforeEvent
    \/ TraceWritten
    \/ Printed
    \/ OwnerMerges
    \/ Done

Spec == Init /\ [][Next]_vars

-----------------------------------------------------------------------------
(* What the owner's hyp show ACTOR prints in worktree w: the old read     *)
(* events of w's log, and the reads of the trace of w's clone.            *)
Show(w, a) ==
    {e.id : e \in {x \in repo[w].ev : x.type = "read" /\ x.actor = a}}
    \cup {r.id : r \in {x \in trace[CloneOf[w]] : x.actor = a}}

\* What actor a got from commands in clone c.
GotIn(c, a) == {g.id : g \in {x \in got[a] : x.clone = c}}

(* Rules. Each is an INVARIANT of trace.cfg.                              *)

(* Decision 1: the log holds only acts that change a Goal; the one read   *)
(* in it is the old read an earlier hyp wrote.                            *)
LogHoldsOnlyGoalActs ==
    \A w \in Worktrees : \A e \in repo[w].ev :
        e.type \in GoalActs \/ e = OldRead

(* Line 1 and decision 2: a read, a refusal or a bad argument leaves the  *)
(* files of its worktree as they were.                                    *)
ALeaveNothingCommandLeavesNothing ==
    last.kind \in LeaveNothing => last.after = last.before

(* Decision 2: a recorded act, a Stable refusal included, adds its own    *)
(* event and only the outputs that event names.                           *)
AnActLeavesItsEventAndItsOutputs ==
    last.kind \in Recorded =>
        \E e \in last.after.ev :
            /\ e.id = last.id /\ e.type \in GoalActs
            /\ last.after.ev = last.before.ev \cup {e}
            /\ last.after.out \ last.before.out \subseteq e.names

(* Decision 2: every output file is named by an event, or was left by an  *)
(* act that crashed before its event.                                     *)
EveryOutputIsNamedOrCrashed ==
    \A w \in Worktrees : crashes = 0 /\ phase = "idle" =>
        repo[w].out \subseteq UNION {e.names : e \in repo[w].ev}

(* The limit of §13: hyp hides no file from git.                          *)
NothingIsHiddenFromGit == skipped = {}

(* Line 2: once a command has ended, show in any worktree holds exactly   *)
(* what the actor got from commands in that worktree's clone.             *)
ShowHoldsExactlyWhatTheActorGot ==
    phase = "idle" =>
        \A w \in Worktrees, a \in Actors : Show(w, a) = GotIn(CloneOf[w], a)

(* Decision 3: every worktree of one clone shows the same reads, before   *)
(* any merge.                                                             *)
AnyWorktreeOfTheClone ==
    \A w1, w2 \in Worktrees, a \in Actors :
        CloneOf[w1] = CloneOf[w2] => Show(w1, a) = Show(w2, a)

(* The old read in the log still shows, in every worktree.                *)
OldReadStillShows ==
    \A w \in Worktrees : 0 \in Show(w, "runner")

TypeOK ==
    /\ phase \in {"idle", "probing", "checked", "outputs written", "ended", "traced"}
    /\ n \in 0..MaxCommands
    /\ crashes \in 0..MaxCommands

=============================================================================
