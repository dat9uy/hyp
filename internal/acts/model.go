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
// in first-seen order, and each top-level definition of its .tla files, with
// the names of the definitions in module order. A commit with no kept model
// has no rules.
func keptRules(root, commit string) ([]string, map[string]definition, []string, error) {
	out, err := gitread.Git(root, "ls-tree", "-r", "-z", commit, KeptDir)
	if err != nil {
		return nil, nil, nil, err
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
			return nil, nil, nil, err
		}
		if ext == ".cfg" {
			rules = appendNew(rules, CfgRules(string(data))...)
		} else {
			modules = append(modules, string(data))
		}
	}
	defs := map[string]definition{}
	var order []string
	for _, m := range modules {
		for _, d := range definitions(m) {
			if _, had := defs[d.Name]; !had {
				order = append(order, d.Name)
			}
			defs[d.Name] = d
		}
	}
	return rules, defs, order, nil
}

// definition is one top-level definition of a TLA+ module: its name, its
// text, and what the line comments right above it name: the source of a
// rule (\* source: X) and the command of a step (\* command: X).
type definition struct {
	Name, Text, Source, Command string
}

// definitions is each top-level definition of a TLA+ module, in module
// order: a line "Name ==" or "Name(args) ==" at the start of a line and the
// indented lines after it, with trailing spaces of each line removed. The
// line comments right above the definition give its source and command.
func definitions(tla string) []definition {
	var defs []definition
	var cur *definition
	var body []string
	source, command := "", ""
	flush := func() {
		if cur != nil {
			cur.Text = strings.Join(body, "\n")
			defs = append(defs, *cur)
		}
		cur, body = nil, nil
	}
	for _, line := range strings.Split(tla, "\n") {
		line = strings.TrimRight(line, " \t\r")
		if cur != nil && line != "" && (line[0] == ' ' || line[0] == '\t') {
			body = append(body, line)
			continue
		}
		flush()
		if c, ok := strings.CutPrefix(line, `\*`); ok {
			c = strings.TrimSpace(c)
			if v, ok := strings.CutPrefix(c, "source:"); ok {
				source = strings.TrimSpace(v)
			}
			if v, ok := strings.CutPrefix(c, "command:"); ok {
				command = strings.TrimSpace(v)
			}
			continue
		}
		if m := defLine.FindStringSubmatch(line); m != nil {
			cur, body = &definition{Name: m[1], Source: source, Command: command}, []string{line}
		}
		source, command = "", ""
	}
	flush()
	return defs
}

var defLine = regexp.MustCompile(`^(\w+)\s*(?:\([^)]*\))?\s*==`)

// KeptChange is each rule of the kept model at commit that is new, or whose
// text differs, against the kept model at from, in the order of the
// commit's check configs.
func KeptChange(root, from, commit string) ([]string, error) {
	d, err := KeptDiff(root, from, commit)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, it := range d {
		if it.Rule {
			out = append(out, it.Name)
		}
	}
	return out, nil
}

// KeptItem is one step or rule of the kept model that a commit adds or
// changes. A step is a definition tagged with the command it takes; a rule
// is a name its check configs check, with the source its comment names.
type KeptItem struct {
	Rule    bool   // a rule; else a step
	Name    string // the name of its definition
	Command string // a step's command
	Source  string // a rule's source, as its comment names it; empty when none
}

// KeptDiff is what the kept model at commit adds or changes against the
// kept model at from: each step that is new, or whose text or command
// differs, in module order, then each rule that is new, or whose text
// differs, in the order of the commit's check configs.
func KeptDiff(root, from, commit string) ([]KeptItem, error) {
	rules, defs, order, err := keptRules(root, commit)
	if err != nil {
		return nil, err
	}
	old, oldDefs, _, err := keptRules(root, from)
	if err != nil {
		return nil, err
	}
	var out []KeptItem
	for _, name := range order {
		d := defs[name]
		if d.Command == "" || slices.Contains(rules, name) {
			continue
		}
		if o, had := oldDefs[name]; had && o.Text == d.Text && o.Command == d.Command {
			continue
		}
		out = append(out, KeptItem{Name: name, Command: d.Command})
	}
	for _, r := range rules {
		o, had := oldDefs[r]
		if slices.Contains(old, r) && had && o.Text == defs[r].Text {
			continue
		}
		out = append(out, KeptItem{Rule: true, Name: r, Source: defs[r].Source})
	}
	return out, nil
}
