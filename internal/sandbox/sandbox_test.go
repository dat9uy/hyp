package sandbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/store"
)

// pathDir makes a PATH directory holding only the named real tools.
func pathDir(t *testing.T, tools ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range tools {
		real, err := exec.LookPath(name)
		if err != nil {
			t.Skipf("%s not on PATH", name)
		}
		if err := os.Symlink(real, filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// setup points HOME at a temp dir and PATH at git and lsof only.
func setup(t *testing.T) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := pathDir(t, "git", "lsof")
	t.Setenv("PATH", dir)
	return dir
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func host(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(dir, "a.txt"), "a\n")
	run(t, dir, "add", "a.txt")
	run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "init")
	return dir, run(t, dir, "rev-parse", "HEAD")
}

func script(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "stub")
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func stubHyp(t *testing.T) string { return script(t, "#!/bin/sh\necho hyp 0.0.0-test\n") }

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

// code numbers a kind the way hyp's exit codes do: 1 refusal, 2 bad arguments, 3 environment.
func code(err error) int {
	if _, ok := errors.AsType[*herr.Error](err); ok {
		return int(herr.KindOf(err))
	}
	return -1
}

func boxDir(t *testing.T, id string) string {
	t.Helper()
	r, err := root()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(r, id)
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// create makes sandbox id and fails the test on error.
func create(t *testing.T, id string) (*Record, string) {
	t.Helper()
	h, base := host(t)
	rec, err := Create(CreateOptions{Refs: []string{"none"}, ID: id, Host: h, Base: base, Hyp: stubHyp(t)})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	return rec, boxDir(t, id)
}

func TestCreateRemoveRoundTrip(t *testing.T) {
	setup(t)
	h, base := host(t)
	rec, err := Create(CreateOptions{Refs: []string{"none"}, ID: "s1", Host: h, Base: base, Hyp: stubHyp(t)})
	if err != nil {
		t.Fatal(err)
	}
	box := boxDir(t, "s1")
	wt := filepath.Join(box, "wt")
	if !exists(filepath.Join(wt, "a.txt")) {
		t.Error("wt/a.txt missing")
	}
	if out := run(t, wt, "remote"); out != "" {
		t.Errorf("remotes = %q, want none", out)
	}
	if exec.Command("git", "-C", wt, "symbolic-ref", "-q", "HEAD").Run() == nil {
		t.Error("HEAD is not detached")
	}
	if ents, _ := os.ReadDir(filepath.Join(box, "tmp")); len(ents) != 0 {
		t.Errorf("tmp has %d entries", len(ents))
	}
	if rec.HypVersion != "hyp 0.0.0-test" || rec.Base != base || rec.Host != filepath.Base(h) {
		t.Errorf("record = %+v", rec)
	}
	var onDisk map[string]any
	if data, _ := os.ReadFile(filepath.Join(box, "sandbox.json")); json.Unmarshal(data, &onDisk) != nil {
		t.Fatalf("sandbox.json = %s", data)
	}
	want := []string{"base", "created", "format", "host", "hyp_sha256", "hyp_version", "id", "recipes"}
	if keys := slices.Sorted(maps.Keys(onDisk)); !slices.Equal(keys, want) {
		t.Errorf("record keys = %v, want %v", keys, want)
	}

	writeFile(t, filepath.Join(wt, ".tree/events/x.jsonl"), "{}\n")
	writeFile(t, filepath.Join(wt, ".tree/evidence/E-1/n.txt"), "n\n")
	writeFile(t, filepath.Join(wt, ".tree/index.sqlite"), "db")
	writeFile(t, filepath.Join(box, "tmp/keep.txt"), "k\n")
	meta := filepath.Join(box, "sandbox.json")
	var m map[string]any
	data, _ := os.ReadFile(meta)
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	m["store"] = "s"
	data, _ = json.Marshal(m)
	writeFile(t, meta, string(data))

	out := filepath.Join(t.TempDir(), "out")
	var warn bytes.Buffer
	if _, err := Remove(RemoveOptions{ID: "s1", OutDir: out, Keep: []string{"tmp/keep.txt", "tmp/missing.txt"}, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"tree/events/x.jsonl", "tree/evidence/E-1/n.txt", "keep.txt"} {
		if !exists(filepath.Join(out, p)) {
			t.Errorf("%s missing from OUT_DIR", p)
		}
	}
	if exists(filepath.Join(out, "tree/index.sqlite")) {
		t.Error("index.sqlite was copied")
	}
	var saved map[string]any
	data, _ = os.ReadFile(filepath.Join(out, "sandbox.json"))
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved["removed"] == nil || saved["store"] != "s" {
		t.Errorf("saved record = %v", saved)
	}
	if !strings.Contains(warn.String(), "tmp/missing.txt") {
		t.Errorf("warn = %q", warn.String())
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

func TestCreateRejectsBadArguments(t *testing.T) {
	setup(t)
	h, base := host(t)
	sub := filepath.Join(h, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	hyp := stubHyp(t)
	if _, err := Create(CreateOptions{Refs: []string{"none"}, ID: "taken", Host: h, Base: base, Hyp: hyp}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		o    CreateOptions
	}{
		{"bad id", CreateOptions{Refs: []string{"none"}, ID: "Bad_ID", Host: h, Base: base, Hyp: hyp}},
		{"id taken", CreateOptions{Refs: []string{"none"}, ID: "taken", Host: h, Base: base, Hyp: hyp}},
		{"host subdirectory", CreateOptions{Refs: []string{"none"}, ID: "c1", Host: sub, Base: base, Hyp: hyp}},
		{"base not a commit", CreateOptions{Refs: []string{"none"}, ID: "c2", Host: h, Base: "deadbeef", Hyp: hyp}},
		{"hyp exits 1", CreateOptions{Refs: []string{"none"}, ID: "c3", Host: h, Base: base, Hyp: script(t, "#!/bin/sh\nexit 1\n")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Create(tc.o); code(err) != 2 {
				t.Fatalf("err = %v, want code 2", err)
			}
			if tc.o.ID != "taken" && exists(boxDir(t, tc.o.ID)) {
				t.Error("sandbox created")
			}
		})
	}
}

func TestCreateFailureDeletesPartial(t *testing.T) {
	dir := setup(t)
	h, base := host(t)
	gitPath := filepath.Join(dir, "git")
	real, err := os.Readlink(gitPath) // the real git, not this PATH entry
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(gitPath); err != nil {
		t.Fatal(err)
	}
	body := "#!/bin/sh\ncase \"$*\" in *\"remote remove\"*) echo boom >&2; exit 1;; esac\nexec " + real + " \"$@\"\n"
	if err := os.WriteFile(gitPath, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err = Create(CreateOptions{Refs: []string{"none"}, ID: "part", Host: h, Base: base, Hyp: stubHyp(t)})
	if code(err) != 3 || !strings.Contains(err.Error(), "boom") || !strings.Contains(err.Error(), "partial sandbox removed") {
		t.Fatalf("err = %v", err)
	}
	if exists(boxDir(t, "part")) {
		t.Error("partial sandbox left behind")
	}
}

func TestCreateNeedsLsof(t *testing.T) {
	setup(t)
	h, base := host(t)
	t.Setenv("PATH", pathDir(t, "git"))
	_, err := Create(CreateOptions{Refs: []string{"none"}, ID: "nl", Host: h, Base: base, Hyp: stubHyp(t)})
	if code(err) != 3 || !strings.Contains(err.Error(), "lsof") || strings.Contains(err.Error(), "\n") {
		t.Fatalf("err = %v", err)
	}
	r, _ := root()
	if ents, _ := os.ReadDir(r); len(ents) != 0 {
		t.Errorf("%s has %d entries", r, len(ents))
	}
}

func TestRemoveNeedsLsof(t *testing.T) {
	setup(t)
	_, box := create(t, "nl")
	t.Setenv("PATH", pathDir(t, "git"))
	out := filepath.Join(t.TempDir(), "out")
	_, err := Remove(RemoveOptions{ID: "nl", OutDir: out})
	if code(err) != 3 || !strings.Contains(err.Error(), "lsof") {
		t.Fatalf("err = %v", err)
	}
	if !exists(filepath.Join(box, "sandbox.json")) || exists(out) {
		t.Error("remove changed something")
	}
}

func TestRemoveDiscard(t *testing.T) {
	setup(t)
	_, box := create(t, "d1")
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "d1", OutDir: out, Discard: true}); code(err) != 2 {
		t.Fatalf("OutDir and Discard: err = %v, want code 2", err)
	}
	if _, err := Remove(RemoveOptions{ID: "d1", Discard: true}); err != nil {
		t.Fatal(err)
	}
	if exists(box) || exists(out) {
		t.Error("discard left the sandbox or wrote OUT_DIR")
	}
}

func TestRemoveRejectsKeepOutside(t *testing.T) {
	setup(t)
	_, box := create(t, "k1")
	for _, k := range []string{"../x", "/etc/hosts"} {
		_, err := Remove(RemoveOptions{ID: "k1", OutDir: filepath.Join(t.TempDir(), "out"), Keep: []string{k}})
		if code(err) != 2 {
			t.Errorf("--keep %s: err = %v, want code 2", k, err)
		}
	}
	if !exists(filepath.Join(box, "sandbox.json")) {
		t.Error("sandbox removed")
	}
}

func TestRemoveRejectsOutDirInsideSandbox(t *testing.T) {
	setup(t)
	_, box := create(t, "o1")
	if _, err := Remove(RemoveOptions{ID: "o1", OutDir: filepath.Join(box, "out")}); code(err) != 2 {
		t.Fatalf("err = %v, want code 2", err)
	}
	if !exists(filepath.Join(box, "sandbox.json")) || exists(filepath.Join(box, "out")) {
		t.Error("remove changed something")
	}
}

func TestRemoveCorruptRecord(t *testing.T) {
	setup(t)
	_, box := create(t, "c1")
	writeFile(t, filepath.Join(box, "sandbox.json"), `{"id": "c1", "port": 1}`+"\n")
	_, err := Remove(RemoveOptions{ID: "c1", Discard: true})
	if code(err) != 3 || !strings.Contains(err.Error(), "corrupt record") {
		t.Fatalf("err = %v", err)
	}
	if !exists(box) {
		t.Error("sandbox removed")
	}
}

// recipeRoot writes a bundle at <tmp>/recipes/bundleDir(name)/ with mode 0o755 for the
// names in exec, and returns <tmp>.
func recipeRoot(t *testing.T, name string, files map[string]string, exec ...string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(dir, "recipes", bundleDir(name), rel)
		writeFile(t, p, body)
		if slices.Contains(exec, rel) {
			if err := os.Chmod(p, 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	return dir
}

// bundleDir is the library directory for a fixture bundle: a name with "@" is
// used as given, any other name is version 1.
func bundleDir(name string) string {
	if strings.Contains(name, "@") {
		return name
	}
	return name + "@1"
}

// createWith fills Host, Base and Hyp the way create does.
func createWith(t *testing.T, id string, o CreateOptions) (*Record, string, error) {
	t.Helper()
	o.ID = id
	if o.Host == "" {
		o.Host, o.Base = host(t)
	}
	if o.Hyp == "" {
		o.Hyp = stubHyp(t)
	}
	rec, err := Create(o)
	return rec, boxDir(t, id), err
}

// bundle is a recipe.md with the given Launch, Doctor and Cleanup lines.
func bundle(name, launch, doctor, cleanup string) string {
	s := "---\nname: " + name + "\nversion: 1\nformat: 1\n---\n"
	for _, sec := range [][2]string{{"Launch", launch}, {"Doctor", doctor}, {"Cleanup", cleanup}} {
		if sec[1] != "" {
			s += "\n## " + sec[0] + "\n\n" + sec[1] + "\n"
		}
	}
	return s
}

func TestCreateRecipeNone(t *testing.T) {
	setup(t)
	rec, box := create(t, "nr")
	if !reflect.DeepEqual(rec.Recipes, []Entry{{Ref: event.RefNone}}) {
		t.Errorf("recipes = %+v", rec.Recipes)
	}
	if data, _ := os.ReadFile(filepath.Join(box, "sandbox.json")); !strings.Contains(string(data), `"recipes": [
    {
      "ref": "none"
    }
  ]`) {
		t.Errorf("sandbox.json = %s", data)
	}
	if exists(filepath.Join(box, "recipes")) || exists(filepath.Join(box, "recipe.log")) {
		t.Error("none made recipes/ or recipe.log")
	}
}

func TestCreateNeedsRecipe(t *testing.T) {
	setup(t)
	h, base := host(t)
	_, err := Create(CreateOptions{ID: "nor", Host: h, Base: base, Hyp: stubHyp(t)})
	if code(err) != 2 || !strings.Contains(err.Error(), "missing --recipe") {
		t.Fatalf("err = %v", err)
	}
	if exists(boxDir(t, "nor")) {
		t.Error("box exists")
	}
}

func TestCreateRunsLaunchAndDoctor(t *testing.T) {
	path := pathDir(t, "git", "lsof", "env", "grep", "cut", "sort")
	setup(t)
	t.Setenv("PATH", path)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md": bundle("boot", "Run: ./launch.sh", "Run: ./doctor.sh", ""),
		"launch.sh": "#!/bin/sh\nprintf '%s\\n' \"$HYP_SANDBOX_ID\" \"$HYP_SANDBOX_WT\" \"$HYP_SANDBOX_TMP\" \"$PATH\" \"$PWD\" > \"$HYP_SANDBOX_TMP/env.txt\"\n" +
			"env | grep '^HYP_SANDBOX_' | cut -d= -f1 | LC_ALL=C sort > \"$HYP_SANDBOX_TMP/vars.txt\"\n",
		"doctor.sh": "#!/bin/sh\n[ -f \"$HYP_SANDBOX_TMP/env.txt\" ]\n",
	}, "launch.sh", "doctor.sh")
	rec, box, err := createWith(t, "rl", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	cp := filepath.Join(box, "recipes", "boot")
	if r := rec.Recipes; len(r) != 1 || r[0].Ref.String() != "boot@1" || r[0].Name != "boot" || r[0].Version != "1" ||
		len(r[0].SHA256) != 64 || rec.CopyDir(r[0]) != cp || r[0].Setup != "passed" || r[0].Cleanup != "" {
		t.Errorf("recipes = %+v", r)
	}
	if !exists(filepath.Join(cp, "recipe.md")) {
		t.Error("recipes/boot/recipe.md missing")
	}
	// The agent works in wt/; the copy it could read or edit is not there.
	if exists(filepath.Join(box, "wt", "recipes")) {
		t.Error("wt/recipes exists")
	}
	data, err := os.ReadFile(filepath.Join(box, "tmp", "env.txt"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("env.txt = %q", data)
	}
	if want := []string{"rl", rec.WTDir(), rec.TmpDir()}; !reflect.DeepEqual(lines[:3], want) {
		t.Errorf("env = %q, want %q", lines[:3], want)
	}
	if !strings.HasPrefix(lines[3], filepath.Join(box, "bin")+":") {
		t.Errorf("PATH = %q", lines[3])
	}
	if vars, _ := os.ReadFile(filepath.Join(box, "tmp", "vars.txt")); string(vars) != "HYP_SANDBOX_ID\nHYP_SANDBOX_RECIPE_TMP\nHYP_SANDBOX_TMP\nHYP_SANDBOX_WT\n" {
		t.Errorf("HYP_SANDBOX_ variables = %q", vars)
	}
	pwd, _ := filepath.EvalSymlinks(lines[4])
	want, _ := filepath.EvalSymlinks(cp)
	if pwd != want {
		t.Errorf("PWD = %q, want %q", pwd, want)
	}
	log, _ := os.ReadFile(filepath.Join(box, "recipe.log"))
	for _, s := range []string{"== boot@1 launch: ./launch.sh", "== boot@1 doctor: ./doctor.sh"} {
		if !strings.Contains(string(log), s) {
			t.Errorf("recipe.log = %q, missing %q", log, s)
		}
	}
}

func TestCreateDoctorFailsRemovesSandbox(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "cleaned")
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "Run: ./doctor.sh", "Run: ./cleanup.sh"),
		"doctor.sh":  "#!/bin/sh\necho not ready; exit 1\n",
		"cleanup.sh": "#!/bin/sh\necho done > " + marker + "\n",
	}, "doctor.sh", "cleanup.sh")
	_, box, err := createWith(t, "df", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if code(err) != 3 {
		t.Fatalf("err = %v, want code 3", err)
	}
	for _, s := range []string{"doctor: ./doctor.sh exited 1", "not ready", "Cleanup run; sandbox removed"} {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("err = %v, missing %q", err, s)
		}
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
	if !exists(marker) {
		t.Error("Cleanup did not run")
	}
}

func TestCreateDoctorFailsWithoutCleanup(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md": bundle("boot", "", "Run: ./doctor.sh", ""),
		"doctor.sh": "#!/bin/sh\nexit 1\n",
	}, "doctor.sh")
	_, box, err := createWith(t, "dn", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if code(err) != 3 || !strings.Contains(err.Error(), "recipe setup failed; no Cleanup to run; sandbox removed") {
		t.Fatalf("err = %v, want code 3 naming no Cleanup", err)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

func TestCreateNamesFailedCleanup(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "Run: ./doctor.sh", "Run: ./cleanup.sh"),
		"doctor.sh":  "#!/bin/sh\necho not ready; exit 1\n",
		"cleanup.sh": "#!/bin/sh\necho stopping; exit 7\n",
	}, "doctor.sh", "cleanup.sh")
	_, box, err := createWith(t, "nc", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if code(err) != 3 {
		t.Fatalf("err = %v, want code 3", err)
	}
	for _, s := range []string{"boot@1 doctor: ./doctor.sh exited 1", "sandbox: boot@1 cleanup: ./cleanup.sh exited 7", "sandbox removed"} {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("err = %v, missing %q", err, s)
		}
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

func TestCreateSetupTimeout(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	prev := setupTimeout
	setupTimeout = time.Second
	t.Cleanup(func() { setupTimeout = prev })
	root := recipeRoot(t, "slow", map[string]string{
		"recipe.md": bundle("slow", "Run: ./launch.sh", "", ""),
		"launch.sh": "#!/bin/sh\nexec sleep 30\n",
	}, "launch.sh")
	start := time.Now()
	_, box, err := createWith(t, "to", CreateOptions{Refs: []string{"slow@1"}, Recipes: root})
	if code(err) != 3 || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("err = %v", err)
	}
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("Create took %v", d)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

func TestCreateRecipeArgumentErrors(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{"recipe.md": bundle("boot", "", "", "")})
	h, base := host(t)
	cases := []struct {
		name string
		o    CreateOptions
		want string
	}{
		{"no recipes dir", CreateOptions{Refs: []string{"x@1"}}, "needs --recipes"},
		{"unknown recipe", CreateOptions{Refs: []string{"nope@1"}, Recipes: root}, "nope"},
	}
	for i, c := range cases {
		id := fmt.Sprintf("ra%d", i)
		if c.name == "unknown recipe" {
			c.o.Host, c.o.Base = h, base
		}
		_, box, err := createWith(t, id, c.o)
		if code(err) != 2 || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want code 2 with %q", c.name, err, c.want)
		}
		if exists(box) {
			t.Errorf("%s: box exists", c.name)
		}
	}
}

// The copy goes to <box>/recipes, so a host that has its own recipes/<name>
// at the base is cloned as it is.
func TestCreateHostWithSameRecipeName(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{"recipe.md": bundle("boot", "", "", "")})
	h, _ := host(t)
	writeFile(t, filepath.Join(h, "recipes", "boot", "recipe.md"), "x\n")
	run(t, h, "add", "recipes")
	run(t, h, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "recipe")
	rec, box, err := createWith(t, "hs", CreateOptions{Refs: []string{"boot@1"}, Recipes: root, Host: h, Base: run(t, h, "rev-parse", "HEAD")})
	if err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(filepath.Join(rec.WTDir(), "recipes", "boot", "recipe.md")); string(data) != "x\n" {
		t.Errorf("wt/recipes/boot/recipe.md = %q, want the host's", data)
	}
	if !exists(filepath.Join(box, "recipes", "boot", "recipe.md")) {
		t.Error("recipes/boot/recipe.md missing")
	}
}

func TestRemoveRunsRecipeCleanup(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "cleaned")
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "", "Run: ./cleanup.sh"),
		"cleanup.sh": "#!/bin/sh\necho done > " + marker + "\n",
	}, "cleanup.sh")
	_, box, err := createWith(t, "rc", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "rc", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	if !exists(marker) {
		t.Error("Cleanup did not run")
	}
	if log, _ := os.ReadFile(filepath.Join(out, "recipe.log")); !strings.Contains(string(log), "== boot@1 cleanup: ./cleanup.sh") {
		t.Errorf("OUT_DIR/recipe.log = %q", log)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

func TestRemoveWarnsOnFailedCleanup(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "", "Run: ./cleanup.sh"),
		"cleanup.sh": "#!/bin/sh\necho stopping; exit 7\n",
	}, "cleanup.sh")
	_, box, err := createWith(t, "cw", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	var warn bytes.Buffer
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "cw", OutDir: out, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(warn.String(), "sandbox:"); n != 1 {
		t.Errorf("warning = %q, want one sandbox: line", warn.String())
	}
	if !strings.Contains(warn.String(), "cleanup: ./cleanup.sh exited 7") {
		t.Errorf("warning = %q", warn.String())
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

func TestRemoveWarnsOnCleanupTimeout(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	prev := cleanupTimeout
	cleanupTimeout = time.Second
	t.Cleanup(func() { cleanupTimeout = prev })
	root := recipeRoot(t, "slow", map[string]string{
		"recipe.md":  bundle("slow", "", "", "Run: ./cleanup.sh"),
		"cleanup.sh": "#!/bin/sh\nexec sleep 30\n",
	}, "cleanup.sh")
	_, box, err := createWith(t, "ct", CreateOptions{Refs: []string{"slow@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	var warn bytes.Buffer
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "ct", OutDir: out, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(warn.String(), "timed out after 1s") {
		t.Errorf("warning = %q", warn.String())
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

func TestRemoveRefusesKeepRecipeLog(t *testing.T) {
	setup(t)
	_, box := create(t, "kr")
	if _, err := Remove(RemoveOptions{ID: "kr", OutDir: filepath.Join(t.TempDir(), "out"), Keep: []string{"recipe.log"}}); code(err) != 2 {
		t.Fatalf("err = %v, want code 2", err)
	}
	if !exists(box) {
		t.Error("sandbox removed")
	}
}

// A single file recipes/<name>.md is not a recipe: create finds no recipe,
// exits 2 and makes no sandbox.
func TestCreateRefusesSingleFileRecipe(t *testing.T) {
	setup(t)
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "recipes", "plain.md"), "---\nname: plain\nversion: 2\nformat: 1\n---\n\nJust prose.\n")
	_, box, err := createWith(t, "sf", CreateOptions{Refs: []string{"plain@2"}, Recipes: root})
	if code(err) != 2 || !strings.Contains(err.Error(), "recipe plain@2 not found") {
		t.Fatalf("err = %v (code %d), want exit 2 with not found", err, code(err))
	}
	if exists(box) {
		t.Error("sandbox made")
	}
}

func TestRemoveUnreadableRecipeWarns(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "", "Run: ./cleanup.sh"),
		"cleanup.sh": "#!/bin/sh\n",
	}, "cleanup.sh")
	_, box, err := createWith(t, "ur", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(box, "recipes", "boot", "recipe.md")); err != nil {
		t.Fatal(err)
	}
	var warn bytes.Buffer
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "ur", OutDir: out, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(warn.String(), "unreadable") || !strings.Contains(warn.String(), "no Cleanup ran") {
		t.Errorf("warn = %q", warn.String())
	}
	if saved := savedRecord(t, out); saved.Recipes[0].Cleanup != "failed: no Cleanup ran" {
		t.Errorf("cleanup = %q", saved.Recipes[0].Cleanup)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

func TestCopyRecipeOpensByRef(t *testing.T) {
	src := filepath.Join(t.TempDir(), "boot@1")
	writeFile(t, filepath.Join(src, "recipe.md"), strings.Replace(bundle("boot", "", "", ""), "version: 1", "version: 2", 1))
	rcp := &recipe.Recipe{Name: "boot", Version: "1", Dir: src}
	_, err := copyRecipe("boot@1", rcp, t.TempDir())
	if code(err) != 1 || !strings.Contains(err.Error(), "recipe boot@1: recipe.md has version 2") {
		t.Fatalf("copyRecipe = %v, want a refusal naming the version", err)
	}
}

func TestRemoveDriftedCopyRunsNoCleanup(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "cleaned")
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "", "Run: ./cleanup.sh"),
		"cleanup.sh": "#!/bin/sh\n",
	}, "cleanup.sh")
	rec, box, err := createWith(t, "dc", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(box, "recipes", "boot", "cleanup.sh"), "#!/bin/sh\n: > "+marker+"\n")
	var warn bytes.Buffer
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "dc", OutDir: out, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	if exists(marker) {
		t.Error("Cleanup ran from a copy that no longer hashes as the record says")
	}
	if w := warn.String(); !strings.Contains(w, "boot@1 at "+rec.CopyDir(rec.Recipes[0])) || !strings.Contains(w, "the record says "+rec.Recipes[0].SHA256) ||
		!strings.Contains(w, "copy hashes ") || !strings.Contains(w, "no Cleanup ran") {
		t.Errorf("warn = %q", w)
	}
	if saved := savedRecord(t, out); saved.Recipes[0].Cleanup != "failed: no Cleanup ran" {
		t.Errorf("cleanup = %q", saved.Recipes[0].Cleanup)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

// waitGone waits up to 5s for pid to exit.
func waitGone(t *testing.T, pid int) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); syscall.Kill(pid, 0) == nil; time.Sleep(100 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatalf("process %d still runs", pid)
		}
	}
}

