// Package channel is hyp's command line: it parses arguments, runs the act and
// turns what comes back into output and an exit code. It is the only place an
// error kind becomes an exit code.
package channel

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/acts"
	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
)

// versionLine is hyp's version, followed by the commit Go stamped into this
// build (first 12 hex digits, -dirty for a checkout with uncommitted
// changes). A build without VCS info (go test, go run, -buildvcs=false)
// prints the bare line.
func versionLine() string {
	line := "hyp 0.2.0-dev (events v2)"
	if info, ok := debug.ReadBuildInfo(); ok {
		line += buildSuffix(info.Settings)
	}
	return line
}

func buildSuffix(settings []debug.BuildSetting) string {
	var rev string
	dirty := false
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return ""
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if dirty {
		rev += "-dirty"
	}
	return " " + rev
}

const usage = `usage: hyp <command> [flags]

An act that writes needs an actor (--actor or $HYP_ACTOR) and a role
(--role or $HYP_ROLE): owner, grill, probe-writer or runner.

commands:
  add goal [G] --base COMMIT --words TEXT --line gap:TEXT [--line guard:TEXT ...]
           [--from G] [--why TEXT]
                                         record a Goal; with G, replace a Goal
                                         that is still written
  place G LINE/EXAMPLE [--why TEXT]      place the example folder
                                         goals/G/LINE/EXAMPLE on a line of Goal G
  run COMMIT G/LINE [--why TEXT]         run every placed example of a line of
                                         Goal G on COMMIT and record the Result
  select G TREE [--why TEXT]             select the candidate row TREE of Goal G;
                                         the report is the PR body
  status [G] [--role ROLE]               print the state of every Goal, or of G:
                                         its stop, its waits and the allowed
                                         acts as ready command lines
  help                                   print this text
  version                                print the hyp version

exit codes: 0 recorded or printed, 1 refused by a rule, 2 bad arguments,
3 environment problem
`

// commands are the top-level command names, sorted.
var commands = []string{"add", "help", "place", "run", "select", "status", "version"}

// Commands returns the sorted top-level command names.
func Commands() []string { return slices.Clone(commands) }

// Run executes args (without the program name) and returns the exit code:
// 0 for success, 1 for a refusal, 2 for bad arguments, 3 for anything else.
func Run(args []string, stdout, stderr io.Writer) int {
	err := dispatch(args, stdout, stderr)
	if err == nil {
		return 0
	}
	var done *printed
	if !errors.As(err, &done) {
		fmt.Fprintf(stderr, "hyp: %v\n", err)
	}
	switch herr.KindOf(err) {
	case herr.Refusal:
		return 1
	case herr.BadArguments:
		return 2
	}
	return 3
}

func dispatch(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return herr.BadArgs("no command")
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, versionLine())
		return nil
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	case "place":
		return place(args[1:], stdout, stderr)
	case "run":
		return runCmd(args[1:], stdout, stderr)
	case "select":
		return selectCmd(args[1:], stdout, stderr)
	case "status":
		return statusCmd(args[1:], stdout)
	case "add":
		if len(args) > 1 && args[1] == "goal" {
			return addGoal(args[2:], stdout, stderr)
		}
		fmt.Fprint(stderr, usage)
		return herr.BadArgs("add what? only 'add goal' exists")
	}
	fmt.Fprint(stderr, usage)
	return herr.BadArgs("unknown command %q", args[0])
}

// parseInterleaved parses args with fs, allowing the positional arguments to
// come before or between the flags, and returns them in order.
func parseInterleaved(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for rest := args; ; {
		if err := fs.Parse(rest); err != nil {
			return nil, err
		}
		if fs.NArg() == 0 {
			return positional, nil
		}
		positional = append(positional, fs.Arg(0))
		rest = fs.Args()[1:]
	}
}

// lineFlag collects repeated --line KIND:TEXT in order.
type lineFlag []acts.LineInput

func (l *lineFlag) String() string { return "" }

func (l *lineFlag) Set(s string) error {
	kind, text, ok := strings.Cut(s, ":")
	if !ok || (kind != events.KindGap && kind != events.KindGuard) {
		return fmt.Errorf("--line %q: want gap:TEXT or guard:TEXT", s)
	}
	*l = append(*l, acts.LineInput{Kind: kind, Text: text})
	return nil
}

