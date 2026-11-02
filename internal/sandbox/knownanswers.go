package sandbox

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
)

// SandboxFormat and ChangesFormat are the shape numbers of sandbox.json and
// changes.json (0: not written yet); cmdrecord.Format is command.json's.
const SandboxFormat, ChangesFormat = 2, 1

var caseTimeout = 60 * time.Second // each run of a known-answer case
var gateTimeout = 10 * time.Minute // every case of one bundle

var caseNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// knownCase is one known-answers/NAME/ directory: its expect.json, its
// expect/ tree (nil without one) and the argv its command resolved to in
// the bundle it was read from.
type knownCase struct {
	Name, Dir    string
	Step         recipe.Step
	Command      string
	Args         []string
	Exit         int
	Stdout, Repo string
	Expect       map[string]outEntry
	argv         []string
}

// expectFile is expect.json; pointers tell a missing key from a zero value.
type expectFile struct {
	Step    *string  `json:"step"`
	Command *string  `json:"command"`
	Args    []string `json:"args"`
	Exit    *int     `json:"exit"`
	Stdout  *string  `json:"stdout"`
	Repo    string   `json:"repo"`
}

// readCases reads r's known-answers/ cases, sorted by name, and refuses
// (exit 1) any case that breaks the case contract (RECIPE-FORMAT.md, Known
// answers). A bundle with no known-answers/ has no cases.
func readCases(r *recipe.Recipe) ([]knownCase, error) {
	ref := r.Ref()
	root := filepath.Join(r.Dir, "known-answers")
	fi, err := os.Stat(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, herr.Env("recipe %s: %v", ref, err)
	}
	if !fi.IsDir() {
		return nil, herr.Refused("recipe %s: known-answers is not a directory", ref)
	}
	ents, err := os.ReadDir(root)
	if err != nil {
		return nil, herr.Env("recipe %s: %v", ref, err)
	}
	var cases []knownCase
	for _, ent := range ents {
		c, err := readCase(r, ref.String(), root, ent)
		if err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, nil
}

// readCase reads and checks one entry of known-answers/.
func readCase(r *recipe.Recipe, ref, root string, ent os.DirEntry) (knownCase, error) {
	name := ent.Name()
	refuse := func(format string, a ...any) (knownCase, error) {
		return knownCase{}, herr.Refused("recipe %s: known answer %s: %s", ref, name, fmt.Sprintf(format, a...))
	}
	if !ent.IsDir() {
		return refuse("is not a directory; a case is known-answers/NAME/ holding input/ and expect.json")
	}
	if !caseNameRe.MatchString(name) {
		return refuse("the name must match [a-z0-9][a-z0-9-]*")
	}
	dir := filepath.Join(root, name)
	if fi, err := os.Stat(filepath.Join(dir, "input")); err != nil || !fi.IsDir() {
		return refuse("input/ is missing")
	}
	data, err := os.ReadFile(filepath.Join(dir, "expect.json"))
	if err != nil {
		return refuse("expect.json is missing")
	}
	var e expectFile
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		return refuse("expect.json: %v", err)
	}
	for _, k := range []struct {
		key     string
		missing bool
	}{{"step", e.Step == nil}, {"command", e.Command == nil}, {"exit", e.Exit == nil}, {"stdout", e.Stdout == nil}} {
		if k.missing {
			return refuse("expect.json has no %q", k.key)
		}
	}
	c := knownCase{Name: name, Dir: dir, Step: recipe.Step(*e.Step), Command: *e.Command, Args: e.Args, Exit: *e.Exit, Stdout: *e.Stdout, Repo: e.Repo}
	switch c.Step {
	case recipe.Evidence:
		if e.Args != nil {
			return refuse("args is not allowed with step evidence; an Evidence command gets only OUT_DIR")
		}
		for _, argv := range r.Evidence {
			if shown(r.Dir, argv) == c.Command {
				c.argv = argv
			}
		}
		if c.argv == nil {
			return refuse("command %q matches no Evidence line", c.Command)
		}
	case recipe.Standalone:
		for _, n := range r.Standalone {
			if n.Name == c.Command {
				c.argv = n.Argv
			}
		}
		if c.argv == nil {
			return refuse("command %q matches no Standalone label", c.Command)
		}
	default:
		return refuse("step %q is not evidence or standalone", c.Step)
	}
	if c.Repo != "" {
		fi, err := os.Lstat(filepath.Join(dir, "input", c.Repo))
		if c.Repo != filepath.Base(c.Repo) || c.Repo == "." || c.Repo == ".." || err != nil || !fi.Mode().IsRegular() {
			return refuse("repo %q is not a regular file directly under input/", c.Repo)
		}
		if heads, ok := bundleRefs(filepath.Join(dir, "input", c.Repo)); ok && !slices.Contains(heads, "HEAD") {
			return refuse("repo %q holds no HEAD, so its clone would have no checkout; make it with git bundle create FILE HEAD <branch>", c.Repo)
		}
	}
	if fi, err := os.Lstat(filepath.Join(dir, "expect")); err == nil {
		if !fi.IsDir() {
			return refuse("expect is not a directory; expect/ holds the files the command must leave in $HYP_OUT")
		}
		tree, problem := readTree(filepath.Join(dir, "expect"), "expect", "")
		if problem != "" {
			return refuse("%s; expect/ holds only directories and regular files", problem)
		}
		c.Expect = tree
	}
	for _, f := range []struct {
		file string
		want int
	}{{"sandbox.json", SandboxFormat}, {"command.json", cmdrecord.Format}, {"changes.json", ChangesFormat}} {
		data, err := os.ReadFile(filepath.Join(dir, "input", f.file))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		var rec struct {
			Format int `json:"format"`
		}
		if err != nil || json.Unmarshal(data, &rec) != nil {
			return refuse("input/%s does not read as a record", f.file)
		}
		if rec.Format != f.want {
			return refuse("input/%s has format %d; this hyp writes format %d (remake the input)", f.file, rec.Format, f.want)
		}
	}
	return c, nil
}

