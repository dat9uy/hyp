# Each removal replaces exact text in a clean checkout of the commit; a pattern
# that does not apply whole is a failed removal. Then it runs the line's tests.
import subprocess, sys, os
root = sys.argv[1]
R = [
 ("actor rule (L2)", [("internal/replay/end.go", '''		if actor != n.Actor {
			return fmt.Errorf("bad end: only the actor that added a Goal ends it: %s added %s, and %s did not", n.Actor, g, actor)
		}
''', "")], "TestEndByAnotherActor"),
 ("already-ended rule (L3)", [("internal/replay/end.go", '''		if e, ok := f.Ended[it.Key]; ok {
			return fmt.Errorf("bad end: %s is already ended, by %s in answer %s; a Goal ends once", g, e.Actor, e.Event)
		}
''', "")], "TestSecondOrMixedEnd"),
 ("mixed rule (L3)", [("internal/replay/end.go", '''				if f.keyGoal(m.Key) == it.Key {
					return fmt.Errorf("bad end: one answer ends %s and moves its key %s; an end is an answer of its own", g, f.keyDisplay(m.Key))
				}
''', "				_ = m\n")], "TestSecondOrMixedEnd"),
 ("write refusal at run and admission (L5)", [
   ("internal/cli/write.go", '''		if g := p.EndedGoals(ev); len(g) > 0 {
			return event.Event{}, refusedErr("%s", p.EndedRefusal(g[0]))
		}
''', ""),
   ("internal/replay/end.go", '''	if e, ok := f.Ended[g]; ok {
		return []Finding{{FindRefuse, endedText(g, e)}}
	}
	return nil
}''', '''	_ = g
	return nil
}''')], "TestWriteUnderAnEndedGoalRefused"),
 ("prompt refusal (L6)", [("internal/cli/prompt.go", '''	if e, ended := r.full.EndOf(id); ended {
		return herr.Refused("hyp prompt %s: the Goal is ended, by %s in answer %s: %q; an ended Goal has no run", g, e.Actor, e.Event, e.Text)
	}
''', "")], "TestPromptForAnEndedGoalRefused"),
 ("ended cite (L7)", [("internal/replay/keyed.go", '''	if _, ok := f.Ended[f.citedGoal(c)]; ok {
		return CiteEnded, ""
	}
''', "")], "TestEndedGoalIsCited"),
 ("Owed skip (L4)", [("internal/replay/owed.go", '''		if q == nil || p.Ended(dq) {''', '''		if q == nil {''')], "TestEndedGoalShowsNothingThatWaits"),
]
for name, edits, tests in R:
    subprocess.run(["git","-C",root,"checkout","-q","--","."], check=True)
    ok = True
    for path, old, new in edits:
        p = os.path.join(root, path); s = open(p).read()
        if s.count(old) != 1:
            ok = False; print(f"## {name}: FAILED REMOVAL, pattern does not apply whole in {path}"); break
        open(p,"w").write(s.replace(old,new))
    if not ok: continue
    r = subprocess.run(["go","test","./internal/cli","-run",tests,"-count=1"], cwd=root, capture_output=True, text=True)
    fails = [l for l in r.stdout.splitlines() if l.startswith("--- FAIL")]
    print(f"## {name}: removed; go test -run {tests}: exit {r.returncode}")
    for l in fails: print("  " + l)
    if r.returncode == 0: print("  NO TEST FAILED: the removal does not bite")
    elif not fails: print("  BUILD FAILED, not a bite:\n" + r.stdout[-400:])
subprocess.run(["git","-C",root,"checkout","-q","--","."], check=True)
