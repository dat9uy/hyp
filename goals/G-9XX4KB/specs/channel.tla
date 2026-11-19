---------------------------- MODULE channel ----------------------------
(***************************************************************************)
(* The channel of hyp v2, Goal G-9XX4KB: what an agent gets from hyp, and  *)
(* what the owner's `hyp show <agent>` holds of it.                        *)
(*                                                                         *)
(* One agent, the runner, runs commands. Each command ends with bytes the  *)
(* channel wrote: a result that prints the state's words, or a refusal     *)
(* that words a failure code. hyp holds the bytes back until it has        *)
(* written the trace, the read event, and only then prints them. The       *)
(* owner changes the log between the runner's reads. hyp can crash before  *)
(* or after the trace, and the trace can fail to write.                    *)
(*                                                                         *)
(* The rules are the Goal's three lines:                                   *)
(*   line 1  the channel is the one block that writes text for the agent;  *)
(*   line 2  the owner can read exactly what the agent read;               *)
(*   line 3  hyp tells the agent nothing its state does not hold.          *)
(***************************************************************************)
EXTENDS Naturals, Sequences

CONSTANTS
    Words,        \* the words an example can hold; one holds a quote
    Codes,        \* the failure codes a part of hyp can raise
    MaxCommands,  \* how many commands the runner runs
    MaxChanges    \* how many times the owner changes the log

VARIABLES
    version,   \* the log's version; each change of the owner moves it on
    wordsAt,   \* wordsAt[v]: the words the state held at version v
    phase,     \* the runner's command: "idle", "held back" or "traced"
    pending,   \* the bytes the channel wrote for the command in flight
    trace,     \* the runner's read events: the bytes hyp recorded, oldest first
    got,       \* the bytes the runner got, oldest first
    crashes,   \* commands that crashed after their trace, before their print
    commands   \* commands the runner ran

vars == <<version, wordsAt, phase, pending, trace, got, crashes, commands>>

(* Bytes are a record. `writer` is the block of code that wrote them; a   *)
(* result or a refusal names the version it was written from and the      *)
(* words it prints; a notice says only that hyp could not record.         *)
Bytes(kind, code, v, w) ==
    [kind |-> kind, writer |-> "channel", code |-> code, version |-> v, words |-> w]

\* The channel prints a value as the state holds it: no quote is added.
Result(v)     == Bytes("result", "none", v, wordsAt[v])
Refusal(c, v) == Bytes("refusal", c, v, wordsAt[v])
Notice        == Bytes("notice", "not-recorded", 0, "none")
Nothing       == Bytes("none", "none", 0, "none")

\* hyp show runner prints the read events as recorded: never a fresh read.
ShowText == trace

Results(s) == SelectSeq(s, LAMBDA b : b.kind # "notice")

RECURSIVE IsSubSeq(_, _)
IsSubSeq(s, t) ==
    IF s = <<>> THEN TRUE
    ELSE IF t = <<>> THEN FALSE
    ELSE IF Head(s) = Head(t) THEN IsSubSeq(Tail(s), Tail(t))
    ELSE IsSubSeq(s, Tail(t))

Init ==
    /\ version = 0
    /\ wordsAt \in [{0} -> Words]
    /\ phase = "idle"
    /\ pending = Nothing
    /\ trace = <<>>
    /\ got = <<>>
    /\ crashes = 0
    /\ commands = 0

(* The runner runs hyp status: the channel writes the state's words.      *)
RunnerReads ==
    /\ phase = "idle" /\ commands < MaxCommands
    /\ phase' = "held back" /\ pending' = Result(version)
    /\ commands' = commands + 1
    /\ UNCHANGED <<version, wordsAt, trace, got, crashes>>

(* A part of hyp refuses the runner's act with a code and a value; the    *)
(* channel words it.                                                      *)
RunnerIsRefused(c) ==
    /\ phase = "idle" /\ commands < MaxCommands
    /\ phase' = "held back" /\ pending' = Refusal(c, version)
    /\ commands' = commands + 1
    /\ UNCHANGED <<version, wordsAt, trace, got, crashes>>

(* hyp writes the output files and the read event, before any print.      *)
TraceWritten ==
    /\ phase = "held back"
    /\ trace' = Append(trace, pending) /\ phase' = "traced"
    /\ UNCHANGED <<version, wordsAt, pending, got, crashes, commands>>

(* The store cannot be written: hyp prints none of the command's bytes,   *)
(* only a notice that it could not record them, and exits 3.              *)
TraceFails ==
    /\ phase = "held back"
    /\ got' = Append(got, Notice) /\ phase' = "idle" /\ pending' = Nothing
    /\ UNCHANGED <<version, wordsAt, trace, crashes, commands>>

Printed ==
    /\ phase = "traced"
    /\ got' = Append(got, pending) /\ phase' = "idle" /\ pending' = Nothing
    /\ UNCHANGED <<version, wordsAt, trace, crashes, commands>>

CrashBeforeTrace ==
    /\ phase = "held back"
    /\ phase' = "idle" /\ pending' = Nothing
    /\ UNCHANGED <<version, wordsAt, trace, got, crashes, commands>>

CrashAfterTrace ==
    /\ phase = "traced"
    /\ phase' = "idle" /\ pending' = Nothing /\ crashes' = crashes + 1
    /\ UNCHANGED <<version, wordsAt, trace, got, commands>>

(* The owner changes the log: the state now holds other words.            *)
OwnerChangesLog(w) ==
    /\ version < MaxChanges
    /\ version' = version + 1
    /\ wordsAt' = [v \in 0..version + 1 |-> IF v = version + 1 THEN w ELSE wordsAt[v]]
    /\ UNCHANGED <<phase, pending, trace, got, crashes, commands>>

Done ==
    /\ phase = "idle" /\ commands = MaxCommands /\ version = MaxChanges
    /\ UNCHANGED vars

Next ==
    \/ RunnerReads
    \/ \E c \in Codes : RunnerIsRefused(c)
    \/ TraceWritten
    \/ TraceFails
    \/ Printed
    \/ CrashBeforeTrace
    \/ CrashAfterTrace
    \/ \E w \in Words : OwnerChangesLog(w)
    \/ Done

Spec == Init /\ [][Next]_vars

-----------------------------------------------------------------------------
(* Rules. Each is an INVARIANT of channel.cfg.                            *)

AllBytes == {got[i] : i \in 1..Len(got)} \cup {trace[i] : i \in 1..Len(trace)}

(* Line 1: every byte the agent gets, and every byte the owner's show     *)
(* holds, is written by the channel; a refusal is the channel's words for *)
(* a code a part of hyp raised.                                           *)
OnlyTheChannelWrites ==
    \A b \in AllBytes :
        /\ b.writer = "channel"
        /\ b.kind = "refusal" => b.code \in Codes

(* Line 2, as the line says it, in its two halves: once a command has    *)
(* ended, the owner's show holds exactly what the runner got. No less:    *)
(* every byte the runner got is in show, in order.                        *)
ShowLacksNothingTheRunnerGot ==
    phase = "idle" => IsSubSeq(got, ShowText)

(* No more: every read in show is one the runner got. With the half       *)
(* above, show equals what the runner got.                                *)
ShowHoldsNothingTheRunnerDidNotGet ==
    phase = "idle" => IsSubSeq(ShowText, got)

(* Line 2: every result and refusal the runner got is in the owner's      *)
(* show, byte for byte, in the order the runner got them, however the log *)
(* changed after.                                                         *)
ShowHoldsEveryResultTheRunnerGot ==
    IsSubSeq(Results(got), ShowText)

(* Line 2: a read in show that the runner never got is a command that     *)
(* crashed after its trace, or the one printing now.                      *)
ShowHoldsNoOtherRead ==
    Len(ShowText) = Len(Results(got)) + crashes + (IF phase = "traced" THEN 1 ELSE 0)

(* Line 2: the one thing the runner can get that show lacks is the notice *)
(* that hyp could not record, and the notice holds nothing of the state.  *)
ANoticeHoldsNoState ==
    \A i \in 1..Len(got) : got[i].kind = "notice" => got[i].words = "none"

(* Line 3: every result and refusal, as the runner got it and as show     *)
(* holds it, prints the words the state held when it was written.         *)
ValuesAsHeld ==
    \A b \in AllBytes :
        b.kind \in {"result", "refusal"} => b.words = wordsAt[b.version]

TypeOK ==
    /\ version \in 0..MaxChanges
    /\ phase \in {"idle", "held back", "traced"}
    /\ crashes \in 0..MaxCommands
    /\ commands \in 0..MaxCommands
    /\ Len(trace) <= commands /\ Len(got) <= commands

=============================================================================