// bundleRefs returns the ref names a git bundle file lists in its header,
// such as HEAD and refs/heads/main, and false when path does not read as a
// v2 or v3 git bundle; git clone then reports what is wrong with it.
func bundleRefs(path string) ([]string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	br := bufio.NewReader(f)
	sig, err := br.ReadString('\n')
	if err != nil || (sig != "# v2 git bundle\n" && sig != "# v3 git bundle\n") {
		return nil, false
	}
	var refs []string
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return nil, false
		}
		line = strings.TrimSuffix(line, "\n")
		switch {
		case line == "":
			return refs, true
		case strings.HasPrefix(line, "@"), strings.HasPrefix(line, "-"):
			// a v3 capability or a prerequisite commit
		default:
			_, name, found := strings.Cut(line, " ")
			if !found {
				return nil, false
			}
			refs = append(refs, name)
		}
	}
}

// shown is argv as a command row shows it: the script as ./PATH inside
// bundle, then its words.
func shown(bundle string, argv []string) string { return strings.Join(rowArgs(bundle, argv), " ") }

// RunKnownAnswers runs every known-answer case of the bundle r, each twice
// and statelessly, from a temporary copy that hashes the same, and returns
// each case's outcome. Each case's stderr goes to log, after a line naming
// the case. A case finds hyp as the copy hyp holds, first on PATH, never a
// sandbox's --hyp; its hash is the one KnownAnswers names. It returns an
// error only when a case breaks the case contract (exit 1) or I/O fails
// (exit 3); a case that fails is an outcome, not an error.
func RunKnownAnswers(r *recipe.Recipe, hyp *SandboxHyp, log io.Writer) (*cmdrecord.KnownAnswers, error) {
	ref := r.Ref()
	ka := &cmdrecord.KnownAnswers{HypSHA256: hyp.SHA256, Cases: []cmdrecord.Case{}}
	tmp, err := os.MkdirTemp("", "hyp-ka-bundle-")
	if err != nil {
		return nil, herr.Env("%v", err)
	}
	defer os.RemoveAll(tmp)
	cp, err := copyRecipe(ref.String(), r, tmp)
	if err != nil {
		return nil, err
	}
	cases, err := readCases(cp)
	if err != nil {
		return nil, err
	}
	ictx, stop := interrupted()
	defer stop()
	gate, cancel := context.WithTimeout(ictx, gateTimeout)
	defer cancel()
	for _, c := range cases {
		fmt.Fprintf(log, "== %s known answer %s\n", ref, c.Name)
		outcome, err := judgeCase(gate, c, hyp.Dir, log)
		if err == nil && ictx.Err() != nil {
			err = herr.Env("recipe %s: known answers interrupted", ref)
		}
		if err != nil {
			return nil, err
		}
		ka.Cases = append(ka.Cases, cmdrecord.Case{Name: c.Name, Outcome: outcome})
	}
	return ka, nil
}

