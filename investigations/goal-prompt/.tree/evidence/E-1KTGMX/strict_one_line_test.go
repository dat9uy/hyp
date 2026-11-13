package cli

// Strict check of G-V5NP59#2, meaning 2: hyp show prints each CASES draft on
// one physical line of its output. It sets no verdict: it logs PASS or FAIL
// per row and never fails. Known answers on stand-ins first: a one-line body
// passes, a body with a line break fails on a show that prints it raw.

import (
	"strings"
	"testing"
)

func onePhysicalLine(out, prefix, draft string) bool {
	flat := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(draft)
	for _, l := range strings.Split(out, "\n") {
		if l == prefix+draft || l == prefix+flat && !strings.ContainsAny(draft, "\r\n") {
			return true
		}
	}
	return false
}

func TestStrictOneLine(t *testing.T) {
	cases := textCases(t) // read before setup changes the directory
	for _, s := range []struct{ body, want string }{{"one line", "PASS"}, {"two\nlines", "FAIL"}} {
		_, g, _ := setup(t)
		e := must(t, "add", "dq", "--parent", g, "--title", "S", "--body", s.body, "--why", "w")
		got := "FAIL"
		if onePhysicalLine(hyp(t, "show", e).stdout, "body: ", s.body) {
			got = "PASS"
		}
		t.Logf("stand-in %q: %s (known answer %s)", s.body, got, s.want)
	}
	for _, c := range cases {
		dir, g, q, h := textCaseStore(t)
		_, _, show, line := writeCase(t, dir, g, q, h, c)
		prefix := line[:len(line)-len(c.draft)]
		got := "FAIL"
		if onePhysicalLine(hyp(t, "show", show).stdout, prefix, c.draft) {
			got = "PASS"
		}
		t.Logf("%s %s line breaks=%d: %s", c.file, c.field, strings.Count(c.draft, "\n"), got)
	}
}
