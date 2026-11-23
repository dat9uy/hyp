package logstore

import (
	"os"
	"testing"
)

// TestMain keeps every trace a test writes in a home folder of its own, never
// in the home folder of the person who runs the tests: the trace is under
// $HOME/.hyp. $XDG_STATE_HOME is cleared, so a test shows that hyp ignores it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "hyp-home-")
	if err != nil {
		panic(err)
	}
	os.Setenv("HOME", dir)
	os.Unsetenv("XDG_STATE_HOME")
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
