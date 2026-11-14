package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// G-SZCBXZ#5: after hyp sync, both records hold the union of both, every
// event and named file, also when each side holds events the other lacks.
// Two clones share an Experiment E (synced before they split). Clone a's
// Result cites the evidence file .tree/evidence/E/run; clone b's Result
// cites .tree/evidence/E/run/log.txt, a file under a folder of that name.
// Neither names one path with other bytes (the case the owner put out of
// the Goal): one path is a file on one side and a folder on the other. The
// join must hold both events and both files.
func TestReviewSyncJoinsAFileAndAFolderAtOnePath(t *testing.T) {
	gitConfig(t)
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	remote := filepath.Join(top, "remote.git")
	gitIn(t, top, "init", "-q", "--bare", remote)
	a := filepath.Join(top, "a")
	gitIn(t, top, "clone", "-q", remote, a)
	writeFile(t, filepath.Join(a, "s", "README.md"), "a store\n")
	gitIn(t, a, "add", "-A")
	gitIn(t, a, "commit", "-qm", "base")
	gitIn(t, a, "push", "-q", "origin", "main")
	t.Chdir(filepath.Join(a, "s"))
	must(t, "init")
	g := must(t, "add", "goal", "--title", "G", "--why", "w")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--rubric", "reading", "--why", "w")
	want(t, hyp(t, "sync"), 0, "synced")

	b := filepath.Join(top, "b")
	gitIn(t, top, "clone", "-q", remote, b)
	t.Chdir(filepath.Join(b, "s"))
	want(t, hyp(t, "sync", "origin"), 0, "synced")

	t.Chdir(filepath.Join(a, "s"))
	pa := ".tree/evidence/" + e + "/run"
	writeFile(t, pa, "a's run\n")
	ra := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "a ran", "--evidence", pa, "--why", "w")
	want(t, hyp(t, "sync"), 0, "synced")

	t.Chdir(filepath.Join(b, "s"))
	pb := ".tree/evidence/" + e + "/run/log.txt"
	if err := os.MkdirAll(filepath.Dir(pb), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, pb, "b's run\n")
	rb := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "b ran", "--evidence", pb, "--why", "w")
	res := hyp(t, "sync")
	t.Logf("hyp sync in b: exit %d\nstdout: %sstderr: %s", res.code, res.stdout, res.stderr)
	if res.code != 0 {
		t.Fatalf("hyp sync could not join a record whose file %s is a folder in the other record: exit %d", pa, res.code)
	}
	t.Logf("the joined record's evidence of %s:\n%s", e, gitIn(t, b, "ls-tree", "-r", "--name-only", "refs/hyp/record", "s/.tree/evidence/"+e))
	for _, c := range []struct{ id, ptr string }{{ra, pa}, {rb, pb}} {
		want(t, hyp(t, "show", c.id), 0, "")
		want(t, hyp(t, "show", c.ptr), 0, "run")
	}
}
