// Package spec is the one spec of hyp: every command with its usage, roles
// and guards in order, every state with the acts it allows and the acts it
// stops, every guard with its rule and its codes, every exit code, and the
// rules of the budget, the example file and the Probe. The event kinds are
// in internal/events (events.Kinds), which the store reads.
//
// hyp acts on these entries: hyp help prints the commands' usage from them,
// the channel maps a failure's kind to the exit code its entry gives, and
// hyp status offers an act only in a state whose entry allows it. hyp spec
// prints every entry, so changing a rule changes one entry. Each entry holds
// its own words, which the channel prints as they are.
package spec

import (
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/fault"
)

// The acts a ready line names, and the acts of a state's entry.
const (
	ActAddGoal  = "add goal"
	ActPlace    = "place"
	ActRun      = "run"
	ActSelect   = "select"
	ActLand     = "land"
	ActUnselect = "unselect"
	ActClose    = "close"
)

// The states of a Goal.
const (
	StateWritten  = "written"
	StateOpen     = "open"
	StateSelected = "selected"
	StateEnded    = "ended"
)

// Command is the entry of one command: an act that writes an event, or a read.
type Command struct {
	Name   string
	Help   string        // its block in hyp help, as printed; its first line starts with the usage
	Roles  []string      // the roles that may run it; empty for a read, which any role runs
	Guards []fault.Guard // the guards it checks, in order; a refusal names the first that fails
	Writes string        // the event it writes
	Bare   bool          // a read that needs no actor and leaves no trace
}

// Usage is the command's usage line: the first line of its help block, up to
// the run of spaces that starts the description.
func (c Command) Usage() string {
	line, _, _ := strings.Cut(c.Help, "\n")
	line = strings.TrimSpace(line)
	if i := strings.Index(line, "  "); i >= 0 {
		line = line[:i]
	}
	return line
}

// Commands are the entries of every command, in the order hyp help prints
// them.
var Commands = []Command{
	{Name: "add goal", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardDraft, fault.GuardWholeLines, fault.GuardFromEnded}, Writes: "goal.add; a new Goal is in no state before it",
		Help: `  add goal [G] --base COMMIT --words TEXT --line gap:TEXT [--line guard:TEXT ...]
           [--from G] [--why TEXT]
                                         record a Goal; with G, replace a Goal
                                         that is still written
`},
	{Name: "place", Roles: []string{"grill"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardSelected, fault.GuardGrillPlaces, fault.GuardWholeSet,
			fault.GuardCommitted, fault.GuardAddsOnly, fault.GuardKind}, Writes: "examples.place",
		Help: `  place G [--why TEXT]                   place the whole example set of Goal G:
                                         every folder goals/G/LINE/EXAMPLE, each
                                         run on H0; it places the set only when
                                         every example agrees with its line, and
                                         otherwise places none and names each
                                         cause
`},
	{Name: "run", Roles: []string{"runner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardSelected, fault.GuardBytes, fault.GuardOpen,
			fault.GuardOwnPath, fault.GuardLeafOnly, fault.GuardBudget, fault.GuardStable}, Writes: "result.record",
		Help: `  run COMMIT G/LINE [--why TEXT]         run every placed example of a line of
                                         Goal G on COMMIT and record the Result
`},
	{Name: "try",
		Help: `  try G LINE/EXAMPLE [--worktree]        run the example folder goals/G/LINE/EXAMPLE
                                         as it is on disk, on H0, or with
                                         --worktree on this checkout's files
                                         with what is not committed, and print
                                         its output, stderr, exit code and
                                         verdict as place and run would record
                                         them; records nothing and uses no run
`},
	{Name: "select", Roles: []string{"runner", "owner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardSelected, fault.GuardCandidate, fault.GuardNewCell,
			fault.GuardGoalsActor, fault.GuardSpec, fault.GuardOwnPath}, Writes: "select",
		Help: `  select G HASH [--why TEXT]             select the candidate row of Goal G whose
                                         subject hash is HASH; the report is the
                                         PR body
`},
	{Name: "land", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardOwnerLands, fault.GuardGoalsActor, fault.GuardStop}, Writes: "land",
		Help: `  land G [--why TEXT]                    the owner lands Goal G at its stop: the
                                         Goal ends, and its events up to the
                                         land are the snapshot that a later
                                         Goal's --from cites
`},
	{Name: "close", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardOwnerClose, fault.GuardGoalsActor}, Writes: "close",
		Help: `  close G [--why TEXT]                   the owner closes Goal G at any stop or
                                         none: the Goal ends as a land ends it,
                                         and a later Goal's --from cites it
`},
	{Name: "unselect", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardOwnerUnsel, fault.GuardGoalsActor, fault.GuardSelected}, Writes: "unselect",
		Help: `  unselect G [--why TEXT]                the owner unselects Goal G: G is open
                                         again with its rows and cells, its
                                         budget starts again, and the runner
                                         selects again only after a run that
                                         records a new cell
`},
	{Name: "status",
		Help: `  status [G] [--role ROLE]               print the state of every Goal, or of G:
                                         its stop, its waits and the allowed
                                         acts as ready command lines; of a
                                         written or open G, also its words,
                                         lines and placed examples
`},
	{Name: "show",
		Help: `  show ACTOR                             print every output hyp printed to
                                         ACTOR, byte for byte, oldest first
`},
	{Name: "prune",
		Help: `  prune [--remove]                       list the traces under ~/.hyp/traces of
                                         repositories that no longer exist;
                                         with --remove, remove those traces
`},
	{Name: "spec", Bare: true,
		Help: `  spec                                   print the spec: every event kind, state,
                                         act, guard and exit code, and the rules
                                         hyp acts on
`},
	{Name: "help", Bare: true,
		Help: `  help                                   print this text
`},
	{Name: "version", Bare: true,
		Help: `  version                                print the hyp version
`},
}

