// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// sdkmap walks internal/service/* and reports every resource annotated with @SDKResource,
// along with the functions it uses for CRUD, CustomizeDiff, and custom import.
//
// Annotations are recognized using the same rules as the service package generator
// (internal/generate/servicepackage), so the output matches what the provider registers.
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// These mirror internal/generate/servicepackage.
var (
	annotationRegexp = regexp.MustCompile(`^//\s*@([0-9A-Za-z]+)(\(([^)]*)\))?\s*$`)
	typeNameRegexp   = regexp.MustCompile(`^aws(?:_[a-z0-9]+)+$`)
)

// servicePackage is a service's Go package, with all of its non-test files parsed.
type servicePackage struct {
	Service string
	Files   []*ast.File
	// Decls maps the name of each package-level function, variable, and constant to its declaration.
	Decls map[string]token.Pos
}

// sdkResource is an SDK resource and the functions that implement it.
type sdkResource struct {
	Service  string
	TypeName string
	// Function is the factory function annotated with @SDKResource.
	Function string
	Position token.Position

	Create, Read, Update, Delete *funcRef
	CustomizeDiff                []funcRef
	// Import is nil if the resource has no importer or uses schema.ImportStatePassthroughContext.
	Import *funcRef
}

// funcRef is a reference to a function from a resource's schema.Resource.
type funcRef struct {
	// Expr is the reference as written, e.g. "resourceObjectCreate" or "verify.SetTagsDiff".
	Expr string
	// Position is where the function is defined: a package-level declaration (for a call, the called
	// function's declaration), or else the function literal or call expression itself.
	// It is the zero value if the function is not declared in the package, e.g. verify.SetTagsDiff or a method value.
	Position token.Position
}

func main() {
	root := flag.String("root", ".", "path to the root of the terraform-provider-aws repository")
	flag.Parse()

	fset := token.NewFileSet()
	pkgs, err := loadServicePackages(fset, filepath.Join(*root, "internal", "service"))
	if err != nil {
		log.Fatal(err)
	}

	var resources []sdkResource
	for _, pkg := range pkgs {
		resources = append(resources, findSDKResources(fset, pkg)...)
	}

	location := func(pos token.Position) string {
		if !pos.IsValid() {
			return ""
		}
		path, err := filepath.Rel(*root, pos.Filename)
		if err != nil {
			log.Fatal(err)
		}
		return fmt.Sprintf("%s:%d", filepath.ToSlash(path), pos.Line)
	}
	printRef := func(kind string, ref *funcRef) {
		if ref != nil {
			fmt.Printf("\t%s\t%s\t%s\n", kind, ref.Expr, location(ref.Position))
		}
	}

	for _, r := range resources {
		fmt.Printf("%s\t%s\t%s\t%s\n", r.Service, r.TypeName, r.Function, location(r.Position))
		printRef("Create", r.Create)
		printRef("Read", r.Read)
		printRef("Update", r.Update)
		printRef("Delete", r.Delete)
		for _, ref := range r.CustomizeDiff {
			printRef("CustomizeDiff", &ref)
		}
		printRef("Import", r.Import)
	}
	fmt.Fprintf(os.Stderr, "found %d @SDKResource resources in %d packages\n", len(resources), len(pkgs))
}

// findSDKResources returns the resources annotated with @SDKResource in pkg.
func findSDKResources(fset *token.FileSet, pkg *servicePackage) []sdkResource {
	var resources []sdkResource

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Doc == nil {
				continue
			}
			for _, c := range fn.Doc.List {
				typeName, ok := sdkResourceTypeName(c.Text)
				if !ok {
					continue
				}
				r := sdkResource{
					Service:  pkg.Service,
					TypeName: typeName,
					Function: fn.Name.Name,
					Position: fset.Position(fn.Pos()),
				}
				if lit := returnedResource(fn); lit != nil {
					addFuncRefs(fset, pkg, &r, lit)
				} else {
					fmt.Fprintf(os.Stderr, "%s: %s does not return a schema.Resource literal\n", r.Position, fn.Name.Name)
				}
				resources = append(resources, r)
			}
		}
	}

	return resources
}

// returnedResource returns the schema.Resource literal returned at the end of fn, or nil.
func returnedResource(fn *ast.FuncDecl) *ast.CompositeLit {
	if fn.Body == nil || len(fn.Body.List) == 0 {
		return nil
	}
	ret, ok := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return nil
	}
	return compositeLit(ret.Results[0], "schema", "Resource")
}

