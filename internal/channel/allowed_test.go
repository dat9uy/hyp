package channel

import (
	"strings"
	"testing"
)

type ready struct{ role, line string }

// readyLines parses the allowed acts out of a state block.
func readyLines(out string) []ready {
	var found []ready
	role := ""
	for _, l := range strings.Split(out, "\n") {
		switch {
		case strings.HasSuffix(l, " may:") && !strings.HasPrefix(l, " "):
			role = strings.TrimSuffix(l, " may:")
		case strings.HasPrefix(l, "  hyp ") && role != "":
			found = append(found, ready{role, strings.TrimSpace(l)})
		case !strings.HasPrefix(l, " "):
			if !strings.HasSuffix(l, " may:") {
				role = ""
			}
		}
	}
	return found
}

var actors = map[string]string{"owner": "owner-1", "grill": "grill-1", "probe-writer": "pw-1", "runner": "runner-1"}

// fill replaces the placeholders hyp prints with real keys.
func fill(line, commit string) []string {
	r := strings.NewReplacer("<commit>", commit, "<example>", "b", "<line>", "line-1",
		"<words>", "w", "<kind>:<text>", "gap:t")
	return strings.Fields(r.Replace(line))
}

// every printed line, with its placeholders filled, is a command hyp parses:
// none of them is exit 2.
func runEveryLine(t *testing.T, out, commit string) int {
	t.Helper()
	lines := readyLines(out)
	for _, l := range lines {
		t.Setenv("HYP_ROLE", l.role)
		t.Setenv("HYP_ACTOR", actors[l.role])
		args := fill(l.line, commit)
		if args[0] != "hyp" {
			t.Fatalf("a ready line starts with %q: %s", args[0], l.line)
		}
		if code, o, errs := run(t, args[1:]...); code == 2 {
			t.Errorf("%s as %s: exit 2, stdout %q, stderr %q", l.line, l.role, o, errs)
		}
	}
	return len(lines)
}

func TestAllowedLinesParseAsCommands(t *testing.T) {
	// a Goal that is written: the owner may replace it and place
	f := newFx(t, nil)
	_, out, _ := run(t, "status")
	if n := runEveryLine(t, out, f.base); n < 2 {
		t.Fatalf("ran %d ready lines of a written Goal:\n%s", n, out)
	}

	// a Goal with one candidate is shown selected: the owner may land and
	// unselect its commit
	g := newFx(t, map[string]string{"line-1/a": counterEx})
	c := g.leaf(t, map[string]string{"counter": "ok\n", g.spec(): "spec\n"})
	if code, out, errs := run(t, "run", c, g.goal+"/line-1"); code != 0 {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	_, out, _ = run(t, "status")
	if !strings.Contains(out, "hyp land "+g.goal+" "+c[:12]) || !strings.Contains(out, "hyp unselect "+g.goal+" "+c[:12]) {
		t.Fatalf("no land and unselect lines:\n%s", out)
	}
	if n := runEveryLine(t, out, c); n < 3 {
		t.Fatalf("ran %d ready lines of a selected Goal:\n%s", n, out)
	}
}
