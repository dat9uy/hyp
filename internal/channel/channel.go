// Package channel is hyp's command line: it parses arguments, runs the act and
// turns what comes back into output and an exit code. It is the only place an
// error kind becomes an exit code.
package channel

import (
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
  help                                   print this text
  version                                print the hyp version

exit codes: 0 recorded or printed, 1 refused by a rule, 2 bad arguments,
3 environment problem
`

// commands are the top-level command names, sorted.
var commands = []string{"add", "help", "place", "run", "version"}

// Commands returns the sorted top-level command names.
func Commands() []string { return slices.Clone(commands) }

// Run executes args (without the program name) and returns the exit code:
// 0 for success, 1 for a refusal, 2 for bad arguments, 3 for anything else.
func Run(args []string, stdout, stderr io.Writer) int {
	err := dispatch(args, stdout, stderr)
	if err == nil {
		return 0
	}
	fmt.Fprintf(stderr, "hyp: %v\n", err)
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
		return place(args[1:], stdout)
	case "run":
		return runCmd(args[1:], stdout)
	case "add":
		if len(args) > 1 && args[1] == "goal" {
			return addGoal(args[2:], stdout)
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

func addGoal(args []string, stdout io.Writer) error {
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
		return err
	}
	verb := "recorded"
	if rep.Replaced {
		verb = "replaced"
	}
	fmt.Fprintf(stdout, "%s %s\n", verb, rep.Goal)
	fmt.Fprintf(stdout, "words: %s\n", rep.Words)
	fmt.Fprintf(stdout, "H0: %s\n", rep.Base)
	for _, l := range rep.Lines {
		fmt.Fprintf(stdout, "%s %s: %s\n", l.Key, l.Kind, l.Text)
	}
	return nil
}

func place(args []string, stdout io.Writer) error {
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
		return err
	}
	fmt.Fprintf(stdout, "placed %s %s @%d\n", rep.Goal, rep.Key, rep.Version)
	fmt.Fprintf(stdout, "hash: %s\n", rep.Hash)
	fmt.Fprintf(stdout, "H0: %s (exit %d)\n", rep.Verdict, rep.Exit)
	fmt.Fprintf(stdout, "%s is %s\n", rep.Goal, rep.State)
	return nil
}

func runCmd(args []string, stdout io.Writer) error {
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
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, rep.String())
	return nil
}
