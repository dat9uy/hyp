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
	"regexp"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/acts"
	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/spec"
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

// usageHead and usageTail frame the commands' help blocks in hyp help; each
// block is the command's entry in the spec (internal/spec).
const usageHead = `usage: hyp <command> [flags]

Every command but help, spec and version needs an actor (--actor or
$HYP_ACTOR). An act that writes also needs a role (--role or $HYP_ROLE):
owner, grill, probe-writer or runner. hyp records what it prints to each actor
in the trace, one folder per clone under ~/.hyp/traces; hyp show prints it.
hyp writes nothing under ~/.local and does not read $XDG_STATE_HOME.

commands:
`

// usage is the text of hyp help: the help block of every command's entry,
// then the exit codes' entries.
var usage = func() string {
	var b strings.Builder
	b.WriteString(usageHead)
	for _, c := range spec.Commands {
		b.WriteString(c.Help)
	}
	b.WriteString("\nexit codes:")
	for i, e := range spec.Exits {
		sep := ","
		if i == len(spec.Exits)-1 {
			sep = ""
		}
		fmt.Fprintf(&b, " %d %s%s", e.Code, e.Name, sep)
	}
	b.WriteString("\n")
	return b.String()
}()

// commands are the top-level command names, sorted: the first word of each
// command's entry in the spec.
var commands = func() []string {
	var out []string
	for _, c := range spec.Commands {
		name, _, _ := strings.Cut(c.Name, " ")
		out = append(out, name)
	}
	slices.Sort(out)
	return slices.Compact(out)
}()

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
	return spec.ExitCode(fault.KindOf(err))
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
	case "spec":
		if len(args) > 1 {
			fmt.Fprint(stderr, usage)
			return fault.BadArg(codeSpecArgs, len(args)-1)
		}
		fmt.Fprint(stdout, specText())
		return nil
	case "place":
		return place(c, args[1:], stdout, stderr)
	case "run":
		return runCmd(c, args[1:], stdout, stderr)
	case "select":
		return selectCmd(c, args[1:], stdout, stderr)
	case "land":
		return landCmd(c, args[1:], stdout, stderr)
	case "unselect":
		return unselectCmd(c, args[1:], stdout, stderr)
	case "close":
		return closeCmd(c, args[1:], stdout, stderr)
	case "status":
		return statusCmd(c, args[1:], stdout)
	case "try":
		return tryCmd(c, args[1:], stdout, stderr)
	case "show":
		return showCmd(c, args[1:], stdout)
	case "prune":
		return pruneCmd(c, args[1:], stdout)
	case "finding":
		if len(args) > 1 {
			switch args[1] {
			case "add", "refute", "decline":
				return findingCmd(c, args[1], args[2:], stdout, stderr)
			case "list":
				return findingListCmd(c, args[2:], stdout)
			}
		}
		fmt.Fprint(stderr, usage)
		return fault.BadArg(codeFindingWhat)
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

// listFlag collects a repeated flag's values in order.
type listFlag []string

func (l *listFlag) String() string { return "" }

func (l *listFlag) Set(s string) error {
	*l = append(*l, s)
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
	var names listFlag
	actorFlags(fs, &o.Actor, &o.Role, &o.Why)
	fs.Var(&names, "finding", "")
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
	o.Lines, o.Findings = lines, names
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
	if len(positional) != 1 {
		return fault.BadArg(codePlaceArgs, len(positional))
	}
	o.Goal = positional[0]
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

// landCmd records the owner's land. Its report names the stop it ended; the
// state block of the landed Goal below it prints the record of the decision.
func landCmd(c *call, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("land", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.LandOpts
	actorFlags(fs, &o.Actor, &o.Role, &o.Why)
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return flagsFailed("land", err)
	}
	c.declare(o.Actor, o.Role)
	if len(positional) != 1 {
		return fault.BadArg(codeLandArgs, len(positional))
	}
	o.Goal = positional[0]
	if err := checkWhy(o.Why); err != nil {
		return err
	}
	o.Dir = c.dir
	rep, err := acts.Land(o)
	if err != nil {
		return finish(stdout, stderr, c.dir, o.Goal, "", err)
	}
	return finish(stdout, stderr, c.dir, rep.Goal, landText(rep), nil)
}

// readFlags parses the flags of a read: --actor, read from $HYP_ACTOR when
// absent, and --role, read from $HYP_ROLE. A read needs an actor, because
// hyp records what it prints to that actor; the role is recorded when it is
// one of the four. filter is the role --role named on the command line, for
// status to filter its allowed acts by.
func readFlags(c *call, command string, args []string) (positional []string, filter string, err error) {
	return readFlagsWith(c, command, args, flag.NewFlagSet(command, flag.ContinueOnError))
}

// readFlagsWith is readFlags over fs, which may hold flags of the command's
// own.
func readFlagsWith(c *call, command string, args []string, fs *flag.FlagSet) (positional []string, filter string, err error) {
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
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	var all bool
	fs.BoolVar(&all, "all", false, "")
	positional, role, err := readFlagsWith(c, "status", args, fs)
	if err != nil {
		return err
	}
	if len(positional) > 1 {
		return fault.BadArg(codeStatusArgs, len(positional))
	}
	o := BlockOpts{Role: role, All: all}
	if len(positional) == 1 {
		o.Goal, o.Ask = positional[0], true
	}
	block, err := blockFor(c.dir, o)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, block)
	return nil
}

