#!/usr/bin/env python3
import sys

import gatelib as g

rows, files = [], []
for a in sys.argv[1:]:
    (files if a.startswith("/") else rows).append(a)
for r in rows:
    if r not in g.FINGERPRINTS:
        g.die("verdict: unknown row %r (want one of %s)" % (r, ", ".join(g.FINGERPRINTS)))
if len(set(rows)) != len(rows):
    g.die("verdict: a row given twice")
scores = [g.parse_score(p) for p in files]
for s in scores:
    for k in ("arm", "trial"):
        if k not in s:
            g.die("verdict: a score has no %s" % k)
print("\n".join(g.verdict(rows, scores)))
