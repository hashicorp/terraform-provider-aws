// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package testparser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/tools/check-basic-test/attrnames"
)

// CheckedAttribute represents an attribute that is checked in a test.
type CheckedAttribute struct {
	Path      string // dot-separated path
	Source    string // "Check" or "ConfigStateChecks"
	CountOnly bool   // true if only the count (# or %) is checked, not values
	Value     string // the asserted value (only populated for count-only checks)
}

// ParseBasicTest parses a test file and extracts checked attributes from the _basic test function.
// resourceName is the Terraform resource address (e.g., "aws_dynamodb_global_secondary_index.test").
func ParseBasicTest(filename string, testFuncName string, resourceName string) ([]CheckedAttribute, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Find the test function
	var testFunc *ast.FuncDecl
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name == testFuncName {
			testFunc = fn
			break
		}
	}
	if testFunc == nil {
		return nil, nil
	}

	// Resolve the resourceName variable value
	resolvedResourceName := resolveResourceName(testFunc, resourceName)

	var checked []CheckedAttribute

	// Walk the function body looking for TestStep structs
	ast.Inspect(testFunc.Body, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Look for TestStep-like composite literals with Check or ConfigStateChecks fields
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key := identName(kv.Key)
			switch key {
			case "Check":
				checked = append(checked, parseCheckField(kv.Value, resolvedResourceName)...)
			case "ConfigStateChecks":
				checked = append(checked, parseConfigStateChecks(kv.Value, resolvedResourceName)...)
			}
		}
		return true
	})

	return checked, nil
}

// resolveResourceName finds the value of a variable like `resourceName := "aws_example.test"`
func resolveResourceName(fn *ast.FuncDecl, varNameOrLiteral string) string {
	// If it looks like a resource address already, return it
	if strings.Contains(varNameOrLiteral, ".") {
		return varNameOrLiteral
	}

	// Search for assignment in the function body
	var resolved string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range assign.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok || id.Name != varNameOrLiteral {
				continue
			}
			if i < len(assign.Rhs) {
				if lit, ok := assign.Rhs[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					resolved = strings.Trim(lit.Value, `"`)
				}
			}
		}
		return true
	})
	if resolved != "" {
		return resolved
	}
	return varNameOrLiteral
}

// parseCheckField extracts attributes from resource.ComposeTestCheckFunc(...) or ComposeAggregateTestCheckFunc(...)
func parseCheckField(expr ast.Expr, resourceName string) []CheckedAttribute {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil
	}

	var attrs []CheckedAttribute
	for _, arg := range call.Args {
		argCall, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}
		attrs = append(attrs, parseCheckCall(argCall, resourceName)...)
	}
	return attrs
}

// parseCheckCall handles individual check function calls within Check.
func parseCheckCall(call *ast.CallExpr, resourceName string) []CheckedAttribute {
	funcName := callFuncName(call)

	switch funcName {
	case "resource.TestCheckResourceAttr",
		"resource.TestCheckResourceAttrSet",
		"resource.TestCheckNoResourceAttr",
		"resource.TestCheckResourceAttrPair",
		"resource.TestMatchResourceAttr":
		// Pattern: func(resourceName, attrName, ...)
		if len(call.Args) >= 2 && matchesResource(call.Args[0], resourceName) {
			if name := resolveStringArg(call.Args[1]); name != "" {
				path, countOnly := normalizePath(name)
				var value string
				if countOnly && len(call.Args) >= 3 {
					value = resolveStringArg(call.Args[2])
				}
				return []CheckedAttribute{{Path: path, Source: "Check", CountOnly: countOnly, Value: value}}
			}
		}
	case "acctest.CheckResourceAttrRegionalARN",
		"acctest.CheckResourceAttrRegionalARNFormat",
		"acctest.CheckResourceAttrGlobalARN",
		"acctest.CheckResourceAttrGlobalARNFormat",
		"acctest.MatchResourceAttrRegionalARN",
		"acctest.MatchResourceAttrRegionalARNRegion",
		"acctest.MatchResourceAttrGlobalARN",
		"acctest.CheckResourceAttrAccountID":
		// Pattern: func(ctx, resourceName, attrName, ...) or func(resourceName, attrName, ...)
		// Find the attr name argument (after ctx and resourceName)
		for i, arg := range call.Args {
			if matchesResource(arg, resourceName) && i+1 < len(call.Args) {
				if name := resolveStringArg(call.Args[i+1]); name != "" {
					path, countOnly := normalizePath(name)
					return []CheckedAttribute{{Path: path, Source: "Check", CountOnly: countOnly}}
				}
			}
		}
	case "acctest.CheckResourceAttrRFC3339",
		"acctest.CheckResourceAttrEquivalentJSON":
		// Pattern: func(resourceName, attrName)
		if len(call.Args) >= 2 && matchesResource(call.Args[0], resourceName) {
			if name := resolveStringArg(call.Args[1]); name != "" {
				path, countOnly := normalizePath(name)
				return []CheckedAttribute{{Path: path, Source: "Check", CountOnly: countOnly}}
			}
		}
	}
	return nil
}

