# G-73R704 Record the shared-store design: what hyp stores and derives when one store holds many Goals — complete

Design record for G1b (investigations/checkpoint-store), which builds it. Each question was settled in session e6a35a24 (2026-11-11) and is recorded as selected with that session as evidence, on the owner's decision. The decisions land in ADR 0020. No lines: no run checks a design argued in a session; G1b's lines carry the cases.

## Q-D906CD What does a Result keep about the commits and files its cited runs were given? — answered

Selected: H-Q7ZRZJ

- H-Q7ZRZJ One subject per cited run, {base, files}; a Hypothesis is covered when one entry holds its whole subject — supported
  - E-DA7082 Two scenarios argued in the session decide which shape keeps the claims true — run (recipe none, rule by reading)
    - R-HW9B05 recorded H-Q7ZRZJ=supports H-X1Z61J=contradicts H-CXH63T=contradicts: S1 and S2 as argued in the session: one subject per run keeps both true; merged lists cover bare B through a patched run (S1); one base per Result leaves red on base with no single checked closer (S2). The owner chose one subject per run.
- H-X1Z61J One base per Result: a Result cites runs at one base only, beside the merged fingerprints — contradicted
- H-CXH63T Merged lists: a list of bases beside the one merged fingerprints list — contradicted

## Q-HC4V9T Where is the commit a build Goal starts from recorded? — answered

Selected: H-KVCHND

- H-KVCHND In the subject of the red-on-base Result; the Goal records no commit — supported
  - E-WBQBPP Two scenarios argued in the session decide where the start commit lives — run (recipe none, rule by reading)
    - R-V8032Z recorded H-KVCHND=supports H-P5BA75=contradicts: S1: a Goal field goes wrong when the base moves; the runs' subjects stay true. S2: only builds would fill it, a delivery kind. The owner agreed the Goal records no start commit, and asked that a Goal's origin be known, which is derived from its cited keys.
- H-P5BA75 On the Goal: a start-commit field written when the Goal is added — contradicted

## Q-H8PJ20 What is the owner's act attached to when one store holds many Goals? — answered

Selected: H-J23DRM

- H-J23DRM To no node: one owner act may move keys of several Goals; each key leads to its Goal — supported
  - E-BC62A9 One owner sitting across two Goals decides what the act is attached to — run (recipe none, rule by reading)
    - R-8CJCV6 recorded H-J23DRM=supports H-379Z1S=contradicts: One sitting across two Goals is one act with no node; each key finds its Goal. Per-Goal acts split one sitting into several events. The owner chose no node; lines moved into the Goal's node.add, so the owner act never names a Goal.
- H-379Z1S To one Goal: each owner act names one Goal, so one sitting across Goals is several events — contradicted

## Q-R21S8R How many event types carry the owner's answers and the runner's checkpoints? — answered

Selected: H-H11D63

- H-H11D63 Two: answer (owner; no node, no why) and checkpoint (runner; node is the Goal, why required) — supported
  - E-YNGH8F Whether the envelope rule can stay per type decides the number of types — run (recipe none, rule by reading)
    - R-AWPQ4X recorded H-H11D63=supports H-MATM0C=contradicts: The owner and runner acts differ in node and why, which the spec defines per type. Two types keep one envelope rule each and make field mixing a decode error; one type makes the envelope depend on the payload. The owner accepted two types, answer and checkpoint.
- H-MATM0C One type, checkpoint, its kind derived from its fields — contradicted
