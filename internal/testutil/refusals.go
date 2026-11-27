package testutil

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Refusal is one call of fault.Refuse in hyp's code: where it is, and the
// values of its guard and its code. A guard or a code that is not a constant
// of package fault, such as a guard taken from a cause, has Const false and
// the source text of its argument as its value.
type Refusal struct {
	At         string
	Guard      string
	Code       string
	GuardConst bool
	CodeConst  bool
}

// Code is the parsed source of hyp's binary: every function, method and
// package-level variable of every package of the module, outside tests and
// testdata, with what each refers to and the refusals its body makes.
type Code struct {
	module string
	nodes  map[string]*codeNode
}

// codeNode is one function, method or package-level variable. A method's key
// is its package path, ".(*).", and its name, whatever its receiver.
type codeNode struct {
	refs     []string
	refusals []Refusal
}

const faultPath = "/internal/fault"

// ReadCode parses every non-test Go file of the module that holds the
// working directory.
func ReadCode(t testing.TB) *Code {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		up := filepath.Dir(root)
		if up == root {
			t.Fatal("no go.mod above the working directory")
		}
		root = up
	}
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	module := ""
	for _, line := range strings.Split(string(mod), "\n") {
		if m, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			module = strings.TrimSpace(m)
		}
	}
	if module == "" {
		t.Fatal("go.mod names no module")
	}

	fs := token.NewFileSet()
	pkgs := map[string][]*ast.File{}
	err = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != root && (d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fs, p, nil, 0)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filepath.Dir(p))
		if err != nil {
			return err
		}
		pkg := module
		if rel != "." {
			pkg = module + "/" + filepath.ToSlash(rel)
		}
		pkgs[pkg] = append(pkgs[pkg], f)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs[module+faultPath]) == 0 {
		t.Fatalf("no files of package %s%s", module, faultPath)
	}

	c := &Code{module: module, nodes: map[string]*codeNode{}}
	consts := map[string]string{}
	top := map[string]map[string]bool{} // package path: its functions and variables
	methods := map[string]map[string]bool{}
	for pkg, files := range pkgs {
		top[pkg], methods[pkg] = map[string]bool{}, map[string]bool{}
		for _, f := range files {
			for _, d := range f.Decls {
				switch d := d.(type) {
				case *ast.FuncDecl:
					if d.Recv != nil {
						methods[pkg][d.Name.Name] = true
					} else {
						top[pkg][d.Name.Name] = true
					}
				case *ast.GenDecl:
					for _, s := range d.Specs {
						v, ok := s.(*ast.ValueSpec)
						if !ok {
							continue
						}
						for i, n := range v.Names {
							if d.Tok == token.VAR {
								top[pkg][n.Name] = true
							}
							if d.Tok == token.CONST && pkg == module+faultPath && i < len(v.Values) {
								if lit, ok := v.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
									consts[n.Name], _ = strconv.Unquote(lit.Value)
								}
							}
						}
					}
				}
			}
		}
	}

	for pkg, files := range pkgs {
		for _, f := range files {
			imports := map[string]string{} // name in the file: module package path
			for _, im := range f.Imports {
				p, _ := strconv.Unquote(im.Path.Value)
				if p != module && !strings.HasPrefix(p, module+"/") {
					continue
				}
				name := path.Base(p)
				if im.Name != nil {
					name = im.Name.Name
				}
				imports[name] = p
			}
			w := walker{c: c, fs: fs, pkg: pkg, imports: imports, top: top, methods: methods, consts: consts}
			for _, d := range f.Decls {
				switch d := d.(type) {
				case *ast.FuncDecl:
					key := pkg + "." + d.Name.Name
					if d.Recv != nil {
						key = pkg + ".(*)." + d.Name.Name
					}
					if d.Body != nil {
						w.walk(key, d.Body)
					}
				case *ast.GenDecl:
					if d.Tok != token.VAR {
						continue
					}
					for _, s := range d.Specs {
						v := s.(*ast.ValueSpec)
						for _, n := range v.Names {
							for _, e := range v.Values {
								w.walk(pkg+"."+n.Name, e)
							}
						}
					}
				}
			}
		}
	}
	return c
}

