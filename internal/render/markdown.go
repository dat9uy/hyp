package render

import (
	"fmt"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// Markdown renders the whole tree as a nested outline for owner review:
// headings for the Goal and DesignQuestions, a question's selected design and
// its limits, and bullets for Hypotheses, Experiments and Results. Nothing is
// truncated.
func Markdown(p replay.Projection) string {
	if p.Goal.GoalID == "" {
		return "# (no Goal)\n"
	}
	r := renderer{p: p}
	var b strings.Builder
	g := p.Nodes[p.Goal.GoalID]
	fmt.Fprintf(&b, "# %s %s — %s\n", r.id(g.ID), g.Title, p.Goal.Status)
	if g.Body != "" {
		fmt.Fprintf(&b, "\n%s\n", g.Body)
	}
	var dq func(id string, level int)
	dq = func(id string, level int) {
		n := p.Nodes[id]
		fmt.Fprintf(&b, "\n%s %s %s — %s\n", strings.Repeat("#", min(level, 6)), r.id(id), n.Title, statusText(r, p.Status[id], nil))
		if n.Body != "" {
			fmt.Fprintf(&b, "\n%s\n", n.Body)
		}
		hyp := SelectedHyp(p, id)
		if hyp != "" {
			fmt.Fprintf(&b, "\nSelected: %s\n", r.id(hyp))
			if body := p.Nodes[hyp].Body; body != "" {
				fmt.Fprintf(&b, "Design: %s\n", body)
			}
		}
		if ls := p.Limits[id]; len(ls) > 0 {
			b.WriteString("\nLimits:\n\n")
			for _, l := range ls {
				fmt.Fprintf(&b, "- %s\n", r.limitLine(l, hyp))
			}
		}
		b.WriteString("\n")
		var nested []string
		for _, h := range n.Children {
			hn := p.Nodes[h]
			fmt.Fprintf(&b, "- %s %s — %s\n", r.id(h), hn.Title, statusText(r, p.Status[h], nil))
			for _, c := range hn.Children {
				cn := p.Nodes[c]
				switch cn.Kind {
				case event.KindExperiment:
					fmt.Fprintf(&b, "  - %s %s — %s (recipe %s)\n", r.id(c), cn.Title, p.Status[c].Status, cn.Recipe)
					for _, res := range cn.Children {
						rs, rr := p.Status[res], resOf(p, res)
						flag := ""
						if rr.Questionable {
							flag = " — questionable"
						}
						if len(rr.Uncontained) > 0 {
							flag += " — uncontained"
						}
						fmt.Fprintf(&b, "    - %s %s %s: %s%s\n", r.id(res), rs.Status, r.verdicts(p.Nodes[res].Verdicts),
							p.Nodes[res].Observation, flag)
					}
				case event.KindDesignQuestion:
					fmt.Fprintf(&b, "  - sub-question %s\n", r.id(c))
					nested = append(nested, c)
				}
			}
		}
		for _, c := range nested {
			dq(c, level+1)
		}
	}
	for _, c := range g.Children {
		dq(c, 2)
	}
	return b.String()
}
