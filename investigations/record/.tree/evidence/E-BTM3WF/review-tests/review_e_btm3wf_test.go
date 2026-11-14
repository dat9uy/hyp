package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// G-SZCBXZ#8: hyp import PATH names a work-tree store by its folder. A
// folder whose name holds a colon, which git and the filesystem allow, is
// read as REF:PATH in every spelling (relative, ./ and absolute), so no
// form of the command imports that store.
func TestReviewImportStoreFolderWithAColon(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, "a:b", "README.md"), "r\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, "a:b"))
	must(t, "init")
	dropRecord(t, repo)
	g := must(t, "add", "goal", "--title", "colon goal", "--why", "w")
	t.Chdir(repo)
	for _, arg := range []string{"a:b", "./a:b", filepath.Join(repo, "a:b")} {
		r := hyp(t, "import", arg)
		t.Logf("hyp import %s: exit %d, stdout %q, stderr %q", arg, r.code, r.stdout, r.stderr)
		if r.code == 0 {
			t.Chdir(filepath.Join(repo, "a:b"))
			want(t, hyp(t, "show", g), 0, "colon goal")
			return
		}
	}
	t.Fatalf("no spelling of hyp import PATH imports the store in folder a:b")
}

// G-SZCBXZ#6: hyp import untracks each file under the store's evidence
// folder that no event names, leaving it on disk. A tracked unnamed file
// whose staged bytes differ from both HEAD and the work tree makes
// git rm --cached refuse, so the import exits non-zero after its record
// write and the file stays tracked, on every later run too.
func TestReviewImportUntracksAnUnnamedFileWithStagedChanges(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, "s", "README.md"), "r\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, "s"))
	must(t, "init")
	dropRecord(t, repo)
	chain(t, "cited\n")
	stray := filepath.Join(repo, "s", ".tree", "evidence", "stray.txt")
	writeFile(t, stray, "one\n")
	gitIn(t, repo, "add", "-f", "-A")
	gitIn(t, repo, "commit", "-qm", "stores")
	writeFile(t, stray, "two\n")
	gitIn(t, repo, "add", "-f", "s/.tree/evidence/stray.txt")
	writeFile(t, stray, "three\n")
	t.Chdir(repo)
	for i := 0; i < 2; i++ {
		r := hyp(t, "import", "s")
		t.Logf("run %d: hyp import s: exit %d, stdout %q, stderr %q", i+1, r.code, r.stdout, r.stderr)
	}
	if out := gitIn(t, repo, "ls-files", "s/.tree/evidence/stray.txt"); out != "" {
		t.Fatalf("after hyp import, git still tracks %s, which no event names", out)
	}
	if _, err := os.Stat(stray); err != nil {
		t.Fatalf("hyp import removed the unnamed file from disk: %v", err)
	}
}

// G-SZCBXZ#10: the go test over the source must fail on any non-test Go
// file outside packages record and sandbox that runs git. A dot import of
// os, a raw execve system call and golang.org/x/sys/unix.Exec each start
// a program, and the boundary scan reports none of them.
func TestReviewBoundaryFindsEveryProgramStart(t *testing.T) {
	cases := map[string]string{
		"dot import of os": "package x\n\nimport . \"os\"\n\nfunc f() { StartProcess(\"/bin/sh\", []string{\"sh\", \"-c\", \"g\" + \"it status\"}, &ProcAttr{}) }\n",
		"raw execve":       "package x\n\nimport \"syscall\"\n\nfunc f(p, a, e *byte) { syscall.Syscall(syscall.SYS_EXECVE, 0, 0, 0) }\n",
		"x/sys/unix.Exec":  "package x\n\nimport \"golang.org/x/sys/unix\"\n\nfunc f() { unix.Exec(\"/bin/sh\", nil, nil) }\n",
	}
	var missed []string
	for name, src := range cases {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "go.mod"), "module m\n")
		writeFile(t, filepath.Join(root, "internal", "render", "probe.go"), src)
		if bad := boundaryFindings(t, root); len(bad) == 0 {
			missed = append(missed, name)
		}
	}
	if len(missed) > 0 {
		t.Fatalf("the boundary scan finds nothing for: %s", strings.Join(missed, ", "))
	}
}

var _ = exec.Command