func TestRemoveRefusesLeftoverProcess(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	rec, box := create(t, "lo")
	cmd := exec.Command("sleep", "300")
	cmd.Dir = rec.WTDir()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	out := filepath.Join(t.TempDir(), "out")
	_, err := Remove(RemoveOptions{ID: "lo", OutDir: out})
	if code(err) != 3 || !strings.Contains(err.Error(), strconv.Itoa(cmd.Process.Pid)+" sleep ") || !strings.Contains(err.Error(), "nothing deleted") ||
		!strings.Contains(err.Error(), "A shell or editor of your own whose working directory is in wt/ counts") {
		t.Fatalf("err = %v", err)
	}
	if !exists(filepath.Join(box, "sandbox.json")) || exists(out) {
		t.Error("remove changed something")
	}
	cmd.Process.Kill()
	cmd.Wait()
	if _, err := Remove(RemoveOptions{ID: "lo", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

func TestCreateSetupLeftoverKeepsSandbox(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep", "python3")
	setup(t)
	t.Setenv("PATH", path)
	root := recipeRoot(t, "srv", map[string]string{
		"recipe.md": bundle("srv", "Run: ./launch.sh", "Run: ./doctor.sh", ""),
		// setsid: it leaves the Launch group, so only the leftover check can see it.
		"launch.sh": "#!/bin/sh\npython3 -c 'import os; os.setsid(); os.execvp(\"sleep\", [\"sleep\", \"300\"])' >/dev/null 2>&1 &\necho $! > \"$HYP_SANDBOX_TMP/pid\"\n",
		"doctor.sh": "#!/bin/sh\nexit 1\n",
	}, "launch.sh", "doctor.sh")
	rec, box, err := createWith(t, "lk", CreateOptions{Refs: []string{"srv@1"}, Recipes: root})
	data, _ := os.ReadFile(filepath.Join(box, "tmp", "pid"))
	pid, perr := strconv.Atoi(strings.TrimSpace(string(data)))
	if perr == nil {
		t.Cleanup(func() { syscall.Kill(pid, syscall.SIGKILL) })
	}
	if code(err) != 3 || !strings.Contains(err.Error(), "still run in") || !strings.Contains(err.Error(), "--discard") {
		t.Fatalf("err = %v", err)
	}
	if !exists(box) || rec != nil {
		t.Fatal("sandbox not kept")
	}
	if perr != nil {
		t.Fatalf("pid file: %v", perr)
	}
	if kept, err := readRecord(filepath.Join(box, "sandbox.json"), "lk"); err != nil || kept.Recipes[0].Setup != "failed" {
		t.Errorf("kept record = %+v, %v; want setup failed", kept, err)
	}
	if !strings.Contains(err.Error(), strconv.Itoa(pid)+" sleep ") {
		t.Errorf("err does not name the leftover: %v", err)
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	waitGone(t, pid)
	if _, err := Remove(RemoveOptions{ID: "lk", Discard: true}); err != nil {
		t.Fatal(err)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

// An outbound connection takes a local port from the ephemeral range. No
// socket, held by hyp or by another process, decides create or remove.
func TestOutboundSocketIgnored(t *testing.T) {
	path := pathDir(t, "git", "lsof", "python3")
	setup(t)
	t.Setenv("PATH", path)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan net.Conn, 2)
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			accepted <- c
		}
	}()
	c, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	py := exec.Command("python3", "-c", "import socket,sys,time; s=socket.create_connection(('127.0.0.1', int(sys.argv[1]))); time.sleep(300)",
		strconv.Itoa(l.Addr().(*net.TCPAddr).Port))
	py.Dir = t.TempDir()
	if err := py.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { py.Process.Kill(); py.Wait() })
	for range 2 {
		select {
		case conn := <-accepted:
			defer conn.Close()
		case <-time.After(10 * time.Second):
			t.Fatal("connection not accepted")
		}
	}
	_, box := create(t, "os")
	if _, err := Remove(RemoveOptions{ID: "os", OutDir: filepath.Join(t.TempDir(), "out")}); err != nil {
		t.Fatal(err)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

// savedRecord reads OUT_DIR/sandbox.json.
func savedRecord(t *testing.T, out string) *Record {
	t.Helper()
	var rec Record
	data, err := os.ReadFile(filepath.Join(out, "sandbox.json"))
	if err != nil || json.Unmarshal(data, &rec) != nil {
		t.Fatalf("OUT_DIR/sandbox.json = %s (%v)", data, err)
	}
	return &rec
}

// An older hyp wrote one recipe object, a port, and copied the bundle into
// wt/recipes. Remove runs Cleanup from that copy and keeps the port.
func TestRemoveReadsOldRecord(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "cleaned")
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "", "Run: ./cleanup.sh"),
		"cleanup.sh": "#!/bin/sh\necho done > " + marker + "\n",
	}, "cleanup.sh")
	rec, box, err := createWith(t, "op", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(box, "wt", "recipes", "boot")
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(box, "recipes", "boot"), old); err != nil {
		t.Fatal(err)
	}
	legacy := map[string]any{
		"id": "op", "host": rec.Host, "base": rec.Base, "wt": rec.WTDir(), "tmp": rec.TmpDir(), "bin": rec.BinDir(),
		"hyp_sha256": rec.HypSHA256, "hyp_version": rec.HypVersion, "created": rec.Created, "port": 55471,
		"recipe": map[string]any{"ref": "boot@1", "name": "boot", "version": "1", "sha256": rec.Recipes[0].SHA256, "path": old, "setup": "passed"},
	}
	data, _ := json.Marshal(legacy)
	writeFile(t, filepath.Join(box, "sandbox.json"), string(data))
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "op", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	if !exists(marker) {
		t.Error("Cleanup did not run from wt/recipes/boot")
	}
	raw, _ := os.ReadFile(filepath.Join(out, "sandbox.json"))
	if !strings.Contains(string(raw), `"port": 55471`) || strings.Contains(string(raw), `"recipe":`) {
		t.Errorf("saved record = %s", raw)
	}
	want := Entry{Ref: event.Ref{Name: "boot", Version: "1"}, Name: "boot", Version: "1", SHA256: rec.Recipes[0].SHA256, Setup: "passed", Cleanup: "ok", legacyPath: old}
	if saved := savedRecord(t, out); !reflect.DeepEqual(saved.Recipes, []Entry{want}) || saved.Removed == "" {
		t.Errorf("saved = %+v", saved)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

// A record with both shapes reads recipes and drops the old recipe object; one
// with neither reads as an empty list.
func TestRecordRecipeShapes(t *testing.T) {
	var both Record
	if err := json.Unmarshal([]byte(`{"id":"x","recipes":[{"ref":"none"}],"recipe":{"ref":"a@1","path":"p"},"port":1}`), &both); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(both.Recipes, []Entry{{Ref: event.RefNone}}) {
		t.Errorf("recipes = %+v", both.Recipes)
	}
	out, _ := json.Marshal(both)
	if strings.Contains(string(out), `"recipe":`) || !strings.Contains(string(out), `"port":1`) {
		t.Errorf("written = %s", out)
	}
	var neither Record
	if err := json.Unmarshal([]byte(`{"id":"x"}`), &neither); err != nil {
		t.Fatal(err)
	}
	if out, _ := json.Marshal(neither); !strings.Contains(string(out), `"recipes":[]`) {
		t.Errorf("written = %s", out)
	}
	if out, _ := json.Marshal(Record{ID: "x", Host: "h", Base: "b"}); !strings.HasPrefix(string(out), `{"base":"b",`) {
		t.Errorf("keys not sorted: %s", out)
	}
}

func TestCreateRecordsSetupNone(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "", "Run: ./cleanup.sh"),
		"cleanup.sh": "#!/bin/sh\n",
	}, "cleanup.sh")
	rec, box, err := createWith(t, "sn", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Recipes[0].Setup != "none" {
		t.Errorf("setup = %q, want none", rec.Recipes[0].Setup)
	}
	if kept, err := readRecord(filepath.Join(box, "sandbox.json"), "sn"); err != nil || kept.Recipes[0].Setup != "none" {
		t.Errorf("sandbox.json = %+v, %v", kept, err)
	}
}

