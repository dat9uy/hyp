package acts

import (
	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/tables"
)

// CommitView is what the prints read of one commit's kept model: whether
// it keeps a model in goals/model/, and the walk of an example on it. It
// reads the kept model from git once, at the first ask, and reads only
// objects of the named commit and of the example folders, no history.
type CommitView struct {
	root, commit string
	loaded       bool
	kept         bool
	steps        []keptStep
	err          error
}

// CommitAt is the view of commit in the repository at root.
func CommitAt(root, commit string) *CommitView {
	return &CommitView{root: root, commit: commit}
}

func (c *CommitView) load() {
	if !c.loaded {
		c.kept, c.steps, c.err = keptSteps(c.root, c.commit)
		c.loaded = true
	}
}

// Keeps reports whether the commit keeps a model in goals/model/.
func (c *CommitView) Keeps() (bool, error) {
	c.load()
	return c.kept, c.err
}

// Walk is the walk of the run section of the example folder with tree hash
// hash on the commit. The example is read before the kept model, so an
// unreadable example is the error even when the kept model is unreadable
// too.
func (c *CommitView) Walk(hash string) (Walk, error) {
	ex, _, err := readExample(c.root, hash)
	if err != nil {
		return Walk{}, err
	}
	c.load()
	if c.err != nil || !c.kept {
		return Walk{}, c.err
	}
	return walkOn(c.steps, ex.Run), nil
}

// GoalView is what the prints of one Goal read of git: whether its selected
// commit is on no ref, what was read of that commit (its kept-model change
// and its charts), the walk of each placed example and the words of each
// example. Each read is made once, at the first ask. Walks are on the
// selected commit, from what its run or land read, while the Goal has a
// select, else on H0 from git.
type GoalView struct {
	root     string
	g        *tables.Goal
	readDone bool
	read     *events.CommitRead
	readErr  error
	base     *CommitView
}

// ViewOf is the view of g in the repository at root.
func ViewOf(root string, g *tables.Goal) *GoalView {
	return &GoalView{root: root, g: g}
}

// OnNoRef reports whether g has a select whose commit is in this clone and
// on no branch or tag (OnNoRef).
func (v *GoalView) OnNoRef() bool {
	return v.g.Select != nil && OnNoRef(v.root, v.g.Select.Commit)
}

// Selected is what was read of the commit g is selected on (SelectedRead).
// A Goal with no select has read nothing of a selected commit: its read is
// empty, so it holds no kept change and no chart.
func (v *GoalView) Selected() (*events.CommitRead, error) {
	if !v.readDone {
		if v.g.Select == nil {
			v.read = &events.CommitRead{}
		} else {
			v.read, v.readErr = SelectedRead(v.root, v.g)
		}
		v.readDone = true
	}
	return v.read, v.readErr
}

// Kept is what the selected commit adds or changes in the kept model
// against H0.
func (v *GoalView) Kept() ([]events.KeptItem, error) {
	read, err := v.Selected()
	if err != nil {
		return nil, err
	}
	return read.Kept, nil
}

// Charts is each chart of the selected commit's spec folder.
func (v *GoalView) Charts() ([]events.Chart, error) {
	read, err := v.Selected()
	if err != nil {
		return nil, err
	}
	return read.Charts, nil
}

// WalkCommit is the commit the walks of g are on, and whether it is the
// selected commit rather than H0.
func (v *GoalView) WalkCommit() (commit string, selected bool) {
	if v.g.Select != nil {
		return v.g.Select.Commit, true
	}
	return v.g.Base, false
}

// Keeps reports whether the commit the walks are on keeps a model in
// goals/model/.
func (v *GoalView) Keeps() (bool, error) {
	if v.g.Select != nil {
		read, err := v.Selected()
		if err != nil {
			return false, err
		}
		return read.Keeps, nil
	}
	return v.baseView().Keeps()
}

// Walk is the walk of the run section of the example folder with tree hash
// hash on the commit the walks are on.
func (v *GoalView) Walk(hash string) (Walk, error) {
	if v.g.Select != nil {
		read, err := v.Selected()
		if err != nil {
			return Walk{}, err
		}
		return WalkRead(v.root, read, hash)
	}
	return v.baseView().Walk(hash)
}

// ExampleWords is the words section of the example folder with tree hash
// hash (ExampleWords).
func (v *GoalView) ExampleWords(hash string) (string, error) {
	return ExampleWords(v.root, hash)
}

func (v *GoalView) baseView() *CommitView {
	if v.base == nil {
		v.base = CommitAt(v.root, v.g.Base)
	}
	return v.base
}
