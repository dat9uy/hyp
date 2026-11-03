---
name: id-citation-arms
version: 2
from: id-citation-arms@1 3fb1e1395f30d6201091602256a4e5d5261e201ec9ca27cc7178611a60210d7b
format: 1
subject: Rules for write-command text that cites an id that is no node, as Arms built in Launch from the sandbox base (one unpatched, each other a --patch over it), run on a frozen hyp store, and e-arms' verdict rule (self-review)
---

# id-citation-arms

The instrument of the id-citation tree's e-arms (`investigations/id-citation`).
Each Arm is a hyp binary: the sandbox base, or the base with one patch
applied. Launch builds every Arm and names it by sha256; Drive runs every
case of the bundle's cases file with every Arm, each on a fresh copy of the
bundle's fixture store, and writes what each Arm did; `verdict` applies
e-arms' pre-registered rule to that table. The Arms are the subject under
test: `cases` runs only the binaries it is given and never the `hyp` on
PATH.

Create a sandbox with the base as `--base`, one `--patch ARM=FILE` per
patched Arm, and an inputs file naming the unpatched Arm:

```bash
printf 'base_arm=text-refuse\n' > inputs.env
hyp sandbox create ID --host . --base <commit> --recipes . \
  --recipe id-citation-arms@2 --inputs inputs.env \
  --patch today=investigations/id-citation/arms/today.patch \
  --patch text-warn=investigations/id-citation/arms/text-warn.patch \
  --patch cite=investigations/id-citation/arms/cite.patch
```

## Launch

Run: ./launch.sh

Reads `base_arm=NAME` from the inputs file (`--inputs FILE`): the Arm the
unpatched base is. It exits 2 when there is no inputs file or no such line,
when NAME is not an Arm name, or when a patch has NAME as its Arm. It builds
the base, and the base with each `$HYP_SANDBOX_PATCHES/<ARM>.patch` applied
alone, each in two copies of `wt/` at different paths with
`CGO_ENABLED=0 go build -trimpath`, and exits 2 when a patch does not apply,
a build fails or the two copies give different sha256 values. The binaries
go to this recipe's tmp as `arms/arm-<arm>`, and `arms.sha256` names each
by its sha256, one line per Arm sorted by name. `wt/` is not changed.

## Doctor

Run: ./doctor.sh

Passes when `python3` and `go` run, every binary matches `arms.sha256`, and
the Arms are exactly the base Arm plus one per patch.

## Drive

Run: ./drive.sh

Runs `./cases.py`, the script of the Standalone label `cases`, over
`fixture/cases.tsv` and `fixture/fixture.tar` with every Arm in
`arms.sha256`, in that file's order. The table goes to this recipe's tmp as
`table.tsv`, and each step's stderr, the `$HYP_OUT` tree of `cases`, to
`stderr.tar`. Drive the sandbox again to run the cases again.

## Evidence

Kept by hand at remove: `--keep tmp/id-citation-arms/table.tsv --keep
tmp/id-citation-arms/arms.sha256 --keep tmp/id-citation-arms/stderr.tar`.
The patches and the inputs file reach OUT_DIR as `patches/` and `inputs/`.
Then `verdict` runs on the kept table and `fixture/cases.tsv`, as a
Standalone command.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Launch and Drive start no process that outlives
them.

## Standalone

Run cases: ./cases.py
External: the Arm binaries given as NAME=/abs/path (the subject), run on copies of the cases file and fixture; writes only $HYP_OUT

Run verdict: ./verdict.py
External: none; reads only the table and cases file given as arguments

- `cases CASES FIXTURE NAME=/abs/path...` prints a header and one row per
  Arm and case: `arm case exit appended warned`, for the last step that
  ran. `appended` is `y` when the bytes under `fixture/.tree/events`
  changed during that step, and `warned` is `y` when its stderr holds a
  line starting with `warn:`. A case whose last step exits 3 runs once
  more. Each step's stderr goes to `$HYP_OUT/<arm>/<case>.<n>.stderr`.
  Drive runs this same script; the label lets the known-answer gate check
  it, with stand-in scripts for the Arms.
- `verdict TABLE CASES` reads that table and the cases' classes (`slip`,
  `prose-slip`, `guard`, `cite-slip`, `cite-guard`), for the Arms
  `today`, `text-warn`, `text-refuse` and `cite`, and prints a summary
  line per Arm and `verdict: h-text-warn=…`, `verdict: h-text-refuse=…`,
  `verdict: h-cite=…` and `verdict: h-null=…` (the rule is in the
  script's docstring). `known-answers/` holds a supports and a
  contradicts case for each label.

`fixture/` is e-arms' fixture: `cases.tsv`, `fixture.tar` (the store the
cases run in) and `ids.txt` (the ids the cases name).

Version 2: version 1 read the Arm binaries from absolute paths outside the
bundle, built by hand from patches that were never published, so the
known-answer gate could not see them and the repository could not rebuild
them. Version 2 builds every Arm in Launch from the sandbox base and its
patches, names each by sha256, and carries the fixture and cases file in
the bundle. `cases.py`, `verdict.py` and their known answers are unchanged.
