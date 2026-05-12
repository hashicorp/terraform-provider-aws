// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/tools/check-basic-test/attrnames"
)

// Attribute represents a schema attribute with its path and properties.
type Attribute struct {
	Path     string // dot-separated path, e.g. "key_schema.attribute_name"
	Required bool
	Optional bool
	Computed bool
}

// ParseFrameworkSchema parses a Go source file and extracts the Framework schema attributes.
// It looks for the Schema method pattern: response.Schema = schema.Schema{...}
func ParseFrameworkSchema(filename string) ([]Attribute, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var attrs []Attribute
	ast.Inspect(f, func(n ast.Node) bool {
		// Look for: response.Schema = s (where s was assigned schema.Schema{...})
		// or: response.Schema = schema.Schema{...}
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Check if LHS is "response.Schema"
		if !isResponseSchema(assign) {
			return true
		}

		// RHS could be a composite literal or an identifier
		switch rhs := assign.Rhs[0].(type) {
		case *ast.CompositeLit:
			attrs = parseSchemaLiteral(rhs, "")
		case *ast.Ident:
			// Find the variable assignment in the same function
			attrs = findSchemaVar(f, rhs.Name)
		}
		return true
	})

	return attrs, nil
}

func isResponseSchema(assign *ast.AssignStmt) bool {
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return false
	}
	sel, ok := assign.Lhs[0].(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "response" && sel.Sel.Name == "Schema"
}

// findSchemaVar finds a variable assignment like `s := schema.Schema{...}` in the file.
func findSchemaVar(f *ast.File, varName string) []Attribute {
	var attrs []Attribute
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		if len(assign.Lhs) != 1 {
			return true
		}
		ident, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || ident.Name != varName {
			return true
		}
		lit, ok := assign.Rhs[0].(*ast.CompositeLit)
		if !ok {
			return true
		}
		attrs = parseSchemaLiteral(lit, "")
		return false
	})
	return attrs
}

// parseSchemaLiteral parses a schema.Schema{} composite literal.
func parseSchemaLiteral(lit *ast.CompositeLit, prefix string) []Attribute {
	var attrs []Attribute
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := identName(kv.Key)
		if key == "Attributes" {
			attrs = append(attrs, parseAttributesMap(kv.Value, prefix)...)
		} else if key == "Blocks" {
			attrs = append(attrs, parseBlocksMap(kv.Value, prefix)...)
		}
	}
	return attrs
}

// parseAttributesMap parses map[string]schema.Attribute{...}
func parseAttributesMap(expr ast.Expr, prefix string) []Attribute {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var attrs []Attribute
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		name := resolveAttrName(kv.Key)
		if name == "" {
			continue
		}
		path := joinPath(prefix, name)

		// Skip timeouts block - it's handled by the framework
		if name == "timeouts" {
			continue
		}

		attr := Attribute{Path: path}
		// Parse the attribute literal to get Required/Optional/Computed
		if attrLit, ok := kv.Value.(*ast.CompositeLit); ok {
			attr.Required, attr.Optional, attr.Computed = parseAttrProperties(attrLit)
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

// parseBlocksMap parses map[string]schema.Block{...}
func parseBlocksMap(expr ast.Expr, prefix string) []Attribute {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var attrs []Attribute
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		name := resolveAttrName(kv.Key)
		if name == "" {
			continue
		}
		path := joinPath(prefix, name)

		// Skip timeouts block
		if name == "timeouts" {
			continue
		}

		// The block itself is an attribute that should be checked
		attrs = append(attrs, Attribute{Path: path})

		// Parse nested attributes within the block
		if blockLit, ok := kv.Value.(*ast.CompositeLit); ok {
			attrs = append(attrs, parseBlockLiteral(blockLit, path)...)
		}
	}
	return attrs
}

// parseBlockLiteral parses a block composite literal (e.g., schema.ListNestedBlock{...})
// to extract nested attributes from NestedObject.
func parseBlockLiteral(lit *ast.CompositeLit, prefix string) []Attribute {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := identName(kv.Key)
		if key == "NestedObject" {
			if nestedLit, ok := kv.Value.(*ast.CompositeLit); ok {
				return parseNestedBlockObject(nestedLit, prefix)
			}
		}
	}
	return nil
}

// parseNestedBlockObject parses schema.NestedBlockObject{Attributes: ...}
func parseNestedBlockObject(lit *ast.CompositeLit, prefix string) []Attribute {
	var attrs []Attribute
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := identName(kv.Key)
		if key == "Attributes" {
			attrs = append(attrs, parseAttributesMap(kv.Value, prefix)...)
		} else if key == "Blocks" {
			attrs = append(attrs, parseBlocksMap(kv.Value, prefix)...)
		}
	}
	return attrs
}

// parseAttrProperties extracts Required, Optional, Computed from an attribute literal.
func parseAttrProperties(lit *ast.CompositeLit) (required, optional, computed bool) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := identName(kv.Key)
		switch key {
		case "Required":
			required = isTrueIdent(kv.Value)
		case "Optional":
			optional = isTrueIdent(kv.Value)
		case "Computed":
			computed = isTrueIdent(kv.Value)
		}
	}
	return
}

// resolveAttrName resolves an attribute name from a map key expression.
// Handles both string literals ("name") and constant references (names.AttrARN).
func resolveAttrName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			return strings.Trim(e.Value, `"`)
		}
	case *ast.SelectorExpr:
		// names.AttrARN -> resolve from the constant name
		return resolveNameConstant(e.Sel.Name)
	case *ast.Ident:
		// Local constant
		return ""
	}
	return ""
}

// resolveNameConstant converts a constant name like "AttrARN" to its value "arn".
func resolveNameConstant(name string) string {
	return attrnames.Resolve(name)
}

func identName(expr ast.Expr) string {
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func isTrueIdent(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "true"
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
