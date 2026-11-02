// Package sandbox makes and removes a disposable clone of a host repository for one Experiment.
//
// Create makes ~/.hyp-sandbox/ID with:
//
//	wt/           a clone of the host detached at the base commit, origin removed;
//	              with a store cut, wt/.tree is that store as it was at the cut event
//	tmp/          scratch, holding an empty tmp/<name>/ for each recipe
//	recipes/      the sandbox's copy of each recipe (recipes/<name>/), unless it is none;
//	              outside wt/, so the agent working there does not find it; a copy is opened by
//	              its ref and checked against the record's sha256 (docs/adr/0005-a-bundle-is-its-ref-and-hash.md)
//	bin/hyp       a copy of the pinned hyp binary
//	inputs/       with --inputs, the sandbox's copy of the inputs file (HYP_SANDBOX_INPUTS)
//	manifest.json every path in wt/ when create finished, the baseline of changes.json
//	recipe.log    the output of every recipe command, each after a "== REF phase: command" line
//	sandbox.json  the Record, written only by hyp
//	sandbox.lock  held by the one hyp command changing the sandbox
//
// With Store and Cut, create writes wt/.tree from the store as replay sees it
// as of the cut event (store.CutAt), so hyp status in wt/ equals hyp status
// --as-of CUT in the store, before any recipe runs.
//
// A create that fails midway deletes its partial sandbox. A clone, not
// `git worktree add`: a worktree and its branches are host refs, and a clone
// never writes to the host's objects.
//
// The recipes stack in the order given, and none stands alone. Once the
// record is written, create runs each recipe's Launch and then Doctor
// commands in its copy, under that recipe's own deadline, before the next
// recipe's Launch, and records setup "passed" (or "none" when there is
// nothing to run). If one fails or times out, it runs the Cleanup of that
// recipe and of every recipe before it, last first, and deletes the sandbox
// only when no process still runs in it; a sandbox it keeps records setup
// "failed" for that recipe.
//
// Remove runs every recipe's Cleanup commands, last recipe first, every
// command even after one fails, each under its own deadline. It then copies
// wt/.tree (without index.sqlite*, which older hyp binaries write; ADR 0006),
// recipe.log, each kept file and the record (with removed), the inputs copy, and changes.json with changes/ (every path
// in wt/ that changed since create finished) to OUT_DIR, and deletes the
// sandbox. With Discard it deletes without writing a
// record. A record written by an older hyp may hold a single recipe object
// and fields this one no longer writes; remove reads it, and the copy lists
// the recipe under recipes and keeps the other fields.
//
// Drive runs every recipe's Drive commands while the sandbox exists, in stack
// order, each in its recipe's copy with the same variables as Launch, and
// stops at the first that fails. Run runs a recipe's Evidence commands, or
// one Standalone command, with no sandbox: from a temporary copy of the
// bundle that hashes the same, in the caller's working directory, with no
// HYP_SANDBOX_ variable. It writes a command record, command.json and
// command.log, to a new directory. Every command hyp runs, in the sandbox or
// not, is a row of its record with its arguments and exit code, written
// before the command starts; a row with no ended is one hyp stopped during.
// Records are written through a temporary file and a rename, so a reader
// sees the old record or the new one, never part of one.
//
// Every command hyp starts leads its own process group, which its row
// records as pgid. Remove stops, after Cleanup, the groups of Launch,
// Doctor, Drive and Cleanup commands that still have members (SIGTERM, then
// SIGKILL after stopWait). A timeout or Ctrl-C kills the running command's
// group; an Evidence or Standalone group ends with its command. The sandbox
// owns no network resource: a recipe that needs one sets it up in its Launch
// and stops it in its Cleanup, and a daemon that leaves its group (setsid)
// is stopped only by its Cleanup. Its one leftover check is the
// working directory of every other process (lsof -d cwd): create after a failed
// setup, and remove, delete nothing while a process runs in the sandbox,
// whatever it listens on. Limits: an unprivileged lsof on macOS may not list
// other users' processes, and a process that only holds a file open in the
// sandbox, with its working directory elsewhere, is not seen.
//
// Exit codes: 2 the arguments are wrong (bad id, id taken, host not a
// repository root, base not a commit, a hyp that does not answer version,
// OUT_DIR exists, a bad keep path, no such sandbox, a missing --recipe, a
// recipe name given twice, none stacked with a recipe, a recipe that does
// not resolve, --store without --cut or the reverse, a store or cut event
// that does not exist, a base that already holds .tree, a sandbox with no
// Drive command, a record directory that exists, an OUT_DIR without
// sandbox.json, an unknown Standalone name, a recipe with no Evidence
// command, an --inputs that is not a regular file); 3 the environment is wrong or I/O failed (a
// process still running in the sandbox, or one hyp could not check for,
// a sandbox another hyp command is changing,
// missing git or lsof, a failed create step, a failed or timed-out Launch or
// Doctor, a sandbox whose setup failed, a Drive, Evidence or Standalone
// command that timed out, was interrupted or was killed, a corrupt record, a store line that
// does not parse). A Drive, Evidence or Standalone script's own non-zero exit
// is returned as *CommandFailed. Every check runs before anything is
// created or deleted.
package sandbox

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/store"
)

// Record is sandbox.json. Recipes lists the sandbox's recipes in stack order;
// --recipe none is the one entry {"ref": "none"}. Store and Cut name the
// store wt/.tree was cut from and the cut event, when create made one.
// Commands lists every command hyp ran in the sandbox, in order. Extra holds every other
// field of a record read from disk (an older hyp's port, or a key a recipe
// script added), so it is written back as read.
type Record struct {
	Format     int                        `json:"format"`
	ID         string                     `json:"id"`
	Host       string                     `json:"host"`
	Base       string                     `json:"base"`
	WT         string                     `json:"wt"`
	Tmp        string                     `json:"tmp"`
	Bin        string                     `json:"bin"`
	HypSHA256  string                     `json:"hyp_sha256"`
	HypVersion string                     `json:"hyp_version"`
	Recipes    []Entry                    `json:"recipes"`
	Inputs     *Inputs                    `json:"inputs,omitempty"`
	Commands   []cmdrecord.Run            `json:"commands,omitempty"`
	Store      string                     `json:"store,omitempty"`
	Cut        string                     `json:"cut,omitempty"`
	Created    string                     `json:"created"`
	Removed    string                     `json:"removed,omitempty"`
	Extra      map[string]json.RawMessage `json:"-"`
}

