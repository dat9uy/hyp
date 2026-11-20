package acts

import (
	"bytes"
	"path"
	"strings"

	"github.com/dat9uy/hyp/internal/gitread"
)

// A Goal's spec folder, goals/<G>/specs/, may hold a model: a TLA+ module
// (.tla), its check configs (.cfg), one switch config per wrong build under
// specs/switches/, and a chart (.mmd). hyp reads these files from a commit
// and runs no model checker.
const switchesDir = "switches/"

// SpecFile is one file of a spec folder at a commit: its path below the spec
// folder and its blob hash.
type SpecFile struct {
	Path string
	Hash string
}

// SpecDir is the spec folder of Goal id, as git names it.
func SpecDir(id string) string { return path.Join("goals", id, "specs") + "/" }

// SpecFiles lists the files under the spec folder of Goal id at commit.
func SpecFiles(root, commit, id string) ([]SpecFile, error) {
	dir := SpecDir(id)
	out, err := gitread.Git(root, "ls-tree", "-r", "-z", commit, dir)
	if err != nil {
		return nil, err
	}
	var files []SpecFile
	for _, rec := range bytes.Split(out, []byte{0}) {
		// <mode> SP <type> SP <hash> TAB <path>
		meta, p, ok := bytes.Cut(rec, []byte{'\t'})
		if !ok {
			continue
		}
		f := strings.Fields(string(meta))
		if len(f) != 3 || f[1] != "blob" {
			continue
		}
		files = append(files, SpecFile{Path: strings.TrimPrefix(string(p), dir), Hash: f[2]})
	}
	return files, nil
}

// Model is what a spec folder holds of a model.
type Model struct {
	Present bool       // the folder holds a .tla file
	Rules   []string   // the rules the check configs check, in first-seen order
	Named   []string   // the rules some switch config names
	Charts  []SpecFile // the .mmd files outside switches/
}

// Unswitched is every rule of the model that no switch config names.
func (m Model) Unswitched() []string {
	named := map[string]bool{}
	for _, r := range m.Named {
		named[r] = true
	}
	var out []string
	for _, r := range m.Rules {
		if !named[r] {
			out = append(out, r)
		}
	}
	return out
}

// ReadModel reads the model of a spec folder from its files: a config under
// switches/ is a switch config, every other .cfg is a check config.
func ReadModel(root string, files []SpecFile) (Model, error) {
	var m Model
	for _, f := range files {
		inSwitches := strings.HasPrefix(f.Path, switchesDir)
		switch path.Ext(f.Path) {
		case ".tla":
			m.Present = true
		case ".mmd":
			if !inSwitches {
				m.Charts = append(m.Charts, f)
			}
		case ".cfg":
			data, err := gitread.Git(root, "cat-file", "blob", f.Hash)
			if err != nil {
				return Model{}, err
			}
			if inSwitches {
				m.Named = appendNew(m.Named, CfgRules(string(data))...)
			} else {
				m.Rules = appendNew(m.Rules, CfgRules(string(data))...)
			}
		}
	}
	return m, nil
}

func appendNew(list []string, names ...string) []string {
	for _, n := range names {
		found := false
		for _, l := range list {
			if l == n {
				found = true
				break
			}
		}
		if !found {
			list = append(list, n)
		}
	}
	return list
}

// cfgKeywords are the keywords of a TLC config file. A name after one of
// ruleKeywords, up to the next keyword, is a rule the config checks.
var cfgKeywords = map[string]bool{
	"SPECIFICATION": true, "INIT": true, "NEXT": true, "CONSTANT": true, "CONSTANTS": true,
	"INVARIANT": true, "INVARIANTS": true, "PROPERTY": true, "PROPERTIES": true,
	"CONSTRAINT": true, "CONSTRAINTS": true, "ACTION_CONSTRAINT": true, "ACTION_CONSTRAINTS": true,
	"SYMMETRY": true, "VIEW": true, "ALIAS": true, "POSTCONDITION": true, "CHECK_DEADLOCK": true,
}

var ruleKeywords = map[string]bool{"INVARIANT": true, "INVARIANTS": true, "PROPERTY": true, "PROPERTIES": true}

// CfgRules is the names a TLC config lists under INVARIANT(S) and
// PROPERTY/PROPERTIES, in order, without its comments.
func CfgRules(cfg string) []string {
	var rules []string
	inRules := false
	for _, tok := range strings.Fields(stripCfgComments(cfg)) {
		if cfgKeywords[tok] {
			inRules = ruleKeywords[tok]
			continue
		}
		if inRules {
			rules = appendNew(rules, tok)
		}
	}
	return rules
}

// stripCfgComments removes \* line comments and (* *) block comments.
func stripCfgComments(cfg string) string {
	var b strings.Builder
	depth := 0
	for i := 0; i < len(cfg); i++ {
		switch {
		case strings.HasPrefix(cfg[i:], "(*"):
			depth++
			i++
		case depth > 0 && strings.HasPrefix(cfg[i:], "*)"):
			depth--
			i++
			b.WriteByte(' ')
		case depth > 0:
		case strings.HasPrefix(cfg[i:], `\*`):
			for i < len(cfg) && cfg[i] != '\n' {
				i++
			}
			b.WriteByte('\n')
		default:
			b.WriteByte(cfg[i])
		}
	}
	return b.String()
}
