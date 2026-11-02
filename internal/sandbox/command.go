package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
)

var driveTimeout = 60 * time.Minute      // each Drive command
var evidenceTimeout = 10 * time.Minute   // each Evidence command
var standaloneTimeout = 60 * time.Minute // a Standalone command

// CommandFailed is a Drive, Evidence or Standalone command that exited
// non-zero. hyp exits with Row's exit code unchanged (spec §9.3).
type CommandFailed struct {
	Row  cmdrecord.Run
	Log  string // the log holding its output
	Tail string // the log's last lines, or ""
}

func (e *CommandFailed) Error() string {
	msg := fmt.Sprintf("%s %s: %s %s; see %s", e.Row.Ref, e.Row.Step, e.Row.Shown(), e.Row.Outcome(), e.Log)
	if e.Tail != "" {
		msg = e.Tail + "\n" + msg
	}
	return msg
}

// stepFailed is the error for a failed row: CommandFailed when the
// script exited, exit 3 when it did not (no exit code to pass through).
func stepFailed(row cmdrecord.Run, log, tail string) error {
	cf := &CommandFailed{row, log, tail}
	if row.Exit != nil {
		return cf
	}
	return herr.Env("%s", cf.Error())
}

// Drive runs every Drive command of the sandbox id's recipes, in stack
// order, each in its recipe's copy with the sandbox's variables, and stops
// at the first that fails. Each command is a row of sandbox.json, so a
// second drive adds rows. It refuses with 2 a sandbox none of whose recipes
// has a Drive command, and with 3 one whose setup failed. Every check runs
// before any command does.
func Drive(id string, out io.Writer) (*Record, error) {
	r, err := root()
	if err != nil {
		return nil, err
	}
	if !idRe.MatchString(id) {
		return nil, herr.BadArgs("bad sandbox id %q; use [a-z0-9-], at most 40 characters", id)
	}
	box := filepath.Join(r, id)
	meta := filepath.Join(box, "sandbox.json")
	if !isFile(meta) {
		return nil, herr.BadArgs("no sandbox %s (%s missing)", id, meta)
	}
	// A remove may have deleted the sandbox between the check above and the
	// lock, so check again.
	unlock, err := lockBox(box, id, "drive")
	if err != nil {
		return nil, err
	}
	defer unlock()
	if !isFile(meta) {
		return nil, herr.BadArgs("no sandbox %s (%s missing)", id, meta)
	}
	rec, err := readRecord(meta, id)
	if err != nil {
		return nil, err
	}
	for _, e := range rec.Recipes {
		if e.Setup == "failed" {
			return nil, herr.Env("sandbox %s: setup failed for %s; nothing driven; remove it", id, e.Ref)
		}
	}
	type drive struct {
		e  *Entry
		cp *recipe.Recipe
	}
	var drives []drive
	for i := range rec.Recipes {
		e := &rec.Recipes[i]
		cp, err := openCopy(r, *e)
		if err != nil {
			return nil, herr.New(herr.KindOf(err), "sandbox: %s at %s unreadable (%s); nothing driven", e.Ref, e.Path, herr.Text(err))
		}
		if cp != nil && len(cp.Drive) > 0 {
			drives = append(drives, drive{e, cp})
		}
	}
	if len(drives) == 0 {
		return nil, herr.BadArgs("sandbox %s has no Drive command; nothing to drive", id)
	}
	log, err := openLog(box)
	if err != nil {
		return nil, herr.Env("%v", err)
	}
	defer log.Close()
	write := func() error { return writeJSON(meta, rec) }
	ictx, stop := interrupted()
	defer stop()
	for _, d := range drives {
		for _, argv := range d.cp.Drive {
			save := rowSaver(&rec.Commands, write)
			ctx, cancel := context.WithTimeout(ictx, driveTimeout)
			row, err := runStep(ctx, log, d.e.Ref.String(), recipe.Drive, d.cp.Dir, stepCmd(ctx, log, d.cp.Dir, sandboxEnv(rec, *d.e), argv), driveTimeout, save, false)
			cancel()
			if err != nil {
				return rec, herr.Env("%v; record not written; the command did not run", err)
			}
			if err := save(row); err != nil {
				return rec, herr.Env("%v", err)
			}
			fmt.Fprintf(out, "%s drive: %s %s\n", d.e.Ref, row.Shown(), row.Outcome())
			if row.Failed() {
				return rec, stepFailed(row, filepath.Join(box, "recipe.log"), logTail(box, 30))
			}
		}
	}
	return rec, nil
}

