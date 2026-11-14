// Package cli implements the hyp command line.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/record"
	"github.com/dat9uy/hyp/internal/sandbox"
	"github.com/dat9uy/hyp/internal/store"
)

// versionLine is hyp's version, followed by the commit Go stamped into this
// build (first 12 hex digits, -dirty for a checkout with uncommitted
// changes). A build without VCS info (go test, go run, -buildvcs=false)
// prints the bare line.
func versionLine() string {
	line := "hyp 0.1.0 (events v1)"
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

Every write command needs --why and an actor (--actor or $HYP_ACTOR).
Statuses are derived from events; no command can set one.

setup:
  init                                   create a store in the current directory; in a git repository, start its record
  version                                print the hyp version

record (in a git repository, every store's events and the files they name, kept on a hidden ref no branch holds):
  sync [REMOTE]                          join the record with REMOTE's (origin by default) and push the join
  import [REF:]PATH...                   add each work-tree store's events the record lacks, from the checkout or commit REF

write:
  add goal --title T [--body B] [--alias A] [--line L]... [--cite N=KEY]...
  add dq   --parent REF --title T [--body B] [--alias A]
  add hyp  --parent REF --title T [--body B] [--alias A] [--subject PATH]...
  add exp  --parent REF --title T --recipe NAME@V|none [--rubric NAME@V|reading] [--records PATH]... [--target REF|none]... [--alias A]
  record   --exp REF [--verdict REF=supports|contradicts|inconclusive]... --observation O [--evidence P]... [--limit L]... [--closes R#N]...
  retract REF | abandon REF | reaffirm REF
  supersede NEW --old OLD
  amend HYP --body B                     revise a Hypothesis body
  subject HYP [PATH...] [--base C]       tie a Hypothesis to the base and files that embody its claim
  link FROM depends_on|external_ref TO   unlink FROM TYPE TO
  select DQ [HYP]                        admitted only when hyp status owes it (no HYP: print the suggestion only)
  unselect DQ | defer DQ | withdraw DQ|EXP | resume DQ
  checkpoint [--summary S] [--goal G] [--refers K[=TEXT]]...   the runner's report: a summary, refers, or both
  answer [--answer R#rN=TEXT]... [--accept K=TEXT]... [--returns K=TEXT]...   the owner's act; no --why

read (REF/TIME: an event id, a node ref, or an RFC3339 time):
  status [--as-of REF|TIME] [--max-lines N] [--live]   orientation screen, 60 lines by default; --live hides contradicted rivals
  show REF|R-id#n|G-id#n|FILE [--as-of REF|TIME]  one node, or one limit or line, in full, with its history; or the bytes of a FILE an event names
  diff --since REF|TIME                       status changes and events since a point
  render --markdown [--as-of REF|TIME]        the whole tree as Markdown, untruncated
  check                                       rejected events, bad lines, warnings
  signals [--json]                            design marks for the whole tree: one-rival, D1
  events --json                               applied events in merge order, and rejected ones, as JSON for scripts
  prompt G-id --actor A [--branch B] [--pr N] [--experiments N]   the /goal prompt for a Goal, from its body's constraints block

bundle (a recipe in recipes/ or a rubric in rubrics/; a name lives in one library):
  bundle hash DIR [--json]               the bundle hash add exp --recipe or --rubric records for DIR
  bundle check                           every pin to a versioned bundle, in every tree under the repo root, still matches it
  bundle new NAME@1 [--rubric]           write a first recipe, or with --rubric a first rubric, beside the nearest store from hyp's template; fill every hyp:placeholder
  bundle new NAME@N --from NAME@M        copy bundle NAME@M to the next version NAME@N in its library, writing from:
  bundle run --record DIR REF NAME [ARGS...]   run one Standalone command with no sandbox; exits with its exit code
  bundle evidence --record DIR REF OUT_DIR     run recipe REF's Evidence commands on remove's OUT_DIR; exits with the failing one's code
  bundle known-answers REF                     run every known-answers/NAME/ case of bundle REF, twice each; 1 on a wrong answer

sandbox (a disposable clone under ~/.hyp-sandbox/ID for one Experiment):
  sandbox create ID --host H --base REV --recipe NAME@VERSION|none [--recipe NAME@VERSION]... [--recipes DIR] [--store SRC --cut EVENT] [--inputs FILE] [--hyp BIN]
  sandbox drive ID                             run every recipe's Drive commands in the sandbox; exits with the failing one's code
  sandbox remove ID OUT_DIR [--keep REL]... | sandbox remove ID --discard
`

type command func(args []string, stdout, stderr io.Writer) error

var commands = map[string]command{
	"init":       cmdInit,
	"add":        cmdAdd,
	"record":     cmdRecord,
	"retract":    cmdOnNode("retract", event.TypeResultRetract),
	"abandon":    cmdOnNode("abandon", event.TypeHypothesisAbandon),
	"reaffirm":   cmdOnNode("reaffirm", event.TypeReaffirm),
	"defer":      cmdOnNode("defer", event.TypeDQDefer),
	"withdraw":   cmdWithdraw,
	"resume":     cmdOnNode("resume", event.TypeDQResume),
	"supersede":  cmdSupersede,
	"amend":      cmdAmend,
	"subject":    cmdSubject,
	"link":       cmdLink("link", event.TypeEdgeAdd),
	"unlink":     cmdLink("unlink", event.TypeEdgeRemove),
	"select":     cmdSelect,
	"unselect":   cmdUnselect,
	"confirm":    cmdConfirm,
	"checkpoint": cmdCheckpoint,
	"answer":     cmdAnswer,
	"status":     cmdStatus,
	"show":       cmdShow,
	"diff":       cmdDiff,
	"render":     cmdRender,
	"check":      cmdCheck,
	"signals":    cmdSignals,
	"events":     cmdEvents,
	"prompt":     cmdPrompt,
	"sandbox":    cmdSandbox,
	"bundle":     cmdBundle,
	"sync":       cmdSync,
	"import":     cmdImport,
}

// aliases maps an older command name to the one that replaced it. An alias
// runs but is not listed, so frozen bundles whose scripts call it still run
// and nothing new names it (ADR 0018: hyp recipe became hyp bundle).
var aliases = map[string]string{"recipe": "bundle"}

// Commands lists every command name Run dispatches, sorted.
func Commands() []string {
	names := []string{"version", "help"}
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Run executes one hyp invocation and returns the process exit code: 0
// success, 1 refused, 2 usage error, 3 I/O or internal error. It is the only
// place a herr kind becomes an exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, versionLine())
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	}
	name := args[0]
	if a, ok := aliases[name]; ok {
		name = a
	}
	cmd, ok := commands[name]
	if !ok {
		fmt.Fprint(stderr, usage)
		return 2
	}
	err := cmd(args[1:], stdout, stderr)
	if err == nil {
		return 0
	}
	if h, ok := errors.AsType[helpText](err); ok {
		fmt.Fprintln(stdout, string(h))
		return 0
	}
	fmt.Fprintln(stderr, err)
	// A recipe script's own exit code passes through unchanged (sandbox spec §1.3).
	if cf, ok := errors.AsType[*sandbox.CommandFailed](err); ok {
		return *cf.Row.Exit
	}
	switch herr.KindOf(err) {
	case herr.Refusal:
		return 1
	case herr.BadArguments:
		return 2
	default:
		return 3
	}
}

func cmdInit(args []string, stdout, stderr io.Writer) error {
	const use = "usage: hyp init"
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		return helpText(use)
	}
	if len(args) != 0 {
		return usageErr(use)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := store.Init(cwd); err != nil {
		if herr.KindOf(err) == herr.Refusal {
			return err
		}
		return herr.Env("%v", err)
	}
	fmt.Fprintf(stdout, "initialised %s in %s\n", record.TreeDir, cwd)
	return nil
}
