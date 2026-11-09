package sandbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/recipe"
)

// recipeMD is a recipe.md for name at version 1 with body after it.
func recipeMD(name, body string) string {
	return "---\nname: " + name + "\nversion: 1\nformat: 1\n---\n\n" + body
}

// passCase adds known-answers/pass to files: one case of command that the
// fixture's script passes, so the gate lets the bundle run.
func passCase(files map[string]string, step, command string, args []string, exit int, stdout string) map[string]string {
	files["known-answers/pass/expect.json"] = expect(step, command, args, exit, stdout, "")
	files["known-answers/pass/input/a"] = ""
	return files
}

// driveBox makes sandbox id with one recipe boot whose Drive runs files.
func driveBox(t *testing.T, id, drive string, files map[string]string, exec ...string) (*Record, string) {
	t.Helper()
	files["recipe.md"] = recipeMD("boot", "## Drive\n\n"+drive+"\n")
	root := recipeRoot(t, "boot", files, exec...)
	rec, box, err := createWith(t, id, CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	return rec, box
}

func failed(t *testing.T, err error) *CommandFailed {
	t.Helper()
	var cf *CommandFailed
	if !errors.As(err, &cf) {
		t.Fatalf("err = %v, want *CommandFailed", err)
	}
	return cf
}

func TestDriveRunsInSandbox(t *testing.T) {
	setup(t)
	rec, box := driveBox(t, "drv", "Run: ./drive.sh", map[string]string{
		"drive.sh": "#!/bin/sh\n: > \"$HYP_SANDBOX_WT/driven.txt\"\necho \"$PWD\" > \"$HYP_SANDBOX_TMP/pwd.txt\"\necho \"$PATH\" > \"$HYP_SANDBOX_TMP/path.txt\"\n",
	}, "drive.sh")
	var buf bytes.Buffer
	if _, err := Drive("drv", &buf); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(rec.WTDir(), "driven.txt")) {
		t.Error("Drive did not run in the sandbox")
	}
	pwd, _ := os.ReadFile(filepath.Join(rec.TmpDir(), "pwd.txt"))
	got, _ := filepath.EvalSymlinks(strings.TrimSpace(string(pwd)))
	want, _ := filepath.EvalSymlinks(filepath.Join(box, "recipes", "boot"))
	if got != want {
		t.Errorf("PWD = %q, want %q", got, want)
	}
	if path, _ := os.ReadFile(filepath.Join(rec.TmpDir(), "path.txt")); !strings.HasPrefix(string(path), rec.BinDir()+":") {
		t.Errorf("PATH = %q, want the pinned hyp first", path)
	}
	if !strings.Contains(buf.String(), "boot@1 drive: ./drive.sh ok") {
		t.Errorf("out = %q", buf.String())
	}
	drives := func() int {
		n := 0
		for _, c := range savedRecord(t, box).Commands {
			if c.Step == "drive" {
				n++
			}
		}
		return n
	}
	if n := drives(); n != 1 {
		t.Fatalf("drive rows = %d, want 1", n)
	}
	if _, err := Drive("drv", &buf); err != nil {
		t.Fatal(err)
	}
	if n := drives(); n != 2 {
		t.Errorf("drive rows after a second drive = %d, want 2", n)
	}
}

func TestDriveRefusesDriftedCopy(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "driven")
	_, box := driveBox(t, "drd", "Run: ./drive.sh", map[string]string{"drive.sh": "#!/bin/sh\n"}, "drive.sh")
	writeFile(t, filepath.Join(box, "recipes", "boot", "drive.sh"), "#!/bin/sh\n: > "+marker+"\n")
	_, err := Drive("drd", &bytes.Buffer{})
	if code(err) != 1 || !strings.Contains(err.Error(), "copy hashes ") || !strings.Contains(err.Error(), "nothing driven") {
		t.Fatalf("Drive = %v, want a refusal", err)
	}
	if exists(marker) {
		t.Error("Drive ran from a copy that no longer hashes as the record says")
	}
}

func TestDrivePassesExitCode(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "b-ran")
	_, box := driveBox(t, "drx", "Run: ./a.sh\nRun: ./b.sh", map[string]string{
		"a.sh": "#!/bin/sh\nexit 30\n",
		"b.sh": "#!/bin/sh\n: > " + marker + "\n",
	}, "a.sh", "b.sh")
	_, err := Drive("drx", &bytes.Buffer{})
	cf := failed(t, err)
	if *cf.Row.Exit != 30 || cf.Row.Step != "drive" {
		t.Errorf("row = %+v", cf.Row)
	}
	if exists(marker) {
		t.Error("Drive went on after a failed command")
	}
	n := 0
	for _, c := range savedRecord(t, box).Commands {
		if c.Step == "drive" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("drive rows = %d, want 1", n)
	}
}

