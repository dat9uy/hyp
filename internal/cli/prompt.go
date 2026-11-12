package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/store"
)

// The three headings of a Goal body's constraints block: what a run may
// change, what it may only read, and what it never does.
var blockHeadings = []string{"May change:", "Read only:", "Never:"}

const promptUse = "usage: hyp prompt G-id --actor A [--branch B] [--pr N] [--experiments N]"

// constraintsBlock finds the constraints block of a Goal body: the lines from
// the first line that starts with a heading to the end of the body. It
// returns the block's lines with trailing empty lines removed, and the
// problems that make it invalid: no block, a missing heading, or a non-empty
// line that starts with no heading. Lines split on \n, and a line is empty
// only when it holds no byte, so a line of spaces, a tab or a lone \r (the
// blank line of a CRLF body) is a bad line. found reports whether any line
// starts with a heading.
func constraintsBlock(body string) (block []string, problems []string, found bool) {
	lines := strings.Split(body, "\n")
	start := -1
	for i, l := range lines {
		if headingOf(l) != "" {
			start = i
			break
		}
	}
	if start < 0 {
		return nil, []string{`the Goal body has no constraints block: no line starts with "May change:", "Read only:" or "Never:"`}, false
	}
	block = lines[start:]
	for len(block) > 0 && block[len(block)-1] == "" {
		block = block[:len(block)-1]
	}
	seen := map[string]bool{}
	for _, l := range block {
		if l == "" {
			continue
		}
		h := headingOf(l)
		if h == "" {
			problems = append(problems, fmt.Sprintf("the constraints block holds a line that starts with no heading: %q", l))
			continue
		}
		seen[h] = true
	}
	for _, h := range blockHeadings {
		if !seen[h] {
			problems = append(problems, fmt.Sprintf("the constraints block has no %q line", h))
		}
	}
	return block, problems, true
}

func headingOf(line string) string {
	for _, h := range blockHeadings {
		if strings.HasPrefix(line, h) {
			return h
		}
	}
	return ""
}

// blockWarning is the hyp add goal warning for a body that has a heading but
// no valid constraints block; "" when the body has no heading or a valid one.
func blockWarning(body string) string {
	_, problems, found := constraintsBlock(body)
	if !found || len(problems) == 0 {
		return ""
	}
	return "warn: hyp prompt will refuse this Goal: " + strings.Join(problems, "; ")
}

// repoRoot is the nearest directory at or above dir that holds .git; ok is
// false when none does.
func repoRoot(dir string) (root string, ok bool) {
	for d := dir; ; {
		if _, err := os.Stat(filepath.Join(d, ".git")); err == nil {
			return d, true
		}
		up := filepath.Dir(d)
		if up == d {
			return "", false
		}
		d = up
	}
}

// cmdPrompt prints the /goal prompt for a Goal, derived from the log: the
// store's path, the actor, the Goal, the hyp-goal-run skill, the stop clauses
// and the Goal body's constraints block. It reads no owner answer, so the
// same Goal node and flags give the same bytes in any store.
func cmdPrompt(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("prompt", stderr)
	actor := fs.String("actor", "", "the runner's actor name (required; $HYP_ACTOR is not read)")
	branch := fs.String("branch", "", "the branch the run commits on")
	pr := fs.Int("pr", 0, "the number of the PR the run updates")
	experiments := fs.Int("experiments", 7, "stop after this many Results recorded under the actor")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr(promptUse)
	}
	if *actor == "" {
		return usageErr("missing --actor: hyp prompt does not read $HYP_ACTOR")
	}
	if !store.ValidActor(*actor) {
		return usageErr("invalid actor %q (want ^[a-z0-9][a-z0-9-]{0,39}$)", *actor)
	}
	if *experiments < 1 {
		return usageErr("--experiments must be at least 1")
	}
	if *pr < 0 {
		return usageErr("--pr must be a PR number")
	}
	if strings.ContainsFunc(*branch, unicode.IsSpace) || strings.ContainsFunc(*branch, unicode.IsControl) {
		return usageErr("--branch %q holds a space or a control character, which would break the prompt's lines", *branch)
	}
	r, err := openReader()
	if err != nil {
		return err
	}
	id, err := resolveRef(r.full, pos[0])
	if err != nil {
		return usageErr("%v", err)
	}
	n := r.full.Nodes[id]
	if n.Kind != event.KindGoal {
		return usageErr("hyp prompt takes a Goal: %s is a %s", pos[0], n.Kind)
	}
	g := event.DisplayID(event.KindGoal, id)
	block, problems, _ := constraintsBlock(n.Body)
	if len(problems) > 0 {
		return herr.Refused("hyp prompt %s: %s", g, strings.Join(problems, "; "))
	}
	// The working directory may reach the store through a symbolic link;
	// the store's place in its repository is read from its real path.
	storeDir, err := filepath.EvalSymlinks(r.root)
	if err != nil {
		return herr.Env("%v", err)
	}
	repo, ok := repoRoot(storeDir)
	if !ok {
		return herr.Env("hyp prompt %s: no repository holds the store %s, so its path relative to the repository root is not known", g, storeDir)
	}
	rel, err := filepath.Rel(repo, storeDir)
	if err != nil {
		return herr.Env("%v", err)
	}
	rel = filepath.ToSlash(rel)

	var b strings.Builder
	fmt.Fprintf(&b, "/goal Run the hyp tree in %s with the hyp-goal-run skill, as HYP_ACTOR=%s. Its Goal is %s: read it with hyp show %s, read hyp status %s, and name --goal %s on every hyp checkpoint.\n", rel, *actor, g, g, g, g)
	fmt.Fprintf(&b, "Done when the latest printed hyp status %s shows \"runner owes nothing: stop and report to the owner\" under OWED, or %d Results have been recorded under %s, or you cannot take the next step honestly and say why.\n", g, *experiments, *actor)
	for _, l := range block {
		b.WriteString(l)
		b.WriteString("\n")
	}
	if *branch != "" {
		fmt.Fprintf(&b, "Commit on the branch %s.\n", *branch)
	}
	if *pr > 0 {
		fmt.Fprintf(&b, "The run's PR is #%d.\n", *pr)
	}
	fmt.Fprint(stdout, b.String())
	return nil
}
