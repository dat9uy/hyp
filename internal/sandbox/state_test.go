package sandbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
)

// TestHelperCreate is not a test of its own: killedCreate runs it in a child
// process, so a real SIGKILL can stop a hyp during Create.
func TestHelperCreate(t *testing.T) {
	data := os.Getenv("HYP_TEST_HELPER_CREATE")
	if data == "" {
		t.Skip("run by killedCreate")
	}
	var o CreateOptions
	if err := json.Unmarshal([]byte(data), &o); err != nil {
		t.Fatal(err)
	}
	if _, err := Create(o); err != nil {
		t.Fatal(err)
	}
}

// killedCreate runs Create in a child process, waits until the row of its
// first Launch has a process group, and kills the child with SIGKILL. The
// Launch keeps running. It returns the sandbox directory and that group.
func killedCreate(t *testing.T, id string, o CreateOptions) (string, int) {
	t.Helper()
	o.ID = id
	o.Host, o.Base = host(t)
	o.Hyp = stubHyp(t)
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperCreate$")
	cmd.Env = append(os.Environ(), "HYP_TEST_HELPER_CREATE="+string(data))
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	box := boxDir(t, id)
	deadline := time.Now().Add(60 * time.Second)
	pgid := 0
	for pgid == 0 {
		if time.Now().After(deadline) {
			cmd.Process.Kill()
			cmd.Wait()
			t.Fatal("no Launch row with a process group")
		}
		time.Sleep(20 * time.Millisecond)
		if rec, err := readRecord(filepath.Join(box, "sandbox.json"), id); err == nil {
			for _, c := range rec.Commands {
				if c.Step == recipe.Launch && c.PGID != 0 {
					pgid = c.PGID
				}
			}
		}
	}
	t.Cleanup(func() { syscall.Kill(-pgid, syscall.SIGKILL) })
	if err := cmd.Process.Signal(syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	cmd.Wait()
	return box, pgid
}

// A hyp killed during Launch leaves a sandbox that cannot be driven and can
// be removed, and remove stops the Launch that outlived it (sandbox spec §4).
func TestKilledLaunchCannotDriveAndCanBeRemoved(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	marker := filepath.Join(t.TempDir(), "driven")
	root := recipeRoot(t, "slow", map[string]string{
		"recipe.md": recipeMD("slow", "## Launch\n\nRun: ./launch.sh\n\n## Drive\n\nRun: ./drive.sh\n"),
		"launch.sh": "#!/bin/sh\nexec sleep 120\n",
		"drive.sh":  "#!/bin/sh\n: > " + marker + "\n",
	}, "launch.sh", "drive.sh")
	box, pgid := killedCreate(t, "kl", CreateOptions{Refs: []string{"slow@1"}, Recipes: root})

	_, err := Drive("kl", io.Discard)
	if code(err) != 3 || !strings.Contains(err.Error(), "hyp sandbox remove") {
		t.Fatalf("drive: err = %v, want code 3 naming remove", err)
	}
	if exists(marker) {
		t.Fatal("drive ran a command on a setup that never finished")
	}
	if !groupAlive(pgid) {
		t.Fatal("the Launch died with hyp; the test proves nothing about stopping it")
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "kl", OutDir: out}); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if groupAlive(pgid) {
		t.Error("the Launch's group is still alive")
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
	kept, err := readRecord(filepath.Join(out, "sandbox.json"), "kl")
	if err != nil {
		t.Fatal(err)
	}
	if kept.Removed == "" || kept.Removing == "" || kept.Recipes[0].Setup != "" {
		t.Errorf("kept record: removed %q, removing %q, setup %q", kept.Removed, kept.Removing, kept.Recipes[0].Setup)
	}
}

// Remove stops an unfinished command before Cleanup, so a Launch that outlived
// its hyp cannot write state a Cleanup has already looked at (sandbox spec §4).
func TestRemoveStopsUnfinishedGroupBeforeCleanup(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	root := recipeRoot(t, "daemon", map[string]string{
		"recipe.md":  recipeMD("daemon", "## Launch\n\nRun: ./launch.sh\n\n## Cleanup\n\nRun: ./cleanup.sh\n"),
		"launch.sh":  "#!/bin/sh\necho $$ > \"$HYP_SANDBOX_RECIPE_TMP/pid\"\nsleep 120\n: > \"$HYP_SANDBOX_RECIPE_TMP/state\"\n",
		"cleanup.sh": "#!/bin/sh\nread pid < \"$HYP_SANDBOX_RECIPE_TMP/pid\"\nif kill -0 \"$pid\" 2>/dev/null; then echo launch-alive; else echo launch-gone; fi\n",
	}, "launch.sh", "cleanup.sh")
	killedCreate(t, "dm", CreateOptions{Refs: []string{"daemon@1"}, Recipes: root})
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "dm", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	log, err := os.ReadFile(filepath.Join(out, "recipe.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(log), "launch-gone") {
		t.Errorf("recipe.log = %q, want Cleanup to find the Launch gone", log)
	}
}

// liveGroup starts a process leading its own group and returns the group.
func liveGroup(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("sleep", "120")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pid := cmd.Process.Pid
	t.Cleanup(func() { syscall.Kill(-pid, syscall.SIGKILL) })
	go cmd.Wait() // reaps it, so a killed group reads as gone
	return pid
}

// addUnfinishedRow writes a Drive row with no ended, as a killed drive leaves.
func addUnfinishedRow(t *testing.T, box string, pgid int) {
	t.Helper()
	rec := savedRecord(t, box)
	rec.Commands = append(rec.Commands, cmdrecord.Run{Ref: "boot@1", Step: recipe.Drive, Args: []string{"./drive.sh"}, PGID: pgid, Started: now()})
	if err := writeJSON(filepath.Join(box, "sandbox.json"), rec); err != nil {
		t.Fatal(err)
	}
}

// Drive refuses while an unfinished row's group lives, and runs once it is
// gone, so "run it again" still works after a killed drive whose command died.
func TestDriveRefusesWhileUnfinishedGroupLives(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	_, box := driveBox(t, "ug", "Run: ./drive.sh", map[string]string{"drive.sh": "#!/bin/sh\n"}, "drive.sh")
	pgid := liveGroup(t)
	addUnfinishedRow(t, box, pgid)
	_, err := Drive("ug", io.Discard)
	if code(err) != 3 || !strings.Contains(err.Error(), fmt.Sprint(pgid)) || !strings.Contains(err.Error(), "hyp sandbox remove") {
		t.Fatalf("err = %v, want code 3 naming group %d and remove", err, pgid)
	}
	syscall.Kill(-pgid, syscall.SIGKILL)
	if left := waitGroups([]int{pgid}, 5*time.Second); len(left) > 0 {
		t.Fatal("group did not end")
	}
	if _, err := Drive("ug", io.Discard); err != nil {
		t.Fatalf("drive after the group ended: %v", err)
	}
}

// A refusal that changes nothing leaves a sandbox as it was: it can still be
// driven. A remove that reached Cleanup and was refused for a leftover process
// cannot be driven, and can be removed again (sandbox spec §4).
func TestRemovingIsWrittenAfterRefusalsThatChangeNothing(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	rec, box := driveBox(t, "rf", "Run: ./drive.sh", map[string]string{"drive.sh": "#!/bin/sh\n"}, "drive.sh")
	taken := t.TempDir()
	if _, err := Remove(RemoveOptions{ID: "rf", OutDir: taken}); code(err) != 2 {
		t.Fatalf("existing OUT_DIR: %v", err)
	}
	if _, err := Remove(RemoveOptions{ID: "rf", Discard: true, Keep: []string{"../outside"}}); code(err) != 2 {
		t.Fatalf("bad --keep: %v", err)
	}
	t.Setenv("PATH", pathDir(t, "git"))
	if _, err := Remove(RemoveOptions{ID: "rf", Discard: true}); code(err) != 3 {
		t.Fatalf("missing lsof: %v", err)
	}
	t.Setenv("PATH", path)
	if got := savedRecord(t, box).Removing; got != "" {
		t.Fatalf("removing = %q after refusals that change nothing", got)
	}
	if _, err := Drive("rf", io.Discard); err != nil {
		t.Fatalf("drive after refusals: %v", err)
	}

	// A process whose working directory is in the sandbox refuses the remove.
	held := exec.Command("sleep", "120")
	held.Dir = rec.WTDir()
	if err := held.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { held.Process.Kill(); held.Wait() })
	if _, err := Remove(RemoveOptions{ID: "rf", Discard: true}); code(err) != 3 || !strings.Contains(err.Error(), "still run") {
		t.Fatalf("leftover: %v", err)
	}
	if savedRecord(t, box).Removing == "" {
		t.Fatal("removing not written by a remove that reached Cleanup")
	}
	if _, err := Drive("rf", io.Discard); code(err) != 3 || !strings.Contains(err.Error(), "hyp sandbox remove") {
		t.Fatalf("drive while removing: %v", err)
	}
	held.Process.Kill()
	held.Wait()
	if _, err := Remove(RemoveOptions{ID: "rf", Discard: true}); err != nil {
		t.Fatalf("second remove: %v", err)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

// Setup is written as each recipe ends, not when the next one's first row
// happens to write the record, so a hyp killed between two recipes leaves the
// first's outcome (sandbox spec §3).
func TestCreateWritesSetupPerRecipe(t *testing.T) {
	setup(t)
	root := recipesRoot(t, map[string]map[string]string{
		"first@1":  {"recipe.md": bundle("first", "Run: ./l.sh", "", ""), "l.sh": "#!/bin/sh\n"},
		"empty@1":  {"recipe.md": bundle("empty", "", "", "Run: ./c.sh"), "c.sh": "#!/bin/sh\n"},
		"second@1": {"recipe.md": bundle("second", "Run: ./l.sh", "", ""), "l.sh": "#!/bin/sh\n"},
	})
	var writes []Record
	rename = func(from, to string) error {
		if filepath.Base(to) == "sandbox.json" {
			var r Record
			if data, err := os.ReadFile(from); err != nil || json.Unmarshal(data, &r) != nil {
				t.Errorf("record write = %s (%v)", data, err)
			}
			writes = append(writes, r)
		}
		return os.Rename(from, to)
	}
	t.Cleanup(func() { rename = os.Rename })
	_, box, err := createWith(t, "pr", CreateOptions{Refs: []string{"first@1", "empty@1", "second@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	wrote := func(first, second string) bool {
		for _, r := range writes {
			if r.Recipes[0].Setup == first && r.Recipes[1].Setup == second && r.Recipes[2].Setup == "" && len(r.Commands) == 1 {
				return true
			}
		}
		return false
	}
	if !wrote("passed", "") {
		t.Error("no record write holds the first recipe's setup before the next recipe ran")
	}
	if !wrote("passed", "none") {
		t.Error("no record write holds the second recipe's setup before the third ran")
	}
	final := savedRecord(t, box)
	if final.Recipes[2].Setup != "passed" {
		t.Errorf("final setup = %q", final.Recipes[2].Setup)
	}
}

// The manifest is written before the last recipe's setup, so a sandbox that
// reads as finished always has its baseline.
func TestManifestComesBeforeLastSetup(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md": recipeMD("boot", "## Launch\n\nRun: ./l.sh\n\n## Drive\n\nRun: ./d.sh\n"),
		"l.sh":      "#!/bin/sh\n",
		"d.sh":      "#!/bin/sh\n",
	}, "l.sh", "d.sh")
	rename = func(from, to string) error {
		if filepath.Base(to) == "sandbox.json" && exists(filepath.Join(filepath.Dir(to), "manifest.json")) {
			return fmt.Errorf("record write refused after the manifest")
		}
		return os.Rename(from, to)
	}
	t.Cleanup(func() { rename = os.Rename })
	_, box, err := createWith(t, "mf", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	rename = os.Rename
	if code(err) != 3 || !strings.Contains(err.Error(), "sandbox kept") {
		t.Fatalf("err = %v, want the record write after the manifest to fail", err)
	}
	if !exists(filepath.Join(box, "manifest.json")) {
		t.Fatal("no manifest")
	}
	if _, err := Drive("mf", io.Discard); code(err) != 3 {
		t.Errorf("drive: %v, want code 3", err)
	}
	if _, err := Remove(RemoveOptions{ID: "mf", OutDir: filepath.Join(t.TempDir(), "out")}); err != nil {
		t.Errorf("remove: %v", err)
	}
}

// A directory with no record clears with --discard, unless something still
// runs in it.
func TestDiscardNoRecordChecksLeftovers(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	box := boxDir(t, "nr")
	if err := os.MkdirAll(box, 0o755); err != nil {
		t.Fatal(err)
	}
	held := exec.Command("sleep", "120")
	held.Dir = box
	if err := held.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { held.Process.Kill(); held.Wait() })
	if _, err := Remove(RemoveOptions{ID: "nr", Discard: true}); code(err) != 3 || !strings.Contains(err.Error(), "still run") {
		t.Fatalf("err = %v, want code 3 for a leftover", err)
	}
	if !exists(box) {
		t.Fatal("sandbox deleted while a process ran in it")
	}
	held.Process.Kill()
	held.Wait()
	if _, err := Remove(RemoveOptions{ID: "nr", Discard: true}); err != nil {
		t.Fatal(err)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

// deriveState reads only the record and the lock (sandbox spec §2).
func TestDeriveState(t *testing.T) {
	entry := func(setup string) Entry {
		return Entry{Ref: event.Ref{Name: "r", Version: "1"}, Name: "r", Setup: setup}
	}
	none := Entry{Ref: event.RefNone}
	for _, c := range []struct {
		name string
		rec  *Record
		held bool
		want boxState
	}{
		{"no record, lock held", nil, true, stateBuilding},
		{"no record, lock free", nil, false, stateInterrupted},
		{"unfinished, lock held", &Record{Recipes: []Entry{entry("passed"), entry("")}}, true, stateSettingUp},
		{"unfinished, lock free", &Record{Recipes: []Entry{entry("passed"), entry("")}}, false, stateInterrupted},
		{"passed and none", &Record{Recipes: []Entry{entry("passed"), entry("none")}}, false, stateReady},
		{"ref none", &Record{Recipes: []Entry{none}}, false, stateReady},
		{"no recipes", &Record{Recipes: []Entry{}}, false, stateReady},
		{"failed before unfinished", &Record{Recipes: []Entry{entry("failed"), entry("")}}, false, stateSetupFailed},
		{"removing before failed", &Record{Removing: "t", Recipes: []Entry{entry("failed")}}, false, stateRemoving},
		{"removing before unfinished", &Record{Removing: "t", Recipes: []Entry{entry("")}}, false, stateRemoving},
		{"removing, lock held", &Record{Removing: "t", Recipes: []Entry{entry("")}}, true, stateRemoving},
		{"ready, lock held", &Record{Recipes: []Entry{entry("passed")}}, true, stateReady},
		{"old record that did not say", &Record{Format: 3, Recipes: []Entry{entry("")}}, false, stateInterrupted},
	} {
		if got := deriveState(c.rec, c.held); got != c.want {
			t.Errorf("%s: state = %s, want %s", c.name, got, c.want)
		}
	}
}

// TestStateTable mirrors the state and command table of sandbox spec §2, one
// row per row there and one cell per command. A cell's want is the exit code,
// with 0 meaning the command runs, and what its message must name. There is
// no parser: a change to the table changes this test by hand.
func TestStateTable(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)

	// A recipe with a Drive, its record changed to one row's state.
	boxWith := func(t *testing.T, id string, change func(*Record)) string {
		t.Helper()
		_, box := driveBox(t, id, "Run: ./drive.sh", map[string]string{"drive.sh": "#!/bin/sh\n"}, "drive.sh")
		rec := savedRecord(t, box)
		change(rec)
		if err := writeJSON(filepath.Join(box, "sandbox.json"), rec); err != nil {
			t.Fatal(err)
		}
		return box
	}
	hold := func(t *testing.T, id, box string) {
		t.Helper()
		unlock, err := lockBox(box, id, "create")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(unlock)
	}
	type cell struct {
		code int
		has  string
	}
	busy := cell{3, "is busy"}
	removeNamed := cell{3, "hyp sandbox remove"}
	taken := cell{2, "exists"}
	noBox := cell{2, "no sandbox"}
	for _, row := range []struct {
		name                         string
		make                         func(t *testing.T, id string)
		create, drive, remove, discd cell
	}{
		{"absent", func(t *testing.T, id string) {},
			cell{0, ""}, noBox, noBox, noBox},
		{"building", func(t *testing.T, id string) {
			box := boxDir(t, id)
			if err := os.MkdirAll(box, 0o755); err != nil {
				t.Fatal(err)
			}
			hold(t, id, box)
		}, taken, busy, busy, busy},
		{"setting_up", func(t *testing.T, id string) {
			box := boxWith(t, id, func(r *Record) { r.Recipes[0].Setup = "" })
			hold(t, id, box)
		}, taken, busy, busy, busy},
		{"interrupted, no record", func(t *testing.T, id string) {
			if err := os.MkdirAll(boxDir(t, id), 0o755); err != nil {
				t.Fatal(err)
			}
		}, taken, cell{3, "--discard"}, cell{2, "--discard"}, cell{0, ""}},
		{"interrupted", func(t *testing.T, id string) {
			boxWith(t, id, func(r *Record) { r.Recipes[0].Setup = "" })
		}, taken, removeNamed, cell{0, ""}, cell{0, ""}},
		{"ready", func(t *testing.T, id string) {
			boxWith(t, id, func(r *Record) {})
		}, taken, cell{0, ""}, cell{0, ""}, cell{0, ""}},
		{"ready, an unfinished row's group lives", func(t *testing.T, id string) {
			boxWith(t, id, func(r *Record) {
				r.Commands = append(r.Commands, cmdrecord.Run{Ref: "boot@1", Step: recipe.Drive, Args: []string{"./drive.sh"}, PGID: liveGroup(t), Started: now()})
			})
		}, taken, cell{3, "process group"}, cell{0, ""}, cell{0, ""}},
		{"ready, no Drive command", func(t *testing.T, id string) {
			if _, _, err := createWith(t, id, CreateOptions{Refs: []string{"none"}}); err != nil {
				t.Fatal(err)
			}
		}, taken, cell{2, "no Drive command"}, cell{0, ""}, cell{0, ""}},
		{"setup_failed", func(t *testing.T, id string) {
			boxWith(t, id, func(r *Record) { r.Recipes[0].Setup = "failed" })
		}, taken, removeNamed, cell{0, ""}, cell{0, ""}},
		{"removing", func(t *testing.T, id string) {
			boxWith(t, id, func(r *Record) { r.Removing = now() })
		}, taken, removeNamed, cell{0, ""}, cell{0, ""}},
	} {
		for _, cmd := range []struct {
			name string
			want cell
			run  func(t *testing.T, id string) error
		}{
			{"create", row.create, func(t *testing.T, id string) error {
				_, _, err := createWith(t, id, CreateOptions{Refs: []string{"none"}})
				return err
			}},
			{"drive", row.drive, func(t *testing.T, id string) error {
				_, err := Drive(id, &bytes.Buffer{})
				return err
			}},
			{"remove", row.remove, func(t *testing.T, id string) error {
				_, err := Remove(RemoveOptions{ID: id, OutDir: filepath.Join(t.TempDir(), "out")})
				return err
			}},
			{"discard", row.discd, func(t *testing.T, id string) error {
				_, err := Remove(RemoveOptions{ID: id, Discard: true})
				return err
			}},
		} {
			t.Run(row.name+"/"+cmd.name, func(t *testing.T) {
				id := "st"
				row.make(t, id)
				t.Cleanup(func() { os.RemoveAll(boxDir(t, id)) })
				err := cmd.run(t, id)
				if err == nil {
					if cmd.want.code != 0 {
						t.Fatalf("ran, want code %d", cmd.want.code)
					}
					return
				}
				if code(err) != cmd.want.code {
					t.Fatalf("err = %v, want code %d", err, cmd.want.code)
				}
				if cmd.want.has != "" && !strings.Contains(err.Error(), cmd.want.has) {
					t.Errorf("err = %v, want it to name %q", err, cmd.want.has)
				}
			})
		}
	}
}
