package awsops

import (
	"go/ast"
)

// funcEntry represents a function or method in the package index.
type funcEntry struct {
	Node     *ast.FuncDecl
	Receiver string // empty for standalone functions
}

// buildFuncIndex creates a map of function/method names to their AST nodes.
// For methods, the key is "ReceiverType.MethodName".
// For standalone functions, the key is the function name.
func buildFuncIndex(pkg *ast.Package) map[string]*ast.FuncDecl {
	index := make(map[string]*ast.FuncDecl)

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				recvType := receiverTypeName(fn.Recv.List[0].Type)
				if recvType != "" {
					index[recvType+"."+fn.Name.Name] = fn
				}
			} else {
				index[fn.Name.Name] = fn
			}
		}
	}

	return index
}

// receiverTypeName extracts the type name from a receiver expression,
// handling both pointer and value receivers.
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return receiverTypeName(t.X)
	case *ast.IndexListExpr:
		return receiverTypeName(t.X)
	}
	return ""
}

// buildImportIndex maps import aliases to import paths for all files in the package.
// The key is the local name (alias or last path component), the value is the import path.
func buildImportIndex(pkg *ast.Package) map[string]map[string]string {
	index := make(map[string]map[string]string)

	for filename, file := range pkg.Files {
		fileImports := make(map[string]string)
		for _, imp := range file.Imports {
			path := imp.Path.Value
			path = path[1 : len(path)-1] // strip quotes

			var name string
			if imp.Name != nil {
				name = imp.Name.Name
			} else {
				parts := splitImportPath(path)
				name = parts[len(parts)-1]
			}
			fileImports[name] = path
		}
		index[filename] = fileImports
	}

	return index
}

func splitImportPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			parts = append(parts, path[start:i])
			start = i + 1
		}
	}
	parts = append(parts, path[start:])
	return parts
}
