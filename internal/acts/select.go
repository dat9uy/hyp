package acts

import (
	"encoding/json"
	"strings"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/tables"
)

// SelectOpts is the input of select.
type SelectOpts struct {
	Dir   string // where the command runs
	Actor string
	Role  string
	Why   string
	Goal  string // the Goal's id
	Tree  string // the row tree to select, or a unique prefix of it
}

// SelectReport is what select returns: the Goal, now selected, and the
// commit and row tree it recorded. Its report is the PR body, which the
// channel prints from Selected.
type SelectReport struct {
	Goal     string
	Commit   string
	RowTree  string
	Selected *tables.Goal // the Goal as the log holds it after the select
}

// Select records a select, or refuses with the first failing guard, in the
// order its entry in hyp spec names: Not landed (always holds), Selected, Candidate,
// New cell (for the runner), Goal's actor (for the owner), Spec, then Own
// path. Spec and Own path read the
// commit of the row's latest Result in merge order, which is the commit the
// select records. Everything is checked under the lock.
func Select(o SelectOpts) (SelectReport, error) {
	w, err := begin(o.Dir, o.Actor, o.Role, o.Why)
	if err != nil {
		return SelectReport{}, err
	}
	root := w.root

	var rep SelectReport
	ev, err := w.append(calculatedView, func(t tables.Tables, _ []events.Event) (events.Event, error) {
		ev := events.Event{Type: events.TypeSelect, Why: o.Why, Actor: o.Actor}
		if o.Role != "owner" {
			// git holds only the owner's acts: the runner's Results show
			// the Goal selected, and no event of the runner's holds it
			return ev, fault.Refuse(fault.GuardCandidate, fault.RoleCannotSelect, o.Role)
		}
		g, err := t.Find(o.Goal)
		if err != nil {
			return ev, err
		}
		if err := tables.GuardNotEnded(g); err != nil {
			return ev, err
		}
		if err := tables.GuardNotSelected(g); err != nil {
			return ev, err
		}
		row, err := tables.GuardCandidate(g, o.Tree, o.Role)
		if err != nil {
			return ev, err
		}
		if err := tables.GuardNewCell(g, o.Role); err != nil {
			return ev, err
		}
		if err := tables.GuardGoalActor(g, o.Actor, o.Role); err != nil {
			return ev, err
		}
		commit := tables.LatestCommit(g, row.RowTree)
		if commit == "" {
			return ev, fault.Refuse(fault.GuardCandidate, fault.NoResultOnRow, row.RowTree)
		}
		if err := guardSpec(root, g, commit); err != nil {
			return ev, err
		}
		changes, err := gitread.OwnChange(root, commit, g.Base)
		if err != nil {
			return ev, err
		}
		if err := tables.GuardOwnPath(g, changes); err != nil {
			return ev, err
		}
		payload, err := json.Marshal(events.Select{Role: o.Role, Commit: commit, RowTree: row.RowTree})
		if err != nil {
			return ev, fault.EnvFault(fault.EncodeEvent, events.TypeSelect, err)
		}
		ev.Node, ev.Payload = g.EventID, payload
		if err := tables.GuardSelect(t, ev); err != nil {
			return ev, err
		}
		rep = SelectReport{Goal: g.ID, Commit: commit, RowTree: row.RowTree}
		return ev, nil
	})
	if err != nil {
		return SelectReport{}, err
	}
	t, _, err := ReadTables(root)
	if err != nil {
		return SelectReport{}, err
	}
	g := t.Goals[ev.Node]
	if g == nil || g.Select == nil {
		return SelectReport{}, fault.EnvFault(fault.SelectNotInLog, rep.Goal)
	}
	rep.Selected = g
	return rep, nil
}

// guardSpec is the Spec guard of an act, which reads git: it reads the
// commit (ReadCommit) and checks what it read (tables.GuardSpec).
func guardSpec(root string, g *tables.Goal, commit string) error {
	read, err := readSpec(root, g, commit)
	if err != nil {
		return err
	}
	return tables.GuardSpec(g, commit, &read)
}

// ReadCommit is what a run reads of git about commit, which its Result
// carries: the commits of g's rows it descends from, its spec folder and
// its kept model.
func ReadCommit(root string, g *tables.Goal, commit string) (events.CommitRead, error) {
	read, err := readSpec(root, g, commit)
	if err != nil {
		return read, err
	}
	if read.Above, err = g.Above(gitread.Ancestry{Dir: root}, commit); err != nil {
		return read, err
	}
	return read, nil
}

// readSpec reads the spec folder of g and the kept model at commit.
func readSpec(root string, g *tables.Goal, commit string) (events.CommitRead, error) {
	var read events.CommitRead
	files, err := SpecFiles(root, commit, g.ID)
	if err != nil {
		return read, err
	}
	m, err := ReadModel(root, files)
	if err != nil {
		return read, err
	}
	read.SpecFiles, read.Model, read.Unnamed, read.Named = len(files), m.Present, m.Unnamed(), m.Named
	for _, c := range m.Charts {
		data, err := gitread.Git(root, "cat-file", "blob", c.Hash)
		if err != nil {
			return read, err
		}
		read.Charts = append(read.Charts, events.Chart{Path: c.Path, Text: string(data)})
	}
	keeps, steps, err := keptSteps(root, commit)
	if err != nil {
		return read, err
	}
	read.Keeps = keeps
	for _, st := range steps {
		read.Steps = append(read.Steps, events.KeptStep{Name: st.name, Command: strings.Join(st.words, " ")})
	}
	if read.Kept, err = KeptDiff(root, g.Base, commit); err != nil {
		return read, err
	}
	return read, nil
}

// SelectedRead is what was read of the commit g is selected on, for the
// prints of a selected or landed Goal: what the latest Result of this clone
// on it carries, else what its land carries. A select or land of the older
// form, recorded before acts carried what they read, carries none: then,
// and only then, it reads the commit from git, as those records were
// printed.
func SelectedRead(root string, g *tables.Goal) (*events.CommitRead, error) {
	if read := tables.ReadOf(g, g.Select.Commit); read != nil {
		return read, nil
	}
	if g.Land != nil && g.Land.Read != nil && g.Land.Commit == g.Select.Commit {
		return g.Land.Read, nil
	}
	read, err := readSpec(root, g, g.Select.Commit)
	if err != nil {
		return nil, err
	}
	return &read, nil
}

// ExampleFile is the example.txtar of the example folder with tree hash
// hash, byte for byte.
func ExampleFile(root, hash string) ([]byte, error) {
	if _, _, err := readExample(root, hash); err != nil {
		return nil, err
	}
	return gitread.Git(root, "cat-file", "blob", hash+":example.txtar")
}

// ExampleWords is the words section of the example folder with tree hash
// hash, as the example holds it.
func ExampleWords(root, hash string) (string, error) {
	ex, _, err := readExample(root, hash)
	if err != nil {
		return "", err
	}
	return ex.Words, nil
}