// Gate runs r's known answers before a run of r (ADR 0002). It returns nil,
// nil for a recipe that is not an instrument. An instrument with no case is
// refused; a case that timed out is an environment error (exit 3), and any
// other case that is not ok refuses the run. The KnownAnswers value is
// returned with every refusal it ran cases for. Its cases find hyp as the
// copy hyp holds.
func Gate(r *recipe.Recipe, hyp *SandboxHyp, log io.Writer) (*cmdrecord.KnownAnswers, error) {
	if !r.IsInstrument() {
		return nil, nil
	}
	ref := r.Ref()
	if !recipe.HasKnownAnswers(r.Dir) {
		return nil, herr.Refused("recipe %s is an instrument with no known answers; add known-answers/NAME/ cases (RECIPE-FORMAT.md, Known answers)", ref)
	}
	ka, err := RunKnownAnswers(r, hyp, log)
	if err != nil {
		return nil, err
	}
	var timedOut, failed []string
	for _, c := range ka.Cases {
		switch {
		case c.Outcome == "ok":
		case strings.HasPrefix(c.Outcome, "timed out"):
			timedOut = append(timedOut, c.Name+": "+c.Outcome)
		default:
			failed = append(failed, c.Name+": "+c.Outcome)
		}
	}
	if len(timedOut) > 0 {
		return ka, herr.Env("recipe %s: known answers timed out: %s; nothing ran", ref, strings.Join(timedOut, "; "))
	}
	if len(failed) > 0 {
		return ka, herr.Refused("recipe %s failed its known answers: %s; nothing ran", ref, strings.Join(failed, "; "))
	}
	return ka, nil
}

// ExternalWarnings lists, for hyp recipe check, each Evidence command (its
// shown form) and Standalone label of r whose External: line is missing,
// "<command>: no External: line", or does not state its reach, "<command>:
// External: not stated (TODO)" (recipe.External, ADR 0008).
func ExternalWarnings(r *recipe.Recipe) []string {
	var out []string
	add := func(command string, e recipe.External) {
		switch {
		case !e.Line:
			out = append(out, command+": no External: line")
		case !e.Stated():
			out = append(out, command+": External: not stated (TODO)")
		}
	}
	for i, argv := range r.Evidence {
		add(shown(r.Dir, argv), r.EvidenceExternal[i])
	}
	for _, n := range r.Standalone {
		add(n.Name, n.External)
	}
	return out
}

