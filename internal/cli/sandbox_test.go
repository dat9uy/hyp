package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestSandboxCommand(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	h := t.TempDir()
	gitIn(t, h, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(h, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, h, "add", "a.txt")
	gitIn(t, h, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "init")
	sha := gitIn(t, h, "rev-parse", "HEAD")
	stub := filepath.Join(t.TempDir(), "hyp")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho hyp 0.0.0-test\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	want(t, hyp(t, "sandbox", "create", "t1"), 2, "usage: hyp sandbox create")
	create := []string{"sandbox", "create", "t1", "--host", h, "--base", sha, "--hyp", stub, "--recipe", "none"}
	want(t, hyp(t, append(create[:len(create):len(create)], "--port", "5")...), 2, "-port")
	want(t, hyp(t, create[:len(create)-2]...), 2, "usage: hyp sandbox create")
	want(t, hyp(t, append(create[:len(create)-2:len(create)-2], "--recipe", "x@1")...), 2, "needs --recipes")
	if r := hyp(t, create...); r.code != 0 || !strings.HasPrefix(r.stdout, "sandbox t1:") {
		t.Fatalf("create: %+v", r)
	}
	want(t, hyp(t, create...), 2, "exists")
	want(t, hyp(t, "sandbox", "remove", "t1", "--discard"), 0, "removed t1; discarded")
	if _, err := os.Stat(filepath.Join(home, ".hyp-sandbox", "t1")); !os.IsNotExist(err) {
		t.Fatalf("sandbox still exists: %v", err)
	}

	// --recipe repeats, in stack order.
	want(t, hyp(t, append(create[:len(create):len(create)], "--recipe", "none")...), 2, "recipe none is already in the stack")
	lib := t.TempDir()
	for _, name := range []string{"env", "harness"} {
		dir := filepath.Join(lib, "recipes", name+"@1")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		md := "---\nname: " + name + "\nversion: 1\nformat: 1\n---\n\n## Doctor\n\nRun: ./doctor.sh\n"
		if err := os.WriteFile(filepath.Join(dir, "recipe.md"), []byte(md), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "doctor.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	stack := append(create[:len(create)-2:len(create)-2], "--recipe", "env@1", "--recipe", "harness@1", "--recipes", lib)
	if r := hyp(t, stack...); r.code != 0 || !strings.HasSuffix(strings.TrimSpace(r.stdout), ", recipe env@1 set up, recipe harness@1 set up") {
		t.Fatalf("stacked create: %+v", r)
	}
	want(t, hyp(t, "sandbox", "remove", "t1", "--discard"), 0, "removed t1; discarded")
}

func TestSandboxDriveCommand(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	t.Setenv("HOME", t.TempDir())
	want(t, hyp(t, "sandbox", "drive"), 2, "usage: hyp sandbox drive ID")
	want(t, hyp(t, "sandbox", "drive", "nosuch"), 2, "no sandbox nosuch")
}
