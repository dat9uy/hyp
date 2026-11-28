package gitread

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/testutil"
)

func TestCommitOfAndHasCommit(t *testing.T) {
	r := testutil.NewRepo(t)
	c := r.Commit(t, map[string]string{"a": "a"})
	got, err := CommitOf(r.Dir, "HEAD")
	if err != nil || got != c {
		t.Fatalf("CommitOf(HEAD) = %q, %v; want %s", got, err, c)
	}
	if got, err := CommitOf(r.Dir, c[:8]); err != nil || got != c {
		t.Fatalf("CommitOf(short) = %q, %v; want %s", got, err, c)
	}
	if _, err := CommitOf(r.Dir, "nope"); err == nil {
		t.Fatal("CommitOf(nope) succeeded")
	}
	if _, err := CommitOf(r.Dir, "--output=x"); err == nil {
		t.Fatal("CommitOf took a flag")
	}
	if ok, err := HasCommit(r.Dir, c); err != nil || !ok {
		t.Fatalf("HasCommit(c) = %v, %v; want true", ok, err)
	}
	if ok, err := HasCommit(r.Dir, strings.Repeat("0", 40)); err != nil || ok {
		t.Fatalf("HasCommit(zeros) = %v, %v; want false and no error", ok, err)
	}
}

func TestIsAncestor(t *testing.T) {
	r := testutil.NewRepo(t)
	c1 := r.Commit(t, map[string]string{"a": "1"})
	c2 := r.Commit(t, map[string]string{"a": "2"})
	if ok, err := IsAncestor(r.Dir, c1, c2); err != nil || !ok {
		t.Fatalf("IsAncestor(c1, c2) = %v, %v; want true", ok, err)
	}
	if ok, err := IsAncestor(r.Dir, c2, c1); err != nil || ok {
		t.Fatalf("IsAncestor(c2, c1) = %v, %v; want false", ok, err)
	}
	if ok, err := IsAncestor(r.Dir, c1, c1); err != nil || !ok {
		t.Fatalf("IsAncestor(c1, c1) = %v, %v; want true", ok, err)
	}
	if _, err := IsAncestor(r.Dir, c1, strings.Repeat("0", 40)); err == nil || fault.KindOf(err) != fault.Environment {
		t.Fatalf("IsAncestor with a missing commit: %v, want an environment error", err)
	}
}

func TestExtractTreeKeepsExecBitAndSymlink(t *testing.T) {
	r := testutil.NewRepo(t)
	r.WriteExec(t, "bin/run", "#!/bin/sh\necho hi\n")
	r.Symlink(t, "link", "bin/run")
	c := r.Commit(t, map[string]string{"a/b/c.txt": "deep\x00bytes\r\n", "plain.txt": "p", "empty": ""})
	dest := filepath.Join(t.TempDir(), "out")
	if err := ExtractTree(r.Dir, c, dest); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(dest, "a/b/c.txt")); err != nil || string(b) != "deep\x00bytes\r\n" {
		t.Fatalf("a/b/c.txt = %q, %v", b, err)
	}
	if b, err := os.ReadFile(filepath.Join(dest, "empty")); err != nil || len(b) != 0 {
		t.Fatalf("empty file = %q, %v", b, err)
	}
	if fi, err := os.Stat(filepath.Join(dest, "bin/run")); err != nil || fi.Mode().Perm() != 0o755 {
		t.Fatalf("bin/run mode = %v, %v; want 0755", fi, err)
	}
	if fi, err := os.Stat(filepath.Join(dest, "plain.txt")); err != nil || fi.Mode().Perm() != 0o644 {
		t.Fatalf("plain.txt mode = %v, %v; want 0644", fi, err)
	}
	target, err := os.Readlink(filepath.Join(dest, "link"))
	if err != nil || target != "bin/run" {
		t.Fatalf("link -> %q, %v; want a symlink to bin/run", target, err)
	}
}

func TestExtractTreeFromTreeHash(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"goals/G/line-1/a/example.txtar": "x", "goals/G/line-1/a/probe/run": "y"})
	tree, err := Git(r.Dir, "rev-parse", "HEAD:goals/G/line-1/a")
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "out")
	if err := ExtractTree(r.Dir, strings.TrimSpace(string(tree)), dest); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(dest, "probe/run")); err != nil || string(b) != "y" {
		t.Fatalf("probe/run = %q, %v", b, err)
	}
}

