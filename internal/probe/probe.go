// Package probe runs one example's Probe on one commit and returns what the
// run printed. It gives no verdict.
package probe

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dat9uy/hyp/internal/example"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
)

// DefaultTimeout is how long a Probe may run.
const DefaultTimeout = 10 * time.Minute

// Opts says what to run: the Probe of the example folder ExampleHash on the
// subject of Commit, both read from the repository at Repo. hyp try reads
// either from disk instead: with ExampleDir set, the example folder is that
// folder as it is on disk; with Worktree set, the subject is the files of
// the checkout at Repo as a commit of every change would hold them
// (gitread.CopyWorktree).
type Opts struct {
	Repo        string
	Commit      string
	ExampleHash string
	ExampleDir  string
	Worktree    bool
	Timeout     time.Duration // zero means DefaultTimeout
}

// Out is what the run printed: its stdout with the Probe's last line removed
// and one final newline dropped, the Probe's own stderr, and the run's exit
// code.
type Out struct {
	Stdout, Stderr []byte
	Exit           int
}

var exitLineRe = regexp.MustCompile(`^exit (0|[1-9][0-9]*)$`)

// Run extracts the subject of o.Commit (or copies the checkout's files) and
// the example folder o.ExampleHash (or copies o.ExampleDir) into a fresh
// temporary folder, runs the example's probe/run in the subject
// with empty stdin and only PATH, HOME, the example's external names and the
// HYP_ variables in its environment, and splits the last line "exit N" from
// what it printed. A Probe that cannot start, exits non-zero itself, times
// out or prints no final "exit N" line is an environment error naming the
// example; nothing is recorded. The temporary folder is removed on return.
func Run(o Opts) (Out, error) {
	timeout := o.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	// name is the example in a failure: its hash, or its folder on disk
	name := o.ExampleHash
	if o.ExampleDir != "" {
		name = o.ExampleDir
	}
	tmp, err := os.MkdirTemp("", "hyp-probe-")
	if err != nil {
		return Out{}, fault.EnvFault(fault.ProbeTemp, name, err)
	}
	defer os.RemoveAll(tmp)
	subject, exampleDir := filepath.Join(tmp, "subject"), filepath.Join(tmp, "example")

	if o.Worktree {
		err = gitread.CopyWorktree(o.Repo, subject, "goals", ".tree")
	} else {
		err = gitread.ExtractTree(o.Repo, o.Commit, subject)
	}
	if err != nil {
		return Out{}, err
	}
	for _, p := range []string{"goals", ".tree"} {
		if err := os.RemoveAll(filepath.Join(subject, p)); err != nil {
			return Out{}, fault.EnvFault(fault.ProbeTemp, name, err)
		}
	}
	if o.ExampleDir != "" {
		err = gitread.CopyDir(o.ExampleDir, exampleDir)
	} else {
		err = gitread.ExtractTree(o.Repo, o.ExampleHash, exampleDir)
	}
	if err != nil {
		return Out{}, err
	}
	data, err := os.ReadFile(filepath.Join(exampleDir, "example.txtar"))
	if err != nil {
		return Out{}, fault.EnvFault(fault.ProbeTxtar, name, err)
	}
	ex, err := example.Parse(data)
	if err != nil {
		return Out{}, err
	}
	// Inputs belong to the example: they are written under its folder, never
	// into the subject.
	for _, in := range ex.Inputs {
		path := filepath.Join(exampleDir, "input", filepath.FromSlash(in.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Out{}, fault.EnvFault(fault.ProbeInput, name, in.Path, err)
		}
		if err := os.WriteFile(path, []byte(in.Body), 0o644); err != nil {
			return Out{}, fault.EnvFault(fault.ProbeInput, name, in.Path, err)
		}
	}

	env := []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}
	for _, ext := range ex.ExternalNames() {
		if v, ok := os.LookupEnv(ext); ok {
			env = append(env, ext+"="+v)
		}
	}
	env = append(env, "HYP_SUBJECT="+subject, "HYP_EXAMPLE="+exampleDir, "HYP_RUN="+ex.Run)

	script := filepath.Join(exampleDir, "probe", "run")
	if fi, err := os.Stat(script); err != nil || !fi.Mode().IsRegular() {
		return Out{}, fault.EnvFault(fault.ProbeMissing, name)
	}
	cmd := exec.Command(script)
	cmd.Dir, cmd.Env = subject, env
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// A grandchild that outlives the kill must not hold the pipes open.
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Start(); err != nil {
		return Out{}, fault.EnvFault(fault.ProbeStart, name, err)
	}
	pgid := cmd.Process.Pid
	// Whatever the Probe left running dies with its process group.
	defer syscall.Kill(-pgid, syscall.SIGKILL) //nolint:errcheck // the group may be gone already
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err = <-done:
	case <-timer.C:
		syscall.Kill(-pgid, syscall.SIGKILL) //nolint:errcheck
		<-done
		return Out{}, fault.EnvFault(fault.ProbeTimeout, name, timeout)
	}
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return Out{}, fault.EnvFault(fault.ProbeExit, name, ee.ExitCode(), strings.TrimSpace(stderr.String()))
		}
		return Out{}, fault.EnvFault(fault.ProbeWait, name, err)
	}

	body, code, ok := splitExitLine(stdout.Bytes())
	if !ok {
		return Out{}, fault.EnvFault(fault.ProbeNoExitLine, name)
	}
	return Out{Stdout: body, Stderr: stderr.Bytes(), Exit: code}, nil
}

// splitExitLine splits the last line "exit N" off out. It returns the bytes
// before that line with one final newline dropped, and N.
func splitExitLine(out []byte) (body []byte, code int, ok bool) {
	if !bytes.HasSuffix(out, []byte("\n")) {
		return nil, 0, false
	}
	out = out[:len(out)-1]
	i := bytes.LastIndexByte(out, '\n')
	line := string(out[i+1:])
	if !exitLineRe.MatchString(line) {
		return nil, 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(line, "exit "))
	if err != nil {
		return nil, 0, false
	}
	return bytes.TrimSuffix(out[:i+1], []byte("\n")), n, true
}