// RunOptions names a recipe command run with no sandbox. Dir is the
// caller's working directory: the recipe lookup starts there; each command
// runs in an empty directory of its own (ADR 0003). Step is recipe.Evidence (every Evidence command, with
// Args the absolute OUT_DIR of hyp sandbox remove) or recipe.Standalone (the
// command Label, with Args the caller's ARGS). Record is the command
// record directory, relative to Dir unless absolute; the CLI reads the
// record back from its working directory, so it passes that as Dir. Record
// must not exist and must not be, or resolve inside, a recipes directory the
// lookup searches or, for Evidence, OUT_DIR.
type RunOptions struct {
	Dir, Ref      string
	Step          recipe.Step
	Label, Record string
	Args          []string
}

// argHashes hashes the files args name: a word, or the part after the
// first = of a KEY=VALUE word, that is an absolute path to a regular file,
// symlinks followed. A directory or a relative path is not hashed. A file
// that exists but cannot be read is an environment failure.
func argHashes(args []string) (map[string]string, error) {
	var out map[string]string
	for _, a := range args {
		for _, p := range argPaths(a) {
			if _, seen := out[p]; seen {
				continue
			}
			fi, err := os.Stat(p)
			if err != nil || !fi.Mode().IsRegular() {
				continue
			}
			sum, err := fileSum(p)
			if err != nil {
				return nil, herr.Env("cannot hash argument %s: %v", p, err)
			}
			if out == nil {
				out = map[string]string{}
			}
			out[p] = sum
		}
	}
	return out, nil
}

// argPaths lists the absolute paths word a may name: a itself, and the
// VALUE of a KEY=VALUE word.
func argPaths(a string) []string {
	var out []string
	if filepath.IsAbs(a) {
		out = append(out, a)
	}
	if _, v, ok := strings.Cut(a, "="); ok && filepath.IsAbs(v) {
		out = append(out, v)
	}
	return out
}