// WeakCases lists what the known answers of instrument r lack, for hyp
// recipe check's warnings: "instrument with no known answers", a reason the
// cases do not read, or "<command>: no must-pass case" and "<command>: no
// must-fail case" for each Evidence command (its shown form) and Standalone
// label without a case expecting exit 0, or a non-zero exit. For a verdict
// check, one whose cases' stdout prints a verdict line, it also lists
// "<command>: case <name>: malformed verdict line ...", "<command>: case
// <name>: label <label> printed twice", and "<command>: verdict <label>: no
// supports case" or "no contradicts case" for each label no case prints so
// (ADR 0004). A recipe that is not an instrument lacks nothing.
func WeakCases(r *recipe.Recipe) []string {
	if !r.IsInstrument() {
		return nil
	}
	if !recipe.HasKnownAnswers(r.Dir) {
		return []string{"instrument with no known answers"}
	}
	cases, err := readCases(r)
	if err != nil {
		return []string{herr.Text(err)}
	}
	type key struct {
		step    recipe.Step
		command string
	}
	pass, fail := map[key]bool{}, map[key]bool{}
	lines := map[key][]string{}            // each case's malformed or repeated verdict lines
	words := map[key]map[string][]string{} // label -> the words its cases print
	for _, c := range cases {
		k := key{c.Step, c.Command}
		if c.Exit == 0 {
			pass[k] = true
		} else {
			fail[k] = true
		}
		seen := map[string]int{}
		for line := range strings.Lines(c.Stdout) {
			line = strings.TrimSuffix(line, "\n")
			if !strings.HasPrefix(line, "verdict:") {
				continue
			}
			m := verdictLineRe.FindStringSubmatch(line)
			if m == nil {
				lines[k] = append(lines[k], fmt.Sprintf("case %s: malformed verdict line %q", c.Name, line))
				continue
			}
			if seen[m[1]]++; seen[m[1]] > 1 {
				if seen[m[1]] == 2 {
					lines[k] = append(lines[k], fmt.Sprintf("case %s: label %s printed twice", c.Name, m[1]))
				}
				continue
			}
			if words[k] == nil {
				words[k] = map[string][]string{}
			}
			words[k][m[1]] = append(words[k][m[1]], m[2])
		}
	}
	var keys []key
	for _, argv := range r.Evidence {
		keys = append(keys, key{recipe.Evidence, shown(r.Dir, argv)})
	}
	for _, n := range r.Standalone {
		keys = append(keys, key{recipe.Standalone, n.Name})
	}
	var out []string
	for _, k := range keys {
		if !pass[k] {
			out = append(out, k.command+": no must-pass case")
		}
		if !fail[k] {
			out = append(out, k.command+": no must-fail case")
		}
		for _, l := range lines[k] {
			out = append(out, k.command+": "+l)
		}
		for _, label := range slices.Sorted(maps.Keys(words[k])) {
			for _, w := range []string{event.VerdictSupports, event.VerdictContradicts} {
				if !slices.Contains(words[k][label], w) {
					out = append(out, k.command+": verdict "+label+": no "+w+" case")
				}
			}
		}
	}
	return out
}

// verdictLineRe is a well-formed verdict line, "verdict: LABEL=WORD": LABEL
// matches the alias pattern of internal/replay/validate.go and WORD is a
// Result verdict.
var verdictLineRe = regexp.MustCompile(`^verdict: ([a-z][a-z0-9-]{0,31})=(` +
	strings.Join(slices.Sorted(maps.Keys(event.Verdicts)), "|") + `)$`)

// judgeCase runs c twice, or once when the first run times out, and returns
// its outcome (RECIPE-FORMAT.md, Known answers).
func judgeCase(gate context.Context, c knownCase, bin string, log io.Writer) (string, error) {
	var runs []caseRun
	for range 2 {
		if gate.Err() != nil {
			return "timed out after " + gateTimeout.String(), nil
		}
		r, timedOut, err := runCase(gate, c, bin, log)
		if err != nil {
			return "", err
		}
		if timedOut && gate.Err() != nil {
			return "timed out after " + gateTimeout.String(), nil
		}
		// A run that timed out cannot pass, so there is no second run.
		if timedOut {
			return "timed out after " + caseTimeout.String(), nil
		}
		if r.problem != "" {
			return r.problem, nil
		}
		runs = append(runs, r)
	}
	a, b := runs[0], runs[1]
	switch {
	case a.exit != b.exit || a.stdout != b.stdout || !maps.Equal(a.out, b.out):
		return "not stateless", nil
	case a.exit != c.Exit:
		return fmt.Sprintf("exit %d, want %d", a.exit, c.Exit), nil
	case a.stdout != c.Stdout:
		return "stdout differs", nil
	case c.Expect != nil:
		if p := firstDiff(a.out, c.Expect); p != "" {
			return "out differs: " + p, nil
		}
	case len(a.out) > 0:
		return "out not empty: " + slices.Min(slices.Collect(maps.Keys(a.out))), nil
	}
	return "ok", nil
}