// addFuncRefs records the functions referenced by the fields of a schema.Resource literal.
func addFuncRefs(fset *token.FileSet, pkg *servicePackage, r *sdkResource, lit *ast.CompositeLit) {
	for field, value := range fields(lit) {
		switch field {
		case "Create", "CreateContext", "CreateWithoutTimeout":
			r.Create = newFuncRef(fset, pkg, value)
		case "Read", "ReadContext", "ReadWithoutTimeout":
			r.Read = newFuncRef(fset, pkg, value)
		case "Update", "UpdateContext", "UpdateWithoutTimeout":
			r.Update = newFuncRef(fset, pkg, value)
		case "Delete", "DeleteContext", "DeleteWithoutTimeout":
			r.Delete = newFuncRef(fset, pkg, value)
		case "CustomizeDiff":
			for _, expr := range customizeDiffFuncs(value) {
				r.CustomizeDiff = append(r.CustomizeDiff, *newFuncRef(fset, pkg, expr))
			}
		case "Importer":
			importer := compositeLit(value, "schema", "ResourceImporter")
			if importer == nil {
				continue
			}
			for field, value := range fields(importer) {
				if (field == "State" || field == "StateContext") && !isSelector(value, "schema", "ImportStatePassthrough", "ImportStatePassthroughContext") {
					r.Import = newFuncRef(fset, pkg, value)
				}
			}
		}
	}
}

// customizeDiffFuncs flattens customdiff.All and customdiff.Sequence into the functions they combine.
func customizeDiffFuncs(expr ast.Expr) []ast.Expr {
	if call, ok := expr.(*ast.CallExpr); ok && isSelector(call.Fun, "customdiff", "All", "Sequence") {
		var exprs []ast.Expr
		for _, arg := range call.Args {
			exprs = append(exprs, customizeDiffFuncs(arg)...)
		}
		return exprs
	}
	return []ast.Expr{expr}
}

func newFuncRef(fset *token.FileSet, pkg *servicePackage, expr ast.Expr) *funcRef {
	ref := &funcRef{Expr: types.ExprString(expr)}

	switch e := expr.(type) {
	case *ast.Ident:
		if pos, ok := pkg.Decls[e.Name]; ok {
			ref.Position = fset.Position(pos)
		}
	case *ast.FuncLit:
		ref.Expr = "func literal"
		ref.Position = fset.Position(e.Pos())
	case *ast.CallExpr:
		ref.Expr = types.ExprString(e.Fun) + "(...)"
		ref.Position = fset.Position(e.Pos())
		if fn, ok := e.Fun.(*ast.Ident); ok {
			if pos, ok := pkg.Decls[fn.Name]; ok {
				ref.Position = fset.Position(pos)
			}
		}
	}

	return ref
}

// fields returns the key/value pairs of a struct literal, keyed by field name.
func fields(lit *ast.CompositeLit) map[string]ast.Expr {
	m := make(map[string]ast.Expr)
	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if key, ok := kv.Key.(*ast.Ident); ok {
				m[key.Name] = kv.Value
			}
		}
	}
	return m
}

// compositeLit returns the literal if expr is &pkg.typeName{...}, or nil.
func compositeLit(expr ast.Expr, pkg, typeName string) *ast.CompositeLit {
	if u, ok := expr.(*ast.UnaryExpr); ok && u.Op == token.AND {
		if lit, ok := u.X.(*ast.CompositeLit); ok && isSelector(lit.Type, pkg, typeName) {
			return lit
		}
	}
	return nil
}

// isSelector reports whether expr is pkg.name for any of names.
func isSelector(expr ast.Expr, pkg string, names ...string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	return ok && x.Name == pkg && slices.Contains(names, sel.Sel.Name)
}

// sdkResourceTypeName returns the resource type name from an @SDKResource annotation comment.
func sdkResourceTypeName(comment string) (string, bool) {
	m := annotationRegexp.FindStringSubmatch(comment)
	if m == nil || m[1] != "SDKResource" {
		return "", false
	}

	for arg := range strings.SplitSeq(m[3], ",") {
		arg = strings.TrimSpace(arg)
		if arg == "" || strings.Contains(arg, "=") {
			continue
		}
		if typeName := strings.Trim(arg, `"`); typeNameRegexp.MatchString(typeName) {
			return typeName, true
		}
		return "", false
	}

	return "", false
}

// loadServicePackages parses the Go package in each service directory directly under dir, honoring build constraints.
func loadServicePackages(fset *token.FileSet, dir string) ([]*servicePackage, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var pkgs []*servicePackage
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		bp, err := build.ImportDir(path, 0)
		if _, ok := errors.AsType[*build.NoGoError](err); ok {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("loading package %s: %w", path, err)
		}
		if len(bp.GoFiles) == 0 {
			continue
		}

		pkg := &servicePackage{
			Service: entry.Name(),
			Decls:   make(map[string]token.Pos),
		}
		for _, name := range bp.GoFiles {
			file, err := parser.ParseFile(fset, filepath.Join(path, name), nil, parser.ParseComments|parser.SkipObjectResolution)
			if err != nil {
				return nil, err
			}
			pkg.Files = append(pkg.Files, file)
			addDecls(pkg.Decls, file)
		}
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

// addDecls records the package-level functions, variables, and constants declared in file.
func addDecls(decls map[string]token.Pos, file *ast.File) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil {
				decls[d.Name.Name] = d.Pos()
			}
		case *ast.GenDecl:
			if d.Tok != token.VAR && d.Tok != token.CONST {
				continue
			}
			for _, spec := range d.Specs {
				for _, name := range spec.(*ast.ValueSpec).Names {
					decls[name.Name] = name.Pos()
				}
			}
		}
	}
}