// tryCmd runs one example as place and run would, and records nothing. It
// is a read: it needs an actor, whose trace it leaves, checks no guard, and
// prints its report and the Goal's state block.
func tryCmd(c *call, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("try", flag.ContinueOnError)
	var o acts.TryOpts
	fs.BoolVar(&o.Worktree, "worktree", false, "")
	fs.StringVar(&o.Commit, "commit", "", "")
	positional, role, err := readFlagsWith(c, "try", args, fs)
	if err != nil {
		return err
	}
	if len(positional) != 2 {
		return fault.BadArg(codeTryArgs, len(positional))
	}
	o.Goal, o.Key, o.Dir = positional[0], positional[1], c.dir
	rep, err := acts.Try(o)
	if err != nil {
		return finishWith(stdout, stderr, c.dir, BlockOpts{Goal: o.Goal, Role: role}, "", err)
	}
	return finishWith(stdout, stderr, c.dir, BlockOpts{Goal: rep.Goal, Role: role}, tryText(rep), nil)
}

// goalIDRe matches a Goal id: G- and the last six characters of its
// goal.add event id, a ULID.
var goalIDRe = regexp.MustCompile(`^G-[0-9A-HJKMNP-TV-Z]{6}$`)

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
		if goalIDRe.MatchString(actor) {
			return fault.BadArg(codeShowGoal, actor)
		}
		return fault.BadArg(codeShowActor, actor)
	}
	t, root, err := readStore(c.dir)
	if err != nil {
		return err
	}
	reads, err := readsOf(root, actor, t)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, showText(root, actor, reads))
	return nil
}

// pruneCmd lists the trace folders of clones that no longer exist, and with
// --remove removes them. It reads them from the traces folder, not from the
// store; like every command but add goal it needs a store to run from, and
// like a read it needs an actor, whose trace it leaves.
func pruneCmd(c *call, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("prune", flag.ContinueOnError)
	var remove bool
	fs.BoolVar(&remove, "remove", false, "")
	positional, _, err := readFlagsWith(c, "prune", args, fs)
	if err != nil {
		return err
	}
	if len(positional) != 0 {
		return fault.BadArg(codePruneArgs, len(positional))
	}
	if _, _, err := readStore(c.dir); err != nil {
		return err
	}
	base, gone, err := logstore.GoneTraces()
	if err != nil {
		return err
	}
	if !remove {
		fmt.Fprint(stdout, pruneText(base, gone, false))
		return nil
	}
	removed, err := logstore.RemoveGone(base, gone)
	fmt.Fprint(stdout, pruneText(base, removed, true))
	return err
}

// printed marks a refusal that finish already wrote, so Run does not print it
// a second time.
type printed struct{ err error }

func (p *printed) Error() string { return p.err.Error() }
func (p *printed) Unwrap() error { return p.err }

// readStore reads the store under dir and calculates its tables. No store is
// a refusal.
func readStore(dir string) (tables.Tables, string, error) {
	t, _, root, err := readStoreLines(dir)
	return t, root, err
}

// readStoreLines is readStore that also returns the lines of the events files
// that replay could not read, such as a torn last line.
func readStoreLines(dir string) (tables.Tables, []logstore.LineError, string, error) {
	root, err := logstore.TopLevel(dir)
	if err != nil {
		return tables.Tables{}, nil, "", err
	}
	if err := logstore.Find(root); err != nil {
		return tables.Tables{}, nil, "", err
	}
	t, unread, err := acts.ReadTables(root)
	if err != nil {
		return tables.Tables{}, nil, "", err
	}
	return t, unread, root, nil
}

