package cli

import (
	"fmt"
	"io"

	"github.com/dat9uy/hyp/internal/sandbox"
)

func cmdSandbox(args []string, stdout, stderr io.Writer) error {
	const use = "usage: hyp sandbox create|drive|remove ..."
	if len(args) == 0 {
		return usageErr(use)
	}
	var err error
	switch args[0] {
	case "-h", "--help", "help":
		return helpText(use)
	case "create":
		err = sandboxCreate(args[1:], stdout)
	case "drive":
		err = sandboxDrive(args[1:], stdout)
	case "remove":
		err = sandboxRemove(args[1:], stdout, stderr)
	default:
		return usageErr(use)
	}
	return err
}

// sandboxDrive handles hyp sandbox drive ID: run every recipe's Drive
// commands in the sandbox.
func sandboxDrive(args []string, stdout io.Writer) error {
	const use = "usage: hyp sandbox drive ID"
	fs := readFlags("sandbox drive", nil)
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr(use)
	}
	_, err = sandbox.Drive(pos[0], stdout)
	return err
}

func sandboxCreate(args []string, stdout io.Writer) error {
	const use = "usage: hyp sandbox create ID --host H --base REV --recipe NAME@VERSION|none [--recipe NAME@VERSION]... [--recipes DIR] [--store SRC --cut EVENT] [--inputs FILE] [--hyp BIN]"
	fs := readFlags("sandbox create", nil)
	host := fs.String("host", "", "host repository root (required)")
	base := fs.String("base", "", "commit to detach the clone at (required)")
	bin := fs.String("hyp", "", "hyp binary the sandbox pins (default: this hyp)")
	var refs multi
	fs.Var(&refs, "recipe", "recipe NAME@VERSION to set up, repeatable in stack order, or none (required)")
	recipes := fs.String("recipes", "", "directory that holds recipes/ (required unless --recipe none)")
	src := fs.String("store", "", "store directory (holding .tree) to cut into wt/.tree; needs --cut")
	cut := fs.String("cut", "", "event id: wt/.tree is the store as of this event; needs --store")
	inputs := fs.String("inputs", "", "file of this run's inputs; the sandbox keeps a copy, hashes it and passes it as HYP_SANDBOX_INPUTS")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 || *host == "" || *base == "" || len(refs) == 0 {
		return usageErr(use)
	}
	rec, err := sandbox.Create(sandbox.CreateOptions{ID: pos[0], Host: *host, Base: *base, Hyp: *bin, Refs: refs, Recipes: *recipes, Store: *src, Cut: *cut, Inputs: *inputs})
	if err != nil {
		return err
	}
	line := fmt.Sprintf("sandbox %s: %s @ %s", pos[0], rec.WTDir(), rec.Base[:min(8, len(rec.Base))])
	if rec.Cut != "" {
		line += ", store cut at " + rec.Cut
	}
	for _, r := range rec.Recipes {
		if r.Ref.IsNone() {
			continue
		}
		line += ", recipe " + r.Ref.String()
		if r.Setup == "passed" {
			line += " set up"
		}
	}
	fmt.Fprintln(stdout, line)
	return nil
}

func sandboxRemove(args []string, stdout, stderr io.Writer) error {
	const use = "usage: hyp sandbox remove ID OUT_DIR [--keep REL]... | hyp sandbox remove ID --discard"
	fs := readFlags("sandbox remove", nil)
	discard := fs.Bool("discard", false, "delete without writing a record")
	var keep multi
	fs.Var(&keep, "keep", "file relative to the sandbox to copy to OUT_DIR (repeatable)")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	o := sandbox.RemoveOptions{Keep: keep, Discard: *discard, Warn: stderr}
	switch {
	case len(pos) == 2 && !*discard:
		o.ID, o.OutDir = pos[0], pos[1]
	case len(pos) == 1 && *discard:
		o.ID = pos[0]
	default:
		return usageErr(use)
	}
	if _, err := sandbox.Remove(o); err != nil {
		return err
	}
	if *discard {
		fmt.Fprintf(stdout, "removed %s; discarded\n", o.ID)
	} else {
		fmt.Fprintf(stdout, "removed %s; record -> %s\n", o.ID, o.OutDir)
	}
	return nil
}
