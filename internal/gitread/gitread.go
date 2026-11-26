// Package gitread is the only way hyp runs git. Every command passes one
// allow-list that checks the whole argument list, so hyp can read a
// repository and cannot write one.
package gitread

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/fault"
)

// Git runs one read-only git command in dir and returns its stdout. A command
// that is not on hyp's read list, or that exits non-zero, is an environment
// error; the second holds git's stderr.
func Git(dir string, args ...string) ([]byte, error) {
	return GitStdin(dir, nil, args...)
}

// GitStdin is Git for a command that reads stdin, such as cat-file --batch.
// It passes the same allow-list and runs with the same environment.
func GitStdin(dir string, stdin []byte, args ...string) ([]byte, error) {
	code, stdout, stderr, err := run(dir, stdin, args)
	if err != nil {
		return nil, err
	}
	if code != 0 {
		return nil, fault.EnvFault(fault.GitExit, args, code, strings.TrimSpace(stderr))
	}
	return stdout, nil
}

// GitExit is Git for a caller that reads the exit code, such as
// merge-base --is-ancestor or cat-file -e. A non-zero exit is returned as the
// code with no error; only a command that is off the read list or cannot run
// is an error. The bytes are stdout.
func GitExit(dir string, args ...string) (int, []byte, error) {
	code, stdout, _, err := run(dir, nil, args)
	return code, stdout, err
}

func run(dir string, stdin []byte, args []string) (code int, stdout []byte, stderr string, err error) {
	if err := allowed(args); err != nil {
		return 0, nil, "", err
	}
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	// A read must not take an optional lock, which would touch the index, nor
	// fetch a missing object in a partial clone, which would write a pack.
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0", "GIT_NO_LAZY_FETCH=1")
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	switch err := cmd.Run(); {
	case err == nil:
		return 0, out.Bytes(), errb.String(), nil
	default:
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode(), out.Bytes(), errb.String(), nil
		}
		return 0, nil, "", fault.EnvFault(fault.GitRun, args, err)
	}
}

// allowed checks the whole argument list against the forms hyp reads with,
// never only the first word. Anything else, a write command or a flag that
// makes a read write (--output, -w, --remote), is refused.
func allowed(args []string) error {
	if len(args) > 0 && permits(args[0], args[1:]) {
		return nil
	}
	return fault.EnvFault(fault.GitNotAllowed, args)
}

// permits reports whether cmd with rest is one of hyp's read forms.
func permits(cmd string, rest []string) bool {
	switch cmd {
	case "rev-parse":
		pos, ok := split(rest, "--verify", "--show-toplevel", "--show-object-format", "--path-format=absolute", "--git-common-dir")
		return ok && len(rest) > 0 && noDash(pos)
	case "cat-file":
		if len(rest) == 1 && rest[0] == "--batch" {
			return true
		}
		return len(rest) == 2 && contains([]string{"-e", "-t", "-p", "blob"}, rest[0]) && noDash(rest[1:])
	case "ls-tree":
		pos, ok := split(rest, "-z", "-r")
		return ok && len(pos) >= 1 && noDash(pos)
	case "merge-base":
		// --is-ancestor A B, or the merge base of two commits
		return (len(rest) == 3 && rest[0] == "--is-ancestor" && noDash(rest[1:])) ||
			(len(rest) == 2 && noDash(rest))
	case "diff-tree":
		pos, ok := split(rest, "--name-only", "-r", "-z", "-c")
		return ok && (len(pos) == 1 || len(pos) == 2) && noDash(pos)
	case "for-each-ref":
		// one fixed form: the first branch or tag that contains a commit
		return len(rest) == len(refsForm)+1 && slices.Equal(rest[:3], refsForm[:3]) &&
			noDash(rest[3:4]) && slices.Equal(rest[4:], refsForm[3:])
	case "ls-files":
		// one fixed form: the files of the checkout a commit of every
		// change would hold
		return slices.Equal(rest, lsFilesForm)
	}
	return false
}

// split separates the leading flags of rest from its positional tail. It
// reports false when a leading flag is not in allowedFlags.
func split(rest []string, allowedFlags ...string) (pos []string, ok bool) {
	for i, a := range rest {
		if !strings.HasPrefix(a, "-") {
			return rest[i:], true
		}
		if !contains(allowedFlags, a) {
			return nil, false
		}
	}
	return nil, true
}

// noDash reports that no argument is empty or starts with a dash, so a flag
// cannot hide among the positionals.
func noDash(ss []string) bool {
	for _, s := range ss {
		if s == "" || strings.HasPrefix(s, "-") {
			return false
		}
	}
	return true
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// refsForm is the one for-each-ref form hyp reads with, the commit going
// after --contains.
var refsForm = []string{"--count=1", "--format=%(refname)", "--contains", "refs/heads", "refs/remotes", "refs/tags"}

// OnARef reports whether a branch, a remote branch or a tag of the clone at
// dir contains commit.
func OnARef(dir, commit string) (bool, error) {
	args := append([]string{"for-each-ref"}, refsForm[:3]...)
	args = append(append(args, commit), refsForm[3:]...)
	out, err := Git(dir, args...)
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}
