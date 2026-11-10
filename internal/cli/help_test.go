package cli

import (
	"os"
	"testing"
)

// groupSubs lists each group command's subcommands. A help request on any of
// them, or on the group itself, must exit 0 with usage on stdout.
var groupSubs = map[string][]string{
	"add":     {"goal", "dq", "hyp", "exp"},
	"bundle":  {"hash", "check", "new", "run", "evidence", "known-answers"},
	"sandbox": {"create", "drive", "remove"},
}

func TestHelpRequestExitsZeroOnStdout(t *testing.T) {
	t.Chdir(t.TempDir())
	var cases [][]string
	for _, name := range Commands() {
		for _, flag := range []string{"-h", "--help"} {
			cases = append(cases, []string{name, flag})
			for _, sub := range groupSubs[name] {
				cases = append(cases, []string{name, sub, flag})
			}
		}
	}
	for group := range groupSubs {
		cases = append(cases, []string{group, "help"})
	}
	for _, args := range cases {
		r := hyp(t, args...)
		if r.code != 0 || r.stdout == "" || r.stderr != "" {
			t.Errorf("hyp %v: exit %d, stdout %q, stderr %q", args, r.code, r.stdout, r.stderr)
		}
	}
	if _, err := os.Stat(".tree"); !os.IsNotExist(err) {
		t.Fatalf("a help request created .tree (stat err %v)", err)
	}
}

func TestWrongGroupCommandLineStillExitsTwo(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, args := range [][]string{
		{"add"}, {"bundle"}, {"recipe"}, {"sandbox"},
		{"add", "frob"}, {"bundle", "frob"}, {"recipe", "frob"}, {"sandbox", "frob"},
		{"init", "extra"}, {"init", "-h", "extra"}, {"check", "extra"},
	} {
		r := hyp(t, args...)
		if r.code != 2 || r.stdout != "" || r.stderr == "" {
			t.Errorf("hyp %v: exit %d, stdout %q, stderr %q", args, r.code, r.stdout, r.stderr)
		}
	}
}
