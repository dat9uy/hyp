package spec

import (
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/testutil"
)

// TestGuardsHoldEveryRefusal: each code a guard can refuse with is in that
// guard's entry, and each code of an entry is one that guard refuses with
// somewhere in hyp's code, so a refusal hyp gives is always the one hyp spec
// prints. The pairs are the values of the guard and the code of every
// fault.Refuse call in the source of the binary. The one refusal whose guard
// is no constant is set-not-placed, which names the guard of its first cause
// and which hyp spec prints on its own line.
func TestGuardsHoldEveryRefusal(t *testing.T) {
	entries := map[fault.Guard]map[fault.Code]bool{}
	for _, g := range Guards {
		entries[g.Name] = map[fault.Code]bool{}
		for _, c := range g.Codes {
			entries[g.Name][c] = true
		}
	}
	used := map[string]bool{}
	refusals := testutil.ReadCode(t).Refusals()
	if len(refusals) < 100 {
		t.Fatalf("found only %d refusals in hyp's code", len(refusals))
	}
	for _, r := range refusals {
		if !r.CodeConst {
			t.Errorf("%s: a refusal's code is no constant of package fault: %s", r.At, r.Code)
			continue
		}
		if !r.GuardConst {
			if fault.Code(r.Code) != fault.SetNotPlaced {
				t.Errorf("%s: %s is refused under a guard that is no constant: %s", r.At, r.Code, r.Guard)
			}
			continue
		}
		codes, ok := entries[fault.Guard(r.Guard)]
		if !ok {
			t.Errorf("%s: hyp refuses under guard %q, which has no entry", r.At, r.Guard)
			continue
		}
		if !codes[fault.Code(r.Code)] {
			t.Errorf("%s: hyp refuses with %q under guard %q, whose entry does not name it", r.At, r.Code, r.Guard)
		}
		used[r.Guard+" "+r.Code] = true
	}
	for _, g := range Guards {
		for _, c := range g.Codes {
			if !used[string(g.Name)+" "+string(c)] {
				t.Errorf("the entry of guard %q names %q, and hyp never refuses with it under that guard", g.Name, c)
			}
		}
	}
}

// TestEveryGuardHasOneEntry: each guard is one entry, and each guard an act or
// a state entry names has one.
func TestEveryGuardHasOneEntry(t *testing.T) {
	seen := map[fault.Guard]int{}
	for _, g := range Guards {
		seen[g.Name]++
		if seen[g.Name] > 1 {
			t.Errorf("guard %q has two entries", g.Name)
		}
	}
	for _, c := range Commands {
		for _, g := range c.Guards {
			if seen[g] == 0 {
				t.Errorf("%s names guard %q, which has no entry", c.Name, g)
			}
		}
	}
	for _, s := range States {
		for _, st := range s.Stops {
			if seen[st.Guard] == 0 {
				t.Errorf("state %s names guard %q, which has no entry", s.Name, st.Guard)
			}
		}
	}
}

// TestStatesNameActsAndGuardsOfTheirActs: a state stops an act only by a
// guard that act checks, and allows and stops only acts that exist.
func TestStatesNameActsAndGuardsOfTheirActs(t *testing.T) {
	acts := map[string]Command{}
	for _, c := range Commands {
		acts[c.Name] = c
	}
	for _, s := range States {
		for _, a := range s.Allows {
			if _, ok := acts[a]; !ok {
				t.Errorf("state %s allows %q, which is no act", s.Name, a)
			}
		}
		for _, st := range s.Stops {
			c, ok := acts[st.Act]
			if !ok {
				t.Errorf("state %s stops %q, which is no act", s.Name, st.Act)
				continue
			}
			if !slices.Contains(c.Guards, st.Guard) {
				t.Errorf("state %s stops %s by %q, which %s does not check", s.Name, st.Act, st.Guard, st.Act)
			}
			if st.When == "" && slices.Contains(s.Allows, st.Act) {
				t.Errorf("state %s both allows and stops every %s", s.Name, st.Act)
			}
		}
	}
}

// TestEveryEventKindHasOneEntry: each event type has one entry, with a home.
func TestEveryEventKindHasOneEntry(t *testing.T) {
	seen := map[string]bool{}
	for _, k := range events.Kinds {
		if seen[k.Type] {
			t.Errorf("event type %s has two entries", k.Type)
		}
		seen[k.Type] = true
		// one rule stores every kind: git when it changes a Goal's state
		if want := map[bool]string{true: events.HomeGit, false: events.HomeTrace}[k.Changes]; k.Home() != want {
			t.Errorf("event type %s is stored in %q, want %q", k.Type, k.Home(), want)
		}
	}
}

// TestExitCodes: each kind of failure has its own exit code.
func TestExitCodes(t *testing.T) {
	for k, want := range map[fault.Kind]int{fault.Refusal: 1, fault.BadArguments: 2, fault.Environment: 3} {
		if got := ExitCode(k); got != want {
			t.Errorf("ExitCode(%v) = %d, want %d", k, got, want)
		}
	}
}

// A usage that hyp help wraps onto a second line is whole in Usage: the
// description column adds nothing to it.
func TestUsageJoinsEveryLine(t *testing.T) {
	want := map[string]string{
		"add goal": "add goal [G] --base COMMIT --words TEXT --line gap:TEXT [--line guard:TEXT ...] [--draft NAME] [--from G] [--finding F ...] [--why TEXT]",
		"try":      "try G LINE/EXAMPLE [--worktree | --commit COMMIT] try --draft NAME --base COMMIT LINE/EXAMPLE",
		"run":      "run COMMIT G/LINE [--why TEXT]",
	}
	for _, c := range Commands {
		if w, ok := want[c.Name]; ok && c.Usage() != w {
			t.Errorf("usage of %s = %q, want %q", c.Name, c.Usage(), w)
		}
	}
}
