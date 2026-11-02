package sandbox

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
)

// readPid reads the pid a script wrote, and kills it when the test ends.
func readPid(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { syscall.Kill(pid, syscall.SIGKILL) })
	return pid
}

// bgLaunch is a Launch that leaves body running in the background and
// writes its pid to $HYP_SANDBOX_TMP/pid.
func bgLaunch(body string) map[string]string {
	return map[string]string{
		"recipe.md": bundle("bg", "Run: ./launch.sh", "", ""),
		"launch.sh": "#!/bin/sh\n" + body + " >/dev/null 2>&1 &\necho $! > \"$HYP_SANDBOX_TMP/pid\"\n",
	}
}

func TestRemoveStopsLaunchGroup(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	rec, _, err := createWith(t, "grp", CreateOptions{Refs: []string{"bg@1"}, Recipes: recipeRoot(t, "bg", bgLaunch("sleep 300"), "launch.sh")})
	if err != nil {
		t.Fatal(err)
	}
	pid := readPid(t, filepath.Join(rec.Tmp, "pid"))
	if syscall.Kill(pid, 0) != nil {
		t.Fatal("the Launch's background process did not survive Launch")
	}
	var warn bytes.Buffer
	if _, err := Remove(RemoveOptions{ID: "grp", Discard: true, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	waitGone(t, pid)
	if !strings.Contains(warn.String(), "sandbox: stopped process group "+strconv.Itoa(rec.Commands[0].PGID)+" (SIGTERM)") {
		t.Errorf("warn = %q", warn.String())
	}
}

func TestRemoveEscalatesToKill(t *testing.T) {
	old := stopWait
	stopWait = 200 * time.Millisecond
	t.Cleanup(func() { stopWait = old })
	path := pathDir(t, "git", "lsof", "sleep", "sh")
	setup(t)
	t.Setenv("PATH", path)
	files := bgLaunch("sh -c \"trap '' TERM; while :; do sleep 1; done\"")
	rec, _, err := createWith(t, "kill", CreateOptions{Refs: []string{"bg@1"}, Recipes: recipeRoot(t, "bg", files, "launch.sh")})
	if err != nil {
		t.Fatal(err)
	}
	pid := readPid(t, filepath.Join(rec.Tmp, "pid"))
	var warn bytes.Buffer
	if _, err := Remove(RemoveOptions{ID: "kill", Discard: true, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	waitGone(t, pid)
	if !strings.Contains(warn.String(), "(SIGKILL)") {
		t.Errorf("warn = %q, want SIGKILL", warn.String())
	}
}

func TestDriveTimeoutKillsGroup(t *testing.T) {
	old := driveTimeout
	driveTimeout = time.Second
	t.Cleanup(func() { driveTimeout = old })
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	rec, _ := driveBox(t, "dto", "Run: ./d.sh", map[string]string{
		"d.sh": "#!/bin/sh\nsleep 300 >/dev/null 2>&1 &\necho $! > \"$HYP_SANDBOX_TMP/pid\"\nsleep 300\n",
	}, "d.sh")
	_, err := Drive("dto", &bytes.Buffer{})
	if code(err) != 3 || !strings.Contains(err.Error(), "timed out after 1s") {
		t.Fatalf("err = %v", err)
	}
	waitGone(t, readPid(t, filepath.Join(rec.Tmp, "pid")))
}

func TestEvidenceGroupEndsWithCommand(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	// The case's input has no sandbox.json, so only the run starts sleep.
	root := recipeRoot(t, "ev", passCase(map[string]string{
		"recipe.md": recipeMD("ev", "## Evidence\n\nRun: ./check.sh\n"),
		"check.sh":  "#!/bin/sh\n[ -e \"$1/sandbox.json\" ] || exit 0\nsleep 300 >/dev/null 2>&1 &\necho $! > \"$1/pid\"\n",
	}, "evidence", "./check.sh", nil, 0, ""), "check.sh")
	kept := keptDir(t)
	if _, err := Run(RunOptions{Dir: root, Ref: "ev@1", Step: "evidence", Args: []string{kept}, Record: filepath.Join(t.TempDir(), "r")}); err != nil {
		t.Fatal(err)
	}
	waitGone(t, readPid(t, filepath.Join(kept, "pid")))
}

func TestInterruptRunsCleanup(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	old := interrupted
	interrupted = func() (context.Context, context.CancelFunc) {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(300*time.Millisecond, cancel)
		return ctx, cancel
	}
	t.Cleanup(func() { interrupted = old })
	root := recipeRoot(t, "int", map[string]string{
		"recipe.md": bundle("int", "Run: ./launch.sh", "", "Run: ./stop.sh"),
		"launch.sh": "#!/bin/sh\nexec sleep 30\n",
		"stop.sh":   "#!/bin/sh\n: > \"$HYP_SANDBOX_TMP/stopped\"\n",
	}, "launch.sh", "stop.sh")
	start := time.Now()
	_, box, err := createWith(t, "int", CreateOptions{Refs: []string{"int@1"}, Recipes: root})
	if code(err) != 3 || !strings.Contains(err.Error(), "./launch.sh interrupted") {
		t.Fatalf("err = %v", err)
	}
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("create took %v after the interrupt", d)
	}
	if !strings.Contains(err.Error(), "Cleanup run") || exists(box) {
		t.Errorf("err = %v; sandbox kept = %v", err, exists(box))
	}
}

func TestRowRecordsPGID(t *testing.T) {
	setup(t)
	_, box := driveBox(t, "pg", "Run: ./d.sh", map[string]string{"d.sh": "#!/bin/sh\n"}, "d.sh")
	if _, err := Drive("pg", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	rows := savedRecord(t, box).Commands
	if len(rows) == 0 {
		t.Fatal("no rows")
	}
	for _, c := range rows {
		if c.PGID <= 0 {
			t.Errorf("row %+v has no pgid", c)
		}
	}
}

// A cancelled context is "interrupted"; one past its deadline is a timeout.
func TestRunStepTellsInterruptFromTimeout(t *testing.T) {
	sh := script(t, "#!/bin/sh\nexec sleep 30\n")
	log, err := os.Create(filepath.Join(t.TempDir(), "log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	save := func(cmdrecord.Run) error { return nil }
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)
	row, _ := runStep(ctx, log, "r@1", "drive", filepath.Dir(sh), stepCmd(ctx, log, t.TempDir(), nil, []string{sh}), time.Minute, save, false)
	if row.Error != "interrupted" {
		t.Errorf("cancelled: error = %q, want interrupted", row.Error)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	row, _ = runStep(ctx, log, "r@1", "drive", filepath.Dir(sh), stepCmd(ctx, log, t.TempDir(), nil, []string{sh}), 100*time.Millisecond, save, false)
	if row.Error != "timed out after 100ms" {
		t.Errorf("deadline: error = %q, want a timeout", row.Error)
	}
}