func TestExtractTreeIgnoresExportIgnore(t *testing.T) {
	r := testutil.NewRepo(t)
	c := r.Commit(t, map[string]string{
		".gitattributes": "secret.txt export-ignore\nsubst.txt export-subst\n",
		"secret.txt":     "secret",
		"subst.txt":      "$Format:%H$",
	})
	dest := filepath.Join(t.TempDir(), "out")
	if err := ExtractTree(r.Dir, c, dest); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(dest, "secret.txt")); err != nil || string(b) != "secret" {
		t.Fatalf("secret.txt = %q, %v; an export-ignore file must be extracted", b, err)
	}
	if b, err := os.ReadFile(filepath.Join(dest, "subst.txt")); err != nil || string(b) != "$Format:%H$" {
		t.Fatalf("subst.txt = %q, %v; export-subst must not change bytes", b, err)
	}
}

func TestExtractTreeRefusesSubmodule(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	testutil.Git(t, r.Dir, "update-index", "--add", "--cacheinfo", "160000,"+strings.Repeat("1", 40)+",vendor/sub")
	testutil.Git(t, r.Dir, "commit", "-q", "-m", "sub")
	dest := filepath.Join(t.TempDir(), "out")
	err := ExtractTree(r.Dir, "HEAD", dest)
	if err == nil || fault.KindOf(err) != fault.Refusal || !testutil.Mentions(err, "vendor/sub") {
		t.Fatalf("err = %v, want a refusal naming vendor/sub", err)
	}
}

func TestExtractTreeRefusesPathLeavingDest(t *testing.T) {
	r := testutil.NewRepo(t)
	c := r.Commit(t, map[string]string{"a": "a"})
	// A tree cannot hold a ".." entry through git itself, so the guard is
	// asserted on the path check.
	dest := t.TempDir()
	if _, err := entryPath(dest, "../escape"); err == nil || fault.KindOf(err) != fault.Refusal {
		t.Fatalf("entryPath(../escape) = %v, want a refusal", err)
	}
	if _, err := entryPath(dest, "/abs"); err == nil {
		t.Fatal("entryPath(/abs) was allowed")
	}
	if _, err := entryPath(dest, "a/../../b"); err == nil {
		t.Fatal("entryPath(a/../../b) was allowed")
	}
	if p, err := entryPath(dest, "a/b"); err != nil || p != filepath.Join(dest, "a/b") {
		t.Fatalf("entryPath(a/b) = %q, %v", p, err)
	}
	if err := ExtractTree(r.Dir, c, filepath.Join(dest, "ok")); err != nil {
		t.Fatal(err)
	}
}

func TestExtractTreeWritesNoGitState(t *testing.T) {
	r := testutil.NewRepo(t)
	c := r.Commit(t, map[string]string{"a": "a", "d/b": "b"})
	dest := filepath.Join(t.TempDir(), "out")
	before := testutil.Snapshot(t, r.Dir)
	if err := ExtractTree(r.Dir, c, dest); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(dest, "d/b")); err != nil || string(b) != "b" {
		t.Fatalf("d/b = %q, %v; the extraction must have run", b, err)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("ExtractTree changed the repository")
	}
}

func TestSubtreeHash(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"f/probe/run": "p", "f/example.txtar": "e", "g/other": "o"})
	f, err := Git(r.Dir, "rev-parse", "HEAD:f")
	if err != nil {
		t.Fatal(err)
	}
	want, err := Git(r.Dir, "rev-parse", "HEAD:f/probe")
	if err != nil {
		t.Fatal(err)
	}
	got, err := SubtreeHash(r.Dir, strings.TrimSpace(string(f)), "probe")
	if err != nil || got != strings.TrimSpace(string(want)) {
		t.Fatalf("SubtreeHash = %q, %v; want %s", got, err, want)
	}
	if got, err := SubtreeHash(r.Dir, strings.TrimSpace(string(f)), "nope"); err != nil || got != "" {
		t.Fatalf("a missing entry = %q, %v; want empty and no error", got, err)
	}
	// a name that is a prefix of another entry is not that entry
	if got, err := SubtreeHash(r.Dir, strings.TrimSpace(string(f)), "prob"); err != nil || got != "" {
		t.Fatalf("a prefix = %q, %v; want empty", got, err)
	}
	if _, err := SubtreeHash(r.Dir, "nope", "probe"); err == nil {
		t.Fatal("an unknown tree was read")
	}
}

