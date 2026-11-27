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

	ActFindingAdd     = "finding add"
	ActFindingRefute  = "finding refute"
	ActFindingDecline = "finding decline"
	ActFindingList    = "finding list"
)

// The states of a Goal.
const (
	StateWritten  = "written"
	StateOpen     = "open"
	StateSelected = "selected"
	StateEnded    = "ended"
)

// The states of a finding.
const (
	FindingNotFolded = "not folded"
	FindingFolded    = "folded"
	FindingSettled   = "settled"
	FindingEnded     = "ended"
	FindingRefuted   = "refuted"
	FindingDeclined  = "declined"
)

// Command is the entry of one command: an act that writes an event, or a read.
type Command struct {
	Name   string
	Help   string        // its block in hyp help, as printed; its first line starts with the usage
	Roles  []string      // the roles that may run it; empty for a read, which any role runs
	Guards []fault.Guard // the guards it checks, in order; a refusal names the first that fails
	Writes string        // the event it writes
	Bare   bool          // a read that needs no actor and leaves no trace
	Any    bool          // an act on a finding: any actor runs it, with any role or none
}

// Usage is the command's whole usage: the text of its help block before the
// description column, joined across the block's lines. A line indented to the
// description column holds only description and adds nothing.
func (c Command) Usage() string {
	var parts []string
	for _, line := range strings.Split(c.Help, "\n") {
		text := strings.TrimLeft(line, " ")
		if text == "" || len(line)-len(text) >= descColumn {
			continue
		}
		if i := strings.Index(text, "  "); i >= 0 {
			text = text[:i]
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, " ")
}

// descColumn is the column, counted from zero, where hyp help starts a
// command's description.
const descColumn = 41

// Commands are the entries of every command, in the order hyp help prints
// them.
var Commands = []Command{
	{Name: "add goal", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardDraft, fault.GuardWholeLines, fault.GuardFromEnded,
			fault.GuardNotRefuted, fault.GuardNotDeclined, fault.GuardOneInWork,
			fault.GuardWholeSet, fault.GuardMeetable, fault.GuardCovered, fault.GuardKind}, Writes: "goal.add; a new Goal is in no state before it",
		Help: `  add goal [G] --base COMMIT --words TEXT --line gap:TEXT [--line guard:TEXT ...]
           [--draft NAME] [--from G] [--finding F ...] [--why TEXT]
                                         record a Goal; with --draft, the
                                         example set of draft NAME in the
                                         clone's trace, LINE/EXAMPLE, read once
                                         and each run on H0: the add carries
                                         each example's hash and verdict, the
                                         Goal is open at once, the bytes it
                                         hashed are copied to goals/G/, and an
                                         example that does not agree with its
                                         line records no Goal;
                                         with G, replace a Goal that is still
                                         written; each --finding names a
                                         finding of this clone's trace that the
                                         Goal is about
`},
	{Name: "place", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardSelected, fault.GuardOwnerPlaces, fault.GuardGoalsActor, fault.GuardWholeSet,
			fault.GuardCommitted, fault.GuardAddsOnly, fault.GuardMeetable, fault.GuardNotRefuted, fault.GuardNotDeclined, fault.GuardOneInWork,
			fault.GuardCovered, fault.GuardKind}, Writes: "examples.place",
		Help: `  place G [--draft NAME] [--why TEXT]
                                         the owner places the whole example set
                                         of Goal G: every folder
                                         goals/G/LINE/EXAMPLE, or with --draft,
                                         the draft NAME of the clone's trace,
                                         LINE/EXAMPLE, read once and copied to
                                         goals/G/; each new or changed one run
                                         on H0; it places the
                                         set only when every example agrees with
                                         its line and each finding G names is
                                         named by an example, and otherwise
                                         places none and names each cause
`},
	{Name: "run", Roles: []string{"runner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardSelected, fault.GuardBytes, fault.GuardOpen,
			fault.GuardOwnPath, fault.GuardLeafOnly, fault.GuardBudget, fault.GuardStable}, Writes: "result.record",
		Help: `  run COMMIT G/LINE [--why TEXT]         run every placed example of a line of
                                         Goal G on COMMIT and record the Result
`},
	{Name: "try",
		Help: `  try G LINE/EXAMPLE [--worktree | --commit COMMIT]
  try --draft NAME --base COMMIT LINE/EXAMPLE
                                         run the example folder goals/G/LINE/EXAMPLE
                                         as it is on disk, on H0, on COMMIT, or
                                         with --worktree on this checkout's files
                                         with what is not committed, and print
                                         its output, stderr, exit code and
                                         verdict as place and run would record
                                         them; it uses no run, and saves the
                                         cell in this clone's trace only when
                                         the trace lacks it; with --draft,
                                         the example LINE/EXAMPLE of draft NAME
                                         of a Goal not added yet, on COMMIT as
                                         its H0, and it saves nothing
`},
	{Name: "draft",
		Help: `  draft path NAME                        print the folder of draft NAME in this
                                         clone's trace, drafts/NAME under
                                         ~/.hyp/traces; the grill writes a
                                         draft's examples there as
                                         LINE/EXAMPLE, and hyp try, add goal
                                         and place read them with --draft NAME;
                                         the folder is in no commit, and hyp
                                         creates nothing
`},
	{Name: "select", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardSelected, fault.GuardCandidate, fault.GuardNewCell,
			fault.GuardGoalsActor, fault.GuardSpec, fault.GuardOwnPath}, Writes: "select",
		Help: `  select G HASH [--why TEXT]             the owner selects one of two or more
                                         candidate rows of Goal G, the one whose
                                         subject hash is HASH; a lone candidate
                                         this clone's Results show is selected
                                         without it; the report is the PR body:
                                         the Goal's words, each line, each
                                         placed example's words, each chart and
                                         the selected commit and row, and no
                                         result of any example
`},
	{Name: "land", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardOwnerLands, fault.GuardGoalsActor, fault.GuardEvidence,
			fault.GuardSpec, fault.GuardOwnPath, fault.GuardStop}, Writes: "land",
		Help: `  land G [COMMIT] [--why TEXT]           the owner lands Goal G: on COMMIT, or
                                         on the commit G is selected on, the
                                         land carries the commit, its tree and
                                         each example's verdict there, running
                                         first each example this clone holds no
                                         Result of, and it is refused when one
                                         is not met; with no commit and no
                                         select, at its stop; the Goal ends,
                                         and its events up to the land are the
                                         snapshot that a later Goal's --from
                                         cites
`},
	{Name: "close", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardOwnerClose, fault.GuardGoalsActor}, Writes: "close",
		Help: `  close G [--why TEXT]                   the owner closes Goal G at any stop or
                                         none: the Goal ends as a land ends it,
                                         and a later Goal's --from cites it
`},
	{Name: "unselect", Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardNotEnded, fault.GuardOwnerUnsel, fault.GuardGoalsActor, fault.GuardSelected}, Writes: "unselect",
		Help: `  unselect G [COMMIT] [--why TEXT]       the owner unselects Goal G on the
                                         commit it is selected on, which COMMIT
                                         may name: the unselect names that
                                         commit, no clone shows G selected on
                                         its row again, and a land on it is
                                         refused; G is open again with its rows
                                         and cells, its budget starts again, and
                                         a select is calculated again only after
                                         a run that records a new cell
`},
	{Name: ActFindingAdd, Any: true,
		Guards: nil, Writes: "finding.add, in the trace",
		Help: `  finding add --fact TEXT --evidence TEXT [--example G/line-N/name --commit C]
              [--draft "hyp add goal ..."] [--why TEXT]
                                         record a finding: a fact an agent
                                         found and a pointer to its evidence,
                                         that names no Goal; --example and
                                         --commit, given together, name an
                                         example and a commit that holds it as
                                         evidence too, and a commit that does
                                         not hold it is refused; --draft
                                         attaches an add goal the agent
                                         drafted, which the owner may edit and
                                         run; it prints the finding's id
`},
	{Name: ActFindingRefute, Any: true,
		Guards: []fault.Guard{fault.GuardNotRefuted, fault.GuardNotDeclined}, Writes: "finding.refute, in the trace",
		Help: `  finding refute F [--why TEXT]          refute finding F: its fact is false;
                                         refuted is final
`},
	{Name: ActFindingDecline, Roles: []string{"owner"},
		Guards: []fault.Guard{fault.GuardOwnerDecl, fault.GuardNotRefuted, fault.GuardNotDeclined, fault.GuardNotFolded},
		Writes: "finding.decline, in the trace",
		Help: `  finding decline F --why TEXT           the owner declines finding F, which no
                                         Goal named: its fact is true and the
                                         owner does not want it fixed, for the
                                         why; declined is final
`},
	{Name: ActFindingList,
		Help: `  finding list [--all | --not-folded]    print every finding with its state,
                                         fact, evidence and draft, and the
                                         finding acts it allows as ready
                                         command lines. A finding whose
                                         evidence names an example prints its
                                         walk: each command of the example's
                                         run section with the step of the
                                         kept model, goals/model/, it takes on
                                         the commit the finding names, or no
                                         step; in a repository that keeps a
                                         model, a finding that names no
                                         example reads no walk. A finding settled or
                                         ended, refuted, or declined, is left
                                         out unless --all; --not-folded lists only
                                         the findings not folded
`},
	{Name: "status",
		Help: `  status [G] [--role ROLE] [--all]       print the state of every Goal in work,
                                         or of G: its stop, its waits and the
                                         allowed acts as ready command lines;
                                         of a written or open G, also its
                                         words, why, lines and placed examples,
                                         of an open G, unless --role names a
                                         role other than owner, the runner's
                                         Start prompt ready to copy; of an
                                         ended G, its words, why and lines and
                                         the why of its land or close; and of
                                         G the findings it names. For each
                                         candidate whose select hyp holds
                                         back, with any role, it prints the
                                         guard, its code and the fix that
                                         hyp select would print. Of a
                                         selected G, it lists each step and
                                         rule of the kept model, goals/model/,
                                         that the selected commit adds or
                                         changes against H0: a step by its
                                         command, a rule with the words of
                                         the Goal line its source names as
                                         G:line-N, or no line; no TLA+ text.
                                         Of G, for each line, each command of
                                         its placed examples' run sections
                                         with the step of the kept model it
                                         takes, or no step: on H0, or on the
                                         selected commit while G is selected;
                                         nothing when that commit keeps no
                                         model. With no G, it prints the words of each
                                         written or open Goal, leaves out the
                                         blocks of ended Goals with a line that
                                         counts them (--all prints them too),
                                         says no Goal is in work when none is,
                                         and ends with the count of findings
                                         not folded
`},
	{Name: "show",
		Help: `  show ACTOR                             print every output hyp printed to
                                         ACTOR, byte for byte, oldest first,
                                         then the step log of each of
                                         ACTOR's Results that keeps one;
                                         an argument that is not an actor
                                         name exits 2, and a Goal id names
                                         hyp status G
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
	{Name: StateWritten, Holds: "a goal.add exists that carries no example, and the Goal is not open",
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
			{ActLand, "with no commit, while the Goal has no stop", fault.GuardStop},
			{ActLand, "on a commit where an example is not met, or that an unselect rejected", fault.GuardEvidence},
			{ActUnselect, "the Goal has no select", fault.GuardSelected},
		}},
	{Name: StateSelected, Holds: "this clone's Results show one candidate row, met on every line, holding a spec, that no unselect rejected; or a select event that no unselect undid, of the older form or the owner's among two candidates; and no end",
		Allows: []string{ActPlace, ActLand, ActUnselect, ActClose},
		Stops: []Stop{
			{ActAddGoal, "a replace", fault.GuardDraft},
			{ActAddGoal, "--from G", fault.GuardFromEnded},
			{ActPlace, "while a select event holds the Goal", fault.GuardSelected},
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

// FindingStates are the entries of the six states of a finding. A
// finding's state follows from the Goals that name it and changes no Goal's
// state: its events are in the clone's trace, and a Goal lands, closes or
// ends with findings in any state.
var FindingStates = []State{
	{Name: FindingNotFolded, Holds: "a finding.add exists, no Goal names it, and no refute or decline names it; it waits for a Goal",
		Allows: []string{ActAddGoal, ActPlace, ActFindingRefute, ActFindingDecline}},
	{Name: FindingFolded, Holds: "a Goal in work names it, at add goal, in an example it placed, or by an old fold, and no refute does; it reads folded into G; the old form reads folded into #N",
		Allows: []string{ActFindingRefute},
		Stops: []Stop{
			{ActAddGoal, "of another Goal", fault.GuardOneInWork},
			{ActPlace, "of another Goal", fault.GuardOneInWork},
			{ActFindingDecline, "", fault.GuardNotFolded},
		}},
	{Name: FindingSettled, Holds: "no Goal in work names it, and a Goal that names it landed with a select; it reads settled in G",
		Allows: []string{ActAddGoal, ActPlace, ActFindingRefute},
		Stops: []Stop{
			{ActFindingDecline, "", fault.GuardNotFolded},
		}},
	{Name: FindingEnded, Holds: "every Goal that named it ended with no select: closed, or landed at two candidates or budget used; it reads ended in G, not settled, and does not wait",
		Allows: []string{ActAddGoal, ActPlace, ActFindingRefute},
		Stops: []Stop{
			{ActFindingDecline, "", fault.GuardNotFolded},
		}},
	{Name: FindingRefuted, Holds: "a refute names it: its fact is false; refuted is final",
		Stops: []Stop{
			{ActAddGoal, "", fault.GuardNotRefuted},
			{ActPlace, "", fault.GuardNotRefuted},
			{ActFindingRefute, "", fault.GuardNotRefuted},
			{ActFindingDecline, "", fault.GuardNotRefuted},
		}},
	{Name: FindingDeclined, Holds: "the owner's decline names it, with a why: its fact is true and the owner does not want it fixed; declined is final",
		Stops: []Stop{
			{ActAddGoal, "", fault.GuardNotDeclined},
			{ActPlace, "", fault.GuardNotDeclined},
			{ActFindingRefute, "", fault.GuardNotDeclined},
			{ActFindingDecline, "", fault.GuardNotDeclined},
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

// Guard is the entry of one guard: its rule, the codes of the refusals it
// gives, and the words on whether replay checks it. Replay checks a guard
// only when it reads only events stored in git, so every clone replays the
// same refusals; a guard that reads git objects, files, the Probes or the
// Results and cells of a clone's trace is checked by its act alone. Whether
// replay checks it is no word of the entry: the replay entries of the event
// kinds in internal/tables say it, and hyp spec prints yes or no from them.
type Guard struct {
	Name  fault.Guard
	Rule  string
	Codes []fault.Code
	// ReplayWhy follows the yes or no that hyp spec prints after checked at
	// replay, with its own leading separator: the part replay checks, or why
	// it checks none; empty for a plain yes.
	ReplayWhy string
}

// Guards are the entries of every guard, and of the refusals no guard names.
var Guards = []Guard{
	{fault.GuardDraft, "before the Goal opens, add goal G may replace it, with no example; after, its words and lines are frozen; a new Goal always passes",
		[]fault.Code{fault.NoGoalToReplace, fault.NotWritten, fault.ReplaceExamples}, ""},
	{fault.GuardWholeLines, "every line has a kind, gap or guard; the Goal has one to ten lines and at least one gap line",
		[]fault.Code{fault.TooManyLines, fault.BadLineKind, fault.NoGapLine}, ""},
	{fault.GuardFromEnded, "add goal --from G cites only an ended Goal, landed or closed; checked by the act only, never at replay",
		[]fault.Code{fault.FromNotEnded}, ": it is checked by the act only"},
	{fault.GuardNotEnded, "no write after the owner's land or close ends the Goal; replay also stops an event of the Goal ahead of the end that the end's snapshot does not name",
		[]fault.Code{fault.Ended, fault.NotInSnapshot}, ""},
	{fault.GuardSelected, "place, run and select need the Goal not selected; unselect needs it selected, on the commit it names, which no unselect rejected before",
		[]fault.Code{fault.HasSelect, fault.NoSelect, fault.NotSelectedOn, fault.AlreadyUnselected}, ", against a select event; the act reads the select this clone's Results show"},
	{fault.GuardOwnerPlaces, "only the owner places an example, with the Goal's actor; a place of the older form names the grill and replays as it was",
		[]fault.Code{fault.NotOwnerPlace}, ""},
	{fault.GuardWholeSet, "the set of an add goal --draft or a place is every folder LINE/EXAMPLE of the draft, goals/G/LINE/EXAMPLE by default; a draft is read once, and the bytes hashed, run on H0 and copied to goals/G/ are that one read; every line holds an example folder, every placed example keeps its folder, and every folder's name is an example name",
		[]fault.Code{fault.LineHoldsNoExample, fault.PlacedFolderGone, fault.NotAnExampleName, fault.ExampleOffLine, fault.NotAFolder, fault.FolderOnDisk}, ", that each example an add goal carries is on a line of the Goal; the act reads the example folders"},
	{fault.GuardCommitted, "each example folder's bytes on disk equal its bytes committed at HEAD; a draft of the clone's trace is in no commit, so a place --draft does not check it, and an add goal --draft has no such set",
		[]fault.Code{fault.NotAFolder, fault.FolderOnDisk, fault.NotCommitted, fault.DiskNotCommitted, fault.DiffersAt, fault.SetChangedWhileProbe}, ": it reads the example folders and git"},
	{fault.GuardAddsOnly, "a new version of an example only adds floor or ceiling sections; any other change is a new example key; a hash the key held before never comes back, except the current one measured again on H0 after a replace moved H0; a place that changes no example places nothing",
		[]fault.Code{fault.SetUnchanged, fault.KeyTwiceInSet, fault.AlreadyCurrent, fault.EarlierVersion, fault.SectionDiffers, fault.PathDiffers, fault.DropsFloor, fault.DropsCeiling, fault.AddsNothing}, ", the half that reads the events; the act checks the folders' bytes"},
	{fault.GuardMeetable, "some output can meet each new example: no floor's whole text is an exit line exit N, since hyp cuts the Probe's last line exit N off before it compares; no two equals floors differ; an equals floor holds the text of each contains floor; no floor holds the text of a ceiling",
		[]fault.Code{fault.FloorIsExitLine, fault.EqualsFloorsDiffer, fault.EqualsLacksContains, fault.FloorHoldsCeiling}, ": it reads the example folders"},
	{fault.GuardKind, "H0 gives each new or changed example the verdict its line's kind needs: a gap example fails, a guard example is met; the add goal or the place records each example's hash and that verdict in git, and H0's outputs in the clone's trace",
		[]fault.Code{fault.WrongVerdictOnH0, fault.H0MovedWhilePlacing, fault.AddedVerdictWrong}, ", that the verdict on H0 an add goal records for each example it carries is the one its line's kind needs; the act runs the Probes on H0"},
	{fault.GuardBytes, "the commit exists in this clone",
		[]fault.Code{fault.CommitNotHere, fault.NotACommit}, ": it reads git"},
	{fault.GuardOpen, "the Goal is open, unless the commit's subject hash is H0's; the examples do not change while the Probes run",
		[]fault.Code{fault.NothingPlaced, fault.NotOpen, fault.H0MovedWhileRun, fault.ExamplesChanged, fault.ExampleChanged}, ""},
	{fault.GuardOwnPath, "the runner's own change of the commit writes no path under goals/X/ of another Goal X, and adds or changes no file under a goals/X/ folder of no Goal in the store; a change that only deletes files of such a folder passes; goals/model/, the repository's kept model, is no Goal's folder",
		[]fault.Code{fault.OtherGoalPath, fault.NoGoalFolder}, ": it reads git"},
	{fault.GuardLeafOnly, "a run is on H0, or on a commit that no row of the Goal descends from",
		[]fault.Code{fault.AncestorOfRow}, ": it reads git"},
	{fault.GuardBudget, "at most three Hypotheses run since the boundary, H0 never counted",
		[]fault.Code{fault.BudgetUsed}, ": it counts the Results of the clone's trace"},
	{fault.GuardStable, "two runs on the same bytes give one verdict; a run it stops still records its Result, marked, which counts nothing",
		[]fault.Code{fault.OtherVerdict, fault.VerdictsDiffered}, ": it reads the cells of the clone's trace; replay keeps the mark a Result was recorded with"},
	{fault.GuardCandidate, "the row is a candidate, a leaf row met on every line that no unselect rejected; only the owner selects, and a lone candidate is selected by the Results that show it, with no select event",
		[]fault.Code{fault.RoleCannotSelect, fault.NoRow, fault.OwnerSelectsAmong, fault.RowNotLeaf, fault.RowNotMet, fault.NoResultOnRow}, ": it reads the cells of the clone's trace"},
	{fault.GuardNewCell, "after an unselect, the runner selects only once a run recorded a cell whose example version, or whose row, had not run before it; the owner selects at any time",
		[]fault.Code{fault.NoNewCell}, ": it reads the cells of the clone's trace"},
	{fault.GuardGoalsActor, "only the actor that added the Goal selects for the owner, lands, closes or unselects it",
		[]fault.Code{fault.NotGoalsActor}, ""},
	{fault.GuardSpec, "the selected commit holds a spec under goals/G/specs/; a model there names each rule of its check configs in a variant config, and holds a chart; when the commit changes the kept model, goals/model/, against H0, whatever commit the clone has checked out, a variant config under goals/G/specs/ names each kept rule that is new or whose definition text differs, the other kept rules need none, and the folder holds a chart",
		[]fault.Code{fault.NoSpec, fault.NoRead, fault.RuleNoVariant, fault.NoChart}, ": run and land read git and their events carry what they read; the calculated select and hyp status check what the Result of the commit carries"},
	{fault.GuardOwnerLands, "only the owner lands a Goal",
		[]fault.Code{fault.NotOwner}, ""},
	{fault.GuardEvidence, "a land on a commit carries, for every current example, its verdict on that commit, from this clone's Result or from a run the land makes first; every verdict is met, the row is not one an unselect rejected, and a select event, when one holds the Goal, is on that row",
		[]fault.Code{fault.LandNotOpen, fault.CommitUnselected, fault.NotTheSelect, fault.EvidenceMissing, fault.NotMetOnCommit}, ": replay reads the verdicts the land carries"},
	{fault.GuardStop, "a land with no commit needs a stop: two candidates or budget used",
		[]fault.Code{fault.NoStop, fault.StopNotHeld}, ", the stop the land names against the select; the act reads the stop the Results hold"},
	{fault.GuardOwnerClose, "only the owner closes a Goal, at any stop or none",
		[]fault.Code{fault.NotOwnerClose}, ""},
	{fault.GuardOwnerUnsel, "only the owner unselects a Goal",
		[]fault.Code{fault.NotOwnerUnselect}, ""},
	{fault.GuardInWork, "an old fold, which no act writes and replay still reads, names one Goal in work, written, open or selected, that the store holds",
		[]fault.Code{fault.NoGoalToFold, fault.FoldIntoEnded}, ": a Goal's state reads only events stored in git"},
	{fault.GuardIssue, "the oldest form of a fold, which no act writes and replay still reads, names an issue by its number, a whole number above zero",
		[]fault.Code{fault.NoIssue}, ""},
	{fault.GuardNotRefuted, "a refuted finding takes no name, no refute and no decline",
		[]fault.Code{fault.Refuted}, " for a refute and a decline, which are in the same clone's trace; a name is checked by its act only, since a Goal's record is in git"},
	{fault.GuardOwnerDecl, "only the owner declines a finding, and with a why",
		[]fault.Code{fault.NotOwnerDecline, fault.DeclineNoWhy}, ""},
	{fault.GuardNotDeclined, "a declined finding takes no name, no refute and no decline",
		[]fault.Code{fault.Declined}, " for a refute and a decline, which are in the same clone's trace; a name is checked by its act only, since a Goal's record is in git"},
	{fault.GuardNotFolded, "the owner declines only a finding that no Goal named, whether that Goal is in work or ended",
		[]fault.Code{fault.DeclineFolded, fault.DeclineFoldedIssue}, ": replay reads the names in git and the decline in the clone's trace"},
	{fault.GuardOneInWork, "add goal and place name a finding that no other Goal in work names",
		[]fault.Code{fault.NamedInWork}, ": it reads the findings of the clone's trace"},
	{fault.GuardCovered, "a place names, in the findings section of at least one example of the set, each finding the Goal names at add goal",
		[]fault.Code{fault.NotCovered}, ": it reads the example folders"},
	{fault.NoGuard, "a file hyp reads breaks a rule of its form, or the act names what the store does not hold",
		[]fault.Code{fault.NoStore, fault.EventVersion, fault.EventType, fault.PayloadDecode, fault.SelectRole, fault.TrailingData,
			fault.IDNotULID, fault.ActorNotFile, fault.NoIDAfter, fault.NoGoalToPlace, fault.NoGoalToRun, fault.NoGoalToSelect,
			fault.NoGoalToLand, fault.NoGoalToClose, fault.NoGoalToUnselect, fault.NoGoalForFinding, fault.NoFinding, fault.NoExampleAtCommit, fault.FoldNoFinding, fault.FoldGoalAndIssue, fault.NoExampleTxtar,
			fault.TextBeforeFirst, fault.SecondSection, fault.ExternalNotName, fault.InputOutside, fault.InputTwice,
			fault.UnknownSection, fault.NoSection, fault.Submodule, fault.EntryLeaves, fault.NotFileOrDir, fault.ReadFolder}, ""},
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
		"The store is .tree/ at the git top level: .tree/events/<actor>/<id>.json, one file per event, each written once and never changed. It holds the owner's acts, each with the evidence it approves, and nothing else: no act of the grill or the runner writes to it. A store written by an earlier hyp also holds the grill's places and the runner's selects, which replay reads as before.",
		"An act adds one new file to the store and changes no file in it, so two branches that each record acts merge with a plain git merge, with no merge driver and no conflict, on GitHub too.",
		"A store written by an earlier hyp also holds .tree/events/<actor>.jsonl, one file per actor of appended lines; hyp reads it as before and appends no line to it.",
		"The first add goal creates the store; .tree/.gitattributes merges the older per-actor files by union. hyp reads from .tree/ only the events of the kinds that change a Goal's state, and no output.",
		"The trace is $HOME/.hyp/traces/<sha256 of the clone's git common directory>/: events/<actor>.jsonl holds the events of every kind that changes no Goal's state, outputs/<sha256> every output, and cells/ H0's cells of the places and the cells a try ran again. Every command that names an actor records there what hyp printed to it, before it prints.",
		"The trace is a cache of this clone: a clone that lacks a cell prints that the cell is not in this clone's trace, with the ready hyp try line that runs it again, and a try of a missing cell saves it there.",
	}},
	{"merge order and replay", []string{
		"Replay reads every event of git and of the clone's trace by id, then by actor, then by line bytes, and calculates every table and state on every command; no state is stored. The states written, open and ended read only events stored in git, so every clone calculates them the same; selected is calculated from the Results of this clone's trace, so a clone that lacks them shows the Goal open, and a land carries the evidence it approves into git.",
		"Replay checks again, at each event's place, the guards whose entry says checked at replay; an event that fails them is listed under refused at replay and never deleted.",
		"An end names its snapshot, the Goal's events stored in git that replay applied when the owner ended it; an event of a kind an end freezes that the snapshot does not name is stopped by Not ended.",
	}},
	{"ids, actors and roles", []string{
		"An event id is a ULID; a Goal id is G- and the last six characters of its goal.add id, a Result id R- and six, a finding id F- and six of its finding.add id.",
		"An actor matches ^[a-z0-9][a-z0-9-]{0,39}$ and names its events file; a write act needs a role: owner, grill, probe-writer or runner. An act on a finding needs only an actor, and records the role it declared when it is one of the four.",
	}},
	{"findings", []string{
		"A finding is a fact an agent found, with a pointer to its evidence, and names no Goal. Any agent adds one, or refutes it when its fact is false; a finding may carry an add goal its agent drafted for the owner. The owner declines, with a why, a finding that no Goal named whose fact is true and that the owner does not want fixed.",
		"A decided want is a line of a written Goal; a fact no line needs is a finding.",
		"A Goal names the findings it is about in its own record, in git, by their ids alone: the owner names them at add goal with --finding F, and the grill in the findings section of an example, one id per line. Add goal and place refuse an id this clone's trace does not hold, a refuted or a declined finding, and a finding another Goal in work names; place refuses the set while a finding the Goal names at add goal has no example that names it. An old fold event counts as a name of each finding it folds by its Goal.",
		"A finding's state follows from the Goals that name it: folded into G while G, a Goal in work, names it; settled once no Goal in work names it and a Goal that names it landed with a select; ended, not settled, when every Goal that named it ended with no select, and a later Goal may name it.",
		"Its events are in the clone's trace and never in git: they change no file under .tree/ and no Goal's state, and an end's snapshot names none of them. A finding of the old form, recorded on a Goal, reads not folded until a Goal names it.",
		"hyp finding list prints each finding with its state: not folded, folded into G, settled in G, ended in G, the old form's folded into #N, refuted, or declined with its why; a settled, ended, refuted or declined finding is left out unless --all. hyp status G prints the id of each finding G names on any clone, with its state and fact when this clone's trace holds it, and says so when it does not; hyp status counts the findings not folded, which wait for a Goal.",
		"A finding's evidence may name an example, as G-XXXXXX/line-N/name, and a commit: finding add refuses a commit that does not hold the example's example.txtar. A finding and a Goal line are about the steps of the kept model, goals/model/, that their examples' run sections take, calculated and never written: each command of the run section takes the step of the kept model tagged \\* command: X whose words start it, the longest when two do, or no step. hyp finding list prints the walk of a finding's example on its commit; hyp status G prints the walk of each placed example on H0, or on the selected commit while G is selected.",
	}},
	{"the subject", []string{
		"A commit's subject is its tree without the top-level entries goals and .tree; its git tree hash is the subject hash, one Hypothesis, one row.",
		"H0 is the Goal's base commit; a row is a leaf when no Result on another row carries one of its commits among the commits of the Goal's rows that its run read its commit descends from.",
	}},
	{"budget and stops", []string{
		"The budget is three Hypotheses since the boundary, H0 never counted; a refused run costs nothing.",
		"The boundary is the latest of the latest example-changing place and the latest unselect, else the Goal's start.",
		"The stops: selected (one candidate this clone's Results show that holds a spec and writes no other Goal's folder, or a select event), two candidates (two or more candidates and no select), budget used (three Hypotheses counted, the third with a live cell for every example, and no candidate). A row an unselect rejected is no candidate.",
		"A stop is calculated, never stored; it waits for the owner. At two candidates and budget used the state block adds the table of outputs.",
		"Runs, rows, candidates and every stop are calculated from the Results of this clone's trace; a select event of the older form, or the owner's select among two candidates, is read from git.",
		"No state or print reads git history: an act that reads git, such as run or land, carries what it read in its event, the commits of the Goal's rows its commit descends from, its spec folder and its kept model, and rows, leaves, candidates, the Spec guard of the calculated select and the prints of the selected commit are calculated from events. A select or land of the older form, which carries no read, is printed from git.",
		"Every state block of a Goal with a select warns when the selected commit is in this clone and on no branch or tag, whatever other commit with the same subject hash is on one.",
	}},
	{"the step log", []string{
		"A step log is a file of step lines, one step per line: step <Action> <json>, where Action is the name of a step and json is one JSON value on the same line. Any program can write it; a writer only appends to it, and never changes or removes a line.",
		"With HYP_STEP_LOG set to a file, hyp appends to it one step line for each event it writes, after the event is stored: Action is the step its kind's entry names and json is the event line as stored. A read writes no step line, and a command that writes no event, such as hyp spec, appends nothing.",
		"With HYP_STEP_LOG unset or empty, hyp writes no step line anywhere, and every command prints and writes the same as without the step log.",
		"For each Probe run hyp sets HYP_STEP_LOG to a fresh file; run keeps that file in the clone's trace, outputs/<sha256>, when the subject's programs wrote to it, and the Result's cell of the example carries its hash. hyp show ACTOR prints, after the reads, the step log of each Result of the actor that keeps one.",
	}},
	{"the example file", []string{
		"An example is goals/G/LINE/EXAMPLE/ with example.txtar and probe/. Its sections: words (one), input/<path> (any), run (one), floor: contains and floor: equals (one or more floors), ceiling (any), external (at most one), findings (at most one: the id of each finding the example names, one per line).",
		"One literal per section, byte for byte; a section body's final newline does not count.",
		"An example is met when the output holds every floor and contains no ceiling, else failed; a line is met on a row when every placed example of it is met there.",
	}},
	{"the Probe", []string{
		"hyp runs probe/run with empty stdin and a 10 minute timeout, in a fresh copy of the commit's subject ($HYP_SUBJECT), with $HYP_EXAMPLE a copy of the example folder and $HYP_RUN the run section.",
		"Only PATH, HOME and the names of the external section pass from the environment; HYP_STEP_LOG names a fresh empty step log of that run (the step log).",
		"The Probe prints the stdout of run, then a last line exit N; the output is the rest, with one final newline dropped. Its stderr is kept beside it and never compared.",
		"A Probe that exits non-zero, times out or prints no exit line records nothing: exit 3.",
	}},
}
