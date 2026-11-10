package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lines writes the Goal's lines as the owner o and returns their keys.
func lines(t *testing.T, g string, texts ...string) []string {
	t.Helper()
	args := []string{"checkpoint", "--summary", "the lines", "--role", "owner", "--actor", "o", "--why", "w"}
	for _, l := range texts {
		args = append(args, "--line", l)
	}
	r := hyp(t, args...)
	if r.code != 0 {
		t.Fatalf("lines: exit %d, stderr %q", r.code, r.stderr)
	}
	var keys []string
	for i := range texts {
		keys = append(keys, g+"#"+string(rune('1'+i)))
		want(t, r, 0, "  line "+keys[i]+": "+texts[i]+"\n")
	}
	return keys
}

// supported adds a Hypothesis under q with a supporting Result.
func supported(t *testing.T, q, title string) string {
	t.Helper()
	h := must(t, "add", "hyp", "--parent", q, "--title", title, "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w")
	return h
}

// A line of the Goal is a limit the owner writes once. It is owed on each
// top-level question with a candidate no closer covers, it holds the
// select there, and a Result that runs its case on the candidate closes it
// for that candidate.
func TestGoalLineIsOwedUntilACloserCoversTheCandidate(t *testing.T) {
	_, g, q := setup(t)
	want(t, hyp(t, "checkpoint", "--summary", "s", "--line", "one tenant per row", "--why", "w"), 2, "--line needs --role owner")
	keys := lines(t, g, "one tenant per row", "a deploy that misses a step fails loudly")
	want(t, hyp(t, "checkpoint", "--summary", "more", "--role", "owner", "--line", "x", "--actor", "o", "--why", "w"), 1,
		"bad lines: the Goal's lines are written once")
	h := supported(t, q, "H")
	want(t, hyp(t, "status"), 0, "runner: "+q+" limit "+keys[0]+" [owed] — one tenant per row\n")
	want(t, hyp(t, "status"), 0, "LINES\n  "+keys[0]+" uncovered on "+q+": "+h+" — one tenant per row\n")
	want(t, hyp(t, "select", q, h, "--why", "w"), 1, "refused: "+q+" owes the runner first: limit "+keys[0])

	e := must(t, "add", "exp", "--parent", h, "--title", "runs both lines", "--recipe", "none", "--decided-by", "reading", "--aims", keys[0], "--target", "none", "--why", "w")
	r := hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--closes", keys[0], "--closes", keys[1], "--why", "w")
	want(t, r, 0, "  closes "+keys[0]+": one tenant per row\n")
	res := strings.Fields(r.stdout)[0]
	notIn(t, "limit "+keys[0])
	want(t, hyp(t, "show", keys[0]), 0, "  covered\nran in "+res+": ran on "+h+": supports\n")
	want(t, hyp(t, "show", g), 0, "line "+keys[1]+" covered on "+q+"; ran in "+res+" — a deploy that misses a step fails loudly")
	notIn(t, keys[0]+" open")
	must(t, "select", q, h, "--why", "w")
	want(t, hyp(t, "show", g), 0, "status: complete")

	// A new candidate that no closer covers owes the line again.
	h2 := supported(t, q, "H2")
	want(t, hyp(t, "status"), 0, "runner: "+q+" limit "+keys[0]+" [aimed_not_settled by ")
	want(t, hyp(t, "show", keys[0]), 0, "  uncovered: "+h2+"\n")
	want(t, hyp(t, "status"), 0, "  "+keys[0]+" uncovered on "+q+": "+h2+"; ran in "+res+" — ")
	want(t, hyp(t, "check"), 0, "check: 0 errors")
}