func TestDriveRefusals(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	if _, err := Drive("nosuch", &bytes.Buffer{}); code(err) != 2 {
		t.Errorf("unknown id: %v", err)
	}
	create(t, "dnone")
	if _, err := Drive("dnone", &bytes.Buffer{}); code(err) != 2 || !strings.Contains(err.Error(), "has no Drive command") {
		t.Errorf("recipe none: %v", err)
	}
	_, box := driveBox(t, "dfail", "Run: ./a.sh", map[string]string{"a.sh": "#!/bin/sh\n"}, "a.sh")
	rec := savedRecord(t, box)
	rec.Recipes[0].Setup = "failed"
	if err := writeJSON(filepath.Join(box, "sandbox.json"), rec); err != nil {
		t.Fatal(err)
	}
	if _, err := Drive("dfail", &bytes.Buffer{}); code(err) != 3 {
		t.Errorf("setup failed: %v", err)
	}
	prev := driveTimeout
	driveTimeout = time.Second
	t.Cleanup(func() { driveTimeout = prev })
	driveBox(t, "dslow", "Run: ./a.sh", map[string]string{"a.sh": "#!/bin/sh\nexec sleep 5\n"}, "a.sh")
	if _, err := Drive("dslow", &bytes.Buffer{}); code(err) != 3 || !strings.Contains(err.Error(), "timed out after 1s") {
		t.Errorf("timeout: %v", err)
	}
}

func readCommandRecord(t *testing.T, dir string) *cmdrecord.Record {
	t.Helper()
	var rec cmdrecord.Record
	data, err := os.ReadFile(filepath.Join(dir, "command.json"))
	if err != nil || json.Unmarshal(data, &rec) != nil {
		t.Fatalf("command.json = %s (%v)", data, err)
	}
	return &rec
}

func TestRunStandalone(t *testing.T) {
	root := recipeRoot(t, "tool", passCase(map[string]string{
		"recipe.md": recipeMD("tool", "## Standalone\n\nRun echo: ./echo.sh fixed\n"),
		"echo.sh": "#!/bin/sh\n[ \"$2\" = case ] && { echo case; exit 0; }\n" +
			"echo \"args: $*\"\necho \"var=${HYP_SANDBOX_ID:-unset}\"\necho \"pwd=$(pwd)\"\necho \"ls=$(ls -A)\"\n" +
			"echo \"self=$(dirname \"$0\")\"\ntouch \"$(dirname \"$0\")/written\"\n",
	}, "standalone", "echo", []string{"case"}, 0, "case\n"), "echo.sh")
	lib, err := recipe.Load(root, "tool@1")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HYP_SANDBOX_ID", "leak")
	out := filepath.Join(t.TempDir(), "rec")
	if _, err := Run(RunOptions{Dir: root, Ref: "tool@1", Step: "standalone", Label: "echo", Args: []string{"a", "--x"}, Record: out}); err != nil {
		t.Fatal(err)
	}
	logData, _ := os.ReadFile(filepath.Join(out, "command.log"))
	log := string(logData)
	for _, s := range []string{"== tool@1 standalone: ./echo.sh fixed a --x", "args: fixed a --x", "var=unset"} {
		if !strings.Contains(log, s) {
			t.Errorf("command.log = %q, missing %q", log, s)
		}
	}
	lines := map[string]string{}
	for l := range strings.SplitSeq(log, "\n") {
		if k, v, ok := strings.Cut(l, "="); ok {
			lines[k] = v
		}
	}
	// A check runs in an empty directory of its own, never the caller's.
	if pwd := lines["pwd"]; filepath.Base(pwd) != "cwd" || strings.HasPrefix(pwd, root) || exists(pwd) || lines["ls"] != "" {
		t.Errorf("pwd = %q, ls = %q: want an empty cwd outside the library, deleted", pwd, lines["ls"])
	}
	if self := lines["self"]; self == "" || strings.HasPrefix(self, root) || exists(self) {
		t.Errorf("self = %q: want a temporary copy outside the library, deleted", self)
	}
	rec := readCommandRecord(t, out)
	if rec.Ref.String() != "tool@1" || rec.SHA256 != lib.Hash || rec.Path != filepath.Base(root)+"/recipes/tool@1" || rec.Cwd != filepath.Base(root) || rec.Step != "standalone" || len(rec.Commands) != 1 {
		t.Fatalf("command.json = %+v", rec)
	}
	if c := rec.Commands[0]; c.Label != "echo" || !reflect.DeepEqual(c.Args, []string{"./echo.sh", "fixed", "a", "--x"}) || c.Exit == nil || *c.Exit != 0 {
		t.Errorf("row = %+v", c)
	}
	if after, err := recipe.Load(root, "tool@1"); err != nil || after.Hash != lib.Hash {
		t.Errorf("library changed: %v", err)
	}
}