// A failing Cleanup command does not stop the next one, and the warning and
// the record name what happened to each.
func TestRemoveRunsEveryCleanup(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "second")
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md": bundle("boot", "", "", "Run: ./stop1.sh\n\nRun: ./stop2.sh"),
		"stop1.sh":  "#!/bin/sh\necho stopping; exit 7\n",
		"stop2.sh":  "#!/bin/sh\necho done > " + marker + "\n",
	}, "stop1.sh", "stop2.sh")
	if _, _, err := createWith(t, "ec", CreateOptions{Refs: []string{"boot@1"}, Recipes: root}); err != nil {
		t.Fatal(err)
	}
	var warn bytes.Buffer
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "ec", OutDir: out, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	if !exists(marker) {
		t.Error("the second Cleanup command did not run")
	}
	if want := "sandbox: boot@1 cleanup: ./stop1.sh exited 7; ./stop2.sh ok; see recipe.log\n"; warn.String() != want {
		t.Errorf("warning = %q, want %q", warn.String(), want)
	}
	if got := savedRecord(t, out).Recipes[0].Cleanup; got != "failed: ./stop1.sh" {
		t.Errorf("cleanup = %q", got)
	}
}

// After a failed setup, create also runs every Cleanup command and names each.
func TestCreateRunsEveryCleanup(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "second")
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md": bundle("boot", "", "Run: ./doctor.sh", "Run: ./stop1.sh\n\nRun: ./stop2.sh"),
		"doctor.sh": "#!/bin/sh\nexit 1\n",
		"stop1.sh":  "#!/bin/sh\nexit 7\n",
		"stop2.sh":  "#!/bin/sh\necho done > " + marker + "\n",
	}, "doctor.sh", "stop1.sh", "stop2.sh")
	_, box, err := createWith(t, "ce", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if code(err) != 3 || !strings.Contains(err.Error(), "sandbox: boot@1 cleanup: ./stop1.sh exited 7; ./stop2.sh ok") ||
		!strings.Contains(err.Error(), "sandbox removed") {
		t.Fatalf("err = %v", err)
	}
	if !exists(marker) || exists(box) {
		t.Error("the second Cleanup command did not run, or the sandbox is kept")
	}
}

