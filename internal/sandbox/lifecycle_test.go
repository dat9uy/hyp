package sandbox

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A Launch that refuses to start exits 3 with its own message, runs the
// recipe's Cleanup, whose output the error shows, and leaves no sandbox.
func TestLaunchFailureRunsCleanupAndRemovesSandbox(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "Run: ./launch.sh", "", "Run: ./cleanup.sh"),
		"launch.sh":  "#!/bin/sh\necho 'RUN_PARAMS is not set'\nexit 3\n",
		"cleanup.sh": "#!/bin/sh\necho 'cleanup: no launch state; nothing to check'\n",
	}, "launch.sh", "cleanup.sh")
	_, box, err := createWith(t, "lf", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if code(err) != 3 {
		t.Fatalf("err = %v, want code 3", err)
	}
	for _, s := range []string{"launch: ./launch.sh exited 3", "RUN_PARAMS is not set", "== boot@1 cleanup: ./cleanup.sh",
		"cleanup: no launch state; nothing to check", "Cleanup run; sandbox removed"} {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("err = %v, missing %q", err, s)
		}
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

// An Evidence check that crashes is a failed command: its row has exit 1,
// its traceback is in command.log, and it is returned as CommandFailed,
// whose exit code the CLI passes through.
func TestEvidenceCrashRecordsExit(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not on PATH")
	}
	root := recipeRoot(t, "chk", passCase(map[string]string{
		"recipe.md": recipeMD("chk", "## Evidence\n\nRun: ./check.py\n"),
		"check.py":  "#!/usr/bin/env python3\nraise RuntimeError('injected check crash')\n",
	}, "evidence", "./check.py", nil, 1, ""), "check.py")
	dir := filepath.Join(t.TempDir(), "rec")
	rec, err := Run(RunOptions{Dir: root, Ref: "chk@1", Step: "evidence", Args: []string{keptDir(t)}, Record: dir})
	cf := failed(t, err)
	if cf.Row.Exit == nil || *cf.Row.Exit != 1 {
		t.Fatalf("row = %+v, want exit 1", cf.Row)
	}
	if len(rec.Commands) != 1 || rec.Commands[0].Exit == nil || *rec.Commands[0].Exit != 1 {
		t.Errorf("rows = %+v, want one with exit 1", rec.Commands)
	}
	if log, _ := os.ReadFile(filepath.Join(dir, "command.log")); !strings.Contains(string(log), "RuntimeError: injected check crash") {
		t.Errorf("command.log = %q", log)
	}
}

// scratch-pg stacked under a Drive recipe: create passes, drive passes,
// remove runs Cleanup ok, and no postgres is left on the sandbox's data.
func TestStackedScratchPg(t *testing.T) {
	if _, err := exec.LookPath("initdb"); err != nil {
		t.Skip("initdb not on PATH")
	}
	// A Unix socket path must stay under 104 bytes on macOS, which the
	// default test temp directory is too long for.
	home, err := os.MkdirTemp("/tmp", "h")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(home) })
	t.Setenv("HOME", home)

	lib := t.TempDir()
	pg := filepath.Join(lib, "recipes", "scratch-pg@1")
	if err := os.CopyFS(pg, os.DirFS("testdata/recipes/scratch-pg@1")); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"launch.sh", "doctor.sh", "cleanup.sh"} {
		if err := os.Chmod(filepath.Join(pg, f), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	drv := filepath.Join(lib, "recipes", "drv@1")
	writeFile(t, filepath.Join(drv, "recipe.md"), "---\nname: drv\nversion: 1\nformat: 1\n---\n\n## Drive\n\nRun: ./d.sh\n")
	writeFile(t, filepath.Join(drv, "d.sh"), "#!/bin/sh\npsql -h \"$HYP_SANDBOX_TMP\" -U postgres -d postgres -Atc 'select 1'\n")
	if err := os.Chmod(filepath.Join(drv, "d.sh"), 0o755); err != nil {
		t.Fatal(err)
	}

	rec, _, err := createWith(t, "stk", CreateOptions{Refs: []string{"scratch-pg@1", "drv@1"}, Recipes: lib})
	if err != nil {
		t.Fatal(err)
	}
	db := filepath.Join(rec.TmpDir(), "db")
	t.Cleanup(func() { exec.Command("pg_ctl", "-D", db, "stop", "-m", "immediate").Run() })
	if rec.Recipes[0].Setup != "passed" || rec.Recipes[1].Setup != "none" {
		t.Errorf("setup = %+v", rec.Recipes)
	}
	if _, err := Drive("stk", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	got, err := Remove(RemoveOptions{ID: "stk", OutDir: out})
	if err != nil {
		t.Fatal(err)
	}
	if got.Recipes[0].Cleanup != "ok" {
		t.Errorf("cleanup = %q, want ok", got.Recipes[0].Cleanup)
	}
	if b, _ := exec.Command("pgrep", "-f", db).Output(); len(bytes.TrimSpace(b)) != 0 {
		t.Errorf("postgres still runs on %s: %s", db, b)
	}
}