// The runner refers a line it cannot run, and the owner who wrote the
// Goal's lines answers that refer. An accept answers a refer by an actor
// other than the referrer; the author of the Goal or of a Result does not
// matter.
func TestOwnerAnswersARefusedLine(t *testing.T) {
	_, g, q := setup(t)
	keys := lines(t, g, "survives a region outage")
	h := supported(t, q, "H")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--accept", keys[0], "--actor", "o", "--why", "w"), 1,
		"bad accepts: "+keys[0]+" is not referred; only a referred limit is accepted, so the runner refers it first")
	must(t, "checkpoint", "--summary", "no region in the test bed", "--refers", keys[0], "--why", "w")
	want(t, hyp(t, "status"), 0, "  "+q+" limit referred, accept or return "+keys[0]+" — survives a region outage\n")
	want(t, hyp(t, "select", q, h, "--why", "w"), 1, "refused: referred limit "+keys[0]+" waits for the owner")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--accept", keys[0], "--why", "w"), 1,
		"separation of duties: a referred the limit it accepts")
	must(t, "checkpoint", "--summary", "not this Goal", "--role", "owner", "--accept", keys[0], "--actor", "o", "--why", "w")
	want(t, hyp(t, "show", keys[0]), 0, "hand: accepted by o\n")
	want(t, hyp(t, "status"), 0, "  "+keys[0]+" accepted by o; uncovered on "+q+": "+h+" — ")
	must(t, "select", q, h, "--why", "w")
	must(t, "checkpoint", "--summary", "run it after all", "--role", "owner", "--returns", keys[0], "--actor", "o", "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" limit "+keys[0]+" [owed]")
	want(t, hyp(t, "check"), 0, "check: 0 errors")
}

// writeCommandRecord writes a format-6 command record under the store's
// evidence that says a directory argument held the files with these bytes.
func writeCommandRecord(t *testing.T, dir, exp string, files map[string]string) string {
	t.Helper()
	type entry struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	}
	var entries []entry
	for name, body := range files {
		sum := sha256.Sum256([]byte(body))
		entries = append(entries, entry{name, hex.EncodeToString(sum[:])})
	}
	rec := map[string]any{"format": 6, "dirs": []map[string]any{{"arg": 1, "path": "repo/structures", "sha256": strings.Repeat("0", 64), "files": len(entries), "entries": entries}}}
	data, _ := json.Marshal(rec)
	rel := filepath.Join(".tree", "evidence", exp, "command.json")
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(rel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, rel), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(rel)
}

// A Hypothesis may name its subject, the files that embody its claim. A
// Result covers it when it judges it or when its cited command record says
// those bytes ran; hyp record copies the fingerprints, and no runner types
// them. Once a Result's fingerprints hold the subject, a different subject
// is refused: it is a new Hypothesis.
func TestSubjectCoversALineByFingerprint(t *testing.T) {
	dir, g, q := setup(t)
	keys := lines(t, g, "one tenant per row")
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	if err := os.WriteFile(filepath.Join(dir, "h-row.py"), []byte("row = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := supported(t, q, "H")
	h2 := must(t, "add", "hyp", "--parent", q2, "--title", "H2", "--subject", "h-row.py", "--why", "w")
	want(t, hyp(t, "show", h2), 0, "subject: h-row.py ")
	want(t, hyp(t, "subject", h2, "h-row.py", "--why", "w"), 1, "bad subject: unchanged")
	e2 := must(t, "add", "exp", "--parent", h2, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	must(t, "record", "--exp", e2, "--verdict", h2+"=supports", "--observation", "o", "--why", "w")

	// One run judged H and ran H2's bytes: it covers both.
	e := must(t, "add", "exp", "--parent", h, "--title", "runs the line", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	rec := writeCommandRecord(t, dir, e, map[string]string{"h-row.py": "row = 1\n"})
	res := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", rec, "--closes", keys[0], "--why", "w")
	want(t, hyp(t, "show", keys[0]), 0, "ran in "+res+": ran on "+h+": supports; ran on "+h2+"\n")
	want(t, hyp(t, "status"), 0, "  "+keys[0]+" covered on "+q+", "+q2+"; ran in "+res+" — one tenant per row\n")
	notIn(t, "limit "+keys[0])

	// The bytes ran as H2, so a different subject is a new Hypothesis.
	if err := os.WriteFile(filepath.Join(dir, "h-row.py"), []byte("row = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	want(t, hyp(t, "subject", h2, "h-row.py", "--why", "w"), 1, "bad subject: "+h2+"'s subject ran in "+res+"; a different subject is a new Hypothesis")
	want(t, hyp(t, "check"), 0, "check: 0 errors")
}

// Separation of duties compares the answer with the latest refer: the
// runner that recorded a limit may not have referred it, and the owner may
// have recorded it.
func TestAcceptComparesWithTheReferrer(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "case C", "--why", "w", "--actor", "o")
	key := r + "#1"
	must(t, "checkpoint", "--summary", "no test", "--refers", key, "--why", "w", "--actor", "b")
	must(t, "checkpoint", "--summary", "skip", "--role", "owner", "--accept", key, "--actor", "o", "--why", "w")
	want(t, hyp(t, "show", key), 0, "state: accepted by o\n")
}
