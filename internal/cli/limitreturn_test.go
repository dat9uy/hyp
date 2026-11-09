package cli

import "testing"

// The owner returns a referred limit whose case the Goal covers: it leaves
// AWAITING OWNER and is the runner's again, owed and holding the select,
// until a Result closes it. The latest refer or return names whose turn it is.
func TestOwnerReturnsAReferredLimit(t *testing.T) {
	_, g, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "missed deploy", "--why", "w")
	key := r + "#1"

	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--returns", key, "--actor", "owner-dat", "--why", "w"), 1, "bad returns: "+key+" is not referred or accepted")
	must(t, "checkpoint", "--summary", "outside the Goal", "--refers", key, "--why", "w")
	want(t, hyp(t, "status"), 0, q+" limit referred, accept or return "+key)

	want(t, hyp(t, "checkpoint", "--summary", "s", "--returns", key, "--why", "w"), 2, "--returns needs --role owner")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--returns", key, "--why", "w"), 1, "separation of duties")
	must(t, "checkpoint", "--summary", "Line B covers a missed deploy; run it", "--role", "owner", "--returns", key, "--actor", "owner-dat", "--why", "w")
	want(t, hyp(t, "show", g), 0, "    returns: "+key+"\n")
	notIn(t, "limit referred")
	want(t, hyp(t, "status"), 0, "runner: "+q+" limit "+key+" [owed] — missed deploy\n")
	notIn(t, "select "+h)
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--returns", key, "--actor", "owner-dat", "--why", "w"), 1, "is not referred or accepted")

	// A second refer hands it to the owner again.
	must(t, "checkpoint", "--summary", "still no test", "--refers", key, "--why", "w")
	want(t, hyp(t, "status"), 0, q+" limit referred, accept or return "+key)
	must(t, "checkpoint", "--summary", "run it", "--role", "owner", "--returns", key, "--actor", "owner-dat", "--why", "w")

	// A closed limit is not returned.
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--why", "w")
	must(t, "record", "--exp", e2, "--verdict", h+"=contradicts", "--observation", "failed", "--closes", key, "--why", "w")
	must(t, "checkpoint", "--summary", "s", "--refers", key, "--why", "w")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--returns", key, "--actor", "owner-dat", "--why", "w"), 1, "bad returns: "+key+" is closed by ")
	want(t, hyp(t, "check"), 0, "check: 0 errors")
}

// An owner who accepted a limit can reopen it with a return: the latest of
// accept and return stands. A runner cannot refer an accepted limit, since
// only the owner reopens one.
func TestOwnerReturnReopensAnAcceptedLimit(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "torn write", "--why", "w")
	key := r + "#1"

	must(t, "checkpoint", "--summary", "durability is its own Goal", "--refers", key, "--why", "w")
	must(t, "checkpoint", "--summary", "accept", "--role", "owner", "--accept", key, "--actor", "owner-dat", "--why", "w")
	want(t, hyp(t, "show", key), 0, "state: accepted by owner-dat\n")
	notIn(t, "limit "+key)
	want(t, hyp(t, "checkpoint", "--summary", "s", "--refers", key, "--why", "w"), 1, "bad refers: "+key+" is accepted; only the owner reopens it, with returns")

	must(t, "checkpoint", "--summary", "I changed my mind: run it", "--role", "owner", "--returns", key, "--actor", "owner-dat", "--why", "w")
	want(t, hyp(t, "show", key), 0, "state: open, returned: the runner runs its case\n")
	want(t, hyp(t, "status"), 0, "runner: "+q+" limit "+key+" [owed] — torn write\n")

	// An accept answers a refer, so accepting it again needs a fresh refer.
	want(t, hyp(t, "checkpoint", "--summary", "skip after all", "--role", "owner", "--accept", key, "--actor", "owner-dat", "--why", "w"), 1,
		"bad accepts: "+key+" is not referred; only a referred limit is accepted, so the runner refers it first")
	must(t, "checkpoint", "--summary", "still no test", "--refers", key, "--why", "w")
	must(t, "checkpoint", "--summary", "skip after all", "--role", "owner", "--accept", key, "--actor", "owner-dat", "--why", "w")
	want(t, hyp(t, "show", key), 0, "state: accepted by owner-dat\n")
	want(t, hyp(t, "check"), 0, "check: 0 errors")
}

// A closed limit's case ran, so an owner cannot accept it: the accept would
// change nothing.
func TestAcceptRefusesAClosedLimit(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "case C", "--why", "w")
	key := r + "#1"
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--why", "w")
	c := must(t, "record", "--exp", e2, "--verdict", h+"=contradicts", "--observation", "ran", "--closes", key, "--why", "w")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--accept", key, "--actor", "owner-dat", "--why", "w"), 1,
		"bad accepts: "+key+" is closed by "+c+"; its case ran, so there is nothing to accept")
	want(t, hyp(t, "show", key), 0, "state: closed by "+c)
}
