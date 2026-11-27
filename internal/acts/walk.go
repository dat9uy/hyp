package acts

import (
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
)

// A walk is what the kept model says of an example's run section on a
// commit: each command of the run section, in order, with the step of the
// kept model it takes there. A step is a definition of the kept model tagged
// with the command it takes (\* command: X); a command takes the step whose
// command its words start with, the longest when two do. A command that
// takes no step, such as echo, reads "no step". hyp calculates every walk
// from the example and the kept model at the commit; no one writes it.

// WalkStep is one command of a run section and the name of the step of the
// kept model it takes, or "" when it takes none.
type WalkStep struct {
	Command string
	Step    string
}

// Walk is the walk of one run section on one commit. Kept is false when the
// commit keeps no model in goals/model/: then the walk has no steps.
type Walk struct {
	Kept  bool
	Steps []WalkStep
}

// keptStep is a step of the kept model: its name and the words of its
// command.
type keptStep struct {
	name  string
	words []string
}

// keptSteps is every step of the kept model at commit, in module order, and
// whether the commit keeps a model: a .tla file in goals/model/ with a
// definition. A rule, which a check config checks, is no step.
func keptSteps(root, commit string) (bool, []keptStep, error) {
	rules, defs, order, err := keptRules(root, commit)
	if err != nil {
		return false, nil, err
	}
	var steps []keptStep
	for _, name := range order {
		d := defs[name]
		if d.Command == "" || slices.Contains(rules, name) {
			continue
		}
		steps = append(steps, keptStep{name: name, words: strings.Fields(d.Command)})
	}
	return len(order) > 0, steps, nil
}

// KeepsModel reports whether commit keeps a model in goals/model/.
func KeepsModel(root, commit string) (bool, error) {
	kept, _, err := keptSteps(root, commit)
	return kept, err
}

// WalkRun is the walk of the run section run on commit.
func WalkRun(root, commit, run string) (Walk, error) {
	kept, steps, err := keptSteps(root, commit)
	if err != nil || !kept {
		return Walk{}, err
	}
	w := Walk{Kept: true}
	for _, cmd := range RunCommands(run) {
		words := commandWords(cmd)
		best, n := "", 0
		for _, s := range steps {
			if len(s.words) > n && len(s.words) <= len(words) && slices.Equal(words[:len(s.words)], s.words) {
				best, n = s.name, len(s.words)
			}
		}
		w.Steps = append(w.Steps, WalkStep{Command: cmd, Step: best})
	}
	return w, nil
}

// WalkExample is the walk of the run section of the example folder with
// tree hash hash, on commit.
func WalkExample(root, commit, hash string) (Walk, error) {
	ex, _, err := readExample(root, hash)
	if err != nil {
		return Walk{}, err
	}
	return WalkRun(root, commit, ex.Run)
}

// exampleRef is the form of an example a finding names: a Goal id, a line
// key and an example name, as G-XXXXXX/line-N/name.
var exampleRef = regexp.MustCompile(`^G-[0-9A-Z]+/line-[1-9][0-9]*/[a-z0-9][a-z0-9-]*$`)

// ExampleFolder is the folder of the example ref names, as git names it.
func ExampleFolder(ref string) string { return path.Join("goals", ref) }

// ExampleAt resolves the example ref and the commit rev a finding names: the
// commit's full sha and the tree hash of the example folder in it. A ref
// that is not of the form G-XXXXXX/line-N/name, or a rev that names no
// commit of this clone, is a bad argument; a commit that does not hold the
// example's example.txtar is a refusal.
func ExampleAt(root, ref, rev string) (commit, hash string, err error) {
	if !exampleRef.MatchString(ref) {
		return "", "", fault.BadArg(fault.BadExampleRef, ref)
	}
	if rev == "" || strings.HasPrefix(rev, "-") {
		return "", "", fault.BadArg(fault.ExampleCommitNotHere, rev)
	}
	ok, err := gitread.HasCommit(root, rev)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", fault.BadArg(fault.ExampleCommitNotHere, rev)
	}
	commit, err = gitread.CommitOf(root, rev)
	if err != nil {
		return "", "", err
	}
	dir := ExampleFolder(ref)
	code, _, err := gitread.GitExit(root, "cat-file", "-e", commit+":"+dir+"/example.txtar")
	if err != nil {
		return "", "", err
	}
	if code != 0 {
		return "", "", fault.Refuse(fault.NoGuard, fault.NoExampleAtCommit, ref, commit, dir+"/example.txtar")
	}
	out, err := gitread.Git(root, "rev-parse", "--verify", commit+":"+dir)
	if err != nil {
		return "", "", err
	}
	return commit, strings.TrimSpace(string(out)), nil
}

// RunCommands is each command of a run section, in order, as its text reads
// with the spaces around it removed: the run section split at each newline
// and each ;, &, &&, || and | outside quotes. A line comment and an empty
// command are left out. A & right after < or > is part of a redirection.
func RunCommands(run string) []string {
	var cmds []string
	var cur strings.Builder
	flush := func() {
		if c := strings.TrimSpace(cur.String()); c != "" && !strings.HasPrefix(c, "#") {
			cmds = append(cmds, c)
		}
		cur.Reset()
	}
	var quote byte
	for i := 0; i < len(run); i++ {
		c := run[i]
		switch {
		case quote != 0:
			if c == '\\' && quote == '"' && i+1 < len(run) {
				cur.WriteByte(c)
				i++
				c = run[i]
			} else if c == quote {
				quote = 0
			}
			cur.WriteByte(c)
		case c == '\\' && i+1 < len(run):
			cur.WriteByte(c)
			i++
			cur.WriteByte(run[i])
		case c == '\'' || c == '"':
			quote = c
			cur.WriteByte(c)
		case c == '#' && strings.TrimSpace(cur.String()) == "":
			for i < len(run) && run[i] != '\n' {
				i++
			}
			flush()
		case c == '\n' || c == ';':
			flush()
		case c == '|':
			flush()
			if i+1 < len(run) && run[i+1] == '|' {
				i++
			}
		case c == '&':
			s := cur.String()
			if len(s) > 0 && (s[len(s)-1] == '>' || s[len(s)-1] == '<') {
				cur.WriteByte(c)
				continue
			}
			flush()
			if i+1 < len(run) && run[i+1] == '&' {
				i++
			}
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return cmds
}

// assignment is a shell word that sets a variable for the command after it.
var assignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// commandWords is the words of one command with their quotes removed, after
// the variable assignments that lead it.
func commandWords(cmd string) []string {
	var words []string
	var cur strings.Builder
	in := false
	var quote byte
	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
		case c == '\'' || c == '"':
			quote, in = c, true
		case c == ' ' || c == '\t':
			if in {
				words = append(words, cur.String())
				cur.Reset()
				in = false
			}
		default:
			cur.WriteByte(c)
			in = true
		}
	}
	if in {
		words = append(words, cur.String())
	}
	for len(words) > 0 && assignment.MatchString(words[0]) {
		words = words[1:]
	}
	return words
}
