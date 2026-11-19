package channel

import (
	"os"
	"testing"
)

// TestMain keeps every trace a test writes in a folder of its own, never in
// the trace folder of the person who runs the tests.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "hyp-trace-")
	if err != nil {
		panic(err)
	}
	os.Setenv("XDG_STATE_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
