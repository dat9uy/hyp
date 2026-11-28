# Specification by example

Specification by example is a practice in which the people who want a change
and the people who build it agree on concrete examples first. Each example is
an input and the outcome a person would see. The examples then run as the
tests. Gojko Adzic's book *Specification by Example* (Manning, 2011) names the
practice. Ward Cunningham's Fit, the Framework for Integrated Test (about
2002), ran tables of examples that customers wrote; developers wrote the
fixtures that connect a table to the software.

Sources: https://www.manning.com/books/specification-by-example and
https://en.wikipedia.org/wiki/Framework_for_integrated_test

A Bayes reading uses it for one job: to keep the step from the owner's words
to a verdict honest. That step is a translation. When an agent writes it as
code, part of the prior is a reading the owner never stated and cannot check.

## How it fills the slots

| Slot | What specification by example gives |
|---|---|
| **Prior** | The owner's key examples, one per want, in plain data. The translation is now part of the prior the owner can read. |
| **Evidence** | An example run on a real subject. The example states what the pass looks like before any run; the baseline shows the fail. |
| **Update** | The tool compares the actual output with the expected output by a fixed rule. The code that runs an example records the output and gives no verdict. |
| **Miss** | A real output that no example decides. The owner settles it by choosing between concrete outputs; the choice becomes a new example. |
| **Stop** | Unchanged. |

## Rules taken from it

- **Start from key examples.** One example per want is enough to start. Ask
  "What would you do, and what would you see?" and write the answer as one row.
- **An example states an observable outcome.** A fixed rule must compare it,
  such as equals, contains, exit code or a file that exists. For an agent, the
  outcome is usually a plain fact: the command it ran, the file it wrote, the
  text in its last message.
- **An example with no observable outcome is unfinished.** Sharpen the words,
  or move the want to a later Goal. Do not hand it to a judge, agent or model,
  and do not ask the owner to approve every output that varies.
- **Examples grow from real outputs.** When an output falls outside every
  example, or two readings of the words give different outputs, build both and
  let the owner choose. No agent writes a new want on its own idea of "correct".
- **The owner reads examples, never the code that runs them.**

## Where it stops fitting

- A want about quality, such as "the explanation is clear", may have no
  observable outcome. Treat that as a sign the want is not ready to measure.
- Output that varies on every run cannot be an expected output. State a fact
  about it instead, or the check breaks the rule of one verdict per case.
- The cost moves to the owner's first examples. Keep them few; a long list of
  examples at the start makes the owner the bottleneck again.

## The event

The 2026-11-17 hyp session reached this practice after three drafts failed.
`session-lessons.md` gives the steps, in the owner's words.