// caseInsensitive reports whether dir's file system treats a and A as one name.
func caseInsensitive(t *testing.T, dir string) bool {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "probe-case"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := os.Lstat(filepath.Join(dir, "PROBE-CASE"))
	return err == nil
}

// commitRaw commits the given entries (path -> blob hash and mode) without
// touching the work tree.
func commitRaw(t *testing.T, r *testutil.Repo, entries map[string][2]string) string {
	t.Helper()
	for path, e := range entries {
		testutil.Git(t, r.Dir, "update-index", "--add", "--cacheinfo", e[1]+","+e[0]+","+path)
	}
	tree := testutil.Git(t, r.Dir, "write-tree")
	return testutil.Git(t, r.Dir, "commit-tree", "-m", "x", tree)
}

func blobOf(t *testing.T, r *testutil.Repo, body string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "b")
	if err := os.WriteFile(f, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return testutil.Git(t, r.Dir, "hash-object", "-w", f)
}

func TestExtractTreeNeverWritesThroughASymlink(t *testing.T) {
	r := testutil.NewRepo(t)
	outside := t.TempDir()
	dest := filepath.Join(t.TempDir(), "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if !caseInsensitive(t, dest) {
		t.Skip("the case-insensitive spelling of the escape needs a case-insensitive file system")
	}
	c := commitRaw(t, r, map[string][2]string{
		"A":   {blobOf(t, r, outside), "120000"},
		"a/b": {blobOf(t, r, "escaped"), "100644"},
	})
	dest = filepath.Join(t.TempDir(), "out2")
	err := ExtractTree(r.Dir, c, dest)
	if err == nil || fault.KindOf(err) != fault.Environment {
		t.Fatalf("err = %v, want an environment error", err)
	}
	if left, _ := os.ReadDir(outside); len(left) != 0 {
		t.Fatalf("a file landed outside dest: %v", left)
	}
}

func TestExtractTreeRefusesAnExistingParentSymlink(t *testing.T) {
	r := testutil.NewRepo(t)
	outside := t.TempDir()
	c := r.Commit(t, map[string]string{"d/f": "x"})
	dest := filepath.Join(t.TempDir(), "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dest, "d")); err != nil {
		t.Fatal(err)
	}
	err := ExtractTree(r.Dir, c, dest)
	if err == nil || fault.KindOf(err) != fault.Environment || !testutil.Mentions(err, "d") {
		t.Fatalf("err = %v, want an environment error naming d", err)
	}
	if left, _ := os.ReadDir(outside); len(left) != 0 {
		t.Fatalf("a file landed outside dest: %v", left)
	}
}

func TestExtractTreeDoesNotOverwriteCaseCollision(t *testing.T) {
	r := testutil.NewRepo(t)
	dest := filepath.Join(t.TempDir(), "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if !caseInsensitive(t, dest) {
		t.Skip("a case collision needs a case-insensitive file system")
	}
	c := commitRaw(t, r, map[string][2]string{
		"X": {blobOf(t, r, "first"), "100644"},
		"x": {blobOf(t, r, "second"), "100644"},
	})
	dest = filepath.Join(t.TempDir(), "out2")
	err := ExtractTree(r.Dir, c, dest)
	if err == nil || fault.KindOf(err) != fault.Environment {
		t.Fatalf("err = %v, want an environment error", err)
	}
	if b, _ := os.ReadFile(filepath.Join(dest, "X")); string(b) != "first" {
		t.Fatalf("X = %q; the first file was overwritten", b)
	}
}

func goalPaths(t *testing.T, r *testutil.Repo, commit, h0 string) []string {
	t.Helper()
	got, err := GoalChange(r.Dir, commit, h0)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, p := range got {
		out = append(out, p.Path)
	}
	return out
}