// A Standalone run keeps each file its arguments name: absolute, relative
// with a /, or after KEY=, symlinks followed, as DIR/inputs/<n>/<base>, hashed,
// and gives the script a copy, so a script that writes to its argument leaves
// the original and the kept bytes alone. The row shows the kept path. A bare
// word is passed as given even when a file of that name is in the caller's
// directory; a directory is passed absolute and shown anchored to the
// repository; a missing path is passed as given.
func TestRunKeepsInputs(t *testing.T) {
	root := recipeRoot(t, "tool", passCase(map[string]string{
		"recipe.md": recipeMD("tool", "## Standalone\n\nRun go: ./go.sh\n"),
		"go.sh": "#!/bin/sh\n[ \"$1\" = case ] && { echo case; exit 0; }\necho rewritten > \"$1\"\n" +
			"for a in \"$@\"; do echo \"arg=$a\"; done\n",
	}, "standalone", "go", []string{"case"}, 0, "case\n"), "go.sh")
	writeFile(t, filepath.Join(root, "data", "rel.txt"), "rel\n")
	writeFile(t, filepath.Join(root, "bare"), "bare\n")
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	cases, fixture := filepath.Join(dir, "cases.tsv"), filepath.Join(dir, "fixture.tar")
	writeFile(t, cases, "a\tb\n")
	writeFile(t, fixture, "tar")
	link := filepath.Join(dir, "link")
	if err := os.Symlink(fixture, link); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "rec")
	args := []string{cases, "FIXTURE=" + fixture, link, "data/rel.txt", "bare", "./sub", dir, filepath.Join(dir, "missing"), "--flag"}
	var warn bytes.Buffer
	if _, err := Run(RunOptions{Dir: root, Ref: "tool@1", Step: "standalone", Label: "go", Args: args, Record: out, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(cases); string(data) != "a\tb\n" {
		t.Errorf("the script changed the original %s: %q", cases, data)
	}
	rec := readCommandRecord(t, out)
	name := filepath.Base(root)
	wantRow := []string{"./go.sh", "inputs/1/cases.tsv", "FIXTURE=inputs/2/fixture.tar", "inputs/3/link", "inputs/4/rel.txt", "bare", name + "/sub", dir, filepath.Join(dir, "missing"), "--flag"}
	if rec.Format != cmdrecord.Format || len(rec.Commands) != 1 || !reflect.DeepEqual(rec.Commands[0].Args, wantRow) {
		t.Fatalf("format %d, rows %+v, want args %q", rec.Format, rec.Commands, wantRow)
	}
	for _, c := range []struct {
		n          int
		path, body string
	}{{1, "inputs/1/cases.tsv", "a\tb\n"}, {2, "inputs/2/fixture.tar", "tar"}, {3, "inputs/3/link", "tar"}, {4, "inputs/4/rel.txt", "rel\n"}} {
		in := rec.Inputs[c.n-1]
		kept := filepath.Join(out, filepath.FromSlash(c.path))
		got, _ := os.ReadFile(kept)
		sum, _ := fileSum(kept)
		if in.Arg != c.n || in.Path != c.path || !in.Kept || in.Size != int64(len(c.body)) || in.SHA256 != sum || string(got) != c.body {
			t.Errorf("input %d = %+v, kept %q (sha256 %s)", c.n, in, got, sum)
		}
	}
	if len(rec.Inputs) != 4 {
		t.Errorf("inputs = %+v, want 4", rec.Inputs)
	}
	log, _ := os.ReadFile(filepath.Join(out, "command.log"))
	if !strings.Contains(string(log), "== tool@1 standalone: "+strings.Join(wantRow, " ")+"\n") || !strings.Contains(string(log), "arg=bare\n") ||
		!strings.Contains(string(log), "arg="+filepath.Join(root, "sub")+"\n") {
		t.Errorf("command.log = %s", log)
	}
	if w := warn.String(); w != "warn: argument 7 names a directory outside the repository; a replay cannot reach it\n" {
		t.Errorf("warn = %q", w)
	}
}

// A directory argument, such as a candidate's subject folder, is passed in
// place and recorded in dirs with the hash of its tree before the first
// command ran: a script that writes into it does not change the recorded
// hash, and a changed file gives the next run another hash.
func TestRunHashesDirectoryArgs(t *testing.T) {
	root := recipeRoot(t, "tool", passCase(map[string]string{
		"recipe.md": recipeMD("tool", "## Standalone\n\nRun go: ./go.sh\n"),
		"go.sh":     "#!/bin/sh\n[ \"$1\" = case ] && { echo case; exit 0; }\necho new > \"$1/written.txt\"\n",
	}, "standalone", "go", []string{"case"}, 0, "case\n"), "go.sh")
	subject := filepath.Join(root, "subject")
	writeFile(t, filepath.Join(subject, "h-module.py"), "x = 1\n")
	writeFile(t, filepath.Join(subject, ".hidden", "note"), "n\n")
	run := func() cmdrecord.Dir {
		t.Helper()
		want, files, err := dirSum(subject)
		if err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(t.TempDir(), "rec")
		if _, err := Run(RunOptions{Dir: root, Ref: "tool@1", Step: "standalone", Label: "go", Args: []string{"./subject"}, Record: out}); err != nil {
			t.Fatal(err)
		}
		rec := readCommandRecord(t, out)
		if len(rec.Dirs) != 1 || rec.Dirs[0] != (cmdrecord.Dir{Arg: 1, Path: filepath.Base(root) + "/subject", SHA256: want, Files: files}) {
			t.Fatalf("dirs = %+v, want one with sha256 %s and %d files", rec.Dirs, want, files)
		}
		return rec.Dirs[0]
	}
	first := run()
	if first.Files != 2 {
		t.Errorf("files = %d, want 2 (dot-entries count)", first.Files)
	}
	if err := os.Remove(filepath.Join(subject, "written.txt")); err != nil {
		t.Fatal(err)
	}
	if again := run(); again.SHA256 != first.SHA256 {
		t.Errorf("the same tree hashed %s, then %s", first.SHA256, again.SHA256)
	}
	os.Remove(filepath.Join(subject, "written.txt"))
	writeFile(t, filepath.Join(subject, "h-module.py"), "x = 2\n")
	if changed := run(); changed.SHA256 == first.SHA256 {
		t.Errorf("a changed file kept the hash %s", first.SHA256)
	}
}

// A file over the cap is not kept: its sha256 and size are recorded, and the
// script still gets a copy.
func TestRunBigInputNotKept(t *testing.T) {
	root := recipeRoot(t, "tool", passCase(map[string]string{
		"recipe.md": recipeMD("tool", "## Standalone\n\nRun go: ./go.sh\n"),
		"go.sh":     "#!/bin/sh\n[ \"$1\" = case ] && { echo case; exit 0; }\n[ -f \"$1\" ]\n",
	}, "standalone", "go", []string{"case"}, 0, "case\n"), "go.sh")
	big := filepath.Join(t.TempDir(), "big.bin")
	writeFile(t, big, strings.Repeat("x", 2<<20))
	out := filepath.Join(t.TempDir(), "rec")
	if _, err := Run(RunOptions{Dir: root, Ref: "tool@1", Step: "standalone", Label: "go", Args: []string{big}, Record: out}); err != nil {
		t.Fatal(err)
	}
	want, _ := fileSum(big)
	rec := readCommandRecord(t, out)
	if len(rec.Inputs) != 1 || rec.Inputs[0].Kept || rec.Inputs[0].SHA256 != want || rec.Inputs[0].Size != 2<<20 || exists(filepath.Join(out, "inputs", "1", "big.bin")) {
		t.Errorf("inputs = %+v", rec.Inputs)
	}
}

func writeTemp(t *testing.T, data string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f")
	writeFile(t, p, data)
	return p
}

func evidenceRoot(t *testing.T) string {
	t.Helper()
	return recipeRoot(t, "chk", passCase(map[string]string{
		"recipe.md": recipeMD("chk", "## Evidence\n\nRun: ./check.sh\n"),
		"check.sh":  "#!/bin/sh\necho \"$1\"\nexit 30\n",
	}, "evidence", "./check.sh", nil, 30, "$CASE/input\n"), "check.sh")
}

func keptDir(t *testing.T) string {
	t.Helper()
	kept := filepath.Join(t.TempDir(), "out")
	writeFile(t, filepath.Join(kept, "sandbox.json"), "{}\n")
	return kept
}

func TestRunEvidence(t *testing.T) {
	root, kept := evidenceRoot(t), keptDir(t)
	for _, name := range []string{"e1", "e2"} {
		out := filepath.Join(t.TempDir(), name)
		_, err := Run(RunOptions{Dir: root, Ref: "chk@1", Step: "evidence", Args: []string{kept}, Record: out})
		cf := failed(t, err)
		if *cf.Row.Exit != 30 || cf.Row.Args[len(cf.Row.Args)-1] != kept {
			t.Errorf("%s: row = %+v", name, cf.Row)
		}
		if rows := readCommandRecord(t, out).Commands; len(rows) != 1 || *rows[0].Exit != 30 {
			t.Errorf("%s: command.json rows = %+v", name, rows)
		}
	}
}

func TestRunRefusals(t *testing.T) {
	tool := recipeRoot(t, "tool", passCase(map[string]string{
		"recipe.md": recipeMD("tool", "## Standalone\n\nRun echo: ./echo.sh\n"),
		"echo.sh":   "#!/bin/sh\n",
	}, "standalone", "echo", nil, 0, ""), "echo.sh")
	single := t.TempDir()
	writeFile(t, filepath.Join(single, "recipes", "one.md"), recipeMD("one", "Notes.\n"))
	exist := filepath.Join(t.TempDir(), "rec")
	if err := os.Mkdir(exist, 0o755); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		o    RunOptions
		want string
	}{
		{"record exists", RunOptions{Dir: tool, Ref: "tool@1", Step: "standalone", Label: "echo", Record: exist}, "exists"},
		{"unknown label", RunOptions{Dir: tool, Ref: "tool@1", Step: "standalone", Label: "nope"}, `has no Standalone command "nope"; it has: echo`},
		{"no evidence", RunOptions{Dir: tool, Ref: "tool@1", Step: "evidence", Args: []string{keptDir(t)}}, "has no Evidence command"},
		{"no sandbox.json", RunOptions{Dir: evidenceRoot(t), Ref: "chk@1", Step: "evidence", Args: []string{t.TempDir()}}, "holds no sandbox.json; give the OUT_DIR of hyp sandbox remove, or, for files no sandbox wrote, run the check as a Standalone command (hyp recipe run --record DIR REF NAME PATH)"},
		{"single file", RunOptions{Dir: single, Ref: "one@1", Step: "standalone", Label: "x"}, "recipe one@1 not found"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.o.Record == "" {
				c.o.Record = filepath.Join(t.TempDir(), "rec")
			}
			rec, err := Run(c.o)
			if code(err) != 2 || !strings.Contains(err.Error(), c.want) || rec != nil {
				t.Fatalf("Run = %v, %v; want exit 2 with %q", rec, err, c.want)
			}
			if c.name != "record exists" && exists(c.o.Record) {
				t.Error("record directory made after a refusal")
			}
		})
	}
}

