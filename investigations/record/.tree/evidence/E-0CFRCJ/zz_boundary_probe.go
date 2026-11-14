package render

import "os/exec"

// probeRunsGit runs git through a shell, outside packages record and sandbox.
func probeRunsGit() ([]byte, error) {
	return exec.Command("sh", "-c", "git rev-parse HEAD").Output()
}

// probeRunsGitByPath runs git by its absolute path.
func probeRunsGitByPath() ([]byte, error) {
	return exec.Command("/usr/bin/git", "rev-parse", "HEAD").Output()
}
