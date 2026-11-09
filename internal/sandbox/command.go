package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
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
// non-zero. hyp exits with Row's exit code unchanged (sandbox spec §1.3).
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
// has a Drive command, with 3 one that is not ready (docs/sandbox-spec.md §2),
// and with 3 one whose unfinished row's group still lives. Every check runs
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
	if !isDir(box) {
		return nil, herr.BadArgs("no sandbox %s (%s missing)", id, box)
	}
	// The record is read under the lock, so a create still making it reads as
	// busy. lockBox reports no sandbox when a remove deleted it since the check
	// above.
	unlock, err := lockBox(box, id, "drive")
	if err != nil {
		return nil, err
	}
	defer unlock()
	rec, err := readBox(box, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, herr.Env("sandbox %s has no record: hyp stopped while making it; nothing driven; run hyp sandbox remove %s --discard", id, id)
	}
	switch deriveState(rec, false) {
	case stateInterrupted:
		return nil, herr.Env("sandbox %s: hyp stopped before setup finished; nothing driven; run hyp sandbox remove %s", id, id)
	case stateSetupFailed:
		return nil, herr.Env("sandbox %s: setup failed; nothing driven; run hyp sandbox remove %s", id, id)
	case stateRemoving:
		return nil, herr.Env("sandbox %s: a remove already started and did not finish; nothing driven; run hyp sandbox remove %s", id, id)
	}
	if live := liveGroups(unfinished(rec.Commands)); len(live) > 0 {
		return nil, herr.Env("sandbox %s: process group %d of a command whose hyp stopped still runs; nothing driven; stop it and drive again, or run hyp sandbox remove %s, which stops it", id, live[0], id)
	}
	type drive struct {
		e  *Entry
		cp *recipe.Recipe
	}
	var drives []drive
	for i := range rec.Recipes {
		e := &rec.Recipes[i]
		cp, err := openCopy(r, rec.CopyDir(*e), *e)
		if err != nil {
			return nil, herr.New(herr.KindOf(err), "sandbox: %s at %s unreadable (%s); nothing driven", e.Ref, rec.CopyDir(*e), herr.Text(err))
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
			row, err := runStep(ctx, log, d.e.Ref.String(), recipe.Drive, d.cp.Dir, stepCmd(ctx, log, d.cp.Dir, sandboxEnv(rec, *d.e), argv), nil, driveTimeout, save, false)
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
	// Warn receives one line per directory argument outside the repository.
	Warn io.Writer
}

// maxKeptInput caps a kept input. It is far below the changes/ cap
// (maxCopy) because kept inputs sit beside every record in git, while
// changes/ is one sandbox's output.
const maxKeptInput = 1 << 20

// pathWord reports whether s is written as a path: absolute, or relative
// with a /. A bare word is never a path, even when a file of that name
// exists, so a plain word cannot silently become one.
func pathWord(s string) bool { return filepath.IsAbs(s) || strings.Contains(s, "/") }

// repoOf is the repository a run's paths are anchored to: the top of the git
// work tree that holds cwd, or cwd outside one.
func repoOf(cwd string) string {
	out, err := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel").Output()
	if top := strings.TrimSpace(string(out)); err == nil && top != "" {
		return top
	}
	return cwd
}

// anchored is p as a record writes it: <repo>/REL inside repo, as given
// outside it.
func anchored(repo, p, given string) (string, bool) {
	if a, ok := anchor(repo, filepath.Base(repo), p); ok {
		return a, true
	}
	return given, false
}

// stageArgs reads each ARGS word (counted from 1) as a file argument, a
// directory argument or a plain word. A word, or the VALUE of a KEY=VALUE
// word, written as a path (pathWord) that names a regular file, symlinks
// followed, is a file argument: hyp copies it under tmp, which the script
// gets, hashes that copy, and keeps a second copy at dir/inputs/<n>/<base>
// unless it is over maxKeptInput; the row shows inputs/<n>/<base>. One that
// names a directory is passed as an absolute path and shown anchored to
// repo, with a warning when it is outside. A relative path is read against
// cwd. Every other word is passed and shown as given. A file argument that
// cannot be read is an environment failure.
func stageArgs(args []string, cwd, repo, dir, tmp string, warn io.Writer) (run, shown []string, inputs []cmdrecord.Input, err error) {
	for i, a := range args {
		n := i + 1
		prefix, val := "", a
		fi, abs := statPath(a, cwd)
		if fi == nil {
			if k, v, ok := strings.Cut(a, "="); ok {
				if fi, abs = statPath(v, cwd); fi != nil {
					prefix, val = k+"=", v
				}
			}
		}
		switch {
		case fi == nil:
			run, shown = append(run, a), append(shown, a)
		case fi.IsDir():
			show, in := anchored(repo, abs, val)
			if !in && warn != nil {
				fmt.Fprintf(warn, "warn: argument %d names a directory outside the repository; a replay cannot reach it\n", n)
			}
			run, shown = append(run, prefix+abs), append(shown, prefix+show)
		default:
			rel := filepath.ToSlash(filepath.Join("inputs", strconv.Itoa(n), filepath.Base(abs)))
			in, err := keepInput(abs, fi.Mode().Perm(), filepath.Join(tmp, filepath.FromSlash(rel)), filepath.Join(dir, filepath.FromSlash(rel)))
			if err != nil {
				return nil, nil, nil, herr.Env("cannot read argument %d (%s): %v", n, val, err)
			}
			in.Arg, in.Path = n, rel
			inputs = append(inputs, in)
			run, shown = append(run, prefix+filepath.Join(tmp, filepath.FromSlash(rel))), append(shown, prefix+rel)
		}
	}
	return run, shown, inputs, nil
}

// statPath stats s, read against cwd when relative, when it is written as a
// path and names a regular file or a directory; nil otherwise.
func statPath(s, cwd string) (fs.FileInfo, string) {
	if !pathWord(s) {
		return nil, ""
	}
	p := s
	if !filepath.IsAbs(p) {
		p = filepath.Join(cwd, p)
	}
	fi, err := os.Stat(p)
	if err != nil || !(fi.Mode().IsRegular() || fi.IsDir()) {
		return nil, ""
	}
	return fi, filepath.Clean(p)
}

// keepInput copies src to pass, the copy the script gets, hashes it, and,
// when it is at most maxKeptInput, copies it again to keep, mode kept, so a
// script that writes to its argument cannot change the kept bytes.
func keepInput(src string, mode fs.FileMode, pass, keep string) (cmdrecord.Input, error) {
	var in cmdrecord.Input
	if err := os.MkdirAll(filepath.Dir(pass), 0o755); err != nil {
		return in, err
	}
	if err := copyFile(src, pass, mode); err != nil {
		return in, err
	}
	fi, err := os.Stat(pass)
	if err != nil {
		return in, err
	}
	if in.SHA256, err = fileSum(pass); err != nil {
		return in, err
	}
	in.Size = fi.Size()
	if in.Size > maxKeptInput {
		return in, nil
	}
	if err := os.MkdirAll(filepath.Dir(keep), 0o755); err != nil {
		return in, err
	}
	if err := copyFile(pass, keep, mode); err != nil {
		return in, err
	}
	in.Kept = true
	return in, nil
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
	cp, tmp, removeCopy, err := newBundleCopy(o.Ref, rcp)
	if err != nil {
		return nil, err
	}
	defer removeCopy()
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
	repo := repoOf(o.Dir)
	runArgs, shownArgs, inputs, err := stageArgs(o.Args, o.Dir, repo, dir, filepath.Join(tmp, "args"), o.Warn)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	path, _ := anchored(repo, rcp.Path, rcp.Path)
	cwd, _ := anchored(repo, o.Dir, o.Dir)
	rec := &cmdrecord.Record{Format: cmdrecord.Format, Ref: rcp.Ref(), Name: rcp.Name, Version: rcp.Version, SHA256: rcp.Hash, Path: path,
		Step: o.Step, Cwd: cwd, Created: now(), Commands: []cmdrecord.Run{}, KnownAnswers: ka, Inputs: inputs}
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
		full := append(slices.Clone(argv), runArgs...)
		shown := append(rowArgs(cp.Dir, argv), shownArgs...)
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
		row, product, err := runCheck(ctx, log, o.Ref, o.Step, cp.Dir, hyp.Dir, full, shown, timeout, save, out, stdout)
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
			return rec, stepFailed(row, filepath.Join(o.Record, "command.log"), fileTail(logPath, 30))
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
func runCheck(ctx context.Context, log *os.File, ref string, step recipe.Step, bundle, bin string, argv, shown []string, timeout time.Duration, save func(cmdrecord.Run) error, out, stdout string) (row cmdrecord.Run, product, err error) {
	root, removeRoot, err := newCheckRoot()
	if err != nil {
		return row, nil, err
	}
	defer removeRoot()
	// Outside root, so the command cannot see or change it, as in a case.
	outs, err := os.MkdirTemp("", "hyp-check-out-")
	if err != nil {
		return row, nil, err
	}
	defer os.RemoveAll(outs)
	outF, err := os.Create(filepath.Join(outs, "stdout"))
	if err != nil {
		return row, nil, err
	}
	defer outF.Close()
	if row, err = runStep(ctx, log, ref, step, bundle, checkCmd(ctx, root, bin, argv, outF, log), shown, timeout, save, true); err != nil {
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
