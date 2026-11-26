------------------------------- MODULE finding -------------------------------
\* The kept model of hyp: the one model of its machine. A Goal's model is a
\* change to it. It starts with the acts on one finding; its states and
\* guards are those hyp spec prints under "finding states".

EXTENDS Sequences

CONSTANTS DeclineFolded, NameRefuted, NameDeclined

VARIABLES st, prev, out

vars == <<st, prev, out>>

States == {"none", "not folded", "folded", "settled", "ended", "refuted", "declined"}

TypeOK == st \in States /\ prev \in States

Init == st = "none" /\ prev = "none" /\ out = "none"

Go(to, print) == st' = to /\ prev' = st /\ out' = print

\* command: hyp finding add
FindingAdd == st = "none" /\ Go("not folded", "recorded F-1")

\* command: hyp add goal
AddGoal ==
    /\ \/ st \in {"not folded", "settled", "ended"}
       \/ NameRefuted /\ st = "refuted"
       \/ NameDeclined /\ st = "declined"
    /\ Go("folded", "recorded G-1")

\* command: hyp land
Land == st = "folded" /\ Go("settled", "landed G-1")

\* command: hyp close
Close == st = "folded" /\ Go("ended", "closed G-1")

\* command: hyp finding refute
FindingRefute == st \in {"not folded", "folded", "settled", "ended"} /\ Go("refuted", "refuted F-1")

\* command: hyp finding decline
FindingDecline ==
    /\ \/ st = "not folded"
       \/ DeclineFolded /\ st = "folded"
    /\ Go("declined", "declined F-1")

\* command: hyp finding list
FindingList == st # "none" /\ Go(st, "F-1 " \o st)

Next == FindingAdd \/ AddGoal \/ Land \/ Close \/ FindingRefute \/ FindingDecline \/ FindingList

Spec == Init /\ [][Next]_vars

\* source: hyp spec, guard Not folded
DeclineOnlyNotFolded == (st = "declined" /\ prev # "declined") => prev = "not folded"

\* source: hyp spec, guard Not refuted
RefutedTakesNoName == prev = "refuted" => st = "refuted"

\* source: hyp spec, guard Not declined
DeclinedTakesNoName == prev = "declined" => st = "declined"
=============================================================================