// Stop is one act a state does not allow, with the guard that refuses it.
type Stop struct {
	Act   string
	When  string // the case of the act it stops; empty for every case
	Guard fault.Guard
}

// State is the entry of one state of a Goal.
type State struct {
	Name   string
	Holds  string
	Allows []string // the acts hyp status may offer in it
	Stops  []Stop
}

// States are the entries of the four states, in the order a Goal moves
// through them. ended wins over selected, and selected over open.
var States = []State{
	{Name: StateWritten, Holds: "a goal.add exists, and the Goal is not open",
		Allows: []string{ActAddGoal, ActPlace, ActRun, ActClose},
		Stops: []Stop{
			{ActRun, "on a commit that is not H0", fault.GuardOpen},
			{ActSelect, "no row is met yet", fault.GuardCandidate},
			{ActLand, "a written Goal has no stop", fault.GuardStop},
			{ActUnselect, "the Goal has no select", fault.GuardSelected},
		}},
	{Name: StateOpen, Holds: "every line holds a placed example, and every current example gives on H0 the verdict its line's kind needs",
		Allows: []string{ActPlace, ActRun, ActSelect, ActLand, ActClose},
		Stops: []Stop{
			{ActAddGoal, "a replace", fault.GuardDraft},
			{ActSelect, "a row that is not a candidate", fault.GuardCandidate},
			{ActSelect, "the runner's, after an unselect with no new cell since", fault.GuardNewCell},
			{ActLand, "while the Goal has no stop", fault.GuardStop},
			{ActUnselect, "the Goal has no select", fault.GuardSelected},
		}},
	{Name: StateSelected, Holds: "the Goal has a select that no unselect undid, and no end",
		Allows: []string{ActLand, ActUnselect, ActClose},
		Stops: []Stop{
			{ActAddGoal, "a replace", fault.GuardDraft},
			{ActAddGoal, "--from G", fault.GuardFromEnded},
			{ActPlace, "", fault.GuardSelected},
			{ActRun, "", fault.GuardSelected},
			{ActSelect, "", fault.GuardSelected},
		}},
	{Name: StateEnded, Holds: "the Goal has a land or a close",
		Stops: []Stop{
			{ActAddGoal, "a replace", fault.GuardDraft},
			{ActPlace, "", fault.GuardNotEnded},
			{ActRun, "", fault.GuardNotEnded},
			{ActSelect, "", fault.GuardNotEnded},
			{ActLand, "", fault.GuardNotEnded},
			{ActUnselect, "", fault.GuardNotEnded},
			{ActClose, "", fault.GuardNotEnded},
		}},
}

