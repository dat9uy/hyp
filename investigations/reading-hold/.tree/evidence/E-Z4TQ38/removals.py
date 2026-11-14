# Removal checks of E-Z4TQ38: each patch must apply whole on a clean checkout
# of the commit, and then the named test must fail.
import subprocess, sys
root = sys.argv[1]
R = [
 ("R1", "internal/cli/write.go", '''		if kind == event.KindHypothesis {
			if _, err = admitted(w.p.AdmitAddHyp(p.Parent)); err != nil {
				return draft{}, err
			}
		}
''', "", "TestOpenReadingHoldsNextHypothesis"),
 ("R2", "internal/cli/write.go", '''			if _, err := admitted(w.p.AdmitRetire("supersede", to)); err != nil {
				return draft{}, err
			}''', "", "TestOpenReadingHoldsNextHypothesis"),
 ("R3", "internal/cli/write.go", '''		if n := p.Nodes[id]; n != nil && n.Kind == event.KindHypothesis {
			return p.AdmitRetire("abandon", id)
		}''', "", "TestReadingOnInconclusiveHoldsNextHypothesis"),
 ("R4", "internal/cli/write.go", '''		return p.AdmitReaffirm(id)''', "", "TestOpenReadingHoldsStaleAndReaffirm"),
 ("R5", "internal/replay/turn.go", '''		if l.Question == dq && l.Open() {''', '''		if l.Question == dq && l.Open() && len(l.Reach) > 0 {''', "TestOpenReadingWithNoReachHoldsSelect"),
 ("R6", "internal/replay/owed.go", ''' && !p.movedBy(h):''', ''':''', "TestReadingOnInconclusiveOwesNoUndecided"),
 ("R7", "internal/replay/admit.go", '''	if !ok || p.owesPremise(q, "") {''', '''	if !ok {''', "TestPremiseReleasesNextHypothesis"),
]
for name, f, old, new, test in R:
    subprocess.run(["git", "checkout", "--", "."], cwd=root, check=True)
    path = root + "/" + f
    s = open(path).read()
    if s.count(old) != 1:
        print(f"{name}: failed removal: pattern does not apply whole ({s.count(old)} matches)")
        continue
    open(path, "w").write(s.replace(old, new))
    r = subprocess.run(["go", "test", "./internal/cli", "-count=1", "-run", "^" + test + "$"], cwd=root, capture_output=True, text=True)
    out = (r.stdout + r.stderr).strip().splitlines()
    print(f"{name}: {f} patched; {test} {'FAILS' if r.returncode else 'PASSES'} (exit {r.returncode})")
    for l in out[:6]:
        print("    " + l[:300])
subprocess.run(["git", "checkout", "--", "."], cwd=root, check=True)
