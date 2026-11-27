------------------------------- MODULE finding -------------------------------
\* The kept model of hyp: the one model of its machine. A Goal's model is a
\* change to it. It starts with the acts on one finding; its states and
\* guards are those hyp spec prints under "finding states". A finding's
\* evidence is free text, or also names an example and a commit; hyp finding
\* list prints the walk of that example, the steps of this model its run
\* section takes on the commit.

EXTENDS Sequences

CONSTANTS DeclineFolded, NameRefuted, NameDeclined, AcceptUnheldExample, ListNoWalk,
          ListShowsRefuted, AllHidesRefuted, ListShowsDeclined

VARIABLES st, prev, out, ev, shown, listed

vars == <<st, prev, out, ev, shown, listed>>

States == {"none", "not folded", "folded", "settled", "ended", "refuted", "declined"}

\* The evidence of the finding: none before it is added, free text only, an
\* example the named commit holds, or an example it does not hold.
Evidence == {"none", "free", "example", "unheld"}

\* What the latest hyp finding list printed of the finding's walk.
Shown == {"none", "walk", "no walk", "nothing"}

\* Whether the latest hyp finding list, since the finding's last act, printed
\* the finding: none before such a list, shown or hidden by the default
\* list, all shown or all hidden by hyp finding list --all.
Listed == {"none", "shown", "hidden", "all shown", "all hidden"}

TypeOK == st \in States /\ prev \in States /\ ev \in Evidence /\ shown \in Shown /\ listed \in Listed

Init == st = "none" /\ prev = "none" /\ out = "none" /\ ev = "none" /\ shown = "none" /\ listed = "none"

Go(to, print) == st' = to /\ prev' = st /\ out' = print /\ listed' = "none" /\ UNCHANGED <<ev, shown>>

Record(e) == st' = "not folded" /\ prev' = st /\ out' = "recorded F-1" /\ ev' = e /\ listed' = "none" /\ UNCHANGED shown

\* command: hyp finding add
FindingAdd ==
    /\ st = "none"
    /\ \/ Record("free")
       \/ Record("example")
       \/ AcceptUnheldExample /\ Record("unheld")
       \/ /\ ~AcceptUnheldExample
          /\ out' = "refused: the commit does not hold the example"
          /\ UNCHANGED <<st, prev, ev, shown, listed>>

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

WalkShown ==
    IF ev = "free" THEN "no walk"
    ELSE IF ListNoWalk THEN "nothing"
    ELSE "walk"

\* The states the default list leaves out.
HiddenByDefault ==
    {"settled", "ended"}
    \cup (IF ListShowsRefuted THEN {} ELSE {"refuted"})
    \cup (IF ListShowsDeclined THEN {} ELSE {"declined"})

\* command: hyp finding list
FindingList ==
    /\ st # "none"
    /\ st' = st /\ prev' = st /\ ev' = ev
    /\ IF st \in HiddenByDefault
       THEN /\ listed' = "hidden"
            /\ shown' = shown
            /\ out' = "1 finding " \o st \o " not shown"
       ELSE /\ listed' = "shown"
            /\ shown' = WalkShown
            /\ out' = "F-1 " \o st \o " " \o WalkShown

\* command: hyp finding list --all
FindingListAll ==
    /\ st # "none"
    /\ st' = st /\ prev' = st /\ ev' = ev
    /\ IF AllHidesRefuted /\ st = "refuted"
       THEN /\ listed' = "all hidden"
            /\ shown' = shown
            /\ out' = "1 finding refuted not shown"
       ELSE /\ listed' = "all shown"
            /\ shown' = WalkShown
            /\ out' = "F-1 " \o st \o " " \o WalkShown

Next == FindingAdd \/ AddGoal \/ Land \/ Close \/ FindingRefute \/ FindingDecline \/ FindingList \/ FindingListAll

Spec == Init /\ [][Next]_vars

\* source: hyp spec, guard Not folded
DeclineOnlyNotFolded == (st = "declined" /\ prev # "declined") => prev = "not folded"

\* source: hyp spec, guard Not refuted
RefutedTakesNoName == prev = "refuted" => st = "refuted"

\* source: hyp spec, guard Not declined
DeclinedTakesNoName == prev = "declined" => st = "declined"

\* source: G-FZ9E1G:line-1
NoUnheldExample == ev # "unheld"

\* source: G-FZ9E1G:line-2
WalkFollowsEvidence ==
    shown # "none" =>
        /\ (shown = "walk" <=> ev \in {"example", "unheld"})
        /\ (shown = "no walk" <=> ev = "free")

\* source: G-TYA4YB:line-1
DefaultHidesRefuted == (st = "refuted" /\ listed \in {"shown", "hidden"}) => listed = "hidden"

\* source: G-TYA4YB:line-2
AllShowsRefuted == (st = "refuted" /\ listed \in {"all shown", "all hidden"}) => listed = "all shown"

\* source: G-TYA4YB:line-3
DefaultShowsNotFoldedHidesDeclined ==
    /\ (st = "not folded" /\ listed \in {"shown", "hidden"}) => listed = "shown"
    /\ (st = "declined" /\ listed \in {"shown", "hidden"}) => listed = "hidden"
=============================================================================