// StateOf returns the entry of a state.
func StateOf(name string) (State, bool) {
	i := slices.IndexFunc(States, func(s State) bool { return s.Name == name })
	if i < 0 {
		return State{}, false
	}
	return States[i], true
}

// Allows reports whether a state's entry allows an act: hyp status offers
// only those.
func Allows(state, act string) bool {
	s, ok := StateOf(state)
	return ok && slices.Contains(s.Allows, act)
}

// StatesAllowing lists the states whose entry allows act.
func StatesAllowing(act string) []string {
	var out []string
	for _, s := range States {
		if slices.Contains(s.Allows, act) {
			out = append(out, s.Name)
		}
	}
	return out
}

// Guard is the entry of one guard: its rule, and the codes of the refusals
// it gives.
type Guard struct {
	Name  fault.Guard
	Rule  string
	Codes []fault.Code
}

// Guards are the entries of every guard, and of the refusals no guard names.
var Guards = []Guard{
	{fault.GuardDraft, "before the Goal opens, add goal G may replace it; after, its words and lines are frozen; a new Goal always passes",
		[]fault.Code{fault.NoGoalToReplace, fault.NotWritten}},
	{fault.GuardWholeLines, "every line has a kind, gap or guard; the Goal has one to ten lines and at least one gap line",
		[]fault.Code{fault.TooManyLines, fault.BadLineKind, fault.NoGapLine, fault.PayloadInvalid}},
	{fault.GuardFromEnded, "add goal --from G cites only an ended Goal, landed or closed; checked by the act only, never at replay",
		[]fault.Code{fault.FromNotEnded}},
	{fault.GuardNotEnded, "no write after the owner's land or close ends the Goal; replay also stops an event of the Goal ahead of the end that the end's snapshot does not name",
		[]fault.Code{fault.Ended, fault.NotInSnapshot}},
	{fault.GuardSelected, "place, run and select need the Goal to have no select; unselect needs it to have one",
		[]fault.Code{fault.HasSelect, fault.NoSelect, fault.PayloadInvalid}},
	{fault.GuardGrillPlaces, "only the grill places an example",
		[]fault.Code{fault.NotGrill, fault.PayloadInvalid}},
	{fault.GuardWholeSet, "the set of a place is every folder goals/G/LINE/EXAMPLE; every line holds an example folder, every placed example keeps its folder, and every folder's name is an example name",
		[]fault.Code{fault.LineHoldsNoExample, fault.PlacedFolderGone, fault.NotAnExampleName}},
	{fault.GuardCommitted, "each example folder's bytes on disk equal its bytes committed at HEAD",
		[]fault.Code{fault.NotAFolder, fault.FolderOnDisk, fault.NotCommitted, fault.DiskNotCommitted, fault.DiffersAt, fault.SetChangedWhileProbe}},
	{fault.GuardAddsOnly, "a new version of an example only adds floor or ceiling sections; any other change is a new example key; a hash the key held before never comes back; a place that changes no example places nothing",
		[]fault.Code{fault.SetUnchanged, fault.KeyTwiceInSet, fault.AlreadyCurrent, fault.EarlierVersion, fault.SectionDiffers, fault.PathDiffers, fault.DropsFloor, fault.DropsCeiling, fault.AddsNothing}},
	{fault.GuardKind, "H0 gives each changed example the verdict its line's kind needs: a gap example fails, a guard example is met",
		[]fault.Code{fault.WrongVerdictOnH0, fault.H0MovedWhilePlacing}},
	{fault.GuardBytes, "the commit exists in this clone",
		[]fault.Code{fault.CommitNotHere, fault.NotACommit}},
	{fault.GuardOpen, "the Goal is open, unless the commit's subject hash is H0's; the examples do not change while the Probes run",
		[]fault.Code{fault.NothingPlaced, fault.NotOpen, fault.H0MovedWhileRun, fault.ExamplesChanged, fault.ExampleChanged}},
	{fault.GuardOwnPath, "the runner's own change of the commit writes no path under goals/X/ of another Goal X",
		[]fault.Code{fault.OtherGoalPath}},
	{fault.GuardLeafOnly, "a run is on H0, or on a commit that no row of the Goal descends from",
		[]fault.Code{fault.AncestorOfRow}},
	{fault.GuardBudget, "at most three Hypotheses run since the boundary, H0 never counted",
		[]fault.Code{fault.BudgetUsed}},
	{fault.GuardStable, "two runs on the same bytes give one verdict; a run it stops still records its Result, marked, which counts nothing",
		[]fault.Code{fault.OtherVerdict, fault.VerdictsDiffered, fault.PayloadInvalid}},
	{fault.GuardCandidate, "the row is a candidate, a leaf row met on every line; with two or more candidates only the owner selects",
		[]fault.Code{fault.RoleCannotSelect, fault.NoRow, fault.OwnerSelectsAmong, fault.RowNotLeaf, fault.RowNotMet, fault.RowNotMetOnGoal, fault.NoResultOnRow}},
	{fault.GuardNewCell, "after an unselect, the runner selects only once a run recorded a cell whose example version, or whose row, had not run before it; the owner selects at any time",
		[]fault.Code{fault.NoNewCell}},
	{fault.GuardGoalsActor, "only the actor that added the Goal selects for the owner, lands, closes or unselects it",
		[]fault.Code{fault.NotGoalsActor}},
	{fault.GuardSpec, "the selected commit holds a spec under goals/G/specs/; a model there names each rule of its check configs in a variant config, and holds a chart",
		[]fault.Code{fault.NoSpec, fault.RuleNoVariant, fault.NoChart}},
	{fault.GuardOwnerLands, "only the owner lands a Goal",
		[]fault.Code{fault.NotOwner}},
	{fault.GuardStop, "a land needs a stop: selected, two candidates or budget used",
		[]fault.Code{fault.NoStop, fault.StopNotHeld, fault.PayloadInvalid}},
	{fault.GuardOwnerClose, "only the owner closes a Goal, at any stop or none",
		[]fault.Code{fault.NotOwnerClose, fault.PayloadInvalid}},
	{fault.GuardOwnerUnsel, "only the owner unselects a Goal",
		[]fault.Code{fault.NotOwnerUnselect, fault.PayloadInvalid}},
	{fault.GuardOwnResult, "only the actor that ran a Result retracts it",
		[]fault.Code{fault.NotOwnResult}},
	{fault.NoGuard, "a file hyp reads breaks a rule of its form, or the act names what the store does not hold",
		[]fault.Code{fault.NoStore, fault.EventVersion, fault.EventType, fault.PayloadDecode, fault.PayloadInvalid, fault.TrailingData,
			fault.IDNotULID, fault.ActorNotFile, fault.NoIDAfter, fault.NoGoalToPlace, fault.NoGoalToRun, fault.NoGoalToSelect,
			fault.NoGoalToLand, fault.NoGoalToClose, fault.NoGoalToUnselect, fault.NoResultToRetract, fault.NoExampleTxtar,
			fault.TextBeforeFirst, fault.SecondSection, fault.ExternalNotName, fault.InputOutside, fault.InputTwice,
			fault.UnknownSection, fault.NoSection, fault.Submodule, fault.EntryLeaves, fault.NotFileOrDir, fault.ReadFolder}},
}

