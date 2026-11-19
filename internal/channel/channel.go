// Package channel is how hyp talks to the owner and the agents: it parses the
// command line, runs the act, and words everything hyp prints, from a
// report and a state block to every refusal. No other package writes text
// for an agent: the other parts return values, and their failures carry a
// code and the values it names (internal/fault). The channel also keeps the
// trace of what it prints: before a command's output reaches the actor that
// ran it, hyp records those bytes, so hyp show prints exactly what that actor
// got. It is the only place a failure becomes an exit code.
package channel

import (
	"bytes"
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
	"github.com/dat9uy/hyp/internal/fault"
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

Every command but help and version needs an actor (--actor or $HYP_ACTOR).
An act that writes also needs a role (--role or $HYP_ROLE): owner, grill,
probe-writer or runner. hyp records what it prints to each actor; hyp show
prints it.

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
  show ACTOR                             print every output hyp printed to
                                         ACTOR, byte for byte, oldest first
  help                                   print this text
  version                                print the hyp version

exit codes: 0 recorded or printed, 1 refused by a rule, 2 bad arguments,
3 environment problem
`

// commands are the top-level command names, sorted.
var commands = []string{"add", "help", "place", "run", "select", "show", "status", "version"}

// Commands returns the sorted top-level command names.
func Commands() []string { return slices.Clone(commands) }

// call is one command as the channel runs it. A command sets actor once its
// flags parse and name a valid actor; the channel then records the trace of
// what it printed under that actor.
type call struct {
	dir   string // the working directory
	actor string
	role  string // the role the command declared, when it is one of the four
}

// declare sets the actor and the role whose trace the command leaves. An
// invalid actor leaves no trace: no events file can carry its name.
func (c *call) declare(actor, role string) {
	if logstore.ValidActor(actor) {
		c.actor = actor
	}
	if acts.ValidRole(role) {
		c.role = role
	}
}

// Run executes args (without the program name) and returns the exit code:
// 0 for success, 1 for a refusal, 2 for bad arguments, 3 for anything else.
// It writes nothing until the command ends: when the command named an actor
// and the repository has a store, hyp first records the trace of what the
// command printed, and only then prints it. A trace that cannot be recorded
// prints in place of the output, exit 3, so no actor gets bytes the trace
// does not hold.
func Run(args []string, stdout, stderr io.Writer) int {
	var out, errs bytes.Buffer
	c := &call{}
	if wd, err := os.Getwd(); err == nil {
		c.dir = wd
	}
	err := dispatch(c, args, &out, &errs)
	if err != nil {
		var done *printed
		if !errors.As(err, &done) {
			fmt.Fprintf(&errs, "hyp: %s\n", failText(err, ""))
		}
	}
	code := exitCode(err)
	if c.actor != "" && c.dir != "" {
		_, rerr := acts.RecordRead(acts.ReadOpts{
			Dir: c.dir, Actor: c.actor, Role: c.role, Command: slices.Clone(args),
			Exit: code, Stdout: out.Bytes(), Stderr: errs.Bytes(),
		})
		if rerr != nil {
			fmt.Fprintf(stderr, "hyp: %s\n", failText(fault.EnvFault(codeNotRecorded, rerr), ""))
			return 3
		}
	}
	stderr.Write(errs.Bytes())
	stdout.Write(out.Bytes())
	return code
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	switch fault.KindOf(err) {
	case fault.Refusal:
		return 1
	case fault.BadArguments:
		return 2
	}
	return 3
}

func dispatch(c *call, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return fault.BadArg(codeNoCommand)
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, versionLine())
		return nil
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	case "place":
		return place(c, args[1:], stdout, stderr)
	case "run":
		return runCmd(c, args[1:], stdout, stderr)
	case "select":
		return selectCmd(c, args[1:], stdout, stderr)
	case "status":
		return statusCmd(c, args[1:], stdout)
	case "show":
		return showCmd(c, args[1:], stdout)
	case "add":
		if len(args) > 1 && args[1] == "goal" {
			return addGoal(c, args[2:], stdout, stderr)
		}
		fmt.Fprint(stderr, usage)
		return fault.BadArg(codeAddWhat)
	}
	fmt.Fprint(stderr, usage)
	return fault.BadArg(codeUnknownCommand, args[0])
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
		return fault.BadArg(codeLineFlag, s)
	}
	*l = append(*l, acts.LineInput{Kind: kind, Text: text})
	return nil
}

// flagsFailed is the bad argument of a command whose flags do not parse.
func flagsFailed(command string, err error) error {
	if _, ok := fault.As(err); ok {
		return err
	}
	return fault.BadArg(codeFlags, command, err)
}

// actorFlags adds --actor and --role to fs, read from $HYP_ACTOR and
// $HYP_ROLE when absent, and --why when why is not nil.
func actorFlags(fs *flag.FlagSet, actor, role, why *string) {
	fs.StringVar(actor, "actor", os.Getenv("HYP_ACTOR"), "")
	fs.StringVar(role, "role", os.Getenv("HYP_ROLE"), "")
	if why != nil {
		fs.StringVar(why, "why", "", "")
	}
}

func checkWhy(why string) error {
	if n := len([]rune(why)); n > logstore.MaxWhy {
		return fault.BadArg(fault.WhyTooLong, n, logstore.MaxWhy)
	}
	return nil
}

func addGoal(c *call, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("add goal", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.AddGoalOpts
	var lines lineFlag
	actorFlags(fs, &o.Actor, &o.Role, &o.Why)
	fs.StringVar(&o.From, "from", "", "")
	fs.StringVar(&o.Words, "words", "", "")
	fs.StringVar(&o.Base, "base", "", "")
	fs.Var(&lines, "line", "")
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return flagsFailed("add goal", err)
	}
	c.declare(o.Actor, o.Role)
	if len(positional) > 1 {
		return fault.BadArg(codeReplaceTwice, len(positional))
	}
	if len(positional) == 1 {
		o.Goal = positional[0]
	}
	if o.Base == "" {
		return fault.BadArg(codeNeedsBase)
	}
	if o.Words == "" {
		return fault.BadArg(codeNeedsWords)
	}
	if err := checkWhy(o.Why); err != nil {
		return err
	}
	o.Lines = lines
	o.Dir = c.dir
	rep, err := acts.AddGoal(o)
	if err != nil {
		return finish(stdout, stderr, c.dir, o.Goal, "", err)
	}
	return finish(stdout, stderr, c.dir, rep.Goal, addGoalText(rep), nil)
}

func place(c *call, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("place", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.PlaceOpts
	actorFlags(fs, &o.Actor, &o.Role, &o.Why)
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return flagsFailed("place", err)
	}
	c.declare(o.Actor, o.Role)
	if len(positional) != 2 {
		return fault.BadArg(codePlaceArgs, len(positional))
	}
	o.Goal, o.Key = positional[0], positional[1]
	if err := checkWhy(o.Why); err != nil {
		return err
	}
	o.Dir = c.dir
	rep, err := acts.Place(o)
	if err != nil {
		return finish(stdout, stderr, c.dir, o.Goal, "", err)
	}
	return finish(stdout, stderr, c.dir, rep.Goal, placeText(rep), nil)
}

func runCmd(c *call, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.RunOpts
	actorFlags(fs, &o.Actor, &o.Role, &o.Why)
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return flagsFailed("run", err)
	}
	c.declare(o.Actor, o.Role)
	if len(positional) != 2 {
		return fault.BadArg(codeRunArgs, len(positional))
	}
	o.Commit, o.Line = positional[0], positional[1]
	if err := checkWhy(o.Why); err != nil {
		return err
	}
	o.Dir = c.dir
	rep, err := acts.Run(o)
	goal := ""
	if i := strings.LastIndex(o.Line, "/"); i > 0 {
		goal = o.Line[:i]
	}
	if err != nil {
		return finish(stdout, stderr, c.dir, goal, "", err)
	}
	return finish(stdout, stderr, c.dir, goal, runText(rep), nil)
}

func selectCmd(c *call, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("select", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.SelectOpts
	actorFlags(fs, &o.Actor, &o.Role, &o.Why)
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return flagsFailed("select", err)
	}
	c.declare(o.Actor, o.Role)
	if len(positional) != 2 {
		return fault.BadArg(codeSelectArgs, len(positional))
	}
	o.Goal, o.Tree = positional[0], positional[1]
	if err := checkWhy(o.Why); err != nil {
		return err
	}
	o.Dir = c.dir
	rep, err := acts.Select(o)
	if err != nil {
		return finish(stdout, stderr, c.dir, o.Goal, "", err)
	}
	root, err := logstore.TopLevel(c.dir)
	if err != nil {
		return err
	}
	// the report is the PR body, so the block does not print it again
	return finishWith(stdout, stderr, c.dir, BlockOpts{Goal: rep.Goal, NoBody: true}, prBody(root, rep.Selected), nil)
}

// readFlags parses the flags of a read: --actor, read from $HYP_ACTOR when
// absent, and --role, read from $HYP_ROLE. A read needs an actor, because
// hyp records what it prints to that actor; the role is recorded when it is
// one of the four. filter is the role --role named on the command line, for
// status to filter its allowed acts by.
func readFlags(c *call, command string, args []string) (positional []string, filter string, err error) {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var actor, role string
	fs.StringVar(&actor, "actor", os.Getenv("HYP_ACTOR"), "")
	fs.StringVar(&filter, "role", "", "")
	if positional, err = parseInterleaved(fs, args); err != nil {
		return nil, "", flagsFailed(command, err)
	}
	role = filter
	if role == "" {
		role = os.Getenv("HYP_ROLE")
	}
	switch {
	case actor == "":
		return nil, "", fault.BadArg(codeNeedsActor, command)
	case !logstore.ValidActor(actor):
		return nil, "", fault.BadArg(fault.BadActor, actor)
	}
	// from here the read is traced, a bad role included
	c.declare(actor, role)
	if filter != "" && !acts.ValidRole(filter) {
		return nil, "", fault.BadArg(fault.BadRole, filter, acts.Roles)
	}
	return positional, filter, nil
}

func statusCmd(c *call, args []string, stdout io.Writer) error {
	positional, role, err := readFlags(c, "status", args)
	if err != nil {
		return err
	}
	if len(positional) > 1 {
		return fault.BadArg(codeStatusArgs, len(positional))
	}
	o := BlockOpts{Role: role}
	if len(positional) == 1 {
		o.Goal = positional[0]
	}
	block, err := blockFor(c.dir, o)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, block)
	return nil
}

func showCmd(c *call, args []string, stdout io.Writer) error {
	positional, _, err := readFlags(c, "show", args)
	if err != nil {
		return err
	}
	if len(positional) != 1 {
		return fault.BadArg(codeShowArgs, len(positional))
	}
	actor := positional[0]
	if !logstore.ValidActor(actor) {
		return fault.BadArg(codeShowActor, actor)
	}
	t, root, err := readStore(c.dir)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, showText(root, actor, t.Reads[actor]))
	return nil
}

// printed marks a refusal that finish already wrote, so Run does not print it
// a second time.
type printed struct{ err error }

func (p *printed) Error() string { return p.err.Error() }
func (p *printed) Unwrap() error { return p.err }

// readStore reads the store under dir and calculates its tables. No store is
// a refusal.
func readStore(dir string) (tables.Tables, string, error) {
	root, err := logstore.TopLevel(dir)
	if err != nil {
		return tables.Tables{}, "", err
	}
	if err := logstore.Find(root); err != nil {
		return tables.Tables{}, "", err
	}
	evs, _, err := logstore.Events(root)
	if err != nil {
		return tables.Tables{}, "", err
	}
	return tables.Replay(evs), root, nil
}

// blockFor reads the store under dir and prints its state block. A Goal named
// in o that the store does not know is a bad argument; no store is a refusal.
func blockFor(dir string, o BlockOpts) (string, error) {
	t, root, err := readStore(dir)
	if err != nil {
		return "", err
	}
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
	if err != nil && fault.KindOf(err) != fault.Refusal {
		return err
	}
	if err != nil {
		// the refusal is the report: it names the guard and carries what the
		// guard showed
		root, _ := logstore.TopLevel(dir)
		fmt.Fprintf(stderr, "hyp: %s\n", failText(err, root))
	} else {
		fmt.Fprint(stdout, report)
	}
	if o.Goal != "" {
		block, berr := blockFor(dir, o)
		switch {
		case berr == nil:
			fmt.Fprint(stdout, block)
		case err == nil:
			return fault.EnvFault(codeStateBlock, berr)
		}
	}
	if err != nil {
		return &printed{err: err}
	}
	return nil
}
