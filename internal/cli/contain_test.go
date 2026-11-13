package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/contain"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/store"
)

// cstore is a store with a runnable bundle ana@1, whose known answer passes,
// and a harn@1 with no commands. Under qi, Experiment ei of hi follows ana@1
// and Result i0 cites its cases run. Under qm, the measurement Experiment em
// follows harn@1 from hf and targets ht. r0 is the command.json of an ana@1
// Evidence run.
type cstore struct {
	dir, g, qi, hi, ei, i0, qm, hf, ht, em, r0, hash, h12 string
}

func containStore(t *testing.T) cstore {
	t.Helper()
	dir, g, _ := setup(t)
	bundleIn(t, dir, "ana", "## Evidence\n\nRun: ./ev.py\n\n## Standalone\n\nRun cases: ./cases.py\n", map[string]string{
		"ev.py":                        "#!/bin/sh\necho ok\n",
		"cases.py":                     "#!/bin/sh\necho ok\n",
		"known-answers/ok/expect.json": `{"step": "evidence", "command": "./ev.py", "exit": 0, "stdout": "ok\n"}`,
		"known-answers/ok/input/a":     "",
	})
	bundleIn(t, dir, "harn", "Run the harness by hand.\n", nil)
	s := cstore{dir: dir, g: g}
	_, hash, err := recipe.Resolve(dir, "ana@1")
	if err != nil {
		t.Fatal(err)
	}
	s.hash, s.h12 = hash, hash[:12]

	s.qi = must(t, "add", "dq", "--parent", g, "--title", "Does ana@1 run?", "--why", "w")
	s.hi = must(t, "add", "hyp", "--parent", s.qi, "--title", "ana@1 gives them", "--why", "w")
	s.ei = must(t, "add", "exp", "--parent", s.hi, "--title", "cases", "--recipe", "ana@1", "--rubric", "reading", "--target", "none", "--why", "w")
	must(t, "bundle", "run", "--record", ".tree/evidence/"+s.ei+"/R1", "ana@1", "cases")
	s.i0 = must(t, "record", "--exp", s.ei, "--verdict", s.hi+"=supports", "--observation", "cases ok",
		"--evidence", ".tree/evidence/"+s.ei+"/R1/command.json", "--why", "w")

	s.qm = must(t, "add", "dq", "--parent", g, "--title", "Which?", "--why", "w")
	s.hf = must(t, "add", "hyp", "--parent", s.qm, "--title", "finds", "--why", "w")
	s.ht = must(t, "add", "hyp", "--parent", s.qm, "--title", "told", "--why", "w")
	s.em = must(t, "add", "exp", "--parent", s.hf, "--title", "measure", "--recipe", "harn@1", "--decided-by", "reading", "--target", s.ht, "--why", "w")

	out := filepath.Join(t.TempDir(), "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "sandbox.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	must(t, "bundle", "evidence", "--record", ".tree/evidence/"+s.em+"/R0", "ana@1", out)
	s.r0 = ".tree/evidence/" + s.em + "/R0/command.json"
	return s
}

// file writes .tree/evidence/<em>/name with mode and returns its pointer.
func (s cstore) file(t *testing.T, name string, mode os.FileMode, body string) string {
	t.Helper()
	ptr := ".tree/evidence/" + s.em + "/" + name
	if err := os.WriteFile(filepath.Join(s.dir, ptr), []byte(body), mode); err != nil {
		t.Fatal(err)
	}
	return ptr
}

// record records a Result of exp judging ht supports and hf contradicts,
// citing cites.
func (s cstore) record(t *testing.T, exp string, cites []string) string {
	t.Helper()
	args := []string{"record", "--exp", exp, "--verdict", s.ht + "=supports", "--verdict", s.hf + "=contradicts",
		"--observation", "o", "--why", "w"}
	for _, c := range cites {
		args = append(args, "--evidence", c)
	}
	return must(t, args...)
}

func reasonsOf(t *testing.T, dir, ref string) []string {
	t.Helper()
	p, _, err := store.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	id, err := resolveRef(p, "", ref)
	if err != nil {
		t.Fatal(err)
	}
	return contain.Reasons(dir, p, id)
}

func wantReasons(t *testing.T, got []string, want ...string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("reasons %q, want %q", got, want)
	}
}