// TestGoalChangeReadsTheTrees: the change is each goals/ file whose entry
// differs from H0's, whichever branch brought it; a file the commit lacks is
// marked deleted.
func TestGoalChangeReadsTheTrees(t *testing.T) {
	r := testutil.NewRepo(t)
	h0 := r.Commit(t, map[string]string{"main.go": "1", "goals/G-OLD000/specs/spec.md": "s"})
	testutil.Git(t, r.Dir, "checkout", "-q", "-b", "side")
	r.Commit(t, map[string]string{"goals/G-XXXXXX/line-1/a/example.txtar": "x"})
	testutil.Git(t, r.Dir, "checkout", "-q", "main")
	testutil.Git(t, r.Dir, "checkout", "-q", "-b", "runner")
	r.Commit(t, map[string]string{"own1.go": "1"})
	testutil.Git(t, r.Dir, "merge", "-q", "--no-ff", "-m", "take in main", "side")
	testutil.Git(t, r.Dir, "rm", "-q", "goals/G-OLD000/specs/spec.md")
	tip := r.Commit(t, map[string]string{"goals/G-YYYYYY/specs/spec.md": "s"})

	got, err := GoalChange(r.Dir, tip, h0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"goals/G-OLD000/specs/spec.md": true, "goals/G-XXXXXX/line-1/a/example.txtar": false, "goals/G-YYYYYY/specs/spec.md": false}
	if len(got) != len(want) {
		t.Fatalf("goal change = %v, want the paths of %v", got, want)
	}
	for _, p := range got {
		if del, ok := want[p.Path]; !ok || del != p.Deleted {
			t.Fatalf("goal change = %v, want %v (path: deleted)", got, want)
		}
	}
	if got := goalPaths(t, r, h0, h0); len(got) != 0 {
		t.Fatalf("goal change of H0 = %v", got)
	}
}

// TestWriterFollowsTheBytes: the writer of a path is the commit that wrote
// the bytes it holds, through a merge that took them in, with the ids of the
// event files it adds under .tree/events/.
func TestWriterFollowsTheBytes(t *testing.T) {
	r := testutil.NewRepo(t)
	h0 := r.Commit(t, map[string]string{"main.go": "1"})
	act := r.Commit(t, map[string]string{".tree/events/owner/EV1.json": "{}", "goals/G-XXXXXX/specs/model.tla": "m"})
	plain := r.Commit(t, map[string]string{"goals/G-XXXXXX/specs/x.tla": "x", ".tree/note": "n"})
	testutil.Git(t, r.Dir, "checkout", "-q", "-b", "runner", h0)
	r.Commit(t, map[string]string{"own.go": "1"})
	testutil.Git(t, r.Dir, "merge", "-q", "--no-ff", "-m", "take in main", "main")
	tip := r.Commit(t, map[string]string{"own.go": "2"})

	tree := &CommitTree{Dir: r.Dir, Commit: tip, H0: h0}
	for _, c := range []struct {
		path, writer string
		old          bool
		events       string
	}{
		{"goals/G-XXXXXX/specs/model.tla", act, false, "EV1"},
		{"goals/G-XXXXXX/specs/x.tla", plain, false, ""},
		{"main.go", h0, true, ""}, // H0's history holds it
	} {
		w, err := tree.Writer(c.path)
		if err != nil {
			t.Fatal(err)
		}
		if w.Commit != c.writer || w.Old != c.old || strings.Join(w.Events, ",") != c.events {
			t.Errorf("writer of %s = %+v; want %s, old %v, events %q", c.path, w, c.writer, c.old, c.events)
		}
	}
	h, err := tree.TreeHash("goals/G-XXXXXX/specs")
	if err != nil || h == "" {
		t.Fatalf("tree hash of a folder = %q, %v", h, err)
	}
	if h, err := tree.TreeHash("goals/G-NONE00"); err != nil || h != "" {
		t.Fatalf("tree hash of no folder = %q, %v; want empty", h, err)
	}
}

// TestOnARef: a commit a branch contains is on a ref; once the branch is
// deleted it is on none, though the clone still holds it.
func TestOnARef(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	testutil.Git(t, r.Dir, "checkout", "-q", "-b", "side")
	c := r.Commit(t, map[string]string{"a": "b"})
	if on, err := OnARef(r.Dir, c); err != nil || !on {
		t.Fatalf("on branch side: %v, %v; want true", on, err)
	}
	testutil.Git(t, r.Dir, "checkout", "-q", "--detach", "HEAD~1")
	testutil.Git(t, r.Dir, "branch", "-q", "-D", "side")
	if on, err := OnARef(r.Dir, c); err != nil || on {
		t.Fatalf("after the branch is deleted: %v, %v; want false", on, err)
	}
	if ok, _ := HasCommit(r.Dir, c); !ok {
		t.Fatal("the clone lost the commit")
	}
}
