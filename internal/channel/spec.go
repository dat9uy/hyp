package channel

import (
	"fmt"
	"strings"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/spec"
)

// specText is what hyp spec prints: every entry hyp acts on, one line each,
// so that a line names the entry it is about. The event kinds come from
// internal/events, everything else from internal/spec.
func specText() string {
	var b strings.Builder
	b.WriteString("hyp spec: the one spec of hyp. hyp acts on each entry below, so a change to a\n")
	b.WriteString("rule is a change to its entry.\n")

	b.WriteString("\nevents: each kind, where it is stored, the Goal it belongs to, whether an end\n")
	b.WriteString("freezes it, and the act that writes it\n")
	for _, k := range events.Kinds {
		frozen := "no"
		if k.Frozen {
			frozen = "yes"
		}
		fmt.Fprintf(&b, "  %s  stored in %s  Goal: %s  frozen by an end: %s  written by: %s\n",
			k.Type, k.Home, k.Goal, frozen, k.Writer)
	}

	b.WriteString("\nstates: each state, when it holds, the acts hyp status may offer in it, and\n")
	b.WriteString("each act it does not allow with the guard that stops that act\n")
	for _, s := range spec.States {
		fmt.Fprintf(&b, "  %s  holds when: %s\n", s.Name, s.Holds)
		fmt.Fprintf(&b, "  %s  allows: %s\n", s.Name, orNone(s.Allows))
		var stops []string
		for _, st := range s.Stops {
			what := st.Act
			if st.When != "" {
				what += ", " + st.When
			}
			stops = append(stops, fmt.Sprintf("%s (guard %s)", what, st.Guard))
		}
		fmt.Fprintf(&b, "  %s  does not allow: %s\n", s.Name, strings.Join(stops, "; "))
	}

	b.WriteString("\nacts: each command, its usage, its roles, the states that allow it, its\n")
	b.WriteString("guards in order, and what it writes\n")
	for _, c := range spec.Commands {
		if c.Bare {
			fmt.Fprintf(&b, "  %s  usage: %s  roles: any, a read with no actor  guards: none  writes: nothing\n", c.Name, c.Usage())
			continue
		}
		if len(c.Roles) == 0 {
			fmt.Fprintf(&b, "  %s  usage: %s  roles: any, a read  guards: none  writes: a read event in the trace\n", c.Name, c.Usage())
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

	b.WriteString("\nguards: each guard, its rule, and the codes of the refusals it gives, each\n")
	fmt.Fprintf(&b, "with exit %d\n", spec.ExitCode(fault.Refusal))
	for _, g := range spec.Guards {
		name := string(g.Name)
		if g.Name == fault.NoGuard {
			name = "(no guard)"
		}
		var codes []string
		for _, c := range g.Codes {
			codes = append(codes, string(c))
		}
		fmt.Fprintf(&b, "  %s  rule: %s  codes: %s  exit %d\n", name, g.Rule, strings.Join(codes, ", "), spec.ExitCode(fault.Refusal))
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

func orNone(xs []string) string {
	if len(xs) == 0 {
		return "none"
	}
	return strings.Join(xs, ", ")
}