// measured builds a fresh store, records one Result under em, and checks
// its reasons; want builds them from the store.
func measured(t *testing.T, cites func(cstore) []string, want func(cstore) []string) {
	t.Helper()
	s := containStore(t)
	r := s.record(t, s.em, cites(s))
	if got, w := reasonsOf(t, s.dir, r), want(s); !slices.Equal(got, w) {
		t.Fatalf("reasons %q, want %q", got, w)
	}
}

func evPy(cstore) string { return "recipes/ana@1/ev.py" }

func TestContainContainedResultHasNoReason(t *testing.T) {
	measured(t,
		func(s cstore) []string { return []string{evPy(s), s.r0} },
		func(cstore) []string { return nil })
}

func TestContainScriptOutsideRecipes(t *testing.T) {
	s := containStore(t)
	c := s.file(t, "c.py", 0o644, "print('x')\n")
	out := s.file(t, "out.txt", 0o644, "x\n")
	r := s.record(t, s.em, []string{c, out})
	wantReasons(t, reasonsOf(t, s.dir, r), "script outside recipes/: "+c)
}

func TestContainScriptByExecBitAndShebang(t *testing.T) {
	s := containStore(t)
	run := s.file(t, "run", 0o755, "echo x\n")
	tool := s.file(t, "tool", 0o644, "#!/bin/sh\necho x\n")
	r := s.record(t, s.em, []string{run, tool})
	wantReasons(t, reasonsOf(t, s.dir, r), "script outside recipes/: "+run, "script outside recipes/: "+tool)
}

func TestContainPinWithoutCommandRecord(t *testing.T) {
	measured(t,
		func(s cstore) []string { return []string{evPy(s)} },
		func(s cstore) []string {
			return []string{"no command record of ana@1 at " + s.h12 + " with passing known answers"}
		})
}

func TestContainCommandRecordMatchesNoPin(t *testing.T) {
	measured(t,
		func(s cstore) []string { return []string{s.r0} },
		func(s cstore) []string {
			return []string{"command record of ana@1 at " + s.h12 + " matches no pin of this Result"}
		})
}

func TestContainBundleWithNoCommandsNeedsNoRecord(t *testing.T) {
	s := containStore(t)
	r := s.record(t, s.em, []string{"recipes/harn@1/recipe.md"})
	wantReasons(t, reasonsOf(t, s.dir, r))
}

func TestContainFixturePinNeedsNoInstrument(t *testing.T) {
	s := containStore(t)
	bundleIn(t, s.dir, "fix", "Known answers.\n", nil)
	writeFile(t, filepath.Join(s.dir, "recipes", "fix@1", "known.txt"), "k\n")
	r := s.record(t, s.em, []string{"recipes/fix@1/known.txt", s.file(t, "out.txt", 0o644, "x\n")})
	wantReasons(t, reasonsOf(t, s.dir, r))
}

func TestContainBundleFilesGiveEachReasonOnce(t *testing.T) {
	s := containStore(t)
	e := must(t, "add", "exp", "--parent", s.hf, "--title", "own", "--recipe", "ana@1", "--rubric", "reading", "--target", s.ht, "--why", "w")
	r := s.record(t, e, []string{evPy(s), "recipes/ana@1/recipe.md"})
	wantReasons(t, reasonsOf(t, s.dir, r), "no command record of ana@1 at "+s.h12+" with passing known answers")
}

