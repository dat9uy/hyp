# id-citation arms

The four Arms of this tree's e-arms, rebuilt as patches against public main.
E-YKJ70D built its arms from a commit and patches that were never published;
these patches reproduce what those arms did, over the rule main now holds.

Base commit: `2d7300e982383e4ce59dc8af0ac4e9b877fd8e46` (main). Every patch
applies to it with `git apply`, and none is stacked on another.

| Arm | Patch | What it does |
|---|---|---|
| `text-refuse` | none | The base itself: main refuses (exit 2, nothing appended) a write whose text flags cite an id that is no node of the store. |
| `today` | `today.patch` | The write commands skip that check, as hyp did before it. |
| `text-warn` | `text-warn.patch` | The check still runs, but a cite that is no node prints a `warn:` line on stderr, and the event is appended with exit 0. |
| `cite` | `cite.patch` | No text is scanned. `request-verify`, `select` and `confirm` take a repeatable `--result REF`, which must resolve to a Result of the store or the command exits 2 with nothing appended; the ids are stored in the payload as `results`. |

Build each Arm from a clean copy of the base, with its patch applied:

```bash
CGO_ENABLED=0 go build -trimpath -o arm-<arm> ./cmd/hyp
```

The build is reproducible: two copies at different paths give the same
sha256. A Result names each binary by that sha256, never by this commit.
`recipes/id-citation-arms@2` builds the Arms in its Launch from these patches,
given as `hyp sandbox create --patch ARM=FILE`.
