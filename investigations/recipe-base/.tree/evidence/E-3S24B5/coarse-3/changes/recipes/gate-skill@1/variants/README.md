# Variants

An arm other than `base` is a directory here named for the arm:
`variants/ARM/SKILL-NAME/SKILL.md`, one file per skill the arm changes
(`hyp-prototype`, `hyp-verification`). Launch lays each file over that skill
in the sandbox. Launch refuses an arm with no directory, an empty directory,
or a file for another skill. Add one directory per change under test; this
README is the only example shipped.
