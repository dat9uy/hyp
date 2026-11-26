package channel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/acts"
	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
)

// The codes of the failures the channel itself reports, about the command
// line it parses.
const (
	codeNoCommand      fault.Code = "no-command"      // no values
	codeUnknownCommand fault.Code = "unknown-command" // the command
	codeAddWhat        fault.Code = "add-what"        // no values
	codeFlags          fault.Code = "flags"           // command, cause
	codePlaceArgs      fault.Code = "place-args"      // count
	codeRunArgs        fault.Code = "run-args"        // count
	codeSelectArgs     fault.Code = "select-args"     // count
	codeStatusArgs     fault.Code = "status-args"     // count
	codeShowArgs       fault.Code = "show-args"       // count
	codeShowActor      fault.Code = "show-actor"      // actor
	codeReplaceTwice   fault.Code = "replace-twice"   // count
	codeNeedsBase      fault.Code = "needs-base"      // no values
	codeNeedsWords     fault.Code = "needs-words"     // no values
	codeLineFlag       fault.Code = "line-flag"       // the --line value
	codeWorkDir        fault.Code = "work-dir"        // cause
	codeStateBlock     fault.Code = "state-block"     // cause
	codeNeedsActor     fault.Code = "needs-actor"     // command
	codeNotRecorded    fault.Code = "not-recorded"    // cause
	codePruneArgs      fault.Code = "prune-args"      // count
	codeTryArgs        fault.Code = "try-args"        // count
	codeLandArgs       fault.Code = "land-args"       // count
	codeCloseArgs      fault.Code = "close-args"      // count
	codeUnselectArgs   fault.Code = "unselect-args"   // count
	codeSpecArgs       fault.Code = "spec-args"       // count
	codeFindingWhat    fault.Code = "finding-what"    // no values
	codeFindingArgs    fault.Code = "finding-args"    // act, count
)

