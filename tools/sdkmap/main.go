// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// sdkmap walks internal/service/* and reports every resource annotated with @SDKResource,
// along with the functions it uses for CRUD, CustomizeDiff, and custom import,
// and the AWS SDK client methods it calls.
//
// Annotations are recognized using the same rules as the service package generator
// (internal/generate/servicepackage), so the output matches what the provider registers.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	sdkPathPrefix     = "github.com/aws/aws-sdk-go-v2/service/"
	servicePathPrefix = "github.com/hashicorp/terraform-provider-aws/internal/service/"
)

var (
	// These mirror internal/generate/servicepackage.
	annotationRegexp = regexp.MustCompile(`^//\s*@([0-9A-Za-z]+)(\(([^)]*)\))?\s*$`)
	typeNameRegexp   = regexp.MustCompile(`^aws(?:_[a-z0-9]+)+$`)

	paginatorRegexp = regexp.MustCompile(`^New([0-9A-Za-z]+)Paginator$`)
)

// plugin is the Terraform plugin library a resource is implemented with.
type plugin string

const (
	pluginSDKv2 plugin = "SDKv2"
)

// resource is a resource and the functions that implement it.
type resource struct {
	Service  string `json:"service"`
	TypeName string `json:"type_name"`
	Plugin   plugin `json:"plugin"`
	// Function is the factory function annotated with the resource's annotation, e.g. @SDKResource.
	Function string `json:"function"`
	// Location is where Function is declared, as "<file>:<line>" relative to the repository root.
	Location string `json:"location"`

	Create        *funcRef   `json:"create,omitempty"`
	Read          *funcRef   `json:"read,omitempty"`
	Update        *funcRef   `json:"update,omitempty"`
	Delete        *funcRef   `json:"delete,omitempty"`
	CustomizeDiff []*funcRef `json:"customize_diff,omitempty"`
	// Import is nil if the resource has no importer or uses schema.ImportStatePassthroughContext.
	Import *funcRef `json:"import,omitempty"`

	// SDKCalls is the sorted set of AWS SDK client methods reachable from the factory function, e.g. "s3.PutObjectAcl".
	SDKCalls []string `json:"sdk_calls"`
}

// funcRef is a reference to a function from a resource's schema.Resource.
type funcRef struct {
	// Expr is the reference as written, e.g. "resourceObjectCreate" or "verify.SetTagsDiff".
	Expr string `json:"expr"`
	// Location is where the function is defined, as "<file>:<line>" relative to the repository root:
	// a package-level declaration (for a call, the called function's declaration), or else the function literal
	// or call expression itself. It is empty if the function is not declared in the package, e.g. verify.SetTagsDiff or a method value.
	Location string `json:"location,omitempty"`
}

func main() {
	log.SetFlags(0)
	root := flag.String("root", ".", "path to the root of the terraform-provider-aws repository")
	out := flag.String("out", "sdkmap.json", "path of the JSON file to write")
	flag.Parse()

	a, err := load(*root)
	if err != nil {
		log.Fatal(err)
	}

	resources := []resource{}
	for _, pkg := range a.services {
		resources = append(resources, a.findSDKResources(pkg)...)
	}

	data, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, append(data, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("found %d @SDKResource resources in %d packages, wrote %s", len(resources), len(a.services), *out)
}

// analyzer holds the parsed source shared by every resource.
type analyzer struct {
	// root is the path to the root of the repository, which locations are relative to.
	root string
	fset *token.FileSet
	// imports maps each parsed file to its imports, keyed by the name each import is referred to by.
	imports map[*token.File]map[string]string
	// clientMethods maps the conns.AWSClient methods that return an AWS SDK client to the client's type,
	// e.g. "S3Client" to "github.com/aws/aws-sdk-go-v2/service/s3".
	clientMethods map[string]string
	// services lists the service packages in directory order.
	services []*servicePackage
	// servicesByPath maps import paths to service packages.
	servicesByPath map[string]*servicePackage
}

// servicePackage is a service's Go package, with all of its non-test files parsed and its declarations indexed.
type servicePackage struct {
	name  string
	files []*ast.File
	// decls maps the names of package-level functions (*ast.FuncDecl), variables, and constants (*ast.ValueSpec) to their declarations.
	decls map[string]ast.Node
	// methods maps receiver type names to method names to declarations.
	methods map[string]map[string]*ast.FuncDecl
	// interfaces maps interface type names to their method names.
	interfaces map[string][]string
	// fieldTypes maps struct field names, on any struct type, to their types.
	fieldTypes map[string]string
}

// load parses conns.AWSClient and every service package under root.
func load(root string) (*analyzer, error) {
	a := &analyzer{
		root:           root,
		fset:           token.NewFileSet(),
		imports:        make(map[*token.File]map[string]string),
		clientMethods:  make(map[string]string),
		servicesByPath: make(map[string]*servicePackage),
	}

	// Record the conns.AWSClient methods that return an AWS SDK client, e.g. S3Client.
	files, err := a.parsePackage(filepath.Join(root, "internal", "conns"))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || resolveType(fn.Recv.List[0].Type, nil) != "AWSClient" || fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
				continue
			}
			if typ := resolveType(fn.Type.Results.List[0].Type, a.importsOf(file)); isSDKPackage(typ) {
				a.clientMethods[fn.Name.Name] = typ
			}
		}
	}

	// Load the Go package in each service directory.
	serviceDir := filepath.Join(root, "internal", "service")
	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		files, err := a.parsePackage(filepath.Join(serviceDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			continue
		}

		pkg := a.indexPackage(entry.Name(), files)
		a.services = append(a.services, pkg)
		a.servicesByPath[servicePathPrefix+pkg.name] = pkg
	}

	return a, nil
}