// Entry is one recipe of a Record. Setup is passed, failed or none (nothing
// to run), and absent for none and in an older record that did not say.
// Cleanup is ok, or "failed: " and each command that failed; it is absent
// until remove, or a failed setup, ran Cleanup commands.
type Entry struct {
	Ref     event.Ref `json:"ref"`
	Name    string    `json:"name,omitempty"`
	Version string    `json:"version,omitempty"`
	SHA256  string    `json:"sha256,omitempty"`
	Path    string    `json:"path,omitempty"`
	Setup   string    `json:"setup,omitempty"`
	Cleanup string    `json:"cleanup,omitempty"`
	// KnownAnswers is the outcome of the recipe's known answers, run before
	// its commands; nil until the gate runs them.
	KnownAnswers *cmdrecord.KnownAnswers `json:"known_answers,omitempty"`
}

// Inputs is a run's inputs file: the path given to --inputs, the sandbox's
// copy every command gets as HYP_SANDBOX_INPUTS, and the copy's sha256.
type Inputs struct {
	Source string `json:"source"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// plainRecord is Record without its JSON methods.
type plainRecord Record

// MarshalJSON writes the fields and Extra as one object, keys sorted.
func (r Record) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(plainRecord(r))
	if err != nil {
		return nil, err
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	for k, v := range r.Extra {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
	return json.Marshal(m)
}

// UnmarshalJSON reads the fields, puts every other key in Extra, and reads
// an older record's single recipe object as the one entry of Recipes. When
// both are present, recipes wins; the recipe object is never written back.
func (r *Record) UnmarshalJSON(data []byte) error {
	var p plainRecord
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	t := reflect.TypeFor[plainRecord]()
	for i := range t.NumField() {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		delete(m, name)
	}
	if old, ok := m["recipe"]; ok {
		if p.Recipes == nil {
			var e Entry
			if err := json.Unmarshal(old, &e); err != nil {
				return err
			}
			p.Recipes = []Entry{e}
		}
		delete(m, "recipe")
	}
	if p.Recipes == nil {
		p.Recipes = []Entry{}
	}
	if len(m) > 0 {
		p.Extra = m
	}
	*r = Record(p)
	return nil
}

// CreateOptions names the sandbox, its host repository and base commit, and
// the hyp binary it pins (empty: the running hyp).
type CreateOptions struct {
	ID, Host, Base, Hyp string
	// Refs is the recipes in stack order, each NAME@VERSION, or the one
	// ref none, and is required; Recipes is the directory that holds recipes/.
	Refs    []string
	Recipes string
	// Store and Cut, both or neither: write wt/.tree from the store at
	// Store as it was at the event Cut.
	Store, Cut string
	// Inputs is a file of this run's inputs, or "": the sandbox keeps a
	// copy and passes it to every command as HYP_SANDBOX_INPUTS.
	Inputs string
}

// setupTimeout bounds one recipe's Launch and Doctor together.
var setupTimeout = 10 * time.Minute

// cleanupTimeout bounds one run of Cleanup.
var cleanupTimeout = 2 * time.Minute

// RemoveOptions names the sandbox and where its record goes: OutDir, or
// nowhere with Discard. Warn receives one line per missing Keep file.
type RemoveOptions struct {
	ID, OutDir string
	Keep       []string // paths relative to the sandbox directory
	Discard    bool
	Warn       io.Writer
}

var idRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)

// root is resolved per call, so a changed HOME takes effect.
func root() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", herr.Env("sandbox: %v", err)
	}
	return filepath.Join(home, ".hyp-sandbox"), nil
}

func needTools(names ...string) error {
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			return herr.Env("sandbox: %s not found on PATH; nothing changed", name)
		}
	}
	return nil
}

// leftovers lists each process other than this hyp whose working directory
// is box or under it, one "PID COMMAND CWD" line each, whatever it listens
// on. lsof -d cwd lists every process's working directory, symlink-resolved,
// and exits 1 when it could not read some; its output still counts. It
// always lists itself, so no output means it failed. It runs from /, so it
// never lists itself as a leftover when hyp runs inside the sandbox.
func leftovers(box string) (string, error) {
	rbox, err := filepath.EvalSymlinks(box)
	if err != nil {
		return "", herr.Env("sandbox: %v", err)
	}
	cmd := exec.Command("lsof", "-nP", "-d", "cwd", "-F", "pcn")
	cmd.Dir = "/"
	out, err := cmd.Output()
	if len(out) == 0 {
		return "", herr.Env("lsof -d cwd listed no processes (%v); cannot tell what still runs in %s", err, box)
	}
	self := strconv.Itoa(os.Getpid())
	var held []string
	var pid, command string
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		switch line[0] {
		case 'p':
			pid, command = line[1:], ""
		case 'c':
			command = line[1:]
		case 'n':
			cwd := line[1:]
			if pid != self && (cwd == rbox || strings.HasPrefix(cwd, rbox+string(filepath.Separator))) {
				held = append(held, pid+" "+command+" "+cwd)
			}
		}
	}
	return strings.Join(held, "\n"), nil
}

// inside reports whether path is strictly under root once both are
// symlink-resolved (macOS /var is /private/var). path is resolved as
// resolve does, so a hand-deleted tmp/ does not block remove and a directory
// MkdirAll would make is checked where it would land.
func inside(root, path string) (bool, error) {
	r, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false, err
	}
	p, err := resolve(path)
	if err != nil {
		return false, err
	}
	return under(r, p), nil
}

// resolve symlink-resolves path. A missing absolute path is resolved through
// its deepest existing parent, with the missing names joined back.
func resolve(path string) (string, error) {
	base, missing, err := split(path)
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{base}, missing...)...), nil
}

// split returns the symlink-resolved deepest existing parent of path (path
// itself when it exists) and the names below it that do not exist. Only an
// absolute path is walked up.
func split(path string) (string, []string, error) {
	p, err := filepath.EvalSymlinks(path)
	var missing []string
	for dir := path; errors.Is(err, fs.ErrNotExist) && filepath.IsAbs(path) && dir != filepath.Dir(dir); {
		missing = append([]string{filepath.Base(dir)}, missing...)
		dir = filepath.Dir(dir)
		p, err = filepath.EvalSymlinks(dir)
	}
	return p, missing, err
}

// covers reports whether path is root or under it; either may be missing.
// Existing directories are compared by identity, so neither a symlink nor
// another spelling on a case-insensitive disk (macOS) gets past it, and
// missing names ignoring case, which refuses more on a case-sensitive disk,
// never less.
func covers(root, path string) (bool, error) {
	rb, rt, err := split(root)
	if err != nil {
		return false, err
	}
	ri, err := os.Stat(rb)
	if err != nil {
		return false, err
	}
	pb, pt, err := split(path)
	if err != nil {
		return false, err
	}
	for cur, below := pb, pt; ; cur, below = filepath.Dir(cur), append([]string{filepath.Base(cur)}, below...) {
		if ci, err := os.Stat(cur); err == nil && os.SameFile(ci, ri) {
			return len(below) >= len(rt) && slices.EqualFunc(below[:len(rt)], rt, strings.EqualFold), nil
		}
		if cur == filepath.Dir(cur) {
			return false, nil
		}
	}
}

// under reports whether the resolved path p is strictly under the resolved r.
func under(r, p string) bool { return strings.HasPrefix(p, r+string(filepath.Separator)) }

// git runs one git command and returns its trimmed output.
func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return "", herr.Env("git %s failed: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func isFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsRegular()
}

func copyFile(src, dst string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// rename is os.Rename; a test swaps it to fail between the temporary file and the rename.
var rename = os.Rename

// writeJSON writes v to path through a temporary file in the same directory,
// renamed over path, so a reader sees the old file or the new one, never
// part of one. On a failure the old file is unchanged and the temporary
// file is removed.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(append(data, '\n'))
	if err == nil {
		err = f.Chmod(0o644) // CreateTemp makes 0600
	}
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
	}
	return err
}

// flock is syscall.Flock; a test swaps it to act between the open and the lock.
var flock = syscall.Flock

// lockBox takes the sandbox's lock, an flock on box/sandbox.lock that the
// kernel drops when this hyp exits, however it exits. It writes "pid PID
// hyp sandbox STEP since TIME" into the file for a refused command to name.
// A sandbox another hyp command holds is refused with 3; one whose directory
// is gone is "no sandbox" (2). The caller defers unlock.
//
// A remove may delete the sandbox between the open and the lock, and a
// create may make it again, so the lock taken may be on a deleted file.
// lockBox then opens the path again, and holds a lock only on the file the
// path names.
func lockBox(box, id, step string) (unlock func(), err error) {
	path := filepath.Join(box, "sandbox.lock")
	for {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
		if errors.Is(err, fs.ErrNotExist) {
			return nil, herr.BadArgs("no sandbox %s (%s missing)", id, box)
		}
		if err != nil {
			return nil, herr.Env("%v; nothing changed", err)
		}
		if err := flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			defer f.Close()
			if !errors.Is(err, syscall.EWOULDBLOCK) {
				return nil, herr.Env("lock %s: %v; nothing changed", path, err)
			}
			// Empty when the holder has not written its line yet.
			holder := "another hyp command"
			if data, _ := io.ReadAll(f); len(strings.TrimSpace(string(data))) > 0 {
				holder = strings.TrimSpace(string(data))
			}
			return nil, herr.Env("sandbox %s is busy: %s; wait for it to end or stop it; nothing changed", id, holder)
		}
		held, err1 := f.Stat()
		named, err2 := os.Stat(path)
		if err1 != nil || err2 != nil || !os.SameFile(held, named) {
			f.Close()
			continue
		}
		// The lock holds without the line, so a failed write is not fatal.
		if f.Truncate(0) == nil {
			f.WriteAt(fmt.Appendf(nil, "pid %d hyp sandbox %s since %s\n", os.Getpid(), step, now()), 0)
		}
		// Closing the file releases its flock.
		return func() { f.Close() }, nil
	}
}

// Create makes the sandbox o.ID and returns its record. Every check runs
// before the sandbox directory is made.
func Create(o CreateOptions) (*Record, error) {
	if err := needTools("git", "lsof"); err != nil {
		return nil, err
	}
	if !idRe.MatchString(o.ID) {
		return nil, herr.BadArgs("bad sandbox id %q; use [a-z0-9-], at most 40 characters", o.ID)
	}
	if len(o.Refs) == 0 {
		return nil, herr.BadArgs("missing --recipe (NAME@VERSION or none)")
	}
	// Each recipe has one copy and one tmp directory, both named by its name.
	names := map[string]bool{}
	for _, ref := range o.Refs {
		// A ref that does not parse is refused when it loads; until then its
		// whole text is its name.
		name := ref
		if r, err := event.ParseRef(ref); err == nil {
			name = r.Name
		}
		if names[name] {
			return nil, herr.BadArgs("--recipe %s: recipe %s is already in the stack; give each name once", ref, name)
		}
		names[name] = true
	}
	if names["none"] && len(o.Refs) > 1 {
		return nil, herr.BadArgs("--recipe none sets nothing up and is given alone, not with another recipe")
	}
	if (o.Store == "") != (o.Cut == "") {
		return nil, herr.BadArgs("--store and --cut go together: the store to cut and the event to cut it at")
	}
	r, err := root()
	if err != nil {
		return nil, err
	}
	box := filepath.Join(r, o.ID)
	if _, err := os.Lstat(box); err == nil {
		return nil, herr.BadArgs("%s exists; remove it or choose a new id", box)
	}
	if !isRepoRoot(o.Host) {
		return nil, herr.BadArgs("%s is not the root of a git work tree", o.Host)
	}
	if _, err := git("-C", o.Host, "cat-file", "-e", o.Base+"^{commit}"); err != nil {
		return nil, herr.BadArgs("%s is not a commit in %s", o.Base, o.Host)
	}
	if o.Inputs != "" {
		if fi, err := os.Stat(o.Inputs); err != nil || !fi.Mode().IsRegular() {
			return nil, herr.BadArgs("--inputs %s is not a regular file", o.Inputs)
		}
	}
	var cut *store.Cut
	if o.Store != "" {
		if _, err := git("-C", o.Host, "cat-file", "-e", o.Base+":.tree"); err == nil {
			return nil, herr.BadArgs("%s already holds .tree at %s; a store cut needs a base without one", o.Host, o.Base)
		}
		if cut, err = store.CutAt(o.Store, o.Cut); err != nil {
			if errors.Is(err, store.ErrNoStore) {
				return nil, herr.BadArgs("--store %s holds no .tree/events", o.Store)
			}
			if errors.Is(err, store.ErrNoEvent) {
				return nil, herr.BadArgs("--cut: %v", err)
			}
			return nil, herr.Env("--cut: %v; nothing created", err)
		}
	}
	var rcps []*recipe.Recipe
	if o.Refs[0] != "none" {
		if o.Recipes == "" {
			return nil, herr.BadArgs("--recipe %s needs --recipes DIR (the directory that holds recipes/)", o.Refs[0])
		}
		for _, ref := range o.Refs {
			rcp, err := recipe.Load(o.Recipes, ref)
			if err != nil {
				return nil, err
			}
			rcps = append(rcps, rcp)
		}
	}
	if o.Hyp == "" {
		if o.Hyp, err = os.Executable(); err != nil {
			return nil, herr.Env("sandbox: %v", err)
		}
	}
	// Absolute, so version is asked of the same file that is copied.
	if o.Hyp, err = filepath.Abs(o.Hyp); err != nil {
		return nil, herr.BadArgs("%s is not a hyp binary (%v)", o.Hyp, err)
	}
	version, err := hypVersion(o.Hyp)
	if err != nil {
		return nil, herr.BadArgs("%s is not a hyp binary (%v)", o.Hyp, err)
	}
	// Each recipe's known answers, in stack order, before anything is made.
	var kalog bytes.Buffer
	kas := make([]*cmdrecord.KnownAnswers, len(rcps))
	if len(rcps) > 0 {
		self, err := NewSandboxHyp()
		if err != nil {
			return nil, herr.New(herr.KindOf(err), "%s; nothing created", herr.Text(err))
		}
		defer self.Remove()
		for i, rcp := range rcps {
			if kas[i], err = Gate(rcp, self, &kalog); err != nil {
				return nil, herr.New(herr.KindOf(err), "%s; nothing created%s", herr.Text(err), lastLines(kalog.String(), 30))
			}
		}
	}

	if err := os.MkdirAll(r, 0o755); err != nil {
		return nil, herr.Env("%v; nothing created", err)
	}
	if err := os.Mkdir(box, 0o755); err != nil {
		return nil, herr.Env("%v; nothing created", err)
	}
	// Held through Launch and Doctor, so no drive or remove runs during them.
	unlock, err := lockBox(box, o.ID, "create")
	if err != nil {
		os.RemoveAll(box)
		return nil, err
	}
	defer unlock()
	rec, cps, err := build(o, box, version, rcps, kas, cut)
	if err != nil {
		os.RemoveAll(box)
		return nil, herr.New(herr.KindOf(err), "%s; partial sandbox removed", herr.Text(err))
	}
	if kalog.Len() > 0 {
		if err := os.WriteFile(filepath.Join(box, "known-answers.log"), kalog.Bytes(), 0o644); err != nil {
			os.RemoveAll(box)
			return nil, herr.Env("%v; partial sandbox removed", err)
		}
	}
	ictx, stop := interrupted()
	defer stop()
	ran := false
	for i, cp := range cps {
		e := &rec.Recipes[i]
		if len(cp.Launch)+len(cp.Doctor) == 0 {
			continue
		}
		ran = true
		if err := runSetup(ictx, box, rec, *e, cp); err != nil {
			// A second Ctrl-C during Cleanup stops hyp.
			stop()
			e.Setup = "failed"
			return nil, setupFailed(o.ID, box, rec, cps[:i+1], err)
		}
		e.Setup = "passed"
	}
	if ran {
		if err := writeJSON(filepath.Join(box, "sandbox.json"), rec); err != nil {
			return nil, herr.Env("%v; sandbox kept", err)
		}
	}
	// The baseline remove diffs wt/ against: what create made, Launch and
	// Doctor included.
	if err := writeManifest(box, rec.WT); err != nil {
		return nil, herr.Env("%v; sandbox kept", err)
	}
	return rec, nil
}

// setupFailed runs the Cleanup of each recipe in launched, the last first,
// after err, the failed setup of the last one. It deletes the sandbox when no
// process still runs in it, and otherwise keeps it with its record.
func setupFailed(id, box string, rec *Record, launched []*recipe.Recipe, err error) error {
	// A failed Cleanup is named in the message; the leftover check below decides.
	cleanupNote, ran := "", "no Cleanup to run"
	for i := len(launched) - 1; i >= 0; i-- {
		cp := launched[i]
		if len(cp.Cleanup) == 0 {
			continue
		}
		ran = "Cleanup run"
		e := &rec.Recipes[i]
		var report string
		e.Cleanup, report = runCleanup(box, rec, *e, cp)
		if report != "" {
			cleanupNote += "\nsandbox: " + report
		}
	}
	var stopped strings.Builder
	stopGroups(rec.Commands, &stopped)
	if stopped.Len() > 0 {
		cleanupNote += "\n" + strings.TrimSuffix(stopped.String(), "\n")
	}
	tail := logTail(box, 30)
	held, lerr := leftovers(box)
	if lerr == nil && held == "" {
		os.RemoveAll(box)
		return herr.Env("%v\n%s%s\nsandbox: recipe setup failed; %s; sandbox removed", err, tail, cleanupNote, ran)
	}
	werr := writeJSON(filepath.Join(box, "sandbox.json"), rec)
	recNote := ""
	if werr != nil {
		recNote = fmt.Sprintf("\nsandbox: %v; the record does not say setup failed", werr)
	}
	next := fmt.Sprintf("then run hyp sandbox remove %s OUT_DIR (keeps recipe.log) or hyp sandbox remove %s --discard", id, id)
	if lerr != nil {
		// Unknown is treated as running: delete only a sandbox nothing runs in.
		return herr.Env("%v\n%s%s%s\n%v\nsandbox: recipe setup failed and hyp could not check whether processes still run in %s; sandbox kept: make sure none do, %s", err, tail, cleanupNote, recNote, lerr, box, next)
	}
	return herr.Env("%v\n%s%s%s\n%s\nsandbox: recipe setup failed and processes still run in %s; sandbox kept: stop them, %s", err, tail, cleanupNote, recNote, held, box, next)
}

// runSetup runs the Launch, then Doctor, commands of e's copy cp under one
// setupTimeout deadline, and stops at the first command that fails, so its
// error names that command only.
func runSetup(ictx context.Context, box string, rec *Record, e Entry, cp *recipe.Recipe) error {
	ctx, cancel := context.WithTimeout(ictx, setupTimeout)
	defer cancel()
	log, err := openLog(box)
	if err != nil {
		return err
	}
	defer log.Close()
	write := func() error { return writeJSON(filepath.Join(box, "sandbox.json"), rec) }
	for _, phase := range []struct {
		name recipe.Step
		cmds [][]string
	}{{recipe.Launch, cp.Launch}, {recipe.Doctor, cp.Doctor}} {
		for _, argv := range phase.cmds {
			save := rowSaver(&rec.Commands, write)
			row, err := runStep(ctx, log, e.Ref.String(), phase.name, cp.Dir, stepCmd(ctx, log, cp.Dir, sandboxEnv(rec, e), argv), setupTimeout, save, false)
			if err != nil {
				return fmt.Errorf("%s %s: %s: %v; record not written; the command did not run", e.Ref, phase.name, row.Shown(), err)
			}
			if err := save(row); err != nil {
				return fmt.Errorf("%s %s: %s %s: %v; record not written", e.Ref, phase.name, row.Shown(), row.Outcome(), err)
			}
			if row.Failed() {
				return fmt.Errorf("%s %s: %s %s", e.Ref, phase.name, row.Shown(), row.Outcome())
			}
		}
	}
	return nil
}

// runCleanup runs every Cleanup command of e's copy cp, each under its own
// cleanupTimeout, even after one fails, writing each row to the sandbox's
// record before and after the command. Stopping processes matters more than
// the trace, so a failed write does not stop a command; it counts as a failed
// Cleanup. It returns the record's cleanup value, and, when a command or a
// write failed, a report naming each command's outcome.
func runCleanup(box string, rec *Record, e Entry, cp *recipe.Recipe) (status, report string) {
	log, err := openLog(box)
	if err != nil {
		return "failed: no Cleanup ran", fmt.Sprintf("%s cleanup: no command ran (%v)", e.Ref, err)
	}
	defer log.Close()
	write := func() error { return writeJSON(filepath.Join(box, "sandbox.json"), rec) }
	var failed, outcomes []string
	for _, argv := range cp.Cleanup {
		var werr error
		put := rowSaver(&rec.Commands, write)
		save := func(c cmdrecord.Run) error {
			if err := put(c); err != nil && werr == nil {
				werr = err
			}
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		row, _ := runStep(ctx, log, e.Ref.String(), recipe.Cleanup, cp.Dir, stepCmd(ctx, log, cp.Dir, sandboxEnv(rec, e), argv), cleanupTimeout, save, false)
		cancel()
		save(row)
		outcome := row.Shown() + " " + row.Outcome()
		if werr != nil {
			outcome += fmt.Sprintf(", a record write failed (%v)", werr)
		}
		if row.Failed() || werr != nil {
			failed = append(failed, row.Shown())
		}
		outcomes = append(outcomes, outcome)
	}
	if len(failed) == 0 {
		return "ok", ""
	}
	return "failed: " + strings.Join(failed, "; "), e.Ref.String() + " cleanup: " + strings.Join(outcomes, "; ")
}

func openLog(box string) (*os.File, error) {
	return os.OpenFile(filepath.Join(box, "recipe.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
}

// rowSaver returns a save that puts one row in *rows, appended the first
// time and replaced after, and then writes the record. A pending row (no
// ended) whose write fails is taken out again: a row that never started is
// never written.
func rowSaver(rows *[]cmdrecord.Run, write func() error) func(cmdrecord.Run) error {
	i := -1
	return func(c cmdrecord.Run) error {
		if i < 0 {
			*rows = append(*rows, c)
			i = len(*rows) - 1
		} else {
			(*rows)[i] = c
		}
		err := write()
		if err != nil && c.Ended == "" {
			*rows, i = (*rows)[:i], -1
		}
		return err
	}
}

// runStep runs cmd, appending a "== REF step: command" line to log first,
// and returns its row, the script shown as ./PATH inside bundle. It first
// passes the row, with started and nothing after, to save, and runs nothing
// when save fails: it returns that pending row and the error. cmd's
// environment, directory and output are set by stepCmd or checkCmd; its
// output goes to open files, never a pipe, so a server a command leaves
// running in the background cannot keep hyp waiting. The command leads its
// own process group, which its row records once it starts; when ctx (the
// one cmd was made with) ends the whole group is killed, and the row reports
// "interrupted" when ctx was cancelled (Ctrl-C) or a timeout after timeout
// when its deadline passed. With endGroup, the group is also killed when the
// command ends, so nothing it left in the background outlives it.
func runStep(ctx context.Context, log *os.File, ref string, step recipe.Step, bundle string, cmd *exec.Cmd, timeout time.Duration, save func(cmdrecord.Run) error, endGroup bool) (cmdrecord.Run, error) {
	row := cmdrecord.Run{Ref: ref, Step: step, Args: rowArgs(bundle, cmd.Args)}
	row.Started = now()
	if err := save(row); err != nil {
		return row, err
	}
	fmt.Fprintf(log, "== %s %s: %s\n", ref, step, row.Shown())
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	err := cmd.Start()
	if err == nil {
		pid := cmd.Process.Pid
		row.PGID = pid
		if serr := save(row); serr != nil {
			syscall.Kill(-pid, syscall.SIGKILL)
			cmd.Wait()
			return row, serr
		}
		err = cmd.Wait()
		if endGroup {
			syscall.Kill(-pid, syscall.SIGKILL)
		}
	}
	row.Ended = now()
	var ee *exec.ExitError
	switch {
	case err != nil && errors.Is(ctx.Err(), context.Canceled):
		row.Error = "interrupted"
	case err != nil && ctx.Err() != nil:
		row.Error = "timed out after " + timeout.String()
	case errors.As(err, &ee) && ee.ExitCode() >= 0:
		code := ee.ExitCode()
		row.Exit = &code
	case err != nil:
		row.Error = err.Error()
	default:
		zero := 0
		row.Exit = &zero
	}
	return row, nil
}

// rowArgs is argv as a row records it: the script as ./PATH inside bundle,
// then its words.
func rowArgs(bundle string, argv []string) []string {
	rel, err := filepath.Rel(bundle, argv[0])
	if err != nil {
		rel = argv[0]
	}
	return append([]string{"./" + filepath.ToSlash(rel)}, argv[1:]...)
}

// withoutSandboxVars drops every HYP_SANDBOX_ entry from env.
func withoutSandboxVars(env []string) []string {
	return slices.DeleteFunc(env, func(kv string) bool { return strings.HasPrefix(kv, "HYP_SANDBOX_") })
}

// stepCmd is argv as a sandbox step: in cwd, with the caller's environment
// without any HYP_SANDBOX_ variable, then extra, and its output to log.
func stepCmd(ctx context.Context, log *os.File, cwd string, extra, argv []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = cwd
	cmd.Stdout, cmd.Stderr = log, log
	// Environ sets PWD to Dir; later entries win over the caller's.
	cmd.Env = append(withoutSandboxVars(cmd.Environ()), extra...)
	return cmd
}

// sandboxEnv is what a command run in the sandbox gets: the sandbox's
// variables, e's own tmp directory in HYP_SANDBOX_RECIPE_TMP, and the pinned
// hyp first on PATH.
func sandboxEnv(rec *Record, e Entry) []string {
	env := []string{
		"HYP_SANDBOX_ID=" + rec.ID, "HYP_SANDBOX_WT=" + rec.WT, "HYP_SANDBOX_TMP=" + rec.Tmp,
		"HYP_SANDBOX_RECIPE_TMP=" + filepath.Join(rec.Tmp, e.Name),
		"PATH=" + rec.Bin + string(os.PathListSeparator) + os.Getenv("PATH"),
	}
	if rec.Inputs != nil {
		env = append(env, "HYP_SANDBOX_INPUTS="+rec.Inputs.Path)
	}
	return env
}

// logTail returns the last n lines of <box>/recipe.log, or "" without one.
func logTail(box string, n int) string {
	data, err := os.ReadFile(filepath.Join(box, "recipe.log"))
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	return strings.Join(lines[max(0, len(lines)-n):], "\n")
}

func isRepoRoot(host string) bool {
	top, err := git("-C", host, "rev-parse", "--show-toplevel")
	if err != nil {
		return false
	}
	h, err := filepath.Abs(host)
	if err != nil {
		return false
	}
	h, err1 := filepath.EvalSymlinks(h)
	t, err2 := filepath.EvalSymlinks(top)
	return err1 == nil && err2 == nil && h == t
}

func hypVersion(bin string) (string, error) {
	if !isFile(bin) {
		return "", errors.New("not a file")
	}
	out, err := exec.Command(bin, "version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// build makes the sandbox's contents and record, writing cut, when there is
// one, to wt/.tree. For each recipe it also returns the recipe opened from
// its copy in <box>/recipes, whose commands point there, in stack order.
func build(o CreateOptions, box, version string, rcps []*recipe.Recipe, kas []*cmdrecord.KnownAnswers, cut *store.Cut) (*Record, []*recipe.Recipe, error) {
	wt, tmp, bin := filepath.Join(box, "wt"), filepath.Join(box, "tmp"), filepath.Join(box, "bin")
	for _, args := range [][]string{
		{"clone", "--quiet", "--no-checkout", o.Host, wt},
		{"-C", wt, "checkout", "--quiet", "--detach", o.Base},
		{"-C", wt, "remote", "remove", "origin"},
	} {
		if _, err := git(args...); err != nil {
			return nil, nil, err
		}
	}
	for _, d := range []string{tmp, bin} {
		if err := os.Mkdir(d, 0o755); err != nil {
			return nil, nil, err
		}
	}
	var src string
	if cut != nil {
		var err error
		if src, err = filepath.Abs(cut.Src); err != nil {
			return nil, nil, err
		}
		if err := cut.Write(wt); err != nil {
			return nil, nil, err
		}
	}
	hyp := filepath.Join(bin, "hyp")
	if err := copyFile(o.Hyp, hyp, 0o755); err != nil {
		return nil, nil, err
	}
	var inputs *Inputs
	if o.Inputs != "" {
		src, err := filepath.Abs(o.Inputs)
		if err != nil {
			return nil, nil, err
		}
		dst := filepath.Join(box, "inputs", filepath.Base(src))
		if err := os.Mkdir(filepath.Dir(dst), 0o755); err != nil {
			return nil, nil, err
		}
		if err := copyFile(src, dst, 0o644); err != nil {
			return nil, nil, err
		}
		h, err := fileSum(dst)
		if err != nil {
			return nil, nil, err
		}
		inputs = &Inputs{Source: src, Path: dst, SHA256: h}
	}
	data, err := os.ReadFile(hyp)
	if err != nil {
		return nil, nil, err
	}
	sum := sha256.Sum256(data)
	base, err := git("-C", wt, "rev-parse", "HEAD")
	if err != nil {
		return nil, nil, err
	}
	host, err := filepath.Abs(o.Host)
	if err != nil {
		return nil, nil, err
	}
	rec := &Record{
		Format: SandboxFormat, ID: o.ID, Host: host, Base: base, WT: wt, Tmp: tmp, Bin: bin,
		HypSHA256: hex.EncodeToString(sum[:]), HypVersion: version,
		Recipes: []Entry{{Ref: event.RefNone}}, Inputs: inputs, Store: src, Cut: o.Cut, Created: now(),
	}
	var cps []*recipe.Recipe
	if len(rcps) > 0 {
		rec.Recipes = nil
		if err := os.Mkdir(filepath.Join(box, "recipes"), 0o755); err != nil {
			return nil, nil, err
		}
	}
	for i, rcp := range rcps {
		cp, err := copyRecipe(o.Refs[i], rcp, filepath.Join(box, "recipes"))
		if err != nil {
			return nil, nil, err
		}
		if err := os.Mkdir(filepath.Join(tmp, rcp.Name), 0o755); err != nil {
			return nil, nil, err
		}
		e := Entry{Ref: rcp.Ref(), Name: rcp.Name, Version: rcp.Version, SHA256: rcp.Hash, Path: cp.Path, KnownAnswers: kas[i]}
		if len(cp.Launch)+len(cp.Doctor) == 0 {
			e.Setup = "none"
		}
		rec.Recipes = append(rec.Recipes, e)
		cps = append(cps, cp)
	}
	if err := writeJSON(filepath.Join(box, "sandbox.json"), rec); err != nil {
		return nil, nil, err
	}
	return rec, cps, nil
}

// copyRecipe copies rcp into dir (a sandbox's recipes/, outside wt/ so the
// agent working there does not find it, or a temporary directory), opens the
// copy by ref and checks that it hashes the same.
func copyRecipe(ref string, rcp *recipe.Recipe, dir string) (*recipe.Recipe, error) {
	dst := filepath.Join(dir, rcp.Name)
	if err := os.CopyFS(dst, os.DirFS(rcp.Dir)); err != nil {
		return nil, err
	}
	cp, err := recipe.OpenAs(dst, ref)
	if err != nil {
		return nil, err
	}
	if cp.Hash != rcp.Hash {
		return nil, herr.Env("recipe %s changed while it was copied", ref)
	}
	return cp, nil
}

// Remove deletes the sandbox o.ID, first copying its tree, kept files and
// record to o.OutDir unless o.Discard. Every check runs before anything is
// written or deleted.
func Remove(o RemoveOptions) (*Record, error) {
	if (o.OutDir == "") == !o.Discard {
		return nil, herr.BadArgs("remove needs OUT_DIR or --discard, not both")
	}
	r, err := root()
	if err != nil {
		return nil, err
	}
	if !idRe.MatchString(o.ID) {
		return nil, herr.BadArgs("bad sandbox id %q; use [a-z0-9-], at most 40 characters", o.ID)
	}
	box := filepath.Join(r, o.ID)
	meta := filepath.Join(box, "sandbox.json")
	if !isFile(meta) {
		return nil, herr.BadArgs("no sandbox %s (%s missing)", o.ID, meta)
	}
	if ok, _ := inside(r, box); !ok {
		return nil, herr.Refused("%s is not under %s", box, r)
	}
	if o.OutDir != "" {
		if _, err := os.Lstat(o.OutDir); err == nil {
			return nil, herr.BadArgs("%s exists; choose a new path", o.OutDir)
		}
		// The record would be deleted with the sandbox it describes.
		abs, err := filepath.Abs(o.OutDir)
		if in, _ := inside(box, abs); err != nil || in || strings.HasPrefix(abs, box+string(filepath.Separator)) {
			return nil, herr.BadArgs("%s is inside the sandbox; choose a path outside %s", o.OutDir, box)
		}
	}
	// Held through the delete. Another remove may have deleted the sandbox
	// between the check above and the lock, so check again.
	unlock, err := lockBox(box, o.ID, "remove")
	if err != nil {
		return nil, err
	}
	defer unlock()
	if !isFile(meta) {
		return nil, herr.BadArgs("no sandbox %s (%s missing)", o.ID, meta)
	}
	rec, err := readRecord(meta, o.ID)
	if err != nil {
		return nil, err
	}
	for _, p := range []string{rec.WT, rec.Tmp, rec.Bin} {
		if ok, _ := inside(r, p); !ok {
			return nil, herr.Refused("%v is not under %s", p, r)
		}
	}
	// Recipes are cleaned up in reverse stack order.
	type cleanup struct {
		e  *Entry
		cp *recipe.Recipe
	}
	var cleanups []cleanup
	for i := len(rec.Recipes) - 1; i >= 0; i-- {
		e := &rec.Recipes[i]
		cp, err := openCopy(r, *e)
		if err != nil {
			e.Cleanup = "failed: no Cleanup ran"
			if o.Warn != nil {
				fmt.Fprintf(o.Warn, "sandbox: %s at %s unreadable (%v); no Cleanup ran\n", e.Ref, e.Path, err)
			}
		} else if cp != nil && len(cp.Cleanup) > 0 {
			cleanups = append(cleanups, cleanup{e, cp})
		}
	}
	if err := needTools("lsof"); err != nil {
		return nil, err
	}
	keep, err := keepPaths(box, o.Keep)
	if err != nil {
		return nil, err
	}

	// A failed Cleanup only warns: the leftover check below decides.
	for _, c := range cleanups {
		var report string
		c.e.Cleanup, report = runCleanup(box, rec, *c.e, c.cp)
		if report != "" && o.Warn != nil {
			fmt.Fprintf(o.Warn, "sandbox: %s; see recipe.log\n", report)
		}
	}
	// After Cleanup, so a recipe stops its own processes first.
	stopGroups(rec.Commands, o.Warn)
	// So a sandbox kept for a leftover process records its cleanup values.
	if len(cleanups) > 0 {
		if err := writeJSON(meta, rec); err != nil && o.Warn != nil {
			fmt.Fprintf(o.Warn, "sandbox: %v; the sandbox's record does not say how Cleanup went\n", err)
		}
	}
	held, err := leftovers(box)
	if err != nil {
		return nil, err
	}
	if held != "" {
		return nil, herr.Env("%s\nsandbox: processes still run in %s; stop them, then remove again; nothing deleted. A shell or editor of your own whose working directory is in wt/ counts: cd out of the sandbox first", held, box)
	}

	if !o.Discard {
		if err := saveRecord(o, rec, rec.WT, box, keep); err != nil {
			os.RemoveAll(o.OutDir) // proved absent above, so it is ours
			return nil, herr.Env("%v; sandbox kept", err)
		}
	}
	if err := os.RemoveAll(box); err != nil {
		return nil, herr.Env("%v", err)
	}
	return rec, nil
}

// openCopy opens the bundle copy e names in the sandbox root r by e's ref; nil
// for none, which has no commands. A copy that is missing
// is an environment error; one outside r, one that does not open, or one that
// no longer hashes as the record's sha256 is refused, so no command runs from
// a copy the record does not vouch for.
func openCopy(r string, e Entry) (*recipe.Recipe, error) {
	if e.Ref.IsNone() {
		return nil, nil
	}
	var got *recipe.Recipe
	err := herr.Refused("not a directory under %s", r)
	if fi, serr := os.Stat(e.Path); serr != nil {
		err = herr.Env("%v", serr)
	} else if ok, _ := inside(r, e.Path); ok && fi.IsDir() {
		got, err = recipe.OpenAs(e.Path, e.Ref.String())
	}
	if err != nil {
		return nil, err
	}
	if e.SHA256 != "" && got.Hash != e.SHA256 {
		return nil, herr.Refused("copy hashes %s, the record says %s", got.Hash, e.SHA256)
	}
	return got, nil
}

// readRecord reads the record at path. A field this hyp does not write is
// kept in Extra, so it is copied to OUT_DIR as it was read.
func readRecord(path, id string) (*Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, herr.Env("%v", err)
	}
	var rec Record
	ok := json.Unmarshal(data, &rec) == nil && rec.ID == id && rec.WT != "" && rec.Tmp != "" && rec.Bin != ""
	for _, e := range rec.Recipes {
		if e.Ref.IsZero() || (!e.Ref.IsNone() && e.Path == "") {
			ok = false
		}
	}
	if !ok {
		return nil, herr.Env("corrupt record %s", path)
	}
	return &rec, nil
}

// The names hyp writes into OUT_DIR when it removes a sandbox.
const (
	keptTree         = "tree"
	keptRecord       = "sandbox.json"
	keptRecipeLog    = "recipe.log"
	keptInputs       = "inputs"
	keptChanges      = "changes"
	keptChangesJSON  = "changes.json"
	keptKnownAnswers = "known-answers.log"
)

// KeptFiles is every name hyp writes into OUT_DIR on remove; no --keep path
// may take one as its basename.
var KeptFiles = []string{keptTree, keptRecord, keptRecipeLog, keptInputs, keptChanges, keptChangesJSON, keptKnownAnswers}

// keepPaths checks that each kept path is relative, stays inside box and
// has a basename no other kept file or the record uses.
func keepPaths(box string, keep []string) ([]string, error) {
	names := map[string]bool{}
	for _, n := range KeptFiles {
		names[n] = true
	}
	var out []string
	for _, k := range keep {
		p := filepath.Join(box, k)
		if filepath.IsAbs(k) || !strings.HasPrefix(p, box+string(filepath.Separator)) {
			return nil, herr.BadArgs("--keep %s is not a path inside the sandbox", k)
		}
		if _, err := os.Lstat(p); err == nil {
			if ok, _ := inside(box, p); !ok {
				return nil, herr.BadArgs("--keep %s is not a path inside the sandbox", k)
			}
		}
		name := filepath.Base(p)
		if names[name] {
			return nil, herr.BadArgs("--keep %s: name %s is already taken in OUT_DIR", k, name)
		}
		names[name] = true
		out = append(out, k)
	}
	return out, nil
}

func saveRecord(o RemoveOptions, rec *Record, wt, box string, keep []string) error {
	if err := os.MkdirAll(o.OutDir, 0o755); err != nil {
		return err
	}
	tree := filepath.Join(wt, ".tree")
	if fi, err := os.Stat(tree); err == nil && fi.IsDir() {
		if err := copyTree(tree, filepath.Join(o.OutDir, keptTree)); err != nil {
			return err
		}
	}
	for _, name := range []string{keptRecipeLog, keptKnownAnswers} {
		if log := filepath.Join(box, name); isFile(log) {
			if err := copyFile(log, filepath.Join(o.OutDir, name), 0o644); err != nil {
				return err
			}
		}
	}
	if rec.Inputs != nil {
		dst := filepath.Join(o.OutDir, keptInputs, filepath.Base(rec.Inputs.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := copyFile(rec.Inputs.Path, dst, 0o644); err != nil {
			return err
		}
	}
	for _, k := range keep {
		src := filepath.Join(box, k)
		if !isFile(src) {
			if o.Warn != nil {
				fmt.Fprintf(o.Warn, "sandbox: --keep %s not found; skipped\n", k)
			}
			continue
		}
		if err := copyFile(src, filepath.Join(o.OutDir, filepath.Base(src)), 0o644); err != nil {
			return err
		}
	}
	if err := writeChanges(box, wt, o.OutDir, o.Warn); err != nil {
		return err
	}
	rec.Removed = now()
	return writeJSON(filepath.Join(o.OutDir, keptRecord), rec)
}

// copyTree copies a hyp tree without the index.sqlite* files older hyp
// binaries write (ADR 0006).
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(d.Name(), "index.sqlite") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

// lastLines is "\n" and the last n lines of text, or "" for empty text.
func lastLines(text string, n int) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if strings.TrimSpace(text) == "" {
		return ""
	}
	return "\n" + strings.Join(lines[max(0, len(lines)-n):], "\n")
}
