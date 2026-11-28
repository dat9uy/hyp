# Session lessons

Two sessions wrote this skill. Both redesigned hyp, a tool that stores an
investigation as rival Hypotheses and Experiments: the first on 2026-11-16, the
second on 2026-11-17. Each lesson names what the agent proposed, what the owner
said, and the rule that followed.

## The prior is H0 and the lines

- **Proposed:** a stored Goal kind, discovery or build.
- **Owner said:** "Choosing between a coarse and a normal Goal is not a gap, it is
  a trade-off in what the owner knows. To discover, they give up what they think
  is right and set a very small prerequisite. This is inspired by Bayes."
- **Rule:** the size of the prior is a choice. A Goal with no guard line is a
  discovery; more than three gap lines without a guard is a warning, because it
  shows the owner already knows a lot.

## A check must be able to give both answers

- **Proposed:** each line needs a "hand-made case" as its known pass, beside H0.
- **Owner asked:** "What does a hand-made case look like? Is it a Hypothesis? If
  not, how does the Experiment run on it?" It turned out to be a frozen output
  that only the scoring part reads, which cannot exist for a test of code.
- **Owner's model:** every new version runs on H0 and the candidate. A real run
  did that: `red-on-base.log` and `green-on-commit.log`.
- **Rule:** the baseline shows the fail before the run; the first met subject
  shows the pass. Do not invent a third subject to fill the slot.

## The same bytes teach nothing new

- **Proposed:** a new version of an Experiment makes every cell "not run".
- **Event:** a run's 10 tests passed on one commit; a reviewer then added 4. The
  rule would have rerun the 10 on the same bytes.
- **Rule:** a Result records a verdict per case; an unchanged case keeps its
  verdict on the same subject. Only new cases run.

## One case, one verdict

- **Proposed:** a sampled Experiment, "5 runs, met when 4 pass".
- **Owner said:** "I don't like the '3 of 5' line of reasoning at all since the
  beginning of this project, since I don't learn anything from that kind of
  reasoning."
- **Rule:** each case is one concrete scenario with one verdict. Every new case
  runs twice; two verdicts on the same bytes mean it is refused as unstable.

## No ranking between rivals

- **Proposed:** a row is "unbeaten" unless another row dominates it.
- **Owner said:** "Thinking of it as 'unbeaten' introduces a scoring concept we
  have to prevent. Sometimes the agent needs to try again, and H2' may be worse
  than H2. We should not limit the behaviour of /goal for scoring that is not
  related to the Goal at all."
- **Rule:** a claim is read against its own lines only.

## The reviewer tests the instrument, not the claim

- **Owner said:** in the old design the runner assigned a reviewer's gaps to a
  Hypothesis, "a lot of prose and judgement on top of each other".
- **Rule:** the reviewer writes a stored check plus the cell it disagrees with.
  hyp runs the check itself. The owner places the case in plain words, and only
  that act adds it to the evidence. The update stays mechanical.
- **Owner said, later (2026-11-17):** a reviewer that writes cases invents
  wants. The rule above is replaced by "Evidence never creates the prior" and
  the lessons of 2026-11-17 below.

## A loop needs named stops

- **Owner said:** the old design stopped when an item waited for the owner and
  after five Results; the redesign must keep both.
- **Rule:** stops are waits for the owner, done, a budget of Hypotheses, and no
  honest next step. Bayes updating never ends by itself.

## 2026-11-17: the reviewer, the case and the translation

The talk started from one open item: a wrong Probe, the test code of a line,
can fail a good Hypothesis before any review opens.

### A patch is not a place on the map

- **Proposed:** key the review to the line, valid while the bytes it read stay.
- **Owner said:** "I don't know we need a key anymore, since we have Case
  already", then "your suggestion is just ad-hoc on top of the handoff, not
  think in how it's fit in architecture".
- **Rule:** find where each slot lives on the system's map before choosing a
  mechanism. The review was a stored state on a Hypothesis, the one column that
  stores no state.

### One table held the prior and the miss

- **Proposed:** the review is the reviewer's cases, beside the owner's job
  stories.
- **Owner said:** "I still feel uneasy about thinking the Case as JTBD, and a
  Case as where review lives, this is a problem about the concept".
- **Rule:** one slot, one concept. The owner's want and an agent's attack in
  one table mix the prior with the evidence.

### A reviewer invents wants

- **Owner said:** the reviewer "could just invent a JTBD without any signal that
  it 'meets the Goal'". Cases exist "for filling/expaining the Goal line", "Not
  to achive the 'correct' in some evaluation that reviewer invented themselves."
- **Rule:** evidence never creates the prior. Only the owner adds a want.

### Reading words is not a check

- **Proposed:** a reviewer checks that the Probe reads each case "the way its
  words say".
- **Owner said:** "I still finding 'Does the Probe read each case the way its
  words say?' sounds wrong, but I don't know why". Then: "(1) it's agent
  problem but (2) it's my problem. since I don't think I have bandwith to judge
  the code."
- **Rule:** words allow many readings, so the check is only another agent's
  reading. A verdict is a fact; what it means is the owner's. The owner judges
  no code. Find the hidden translation and make it plain data.

### The case was the wrong abstraction

- **Owner said:** "Now I think even Case is the wrong abstraction", and the
  owner can judge a short plain output: "test table, technically speaking".
  Then: "specification by example sounds like what I expected".
- **Rule:** a case joined the owner's want and an agent's translation in one
  row. Specification by example splits them: examples the owner reads, code
  that records outputs, and a fixed comparison by the tool.

### No judge as the fallback

- **Proposed:** the owner labels real outputs when a want has no plain expected
  output.
- **Owner asked:** "Why you suggest B as fallback to A?"
- **Rule:** outputs that resist a fixed rule vary on every run, so labels
  rarely match and every run returns to the owner. An example with no
  observable outcome is unfinished; sharpen the words.

### The owner decides better from consequences

- **Owner said:** "If the grill process required too fine detail just to get
  started, the human become the bottleneck again, which defeat the purpose of
  agentic coding in the first place." And "if I have to make too much decision
  first hand, my decision as an owner would be low quality".
- **Event:** in an earlier grill the owner approved every agent recommendation
  over two rounds and five read-backs.
- **Rule:** keep the up-front prior to key examples. Add detail when a real
  output shows a gap, as a choice between concrete outputs.