// parsePackage parses the non-test Go files of the package in dir, honoring build constraints, and records their imports.
// It returns no files if dir contains no buildable Go files.
func (a *analyzer) parsePackage(dir string) ([]*ast.File, error) {
	bp, err := build.ImportDir(dir, 0)
	if _, ok := errors.AsType[*build.NoGoError](err); ok {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("loading package %s: %w", dir, err)
	}

	var files []*ast.File
	for _, name := range bp.GoFiles {
		file, err := parser.ParseFile(a.fset, filepath.Join(dir, name), nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}

		imports := make(map[string]string)
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return nil, err
			}
			name := path.Base(importPath)
			if spec.Name != nil {
				name = spec.Name.Name
			}
			imports[name] = importPath
		}
		a.imports[a.fset.File(file.Pos())] = imports

		files = append(files, file)
	}

	return files, nil
}

// importsOf returns the imports of the file containing node.
func (a *analyzer) importsOf(node ast.Node) map[string]string {
	return a.imports[a.fset.File(node.Pos())]
}

// indexPackage indexes the package-level declarations in a service package's files.
func (a *analyzer) indexPackage(name string, files []*ast.File) *servicePackage {
	pkg := &servicePackage{
		name:       name,
		files:      files,
		decls:      make(map[string]ast.Node),
		methods:    make(map[string]map[string]*ast.FuncDecl),
		interfaces: make(map[string][]string),
		fieldTypes: make(map[string]string),
	}

	for _, file := range files {
		imports := a.importsOf(file)
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					pkg.decls[d.Name.Name] = d
					continue
				}
				recv := resolveType(d.Recv.List[0].Type, imports)
				if pkg.methods[recv] == nil {
					pkg.methods[recv] = make(map[string]*ast.FuncDecl)
				}
				pkg.methods[recv][d.Name.Name] = d

			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.ValueSpec:
						for _, id := range s.Names {
							pkg.decls[id.Name] = s
						}
					case *ast.TypeSpec:
						switch t := s.Type.(type) {
						case *ast.StructType:
							for _, field := range t.Fields.List {
								for _, id := range field.Names {
									pkg.fieldTypes[id.Name] = resolveType(field.Type, imports)
								}
							}
						case *ast.InterfaceType:
							for _, method := range t.Methods.List {
								for _, id := range method.Names {
									pkg.interfaces[s.Name.Name] = append(pkg.interfaces[s.Name.Name], id.Name)
								}
							}
						}
					}
				}
			}
		}
	}

	return pkg
}

// findSDKResources returns the resources annotated with @SDKResource in pkg.
func (a *analyzer) findSDKResources(pkg *servicePackage) []resource {
	var resources []resource

	for _, file := range pkg.files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			typeNames, tagged := parseAnnotations(fn.Doc)
			if len(typeNames) == 0 {
				continue
			}

			r := resource{
				Service:  pkg.name,
				Plugin:   pluginSDKv2,
				Function: fn.Name.Name,
				Location: a.location(fn),
			}
			if lit := returnedResource(fn); lit != nil {
				a.addFuncRefs(pkg, &r, lit)
			} else {
				log.Printf("%s: %s does not return a schema.Resource literal", r.Location, r.Function)
			}
			r.SDKCalls = a.findSDKCalls(pkg, fn, tagged)
			if len(r.SDKCalls) == 0 {
				log.Printf("%s: no AWS SDK calls found for %s", r.Location, r.Function)
			}

			// A factory function can be registered under several type names.
			for _, typeName := range typeNames {
				r.TypeName = typeName
				resources = append(resources, r)
			}
		}
	}

	return resources
}

