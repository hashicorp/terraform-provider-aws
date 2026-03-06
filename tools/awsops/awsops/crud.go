package awsops

import (
	"go/ast"
)

// resolveCRUDFunctions returns a map of CRUD method names to their AST function nodes.
func resolveCRUDFunctions(res resourceInfo, pkg *ast.Package, funcIndex map[string]*ast.FuncDecl) map[string]*ast.FuncDecl {
	if res.Type == "sdk" {
		return resolveSDKCRUD(res, pkg, funcIndex)
	}
	return resolveFrameworkCRUD(res, pkg, funcIndex)
}

// resolveSDKCRUD finds CRUD functions for SDK-based resources by looking at
// the schema.Resource struct literal in the same file as the annotation.
func resolveSDKCRUD(res resourceInfo, pkg *ast.Package, funcIndex map[string]*ast.FuncDecl) map[string]*ast.FuncDecl {
	result := make(map[string]*ast.FuncDecl)
	file := pkg.Files[res.File]
	if file == nil {
		return result
	}

	sdkFields := map[string]string{
		"CreateWithoutTimeout": "create",
		"ReadWithoutTimeout":   "read",
		"UpdateWithoutTimeout": "update",
		"DeleteWithoutTimeout": "delete",
	}

	// Walk the file looking for schema.Resource composite literals.
	ast.Inspect(file, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if !isSchemaResourceType(cl.Type) {
			return true
		}

		for _, elt := range cl.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			ident, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}
			method, ok := sdkFields[ident.Name]
			if !ok {
				continue
			}
			// The value should be a function identifier.
			funcName := extractFuncName(kv.Value)
			if funcName == "" {
				continue
			}
			if fn, ok := funcIndex[funcName]; ok {
				result[method] = fn
			}
		}

		return true
	})

	return result
}

// resolveFrameworkCRUD finds CRUD methods for Framework-based resources.
// It identifies the resource struct type from the constructor function that
// appears near the annotation, then looks for Create/Read/Update/Delete methods.
func resolveFrameworkCRUD(res resourceInfo, pkg *ast.Package, funcIndex map[string]*ast.FuncDecl) map[string]*ast.FuncDecl {
	result := make(map[string]*ast.FuncDecl)
	file := pkg.Files[res.File]
	if file == nil {
		return result
	}

	// Find the struct type by looking for the constructor function near the annotation.
	structType := findFrameworkStructType(file)
	if structType == "" {
		return result
	}

	methods := map[string]string{
		"Create": "create",
		"Read":   "read",
		"Update": "update",
		"Delete": "delete",
	}

	for methodName, crudName := range methods {
		key := structType + "." + methodName
		if fn, ok := funcIndex[key]; ok {
			result[crudName] = fn
		}
	}

	return result
}

// findFrameworkStructType finds the resource struct type from a Framework resource file.
// It looks for a function that returns a struct literal with & (pointer), which is the
// common constructor pattern: func newFooResource(...) { r := &fooResource{} ... }
func findFrameworkStructType(file *ast.File) string {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}
		// Look for functions returning ResourceWithConfigure or similar.
		if fn.Type.Results == nil {
			continue
		}
		if fn.Body == nil {
			continue
		}

		// Search function body for &structType{} patterns.
		var found string
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if found != "" {
				return false
			}
			ue, ok := n.(*ast.UnaryExpr)
			if !ok {
				return true
			}
			cl, ok := ue.X.(*ast.CompositeLit)
			if !ok {
				return true
			}
			ident, ok := cl.Type.(*ast.Ident)
			if !ok {
				return true
			}
			// Check if this type has CRUD-like methods in the file.
			if isLikelyResourceStruct(ident.Name, file) {
				found = ident.Name
			}
			return true
		})
		if found != "" {
			return found
		}
	}
	return ""
}

// isLikelyResourceStruct checks if a struct type has methods that look like CRUD methods.
func isLikelyResourceStruct(typeName string, file *ast.File) bool {
	// Check type declarations in file for struct types with this name.
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if ts.Name.Name == typeName {
				_, isStruct := ts.Type.(*ast.StructType)
				return isStruct
			}
		}
	}
	return false
}

// isSchemaResourceType checks if an expression refers to schema.Resource.
func isSchemaResourceType(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "schema" && sel.Sel.Name == "Resource"
}

// extractFuncName extracts a function name from an expression.
func extractFuncName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return v.Sel.Name
	}
	return ""
}
