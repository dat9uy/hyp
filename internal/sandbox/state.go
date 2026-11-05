package sandbox

// boxState is a sandbox's state. It is derived, never stored
// (docs/sandbox-spec.md §2).
type boxState string

const (
	stateBuilding    boxState = "building"     // no record yet, create holds the lock
	stateSettingUp   boxState = "setting_up"   // a recipe's setup is unfinished, create holds the lock
	stateInterrupted boxState = "interrupted"  // no record or an unfinished setup, and no hyp holds the lock
	stateReady       boxState = "ready"        // every recipe's setup finished
	stateSetupFailed boxState = "setup_failed" // a recipe's setup failed
	stateRemoving    boxState = "removing"     // a remove passed its refusals
)

// deriveState is the state of a sandbox whose directory exists, from its
// record (nil when it has none) and whether another hyp holds its lock. The
// lock only tells a setup in progress from one a killed hyp left: a held lock
// is the busy refusal whatever this returns, because a command takes the lock
// before it reads the state.
func deriveState(rec *Record, held bool) boxState {
	if rec == nil {
		if held {
			return stateBuilding
		}
		return stateInterrupted
	}
	if rec.Removing != "" {
		return stateRemoving
	}
	unfinished := false
	for _, e := range rec.Recipes {
		switch {
		case e.Setup == "failed":
			return stateSetupFailed
		case !e.Ref.IsNone() && e.Setup != "passed" && e.Setup != "none":
			unfinished = true
		}
	}
	switch {
	case unfinished && held:
		return stateSettingUp
	case unfinished:
		return stateInterrupted
	}
	return stateReady
}