// blockFor reads the store under dir and prints its state block. A Goal named
// in o that the store does not know is a bad argument; no store is a refusal.
func blockFor(dir string, o BlockOpts) (string, error) {
	t, unread, root, err := readStoreLines(dir)
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
	return StateBlock(t, unread, root, o), nil
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

// closeCmd records the owner's close. Its report names the Goal; the state
// block of the closed Goal below it prints the record of the decision.
func closeCmd(c *call, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("close", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.CloseOpts
	actorFlags(fs, &o.Actor, &o.Role, &o.Why)
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return flagsFailed("close", err)
	}
	c.declare(o.Actor, o.Role)
	if len(positional) != 1 {
		return fault.BadArg(codeCloseArgs, len(positional))
	}
	o.Goal = positional[0]
	if err := checkWhy(o.Why); err != nil {
		return err
	}
	o.Dir = c.dir
	rep, err := acts.Close(o)
	if err != nil {
		return finish(stdout, stderr, c.dir, o.Goal, "", err)
	}
	return finish(stdout, stderr, c.dir, rep.Goal, closeText(rep), nil)
}

// unselectCmd records the owner's unselect. Its report names the select it
// undid; the state block below it shows the Goal open again.
func unselectCmd(c *call, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("unselect", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o acts.UnselectOpts
	actorFlags(fs, &o.Actor, &o.Role, &o.Why)
	positional, err := parseInterleaved(fs, args)
	if err != nil {
		return flagsFailed("unselect", err)
	}
	c.declare(o.Actor, o.Role)
	if len(positional) != 1 {
		return fault.BadArg(codeUnselectArgs, len(positional))
	}
	o.Goal = positional[0]
	if err := checkWhy(o.Why); err != nil {
		return err
	}
	o.Dir = c.dir
	rep, err := acts.Unselect(o)
	if err != nil {
		return finish(stdout, stderr, c.dir, o.Goal, "", err)
	}
	return finish(stdout, stderr, c.dir, rep.Goal, unselectText(rep), nil)
}

// findingCmd runs an act on a finding: finding add, refute or decline. Any
// actor runs it, with any role or none, but only the owner declines; it
// writes its event to the clone's trace. Its report names each finding with
// its state.
func findingCmd(c *call, act string, args []string, stdout, stderr io.Writer) error {
	command := "finding " + act
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	var o acts.FindingOpts
	fs.StringVar(&o.Why, "why", "", "")
	switch act {
	case "add":
		fs.StringVar(&o.Fact, "fact", "", "")
		fs.StringVar(&o.Evidence, "evidence", "", "")
		fs.StringVar(&o.Draft, "draft", "", "")
		fs.StringVar(&o.Example, "example", "", "")
		fs.StringVar(&o.Commit, "commit", "", "")
	}
	positional, _, err := readFlagsWith(c, command, args, fs)
	if err != nil {
		return err
	}
	o.Actor, o.Role, o.Dir = c.actor, c.role, c.dir
	if err := checkWhy(o.Why); err != nil {
		return err
	}
	var rep acts.FindingReport
	switch act {
	case "add":
		if len(positional) != 0 {
			return fault.BadArg(codeFindingArgs, command, len(positional))
		}
		rep, err = acts.FindingAdd(o)
	case "refute":
		if len(positional) != 1 {
			return fault.BadArg(codeFindingArgs, command, len(positional))
		}
		o.Findings = positional
		rep, err = acts.FindingRefute(o)
	case "decline":
		if len(positional) != 1 {
			return fault.BadArg(codeFindingArgs, command, len(positional))
		}
		o.Findings = positional
		rep, err = acts.FindingDecline(o)
	}
	if err != nil {
		return finish(stdout, stderr, c.dir, "", "", err)
	}
	return finish(stdout, stderr, c.dir, "", findingText(rep), nil)
}

// findingListCmd is hyp finding list: a read that prints every finding with
// its state, fact, evidence and draft, and the acts each allows as ready
// lines.
func findingListCmd(c *call, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("finding list", flag.ContinueOnError)
	var o ListOpts
	fs.BoolVar(&o.All, "all", false, "")
	fs.BoolVar(&o.NotFolded, "not-folded", false, "")
	positional, _, err := readFlagsWith(c, "finding list", args, fs)
	if err != nil {
		return err
	}
	if len(positional) != 0 {
		return fault.BadArg(codeFindingArgs, "finding list", len(positional))
	}
	t, _, root, err := readStoreLines(c.dir)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, FindingList(t, root, o))
	return nil
}