// Run runs o.Ref's Evidence commands, or its Standalone command o.Label,
// from a temporary copy of the bundle, and writes the command record to
// o.Record. It stops at the first command that fails. Every check runs
// before the record directory is made, so a refusal returns a nil record.
func Run(o RunOptions) (*cmdrecord.Record, error) {
	// Clean, so a/../b is checked as the b it names. dir is the one path
	// every check and write uses; o.Record stays as given for messages.
	o.Record = filepath.Clean(o.Record)
	dir := o.Record
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(o.Dir, dir)
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, herr.Env("%v", err)
	}
	if _, err := os.Lstat(dir); err == nil {
		return nil, herr.BadArgs("%s exists; choose a new path", o.Record)
	}
	if o.Step == recipe.Evidence && (len(o.Args) != 1 || !isFile(filepath.Join(o.Args[0], "sandbox.json"))) {
		return nil, herr.BadArgs("%s holds no sandbox.json; give the OUT_DIR of hyp sandbox remove, or, for files no sandbox wrote, run the check as a Standalone command (hyp recipe run --record DIR REF NAME PATH)", strings.Join(o.Args, " "))
	}
	rcp, err := recipe.Load(o.Dir, o.Ref)
	if err != nil {
		return nil, err
	}
	if err := recordOutside(o, dir); err != nil {
		return nil, err
	}
	if _, err := pick(rcp, o); err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp("", "hyp-run-")
	if err != nil {
		return nil, herr.Env("%v", err)
	}
	defer os.RemoveAll(tmp)
	cp, err := copyRecipe(o.Ref, rcp, tmp)
	if err != nil {
		return nil, err
	}
	argvs, err := pick(cp, o)
	if err != nil {
		return nil, err
	}
	// One read of the running hyp: the gate's cases and the commands run
	// the same copy, so known_answers.hyp_sha256 names the hyp the commands
	// ran.
	hyp, err := NewSandboxHyp()
	if err != nil {
		return nil, err
	}
	defer hyp.Remove()
	// The known answers run before the record directory exists, so a
	// refusal leaves none.
	var kalog bytes.Buffer
	ka, err := Gate(rcp, hyp, &kalog)
	if err != nil {
		return nil, herr.New(herr.KindOf(err), "%s%s", herr.Text(err), lastLines(kalog.String(), 30))
	}

	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return nil, herr.Env("%v", err)
	}
	// Mkdir, not MkdirAll: it fails on a directory made since the check
	// above, so the record directory removed on a failure below is ours.
	if err := os.Mkdir(dir, 0o755); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil, herr.BadArgs("%s exists; choose a new path", o.Record)
		}
		return nil, herr.Env("%v", err)
	}
	logPath, meta := filepath.Join(dir, "command.log"), filepath.Join(dir, "command.json")
	log, err := os.Create(logPath)
	if err != nil {
		os.RemoveAll(dir)
		return nil, herr.Env("%v", err)
	}
	defer log.Close()
	if kalog.Len() > 0 {
		if err := os.WriteFile(filepath.Join(dir, "known-answers.log"), kalog.Bytes(), 0o644); err != nil {
			os.RemoveAll(dir)
			return nil, herr.Env("%v", err)
		}
	}
	argSums, err := argHashes(o.Args)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	rec := &cmdrecord.Record{Format: cmdrecord.Format, Ref: rcp.Ref(), Name: rcp.Name, Version: rcp.Version, SHA256: rcp.Hash, Path: rcp.Path,
		Step: o.Step, Cwd: o.Dir, Created: now(), Commands: []cmdrecord.Run{}, KnownAnswers: ka, ArgSHA256: argSums}
	if err := writeJSON(meta, rec); err != nil {
		os.RemoveAll(dir)
		return nil, herr.Env("%v", err)
	}
	timeout := standaloneTimeout
	if o.Step == recipe.Evidence {
		timeout = evidenceTimeout
	}
	// A Standalone command's product is DIR/out and DIR/stdout; Evidence
	// command n's is DIR/out/n and DIR/stdout/n.
	outDst, stdoutDst := filepath.Join(dir, "out"), filepath.Join(dir, "stdout")
	if o.Step == recipe.Evidence {
		for _, d := range []string{outDst, stdoutDst} {
			if err := os.Mkdir(d, 0o755); err != nil {
				return rec, herr.Env("%v", err)
			}
		}
	}
	write := func() error { return writeJSON(meta, rec) }
	ictx, stop := interrupted()
	defer stop()
	for n, argv := range argvs {
		full := append(slices.Clone(argv), o.Args...)
		put := rowSaver(&rec.Commands, write)
		save := func(c cmdrecord.Run) error {
			c.Label = o.Label
			return put(c)
		}
		ctx, cancel := context.WithTimeout(ictx, timeout)
		out, stdout := outDst, stdoutDst
		if o.Step == recipe.Evidence {
			out, stdout = filepath.Join(outDst, strconv.Itoa(n)), filepath.Join(stdoutDst, strconv.Itoa(n))
		}
		row, product, err := runCheck(ctx, log, o.Ref, o.Step, cp.Dir, hyp.Dir, full, timeout, save, out, stdout)
		cancel()
		if err != nil {
			return rec, herr.Env("%v; record not written; the command did not run", err)
		}
		row.Label = o.Label
		if err := save(row); err != nil {
			return rec, herr.Env("%v", err)
		}
		// A failed command's exit code wins over a problem with its product.
		if row.Failed() {
			return rec, stepFailed(row, filepath.Join(o.Record, "command.log"), "")
		}
		if product != nil {
			return rec, product
		}
	}
	return rec, nil
}

