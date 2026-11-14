package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The Goal body's decision "a write is durable once it exits 0"; also
// G-SZCBXZ#1 and #3 under an unclean shutdown): the base fsyncs the events
// file before a write exits. A record write instead adds loose objects and
// moves a ref, which git by default does not harden (core.fsync defaults to
// committed,-loose-object, and reference is not in it). This test logs every
// git run of one record write and asks that each one that writes an object
// or the ref ran with core.fsync naming loose objects and refs.
func TestRecordWriteHardensObjectsAndRef(t *testing.T) {
	recordRepo(t)
	real, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not on PATH")
	}
	bin := t.TempDir()
	log := filepath.Join(t.TempDir(), "git.log")
	script := "#!/bin/sh\nprintf '%s|%s|%s=%s|%s\\n' \"$GIT_CONFIG_PARAMETERS\" \"$GIT_CONFIG_COUNT\" \"$GIT_CONFIG_KEY_0\" \"$GIT_CONFIG_VALUE_0\" \"$*\" >> " + log + "\nexec " + real + " \"$@\"\n"
	writeFile(t, filepath.Join(bin, "git"), script)
	if err := os.Chmod(filepath.Join(bin, "git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	must(t, "add", "goal", "--title", "durable", "--why", "w")
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	cfg, _ := exec.Command(real, "config", "--get", "core.fsync").Output()
	hardened := func(line string) bool {
		v := strings.ToLower(line + " " + string(cfg))
		if !strings.Contains(v, "core.fsync") && len(cfg) == 0 {
			return false
		}
		return strings.Contains(v, "all") || (strings.Contains(v, "loose-object") || strings.Contains(v, "objects") || strings.Contains(v, "committed")) && strings.Contains(v, "reference")
	}
	writers := 0
	for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		args := l[strings.LastIndex(l, "|")+1:]
		if !strings.Contains(args, "hash-object -w") && !strings.Contains(args, "write-tree") && !strings.Contains(args, "commit-tree") && !strings.Contains(args, "update-ref") {
			continue
		}
		writers++
		if !hardened(l) {
			t.Errorf("git %s ran with no core.fsync for loose objects and refs; a write that exits 0 may be lost on an unclean shutdown", args)
		}
	}
	if writers == 0 {
		t.Fatalf("no git run wrote an object or the ref; log:\n%s", data)
	}
}