// Each Cleanup command has its own deadline: one that times out leaves the
// next its full time.
func TestRemoveCleanupDeadlineEach(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	prev := cleanupTimeout
	cleanupTimeout = time.Second
	t.Cleanup(func() { cleanupTimeout = prev })
	marker := filepath.Join(t.TempDir(), "second")
	root := recipeRoot(t, "slow", map[string]string{
		"recipe.md": bundle("slow", "", "", "Run: ./stop1.sh\n\nRun: ./stop2.sh"),
		"stop1.sh":  "#!/bin/sh\nexec sleep 30\n",
		"stop2.sh":  "#!/bin/sh\necho done > " + marker + "\n",
	}, "stop1.sh", "stop2.sh")
	if _, _, err := createWith(t, "cd", CreateOptions{Refs: []string{"slow@1"}, Recipes: root}); err != nil {
		t.Fatal(err)
	}
	var warn bytes.Buffer
	if _, err := Remove(RemoveOptions{ID: "cd", OutDir: filepath.Join(t.TempDir(), "out"), Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	if !exists(marker) {
		t.Error("the second Cleanup command did not finish")
	}
	if !strings.Contains(warn.String(), "./stop1.sh timed out after 1s; ./stop2.sh ok") {
		t.Errorf("warning = %q", warn.String())
	}
}

// When the leftover check itself fails after a failed setup, create keeps the
// sandbox and says it could not check, not that processes still run.
func TestCreateCouldNotCheckLeftovers(t *testing.T) {
	dir := setup(t)
	if err := os.Remove(filepath.Join(dir, "lsof")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "lsof"), "#!/bin/sh\nexit 1\n")
	if err := os.Chmod(filepath.Join(dir, "lsof"), 0o755); err != nil {
		t.Fatal(err)
	}
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md": bundle("boot", "", "Run: ./doctor.sh", ""),
		"doctor.sh": "#!/bin/sh\nexit 1\n",
	}, "doctor.sh")
	_, box, err := createWith(t, "cc", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if code(err) != 3 || !strings.Contains(err.Error(), "could not check whether processes still run in") ||
		strings.Contains(err.Error(), "failed and processes still run in") {
		t.Fatalf("err = %v", err)
	}
	if kept, rerr := readRecord(filepath.Join(box, "sandbox.json"), "cc"); rerr != nil || kept.Recipes[0].Setup != "failed" {
		t.Errorf("kept record = %+v, %v; want setup failed", kept, rerr)
	}
}

