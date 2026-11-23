------------------------------ MODULE status ------------------------------
(***************************************************************************)
(* Goal G-MH24W6: "hyp status with no Goal lists only the Goals in work,   *)
(* and says how many ended Goals it hid; hyp status --all lists every      *)
(* Goal."                                                                  *)
(*                                                                         *)
(* A store holds the Goals in Goals, each written, open, selected, landed  *)
(* or closed. Landed and closed are the two ends. Replay can refuse an     *)
(* event of an ended Goal, and an events file can hold a line hyp cannot   *)
(* read. The owner reads hyp status, hyp status --all, hyp status G,       *)
(* hyp help and hyp spec.                                                  *)
(*                                                                         *)
(* Variants, each FALSE in status.cfg, are cheaper builds:                *)
(*   ListsEnded       status lists every Goal and no count line (H0)       *)
(*   NoCountLine      status hides ended Goals and prints no count line    *)
(*   NoInWorkLine     status prints no "no Goal is in work" line           *)
(*   AllIgnored       status --all hides ended Goals as status does        *)
(*   UsageNoAll       hyp help and hyp spec do not name --all              *)
(*   HidesSelected    status hides selected Goals with the ended ones      *)
(*   NamedShort       status G of an ended Goal prints its header only     *)
(*   HidesRefused     status lists only the refused events of shown Goals  *)
(*   CountAlways      status prints the count line with 0 ended Goals      *)
(***************************************************************************)
EXTENDS Integers, FiniteSets

CONSTANTS Goals,
    ListsEnded, NoCountLine, NoInWorkLine, AllIgnored, UsageNoAll,
    HidesSelected, NamedShort, HidesRefused, CountAlways

VARIABLES
    g,        \* per Goal: "written", "open", "selected", "landed", "closed"
    refused,  \* replay refused an event of an ended Goal
    unread,   \* an events file holds a line hyp cannot read
    out       \* what the last command printed

vars == <<g, refused, unread, out>>

Ended == {"landed", "closed"}
InWork == {"written", "open", "selected"}
NoGoal == [i \in Goals |-> FALSE]
EndedCount == Cardinality({i \in Goals : g[i] \in Ended})
InWorkCount == Cardinality({i \in Goals : g[i] \in InWork})

Act(cmd, i) == [cmd |-> cmd, arg |-> i, shown |-> NoGoal, count |-> 0,
                countLine |-> FALSE, noWork |-> FALSE, refusedShown |-> FALSE,
                unreadShown |-> FALSE, full |-> FALSE, usageAll |-> FALSE]

TypeOK ==
    /\ g \in [Goals -> InWork \cup Ended]
    /\ refused \in BOOLEAN /\ unread \in BOOLEAN

Init ==
    /\ g = [i \in Goals |-> "written"]
    /\ refused = FALSE /\ unread = FALSE
    /\ out = Act("none", 0)

Move(i, from, to, cmd) ==
    /\ g[i] \in from
    /\ g' = [g EXCEPT ![i] = to]
    /\ out' = Act(cmd, i)
    /\ UNCHANGED <<refused, unread>>

GrillPlaces(i)   == Move(i, {"written"}, "open", "place")
RunnerSelects(i) == Move(i, {"open"}, "selected", "select")
OwnerLands(i)    == Move(i, {"selected"}, "landed", "land")
OwnerCloses(i)   == Move(i, InWork, "closed", "close")

\* A merged log brings an event that replay refuses, of an ended Goal.
ReplayRefuses ==
    /\ \E i \in Goals : g[i] \in Ended
    /\ ~refused /\ refused' = TRUE
    /\ out' = Act("merge", 0)
    /\ UNCHANGED <<g, unread>>

\* An events file gets a line hyp cannot read.
LineUnread ==
    /\ ~unread /\ unread' = TRUE
    /\ out' = Act("merge", 0)
    /\ UNCHANGED <<g, refused>>

\* Which Goals hyp status lists, with no Goal or with --all.
Hidden(i, all) ==
    /\ ~ListsEnded
    /\ ~(all /\ ~AllIgnored)
    /\ (g[i] \in Ended \/ (HidesSelected /\ g[i] = "selected"))

StatusOut(cmd, all) ==
    LET shown == [i \in Goals |-> ~Hidden(i, all)]
        hidden == Cardinality({i \in Goals : ~shown[i]})
        line == ~ListsEnded /\ ~NoCountLine /\ (hidden > 0 \/ CountAlways)
    IN [cmd |-> cmd, arg |-> 0, shown |-> shown, count |-> hidden,
        countLine |-> line,
        noWork |-> line /\ ~NoInWorkLine /\ hidden > 0 /\ \A i \in Goals : ~shown[i],
        refusedShown |-> refused /\ ~(HidesRefused /\ \A i \in Goals : g[i] \in Ended => ~shown[i]),
        unreadShown |-> unread, full |-> FALSE, usageAll |-> FALSE]

OwnerReadsStatus ==
    /\ out' = StatusOut("status", FALSE)
    /\ UNCHANGED <<g, refused, unread>>

OwnerReadsStatusAll ==
    /\ out' = IF ListsEnded THEN Act("status-all-refused", 0)
              ELSE StatusOut("status-all", TRUE)
    /\ UNCHANGED <<g, refused, unread>>

OwnerReadsGoal(i) ==
    /\ out' = [Act("status-g", i) EXCEPT !.full = ~(NamedShort /\ g[i] \in Ended)]
    /\ UNCHANGED <<g, refused, unread>>

OwnerReadsUsage ==
    /\ \E cmd \in {"help", "spec"} :
         out' = [Act(cmd, 0) EXCEPT !.usageAll = ~(ListsEnded \/ UsageNoAll)]
    /\ UNCHANGED <<g, refused, unread>>

Next ==
    \/ \E i \in Goals : GrillPlaces(i)
    \/ \E i \in Goals : RunnerSelects(i)
    \/ \E i \in Goals : OwnerLands(i)
    \/ \E i \in Goals : OwnerCloses(i)
    \/ ReplayRefuses
    \/ LineUnread
    \/ OwnerReadsStatus
    \/ OwnerReadsStatusAll
    \/ \E i \in Goals : OwnerReadsGoal(i)
    \/ OwnerReadsUsage

Spec == Init /\ [][Next]_vars

\* source: model only
TypeOKRule == TypeOK

\* source: line-1
NoEndedBlock ==
    out.cmd = "status" => \A i \in Goals : g[i] \in Ended => ~out.shown[i]

\* source: line-2
CountsHidden ==
    out.cmd = "status" /\ EndedCount > 0 => out.countLine /\ out.count = EndedCount

\* source: line-2
SaysNoGoalInWork ==
    out.cmd = "status" /\ EndedCount > 0 /\ InWorkCount = 0 => out.noWork

\* source: line-3
AllListsEvery ==
    /\ out.cmd # "status-all-refused"
    /\ out.cmd = "status-all" => \A i \in Goals : out.shown[i]

\* source: line-4
UsageNamesAll == out.cmd \in {"help", "spec"} => out.usageAll

\* source: line-5
InWorkListed ==
    out.cmd = "status" => \A i \in Goals : g[i] \in InWork => out.shown[i]

\* source: line-6
NamedFullBlock == out.cmd = "status-g" => out.full

\* source: line-7
ReplayLinesListed ==
    out.cmd = "status" => out.refusedShown = refused /\ out.unreadShown = unread

\* source: line-8
NoCountWhenNoneEnded ==
    out.cmd = "status" /\ EndedCount = 0 => ~out.countLine
=============================================================================
