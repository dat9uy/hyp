---
name: hyp-bayes-thinking
description: "Read a design through Bayes: the prior the owner states, evidence that can give both answers, an update that is derived and not argued, what the evidence misses, and what stops the loop. Keeps the design of a system that tests, measures or improves itself from drifting into checks that cannot fail, ratio scores, rankings, reruns that teach nothing, or objects invented to fill a slot. Use when designing or reviewing how a system checks, measures, reruns or improves its own work, when the owner asks for a Bayes lens or the Bayes philosophy, or inside hyp-rethinking to find breaks and requirements. Not for computing probabilities or statistics, and not for choosing between approaches, which hyp-rethinking owns."
metadata:
  version: 1.1.0
---

# Bayes thinking

A lens for the design of a system that loops on itself: an agent that runs
experiments, a review that feeds its own instrument, a build that tests its own
changes. It produces a short Bayes reading of one concept, with a finding for
each slot that is empty or broken. The owner uses it to see where the loop can
fool itself. The bar: every finding names a concrete part of the system and an
event, and nothing in the reading is Bayes as decoration.

The lens is qualitative. Most such systems compute no probability, and the
reading never claims one. Bayes supplies the questions, not the arithmetic.

## When to use

- The owner asks for a Bayes lens, the Bayes philosophy, or "test it with Bayes".
- A design decides when a check runs, reruns, counts, or stops counting.
- A design adds a reviewer, a retry, a threshold, a score or a new kind of case.
- Inside `hyp-rethinking`: use the slots to derive breaks (its step 3) and the
  requirements (its step 5). The lens adds no fourth approach.
- Not for: statistics or sample-size math; choosing between approaches
  (`hyp-rethinking`); a plain definition of a loose word (answer it directly).

## The five slots

Map the concept onto these slots. Each slot names a part of the system, the
person or process that fills it, and the event that shows it.

| Slot | Question | Healthy when |
|---|---|---|
| **Prior** | What is stated before any run, and by whom? | The owner states it, in words and concrete examples the owner can read. Its size matches what the owner knows: a small prior to discover, a large one to build. |
| **Evidence** | Can each check give both answers? What shows each one? | A baseline shows the check can fail before the run; a real subject shows it can pass during the run. The owner knows what each answer means: an example states the pass before any run. |
| **Update** | How does one result change the state? | The state is derived from stored results by a rule. A verdict comes from a fixed comparison the tool owns, not from code or prose an agent wrote. |
| **Miss** | What does the evidence not test, and who looks? | A miss shows as a real output the examples do not decide. An agent may show it and ask. Only the owner's act adds a want. |
| **Stop** | What ends the loop? | A short, mechanical list: waits for the owner, done, budget, no honest next step. |

A slot with no part of the system to name is a finding. A slot that only fills
with a new invented object is also a finding: report it, do not invent the object.

## Rules that came from real sessions

Each rule removes a failure the owner rejected. `references/session-lessons.md`
gives the event behind each.

- **A check that gives one answer whatever the truth is teaches nothing.** That
  covers a check that passes on everything and one that fails on everything.
- **Use what exists to show both answers.** The baseline shows the fail; the
  first real subject that passes shows the pass. Do not invent a third subject,
  such as a hand-made sample, to make the pair look symmetric. An example's
  input given to the baseline or a real subject is not a third subject.
- **The same check on the same bytes teaches nothing new.** Keep the verdict of
  each unchanged case. A new version is new evidence only for its new cases.
- **One case, one verdict.** A case that gives two verdicts on the same bytes is
  unstable: sharpen it or drop it. Never average it into a ratio such as
  "met when 4 of 5 pass"; the owner learns nothing from a ratio.
- **No ranking between rivals.** Each claim is read against its own lines. A
  score that orders rivals by something the Goal did not ask for limits what the
  loop may try.
- **The size of the prior is a choice about knowledge, not a gap.** Many lines
  in a discovery mean the owner already knows a lot: it is a build.
- **What a discovery learned becomes the next prior.** Lost sandboxes are not
  lost evidence when the result is written as the next Goal's lines.
- **A loop needs a stop that Bayes does not give.** Updating never ends by
  itself. Name the stops before the loop runs.
- **Evidence never creates the prior.** A case or example that an agent writes
  on its own idea of "correct" is evidence that pretends to be the prior. The
  owner's approval does not change that: an approved agent idea is still the
  agent's. Only the owner adds a want; an agent may ask.
- **Find the hidden translation.** Between the owner's words and a verdict,
  someone turns words into something a machine runs. Name who writes it and
  what makes it trusted. An empty answer is a finding. A healthy translation is
  plain data the owner can read: see `references/specification-by-example.md`.
- **A verdict is a fact; what it means is the owner's.** No agent judges
  whether a verdict is what the owner meant, and the owner never judges code.
- **One slot, one concept.** When one stored thing fills two slots, such as
  the owner's want and an agent's attack in one table, prior and evidence can
  not be told apart. Report it.
- **The owner decides better from consequences.** Keep the up-front prior
  small. Add detail when a real output shows a gap, by asking the owner to
  choose between concrete outputs, never to approve an agent's foresight.

## How to work

1. Name the concept and the event that points at it. With no event, say so: the
   reading is then foresight, and each finding is marked that way.
2. Fill the five slots from the system's own docs, code and records. Cite each.
   Include who turns the owner's words into checks, and what makes that trusted.
3. Run the rules above against the filled slots. Each failure is a finding with
   its evidence and the part of the system it touches.
4. Say where the analogy stops matching the system, such as "no probability is
   computed; a cell is met, failed or not run".
5. Give the findings to the owner, or to `hyp-rethinking` as breaks and
   requirements. The owner decides; the reading records no decision.

## Done when

- Every slot names a part of the system, or is listed as a finding.
- Every finding cites an event or is marked foresight.
- No finding is answered by a new term or object the system does not have.
- The reading names who turns the owner's words into checks, and what makes
  that trusted.
- The reading says where the analogy stops.

## Resources

- `references/session-lessons.md`: the event behind each rule, from the
  sessions that wrote this skill. Open it when a rule's reason is unclear or
  disputed.
- `references/specification-by-example.md`: how concrete examples fill the
  Prior, Evidence, Update and Miss slots. Open it when a design must turn the
  owner's words into checks, or a reading finds the hidden translation.