// runCheck runs argv as a check, the way a known-answer case runs it, under
// a new root it removes after, and returns its row (see runStep). Its stderr
// goes to log. Once it ends, even when it failed, its stdout goes to log and
// to the file stdout, and its $HYP_OUT tree, read by the gate's rules
// (readTree), is written to out. product is nil, an exit 1 when the tree
// broke those rules (what it could read is kept), or an exit 3 when keeping
// the product failed. err means the command did not run.
func runCheck(ctx context.Context, log *os.File, ref string, step recipe.Step, bundle, bin string, argv []string, timeout time.Duration, save func(cmdrecord.Run) error, out, stdout string) (row cmdrecord.Run, product, err error) {
	tmp, err := os.MkdirTemp("", "hyp-check-")
	if err != nil {
		return row, nil, err
	}
	defer os.RemoveAll(tmp)
	if err := os.Mkdir(filepath.Join(tmp, "root"), 0o755); err != nil {
		return row, nil, err
	}
	root, err := checkRoot(filepath.Join(tmp, "root"))
	if err != nil {
		return row, nil, err
	}
	outF, err := os.Create(filepath.Join(tmp, "stdout"))
	if err != nil {
		return row, nil, err
	}
	defer outF.Close()
	if row, err = runStep(ctx, log, ref, step, bundle, checkCmd(ctx, root, bin, argv, outF, log), timeout, save, true); err != nil {
		return row, nil, err
	}
	data, rerr := os.ReadFile(outF.Name())
	if rerr == nil {
		_, rerr = log.Write(data)
	}
	if rerr == nil {
		rerr = os.WriteFile(stdout, data, 0o644)
	}
	tree, problem := readTree(filepath.Join(root, "out"), "out", "")
	if werr := errors.Join(rerr, writeTree(out, tree)); werr != nil {
		return row, herr.Env("%s %s: %s ran, but its product was not kept: %v", ref, step, row.Shown(), werr), nil
	}
	if problem != "" {
		return row, herr.Refused("%s %s: %s left a product hyp does not keep: %s; the rest is in %s", ref, step, row.Shown(), problem, out), nil
	}
	return row, nil, nil
}

// recordOutside refuses with 2 a record at abs that is, or is inside, a
// recipes directory the lookup searches, since a record there changes the
// hash of the bundle it lands in and every pin to it would drift, or, for
// Evidence, inside OUT_DIR, since the check reads it. A path it cannot
// resolve is 3.
func recordOutside(o RunOptions, abs string) error {
	// where names root in the refusal.
	refuse := func(root, where string) error {
		in, err := covers(root, abs)
		if err != nil {
			return herr.Env("cannot resolve --record %s: %v; nothing ran", o.Record, err)
		}
		if in {
			return herr.BadArgs("--record %s is inside %s; choose a path outside it", o.Record, where)
		}
		return nil
	}
	if o.Step == recipe.Evidence {
		if err := refuse(o.Args[0], "OUT_DIR "+o.Args[0]+", which the check reads"); err != nil {
			return err
		}
	}
	// A recipes directory that does not exist yet counts too, so a record
	// cannot start a library where the lookup would find it.
	for _, d := range recipe.Dirs(o.Dir) {
		if err := refuse(d, "the recipe library "+d+", where it would change a recipe's hash"); err != nil {
			return err
		}
	}
	return nil
}

// pick returns the commands o names in rcp: every Evidence command, or the
// one Standalone command labelled o.Label.
func pick(rcp *recipe.Recipe, o RunOptions) ([][]string, error) {
	if o.Step == recipe.Evidence {
		if len(rcp.Evidence) == 0 {
			return nil, herr.BadArgs("recipe %s has no Evidence command", o.Ref)
		}
		return rcp.Evidence, nil
	}
	var names []string
	for _, n := range rcp.Standalone {
		if n.Name == o.Label {
			return [][]string{n.Argv}, nil
		}
		names = append(names, n.Name)
	}
	has := "none"
	if len(names) > 0 {
		has = strings.Join(names, ", ")
	}
	return nil, herr.BadArgs("recipe %s has no Standalone command %q; it has: %s", o.Ref, o.Label, has)
}
