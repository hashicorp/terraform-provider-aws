package awsops

import (
	"go/ast"
	"strings"
)

const awsSDKv2Prefix = "github.com/aws/aws-sdk-go-v2/service/"

// extractAWSOperations walks a function's AST, following intra-package calls,
// and collects all AWS SDK v2 API operation names.
func extractAWSOperations(fn *ast.FuncDecl, funcIndex map[string]*ast.FuncDecl, importIndex map[string]map[string]string, funcFileIndex map[string]string, visited map[string]bool) []string {
	if fn == nil || fn.Body == nil {
		return nil
	}

	key := funcKey(fn)
	if visited[key] {
		return nil
	}
	visited[key] = true

	// Determine which file this function lives in for import resolution.
	filename := funcFileIndex[key]
	fileImports := importIndex[filename]

	var ops []string
	seen := make(map[string]bool)

	// Track variables that hold AWS SDK client references.
	clientVars := findClientVars(fn, fileImports)

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check for method calls: expr.MethodName(args...)
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			methodName := sel.Sel.Name

			if ident, ok := sel.X.(*ast.Ident); ok {
				// Check if caller is a known SDK client variable.
				if clientVars[ident.Name] {
					if !seen[methodName] {
						seen[methodName] = true
						ops = append(ops, methodName)
					}
					return true
				}

				// Check if caller is an imported AWS SDK service package (not sub-packages like /types).
				if path, ok := fileImports[ident.Name]; ok && isAWSServicePackage(path) {
					if !seen[methodName] {
						seen[methodName] = true
						ops = append(ops, methodName)
					}
					return true
				}

				// Follow method calls on package-local types.
				for fname, f := range funcIndex {
					if strings.HasSuffix(fname, "."+methodName) && f != fn {
						childOps := extractAWSOperations(f, funcIndex, importIndex, funcFileIndex, visited)
						for _, op := range childOps {
							if !seen[op] {
								seen[op] = true
								ops = append(ops, op)
							}
						}
						break
					}
				}
			}
		}

		// Check for standalone function calls: funcName(args...)
		if ident, ok := call.Fun.(*ast.Ident); ok {
			if f, ok := funcIndex[ident.Name]; ok {
				childOps := extractAWSOperations(f, funcIndex, importIndex, funcFileIndex, visited)
				for _, op := range childOps {
					if !seen[op] {
						seen[op] = true
						ops = append(ops, op)
					}
				}
			}
		}

		return true
	})

	return ops
}

// findClientVars identifies variables that hold AWS SDK client references.
// It checks:
//  1. Function parameters with AWS SDK client types (e.g., conn *s3.Client)
//  2. Assignments from AWSClient accessor calls (e.g., meta.(*conns.AWSClient).S3Client(ctx))
//  3. Variable declarations with AWS SDK client types (e.g., var conn *s3.Client)
func findClientVars(fn *ast.FuncDecl, fileImports map[string]string) map[string]bool {
	clients := make(map[string]bool)

	// Check function parameters for AWS SDK client types.
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			if isAWSClientType(param.Type, fileImports) {
				for _, name := range param.Names {
					clients[name.Name] = true
				}
			}
		}
	}

	if fn.Body == nil {
		return clients
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			// conn := meta.(*conns.AWSClient).FooClient(ctx)
			for i, rhs := range stmt.Rhs {
				if isClientAccessor(rhs) {
					if i < len(stmt.Lhs) {
						if ident, ok := stmt.Lhs[i].(*ast.Ident); ok {
							clients[ident.Name] = true
						}
					}
				}
			}
		case *ast.DeclStmt:
			// var conn *s3.Client
			gd, ok := stmt.Decl.(*ast.GenDecl)
			if !ok {
				return true
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				if vs.Type != nil && isAWSClientType(vs.Type, fileImports) {
					for _, name := range vs.Names {
						clients[name.Name] = true
					}
				}
			}
		}
		return true
	})

	return clients
}

// isAWSClientType checks if a type expression refers to an AWS SDK client type.
// Matches *pkg.Client where pkg is imported from aws-sdk-go-v2/service/.
func isAWSClientType(expr ast.Expr, fileImports map[string]string) bool {
	// Unwrap pointer.
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}

	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "Client" {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	if path, ok := fileImports[ident.Name]; ok {
		return strings.HasPrefix(path, awsSDKv2Prefix)
	}
	return false
}

// isClientAccessor checks if an expression is a call to an AWS client accessor.
// Matches: meta.(*conns.AWSClient).FooClient(ctx) or r.Meta().FooClient(ctx)
func isClientAccessor(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if !strings.HasSuffix(sel.Sel.Name, "Client") {
		return false
	}

	return containsClientChain(sel.X)
}

// containsClientChain checks if an expression chain references AWSClient or Meta().
func containsClientChain(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.CallExpr:
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Meta" {
				return true
			}
		}
		return containsClientChain(e.Fun)
	case *ast.SelectorExpr:
		if e.Sel.Name == "AWSClient" {
			return true
		}
		return containsClientChain(e.X)
	case *ast.TypeAssertExpr:
		return containsClientChain(e.Type) || containsClientChain(e.X)
	case *ast.StarExpr:
		return containsClientChain(e.X)
	case *ast.ParenExpr:
		return containsClientChain(e.X)
	case *ast.Ident:
		return e.Name == "AWSClient"
	}
	return false
}

// isAWSServicePackage returns true if the import path is a top-level AWS SDK v2
// service package (e.g. ".../service/s3") but not a sub-package (e.g. ".../service/s3/types").
func isAWSServicePackage(path string) bool {
	if !strings.HasPrefix(path, awsSDKv2Prefix) {
		return false
	}
	remainder := path[len(awsSDKv2Prefix):]
	return !strings.Contains(remainder, "/")
}

func funcKey(fn *ast.FuncDecl) string {
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recvType := receiverTypeName(fn.Recv.List[0].Type)
		return recvType + "." + fn.Name.Name
	}
	return fn.Name.Name
}