func TestContainSymlinkedScript(t *testing.T) {
	s := containStore(t)
	target := filepath.Join(t.TempDir(), "run.sh")
	if err := os.WriteFile(target, []byte("echo x\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := ".tree/evidence/" + s.em + "/run.sh"
	if err := os.Symlink(target, filepath.Join(s.dir, link)); err != nil {
		t.Fatal(err)
	}
	r := s.record(t, s.em, []string{link})
	wantReasons(t, reasonsOf(t, s.dir, r), "script outside recipes/: "+link)
}

func TestContainUnreadableCommandRecord(t *testing.T) {
	s := containStore(t)
	if err := os.MkdirAll(filepath.Join(s.dir, ".tree/evidence", s.em, "bad"), 0o755); err != nil {
		t.Fatal(err)
	}
	bad := s.file(t, "bad/command.json", 0o644, "not json\n")
	empty := s.file(t, "command.json", 0o644, "{}\n")
	r := s.record(t, s.em, []string{bad, empty})
	wantReasons(t, reasonsOf(t, s.dir, r),
		"command record "+bad+" names no recipe run",
		"command record "+empty+" names no recipe run")
}

func TestContainUnversionedRecipesScript(t *testing.T) {
	s := containStore(t)
	if err := os.MkdirAll(filepath.Join(s.dir, "recipes", "scratch"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.dir, "recipes", "scratch", "verdict.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := s.record(t, s.em, []string{"recipes/scratch/verdict.sh"})
	wantReasons(t, reasonsOf(t, s.dir, r), "script outside recipes/: recipes/scratch/verdict.sh")
}

func TestContainRepeatedCitationGivesReasonOnce(t *testing.T) {
	s := containStore(t)
	c := s.file(t, "c.py", 0o644, "print('x')\n")
	must(t, "bundle", "run", "--record", ".tree/evidence/"+s.em+"/R1", "ana@1", "cases")
	r1 := ".tree/evidence/" + s.em + "/R1/command.json"
	r := s.record(t, s.em, []string{c, "./" + c, s.r0, r1})
	wantReasons(t, reasonsOf(t, s.dir, r),
		"script outside recipes/: "+c,
		"command record of ana@1 at "+s.h12+" matches no pin of this Result")
}

func TestContainDriveRecordIsNoRun(t *testing.T) {
	s := containStore(t)
	if err := os.MkdirAll(filepath.Join(s.dir, ".tree/evidence", s.em, "d"), 0o755); err != nil {
		t.Fatal(err)
	}
	d := s.file(t, "d/command.json", 0o644, `{"ref":"ana@1","sha256":"`+s.hash+`","step":"drive"}`)
	r := s.record(t, s.em, []string{evPy(s), d})
	wantReasons(t, reasonsOf(t, s.dir, r), "no command record of ana@1 at "+s.h12+" with passing known answers")
}

// Mark shares its caches across Results; each Result's marks must equal
// what Reasons gives it alone.
func TestContainMarkMatchesReasons(t *testing.T) {
	s := containStore(t)
	must(t, "record", "--exp", s.ei, "--verdict", s.hi+"=supports", "--observation", "again",
		"--evidence", ".tree/evidence/"+s.ei+"/R1/command.json", "--why", "w")
	s.record(t, s.em, []string{s.file(t, "c.py", 0o644, "print('x')\n")})
	s.record(t, s.em, []string{evPy(s), s.r0})
	s.record(t, s.em, []string{s.file(t, "out.txt", 0o644, "x\n")})
	s.record(t, s.em, []string{evPy(s)})
	s.record(t, s.em, []string{s.r0})
	p, _, err := store.Load(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	marked := 0
	for id, n := range p.Nodes {
		if n.Kind != event.KindResult {
			continue
		}
		if got, want := p.Results[id].Uncontained, contain.Reasons(s.dir, p, id); !slices.Equal(got, want) {
			t.Errorf("%s: Mark %q, Reasons %q", id, got, want)
		}
		if len(p.Results[id].Uncontained) > 0 {
			marked++
		}
	}
	if marked < 3 {
		t.Fatalf("only %d Results marked; the test should cover U1, U2 and U3", marked)
	}
}

func TestContainFailedRunIsNoRun(t *testing.T) {
	for name, c := range map[string]struct{ commands, why string }{
		"exited 1":    {`[{"step":"evidence","exit":1}]`, "; in the cited record, a command exited 1"},
		"did not end": {`[{"step":"evidence"}]`, "; in the cited record, a command did not end"},
		"no commands": {`[]`, ""},
	} {
		t.Run(name, func(t *testing.T) {
			s := containStore(t)
			if err := os.MkdirAll(filepath.Join(s.dir, ".tree/evidence", s.em, "f"), 0o755); err != nil {
				t.Fatal(err)
			}
			f := s.file(t, "f/command.json", 0o644, `{"ref":"ana@1","sha256":"`+s.hash+`","step":"evidence","commands":`+c.commands+`,`+passingKA+`}`)
			r := s.record(t, s.em, []string{evPy(s), f})
			wantReasons(t, reasonsOf(t, s.dir, r), "no command record of ana@1 at "+s.h12+" with passing known answers"+c.why)
		})
	}
}

// passingKA is the known_answers of a command record whose one case passed.
const passingKA = `"known_answers":{"hyp_sha256":"x","cases":[{"name":"a","outcome":"ok"}]}`

// A clean run record counts only with known answers that ran and passed.
func TestContainRecordKnownAnswers(t *testing.T) {
	for name, ka := range map[string]string{
		"passing":     "," + passingKA,
		"without":     "",
		"no cases":    `,"known_answers":{"hyp_sha256":"x","cases":[]}`,
		"failed case": `,"known_answers":{"hyp_sha256":"x","cases":[{"name":"a","outcome":"ok"},{"name":"b","outcome":"stdout differs"}]}`,
		"null":        `,"known_answers":null`,
	} {
		t.Run(name, func(t *testing.T) {
			s := containStore(t)
			if err := os.MkdirAll(filepath.Join(s.dir, ".tree/evidence", s.em, "k"), 0o755); err != nil {
				t.Fatal(err)
			}
			k := s.file(t, "k/command.json", 0o644, `{"ref":"ana@1","sha256":"`+s.hash+`","step":"evidence","commands":[{"step":"evidence","exit":0}]`+ka+`}`)
			r := s.record(t, s.em, []string{evPy(s), k})
			if name == "passing" {
				wantReasons(t, reasonsOf(t, s.dir, r))
				return
			}
			wantReasons(t, reasonsOf(t, s.dir, r), "no command record of ana@1 at "+s.h12+" with passing known answers; in the cited record, the known answers did not pass")
		})
	}
}

func TestContainUnhashedBundleFileIsOutside(t *testing.T) {
	s := containStore(t)
	for _, f := range []string{".hidden.py", "__pycache__/run.sh"} {
		file := filepath.Join(s.dir, "recipes", "ana@1", f)
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	r := s.record(t, s.em, []string{"recipes/ana@1/.hidden.py", "recipes/ana@1/__pycache__/run.sh", evPy(s), s.r0})
	wantReasons(t, reasonsOf(t, s.dir, r),
		"script outside recipes/: recipes/ana@1/.hidden.py",
		"script outside recipes/: recipes/ana@1/__pycache__/run.sh")
}

func TestContainURLIsNoRecord(t *testing.T) {
	s := containStore(t)
	r := s.record(t, s.em, []string{"https://example.com/x/command.json", "https://example.com/run.sh"})
	wantReasons(t, reasonsOf(t, s.dir, r))
}

func TestContainRetractedResultSkipped(t *testing.T) {
	s := containStore(t)
	c := s.file(t, "c.py", 0o644, "print('x')\n")
	r := s.record(t, s.em, []string{c})
	must(t, "retract", r, "--why", "w")
	p, _, err := store.Load(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	id, err := resolveRef(p, "", r)
	if err != nil {
		t.Fatal(err)
	}
	contain.Mark(s.dir, &p)
	if got := p.Results[id].Uncontained; got != nil {
		t.Fatalf("retracted %s marked: %q", r, got)
	}
}

func TestContainOwnRunnableRecipeNeedsRecord(t *testing.T) {
	s := containStore(t)
	e := must(t, "add", "exp", "--parent", s.hf, "--title", "own", "--recipe", "ana@1", "--rubric", "reading", "--target", s.ht, "--why", "w")
	out := ".tree/evidence/" + e + "/out.txt"
	writeFile(t, filepath.Join(s.dir, out), "x\n")
	r := s.record(t, e, []string{out})
	wantReasons(t, reasonsOf(t, s.dir, r), "no command record of ana@1 at "+s.h12+" with passing known answers")
}

// readers runs every read command that shows the mark, for Result r of
// Experiment em under qm.
func (s cstore) readers(t *testing.T, r string) map[string]result {
	t.Helper()
	return map[string]result{
		"show R":       hyp(t, "show", r),
		"show E":       hyp(t, "show", s.em),
		"show Q":       hyp(t, "show", s.qm),
		"status":       hyp(t, "status", "--max-lines", "0"),
		"check":        hyp(t, "check"),
		"render":       hyp(t, "render", "--markdown"),
		"show R as-of": hyp(t, "show", r, "--as-of", r),
		"status as-of": hyp(t, "status", "--max-lines", "0", "--as-of", r),
		"render as-of": hyp(t, "render", "--markdown", "--as-of", r),
	}
}

func TestUncontainedShowsEverywhere(t *testing.T) {
	s := containStore(t)
	c := s.file(t, "c.py", 0o644, "print('x')\n")
	rec := hyp(t, "record", "--exp", s.em, "--verdict", s.ht+"=supports", "--verdict", s.hf+"=contradicts",
		"--observation", "o", "--evidence", c, "--why", "w")
	want(t, rec, 0, "warn: uncontained R-")
	r := strings.Fields(rec.stdout)[0]
	out := s.readers(t, r)
	for name, sub := range map[string]string{
		"show R":       "uncontained: script outside recipes/:",
		"show E":       " UNCONTAINED",
		"show Q":       "uncontained " + r,
		"status":       "UNCONTAINED 1 results",
		"check":        "warn: uncontained " + r,
		"render":       "— uncontained",
		"show R as-of": "uncontained:",
		"status as-of": "UNCONTAINED 1 results",
		"render as-of": "— uncontained",
	} {
		if res := out[name]; res.code != 0 || !strings.Contains(res.stdout, sub) {
			t.Errorf("%s: exit %d, stdout %q, want %q", name, res.code, res.stdout, sub)
		}
	}
}

func TestContainedShowsNothing(t *testing.T) {
	s := containStore(t)
	rec := hyp(t, "record", "--exp", s.em, "--verdict", s.ht+"=supports", "--verdict", s.hf+"=contradicts",
		"--observation", "o", "--evidence", evPy(s), "--evidence", s.r0, "--why", "w")
	if rec.code != 0 || strings.Contains(rec.stderr, "uncontained") {
		t.Fatalf("record: %+v", rec)
	}
	r := strings.Fields(rec.stdout)[0]
	for name, res := range s.readers(t, r) {
		if res.code != 0 || strings.Contains(strings.ToLower(res.stdout+res.stderr), "uncontained") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", name, res.code, res.stdout, res.stderr)
		}
	}
}

func TestContainCitedSourceFile(t *testing.T) {
	s := containStore(t)
	g := s.file(t, "check.go", 0o644, "package main\n")
	ts := s.file(t, "tool.ts", 0o644, "let x = 1\n")
	out := s.file(t, "out.txt", 0o644, "x\n")
	r := s.record(t, s.em, []string{g, ts, out})
	wantReasons(t, reasonsOf(t, s.dir, r), "script outside recipes/: "+g, "script outside recipes/: "+ts)
}

func TestContainSourceFileInBundle(t *testing.T) {
	s := containStore(t)
	bundleIn(t, s.dir, "src", "Sources.\n", nil)
	writeFile(t, filepath.Join(s.dir, "recipes", "src@1", "check.go"), "package main\n")
	r := s.record(t, s.em, []string{"recipes/src@1/check.go"})
	wantReasons(t, reasonsOf(t, s.dir, r))
}

// A Result pinned to an instrument whose bundle is later deleted is reported
// by hyp bundle check as a drift to a missing bundle, and fails it.
func TestDeletedInstrumentBundleDrifts(t *testing.T) {
	s := containStore(t)
	if err := os.Mkdir(filepath.Join(s.dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	want(t, hyp(t, "bundle", "check"), 0, "0 drift")
	if err := os.RemoveAll(filepath.Join(s.dir, "recipes", "ana@1")); err != nil {
		t.Fatal(err)
	}
	r := hyp(t, "bundle", "check")
	want(t, r, 1, "drift: . "+s.i0+" recipes/ana@1/: recorded "+s.hash+", now missing")
	want(t, r, 1, "recipe check found 1 failures")
}