// parseAnnotations returns the type names from a function's @SDKResource annotations,
// and whether its @Tags annotation means the tags interceptor lists and updates its tags.
func parseAnnotations(doc *ast.CommentGroup) (typeNames []string, tagged bool) {
	if doc == nil {
		return nil, false
	}

	for _, c := range doc.List {
		m := annotationRegexp.FindStringSubmatch(c.Text)
		if m == nil {
			continue
		}
		switch name, args := m[1], m[3]; name {
		case "SDKResource":
			// The type name is the first argument that is not of the form key=value.
			for arg := range strings.SplitSeq(args, ",") {
				arg = strings.TrimSpace(arg)
				if arg == "" || strings.Contains(arg, "=") {
					continue
				}
				if typeName := strings.Trim(arg, `"`); typeNameRegexp.MatchString(typeName) {
					typeNames = append(typeNames, typeName)
				}
				break
			}
		case "Tags":
			tagged = strings.Contains(args, "identifierAttribute")
		}
	}

	return typeNames, tagged
}

// returnedResource returns the &schema.Resource{...} literal returned at the end of fn, or nil.
func returnedResource(fn *ast.FuncDecl) *ast.CompositeLit {
	if len(fn.Body.List) == 0 {
		return nil
	}
	ret, ok := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return nil
	}
	return schemaLiteral(ret.Results[0], "Resource")
}

// addFuncRefs records the functions referenced by the fields of a schema.Resource literal.
func (a *analyzer) addFuncRefs(pkg *servicePackage, r *resource, lit *ast.CompositeLit) {
	for field, value := range fieldValues(lit) {
		switch field {
		case "Create", "CreateContext", "CreateWithoutTimeout":
			r.Create = a.newFuncRef(pkg, value)
		case "Read", "ReadContext", "ReadWithoutTimeout":
			r.Read = a.newFuncRef(pkg, value)
		case "Update", "UpdateContext", "UpdateWithoutTimeout":
			r.Update = a.newFuncRef(pkg, value)
		case "Delete", "DeleteContext", "DeleteWithoutTimeout":
			r.Delete = a.newFuncRef(pkg, value)
		case "CustomizeDiff":
			for _, expr := range customizeDiffFuncs(value) {
				r.CustomizeDiff = append(r.CustomizeDiff, a.newFuncRef(pkg, expr))
			}
		case "Importer":
			importer := schemaLiteral(value, "ResourceImporter")
			if importer == nil {
				continue
			}
			for field, value := range fieldValues(importer) {
				if (field == "State" || field == "StateContext") && !isSelector(value, "schema", "ImportStatePassthrough", "ImportStatePassthroughContext") {
					r.Import = a.newFuncRef(pkg, value)
				}
			}
		}
	}
}

// customizeDiffFuncs flattens customdiff.All and customdiff.Sequence into the functions they combine.
func customizeDiffFuncs(expr ast.Expr) []ast.Expr {
	call, ok := expr.(*ast.CallExpr)
	if !ok || !isSelector(call.Fun, "customdiff", "All", "Sequence") {
		return []ast.Expr{expr}
	}

	var exprs []ast.Expr
	for _, arg := range call.Args {
		exprs = append(exprs, customizeDiffFuncs(arg)...)
	}
	return exprs
}

func (a *analyzer) newFuncRef(pkg *servicePackage, expr ast.Expr) *funcRef {
	ref := &funcRef{Expr: types.ExprString(expr)}

	switch e := expr.(type) {
	case *ast.Ident:
		if d, ok := pkg.decls[e.Name]; ok {
			ref.Location = a.location(d)
		}
	case *ast.FuncLit:
		ref.Expr = "func literal"
		ref.Location = a.location(e)
	case *ast.CallExpr:
		ref.Expr = types.ExprString(e.Fun) + "(...)"
		ref.Location = a.location(e)
		if fn, ok := e.Fun.(*ast.Ident); ok {
			if d, ok := pkg.decls[fn.Name]; ok {
				ref.Location = a.location(d)
			}
		}
	}

	return ref
}

