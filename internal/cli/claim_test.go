package cli

import (
	"strings"
	"testing"
)

// claimStore is a store with one Question, Hypotheses ha (alias h-a) and
// hb, and an Experiment ex of ha with no recipe.
func claimStore(t *testing.T) (ha, hb, ex string) {
	t.Helper()
	_, _, q := setup(t)
	ha = must(t, "add", "hyp", "--parent", q, "--title", "a", "--alias", "h-a", "--why", "w")
	hb = must(t, "add", "hyp", "--parent", q, "--title", "b", "--why", "w")
	ex = must(t, "add", "exp", "--parent", ha, "--title", "x", "--recipe", "none", "--decided-by", "reading", "--target", hb, "--why", "w")
	return
}

func recordClaim(t *testing.T, ex, ha, hb, obs string, files map[string]string) result {
	t.Helper()
	args := []string{"record", "--exp", ex, "--verdict", ha + "=supports", "--verdict", hb + "=contradicts",
		"--observation", obs, "--why", "w"}
	for name, body := range files {
		ptr := writeEvidence(t, ex, name, body)
		args = append(args, "--evidence", ptr)
	}
	return hyp(t, args...)
}

func TestRecordWarnsVerdictMismatch(t *testing.T) {
	ha, hb, ex := claimStore(t)
	rec := recordClaim(t, ex, ha, hb, "o", map[string]string{
		"out.txt": "verdict: h-a=contradicts\nverdict: " + hb + "=contradicts\n",
	})
	want(t, rec, 0, "warn: verdict mismatch R-")
	if !strings.Contains(rec.stderr, ha+" (h-a) typed supports, its cited files say contradicts") {
		t.Fatalf("stderr %q", rec.stderr)
	}
	if strings.Contains(rec.stderr, hb+" typed") {
		t.Fatalf("hb agrees, stderr %q", rec.stderr)
	}
	want(t, hyp(t, "check"), 0, "warn: verdict mismatch R-")
}

func TestRecordAgreeingVerdictsWarnNothing(t *testing.T) {
	ha, hb, ex := claimStore(t)
	rec := recordClaim(t, ex, ha, hb, "commit 1a2b3c4d ran", map[string]string{
		"out.txt": "head 1a2b3c4d\nverdict: h-a=supports\nverdict: other=supports\n",
	})
	if rec.code != 0 || strings.Contains(rec.stderr, "warn:") {
		t.Fatalf("record: %+v", rec)
	}
	if c := hyp(t, "check"); strings.Contains(c.stdout, "mismatch") || strings.Contains(c.stdout, "uncited") {
		t.Fatalf("check: %+v", c)
	}
}

func TestRecordWarnsUncitedHex(t *testing.T) {
	ha, hb, ex := claimStore(t)
	rec := recordClaim(t, ex, ha, hb, "built at 9f8e7d6c, sha deadbeef01, word cafebabe and 1234567", map[string]string{
		"out.txt": "built at 9f8e7d6c\n",
	})
	want(t, rec, 0, "warn: uncited R-")
	if !strings.Contains(rec.stderr, "deadbeef01 is in the observation but in no cited file") {
		t.Fatalf("stderr %q", rec.stderr)
	}
	// A token in a cited file, one with no digit and one with no letter
	// are not warned.
	for _, tok := range []string{"9f8e7d6c", "cafebabe", "1234567"} {
		if strings.Contains(rec.stderr, tok) {
			t.Fatalf("%s warned: %q", tok, rec.stderr)
		}
	}
	want(t, hyp(t, "check"), 0, "warn: uncited R-")
}

func TestUncitedHexNeedsALocalFile(t *testing.T) {
	ha, hb, ex := claimStore(t)
	rec := hyp(t, "record", "--exp", ex, "--verdict", ha+"=supports", "--verdict", hb+"=contradicts",
		"--observation", "sha deadbeef01", "--evidence", "https://example.com/run", "--why", "w")
	if rec.code != 0 || strings.Contains(rec.stderr, "uncited") {
		t.Fatalf("record: %+v", rec)
	}
}

// writeEvidence writes .tree/evidence/<exp>/name and returns its pointer.
func writeEvidence(t *testing.T, exp, name, body string) string {
	t.Helper()
	ptr := ".tree/evidence/" + exp + "/" + name
	writeFile(t, ptr, body)
	return ptr
}

func TestRecordWarnsUncitedHexInALimit(t *testing.T) {
	ha, hb, ex := claimStore(t)
	ptr := writeEvidence(t, ex, "out.txt", "ran\n")
	rec := hyp(t, "record", "--exp", ex, "--verdict", ha+"=supports", "--verdict", hb+"=contradicts",
		"--observation", "o", "--limit", "one subject, built from b085464", "--reading", "Goal line A: does 1a2b3c4 mean one org?",
		"--evidence", ptr, "--why", "w")
	r := strings.Fields(rec.stdout)[0]
	want(t, rec, 0, "warn: uncited "+r+": b085464 is in limit "+r+"#1 but in no cited file")
	if strings.Contains(rec.stderr, "1a2b3c4") {
		t.Fatalf("a reading was read: %q", rec.stderr)
	}
}
