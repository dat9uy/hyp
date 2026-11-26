package acts

import (
	"bytes"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/gitread"
)

// A Goal's spec folder, goals/<G>/specs/, may hold a model: a TLA+ module
// (.tla), its check configs (.cfg), one variant config per potential
// Hypothesis under specs/variants/, and a chart (.mmd). hyp reads these files
// from a commit and runs no model checker. A spec folder written before the
// rename holds its variant configs under specs/switches/; hyp reads both.
const variantsDir = "variants/"

// oldVariantsDir is the folder that held the variant configs before they
// were called variant configs.
const oldVariantsDir = "switches/"

// inVariants reports whether p, a path below the spec folder, is in a folder
// of variant configs.
func inVariants(p string) bool {
	return strings.HasPrefix(p, variantsDir) || strings.HasPrefix(p, oldVariantsDir)
}

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
	Named   []string   // the rules some variant config names
	Charts  []SpecFile // the .mmd files outside the variant folders
}

// Unnamed is every rule of the model that no variant config names.
func (m Model) Unnamed() []string {
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
// variants/ (or the older switches/) is a variant config, every other .cfg is
// a check config.
func ReadModel(root string, files []SpecFile) (Model, error) {
	var m Model
	for _, f := range files {
		isVariant := inVariants(f.Path)
		switch path.Ext(f.Path) {
		case ".tla":
			m.Present = true
		case ".mmd":
			if !isVariant {
				m.Charts = append(m.Charts, f)
			}
		case ".cfg":
			data, err := gitread.Git(root, "cat-file", "blob", f.Hash)
			if err != nil {
				return Model{}, err
			}
			if isVariant {
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

// KeptDir is the folder of the repository's kept model: the one model of its
// machine, a TLA+ module (.tla) and its check configs (.cfg). A Goal's model
// is a change to it. It is no Goal's folder, so a runner's own change may
// write it.
const KeptDir = "goals/model/"

// keptRules is the kept model at commit: each rule its check configs check,
// in first-seen order, and the text of each rule's definition in its .tla
// files. A commit with no kept model has no rules.
func keptRules(root, commit string) ([]string, map[string]string, error) {
	out, err := gitread.Git(root, "ls-tree", "-r", "-z", commit, KeptDir)
	if err != nil {
		return nil, nil, err
	}
	var rules []string
	var modules []string
	for _, rec := range bytes.Split(out, []byte{0}) {
		meta, p, ok := bytes.Cut(rec, []byte{'\t'})
		if !ok {
			continue
		}
		f := strings.Fields(string(meta))
		rel := strings.TrimPrefix(string(p), KeptDir)
		if len(f) != 3 || f[1] != "blob" || inVariants(rel) {
			continue
		}
		ext := path.Ext(rel)
		if ext != ".cfg" && ext != ".tla" {
			continue
		}
		data, err := gitread.Git(root, "cat-file", "blob", f[2])
		if err != nil {
			return nil, nil, err
		}
		if ext == ".cfg" {
			rules = appendNew(rules, CfgRules(string(data))...)
		} else {
			modules = append(modules, string(data))
		}
	}
	text := map[string]string{}
	for _, m := range modules {
		for name, def := range definitions(m) {
			text[name] = def
		}
	}
	return rules, text, nil
}

// definitions is the text of each top-level definition of a TLA+ module: a
// line "Name ==" or "Name(args) ==" at the start of a line and the indented
// lines after it, with trailing spaces of each line removed.
func definitions(tla string) map[string]string {
	defs := map[string]string{}
	name := ""
	var body []string
	flush := func() {
		if name != "" {
			defs[name] = strings.Join(body, "\n")
		}
		name, body = "", nil
	}
	for _, line := range strings.Split(tla, "\n") {
		line = strings.TrimRight(line, " \t\r")
		if name != "" && line != "" && (line[0] == ' ' || line[0] == '\t') {
			body = append(body, line)
			continue
		}
		flush()
		if m := defLine.FindStringSubmatch(line); m != nil {
			name, body = m[1], []string{line}
		}
	}
	flush()
	return defs
}

var defLine = regexp.MustCompile(`^(\w+)\s*(?:\([^)]*\))?\s*==`)

// KeptChange is each rule of the kept model at commit that is new, or whose
// text differs, against the kept model at from, in the order of the
// commit's check configs.
func KeptChange(root, from, commit string) ([]string, error) {
	rules, text, err := keptRules(root, commit)
	if err != nil {
		return nil, err
	}
	old, oldText, err := keptRules(root, from)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, r := range rules {
		t, had := oldText[r]
		if !slices.Contains(old, r) || !had || t != text[r] {
			out = append(out, r)
		}
	}
	return out, nil
}

// KeptFrom is the commit whose kept model the kept change of commit is
// measured against: where commit left the checked-out line, the merge base
// of HEAD and commit, unless that is commit itself or a commit before H0;
// then H0.
func KeptFrom(root, h0, commit string) (string, error) {
	code, out, err := gitread.GitExit(root, "merge-base", "HEAD", commit)
	if err != nil {
		return "", err
	}
	mb := strings.TrimSpace(string(out))
	if code != 0 || mb == "" || mb == commit {
		return h0, nil
	}
	after, err := gitread.IsAncestor(root, h0, mb)
	if err != nil {
		return "", err
	}
	if !after {
		return h0, nil
	}
	return mb, nil
}