// words is the text of every code: a format whose verbs name the code's
// values in order. hyp prints each value as it holds it, so no format quotes
// a value.
var words = map[fault.Code]string{
	// arguments every act checks
	fault.NoActor:     "no actor; give --actor or $HYP_ACTOR",
	fault.BadActor:    "actor %s must match ^[a-z0-9][a-z0-9-]{0,39}$ (--actor or $HYP_ACTOR)",
	fault.BadRole:     "role %s must be one of %s (--role or $HYP_ROLE)",
	fault.NoRole:      "no role; give --role or $HYP_ROLE, one of %s",
	fault.WhyTooLong:  "--why is %d runes; the limit is %d",
	fault.NoGoal:      "no Goal %s",
	fault.GoalTwice:   "%s matches %d Goals; name the event id",
	fault.NoLine:      "%s has no line %s",
	fault.EncodeEvent: "encode %s: %s",

	// add goal
	fault.EmptyWords:    "the Goal's words are empty",
	fault.LineNoText:    "line %d has no text",
	fault.BaseNotCommit: "--base %s is not a commit",
	fault.BaseNotHere:   "--base %s is not a commit of this repository",

	// try
	fault.TryWorktreeCommit: "try takes --worktree or --commit, not both",
	fault.TryNoCommit:       "--commit %s is not a commit of this clone",

	// place
	fault.BadExampleKey:        "%s is not LINE/EXAMPLE, for example line-1/a (an example name matches ^[a-z0-9][a-z0-9-]{0,39}$)",
	fault.SetUnchanged:         "no example of %s changed since its last place; there is nothing to place",
	fault.SetChangedWhileProbe: "the example set of %s changed while the Probes ran; place it again",
	fault.KeyTwiceInSet:        "%s is twice in one place",
	fault.LineHoldsNoExample:   "%s holds no example: %s has no example folder",
	fault.PlacedFolderGone:     "%s was placed, and its folder %s is gone; a placed example stays in the set",
	fault.NotAnExampleName:     "%s is not an example folder: an example name matches ^[a-z0-9][a-z0-9-]{0,39}$",
	fault.ChangedWhileProbe:    "%s changed from %s to %s while the Probe ran; place it again",
	fault.H0MovedWhilePlacing:  "H0 of %s moved from %s to %s while the Probe ran; place it again",
	fault.NotAFolder:           "%s is not a folder on disk",
	fault.FolderOnDisk:         "%s",
	fault.NotCommitted:         "%s is not committed at HEAD",
	fault.DiskNotCommitted:     "%s on disk is not the folder committed at HEAD",
	fault.DiffersAt:            "%s differs from HEAD at %s",
	fault.ObjectMissing:        "object %s not in this clone",
	fault.NoExampleTxtar:       "example folder %s has no example.txtar",

	// run and select
	fault.NoCommit:        "run needs a COMMIT",
	fault.BadGoalLine:     "%s is not <Goal>/<line>, for example G-NXHHVA/line-1",
	fault.GoalVanished:    "Goal %s vanished from the log",
	fault.NotACommit:      "%s is not a commit",
	fault.H0MovedWhileRun: "H0 of %s moved from %s to %s while the Probes ran; run again",
	fault.ExamplesChanged: "the examples of the line changed while the Probes ran; run again",
	fault.ExampleChanged:  "%s changed from %s to %s while the Probes ran; run again",
	fault.SelectNotInLog:  "the select of %s is not in the log",
	fault.NoResultOnRow:   "row %s has no Result to select",
	fault.NoSpec:          "commit %s holds no file under %s; commit the spec, then run one line on that commit",
	fault.RuleNoVariant:   "commit %s holds a model whose rules no variant config names: %s; commit a variant config under %s that lists each rule it breaks, then run one line on that commit",
	fault.NoChart:         "commit %s holds a model and no chart under %s; commit the chart (.mmd), then run one line on that commit",

	// events and their payloads
	fault.EventVersion:   "event version %d, want %d",
	fault.EventType:      "unknown event type %s",
	fault.PayloadDecode:  "%s payload: %s",
	fault.TrailingData:   "trailing data after JSON object",
	fault.PayloadInvalid: "%s",

	// the example file
	fault.SectionDiffers:  "the %s section differs from the current version; a change there is a new example key, not a new version",
	fault.PathDiffers:     "%s differs from the current version; a change there is a new example key, not a new version",
	fault.DropsFloor:      "the new version drops a floor (%s): %s",
	fault.DropsCeiling:    "the new version drops a ceiling: %s",
	fault.AddsNothing:     "the new version adds nothing; it must add at least one floor or ceiling",
	fault.TextBeforeFirst: "example.txtar line %d: text before the first section",
	fault.SecondSection:   "example.txtar line %d: a second %s section",
	fault.ExternalNotName: "example.txtar line %d: external line %s is not an environment variable name",
	fault.InputOutside:    "example.txtar line %d: input path %s must be a relative path inside the folder",
	fault.InputTwice:      "example.txtar line %d: input/%s appears twice",
	fault.UnknownSection:  "example.txtar line %d: unknown section %s",
	fault.NoSection:       "example.txtar has no %s section",

	// git: a command's arguments print as one command line
	fault.GitExit:       "git %s: exit %d: %s",
	fault.GitRun:        "git %s: %s",
	fault.GitNotAllowed: "git %s is not on hyp's read list",
	fault.LsTreeParse:   "git ls-tree %s: cannot parse %s",
	fault.Submodule:     "%s: %s is a submodule (mode 160000)",
	fault.EntryLeaves:   "tree entry %s leaves the destination",
	fault.Extract:       "extract %s: %s",
	fault.CatFileBatch:  "git cat-file --batch %s: %s",
	fault.BatchEnds:     "output ends before the object %s",
	fault.BatchMissing:  "object %s is missing from this clone",
	fault.BatchRecord:   "unexpected record %s for %s",
	fault.BatchSize:     "bad size in record %s",
	fault.NotRealDir:    "extract %s: a parent is not a real directory (symlink or file already written)",
	fault.ReadLink:      "read link %s: %s",
	fault.ReadPath:      "read %s: %s",
	fault.NotFileOrDir:  "%s is not a regular file, a symlink or a folder",
	fault.ReadFolder:    "read folder %s: %s",

	// the store
	fault.NotInRepo:    "%s is not in a git repository: %s",
	fault.ObjectFormat: "object format %s: hyp reads sha1 repositories only",
	fault.CreateStore:  "create store: %s",
	fault.NoStore:      "no store; run hyp add goal",
	fault.IDNotULID:    "id %s is not a ULID",
	fault.ActorNotFile: "actor %s is not the file's actor %s",
	fault.ReadEvents:   "read events: %s",
	fault.OpenStore:    "open store: %s",
	fault.LockStore:    "lock store: %s",
	fault.NoIDAfter:    "no ULID sorts after %s, the newest id the store holds",
	fault.FileOpen:     "open %s: %s",
	fault.FileStat:     "stat %s: %s",
	fault.FileRead:     "read %s: %s",
	fault.FileWrite:    "write %s: %s",
	fault.FileSync:     "sync %s: %s",
	fault.StoreOutput:  "store output: %s",
	fault.EventHome:    "a %s event is not stored in the %s: its entry in hyp spec names another place",

	// the trace
	fault.NoTraceFolder: "no folder for the trace: $HOME is not an absolute path",
	fault.CreateTrace:   "create the trace folder: %s",
	fault.ReadTraces:    "read the traces folder %s: %s",
	fault.RemoveTrace:   "remove the trace folder %s: %s",

	// Probes
	fault.ProbeTemp:       "probe %s: %s",
	fault.ProbeTxtar:      "probe %s: example.txtar: %s",
	fault.ProbeInput:      "probe %s: input/%s: %s",
	fault.ProbeMissing:    "probe %s: probe/run is missing",
	fault.ProbeStart:      "probe %s: cannot start probe/run: %s",
	fault.ProbeTimeout:    "probe %s: probe/run timed out after %s",
	fault.ProbeExit:       "probe %s: probe/run exited %d: %s",
	fault.ProbeWait:       "probe %s: probe/run: %s",
	fault.ProbeNoExitLine: `probe %s: probe/run printed no final line "exit N"`,

	// guards that read the tables
	fault.NothingPlaced:      "%s has no placed example for %s; there is nothing to run",
	fault.NotOpen:            "%s is %s; only H0 may run before the Goal is open",
	fault.CommitNotHere:      "commit %s is not in this clone",
	fault.LeafNeedsCommit:    "Leaf only: the leaf mark of %s needs commit %s, which this clone lacks (%s)",
	fault.AncestorOfRow:      "%s is an ancestor of %s, a commit of a row of %s; run a leaf",
	fault.OtherGoalPath:      "commit %[1]s writes %[2]s, a path of another Goal; %[3]s may write only under goals/%[3]s/",
	fault.BudgetUsed:         "%s has used %d of %d Hypotheses since its last change; the owner decides",
	fault.OtherVerdict:       "%s on this Hypothesis gave another verdict than its live cell: %s",
	fault.NoGoalToRun:        "no Goal %s to run on",
	fault.NoGoalToPlace:      "no Goal %s to place on",
	fault.NoGoalToSelect:     "no Goal %s to select on",
	fault.RoleCannotSelect:   "the role is %s; only the runner or the owner selects",
	fault.NoTree:             "select needs a subject hash",
	fault.TreeTwice:          "%s matches %d rows of %s; give more of the subject hash",
	fault.NoRow:              "%s has no row %s; a candidate is a leaf row met on every line",
	fault.RowLeafUnknown:     "Candidate: the leaf mark of row %s needs commit %s, which this clone lacks",
	fault.OwnerSelectsAmong:  "%s has %d candidates; only the owner selects among them",
	fault.RowNotLeaf:         "row %s is not a leaf; a commit of another row descends from it",
	fault.RowNotMet:          "row %s is not met on every line",
	fault.NotGoalsActor:      "%s was added by %s, not %s; only that actor acts for the owner",
	fault.NoGoalToReplace:    "no Goal %s to replace",
	fault.NotWritten:         "%s is %s; only a written Goal may be replaced",
	fault.TooManyLines:       "%d lines; a Goal lists at most %d",
	fault.BadLineKind:        "line %d (%s) has kind %s; want gap or guard",
	fault.NoGapLine:          "the Goal has no gap line",
	fault.HasSelect:          "%s has a select; it takes no place, run or second select",
	fault.NotGrill:           "only the grill places an example; the role is %s",
	fault.AlreadyCurrent:     "%s is already the current version of %s (@%d); a new version adds a floor or a ceiling",
	fault.EarlierVersion:     "%s was version @%d of %s; an earlier version never comes back",
	fault.NoResultToRetract:  "no Result %s to retract",
	fault.NotOwnResult:       "%s was run by %s, not %s",
	fault.Ended:              "%s is ended; an ended Goal takes no write",
	fault.NotInSnapshot:      "the end of %s, %s, does not name this event: it was written in another clone before the end and reached this log after it, so the ended Goal keeps what the owner saw",
	fault.NotOwnerClose:      "only the owner closes a Goal; the role is %s",
	fault.NoGoalToClose:      "no Goal %s to close",
	fault.NotOwner:           "only the owner lands a Goal; the role is %s",
	fault.NoStop:             "%s has no stop; the owner lands a Goal once it is selected, has two candidates or has used its budget",
	fault.NoGoalToLand:       "no Goal %s to land",
	fault.StopNotHeld:        "the land names the stop %s, which %s does not hold",
	fault.FromNotEnded:       "%s is %s; --from cites only an ended Goal",
	fault.NotOwnerUnselect:   "only the owner unselects a Goal; the role is %s",
	fault.NoSelect:           "%s is %s and has no select; only a selected Goal is unselected",
	fault.NoGoalToUnselect:   "no Goal %s to unselect",
	fault.NoGoalForFinding:   "no Goal %s to add a finding to",
	fault.NoFinding:          "no finding %s",
	fault.FindingTwice:       "%s matches %d findings; give the finding.add event id",
	fault.NoIssue:            "the old form of a fold names the issue it folds into, a whole number above zero",
	fault.NoGoalToFold:       "no Goal %s to fold into; a fold names a Goal in work: written, open or selected",
	fault.FoldIntoEnded:      "%s is ended; a fold names a Goal in work: written, open or selected",
	fault.FoldGoalAndIssue:   "the fold names both a Goal and an issue; it names one Goal",
	fault.Refuted:            "%s is refuted; a refuted finding takes no name, no refute and no decline",
	fault.Declined:           "%s is declined; a declined finding takes no name, no refute and no decline",
	fault.NotOwnerDecline:    "only the owner declines a finding; the role is %s",
	fault.DeclineNoWhy:       "a decline needs --why TEXT: the owner's reason not to fix the finding",
	fault.DeclineFolded:      "%s is named by %s; the owner declines only a finding that no Goal named",
	fault.DeclineFoldedIssue: "%s is folded into #%d; the owner declines only a finding that no Goal named",
	fault.FoldNoFinding:      "the fold names no finding",
	fault.NamedInWork:        "%s is named by %s, a Goal in work; one Goal in work names a finding",
	fault.NotCovered:         "%s names %s, and no example of the set names it in its findings section",
	fault.EmptyFact:          "finding add needs --fact TEXT",
	fault.EmptyEvidence:      "finding add needs --evidence TEXT, a pointer to the evidence",
	fault.DraftNotAddGoal:    "--draft is an add goal command line for the owner: it starts with hyp add goal",
	fault.NoNewCell:          "%s was unselected by %s, and no run since recorded a new cell: an example version or a row that had not run before the unselect; the runner selects again after such a run, the owner at any time",

	// the channel's own
	codeNoCommand:      "no command",
	codeUnknownCommand: "unknown command %s",
	codeAddWhat:        "add what? only add goal exists",
	codeFlags:          "%s: %s",
	codePlaceArgs:      "place takes one Goal and places its whole example set, for example: hyp place G-NXHHVA; got %d arguments",
	codeRunArgs:        "run takes COMMIT and <Goal>/<line>, for example: hyp run 4df2b9e G-NXHHVA/line-3; got %d arguments",
	codeSelectArgs:     "select takes a Goal and a subject hash, for example: hyp select G-NXHHVA 4df2b9e...; got %d arguments",
	codeStatusArgs:     "status takes at most one Goal, got %d",
	codeShowArgs:       "show takes one actor, for example: hyp show runner-1; got %d arguments",
	codeShowActor:      "show %s: an actor matches ^[a-z0-9][a-z0-9-]{0,39}$",
	codeReplaceTwice:   "add goal takes at most one Goal to replace, got %d",
	codeNeedsBase:      "add goal needs --base COMMIT",
	codeNeedsWords:     "add goal needs --words TEXT",
	codeLineFlag:       "--line %s: want gap:TEXT or guard:TEXT",
	codeWorkDir:        "working directory: %s",
	codeStateBlock:     "state block: %s",
	codeNeedsActor:     "%s needs an actor (--actor or $HYP_ACTOR): hyp records what it prints to each actor",
	codePruneArgs:      "prune takes no argument but --remove; got %d arguments",
	codeFindingWhat:    "finding what? finding add, finding refute, finding decline or finding list",
	codeFindingArgs:    "%s takes the arguments its usage in hyp help names; got %d arguments",
	codeLandArgs:       "land takes one Goal, for example: hyp land G-NXHHVA; got %d arguments",
	codeCloseArgs:      "close takes one Goal, for example: hyp close G-NXHHVA; got %d arguments",
	codeUnselectArgs:   "unselect takes one Goal, for example: hyp unselect G-NXHHVA; got %d arguments",
	codeSpecArgs:       "spec takes no argument; got %d",
	codeTryArgs:        "try takes a Goal and LINE/EXAMPLE, for example: hyp try G-NXHHVA line-1/a; got %d arguments",
	codeNotRecorded:    "could not record what this command printed, so hyp prints nothing of it: %s",
}

