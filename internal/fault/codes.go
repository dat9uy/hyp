package fault

// The codes of every failure the parts of hyp outside the channel report.
// Each value is an identifier; internal/channel holds the words of each one,
// and its tests check that every code here has words that name as many
// values as each call passes.
const (
	// arguments every act checks
	NoActor     Code = "no-actor"     // no values
	BadActor    Code = "bad-actor"    // actor
	BadRole     Code = "bad-role"     // role, the roles
	NoRole      Code = "no-role"      // the roles
	WhyTooLong  Code = "why-too-long" // runes, limit
	NoGoal      Code = "no-goal"      // name
	GoalTwice   Code = "goal-twice"   // name, count
	NoLine      Code = "no-line"      // Goal, line key
	EncodeEvent Code = "encode-event" // event type, cause

	// add goal
	EmptyWords    Code = "empty-words"     // no values
	LineNoText    Code = "line-no-text"    // line number
	BaseNotCommit Code = "base-not-commit" // base
	BaseNotHere   Code = "base-not-here"   // base

	// try
	TryWorktreeCommit Code = "try-worktree-commit" // no values
	TryNoCommit       Code = "try-no-commit"       // commit

	// place
	BadExampleKey        Code = "bad-example-key"         // key
	SetNotPlaced         Code = "set-not-placed"          // acts.SetFacts
	SetUnchanged         Code = "set-unchanged"           // Goal
	SetChangedWhileProbe Code = "set-changed-while-probe" // Goal
	KeyTwiceInSet        Code = "key-twice-in-set"        // key
	LineHoldsNoExample   Code = "line-holds-no-example"   // line key, folder
	PlacedFolderGone     Code = "placed-folder-gone"      // key, folder
	NotAnExampleName     Code = "not-an-example-name"     // folder
	ChangedWhileProbe    Code = "changed-while-probe"     // key, from hash, to hash
	H0MovedWhilePlacing  Code = "h0-moved-while-placing"  // Goal, from commit, to commit
	NotAFolder           Code = "not-a-folder"            // path
	FolderOnDisk         Code = "folder-on-disk"          // cause
	NotCommitted         Code = "not-committed"           // path
	DiskNotCommitted     Code = "disk-not-committed"      // path
	DiffersAt            Code = "differs-at"              // path, first differing path
	ObjectMissing        Code = "object-missing"          // hash
	NoExampleTxtar       Code = "no-example-txtar"        // hash
	WrongVerdictOnH0     Code = "wrong-verdict-on-h0"     // acts.KindFacts
	FloorIsExitLine      Code = "floor-is-exit-line"      // op, floor
	EqualsFloorsDiffer   Code = "equals-floors-differ"    // floor, floor
	EqualsLacksContains  Code = "equals-lacks-contains"   // equals floor, contains floor
	FloorHoldsCeiling    Code = "floor-holds-ceiling"     // op, floor, ceiling

	// run
	NoCommit         Code = "no-commit"          // no values
	BadGoalLine      Code = "bad-goal-line"      // the LINE argument
	GoalVanished     Code = "goal-vanished"      // Goal
	NotACommit       Code = "not-a-commit"       // commit
	H0MovedWhileRun  Code = "h0-moved-while-run" // Goal, from commit, to commit
	ExamplesChanged  Code = "examples-changed"   // no values
	ExampleChanged   Code = "example-changed"    // key, from hash, to hash
	VerdictsDiffered Code = "verdicts-differed"  // acts.StableFacts
	SelectNotInLog   Code = "select-not-in-log"  // Goal
	NoResultOnRow    Code = "no-result-on-row"   // row tree
	NoSpec           Code = "no-spec"            // commit, folder
	NoRead           Code = "no-read"            // commit, Goal
	RuleNoVariant    Code = "rule-no-variant"    // commit, rules, variants folder
	NoChart          Code = "no-chart"           // commit, folder

	// events and their payloads
	EventVersion  Code = "event-version"  // version, wanted version
	EventType     Code = "event-type"     // type
	PayloadDecode Code = "payload-decode" // event type, cause
	TrailingData  Code = "trailing-data"  // no values
	SelectRole    Code = "select-role"    // role

	// the example file
	SectionDiffers  Code = "section-differs"   // section name
	PathDiffers     Code = "path-differs"      // path
	DropsFloor      Code = "drops-floor"       // op, literal
	DropsCeiling    Code = "drops-ceiling"     // literal
	AddsNothing     Code = "adds-nothing"      // no values
	TextBeforeFirst Code = "text-before-first" // line
	SecondSection   Code = "second-section"    // line, section name
	ExternalNotName Code = "external-not-name" // line, the external line
	InputOutside    Code = "input-outside"     // line, path
	InputTwice      Code = "input-twice"       // line, path
	UnknownSection  Code = "unknown-section"   // line, section name
	NoSection       Code = "no-section"        // section name

	// git
	GitExit       Code = "git-exit"        // arguments, exit code, stderr
	GitRun        Code = "git-run"         // arguments, cause
	GitNotAllowed Code = "git-not-allowed" // arguments
	LsTreeParse   Code = "ls-tree-parse"   // tree-ish, record
	Submodule     Code = "submodule"       // tree-ish, path
	EntryLeaves   Code = "entry-leaves"    // entry name
	Extract       Code = "extract"         // what, cause
	CatFileBatch  Code = "cat-file-batch"  // tree-ish, cause
	BatchEnds     Code = "batch-ends"      // object
	BatchMissing  Code = "batch-missing"   // object
	BatchRecord   Code = "batch-record"    // record, object
	BatchSize     Code = "batch-size"      // record
	NotRealDir    Code = "not-real-dir"    // path
	ReadLink      Code = "read-link"       // path, cause
	ReadPath      Code = "read-path"       // path, cause
	NotFileOrDir  Code = "not-file-or-dir" // path
	ReadFolder    Code = "read-folder"     // path, cause

	// the store
	NotInRepo    Code = "not-in-repo"    // folder, cause
	ObjectFormat Code = "object-format"  // format
	CreateStore  Code = "create-store"   // cause
	NoStore      Code = "no-store"       // no values
	IDNotULID    Code = "id-not-ulid"    // id
	ActorNotFile Code = "actor-not-file" // actor, the file's actor
	ReadEvents   Code = "read-events"    // cause
	OpenStore    Code = "open-store"     // cause
	LockStore    Code = "lock-store"     // cause
	NoIDAfter    Code = "no-id-after"    // the newest id
	FileOpen     Code = "file-open"      // path, cause
	FileStat     Code = "file-stat"      // path, cause
	FileRead     Code = "file-read"      // path, cause
	FileWrite    Code = "file-write"     // path, cause
	FileSync     Code = "file-sync"      // path, cause
	StoreOutput  Code = "store-output"   // cause
	EventHome    Code = "event-home"     // event type, the home written to

	// the trace
	NoTraceFolder Code = "no-trace-folder" // no values
	CreateTrace   Code = "create-trace"    // cause
	ReadTraces    Code = "read-traces"     // path, cause
	RemoveTrace   Code = "remove-trace"    // path, cause

	// Probes
	ProbeTemp       Code = "probe-temp"         // example hash, cause
	ProbeTxtar      Code = "probe-txtar"        // example hash, cause
	ProbeInput      Code = "probe-input"        // example hash, path, cause
	ProbeMissing    Code = "probe-missing"      // example hash
	ProbeStart      Code = "probe-start"        // example hash, cause
	ProbeTimeout    Code = "probe-timeout"      // example hash, timeout
	ProbeExit       Code = "probe-exit"         // example hash, exit code, stderr
	ProbeWait       Code = "probe-wait"         // example hash, cause
	ProbeNoExitLine Code = "probe-no-exit-line" // example hash

	// guards that read the tables
	NothingPlaced     Code = "nothing-placed"      // Goal, line key
	NotOpen           Code = "not-open"            // Goal, state
	CommitNotHere     Code = "commit-not-here"     // commit
	LeafNeedsCommit   Code = "leaf-needs-commit"   // Goal, commit, cause
	AncestorOfRow     Code = "ancestor-of-row"     // commit, descendant commit, Goal
	OtherGoalPath     Code = "other-goal-path"     // commit, path, Goal
	NoGoalFolder      Code = "no-goal-folder"      // commit, path
	BudgetUsed        Code = "budget-used"         // Goal, trees used, trees allowed
	OtherVerdict      Code = "other-verdict"       // Goal, example keys
	NoGoalToRun       Code = "no-goal-to-run"      // Goal
	NoGoalToPlace     Code = "no-goal-to-place"    // Goal
	NoGoalToSelect    Code = "no-goal-to-select"   // Goal
	RoleCannotSelect  Code = "role-cannot-select"  // role
	NoTree            Code = "no-tree"             // no values
	TreeTwice         Code = "tree-twice"          // tree, count, Goal
	NoRow             Code = "no-row"              // Goal, tree
	OwnerSelectsAmong Code = "owner-selects-among" // Goal, candidates
	RowNotLeaf        Code = "row-not-leaf"        // row tree
	RowNotMet         Code = "row-not-met"         // row tree
	NotGoalsActor     Code = "not-goals-actor"     // Goal, its actor, the act's actor
	NoGoalToReplace   Code = "no-goal-to-replace"  // Goal
	NotWritten        Code = "not-written"         // Goal, state
	TooManyLines      Code = "too-many-lines"      // lines, limit
	BadLineKind       Code = "bad-line-kind"       // line number, line key, kind
	NoGapLine         Code = "no-gap-line"         // no values
	HasSelect         Code = "has-select"          // Goal
	NotOwnerPlace     Code = "not-owner-place"     // role
	BadExamplesFolder Code = "bad-examples-folder" // folder
	CopyExamples      Code = "copy-examples"       // from, to, cause
	TryDraftOnH0      Code = "try-draft-on-h0"     // no values
	BadDraftName      Code = "bad-draft-name"      // name
	NoDraft           Code = "no-draft"            // name, folder
	ExampleOffLine    Code = "example-off-line"    // example key
	AddedVerdictWrong Code = "added-verdict-wrong" // example key, verdict, verdict needed
	ReplaceExamples   Code = "replace-examples"    // Goal
	AlreadyCurrent    Code = "already-current"     // hash, key, version
	EarlierVersion    Code = "earlier-version"     // hash, version, key
	Ended             Code = "ended"               // Goal
	NotInSnapshot     Code = "not-in-snapshot"     // Goal, end event id
	NotOwnerClose     Code = "not-owner-close"     // role
	NoGoalToClose     Code = "no-goal-to-close"    // Goal
	NotOwner          Code = "not-owner"           // role
	NoStop            Code = "no-stop"             // Goal
	NoGoalToLand      Code = "no-goal-to-land"     // Goal
	StopNotHeld       Code = "stop-not-held"       // stop, Goal
	FromNotEnded      Code = "from-not-ended"      // Goal, state
	NotOwnerUnselect  Code = "not-owner-unselect"  // role
	NoSelect          Code = "no-select"           // Goal, state
	NoGoalToUnselect  Code = "no-goal-to-unselect" // Goal
	NoNewCell         Code = "no-new-cell"         // Goal, unselect id
	CommitUnselected  Code = "commit-unselected"   // commit, Goal
	LandNotOpen       Code = "land-not-open"       // Goal, state
	NotTheSelect      Code = "not-the-select"      // commit, Goal, selected commit
	EvidenceMissing   Code = "evidence-missing"    // commit, example key
	NotMetOnCommit    Code = "not-met-on-commit"   // example key, commit, verdict
	NotSelectedOn     Code = "not-selected-on"     // Goal, commit, selected commit or empty
	AlreadyUnselected Code = "already-unselected"  // Goal, commit

	// findings
	NoGoalForFinding     Code = "no-goal-for-finding"     // Goal
	NoFinding            Code = "no-finding"              // finding
	FindingTwice         Code = "finding-twice"           // finding, count
	NoIssue              Code = "no-issue"                // no values
	NoGoalToFold         Code = "no-goal-to-fold"         // Goal
	FoldIntoEnded        Code = "fold-into-ended"         // Goal
	FoldGoalAndIssue     Code = "fold-goal-and-issue"     // no values
	Refuted              Code = "refuted"                 // finding
	Declined             Code = "declined"                // finding
	NotOwnerDecline      Code = "not-owner-decline"       // role
	DeclineNoWhy         Code = "decline-no-why"          // no values
	DeclineFolded        Code = "decline-folded"          // finding, Goal
	DeclineFoldedIssue   Code = "decline-folded-issue"    // finding, issue
	FoldNoFinding        Code = "fold-no-finding"         // no values
	NamedInWork          Code = "named-in-work"           // finding, Goal
	NotCovered           Code = "not-covered"             // Goal, finding
	EmptyFact            Code = "empty-fact"              // no values
	EmptyEvidence        Code = "empty-evidence"          // no values
	DraftNotAddGoal      Code = "draft-not-add-goal"      // no values
	ExampleNoCommit      Code = "example-no-commit"       // no values
	BadExampleRef        Code = "bad-example-ref"         // example
	ExampleCommitNotHere Code = "example-commit-not-here" // commit
	NoExampleAtCommit    Code = "no-example-at-commit"    // example, commit, path
)