// location returns where node is, as "<file>:<line>" relative to the repository root.
func (a *analyzer) location(node ast.Node) string {
	pos := a.fset.Position(node.Pos())
	rel, err := filepath.Rel(a.root, pos.Filename)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s:%d", filepath.ToSlash(rel), pos.Line)
}

// callFinder collects the AWS SDK client methods called from the declarations it follows.
type callFinder struct {
	*analyzer
	visited map[ast.Node]bool
	calls   map[string]bool
}

// findSDKCalls returns the sorted set of AWS SDK client methods reachable from a resource's factory function.
func (a *analyzer) findSDKCalls(pkg *servicePackage, factory *ast.FuncDecl, tagged bool) []string {
	f := &callFinder{analyzer: a, visited: make(map[ast.Node]bool), calls: make(map[string]bool)}

	f.follow(pkg, factory)
	if tagged {
		// The tags interceptor lists and updates tags through the service package's ListTags and UpdateTags.
		for _, name := range []string{"ListTags", "UpdateTags"} {
			if m, ok := pkg.methods["servicePackage"][name]; ok {
				f.follow(pkg, m)
			}
		}
	}

	calls := slices.AppendSeq(make([]string, 0, len(f.calls)), maps.Keys(f.calls))
	slices.Sort(calls)
	return calls
}

// follow walks a declaration in pkg, once.
func (f *callFinder) follow(pkg *servicePackage, decl ast.Node) {
	if decl == nil || f.visited[decl] {
		return
	}
	f.visited[decl] = true
	ast.Walk(&declVisitor{finder: f, pkg: pkg, imports: f.importsOf(decl), varTypes: make(map[string]string)}, decl)
}

// declVisitor records the SDK client method calls in a declaration, and follows the declarations it references:
// package-level declarations in pkg, exported declarations in other service packages,
// and methods in pkg called on a value whose type is known.
//
// Types are tracked syntactically, as strings: an AWS SDK client by its package's import path
// (e.g. "github.com/aws/aws-sdk-go-v2/service/s3"), and any other type by its unqualified name.
type declVisitor struct {
	finder  *callFinder
	pkg     *servicePackage
	imports map[string]string
	// varTypes maps variable names to their types, where known.
	// Scoping is ignored: names are shared by all nested function literals.
	varTypes map[string]string
}

func (v *declVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Recv != nil {
			v.setParamTypes(n.Recv)
		}

	case *ast.FuncType:
		v.setParamTypes(n.Params)

	case *ast.AssignStmt:
		// With several values on the left and one on the right (e.g. conn, err := f()), only the first value's type is known.
		for i, lhs := range n.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && (len(n.Lhs) == len(n.Rhs) || i == 0) {
				v.setVarType(id.Name, v.typeOf(n.Rhs[i]))
			}
		}

	case *ast.ValueSpec:
		for i, id := range n.Names {
			if n.Type != nil {
				v.setVarType(id.Name, resolveType(n.Type, v.imports))
			} else if i < len(n.Values) {
				v.setVarType(id.Name, v.typeOf(n.Values[i]))
			}
		}

	case *ast.CallExpr:
		sel, ok := n.Fun.(*ast.SelectorExpr)
		if !ok {
			break
		}
		if typ := v.typeOf(sel.X); isSDKPackage(typ) {
			if sel.Sel.Name != "Options" {
				v.finder.calls[path.Base(typ)+"."+sel.Sel.Name] = true
			}
		} else if x, ok := sel.X.(*ast.Ident); ok && isSDKPackage(v.imports[x.Name]) {
			// A paginator calls its operation, e.g. s3.NewListObjectsV2Paginator calls ListObjectsV2.
			if m := paginatorRegexp.FindStringSubmatch(sel.Sel.Name); m != nil {
				v.finder.calls[path.Base(v.imports[x.Name])+"."+m[1]] = true
			}
		}

	case *ast.SelectorExpr:
		if x, ok := n.X.(*ast.Ident); ok {
			if importPath, ok := v.imports[x.Name]; ok {
				if other, ok := v.finder.servicesByPath[importPath]; ok {
					v.finder.follow(other, other.decls[n.Sel.Name])
				}
				return nil
			}
		}
		for _, m := range v.pkg.methodsOf(v.typeOf(n.X), n.Sel.Name) {
			v.finder.follow(v.pkg, m)
		}
		// Sel names a field or method, not a package-level declaration, so only X is walked.
		ast.Walk(v, n.X)
		return nil

	case *ast.KeyValueExpr:
		// A struct literal key names a field, not a package-level declaration, so only the value is walked.
		if _, ok := n.Key.(*ast.Ident); ok {
			ast.Walk(v, n.Value)
			return nil
		}

	case *ast.Ident:
		v.finder.follow(v.pkg, v.pkg.decls[n.Name])
	}

	return v
}