// hyp run from inside the sandbox, and the lsof it starts, are not leftovers.
func TestRemoveFromInsideSandbox(t *testing.T) {
	setup(t)
	rec, box := create(t, "in")
	t.Chdir(rec.WTDir())
	if _, err := Remove(RemoveOptions{ID: "in", Discard: true}); err != nil {
		t.Fatal(err)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
}

// recipesRoot writes one bundle per name at <tmp>/recipes/bundleDir(name)/, with every
// .sh file executable, and returns <tmp>.
func recipesRoot(t *testing.T, bundles map[string]map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, files := range bundles {
		for rel, body := range files {
			p := filepath.Join(dir, "recipes", bundleDir(name), rel)
			writeFile(t, p, body)
			if strings.HasSuffix(rel, ".sh") {
				if err := os.Chmod(p, 0o755); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	return dir
}

// appendTo is a script line that appends word to file.
func appendTo(file, word string) string { return "echo " + word + " >> " + file + "\n" }

// Two recipes set up in order, each with its own tmp directory, and the
// second can read what the first left in its tmp directory.
func TestCreateStacksRecipes(t *testing.T) {
	setup(t)
	order := filepath.Join(t.TempDir(), "order")
	sh := func(name, phase, extra string) string {
		return "#!/bin/sh\n" + appendTo(order, name+"-"+phase) + "echo \"$HYP_SANDBOX_RECIPE_TMP\" > \"$HYP_SANDBOX_RECIPE_TMP/" + phase + "\"\n" + extra
	}
	root := recipesRoot(t, map[string]map[string]string{
		"env": {"recipe.md": bundle("env", "Run: ./launch.sh", "Run: ./doctor.sh", ""),
			"launch.sh": sh("env", "launch", ""), "doctor.sh": sh("env", "doctor", "")},
		"harness": {"recipe.md": bundle("harness", "Run: ./launch.sh", "Run: ./doctor.sh", ""),
			"launch.sh": sh("harness", "launch", ""),
			"doctor.sh": sh("harness", "doctor", "[ -f \"$HYP_SANDBOX_TMP/env/launch\" ]\n")},
	})
	rec, box, err := createWith(t, "st", CreateOptions{Refs: []string{"env@1", "harness@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(order); string(got) != "env-launch\nenv-doctor\nharness-launch\nharness-doctor\n" {
		t.Errorf("order = %q", got)
	}
	if len(rec.Recipes) != 2 {
		t.Fatalf("recipes = %+v", rec.Recipes)
	}
	for i, name := range []string{"env", "harness"} {
		e := rec.Recipes[i]
		if e.Ref.String() != name+"@1" || e.Name != name || len(e.SHA256) != 64 || e.Setup != "passed" || rec.CopyDir(e) != filepath.Join(box, "recipes", name) {
			t.Errorf("recipes[%d] = %+v", i, e)
		}
		for _, phase := range []string{"launch", "doctor"} {
			got, _ := os.ReadFile(filepath.Join(rec.TmpDir(), name, phase))
			if strings.TrimSpace(string(got)) != filepath.Join(rec.TmpDir(), name) {
				t.Errorf("%s %s: HYP_SANDBOX_RECIPE_TMP = %q", name, phase, got)
			}
		}
	}
	if rec.Recipes[0].SHA256 == rec.Recipes[1].SHA256 {
		t.Error("both entries have one hash")
	}
	if kept, err := readRecord(filepath.Join(box, "sandbox.json"), "st"); err != nil || !reflect.DeepEqual(kept.Recipes, rec.Recipes) {
		t.Errorf("sandbox.json = %+v, %v", kept, err)
	}
}

// A failing second Launch runs the Cleanup of the second recipe and then the
// first, never launches the third, and the server the first left running is
// stopped, so no process has its working directory in the sandbox.
func TestCreateStackFailedLaunchCleansUp(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	dir := t.TempDir()
	order, pidCopy := filepath.Join(dir, "order"), filepath.Join(dir, "pid")
	root := recipesRoot(t, map[string]map[string]string{
		"srv": {"recipe.md": bundle("srv", "Run: ./launch.sh", "", "Run: ./cleanup.sh"),
			"launch.sh": "#!/bin/sh\n" + appendTo(order, "srv-launch") + "sleep 300 >/dev/null 2>&1 &\n" +
				"echo $! > \"$HYP_SANDBOX_RECIPE_TMP/pid\"\necho $! > " + pidCopy + "\n",
			"cleanup.sh": "#!/bin/sh\n" + appendTo(order, "srv-cleanup") + "read pid < \"$HYP_SANDBOX_RECIPE_TMP/pid\"\nkill $pid\n" +
				"while kill -0 $pid 2>/dev/null; do sleep 0.1; done\n"},
		"bad": {"recipe.md": bundle("bad", "Run: ./launch.sh", "", "Run: ./cleanup.sh"),
			"launch.sh":  "#!/bin/sh\n" + appendTo(order, "bad-launch") + "exit 1\n",
			"cleanup.sh": "#!/bin/sh\n" + appendTo(order, "bad-cleanup")},
		"late": {"recipe.md": bundle("late", "Run: ./launch.sh", "", "Run: ./cleanup.sh"),
			"launch.sh":  "#!/bin/sh\n" + appendTo(order, "late-launch"),
			"cleanup.sh": "#!/bin/sh\n" + appendTo(order, "late-cleanup")},
	})
	_, box, err := createWith(t, "sf", CreateOptions{Refs: []string{"srv@1", "bad@1", "late@1"}, Recipes: root})
	data, _ := os.ReadFile(pidCopy)
	pid, perr := strconv.Atoi(strings.TrimSpace(string(data)))
	if perr == nil {
		t.Cleanup(func() { syscall.Kill(pid, syscall.SIGKILL) })
	}
	if code(err) != 3 || !strings.Contains(err.Error(), "bad@1 launch: ./launch.sh exited 1") ||
		!strings.Contains(err.Error(), "recipe setup failed; Cleanup run; sandbox removed") {
		t.Fatalf("err = %v", err)
	}
	if got, _ := os.ReadFile(order); string(got) != "srv-launch\nbad-launch\nbad-cleanup\nsrv-cleanup\n" {
		t.Errorf("order = %q", got)
	}
	if exists(box) {
		t.Error("sandbox still exists")
	}
	if perr != nil {
		t.Fatalf("pid file: %v", perr)
	}
	if syscall.Kill(pid, 0) == nil {
		t.Errorf("process %d that srv's Launch started still runs", pid)
	}
}

// Remove runs every recipe's Cleanup, the last recipe first; a failure in
// either is reported and does not skip the other.
func TestRemoveStackRunsEveryCleanup(t *testing.T) {
	for _, failing := range []string{"env", "harness"} {
		t.Run(failing, func(t *testing.T) {
			setup(t)
			order := filepath.Join(t.TempDir(), "order")
			cleanup := func(name string) string {
				s := "#!/bin/sh\n" + appendTo(order, name)
				if name == failing {
					s += "exit 7\n"
				}
				return s
			}
			root := recipesRoot(t, map[string]map[string]string{
				"env":     {"recipe.md": bundle("env", "", "", "Run: ./cleanup.sh"), "cleanup.sh": cleanup("env")},
				"harness": {"recipe.md": bundle("harness", "", "", "Run: ./cleanup.sh"), "cleanup.sh": cleanup("harness")},
			})
			_, box, err := createWith(t, "sr", CreateOptions{Refs: []string{"env@1", "harness@1"}, Recipes: root})
			if err != nil {
				t.Fatal(err)
			}
			var warn bytes.Buffer
			out := filepath.Join(t.TempDir(), "out")
			if _, err := Remove(RemoveOptions{ID: "sr", OutDir: out, Warn: &warn}); err != nil {
				t.Fatal(err)
			}
			if got, _ := os.ReadFile(order); string(got) != "harness\nenv\n" {
				t.Errorf("order = %q", got)
			}
			if want := "sandbox: " + failing + "@1 cleanup: ./cleanup.sh exited 7; see recipe.log\n"; warn.String() != want {
				t.Errorf("warning = %q, want %q", warn.String(), want)
			}
			for _, e := range savedRecord(t, out).Recipes {
				want := "ok"
				if e.Name == failing {
					want = "failed: ./cleanup.sh"
				}
				if e.Cleanup != want {
					t.Errorf("%s cleanup = %q, want %q", e.Ref, e.Cleanup, want)
				}
			}
			log, _ := os.ReadFile(filepath.Join(out, "recipe.log"))
			if h, e := strings.Index(string(log), "== harness@1 cleanup:"), strings.Index(string(log), "== env@1 cleanup:"); h < 0 || e < h {
				t.Errorf("recipe.log = %q", log)
			}
			if exists(box) {
				t.Error("sandbox still exists")
			}
		})
	}
}

// Each recipe's Launch and Doctor have their own deadline, so two recipes
// that each fit it pass together although their sum does not.
func TestCreateStackDeadlineEach(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	prev := setupTimeout
	setupTimeout = 2 * time.Second
	t.Cleanup(func() { setupTimeout = prev })
	slow := func(name string) map[string]string {
		return map[string]string{"recipe.md": bundle(name, "Run: ./launch.sh", "", ""), "launch.sh": "#!/bin/sh\nexec sleep 1.2\n"}
	}
	root := recipesRoot(t, map[string]map[string]string{"a": slow("a"), "b": slow("b")})
	rec, _, err := createWith(t, "sd", CreateOptions{Refs: []string{"a@1", "b@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Recipes[0].Setup != "passed" || rec.Recipes[1].Setup != "passed" {
		t.Errorf("recipes = %+v", rec.Recipes)
	}
}

func TestCreateStackRefusals(t *testing.T) {
	setup(t)
	root := recipesRoot(t, map[string]map[string]string{"boot": {"recipe.md": bundle("boot", "", "", "")}})
	for i, c := range []struct {
		refs []string
		want string
	}{
		{[]string{"boot@1", "boot@1"}, "recipe boot is already in the stack"},
		{[]string{"boot@1", "boot@2"}, "recipe boot is already in the stack"},
		{[]string{"none", "none"}, "recipe none is already in the stack"},
		{[]string{"none", "boot@1"}, "--recipe none sets nothing up and is given alone"},
		{[]string{"boot@1", "none"}, "--recipe none sets nothing up and is given alone"},
	} {
		_, box, err := createWith(t, fmt.Sprintf("sx%d", i), CreateOptions{Refs: c.refs, Recipes: root})
		if code(err) != 2 || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%v: err = %v, want code 2 with %q", c.refs, err, c.want)
		}
		if exists(box) {
			t.Errorf("%v: box exists", c.refs)
		}
	}
}

// A run passes its inputs the way the recipe contract says: the caller sets
// an environment variable naming a JSON file, and Launch and Doctor inherit it.
func TestLaunchReadsRunInputs(t *testing.T) {
	setup(t)
	params := filepath.Join(t.TempDir(), "params.json")
	const input = `{"cut":"E-0001","snapshot":"s1"}`
	writeFile(t, params, input+"\n")
	t.Setenv("HYP_TEST_RUN_PARAMS", params)
	root := recipesRoot(t, map[string]map[string]string{"in": {
		"recipe.md": bundle("in", "Run: ./launch.sh", "Run: ./doctor.sh", ""),
		"launch.sh": "#!/bin/sh\nread -r line < \"$HYP_TEST_RUN_PARAMS\"\necho \"$line\" > \"$HYP_SANDBOX_RECIPE_TMP/seen.json\"\n",
		"doctor.sh": "#!/bin/sh\nread -r a < \"$HYP_TEST_RUN_PARAMS\"\nread -r b < \"$HYP_SANDBOX_RECIPE_TMP/seen.json\"\n[ \"$a\" = \"$b\" ]\n",
	}})
	rec, _, err := createWith(t, "ri", CreateOptions{Refs: []string{"in@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(rec.TmpDir(), "in", "seen.json")); string(got) != input+"\n" {
		t.Errorf("Launch saw %q, want %q", got, input)
	}
}

// eventFiles maps each non-empty events file in dir to its bytes.
func eventFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) > 0 {
			out[filepath.Base(p)] = string(data)
		}
	}
	return out
}

// TestCreateStoreCut cuts the synthetic store into wt/.tree before the
// recipe's Launch runs. The cut is at the select of q-cache, the store's 19th
// event, all of them runner's: wt/.tree holds the first 19 lines of
// runner.jsonl and no other actor's file.
func TestCreateStoreCut(t *testing.T) {
	src, err := filepath.Abs("../testutil/testdata/synth")
	if err != nil {
		t.Fatal(err)
	}
	at := "01M6A4QYT7T8JQ6FKN93D82WFX"
	runner, err := os.ReadFile(filepath.Join(src, ".tree", "events", "runner.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitAfter(string(runner), "\n")
	if len(lines) < 19 {
		t.Fatalf("runner.jsonl has %d lines, want at least 19", len(lines))
	}
	cutEvents := map[string]string{"runner.jsonl": strings.Join(lines[:19], "")}

	setup(t)
	files := map[string]string{
		"recipe.md": bundle("cut", "Run: ./launch.sh", "", ""),
		"launch.sh": "#!/bin/sh\ntest -d \"$HYP_SANDBOX_WT/.tree/events\" && echo cut-before-launch\n",
	}
	rec, box, err := createWith(t, "c1", CreateOptions{Refs: []string{"cut@1"}, Recipes: recipeRoot(t, "cut", files, "launch.sh"), Store: src, Cut: at})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Store != src || rec.Cut != at || rec.Recipes[0].Setup != "passed" {
		t.Errorf("record store %q cut %q setup %q", rec.Store, rec.Cut, rec.Recipes[0].Setup)
	}
	if got := eventFiles(t, filepath.Join(box, "wt", ".tree", "events")); !maps.Equal(got, cutEvents) {
		t.Errorf("wt/.tree events differ from the expected cut: %v vs %v", slices.Sorted(maps.Keys(got)), slices.Sorted(maps.Keys(cutEvents)))
	}
	if log, _ := os.ReadFile(filepath.Join(box, "recipe.log")); !strings.Contains(string(log), "cut-before-launch") {
		t.Errorf("Launch did not see wt/.tree: %s", log)
	}
	var saved Record
	data, _ := os.ReadFile(filepath.Join(box, "sandbox.json"))
	if err := json.Unmarshal(data, &saved); err != nil || saved.Store != src || saved.Cut != at {
		t.Errorf("sandbox.json store %q cut %q (%v)", saved.Store, saved.Cut, err)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "c1", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	if got := eventFiles(t, filepath.Join(out, "tree", "events")); !maps.Equal(got, cutEvents) {
		t.Errorf("OUT_DIR/tree events differ from the cut")
	}
}

func TestCreateStoreCutRefusals(t *testing.T) {
	setup(t)
	src := t.TempDir()
	if err := store.Init(src); err != nil {
		t.Fatal(err)
	}
	ev, err := store.Append(src, "a", event.Event{Type: event.TypeNodeAdd, Payload: []byte(`{"kind":"Goal","parent":"","title":"g"}`), Why: "w"})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		o    CreateOptions
		msg  string
	}{
		{"unknown cut", CreateOptions{Store: src, Cut: "01NOPE"}, "no such event 01NOPE"},
		{"no store", CreateOptions{Store: t.TempDir(), Cut: ev.ID}, "holds no .tree/events"},
		{"store alone", CreateOptions{Store: src}, "--store needs --cut"},
		{"cut alone, host holds no store", CreateOptions{Cut: ev.ID}, "which holds no store"},
	} {
		t.Run(c.name, func(t *testing.T) {
			c.o.Refs = []string{"none"}
			_, box, err := createWith(t, "r1", c.o)
			if code(err) != 2 || !strings.Contains(err.Error(), c.msg) {
				t.Fatalf("err = %v (code %d), want exit 2 with %q", err, code(err), c.msg)
			}
			if exists(box) {
				t.Error("sandbox made")
			}
		})
	}
	// A line that does not parse has no id and no place in the merge order,
	// so the cut refuses the store instead of skipping the line.
	t.Run("unparseable line", func(t *testing.T) {
		bad := t.TempDir()
		if err := store.Init(bad); err != nil {
			t.Fatal(err)
		}
		g, err := store.Append(bad, "a", event.Event{Type: event.TypeNodeAdd, Payload: []byte(`{"kind":"Goal","parent":"","title":"g"}`), Why: "w"})
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.OpenFile(filepath.Join(bad, ".tree", "events", "a.jsonl"), os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString("{not json\n")
		f.Close()
		_, box, err := createWith(t, "r1", CreateOptions{Store: bad, Cut: g.ID, Refs: []string{"none"}})
		if code(err) != 3 || !strings.Contains(err.Error(), "a.jsonl:2: ") ||
			!strings.Contains(err.Error(), "run hyp check") || !strings.Contains(err.Error(), "nothing created") {
			t.Fatalf("err = %v (code %d), want exit 3 naming a.jsonl:2", err, code(err))
		}
		if exists(box) {
			t.Error("sandbox made")
		}
	})
}

func TestCreateRecordsCommands(t *testing.T) {
	setup(t)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "Run: ./launch.sh", "Run: ./doctor.sh", "Run: ./cleanup.sh"),
		"launch.sh":  "#!/bin/sh\n",
		"doctor.sh":  "#!/bin/sh\n",
		"cleanup.sh": "#!/bin/sh\n",
	}, "launch.sh", "doctor.sh", "cleanup.sh")
	rec, box, err := createWith(t, "rcmd", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Commands) != 2 {
		t.Fatalf("commands = %+v, want 2", rec.Commands)
	}
	for i, want := range []struct{ step, script string }{{"launch", "./launch.sh"}, {"doctor", "./doctor.sh"}} {
		c := rec.Commands[i]
		if string(c.Step) != want.step || !reflect.DeepEqual(c.Args, []string{want.script}) || c.Exit == nil || *c.Exit != 0 || c.Started == "" || c.Ended == "" {
			t.Errorf("commands[%d] = %+v", i, c)
		}
	}
	data, _ := os.ReadFile(filepath.Join(box, "sandbox.json"))
	var m map[string]json.RawMessage
	if json.Unmarshal(data, &m) != nil || m["commands"] == nil {
		t.Errorf("sandbox.json has no commands: %s", data)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "rcmd", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	if c := savedRecord(t, out).Commands; len(c) != 3 || c[2].Step != "cleanup" {
		t.Errorf("OUT_DIR commands = %+v", c)
	}
}

// failRename makes every record rename fail until the test ends.
func failRename(t *testing.T) {
	t.Helper()
	rename = func(string, string) error { return errors.New("rename refused") }
	t.Cleanup(func() { rename = os.Rename })
}

func TestWriteJSONKeepsOldRecordOnFailedRename(t *testing.T) {
	setup(t)
	_, box := create(t, "atom")
	meta := filepath.Join(box, "sandbox.json")
	before, err := readRecord(meta, "atom")
	if err != nil {
		t.Fatal(err)
	}
	changed := *before
	changed.Created = "changed"
	failRename(t)
	if err := writeJSON(meta, &changed); err == nil {
		t.Fatal("writeJSON succeeded with a failing rename")
	}
	after, err := readRecord(meta, "atom")
	if err != nil || !reflect.DeepEqual(after, before) {
		t.Errorf("record = %+v (%v), want the old one %+v", after, err, before)
	}
	names, _ := os.ReadDir(box)
	for _, n := range names {
		if strings.HasPrefix(n.Name(), ".sandbox.json-") {
			t.Errorf("temporary file %s left", n.Name())
		}
	}
	rename = os.Rename
	if _, err := Remove(RemoveOptions{ID: "atom", Discard: true}); err != nil {
		t.Fatal(err)
	}
}

func TestWriteJSONReplacesWithMode0644(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "r.json")
	for _, v := range []string{"one", "two"} {
		if err := writeJSON(p, map[string]string{"v": v}); err != nil {
			t.Fatal(err)
		}
	}
	data, _ := os.ReadFile(p)
	if string(data) != "{\n  \"v\": \"two\"\n}\n" {
		t.Errorf("content = %q", data)
	}
	if fi, err := os.Stat(p); err != nil || fi.Mode().Perm() != 0o644 {
		t.Errorf("mode = %v (%v), want 0644", fi.Mode(), err)
	}
	if names, _ := os.ReadDir(dir); len(names) != 1 {
		t.Errorf("dir holds %d files, want 1", len(names))
	}
}

func TestLockReleasedOnUnlock(t *testing.T) {
	setup(t)
	_, box := create(t, "lk")
	unlock, err := lockBox(box, "lk", "drive")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lockBox(box, "lk", "remove"); code(err) != 3 || !strings.Contains(err.Error(), "hyp sandbox drive") {
		t.Fatalf("second lock: err = %v, want code 3 naming the drive", err)
	}
	unlock()
	unlock2, err := lockBox(box, "lk", "remove")
	if err != nil {
		t.Fatalf("lock after unlock: %v", err)
	}
	unlock2()
}

func TestLaunchRowWrittenBeforeCommand(t *testing.T) {
	path := pathDir(t, "git", "lsof", "cp", "dirname")
	setup(t)
	t.Setenv("PATH", path)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md": bundle("boot", "Run: ./launch.sh", "", ""),
		"launch.sh": "#!/bin/sh\ncp \"$(dirname \"$HYP_SANDBOX_WT\")/sandbox.json\" \"$HYP_SANDBOX_TMP/during.json\"\n",
	}, "launch.sh")
	rec, _, err := createWith(t, "lpend", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	during := filepath.Join(rec.TmpDir(), "during.json")
	pendingRow(t, during, "launch", func() []cmdrecord.Run {
		var mid Record
		data, _ := os.ReadFile(during)
		if err := json.Unmarshal(data, &mid); err != nil {
			t.Fatal(err)
		}
		return mid.Commands
	})
	if len(rec.Commands) != 1 || rec.Commands[0].Ended == "" {
		t.Errorf("rows after create = %+v", rec.Commands)
	}
}

// A killed hyp leaves a last row with started only; remove reads it and
// copies it to OUT_DIR as it was.
func TestPendingRowRecordStillReads(t *testing.T) {
	setup(t)
	_, box := create(t, "kept")
	meta := filepath.Join(box, "sandbox.json")
	rec, err := readRecord(meta, "kept")
	if err != nil {
		t.Fatal(err)
	}
	rec.Commands = append(rec.Commands, cmdrecord.Run{Ref: "boot@1", Step: "drive", Args: []string{"./drive.sh"}, Started: now()})
	if err := writeJSON(meta, rec); err != nil {
		t.Fatal(err)
	}
	if _, err := readRecord(meta, "kept"); err != nil {
		t.Fatalf("readRecord: %v", err)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "kept", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	got := savedRecord(t, out).Commands
	if len(got) != 1 || !reflect.DeepEqual(got[0], rec.Commands[0]) || got[0].Outcome() != "did not end" {
		t.Errorf("OUT_DIR rows = %+v", got)
	}
	if data, _ := os.ReadFile(filepath.Join(out, "sandbox.json")); strings.Contains(string(data), `"ended"`) {
		t.Errorf("OUT_DIR record has an ended key: %s", data)
	}
}

func TestCleanupRunsWhenRecordWriteFails(t *testing.T) {
	path := pathDir(t, "git", "lsof", "touch")
	setup(t)
	t.Setenv("PATH", path)
	marker := filepath.Join(t.TempDir(), "cleaned")
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "", "Run: ./cleanup.sh"),
		"cleanup.sh": "#!/bin/sh\ntouch " + marker + "\n",
	}, "cleanup.sh")
	if _, _, err := createWith(t, "cw", CreateOptions{Refs: []string{"boot@1"}, Recipes: root}); err != nil {
		t.Fatal(err)
	}
	failRename(t)
	var warn bytes.Buffer
	rec, err := Remove(RemoveOptions{ID: "cw", Discard: true, Warn: &warn})
	if err != nil {
		t.Fatal(err)
	}
	if !exists(marker) {
		t.Error("Cleanup did not run")
	}
	if !strings.Contains(warn.String(), "a record write failed") {
		t.Errorf("warn = %q, want the failed write named", warn.String())
	}
	if !strings.HasPrefix(rec.Recipes[0].Cleanup, "failed: ") {
		t.Errorf("cleanup = %q, want failed", rec.Recipes[0].Cleanup)
	}
	if c := rec.Commands; len(c) != 1 || c[0].Step != "cleanup" || c[0].Ended == "" || c[0].Exit == nil || *c[0].Exit != 0 {
		t.Errorf("rows = %+v, want one finished cleanup row", c)
	}
}

// A remove refused for a leftover process keeps the Cleanup rows and value.
func TestRemoveKeepsCleanupInRecord(t *testing.T) {
	path := pathDir(t, "git", "lsof", "sleep")
	setup(t)
	t.Setenv("PATH", path)
	root := recipeRoot(t, "boot", map[string]string{
		"recipe.md":  bundle("boot", "", "", "Run: ./cleanup.sh"),
		"cleanup.sh": "#!/bin/sh\n",
	}, "cleanup.sh")
	rec, box, err := createWith(t, "kc", CreateOptions{Refs: []string{"boot@1"}, Recipes: root})
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sleep", "300")
	cmd.Dir = rec.WTDir()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	if _, err := Remove(RemoveOptions{ID: "kc", Discard: true}); code(err) != 3 {
		t.Fatalf("err = %v, want code 3 leftover", err)
	}
	got := savedRecord(t, box)
	if len(got.Commands) != 1 || got.Commands[0].Step != "cleanup" || got.Recipes[0].Cleanup != "ok" {
		t.Errorf("kept record = %+v", got)
	}
}

// alias is p spelled the other way round the macOS /var and /tmp links to
// /private, or "" when p has no such spelling.
func alias(p string) string {
	for _, d := range []string{"/var/", "/tmp/"} {
		if strings.HasPrefix(p, d) {
			return "/private" + p
		}
		if strings.HasPrefix(p, "/private"+d) {
			return strings.TrimPrefix(p, "/private")
		}
	}
	return ""
}

func TestInsideResolvesDeepestExistingParent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "root")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	cases := map[string]bool{
		filepath.Join(link, "a", "b", "c"): true,
		filepath.Join(link, "a"):           true,
		root:                               false,
		filepath.Join(filepath.Dir(root), "sib", "a", "b"):   false,
		filepath.Join(filepath.Dir(root), "rootx", "a", "b"): false,
	}
	if p := alias(root); p != "" {
		cases[filepath.Join(p, "a", "b")] = true
	}
	// A dangling link is not followed: MkdirAll fails on it, so nothing is
	// made under root.
	dang := filepath.Join(t.TempDir(), "dang")
	if err := os.Symlink(filepath.Join(root, "sub"), dang); err != nil {
		t.Fatal(err)
	}
	cases[filepath.Join(dang, "rec")] = false
	for path, want := range cases {
		if got, err := inside(root, path); err != nil || got != want {
			t.Errorf("inside(root, %s) = %v, %v; want %v", path, got, err, want)
		}
	}
}

func TestInsideErrorsOnUnresolvablePath(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(t.TempDir(), "file")
	writeFile(t, file, "x\n")
	if _, err := inside(root, filepath.Join(file, "rec")); err == nil {
		t.Error("a path under a regular file resolved")
	}
	if _, err := inside(root, "rel/rec"); err == nil {
		t.Error("a missing relative path resolved")
	}
}

// caseInsensitive reports whether the disk holding dir ignores case.
func caseInsensitive(t *testing.T, dir string) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(filepath.Dir(dir), strings.ToUpper(filepath.Base(dir))))
	return err == nil && strings.ToUpper(filepath.Base(dir)) != filepath.Base(dir)
}

func TestCoversByIdentity(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "lib")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(base, "recipes")
	cases := []struct {
		root, path string
		want       bool
	}{
		{root, root, true},
		{root, filepath.Join(link, "a", "b"), true},
		{root, filepath.Join(base, "libx", "a"), false},
		{root, base, false},
		{missing, missing, true},
		{missing, filepath.Join(missing, "x", "rec"), true},
		{missing, filepath.Join(base, "recipesx"), false},
		{missing, filepath.Join(base, "rec"), false},
	}
	if caseInsensitive(t, root) {
		cases = append(cases,
			struct {
				root, path string
				want       bool
			}{root, filepath.Join(base, "LIB", "rec"), true},
			struct {
				root, path string
				want       bool
			}{missing, filepath.Join(base, "RECIPES", "x"), true})
	}
	for _, c := range cases {
		if got, err := covers(c.root, c.path); err != nil || got != c.want {
			t.Errorf("covers(%s, %s) = %v, %v; want %v", c.root, c.path, got, err, c.want)
		}
	}
}