// caseRun is one run of a case: its exit code, and its stdout and out tree
// with the case directory's path as $CASE; problem is why its out tree is
// not a product (see readTree).
type caseRun struct {
	exit            int
	stdout, problem string
	out             map[string]outEntry
}

// runCase runs c once as a check under a new case directory, which it
// removes after, and returns its run. The command runs in its own process
// group, killed when the command ends or after caseTimeout. Output goes to
// files, never a pipe, so a process left in the background cannot keep hyp
// waiting.
func runCase(gate context.Context, c knownCase, bin string, log io.Writer) (r caseRun, timedOut bool, err error) {
	tmp, err := os.MkdirTemp("", "hyp-ka-")
	if err != nil {
		return r, false, herr.Env("%v", err)
	}
	defer os.RemoveAll(tmp)
	dir, err := checkRoot(tmp)
	if err != nil {
		return r, false, err
	}
	if err := os.CopyFS(filepath.Join(dir, "input"), os.DirFS(filepath.Join(c.Dir, "input"))); err != nil {
		return r, false, herr.Env("known answer %s: copying input: %v", c.Name, err)
	}
	if c.Repo != "" {
		// The clone gets the check's environment, so host git config and
		// GIT_ variables cannot change what the case reads.
		clone := exec.Command("git", "clone", "--quiet", filepath.Join(dir, "input", c.Repo), filepath.Join(dir, "repo"))
		clone.Dir, clone.Env = filepath.Join(dir, "cwd"), checkEnv(dir, bin)
		if out, err := clone.CombinedOutput(); err != nil {
			return r, false, herr.Env("known answer %s: git clone %s: %v: %s", c.Name, c.Repo, err, strings.TrimSpace(string(out)))
		}
	}
	argv := slices.Clone(c.argv)
	if c.Step == recipe.Evidence {
		argv = append(argv, filepath.Join(dir, "input"))
	} else {
		for _, a := range c.Args {
			argv = append(argv, strings.ReplaceAll(a, "$CASE", dir))
		}
	}
	// Outside dir, so the command cannot see or change them.
	outs, err := os.MkdirTemp("", "hyp-ka-out-")
	if err != nil {
		return r, false, herr.Env("%v", err)
	}
	defer os.RemoveAll(outs)
	outF, err1 := os.Create(filepath.Join(outs, "stdout"))
	errF, err2 := os.Create(filepath.Join(outs, "stderr"))
	if err := errors.Join(err1, err2); err != nil {
		return r, false, herr.Env("%v", err)
	}
	defer outF.Close()
	defer errF.Close()

	ctx, cancel := context.WithTimeout(gate, caseTimeout)
	defer cancel()
	cmd := checkCmd(ctx, dir, bin, argv, outF, errF)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	runErr := cmd.Run()
	if cmd.Process != nil {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	if e, err := os.ReadFile(errF.Name()); err == nil {
		log.Write(e)
	}
	out, err := os.ReadFile(outF.Name())
	if err != nil {
		return r, false, herr.Env("%v", err)
	}
	var ee *exec.ExitError
	switch {
	case runErr != nil && ctx.Err() != nil:
		return r, true, nil
	case errors.As(runErr, &ee):
		r.exit = ee.ExitCode()
	case runErr != nil:
		return r, false, herr.Env("known answer %s: %s: %v", c.Name, c.Command, runErr)
	}
	r.stdout = strings.ReplaceAll(string(out), dir, "$CASE")
	r.out, r.problem = readTree(filepath.Join(dir, "out"), "out", dir)
	return r, false, nil
}