func addGoal(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("add goal", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.AddGoalOpts
	var lines lineFlag
	fs.StringVar(&o.Actor, "actor", os.Getenv("HYP_ACTOR"), "")
	fs.StringVar(&o.Role, "role", os.Getenv("HYP_ROLE"), "")
	fs.StringVar(&o.Why, "why", "", "")
	fs.StringVar(&o.From, "from", "", "")
	fs.StringVar(&o.Words, "words", "", "")
	fs.StringVar(&o.Base, "base", "", "")
	fs.Var(&lines, "line", "")
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return herr.BadArgs("add goal: %v", err)
	}
	if len(positional) > 1 {
		return herr.BadArgs("add goal takes at most one Goal to replace, got %d", len(positional))
	}
	if len(positional) == 1 {
		o.Goal = positional[0]
	}
	if o.Base == "" {
		return herr.BadArgs("add goal needs --base COMMIT")
	}
	if o.Words == "" {
		return herr.BadArgs("add goal needs --words TEXT")
	}
	if n := len([]rune(o.Why)); n > logstore.MaxWhy {
		return herr.BadArgs("--why is %d runes; the limit is %d", n, logstore.MaxWhy)
	}
	o.Lines = lines
	cwd, err := os.Getwd()
	if err != nil {
		return herr.Env("working directory: %v", err)
	}
	o.Dir = cwd
	rep, err := acts.AddGoal(o)
	if err != nil {
		return finish(stdout, stderr, cwd, o.Goal, "", err)
	}
	var b strings.Builder
	verb := "recorded"
	if rep.Replaced {
		verb = "replaced"
	}
	fmt.Fprintf(&b, "%s %s\n", verb, rep.Goal)
	fmt.Fprintf(&b, "words: %s\n", rep.Words)
	fmt.Fprintf(&b, "H0: %s\n", rep.Base)
	for _, l := range rep.Lines {
		fmt.Fprintf(&b, "%s %s: %s\n", l.Key, l.Kind, l.Text)
	}
	return finish(stdout, stderr, cwd, rep.Goal, b.String(), nil)
}

func place(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("place", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.PlaceOpts
	fs.StringVar(&o.Actor, "actor", os.Getenv("HYP_ACTOR"), "")
	fs.StringVar(&o.Role, "role", os.Getenv("HYP_ROLE"), "")
	fs.StringVar(&o.Why, "why", "", "")
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return herr.BadArgs("place: %v", err)
	}
	if len(positional) != 2 {
		return herr.BadArgs("place takes a Goal and LINE/EXAMPLE, for example: hyp place G-NXHHVA line-1/a; got %d arguments", len(positional))
	}
	o.Goal, o.Key = positional[0], positional[1]
	if n := len([]rune(o.Why)); n > logstore.MaxWhy {
		return herr.BadArgs("--why is %d runes; the limit is %d", n, logstore.MaxWhy)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return herr.Env("working directory: %v", err)
	}
	o.Dir = cwd
	rep, err := acts.Place(o)
	if err != nil {
		return finish(stdout, stderr, cwd, o.Goal, "", err)
	}
	text := fmt.Sprintf("placed %s %s @%d\nhash: %s\nH0: %s (exit %d)\n%s is %s\n",
		rep.Goal, rep.Key, rep.Version, rep.Hash, rep.Verdict, rep.Exit, rep.Goal, rep.State)
	return finish(stdout, stderr, cwd, rep.Goal, text, nil)
}

func runCmd(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.RunOpts
	fs.StringVar(&o.Actor, "actor", os.Getenv("HYP_ACTOR"), "")
	fs.StringVar(&o.Role, "role", os.Getenv("HYP_ROLE"), "")
	fs.StringVar(&o.Why, "why", "", "")
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return herr.BadArgs("run: %v", err)
	}
	if len(positional) != 2 {
		return herr.BadArgs("run takes COMMIT and <Goal>/<line>, for example: hyp run 4df2b9e G-NXHHVA/line-3; got %d arguments", len(positional))
	}
	o.Commit, o.Line = positional[0], positional[1]
	if n := len([]rune(o.Why)); n > logstore.MaxWhy {
		return herr.BadArgs("--why is %d runes; the limit is %d", n, logstore.MaxWhy)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return herr.Env("working directory: %v", err)
	}
	o.Dir = cwd
	rep, err := acts.Run(o)
	goal := ""
	if i := strings.LastIndex(o.Line, "/"); i > 0 {
		goal = o.Line[:i]
	}
	return finish(stdout, stderr, cwd, goal, rep.String(), err)
}

