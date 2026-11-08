package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// citeStore builds a store with a scope-change select, so an owner confirm
// can be written, and returns its ids.
func citeStore(t *testing.T) (dir, q, h, e, r, rULID string) {
	t.Helper()
	dir, _, q = setup(t)
	h = must(t, "add", "hyp", "--parent", q, "--alias", "h-alpha", "--title", "alpha", "--why", "rival")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "beta", "--why", "rival")
	e = must(t, "add", "exp", "--parent", h, "--target", h2, "--recipe", "none", "--decided-by", "reading", "--title", "E", "--why", "test")
	r = must(t, "record", "--exp", e, "--verdict", h+"=supports", "--verdict", h2+"=contradicts", "--observation", "alpha holds", "--why", "ran")
	must(t, "select", q, h, "--scope-change", "--why", "pick")
	var evs struct {
		Events []struct{ Type, Node string }
	}
	if err := json.Unmarshal([]byte(hyp(t, "events", "--json").stdout), &evs); err != nil {
		t.Fatal(err)
	}
	for _, ev := range evs.Events {
		if ev.Type == "result.record" {
			rULID = ev.Node
		}
	}
	return
}

func TestCiteRefusesUnresolvedIDs(t *testing.T) {
	dir, q, h, e, r, _ := citeStore(t)
	typo := r[:len(r)-1] + map[bool]string{true: "Y", false: "Z"}[strings.HasSuffix(r, "Z")]
	cases := []struct {
		name, token string
		args        []string
	}{
		{"placeholder in why (#240's event)", "--why names R-<new>",
			[]string{"confirm", q, "--role", "owner", "--verdict", "insufficient", "--why", "Done the to-do: R-<new> shows declared honest reading Results clear"}},
		{"ellipsis in why", "--why names E-…", []string{"confirm", q, "--role", "owner", "--verdict", "insufficient", "--why", "done: E-… shows it"}},
		{"dots in summary", "--summary names E-...", []string{"checkpoint", "--summary", "see E-... later", "--why", "x"}},
		{"Xs in observation", "--observation names R-XXXXXX",
			[]string{"record", "--exp", e, "--verdict", h + "=supports", "--observation", "R-XXXXXX shows alpha", "--why", "x"}},
		{"seven Xs", "--why names R-XXXXXXX", []string{"confirm", q, "--role", "owner", "--verdict", "insufficient", "--why", "R-XXXXXXX shows alpha"}},
		{"braces in body", "--body names H-{id}", []string{"add", "hyp", "--parent", q, "--title", "g", "--body", "follows from H-{id}", "--why", "x"}},
		{"angle in amend body", "--body names H-<id>", []string{"amend", h, "--body", "as H-<id> says", "--why", "x"}},
		{"placeholder in limit", "--limit names R-<n>",
			[]string{"record", "--exp", e, "--verdict", h + "=supports", "--observation", "ok", "--limit", "fine", "--limit", "see R-<n>", "--why", "x"}},
		{"typo", "--why names " + typo, []string{"confirm", q, "--role", "owner", "--verdict", "insufficient", "--why", typo + " shows alpha holds"}},
		{"another store's id", "--why names R-64ZZ43", []string{"confirm", q, "--role", "owner", "--verdict", "insufficient", "--why", "R-64ZZ43 shows alpha holds"}},
		{"wrong kind", "--why names H-" + r[2:], []string{"confirm", q, "--role", "owner", "--verdict", "insufficient", "--why", "H-" + r[2:] + " holds"}},
		{"link why", "--why names Q-<x>", []string{"link", h, "external_ref", "gh:1", "--why", "per Q-<x>"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := logLines(t, dir)
			res := hyp(t, c.args...)
			want(t, res, 2, c.token+", which is no node here")
			if n := logLines(t, dir); n != before {
				t.Fatalf("refused write appended: %d lines, was %d", n, before)
			}
		})
	}
}

func TestCiteAcceptsResolvedIDsAndProse(t *testing.T) {
	dir, q, h, _, r, rULID := citeStore(t)
	for _, why := range []string{
		r + " shows alpha holds",
		strings.ToLower(r) + " is fine too",
		"see " + r + ".",
		r + "#1 is a limit key",
		rULID + " shows alpha holds",
		"other-tree:R-ABC123 shows the same",
		"containment:R-<new> was the slip",
		"checked E-mail, R-squared and H-index in the notes",
		"h-alpha holds per the result",
		"the new Result shows alpha holds",
		"R-ABCDEFG is a longer word, and XR-ABCDEF is part of one",
		"cites " + h + " and " + q,
	} {
		before := logLines(t, dir)
		res := hyp(t, "link", h, "external_ref", fmt.Sprintf("gh:%d", before), "--why", why)
		if res.code != 0 || logLines(t, dir) != before+1 {
			t.Fatalf("why %q: exit %d, stderr %q", why, res.code, res.stderr)
		}
	}
}

func TestCiteNamesEveryUnresolvedToken(t *testing.T) {
	_, q, _, _, _, _ := citeStore(t)
	res := hyp(t, "confirm", q, "--role", "owner", "--verdict", "insufficient", "--why", "R-<a> and E-<b>")
	want(t, res, 2, "--why names R-<a>, which is no node here; --why names E-<b>, which is no node here; cite an id this store holds")
}
