package logstore

import (
	"os"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
)

// StepLogVar names the file hyp appends a step line to for each event it
// writes. Unset or empty, hyp writes no step line.
const StepLogVar = "HYP_STEP_LOG"

// StepLine is the step line of the event line ev: step <Action> <json>,
// where Action is the step its kind's entry names and json is the event as
// it was written. A kind with no step, such as a read, has no step line.
func StepLine(typ string, line []byte) []byte {
	k, ok := events.KindOf(typ)
	if !ok || k.Step == "" {
		return nil
	}
	out := append([]byte("step "+k.Step+" "), line...)
	return append(out, '\n')
}

// appendStep appends the step line of the event line just written to the
// file $HYP_STEP_LOG names, when it names one. It only appends.
func appendStep(typ string, line []byte) error {
	path := os.Getenv(StepLogVar)
	step := StepLine(typ, line)
	if path == "" || step == nil {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fault.EnvFault(fault.FileOpen, path, err)
	}
	defer f.Close()
	if _, err := f.Write(step); err != nil {
		return fault.EnvFault(fault.FileWrite, path, err)
	}
	return nil
}