// parseConfigStateChecks extracts attributes from []statecheck.StateCheck{...}
func parseConfigStateChecks(expr ast.Expr, resourceName string) []CheckedAttribute {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}

	var attrs []CheckedAttribute
	for _, elt := range lit.Elts {
		call, ok := elt.(*ast.CallExpr)
		if !ok {
			continue
		}
		attrs = append(attrs, parseStateCheckCall(call, resourceName, "")...)
	}
	return attrs
}

// parseStateCheckCall handles statecheck.ExpectKnownValue, CompareValuePairs, etc.
func parseStateCheckCall(call *ast.CallExpr, resourceName string, prefix string) []CheckedAttribute {
	funcName := callFuncName(call)

	switch funcName {
	case "statecheck.ExpectKnownValue":
		// Pattern: statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("attr"), knownvalue.Check)
		if len(call.Args) >= 3 && matchesResource(call.Args[0], resourceName) {
			path := parseTFJsonPath(call.Args[1])
			if path == "" {
				return nil
			}
			fullPath := joinPath(prefix, path)
			attrs := []CheckedAttribute{{Path: fullPath, Source: "ConfigStateChecks"}}
			// Parse nested object checks for sub-attributes
			attrs = append(attrs, parseKnownValueForNested(call.Args[2], fullPath)...)
			return attrs
		}
	case "statecheck.CompareValuePairs":
		// Pattern: statecheck.CompareValuePairs(resourceName, tfjsonpath, resourceName2, tfjsonpath2, compare)
		if len(call.Args) >= 2 && matchesResource(call.Args[0], resourceName) {
			path := parseTFJsonPath(call.Args[1])
			if path != "" {
				return []CheckedAttribute{{Path: joinPath(prefix, path), Source: "ConfigStateChecks"}}
			}
		}
	case "tfstatecheck.ExpectAttributeFormat":
		// Pattern: tfstatecheck.ExpectAttributeFormat(resourceName, tfjsonpath.New("attr"), format)
		if len(call.Args) >= 2 && matchesResource(call.Args[0], resourceName) {
			path := parseTFJsonPath(call.Args[1])
			if path != "" {
				return []CheckedAttribute{{Path: joinPath(prefix, path), Source: "ConfigStateChecks"}}
			}
		}
	}
	return nil
}

// parseKnownValueForNested extracts nested attribute names from knownvalue checks.
// e.g., knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{...})})
func parseKnownValueForNested(expr ast.Expr, parentPath string) []CheckedAttribute {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil
	}

	funcName := callFuncName(call)
	switch funcName {
	case "knownvalue.ObjectExact", "knownvalue.ObjectPartial":
		return parseObjectCheckMap(call, parentPath)
	case "knownvalue.ListExact", "knownvalue.ListPartial",
		"knownvalue.SetExact", "knownvalue.SetPartial":
		// Look inside the slice for ObjectExact calls
		if len(call.Args) >= 1 {
			return parseListForObjects(call.Args[0], parentPath)
		}
	}
	return nil
}

