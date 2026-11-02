package sandbox

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dat9uy/hyp/internal/cmdrecord"
	"github.com/dat9uy/hyp/internal/recipe"
)

var stopWait = 5 * time.Second // SIGTERM to SIGKILL, for each group remove stops
var killWait = 2 * time.Second // after SIGKILL

// interrupted is the context every command hyp runs for the operator derives
// from: it ends on Ctrl-C or SIGTERM. Each command leads its own process
// group, so a terminal's SIGINT reaches hyp only, and hyp ends the command's
// group itself. Tests replace it.
var interrupted = func() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// stopGroups stops every process group a Launch, Doctor, Drive or Cleanup
// row started that still has a member: SIGTERM, up to stopWait for the group
// to end, then SIGKILL and up to killWait. It writes one line per group it
// signalled to warn, when warn is not nil. A group id below 2, or hyp's own
// group, is never signalled, so a hand-edited row cannot reach every process
// of the user or hyp itself. A group id reused by an unrelated group after
// the recorded group ended is not told apart (a documented limit).
func stopGroups(rows []cmdrecord.Run, warn io.Writer) {
	seen := map[int]bool{}
	var live []int
	for _, c := range rows {
		switch c.Step {
		case recipe.Launch, recipe.Doctor, recipe.Drive, recipe.Cleanup:
		default:
			continue
		}
		if c.PGID < 2 || c.PGID == syscall.Getpgrp() || seen[c.PGID] {
			continue
		}
		seen[c.PGID] = true
		if groupAlive(c.PGID) {
			live = append(live, c.PGID)
		}
	}
	if len(live) == 0 {
		return
	}
	for _, g := range live {
		syscall.Kill(-g, syscall.SIGTERM)
	}
	left := waitGroups(live, stopWait)
	note := func(g int, how string) {
		if warn != nil {
			fmt.Fprintf(warn, "sandbox: stopped process group %d (%s)\n", g, how)
		}
	}
	still := map[int]bool{}
	for _, g := range left {
		still[g] = true
	}
	for _, g := range live {
		if !still[g] {
			note(g, "SIGTERM")
		}
	}
	for _, g := range left {
		syscall.Kill(-g, syscall.SIGKILL)
	}
	waitGroups(left, killWait)
	for _, g := range left {
		note(g, "SIGKILL")
	}
}

// groupAlive reports whether process group g has a member.
func groupAlive(g int) bool { return syscall.Kill(-g, 0) == nil }

// waitGroups polls every 100 ms until no group in groups has a member or
// d passes, and returns the groups that still do.
func waitGroups(groups []int, d time.Duration) []int {
	deadline := time.Now().Add(d)
	for {
		var left []int
		for _, g := range groups {
			if groupAlive(g) {
				left = append(left, g)
			}
		}
		if len(left) == 0 || time.Now().After(deadline) {
			return left
		}
		time.Sleep(100 * time.Millisecond)
	}
}
