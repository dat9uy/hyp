package channel

import (
	"fmt"
	"strings"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/spec"
	"github.com/dat9uy/hyp/internal/tables"
)

// specText is what hyp spec prints: every entry hyp acts on, one line each,
// so that a line names the entry it is about. The event kinds come from
// internal/events, whether replay checks a guard from the replay entries in
// internal/tables, everything else from internal/spec.
func specText() string {
	var b strings.Builder
	b.WriteString("hyp spec: the one spec of hyp. hyp acts on each entry below, so a change to a\n")
	b.WriteString("rule is a change to its entry.\n")

	b.WriteString("\nevents: each kind, whether it changes a Goal's state, where it is stored, the\n")
	b.WriteString("Goal it belongs to, whether an end freezes it, and the act that writes it; a\n")
	b.WriteString("kind that changes a Goal's state is stored in git, every other kind in the trace\n")
	for _, k := range events.Kinds {
		fmt.Fprintf(&b, "  %s  changes a Goal's state: %s  stored in %s  Goal: %s  frozen by an end: %s  written by: %s  step: %s\n",
			k.Type, yesNo(k.Changes), k.Home(), k.Goal, yesNo(k.Frozen), k.Writer, stepName(k.Step))
	}

	b.WriteString("\nstates: each state, when it holds, the acts hyp status may offer in it, and\n")
	b.WriteString("each act it does not allow with the guard that stops that act\n")
	for _, s := range spec.States {
		fmt.Fprintf(&b, "  %s  holds when: %s\n", s.Name, s.Holds)
		fmt.Fprintf(&b, "  %s  allows: %s\n", s.Name, orNone(s.Allows))
		fmt.Fprintf(&b, "  %s  does not allow: %s\n", s.Name, stopsText(s.Stops))
	}

	b.WriteString("\nfinding states: each state of a finding, when it holds, the acts it allows,\n")
	b.WriteString("and each act it does not allow with the guard that stops that act; a finding's\n")
	b.WriteString("state changes no Goal's state\n")
	for _, s := range spec.FindingStates {
		fmt.Fprintf(&b, "  %s  holds when: %s\n", s.Name, s.Holds)
		fmt.Fprintf(&b, "  %s  allows: %s\n", s.Name, orNone(s.Allows))
		fmt.Fprintf(&b, "  %s  does not allow: %s\n", s.Name, stopsText(s.Stops))
	}

	b.WriteString("\nacts: each command, its usage, its roles, the states that allow it, its\n")
	b.WriteString("guards in order, and what it writes\n")
	for _, c := range spec.Commands {
		if c.Bare {
			fmt.Fprintf(&b, "  %s  usage: %s  roles: any, a read with no actor  guards: none  writes: nothing\n", c.Name, c.Usage())
			continue
		}
		if c.Any {
			fmt.Fprintf(&b, "  %s  usage: %s  roles: any actor, with any role or none  states: every state of the Goal  guards in order: %s  writes: %s\n",
				c.Name, c.Usage(), orNone(guardNames(c.Guards)), c.Writes)
			continue
		}
		if len(c.Roles) == 0 {
			writes := "a read event in the trace"
			if c.Name == "try" {
				writes += "; the cell, in the trace, when this clone's trace lacks it"
			}
			fmt.Fprintf(&b, "  %s  usage: %s  roles: any, a read  guards: none  writes: %s\n", c.Name, c.Usage(), writes)
			continue
		}
		var guards []string
		for _, g := range c.Guards {
			guards = append(guards, string(g))
		}
		fmt.Fprintf(&b, "  %s  usage: %s  roles: %s  states: %s  guards in order: %s  writes: %s\n",
			c.Name, c.Usage(), strings.Join(c.Roles, ", "), orNone(spec.StatesAllowing(c.Name)),
			strings.Join(guards, ", "), c.Writes)
	}

	b.WriteString("\nguards: each guard, its rule, whether replay checks it, and the codes of the\n")
	fmt.Fprintf(&b, "refusals it gives, each with exit %d\n", spec.ExitCode(fault.Refusal))
	for _, g := range spec.Guards {
		name := string(g.Name)
		if g.Name == fault.NoGuard {
			name = "(no guard)"
		}
		var codes []string
		for _, c := range g.Codes {
			codes = append(codes, string(c))
		}
		replay := yesNo(tables.ChecksAtReplay(g.Name)) + g.ReplayWhy
		fmt.Fprintf(&b, "  %s  rule: %s  checked at replay: %s  codes: %s  exit %d\n", name, g.Rule, replay, strings.Join(codes, ", "), spec.ExitCode(fault.Refusal))
	}
	fmt.Fprintf(&b, "  a place that places none of its set gives %s, with the guard of its first cause  exit %d\n",
		fault.SetNotPlaced, spec.ExitCode(fault.Refusal))

	b.WriteString("\nexit codes:\n")
	for _, e := range spec.Exits {
		fmt.Fprintf(&b, "  exit %d  %s: %s\n", e.Code, e.Name, e.Text)
	}

	b.WriteString("\nrules:\n")
	for _, r := range spec.Rules {
		fmt.Fprintf(&b, "  %s\n", r.Name)
		for _, t := range r.Text {
			fmt.Fprintf(&b, "    %s\n", t)
		}
	}
	return b.String()
}

// stopsText is the acts a state does not allow, each with the guard that
// stops it, or none.
func stopsText(stops []spec.Stop) string {
	var out []string
	for _, st := range stops {
		what := st.Act
		if st.When != "" {
			what += ", " + st.When
		}
		out = append(out, fmt.Sprintf("%s (guard %s)", what, st.Guard))
	}
	if len(out) == 0 {
		return "none"
	}
	return strings.Join(out, "; ")
}

func guardNames(gs []fault.Guard) []string {
	var out []string
	for _, g := range gs {
		out = append(out, string(g))
	}
	return out
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func orNone(xs []string) string {
	if len(xs) == 0 {
		return "none"
	}
	return strings.Join(xs, ", ")
}

// stepName is the step an event kind's step line names, or none.
func stepName(step string) string {
	if step == "" {
		return "none"
	}
	return step
}
