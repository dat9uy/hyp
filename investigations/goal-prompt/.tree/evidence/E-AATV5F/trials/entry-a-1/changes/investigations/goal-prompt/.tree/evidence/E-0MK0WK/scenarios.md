# Scenarios for q-where, argued in the session of 2026-11-12 (owner-standin)

Question: where is the /goal prompt generated, by a hyp command that reads the log, or by a skill that writes it from the tree?

S1. Print the prompt ten times for one fixture Goal.
- A command gives ten identical outputs, so a clause-for-clause match with a baseline can be tested.
- A skill's wording varies between runs, so no match is assured.

S2. Print the prompt for a Goal whose body has no valid constraints block.
- A command exits 1 and names what is missing, and prints nothing.
- A skill has no exit code and can write a prompt with no constraints.

S3. Count the Results under an actor for the stop clause.
- A command derives the count from the replayed log.
- A skill counts by reading text.

Outcome argued in the session: a command meets S1 to S3, and a skill fails S1 and S2. No run happened.