// walker records the references and refusals of the nodes of one file.
type walker struct {
	c       *Code
	fs      *token.FileSet
	pkg     string
	imports map[string]string
	top     map[string]map[string]bool
	methods map[string]map[string]bool
	consts  map[string]string
}

func (w walker) node(key string) *codeNode {
	n := w.c.nodes[key]
	if n == nil {
		n = &codeNode{}
		w.c.nodes[key] = n
	}
	return n
}

// walk adds to the node key every function, method and variable that root
// refers to, and every refusal it makes. A selector on a package names that
// package's function or variable; any other selector names each method of
// that name in this package and in the module packages the file imports, as
// the type is not known; a plain name names this package's function or
// variable of that name.
func (w walker) walk(key string, root ast.Node) {
	n := w.node(key)
	sel := map[*ast.Ident]bool{} // the names a selector selects, which are no plain names
	ast.Inspect(root, func(x ast.Node) bool {
		switch x := x.(type) {
		case *ast.CallExpr:
			if w.isFault(x.Fun, "Refuse") && len(x.Args) >= 2 {
				r := Refusal{At: w.fs.Position(x.Pos()).String()}
				r.Guard, r.GuardConst = w.value(x.Args[0])
				r.Code, r.CodeConst = w.value(x.Args[1])
				n.refusals = append(n.refusals, r)
			}
		case *ast.SelectorExpr:
			sel[x.Sel] = true
			if id, ok := x.X.(*ast.Ident); ok {
				if p, ok := w.imports[id.Name]; ok {
					if w.top[p][x.Sel.Name] {
						n.refs = append(n.refs, p+"."+x.Sel.Name)
					}
					return false
				}
			}
			if w.methods[w.pkg][x.Sel.Name] {
				n.refs = append(n.refs, w.pkg+".(*)."+x.Sel.Name)
			}
			for _, p := range w.imports {
				if w.methods[p][x.Sel.Name] {
					n.refs = append(n.refs, p+".(*)."+x.Sel.Name)
				}
			}
		case *ast.Ident:
			if !sel[x] && w.top[w.pkg][x.Name] {
				n.refs = append(n.refs, w.pkg+"."+x.Name)
			}
		}
		return true
	})
}

// isFault reports whether e is the selector of package fault's name.
func (w walker) isFault(e ast.Expr, name string) bool {
	s, ok := e.(*ast.SelectorExpr)
	if !ok || s.Sel.Name != name {
		return false
	}
	id, ok := s.X.(*ast.Ident)
	return ok && w.imports[id.Name] == w.c.module+faultPath
}

// value is the value of a constant of package fault that e names, or e's
// source text and false.
func (w walker) value(e ast.Expr) (string, bool) {
	if s, ok := e.(*ast.SelectorExpr); ok && w.isFault(s, s.Sel.Name) {
		if v, ok := w.consts[s.Sel.Name]; ok {
			return v, true
		}
	}
	var b strings.Builder
	ast.Inspect(e, func(x ast.Node) bool {
		if id, ok := x.(*ast.Ident); ok {
			if b.Len() > 0 {
				b.WriteString(".")
			}
			b.WriteString(id.Name)
		}
		return true
	})
	return b.String(), false
}

// Refusals is every refusal of hyp's code, in no order.
func (c *Code) Refusals() []Refusal {
	var out []Refusal
	for _, n := range c.nodes {
		out = append(out, n.refusals...)
	}
	return out
}

// Reachable is every refusal of the functions, methods and variables that
// roots refer to, directly or through others, the roots' own included. A
// root is a function's full name as the runtime prints it, such as
// github.com/dat9uy/hyp/internal/tables.Replay. It fails the test on a root
// the code does not hold.
func (c *Code) Reachable(t testing.TB, roots ...string) []Refusal {
	t.Helper()
	seen := map[string]bool{}
	todo := slices.Clone(roots)
	for _, r := range roots {
		if c.nodes[r] == nil {
			t.Fatalf("no function %s in hyp's code", r)
		}
	}
	var out []Refusal
	for len(todo) > 0 {
		k := todo[len(todo)-1]
		todo = todo[:len(todo)-1]
		if seen[k] {
			continue
		}
		seen[k] = true
		n := c.nodes[k]
		if n == nil {
			continue
		}
		out = append(out, n.refusals...)
		todo = append(todo, n.refs...)
	}
	return out
}