// shaped are the codes whose values do not fit one format: bodyText prints
// each with a function of its own, kindText and stableText.
var shaped = map[fault.Code]bool{fault.WrongVerdictOnH0: true, fault.VerdictsDiffered: true, fault.SetNotPlaced: true}

// failText is the text hyp prints for err: "refused: " and the guard for a
// refusal, then the words of its code with its values. root is the git top
// level, or "" when it is not known.
func failText(err error, root string) string {
	if fault.KindOf(err) == fault.Refusal {
		return "refused: " + bodyText(err, root)
	}
	return bodyText(err, root)
}

// bodyText is the words of err without the "refused: " of a refusal: the
// guard that refused, then the words of its code with its values. An error
// that is no failure of hyp, such as one of the operating system, prints its
// own text.
func bodyText(err error, root string) string {
	e, ok := fault.As(err)
	if !ok {
		return err.Error()
	}
	if e.Code == fault.SetNotPlaced {
		// each cause names its own guard
		return setText(e.Args, root)
	}
	var text string
	if format, ok := words[e.Code]; e.Code == fault.WrongVerdictOnH0 {
		text = kindText(e.Args)
	} else if e.Code == fault.VerdictsDiffered {
		text = stableText(e.Args, root)
	} else if ok {
		values := make([]any, len(e.Args))
		for i, a := range e.Args {
			values[i] = valueText(e.Code, a, root)
		}
		text = fmt.Sprintf(format, values...)
	} else {
		// a code with no words yet: name the code, so the gap shows
		text = "(" + string(e.Code) + ")"
	}
	if e.Guard != fault.NoGuard {
		text = string(e.Guard) + ": " + text
	}
	return text
}

