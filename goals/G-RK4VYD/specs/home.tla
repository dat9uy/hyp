------------------------------ MODULE home ------------------------------
(***************************************************************************)
(* Where hyp keeps the trace, and how the owner cleans it, Goal G-RK4VYD:  *)
(* "Keep the trace in hyp's own folder, and let the owner clean it".       *)
(*                                                                         *)
(* Clone c1 has two worktrees, the main checkout and the runner's          *)
(* worktree; clone c2 has one. An agent reads in a worktree of a clone     *)
(* that exists: hyp records the read in a trace folder and the agent gets  *)
(* it. $XDG_STATE_HOME may be set or unset between commands. A clone can   *)
(* be deleted. From a worktree of a clone that exists the owner runs       *)
(* hyp show, hyp prune, or hyp prune --remove. Under ~/.local the owner    *)
(* keeps an old trace folder that an earlier hyp wrote.                    *)
(*                                                                         *)
(* The rules are the Goal's three lines and the owner's decisions:         *)
(*   line 1 / decision 1  the trace is under ~/.hyp; hyp never uses        *)
(*                        $XDG_STATE_HOME and writes nothing under ~/.local *)
(*   line 2 / decision 2  hyp prune lists the traces of clones that no     *)
(*                        longer exist and removes nothing; --remove       *)
(*                        removes only those; a live clone's trace is      *)
(*                        never listed or removed                          *)
(*   decision 3           hyp never reads, moves or removes the old folder *)
(*                        under ~/.local                                   *)
(*   line 3 / decision 4  the owner reads an agent's reads from any        *)
(*                        worktree of the clone, before any merge          *)
(*                                                                         *)
(* Switches. UseXDG and NoPrune are how H0 worked. PruneListsAll,          *)
(* PruneRemovesAlways, RemoveDeletesAll and PruneAlsoOld are cheaper       *)
(* builds the Goal's examples rule out. home.cfg sets all FALSE.           *)
(* Left out: prune's own read (it lands in the live clone's trace, which   *)
(* prune never lists), and a new clone at a deleted one's path, which the  *)
(* owner accepted as showing the old reads.                                *)
(***************************************************************************)
EXTENDS Naturals, FiniteSets

CONSTANTS
    MaxReads,            \* how many reads run in all
    UseXDG,              \* H0: the trace is under $XDG_STATE_HOME, else ~/.local/state
    NoPrune,             \* H0: hyp prune is an unknown command, exit 2
    PruneListsAll,       \* a prune that lists every trace, live ones too
    PruneRemovesAlways,  \* a prune that removes what it lists without --remove
    RemoveDeletesAll,    \* a --remove that removes every trace
    PruneAlsoOld         \* a prune that also cleans the old folder under ~/.local

Clones    == {"c1", "c2"}
Worktrees == {"c1-main", "c1-runner", "c2-main"}
CloneOf   == [w \in Worktrees |-> IF w = "c2-main" THEN "c2" ELSE "c1"]
Places    == {"hyp", "local", "xdg"}   \* ~/.hyp, ~/.local/state, $XDG_STATE_HOME
NoTrace   == [p \in Places |-> [c \in Clones |-> {}]]

VARIABLES
    trace,        \* trace[p][c]: the reads in clone c's trace folder at place p
    got,          \* got[c]: the reads agents got from commands in clone c
    live,         \* the clones that exist
    xdg,          \* TRUE when $XDG_STATE_HOME is set to an absolute path
    old,          \* the old trace folder under ~/.local: "as left" or "touched"
    n,            \* reads run
    last,         \* the owner's last command: "none", "show", "prune", "remove"
    exit,         \* its exit code
    listed,       \* the clones whose traces it named
    before,       \* the trace when it started
    removedLive   \* TRUE once a prune removed a live clone's trace

vars == <<trace, got, live, xdg, old, n, last, exit, listed, before, removedLive>>

\* Where this build keeps a clone's trace, given the environment now.
Place == IF UseXDG THEN (IF xdg THEN "xdg" ELSE "local") ELSE "hyp"

\* What hyp show prints in worktree w: the trace of w's clone where this
\* build looks for it now.
Show(w) == trace[Place][CloneOf[w]]

\* The clones that have a trace folder under ~/.hyp.
Traced == {c \in Clones : trace["hyp"][c] # {}}

Init ==
    /\ trace = NoTrace
    /\ got = [c \in Clones |-> {}]
    /\ live = Clones
    /\ xdg \in BOOLEAN
    /\ old = "as left"
    /\ n = 0
    /\ last = "none" /\ exit = 0 /\ listed = {} /\ before = NoTrace
    /\ removedLive = FALSE

(* An agent reads in worktree w; hyp records the read, then prints it.    *)
AgentReads(w) ==
    /\ CloneOf[w] \in live /\ n < MaxReads
    /\ n' = n + 1
    /\ trace' = [trace EXCEPT ![Place][CloneOf[w]] = @ \cup {n + 1}]
    /\ got' = [got EXCEPT ![CloneOf[w]] = @ \cup {n + 1}]
    /\ last' = "none" /\ listed' = {}
    /\ UNCHANGED <<live, xdg, old, exit, before, removedLive>>

(* $XDG_STATE_HOME is set or unset between two commands.                  *)
EnvChanges ==
    /\ xdg' = ~xdg
    /\ UNCHANGED <<trace, got, live, old, n, last, exit, listed, before, removedLive>>

(* A clone is deleted, or moved to another path.                          *)
CloneDeleted(c) ==
    /\ c \in live /\ live \ {c} # {}
    /\ live' = live \ {c}
    /\ last' = "none" /\ listed' = {}
    /\ UNCHANGED <<trace, got, xdg, old, n, exit, before, removedLive>>

(* The owner runs hyp show in worktree w. It changes no trace this model  *)
(* keeps; the rules read Show(w) in every state.                          *)
OwnerShows(w) ==
    /\ CloneOf[w] \in live /\ last # "show"
    /\ last' = "show" /\ exit' = 0 /\ listed' = {}
    /\ UNCHANGED <<trace, got, live, xdg, old, n, before, removedLive>>

(* The owner runs hyp prune, or hyp prune --remove, in worktree w.        *)
OwnerPrunes(w, remove) ==
    /\ CloneOf[w] \in live
    /\ last' = IF remove THEN "remove" ELSE "prune"
    /\ before' = trace
    /\ IF NoPrune
         THEN /\ exit' = 2 /\ listed' = {}
              /\ UNCHANGED <<trace, old, removedLive>>
         ELSE LET gone  == IF PruneListsAll THEN Traced ELSE Traced \ live
                  drop  == IF ~(remove \/ PruneRemovesAlways) THEN {}
                           ELSE IF remove /\ RemoveDeletesAll THEN Traced
                           ELSE gone
              IN /\ exit' = 0 /\ listed' = gone
                 /\ trace' = [trace EXCEPT !["hyp"] =
                                [c \in Clones |-> IF c \in drop THEN {} ELSE trace["hyp"][c]]]
                 /\ removedLive' = (removedLive \/ drop \cap live # {})
                 /\ old' = IF PruneAlsoOld /\ (remove \/ PruneRemovesAlways)
                             THEN "touched" ELSE old
    /\ UNCHANGED <<got, live, xdg, n>>

Next ==
    \/ \E w \in Worktrees : AgentReads(w)
    \/ EnvChanges
    \/ \E c \in Clones : CloneDeleted(c)
    \/ \E w \in Worktrees : OwnerShows(w)
    \/ \E w \in Worktrees, r \in BOOLEAN : OwnerPrunes(w, r)

Spec == Init /\ [][Next]_vars

-----------------------------------------------------------------------------
(* Rules. Each is an INVARIANT of home.cfg.                               *)

(* Line 1, decision 1: every read an agent got in a live clone is in that *)
(* clone's trace under ~/.hyp.                                            *)
TraceIsUnderHyp ==
    \A c \in live : got[c] \subseteq trace["hyp"][c]

(* Decision 1: hyp writes nothing under ~/.local.                         *)
NothingUnderLocal ==
    \A c \in Clones : trace["local"][c] = {}

(* Decision 1: hyp never uses $XDG_STATE_HOME.                            *)
NothingUnderXDG ==
    \A c \in Clones : trace["xdg"][c] = {}

(* Line 1: hyp show reads the trace from ~/.hyp and nowhere else, so once *)
(* ~/.hyp is moved away it no longer holds the read.                      *)
ShowReadsHyp ==
    \A w \in Worktrees : Show(w) = trace["hyp"][CloneOf[w]]

(* Line 2: hyp prune, with or without --remove, names exactly the clones  *)
(* that no longer exist and had a trace.                                  *)
PruneNamesEveryGoneClone ==
    last \in {"prune", "remove"} =>
        /\ exit = 0
        /\ listed = {c \in Clones \ live : before["hyp"][c] # {}}

(* Decision 2: hyp prune removes nothing.                                 *)
PruneRemovesNothing ==
    last = "prune" => trace = before

(* Line 2: after hyp prune --remove no trace of a gone clone is left on   *)
(* the disk.                                                              *)
RemoveLeavesNoGoneTrace ==
    last = "remove" => \A c \in Clones \ live : trace["hyp"][c] = {}

(* Decision 2: a live clone's trace is never listed.                      *)
LiveNeverListed ==
    listed \cap live = {}

(* Decision 2: a live clone's trace is never removed.                     *)
LiveNeverRemoved ==
    removedLive = FALSE

(* Decision 3: hyp never reads, moves or removes the old folder under     *)
(* ~/.local.                                                              *)
OldFolderLeftAlone ==
    old = "as left"

(* Line 3, decision 4: in every worktree of a live clone, hyp show holds  *)
(* exactly what agents got from that clone, before any merge.             *)
ShowFromAnyWorktree ==
    \A w \in Worktrees : CloneOf[w] \in live => Show(w) = got[CloneOf[w]]

TypeOK ==
    /\ live \subseteq Clones /\ live # {}
    /\ n \in 0..MaxReads
    /\ last \in {"none", "show", "prune", "remove"}
    /\ exit \in {0, 2}
    /\ old \in {"as left", "touched"}

=============================================================================