// setParamTypes records the types of a function's named parameters or receiver.
func (v *declVisitor) setParamTypes(params *ast.FieldList) {
	for _, field := range params.List {
		for _, id := range field.Names {
			v.setVarType(id.Name, resolveType(field.Type, v.imports))
		}
	}
}

// setVarType records the type of a variable, unless the type is unknown or the variable has an interface type.
// Keeping the interface type means its method calls dispatch to every implementation that may be assigned to it.
func (v *declVisitor) setVarType(name, typ string) {
	if _, isInterface := v.pkg.interfaces[v.varTypes[name]]; typ != "" && !isInterface {
		v.varTypes[name] = typ
	}
}

// typeOf returns the type of the value of expr, or "" if it is not known.
func (v *declVisitor) typeOf(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return v.varTypes[e.Name]
	case *ast.ParenExpr:
		return v.typeOf(e.X)
	case *ast.StarExpr:
		return v.typeOf(e.X)
	case *ast.UnaryExpr:
		return v.typeOf(e.X)
	case *ast.CompositeLit:
		return resolveType(e.Type, nil)
	case *ast.SelectorExpr:
		return v.pkg.fieldTypes[e.Sel.Name]
	case *ast.CallExpr:
		switch fun := e.Fun.(type) {
		case *ast.SelectorExpr:
			return v.finder.clientMethods[fun.Sel.Name]
		case *ast.Ident:
			if fn, ok := v.pkg.decls[fun.Name].(*ast.FuncDecl); ok && fn.Type.Results != nil {
				return resolveType(fn.Type.Results.List[0].Type, v.finder.importsOf(fn))
			}
		}
	}
	return ""
}

// methodsOf returns the methods named name on typ, or, if typ is an interface, on every type in pkg that implements it.
func (pkg *servicePackage) methodsOf(typ, name string) []*ast.FuncDecl {
	if typ == "" {
		return nil
	}
	if m, ok := pkg.methods[typ][name]; ok {
		return []*ast.FuncDecl{m}
	}

	var methods []*ast.FuncDecl
	if names, ok := pkg.interfaces[typ]; ok {
		for _, ms := range pkg.methods {
			implements := !slices.ContainsFunc(names, func(n string) bool { return ms[n] == nil })
			if m, ok := ms[name]; ok && implements {
				methods = append(methods, m)
			}
		}
	}
	return methods
}

// resolveType returns the type denoted by the type expression expr: an SDK client's import path,
// or the name of an unqualified type. It returns "" for any other type.
func resolveType(expr ast.Expr, imports map[string]string) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return resolveType(e.X, imports)
	case *ast.IndexExpr:
		return resolveType(e.X, imports)
	case *ast.IndexListExpr:
		return resolveType(e.X, imports)
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok && e.Sel.Name == "Client" && isSDKPackage(imports[x.Name]) {
			return imports[x.Name]
		}
	}
	return ""
}

// isSDKPackage reports whether importPath is an AWS SDK service package.
func isSDKPackage(importPath string) bool {
	return strings.HasPrefix(importPath, sdkPathPrefix)
}

// schemaLiteral returns the literal if expr is &schema.typeName{...}, or nil.
func schemaLiteral(expr ast.Expr, typeName string) *ast.CompositeLit {
	if u, ok := expr.(*ast.UnaryExpr); ok && u.Op == token.AND {
		if lit, ok := u.X.(*ast.CompositeLit); ok && isSelector(lit.Type, "schema", typeName) {
			return lit
		}
	}
	return nil
}

// fieldValues returns the values set in a struct literal, keyed by field name.
func fieldValues(lit *ast.CompositeLit) map[string]ast.Expr {
	values := make(map[string]ast.Expr)
	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if key, ok := kv.Key.(*ast.Ident); ok {
				values[key.Name] = kv.Value
			}
		}
	}
	return values
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