func selectCmd(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("select", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.SelectOpts
	fs.StringVar(&o.Actor, "actor", os.Getenv("HYP_ACTOR"), "")
	fs.StringVar(&o.Role, "role", os.Getenv("HYP_ROLE"), "")
	fs.StringVar(&o.Why, "why", "", "")
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return herr.BadArgs("select: %v", err)
	}
	if len(positional) != 2 {
		return herr.BadArgs("select takes a Goal and a row tree, for example: hyp select G-NXHHVA 4df2b9e...; got %d arguments", len(positional))
	}
	o.Goal, o.Tree = positional[0], positional[1]
	if n := len([]rune(o.Why)); n > logstore.MaxWhy {
		return herr.BadArgs("--why is %d runes; the limit is %d", n, logstore.MaxWhy)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return herr.Env("working directory: %v", err)
	}
	o.Dir = cwd
	rep, err := acts.Select(o)
	if err != nil {
		return finish(stdout, stderr, cwd, o.Goal, "", err)
	}
	// the report is the PR body, so the block does not print it again
	return finishWith(stdout, stderr, cwd, BlockOpts{Goal: rep.Goal, NoBody: true}, rep.String(), nil)
}

func statusCmd(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var role string
	fs.StringVar(&role, "role", "", "")
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return herr.BadArgs("status: %v", err)
	}
	if len(positional) > 1 {
		return herr.BadArgs("status takes at most one Goal, got %d", len(positional))
	}
	if role != "" && !acts.ValidRole(role) {
		return herr.BadArgs("role %q must be one of %s", role, strings.Join(acts.Roles, ", "))
	}
	o := BlockOpts{Role: role}
	if len(positional) == 1 {
		o.Goal = positional[0]
	}
	cwd, err := os.Getwd()
	if err != nil {
		return herr.Env("working directory: %v", err)
	}
	block, err := blockFor(cwd, o)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, block)
	return nil
}

// printed marks a refusal that finish already wrote, so Run does not print it
// a second time.
type printed struct{ err error }

func (p *printed) Error() string { return p.err.Error() }
func (p *printed) Unwrap() error { return p.err }

// blockFor reads the store under dir and prints its state block. A Goal named
// in o that the store does not know is a bad argument; no store is a refusal.
func blockFor(dir string, o BlockOpts) (string, error) {
	root, err := logstore.TopLevel(dir)
	if err != nil {
		return "", err
	}
	if err := logstore.Find(root); err != nil {
		return "", err
	}
	evs, _, err := logstore.Events(root)
	if err != nil {
		return "", err
	}
	t := tables.Replay(evs)
	if o.Goal != "" {
		g, err := t.Find(o.Goal)
		if err != nil {
			return "", err
		}
		o.Goal = g.EventID
	}
	return StateBlock(t, root, o), nil
}

// finish prints what a write act returned: its report, or on a refusal the
// refusal (to stderr), and then the state block of the Goal it was about. Bad arguments
// and an environment problem print no block.
func finish(stdout, stderr io.Writer, dir, goal, report string, err error) error {
	return finishWith(stdout, stderr, dir, BlockOpts{Goal: goal}, report, err)
}

func finishWith(stdout, stderr io.Writer, dir string, o BlockOpts, report string, err error) error {
	if err != nil && herr.KindOf(err) != herr.Refusal {
		return err
	}
	if err != nil {
		// the refusal is the report: it names the guard and carries what the
		// guard showed
		fmt.Fprintf(stderr, "hyp: %v\n", err)
	} else {
		fmt.Fprint(stdout, report)
	}
	if o.Goal != "" {
		block, berr := blockFor(dir, o)
		switch {
		case berr == nil:
			fmt.Fprint(stdout, block)
		case err == nil:
			return herr.Env("state block: %v", berr)
		}
	}
	if err != nil {
		return &printed{err: err}
	}
	return nil
}
