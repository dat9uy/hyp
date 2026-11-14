package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dat9uy/hyp/internal/check"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/sandbox"
	"github.com/dat9uy/hyp/internal/store"
)

// recipeHash is the --json output of hyp bundle hash.
type recipeHash struct {
	Ref     string `json:"ref"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
}

const recipeUse = "usage: hyp bundle hash DIR [--json] | hyp bundle check | hyp bundle new NAME@1 | hyp bundle new NAME@N --from NAME@M | hyp bundle run --record DIR REF NAME [ARGS...] | hyp bundle evidence --record DIR REF OUT_DIR | hyp bundle known-answers REF"

// cmdRecipe dispatches the hyp bundle subcommands.
func cmdBundle(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return usageErr(recipeUse)
	}
	switch args[0] {
	case "-h", "--help", "help":
		return helpText(recipeUse)
	case "hash":
		return cmdRecipeHash(args[1:], stdout, stderr)
	case "check":
		return cmdRecipeCheck(args[1:], stdout, stderr)
	case "new":
		return cmdRecipeNew(args[1:], stdout, stderr)
	case "run":
		return cmdRecipeRun(args[1:], stdout, stderr)
	case "evidence":
		return cmdRecipeEvidence(args[1:], stdout, stderr)
	case "known-answers":
		return cmdRecipeKnownAnswers(args[1:], stdout, stderr)
	}
	return usageErr(recipeUse)
}

// cmdRecipeHash handles hyp bundle hash DIR: the hash hyp add exp --recipe
// records for the bundle in DIR, which must open as a recipe.
func cmdRecipeHash(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("bundle hash", stderr)
	asJSON := fs.Bool("json", false, "print ref, name, version, path and sha256 as JSON")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr(recipeUse)
	}
	// Absolute, so "." from inside a bundle has the bundle's name.
	dir, err := filepath.Abs(pos[0])
	if err != nil {
		return herr.Env("%v", err)
	}
	r, err := recipe.OpenToHash(dir)
	if err != nil {
		// A DIR that holds no recipe.md names no bundle at all.
		if _, serr := os.Stat(filepath.Join(dir, "recipe.md")); os.IsNotExist(serr) {
			return herr.BadArgs("%s", herr.Text(err))
		}
		return err
	}
	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(recipeHash{Ref: r.Ref().String(), Name: r.Name, Version: r.Version, Path: r.Path, SHA256: r.Hash})
	}
	fmt.Fprintln(stdout, r.Hash)
	return nil
}

// cmdRecipeCheck handles hyp bundle check: every pin to a versioned bundle in
// every store under the repository root must still match the bundle on disk.
func cmdRecipeCheck(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("bundle check", stderr)
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 0 {
		return usageErr("usage: hyp bundle check")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	report, failures, hinted, err := check.Recipes(cwd)
	if errors.Is(err, check.ErrNoRepo) {
		return herr.Env("hyp bundle check runs inside a git repository")
	}
	if err != nil {
		return herr.Env("%v", err)
	}
	for _, line := range report {
		fmt.Fprintln(stdout, line)
	}
	if failures > 0 && hinted == failures {
		return herr.Refused("recipe check found %d failures; rename each folder a hint names", failures)
	}
	if failures > 0 {
		return herr.Refused("recipe check found %d failures; a changed bundle needs a new version (hyp bundle new)", failures)
	}
	return nil
}

// cmdRecipeNew handles hyp bundle new NAME@1, which writes a first version
// from the embedded template beside the nearest store, and hyp bundle new
// NAME@N --from NAME@M, which copies the versioned bundle NAME@M to the next
// version, recording where it came from.
func cmdRecipeNew(args []string, stdout, stderr io.Writer) error {
	const use = "usage: hyp bundle new NAME@1 [--rubric] | hyp bundle new NAME@N --from NAME@M"
	fs := readFlags("bundle new", stderr)
	from := fs.String("from", "", "the version to copy, NAME@M; omitted for a first version NAME@1")
	rubric := fs.Bool("rubric", false, "write a first rubric into rubrics/ instead of a first recipe into recipes/")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr(use)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if *from == "" {
		return newFirst(cwd, pos[0], *rubric, stdout)
	}
	if *rubric {
		return usageErr("--rubric names the library of a first version; --from copies NAME@M within its own library")
	}
	dir, stamped, err := recipe.NewVersion(cwd, pos[0], *from)
	if err != nil {
		return err
	}
	if _, err := recipe.Open(dir); err != nil {
		fmt.Fprintf(stderr, "warn: the copy does not open (%s); fix it before using it\n", herr.Text(err))
	}
	fmt.Fprintln(stdout, dir)
	fmt.Fprintln(stdout, "from: "+stamped)
	return nil
}

// newFirst writes the first version target from the embedded template into
// recipes/ beside the nearest store above cwd and prints its directory.
func newFirst(cwd, target string, rubric bool, stdout io.Writer) error {
	// A bad ref is a bad argument wherever it is typed, store or not.
	if _, err := recipe.FirstRef(target); err != nil {
		return err
	}
	st, err := store.Open(cwd)
	if err != nil {
		return err
	}
	root := st.Dir
	write := recipe.NewFirst
	if rubric {
		write = recipe.NewFirstRubric
	}
	dir, err := write(root, target)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, dir)
	return nil
}

// cmdRecipeRun handles hyp bundle run --record DIR REF NAME [ARGS...]: run
// REF's Standalone command NAME with ARGS appended, with no sandbox.
func cmdRecipeRun(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("bundle run", stderr)
	record := fs.String("record", "", "new directory for the command trace (required)")
	pos, err := parseHead(fs, args)
	if err != nil {
		return err
	}
	if *record == "" || len(pos) < 2 {
		return usageErr("usage: hyp bundle run --record DIR REF NAME [ARGS...]")
	}
	return runOutside(stdout, sandbox.RunOptions{Ref: pos[0], Step: recipe.Standalone, Label: pos[1], Args: pos[2:], Record: *record, Warn: stderr})
}

// cmdRecipeEvidence handles hyp bundle evidence --record DIR REF OUT_DIR:
// run REF's Evidence commands on the files hyp sandbox remove kept in
// OUT_DIR, whose absolute path each command gets as its last argument.
func cmdRecipeEvidence(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("bundle evidence", stderr)
	record := fs.String("record", "", "new directory for the command trace (required)")
	pos, err := parseHead(fs, args)
	if err != nil {
		return err
	}
	if *record == "" || len(pos) != 2 {
		return usageErr("usage: hyp bundle evidence --record DIR REF OUT_DIR")
	}
	kept, err := filepath.Abs(pos[1])
	if err != nil {
		return usageErr("%v", err)
	}
	return runOutside(stdout, sandbox.RunOptions{Ref: pos[0], Step: recipe.Evidence, Args: []string{kept}, Record: *record, Warn: stderr})
}

// cmdRecipeKnownAnswers handles hyp bundle known-answers REF: run every
// known-answer case of the bundle REF with no sandbox, one line per case. It
// exits 3 when a case timed out, 1 when one failed or an instrument has no
// cases, and 0 when every case is ok.
func cmdRecipeKnownAnswers(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("bundle known-answers", stderr)
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("usage: hyp bundle known-answers REF")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	ref := pos[0]
	r, err := recipe.LoadAny(cwd, ref)
	if err != nil {
		return err
	}
	if len(r.Evidence) == 0 && len(r.Standalone) == 0 {
		fmt.Fprintf(stdout, "bundle %s has no Evidence or Standalone command; no known answers to run\n", ref)
		return nil
	}
	hyp, err := sandbox.NewSandboxHyp()
	if err != nil {
		return err
	}
	defer hyp.Remove()
	ka, err := sandbox.RunKnownAnswers(r, hyp, stderr)
	if err != nil {
		return err
	}
	if len(ka.Cases) == 0 {
		return herr.Refused("recipe %s is an instrument with no known answers; add known-answers/NAME/ cases (RECIPE-FORMAT.md, Known answers)", ref)
	}
	failed, timedOut := 0, false
	for _, c := range ka.Cases {
		if c.Outcome == "ok" {
			fmt.Fprintln(stdout, "ok "+c.Name)
			continue
		}
		fmt.Fprintf(stdout, "FAIL %s: %s\n", c.Name, c.Outcome)
		failed++
		timedOut = timedOut || strings.HasPrefix(c.Outcome, "timed out")
	}
	switch {
	case timedOut:
		return herr.Env("recipe %s: a known answer timed out; %d of %d cases failed", ref, failed, len(ka.Cases))
	case failed > 0:
		return herr.Refused("recipe %s: %d of %d known answers failed", ref, failed, len(ka.Cases))
	}
	return nil
}

// runOutside runs o from the working directory and, when a command ran,
// prints its log and where the record is.
func runOutside(stdout io.Writer, o sandbox.RunOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	o.Dir = cwd
	rec, err := sandbox.Run(o)
	if rec != nil {
		if log, rerr := os.ReadFile(filepath.Join(o.Record, "command.log")); rerr == nil {
			stdout.Write(log)
		}
		fmt.Fprintln(stdout, "record -> "+o.Record)
	}
	return err
}

// parseHead reads flags only before the first positional and returns
// everything from it on as given, so a script's own flags in ARGS are never
// read by hyp. (parse re-parses after every positional and would reject
// them.)
func parseHead(fs *flag.FlagSet, args []string) ([]string, error) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			var b strings.Builder
			fs.SetOutput(&b)
			fmt.Fprintf(&b, "usage of hyp %s:\n", fs.Name())
			fs.PrintDefaults()
			return nil, helpText(strings.TrimSuffix(b.String(), "\n"))
		}
		return nil, usageErr("%v", err)
	}
	return fs.Args(), nil
}
