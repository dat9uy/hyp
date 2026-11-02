package cli

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestSignalsReplaysOnce counts, with a coverage build of hyp, how often
// replay applies an event while signals and status run: once per applied
// event means each command replays the log once.
func TestSignalsReplaysOnce(t *testing.T) {
	if testing.Short() {
		t.Skip("builds hyp with coverage")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "hyp")
	build := exec.Command("go", "build", "-o", bin, "-cover", "-covermode=count",
		"-coverpkg=./cmd/hyp,./internal/replay", "./cmd/hyp")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	line := lineOf(t, filepath.Join(root, "internal/replay/replay.go"), "st.Applied = append(st.Applied, ev)")
	copySynth(t)

	var events struct{ Events []json.RawMessage }
	if err := json.Unmarshal([]byte(hyp(t, "events", "--json").stdout), &events); err != nil {
		t.Fatal(err)
	}
	applied := len(events.Events)
	for _, args := range [][]string{{"signals", "--json"}, {"status"}} {
		cov := t.TempDir()
		run := exec.Command(bin, args...)
		run.Env = append(os.Environ(), "GOCOVERDIR="+cov)
		if out, err := run.CombinedOutput(); err != nil {
			t.Fatalf("hyp %v: %v\n%s", args, err, out)
		}
		if steps := stepsAt(t, cov, line); steps != applied {
			t.Errorf("hyp %v: replay applied %d events, the log applies %d", args, steps, applied)
		}
	}
}

// lineOf returns the number of the line in path holding text.
func lineOf(t *testing.T, path, text string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for i, l := range strings.Split(string(data), "\n") {
		if strings.Contains(l, text) {
			return i + 1
		}
	}
	t.Fatalf("no %q in %s", text, path)
	return 0
}

// stepsAt sums the counts of replay.go's coverage blocks that span line.
func stepsAt(t *testing.T, cov string, line int) int {
	t.Helper()
	txt := filepath.Join(cov, "c.txt")
	if out, err := exec.Command("go", "tool", "covdata", "textfmt", "-i="+cov, "-o="+txt).CombinedOutput(); err != nil {
		t.Fatalf("covdata: %v\n%s", err, out)
	}
	f, err := os.Open(txt)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// internal/replay/replay.go:104.2,105.12 2 37
		pos, rest, ok := strings.Cut(sc.Text(), " ")
		file, span, _ := strings.Cut(pos, ":")
		if !ok || !strings.HasSuffix(file, "/internal/replay/replay.go") {
			continue
		}
		from, to, _ := strings.Cut(span, ",")
		start, _ := strconv.Atoi(strings.Split(from, ".")[0])
		end, _ := strconv.Atoi(strings.Split(to, ".")[0])
		fields := strings.Fields(rest)
		count, _ := strconv.Atoi(fields[len(fields)-1])
		if start <= line && line <= end {
			n += count
		}
	}
	return n
}
