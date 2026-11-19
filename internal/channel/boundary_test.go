package channel

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/fault"
)

const module = "github.com/dat9uy/hyp"

// binaryPackages are the packages of this module the hyp binary is built
// from, by import path, with their folders.
func binaryPackages(t *testing.T) map[string]string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	var walk func(path string)
	walk = func(path string) {
		if _, seen := out[path]; seen || !strings.HasPrefix(path, module) {
			return
		}
		dir := filepath.Join(root, strings.TrimPrefix(path, module))
		pkg, err := build.ImportDir(dir, 0)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		out[path] = dir
		for _, imp := range pkg.Imports {
			walk(imp)
		}
	}
	walk(module + "/cmd/hyp")
	return out
}

// sourceFiles parses the non-test Go files of the package in dir.
func sourceFiles(t *testing.T, fset *token.FileSet, dir string) []*ast.File {
	t.Helper()
	pkg, err := build.ImportDir(dir, 0)
	if err != nil {
		t.Fatal(err)
	}
	var files []*ast.File
	for _, name := range pkg.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	return files
}

// constValues maps each string constant of the files to its value.
func constValues(files []*ast.File) map[string]string {
	out := map[string]string{}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			vs, ok := n.(*ast.ValueSpec)
			if !ok {
				return true
			}
			for i, name := range vs.Names {
				if i < len(vs.Values) {
					if lit, ok := vs.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						if v, err := strconv.Unquote(lit.Value); err == nil {
							out[name.Name] = v
						}
					}
				}
			}
			return true
		})
	}
	return out
}

// verbs is how many values a format names: the highest argument a verb
// reads, explicit indexes included.
func verbs(format string) int {
	most, next := 0, 1
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		i++
		if i < len(format) && format[i] == '%' {
			continue
		}
		for i < len(format) && strings.ContainsRune("+-# 0123456789.", rune(format[i])) {
			i++
		}
		if i < len(format) && format[i] == '[' {
			end := strings.IndexByte(format[i:], ']')
			n, _ := strconv.Atoi(format[i+1 : i+end])
			next = n
			i += end + 1
		}
		most = max(most, next)
		next++
	}
	return most
}

// faultCall is one call of a fault constructor: the code it names and how
// many values it passes, or -1 when it spreads a slice.
type faultCall struct {
	pos    string
	code   string
	values int
}

var constructors = map[string]int{"Refuse": 1, "BadArg": 0, "EnvFault": 0} // index of the code argument

// TestOnlyTheChannelWritesText holds the boundary of the channel: no package
// of the binary but internal/channel writes text for the agent. Outside the
// channel no file imports fmt or calls errors.New, and no failure carries a
// sentence among its values; every code has words in the channel, which
// name as many values as each call passes; and the channel prints every
// value as hyp holds it, never through %q or strconv.Quote.
func TestOnlyTheChannelWritesText(t *testing.T) {
	fset := token.NewFileSet()
	pkgs := binaryPackages(t)
	faultDir := pkgs[module+"/internal/fault"]
	channelDir := pkgs[module+"/internal/channel"]
	if faultDir == "" || channelDir == "" {
		t.Fatalf("the binary is not built from internal/fault and internal/channel: %v", pkgs)
	}
	codeOf := map[string]string{} // constant name, as a call names it, to the code
	for name, v := range constValues(sourceFiles(t, fset, faultDir)) {
		codeOf["fault."+name] = v
	}
	channelFiles := sourceFiles(t, fset, channelDir)
	for name, v := range constValues(channelFiles) {
		if strings.HasPrefix(name, "code") {
			codeOf[name] = v
		}
	}

	var calls []faultCall
	for path, dir := range pkgs {
		inChannel := dir == channelDir
		for _, f := range sourceFiles(t, fset, dir) {
			at := func(n ast.Node) string { return fset.Position(n.Pos()).String() }
			if !inChannel {
				for _, imp := range f.Imports {
					if imp.Path.Value == `"fmt"` {
						t.Errorf("%s imports fmt: only internal/channel writes text (%s)", at(imp), path)
					}
				}
			}
			ast.Inspect(f, func(n ast.Node) bool {
				if lit, ok := n.(*ast.BasicLit); ok && inChannel && lit.Kind == token.STRING && strings.Contains(lit.Value, "%q") {
					t.Errorf("%s: a format quotes a value with %%q; print it as hyp holds it", at(lit))
				}
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkg, _ := sel.X.(*ast.Ident)
				if pkg == nil {
					return true
				}
				switch {
				case !inChannel && pkg.Name == "errors" && sel.Sel.Name == "New":
					t.Errorf("%s calls errors.New: a failure outside the channel is a fault code", at(call))
				case inChannel && pkg.Name == "strconv" && strings.HasPrefix(sel.Sel.Name, "Quote"):
					t.Errorf("%s calls strconv.%s; print the value as hyp holds it", at(call), sel.Sel.Name)
				case pkg.Name == "fault":
					ci, ok := constructors[sel.Sel.Name]
					if !ok || len(call.Args) <= ci {
						return true
					}
					var code string
					switch a := call.Args[ci].(type) {
					case *ast.SelectorExpr:
						if x, ok := a.X.(*ast.Ident); ok {
							code = codeOf[x.Name+"."+a.Sel.Name]
						}
					case *ast.Ident:
						code = codeOf[a.Name]
					}
					if code == "" {
						t.Errorf("%s: fault.%s names no code constant", at(call), sel.Sel.Name)
						return true
					}
					values := len(call.Args) - ci - 1
					if call.Ellipsis.IsValid() {
						values = -1
					}
					calls = append(calls, faultCall{pos: at(call), code: code, values: values})
					if !inChannel {
						for _, a := range call.Args[ci+1:] {
							if lit, ok := a.(*ast.BasicLit); ok && lit.Kind == token.STRING && strings.Contains(lit.Value, " ") {
								t.Errorf("%s passes the text %s to a failure; its words belong in the channel", at(lit), lit.Value)
							}
						}
					}
				}
				return true
			})
		}
	}
	if len(calls) < 100 {
		t.Fatalf("found %d fault calls; the walk missed the binary's packages", len(calls))
	}

	for _, c := range calls {
		code := fault.Code(c.code)
		if shaped[code] {
			if c.values != 1 {
				t.Errorf("%s: %s takes one value, its facts; the call passes %d", c.pos, code, c.values)
			}
			continue
		}
		format, ok := words[code]
		if !ok {
			t.Errorf("%s: %s has no words in the channel", c.pos, code)
			continue
		}
		if c.values >= 0 && verbs(format) != c.values {
			t.Errorf("%s: %s names %d values, the call passes %d: %s", c.pos, code, verbs(format), c.values, format)
		}
	}
	for name, v := range codeOf {
		if _, ok := words[fault.Code(v)]; !ok && !shaped[fault.Code(v)] && name != "fault.NoGuard" && !strings.HasPrefix(name, "fault.Guard") {
			t.Errorf("code %s (%s) has no words in the channel", v, name)
		}
	}
}

func TestVerbs(t *testing.T) {
	for format, want := range map[string]int{
		"no values":              0,
		"%s and %d, 100%%":       2,
		"%[1]s then %[3]s %[1]s": 3,
		"%5.2f %-3s":             2,
	} {
		if got := verbs(format); got != want {
			t.Errorf("verbs(%q) = %d, want %d", format, got, want)
		}
	}
	if slices.Contains([]int{verbs("%[2]s")}, 0) {
		t.Fatal("an indexed verb names its index")
	}
}
