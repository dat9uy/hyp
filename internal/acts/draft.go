package acts

import (
	"os"
	"path/filepath"

	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
)

// DraftDir is the folder of draft name in the clone's trace, which holds the
// draft's example set as <line>/<example>. It creates nothing.
func DraftDir(dir, name string) (string, error) {
	if !logstore.ValidDraft(name) {
		return "", fault.BadArg(fault.BadDraftName, name)
	}
	root, err := logstore.TopLevel(dir)
	if err != nil {
		return "", err
	}
	tr, err := logstore.TraceOf(root)
	if err != nil {
		return "", err
	}
	return tr.DraftDir(name), nil
}

// snapshotDraft copies draft name of the clone's trace once, to a new
// temporary folder, and returns the copy and the function that removes it.
// An act reads the draft only through that copy, so the bytes it hashes,
// runs on H0 and writes to a Goal's folder are one set, even when the draft
// changes while a Probe runs.
func snapshotDraft(root, name string) (string, func(), error) {
	if !logstore.ValidDraft(name) {
		return "", nil, fault.BadArg(fault.BadDraftName, name)
	}
	tr, err := logstore.TraceOf(root)
	if err != nil {
		return "", nil, err
	}
	src := tr.DraftDir(name)
	if fi, err := os.Stat(src); err != nil || !fi.IsDir() {
		return "", nil, fault.BadArg(fault.NoDraft, name, src)
	}
	tmp, err := os.MkdirTemp("", "hyp-draft-")
	if err != nil {
		return "", nil, fault.EnvFault(fault.CopyExamples, src, os.TempDir(), err)
	}
	done := func() { os.RemoveAll(tmp) }
	snap := filepath.Join(tmp, name)
	if err := gitread.CopyDir(src, snap); err != nil {
		done()
		return "", nil, err
	}
	return snap, done, nil
}