// Exit is the entry of one exit code.
type Exit struct {
	Code int
	Kind fault.Kind // the kind of failure it is for; 0 for success
	Name string
	Text string
}

// Exits are the entries of the four exit codes.
var Exits = []Exit{
	{0, 0, "recorded or printed", "the act recorded its event, or the read printed"},
	{1, fault.Refusal, "refusal", "a rule of hyp refuses the act, or a file hyp reads breaks a rule; the refusal names its guard"},
	{2, fault.BadArguments, "bad argument", "the command line is wrong"},
	{3, fault.Environment, "environment problem", "I/O, a missing tool, or a Probe that failed or timed out"},
}

// ExitCode is the exit code of a failure of kind k: the code of its entry.
// A kind with no entry is an environment problem.
func ExitCode(k fault.Kind) int {
	for _, e := range Exits {
		if e.Kind == k && k != 0 {
			return e.Code
		}
	}
	return 3
}

// Rule is one rule hyp acts on that is no single event, state, act or guard.
type Rule struct {
	Name string
	Text []string
}

// Rules are the rules of the store, the budget, the example file and the
// Probe.
var Rules = []Rule{
	{"the store", []string{
		"The store is .tree/ at the git top level: .tree/events/<actor>.jsonl, one file per actor, appended only, and .tree/outputs/<sha256>.",
		"The first add goal creates it; .tree/.gitattributes merges the events files by union.",
		"The trace is $HOME/.hyp/traces/<sha256 of the clone's git common directory>/: every command that names an actor records there what hyp printed to it, before it prints.",
	}},
	{"merge order and replay", []string{
		"Replay reads every event by id, then by actor, then by line bytes, and calculates every table and state on every command; no state is stored.",
		"Replay checks each event's guards again at its place; an event that fails them is listed under refused at replay and never deleted.",
		"An end names its snapshot, the Goal's events replay applied when the owner ended it; an event of a kind an end freezes that the snapshot does not name is stopped by Not ended.",
	}},
	{"ids, actors and roles", []string{
		"An event id is a ULID; a Goal id is G- and the last six characters of its goal.add id, a Result id R- and six.",
		"An actor matches ^[a-z0-9][a-z0-9-]{0,39}$ and names its events file; a write act needs a role: owner, grill, probe-writer or runner.",
	}},
	{"the subject", []string{
		"A commit's subject is its tree without the top-level entries goals and .tree; its git tree hash is the subject hash, one Hypothesis, one row.",
		"H0 is the Goal's base commit; a row is a leaf when none of its commits has a descendant among the commits of the Goal's other rows.",
	}},
	{"budget and stops", []string{
		"The budget is three Hypotheses since the boundary, H0 never counted; a refused run costs nothing, a retracted Result still counts.",
		"The boundary is the latest of the latest place, the latest budget.reset and the latest unselect, else the Goal's start.",
		"The stops: selected (a select), two candidates (two or more candidates and no select), budget used (three Hypotheses counted, the third with a live cell for every example, and no candidate).",
		"A stop is calculated, never stored; it waits for the owner. At two candidates and budget used the state block adds the table of outputs.",
	}},
	{"the example file", []string{
		"An example is goals/G/LINE/EXAMPLE/ with example.txtar and probe/. Its sections: words (one), input/<path> (any), run (one), floor: contains and floor: equals (one or more floors), ceiling (any), external (at most one).",
		"One literal per section, byte for byte; a section body's final newline does not count.",
		"An example is met when the output holds every floor and contains no ceiling, else failed; a line is met on a row when every placed example of it is met there.",
	}},
	{"the Probe", []string{
		"hyp runs probe/run with empty stdin and a 10 minute timeout, in a fresh copy of the commit's subject ($HYP_SUBJECT), with $HYP_EXAMPLE a copy of the example folder and $HYP_RUN the run section.",
		"Only PATH, HOME and the names of the external section pass from the environment.",
		"The Probe prints the stdout of run, then a last line exit N; the output is the rest, with one final newline dropped. Its stderr is kept beside it and never compared.",
		"A Probe that exits non-zero, times out or prints no exit line records nothing: exit 3.",
	}},
}