func TestRunTimeout(t *testing.T) {
	prev := standaloneTimeout
	standaloneTimeout = time.Second
	t.Cleanup(func() { standaloneTimeout = prev })
	root := recipeRoot(t, "slow", passCase(map[string]string{
		"recipe.md": recipeMD("slow", "## Standalone\n\nRun wait: ./wait.sh\n"),
		"wait.sh":   "#!/bin/sh\n[ \"$1\" = case ] && exit 0\nexec sleep 5\n",
	}, "standalone", "wait", []string{"case"}, 0, ""), "wait.sh")
	out := filepath.Join(t.TempDir(), "rec")
	_, err := Run(RunOptions{Dir: root, Ref: "slow@1", Step: "standalone", Label: "wait", Record: out})
	if code(err) != 3 {
		t.Fatalf("err = %v, want exit 3", err)
	}
	if rows := readCommandRecord(t, out).Commands; len(rows) != 1 || rows[0].Error != "timed out after 1s" || rows[0].Exit != nil {
		t.Errorf("rows = %+v", rows)
	}
}

// A record path that names an existing directory through .. is refused,
// and a failed record write never deletes a directory hyp did not make.
func TestRunKeepsExistingRecordDir(t *testing.T) {
	root := recipeRoot(t, "tool", passCase(map[string]string{
		"recipe.md": recipeMD("tool", "## Standalone\n\nRun echo: ./echo.sh\n"),
		"echo.sh":   "#!/bin/sh\n",
	}, "standalone", "echo", nil, 0, ""), "echo.sh")
	base := t.TempDir()
	x := filepath.Join(base, "x")
	writeFile(t, filepath.Join(x, "precious"), "keep\n")
	if err := os.Mkdir(filepath.Join(x, "command.log"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Run(RunOptions{Dir: root, Ref: "tool@1", Step: "standalone", Label: "echo", Record: filepath.Join(base, "newdir", "..", "x")})
	if code(err) != 2 || !strings.Contains(err.Error(), "exists") {
		t.Fatalf("err = %v, want exit 2 naming an existing path", err)
	}
	if !exists(filepath.Join(x, "precious")) {
		t.Error("the existing record directory was deleted")
	}
}

// driveRows counts the drive rows of the record in dir.
func driveRows(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	for _, c := range savedRecord(t, dir).Commands {
		if c.Step == "drive" {
			n++
		}
	}
	return n
}

// waitFile waits up to 10s for path to exist.
func waitFile(t *testing.T, path string) {
	t.Helper()
	for deadline := time.Now().Add(10 * time.Second); !exists(path); {
		if time.Now().After(deadline) {
			t.Fatalf("%s did not appear", path)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestDriveRefusesBusySandbox(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	rec, box := driveBox(t, "busy", "Run: ./drive.sh", map[string]string{
		"drive.sh": "#!/bin/sh\n: > \"$HYP_SANDBOX_TMP/started\"\nwhile [ ! -e \"$HYP_SANDBOX_TMP/go\" ]; do sleep 0.1; done\n",
	}, "drive.sh")
	done := make(chan error, 1)
	go func() {
		_, err := Drive("busy", io.Discard)
		done <- err
	}()
	waitFile(t, filepath.Join(rec.TmpDir(), "started"))
	_, err := Drive("busy", io.Discard)
	if code(err) != 3 {
		t.Fatalf("second drive: err = %v, want code 3", err)
	}
	for _, s := range []string{"is busy", fmt.Sprintf("pid %d", os.Getpid()), "hyp sandbox drive"} {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("second drive: err = %q, missing %q", err, s)
		}
	}
	if _, err := Remove(RemoveOptions{ID: "busy", Discard: true}); code(err) != 3 || !strings.Contains(err.Error(), "is busy") {
		t.Errorf("remove during a drive: err = %v, want code 3 busy", err)
	}
	if !exists(filepath.Join(box, "sandbox.json")) {
		t.Fatal("remove during a drive deleted the sandbox")
	}
	writeFile(t, filepath.Join(rec.TmpDir(), "go"), "")
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if n := driveRows(t, box); n != 1 {
		t.Fatalf("drive rows = %d, want 1", n)
	}
	if _, err := Drive("busy", io.Discard); err != nil {
		t.Fatal(err)
	}
	if n := driveRows(t, box); n != 2 {
		t.Fatalf("drive rows = %d, want 2", n)
	}
}

// betweenOpenAndLock runs act once, after lockBox opens sandbox.lock and
// before it locks it: where a remove's delete can land.
func betweenOpenAndLock(t *testing.T, act func()) {
	t.Helper()
	once := false
	flock = func(fd, how int) error {
		if !once {
			once = true
			act()
		}
		return syscall.Flock(fd, how)
	}
	t.Cleanup(func() { flock = syscall.Flock })
}

// A remove deleted the sandbox after drive opened its lock file: drive
// locks a deleted file, and then finds no sandbox.
func TestDriveAfterRemoveIsNoSandbox(t *testing.T) {
	path := pathDir(t, "git", "lsof", "touch")
	setup(t)
	t.Setenv("PATH", path)
	rec, box := driveBox(t, "gone", "Run: ./drive.sh", map[string]string{
		"drive.sh": "#!/bin/sh\ntouch \"$HYP_SANDBOX_TMP/ran\"\n",
	}, "drive.sh")
	betweenOpenAndLock(t, func() { os.RemoveAll(box) })
	if _, err := Drive("gone", io.Discard); code(err) != 2 || !strings.Contains(err.Error(), "no sandbox") {
		t.Fatalf("drive: err = %v, want code 2 no sandbox", err)
	}
	if exists(filepath.Join(rec.TmpDir(), "ran")) {
		t.Error("the Drive script ran")
	}
}

// A remove deleted the sandbox and a create made the same id again, holding
// its lock, after this remove opened the old lock file: the remove is
// refused as busy and leaves the new sandbox alone.
func TestRemoveLeavesRecreatedSandbox(t *testing.T) {
	setup(t)
	_, box := create(t, "again")
	var unlock func()
	betweenOpenAndLock(t, func() {
		os.RemoveAll(box)
		create(t, "again")
		var err error
		if unlock, err = lockBox(box, "again", "create"); err != nil {
			t.Fatal(err)
		}
	})
	if _, err := Remove(RemoveOptions{ID: "again", Discard: true}); code(err) != 3 || !strings.Contains(err.Error(), "hyp sandbox create") {
		t.Fatalf("remove: err = %v, want code 3 busy with the create", err)
	}
	if !exists(filepath.Join(box, "sandbox.json")) {
		t.Error("remove deleted the new sandbox")
	}
	unlock()
}

// pendingRow checks that the record copied to path, read as v, ends in a row
// for step with started and no ended, exit or error, and no "ended" key.
func pendingRow(t *testing.T, path string, step recipe.Step, rows func() []cmdrecord.Run) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	c := rows()
	if len(c) == 0 {
		t.Fatalf("%s holds no row", path)
	}
	last := c[len(c)-1]
	if last.Step != step || last.Started == "" || last.Ended != "" || last.Exit != nil || last.Error != "" {
		t.Errorf("row during the command = %+v", last)
	}
	var raw struct {
		Commands []map[string]json.RawMessage `json:"commands"`
	}
	if json.Unmarshal(data, &raw) != nil || len(raw.Commands) == 0 {
		t.Fatalf("%s = %s", path, data)
	}
	if _, ok := raw.Commands[len(raw.Commands)-1]["ended"]; ok {
		t.Errorf("row during the command has an ended key: %s", data)
	}
}

func TestDriveRowWrittenBeforeCommand(t *testing.T) {
	path := pathDir(t, "git", "lsof", "cp", "dirname")
	setup(t)
	t.Setenv("PATH", path)
	rec, box := driveBox(t, "pend", "Run: ./drive.sh", map[string]string{
		"drive.sh": "#!/bin/sh\ncp \"$(dirname \"$HYP_SANDBOX_WT\")/sandbox.json\" \"$HYP_SANDBOX_TMP/during.json\"\n",
	}, "drive.sh")
	if _, err := Drive("pend", io.Discard); err != nil {
		t.Fatal(err)
	}
	during := filepath.Join(rec.TmpDir(), "during.json")
	var mid Record
	pendingRow(t, during, "drive", func() []cmdrecord.Run {
		data, _ := os.ReadFile(during)
		if err := json.Unmarshal(data, &mid); err != nil {
			t.Fatal(err)
		}
		return mid.Commands
	})
	after := savedRecord(t, box).Commands
	if len(after) != len(mid.Commands) {
		t.Fatalf("rows after = %d, during = %d: the row was appended, not replaced", len(after), len(mid.Commands))
	}
	if last := after[len(after)-1]; last.Ended == "" || last.Exit == nil || *last.Exit != 0 {
		t.Errorf("finished row = %+v", last)
	}
}

func TestRunRowWrittenBeforeCommand(t *testing.T) {
	root := recipeRoot(t, "tool", passCase(map[string]string{
		"recipe.md":                             recipeMD("tool", "## Standalone\n\nRun copy: ./copy.sh\n"),
		"copy.sh":                               "#!/bin/sh\ncp \"$1/command.json\" \"$1/during.json\"\n",
		"known-answers/pass/input/command.json": fmt.Sprintf("{\"format\": %d}\n", cmdrecord.Format),
	}, "standalone", "copy", []string{"$CASE/input"}, 0, ""), "copy.sh")
	out := filepath.Join(t.TempDir(), "rec")
	if _, err := Run(RunOptions{Dir: root, Ref: "tool@1", Step: "standalone", Label: "copy", Args: []string{out}, Record: out}); err != nil {
		t.Fatal(err)
	}
	during := filepath.Join(out, "during.json")
	pendingRow(t, during, "standalone", func() []cmdrecord.Run {
		var mid cmdrecord.Record
		data, _ := os.ReadFile(during)
		if err := json.Unmarshal(data, &mid); err != nil {
			t.Fatal(err)
		}
		if len(mid.Commands) != 1 || mid.Commands[0].Label != "copy" {
			t.Errorf("rows during = %+v, want the one labelled row", mid.Commands)
		}
		return mid.Commands
	})
	if c := readCommandRecord(t, out).Commands; len(c) != 1 || c[0].Ended == "" || c[0].Exit == nil || *c[0].Exit != 0 {
		t.Errorf("finished rows = %+v", c)
	}
}

func TestDriveRefusesWhenRowCannotBeSaved(t *testing.T) {
	path := pathDir(t, "git", "lsof", "touch")
	setup(t)
	t.Setenv("PATH", path)
	rec, box := driveBox(t, "nosave", "Run: ./drive.sh", map[string]string{
		"drive.sh": "#!/bin/sh\ntouch \"$HYP_SANDBOX_TMP/ran\"\n",
	}, "drive.sh")
	failRename(t)
	got, err := Drive("nosave", io.Discard)
	if code(err) != 3 || !strings.Contains(err.Error(), "the command did not run") {
		t.Fatalf("err = %v, want code 3 the command did not run", err)
	}
	if exists(filepath.Join(rec.TmpDir(), "ran")) {
		t.Error("the Drive script ran")
	}
	if got != nil && len(got.Commands) != 0 {
		t.Errorf("returned rows = %+v, want none", got.Commands)
	}
	rename = os.Rename
	if n := driveRows(t, box); n != 0 {
		t.Errorf("drive rows on disk = %d, want 0", n)
	}
}

// listing is every path under dir with its mode, one per line.
func listing(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s %v\n", p, fi.Mode())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestRunRefusesRecordInsideBundle(t *testing.T) {
	root := recipeRoot(t, "tool@1", passCase(map[string]string{
		"recipe.md": recipeMD("tool", "## Standalone\n\nRun touch: ./touch.sh\n"),
		"touch.sh":  "#!/bin/sh\ntouch \"$1\"\n",
	}, "standalone", "touch", []string{"$CASE/x"}, 0, ""), "touch.sh")
	chk := evidenceRoot(t)
	kept := keptDir(t)
	for _, s := range []struct {
		root, ref string
		step      recipe.Step
		label     string
	}{
		{root, "tool@1", "standalone", "touch"},
		{chk, "chk@1", "evidence", ""},
	} {
		rcp, err := recipe.Load(s.root, s.ref)
		if err != nil {
			t.Fatal(err)
		}
		bundle := rcp.Dir
		link := filepath.Join(t.TempDir(), "link")
		if err := os.Symlink(bundle, link); err != nil {
			t.Fatal(err)
		}
		rel := filepath.Join("recipes", filepath.Base(bundle), "rec")
		records := []string{
			filepath.Join(bundle, "rec"),
			filepath.Join(bundle, "a", "b", "rec"),
			filepath.Join(link, "rec"),
			filepath.Join(link, "a", "b", "rec"),
			rel,
		}
		if p := alias(bundle); p != "" {
			records = append(records, filepath.Join(p, "rec"))
		}
		hashes, before := recipe.Hashes(s.root), listing(t, bundle)
		// Every check runs before the temporary copy is made.
		tmp := t.TempDir()
		t.Setenv("TMPDIR", tmp)
		for _, record := range records {
			mark := filepath.Join(t.TempDir(), "mark")
			args := []string{mark}
			if s.step == "evidence" {
				args = []string{kept}
			}
			rec, err := Run(RunOptions{Dir: s.root, Ref: s.ref, Step: s.step, Label: s.label, Args: args, Record: record})
			lib := filepath.Join(s.root, "recipes")
			if rec != nil || code(err) != 2 || !strings.Contains(err.Error(), "recipe library "+lib) {
				t.Errorf("%s --record %s: rec = %v, err = %v; want exit 2 naming %s", s.step, record, rec, err, lib)
			}
			abs := record
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(s.root, abs)
			}
			if exists(abs) || exists(mark) {
				t.Errorf("%s --record %s: record or marker written", s.step, record)
			}
		}
		if !reflect.DeepEqual(recipe.Hashes(s.root), hashes) || listing(t, bundle) != before {
			t.Errorf("%s: bundle changed", s.step)
		}
		if left, _ := os.ReadDir(tmp); len(left) != 0 {
			t.Errorf("%s: temporary copy made: %v", s.step, left)
		}
	}
}

func TestRunRefusesRecordInsideOutDir(t *testing.T) {
	root, kept := evidenceRoot(t), keptDir(t)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(kept, link); err != nil {
		t.Fatal(err)
	}
	records := []string{
		filepath.Join(kept, "rec"),
		filepath.Join(kept, "a", "b", "rec"),
		filepath.Join(link, "rec"),
	}
	if p := alias(kept); p != "" {
		records = append(records, filepath.Join(p, "rec"))
	}
	before := listing(t, kept)
	for _, record := range records {
		rec, err := Run(RunOptions{Dir: root, Ref: "chk@1", Step: "evidence", Args: []string{kept}, Record: record})
		if rec != nil || code(err) != 2 || !strings.Contains(err.Error(), "OUT_DIR "+kept) {
			t.Errorf("--record %s: rec = %v, err = %v; want exit 2 naming OUT_DIR", record, rec, err)
		}
	}
	if listing(t, kept) != before {
		t.Error("OUT_DIR changed")
	}
	beside := filepath.Join(filepath.Dir(kept), "rec")
	_, err := Run(RunOptions{Dir: root, Ref: "chk@1", Step: "evidence", Args: []string{kept}, Record: beside})
	if cf := failed(t, err); *cf.Row.Exit != 30 {
		t.Errorf("beside OUT_DIR: row = %+v", cf.Row)
	}
}

func TestRunRefusesRecordInOuterCopy(t *testing.T) {
	files := passCase(map[string]string{
		"recipe.md": recipeMD("tool", "## Standalone\n\nRun touch: ./touch.sh\n"),
		"touch.sh":  "#!/bin/sh\ntouch \"$1\"\n",
	}, "standalone", "touch", []string{"$CASE/x"}, 0, "")
	repo := recipeRoot(t, "tool@1", files, "touch.sh")
	run(t, repo, "init", "-q")
	sub := filepath.Join(repo, "sub")
	for rel, body := range files {
		writeFile(t, filepath.Join(sub, "recipes", "tool@1", rel), body)
	}
	if err := os.Chmod(filepath.Join(sub, "recipes", "tool@1", "touch.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(repo, "recipes", "tool@1", "rec")
	mark := filepath.Join(t.TempDir(), "mark")
	rec, err := Run(RunOptions{Dir: sub, Ref: "tool@1", Step: "standalone", Label: "touch", Args: []string{mark}, Record: record})
	if rec != nil || code(err) != 2 || !strings.Contains(err.Error(), "recipe library "+filepath.Join(repo, "recipes")+",") {
		t.Fatalf("rec = %v, err = %v; want exit 2 naming the outer bundle", rec, err)
	}
	if exists(record) || exists(mark) {
		t.Error("record or marker written")
	}
}

func TestRunRecordUnresolvable(t *testing.T) {
	root, kept := evidenceRoot(t), keptDir(t)
	file := filepath.Join(t.TempDir(), "file")
	writeFile(t, file, "x\n")
	rec, err := Run(RunOptions{Dir: root, Ref: "chk@1", Step: "evidence", Args: []string{kept}, Record: filepath.Join(file, "rec")})
	if rec != nil || code(err) != 3 || !strings.Contains(err.Error(), "nothing ran") {
		t.Fatalf("rec = %v, err = %v; want exit 3, nothing ran", rec, err)
	}
}

func TestRunRefusesRecordInLibrary(t *testing.T) {
	repo := t.TempDir()
	run(t, repo, "init", "-q")
	sub := filepath.Join(repo, "sub")
	writeFile(t, filepath.Join(sub, "recipes", "tool@1", "recipe.md"), recipeMD("tool", "## Standalone\n\nRun touch: ./touch.sh\n"))
	writeFile(t, filepath.Join(sub, "recipes", "tool@1", "touch.sh"), "#!/bin/sh\ntouch \"$1\"\n")
	for rel, body := range passCase(map[string]string{}, "standalone", "touch", []string{"$CASE/x"}, 0, "") {
		writeFile(t, filepath.Join(sub, "recipes", "tool@1", rel), body)
	}
	if err := os.Chmod(filepath.Join(sub, "recipes", "tool@1", "touch.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "recipes", "other@1", "recipe.md"), recipeMD("other", "Notes.\n"))
	for _, record := range []string{
		filepath.Join(sub, "recipes", "rec"),
		filepath.Join(sub, "recipes", "other@1", "rec"),
		filepath.Join(repo, "recipes", "new", "rec"),
		filepath.Join(repo, "recipes"),
	} {
		mark := filepath.Join(t.TempDir(), "mark")
		rec, err := Run(RunOptions{Dir: sub, Ref: "tool@1", Step: "standalone", Label: "touch", Args: []string{mark}, Record: record})
		if rec != nil || code(err) != 2 || !strings.Contains(err.Error(), "recipe library") {
			t.Errorf("--record %s: rec = %v, err = %v; want exit 2 naming the library", record, rec, err)
		}
		if exists(record) || exists(mark) {
			t.Errorf("--record %s: record or marker written", record)
		}
	}
	if exists(filepath.Join(repo, "recipes")) {
		t.Error("a refused record started a library")
	}
	beside := filepath.Join(sub, "rec")
	if _, err := Run(RunOptions{Dir: sub, Ref: "tool@1", Step: "standalone", Label: "touch", Args: []string{filepath.Join(t.TempDir(), "m")}, Record: beside}); err != nil {
		t.Fatalf("beside the library: %v", err)
	}
}

func TestRunRecordRelativeToDir(t *testing.T) {
	root := recipeRoot(t, "tool", passCase(map[string]string{
		"recipe.md": recipeMD("tool", "## Standalone\n\nRun echo: ./echo.sh\n"),
		"echo.sh":   "#!/bin/sh\n",
	}, "standalone", "echo", nil, 0, ""), "echo.sh")
	cwd := t.TempDir()
	t.Chdir(cwd)
	writeFile(t, filepath.Join(root, "taken", "x"), "keep\n")
	if _, err := Run(RunOptions{Dir: root, Ref: "tool@1", Step: "standalone", Label: "echo", Record: "taken"}); code(err) != 2 {
		t.Errorf("existing Dir/taken: err = %v, want exit 2", err)
	}
	if _, err := Run(RunOptions{Dir: root, Ref: "tool@1", Step: "standalone", Label: "echo", Record: "out/rec"}); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(root, "out", "rec", "command.json")) || exists(filepath.Join(cwd, "out")) {
		t.Error("a relative record was not made under Dir")
	}
}

func TestRunRefusesRecordSpelledInOtherCase(t *testing.T) {
	root, kept := evidenceRoot(t), keptDir(t)
	if !caseInsensitive(t, kept) {
		t.Skip("case-sensitive disk")
	}
	for _, c := range []struct{ record, want string }{
		{filepath.Join(root, "RECIPES", "chk", "rec"), "recipe library"},
		{filepath.Join(root, "Recipes", "rec"), "recipe library"},
		{filepath.Join(filepath.Dir(kept), "OUT", "rec"), "OUT_DIR"},
	} {
		rec, err := Run(RunOptions{Dir: root, Ref: "chk@1", Step: "evidence", Args: []string{kept}, Record: c.record})
		if rec != nil || code(err) != 2 || !strings.Contains(err.Error(), c.want) {
			t.Errorf("--record %s: rec = %v, err = %v; want exit 2 naming %s", c.record, rec, err, c.want)
		}
	}
	if exists(filepath.Join(root, "recipes", "chk", "rec")) || exists(filepath.Join(kept, "rec")) {
		t.Error("a record was written")
	}
}