// valueText prints one value of a failure as hyp holds it: a cause by its own
// words, a list of names joined by ", ", a git command's arguments joined by
// a space as the command line they were.
func valueText(code fault.Code, a any, root string) any {
	switch v := a.(type) {
	case error:
		return bodyText(v, root)
	case []string:
		switch code {
		case fault.GitExit, fault.GitRun, fault.GitNotAllowed:
			return strings.Join(v, " ")
		}
		return strings.Join(v, ", ")
	case []byte:
		return string(v)
	}
	return a
}

// kindText is the Kind refusal: H0's verdict and output beside the
// example's floors and ceilings, each as the bytes hyp holds.
func kindText(args []any) string {
	if len(args) != 1 {
		return "(" + string(fault.WrongVerdictOnH0) + ")"
	}
	f, ok := args[0].(acts.KindFacts)
	if !ok {
		return "(" + string(fault.WrongVerdictOnH0) + ")"
	}
	want := "fail"
	if f.Line.Kind == events.KindGuard {
		want = "be met"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s is a %s line, so its example must %s on H0; it is %s\n", f.Line.Key, f.Line.Kind, want, f.Verdict)
	writeBlock(&b, "H0 output, exit "+strconv.Itoa(f.Out.Exit), f.Out.Stdout, int64(len(f.Out.Stdout)))
	writeBlock(&b, "H0 stderr", f.Out.Stderr, int64(len(f.Out.Stderr)))
	for _, fl := range f.Floors {
		writeBlock(&b, "floor: "+fl.Op, []byte(fl.Text), int64(len(fl.Text)))
	}
	for _, c := range f.Ceilings {
		writeBlock(&b, "ceiling", []byte(c), int64(len(c)))
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// setText is the refusal of a place that places none of the set: the count
// of causes, then each cause under its example key, or its line key for a
// line that holds no example, with its guard and words.
func setText(args []any, root string) string {
	if len(args) != 1 {
		return "(" + string(fault.SetNotPlaced) + ")"
	}
	f, ok := args[0].(acts.SetFacts)
	if !ok || len(f.Causes) == 0 {
		return "(" + string(fault.SetNotPlaced) + ")"
	}
	causes := f.Causes
	var b strings.Builder
	plural := "s"
	if len(causes) == 1 {
		plural = ""
	}
	fmt.Fprintf(&b, "none of the example set of %v is placed; %d cause%s:\n", f.Goal, len(causes), plural)
	for _, c := range causes {
		fmt.Fprintf(&b, "%s: %s\n", c.Key, bodyText(c.Err, root))
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// maxShown is how much of one stored output a Stable refusal prints.
const maxShown = 4096

// stableText is the Stable refusal: for each example whose verdict differed,
// both verdicts and both outputs, read from the store.
func stableText(args []any, root string) string {
	if len(args) != 1 {
		return "(" + string(fault.VerdictsDiffered) + ")"
	}
	f, ok := args[0].(acts.StableFacts)
	if !ok {
		return "(" + string(fault.VerdictsDiffered) + ")"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "two runs of %s on one Hypothesis gave different verdicts; the first stands\n", f.Goal)
	for _, d := range f.Diffs {
		fmt.Fprintf(&b, "%s: %s, then %s\n", d.Key, d.FirstVerdict, d.Verdict)
		writeStored(&b, root, "first output "+d.FirstOutput, d.FirstOutput, maxShown)
		writeStored(&b, root, "second output "+d.Output, d.Output, maxShown)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// writeBlock writes bytes hyp holds under a label, as they are: never quoted,
// escaped or indented. The label line ends in the byte count, so a reader
// knows where the bytes end: "<label>, N bytes:", the bytes, and one newline.
// size is the size of the whole; when data is only its first bytes, the label
// says "first N of M bytes". An empty whole is the label line
// "<label>, 0 bytes" alone.
func writeBlock(b *strings.Builder, label string, data []byte, size int64) {
	switch {
	case size == 0:
		fmt.Fprintf(b, "%s, 0 bytes\n", label)
		return
	case int64(len(data)) < size:
		fmt.Fprintf(b, "%s, first %d of %d bytes:\n", label, len(data), size)
	default:
		fmt.Fprintf(b, "%s, %d bytes:\n", label, size)
	}
	b.Write(data)
	b.WriteString("\n")
}

// writeStored writes the stored output named hash as a block, at most max
// bytes of it, or says why it cannot. It reads the clone's trace, then
// .tree/outputs, where an earlier hyp stored outputs.
func writeStored(b *strings.Builder, root, label, hash string, max int) {
	data, size, err := logstore.ReadOutput(root, hash, max)
	if err != nil {
		fmt.Fprintf(b, "%s: not readable here: %s\n", label, bodyText(err, root))
		return
	}
	writeBlock(b, label, data, size)
}
