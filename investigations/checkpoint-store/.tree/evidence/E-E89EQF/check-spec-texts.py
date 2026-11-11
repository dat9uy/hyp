#!/usr/bin/env python3
"""Check that replay-spec rules 26 and 27 name each rejection text of
validateSubject and checkRunSubjects.

usage: check-spec-texts.py REPO
Each errors.New / fmt.Errorf text in those two functions is split at its
format verbs; each fixed piece of 4 or more characters must appear in the
text of rules 26 and 27. A text that starts "over limit" needs only
`over limit`. Prints one line per text and exits 1 when one is missing.
"""
import re, sys

repo = sys.argv[1]
src = open(f"{repo}/internal/replay/validate.go").read()
spec = open(f"{repo}/docs/replay-spec.md").read()
rules = "\n".join(l for l in spec.splitlines() if l.startswith("26. ") or l.startswith("27. "))
if rules.count("\n") != 1:
    sys.exit("rules 26 and 27 not found")


def body(name):
    a = src.index(name)
    return src[a:src.index("\n}\n", a)]


texts = []
for fn in ("func (f *Facts) validateSubject(", "func checkRunSubjects("):
    texts += re.findall(r'(?:errors\.New|fmt\.Errorf)\("((?:[^"\\]|\\.)*)"', body(fn))
missing = 0
for t in texts:
    t = t.replace('\\"', '"')
    if t.startswith("over limit"):
        ok = "`over limit`" in rules
    else:
        pieces = [p for p in re.split(r"%[a-z]", t) if len(p.strip()) >= 4]
        ok = all(p in rules for p in pieces)
    print(("named  " if ok else "MISSING"), t)
    missing += not ok
print(f"{len(texts)} texts, {missing} missing")
sys.exit(1 if missing else 0)