// parseObjectCheckMap extracts keys from map[string]knownvalue.Check{...}
func parseObjectCheckMap(call *ast.CallExpr, parentPath string) []CheckedAttribute {
	if len(call.Args) < 1 {
		return nil
	}
	lit, ok := call.Args[0].(*ast.CompositeLit)
	if !ok {
		return nil
	}

	var attrs []CheckedAttribute
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		name := resolveStringArg(kv.Key)
		if name == "" {
			continue
		}
		fullPath := joinPath(parentPath, name)
		attrs = append(attrs, CheckedAttribute{Path: fullPath, Source: "ConfigStateChecks"})
		// Recurse into nested checks
		attrs = append(attrs, parseKnownValueForNested(kv.Value, fullPath)...)
	}
	return attrs
}

// parseListForObjects looks inside a slice literal for ObjectExact calls.
func parseListForObjects(expr ast.Expr, parentPath string) []CheckedAttribute {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var attrs []CheckedAttribute
	for _, elt := range lit.Elts {
		call, ok := elt.(*ast.CallExpr)
		if !ok {
			continue
		}
		attrs = append(attrs, parseKnownValueForNested(call, parentPath)...)
	}
	return attrs
}

// parseTFJsonPath extracts the path from tfjsonpath.New("attr").AtMapKey("sub") chains.
func parseTFJsonPath(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.CallExpr:
		funcName := callFuncName(e)
		switch funcName {
		case "tfjsonpath.New":
			if len(e.Args) >= 1 {
				return resolveStringArg(e.Args[0])
			}
		}
		// Handle chained calls like tfjsonpath.New("x").AtMapKey("y")
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
			base := parseTFJsonPath(sel.X)
			switch sel.Sel.Name {
			case "AtMapKey":
				if len(e.Args) >= 1 {
					key := resolveStringArg(e.Args[0])
					if key != "" {
						return joinPath(base, key)
					}
				}
			case "AtSliceIndex":
				// Slice index doesn't change the attribute path for our purposes
				return base
			}
		}
	}
	return ""
}

// matchesResource checks if an expression matches the resource name variable or literal.
func matchesResource(expr ast.Expr, resourceName string) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		// Variable reference like `resourceName`
		return e.Name == "resourceName" || e.Name == resourceName
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			return strings.Trim(e.Value, `"`) == resourceName
		}
	}
	return false
}

// resolveStringArg resolves a string argument that could be a literal or a names.Attr* constant.
func resolveStringArg(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			return strings.Trim(e.Value, `"`)
		}
	case *ast.SelectorExpr:
		if id, ok := e.X.(*ast.Ident); ok {
			switch id.Name {
			case "names":
				// names.AttrARN -> "arn"
				return resolveNameConstant(e.Sel.Name)
			case "acctest":
				// acctest.CtTagsAllPercent -> "tags_all.%"
				return resolveAcctestConstant(e.Sel.Name)
			}
		}
	}
	return ""
}

// resolveAcctestConstant resolves acctest.Ct* constants to their string values.
// e.g., CtTagsAllPercent -> "tags_all.%"
func resolveAcctestConstant(name string) string {
	if !strings.HasPrefix(name, "Ct") {
		return ""
	}
	// Load from the attrnames lookup if available
	return attrnames.Resolve(name)
}

// resolveNameConstant converts names.AttrXxx to the snake_case value.
func resolveNameConstant(name string) string {
	return attrnames.Resolve(name)
}

func callFuncName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		if id, ok := fn.X.(*ast.Ident); ok {
			return id.Name + "." + fn.Sel.Name
		}
	case *ast.Ident:
		return fn.Name
	}
	return ""
}

func identName(expr ast.Expr) string {
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

// normalizePath converts SDKv2-style dot-indexed paths to schema paths.
// e.g., "destination.0.cloudwatch_logs.0.log_group_arn" → "destination.cloudwatch_logs.log_group_arn"
// e.g., "destination.#" → "destination"
// Returns the normalized path and whether it's a count-only check (ends in # or %).
func normalizePath(path string) (string, bool) {
	parts := strings.Split(path, ".")
	countOnly := false
	var result []string
	for _, p := range parts {
		if p == "#" || p == "%" {
			countOnly = true
			continue
		}
		if len(p) > 0 && p[0] >= '0' && p[0] <= '9' {
			continue
		}
		result = append(result, p)
	}
	return strings.Join(result, "."), countOnly
}
