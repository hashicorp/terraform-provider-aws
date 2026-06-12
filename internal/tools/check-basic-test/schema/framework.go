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
// It finds the Schema method by signature and extracts the response parameter name dynamically.
func ParseFrameworkSchema(filename string) ([]Attribute, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	p := &fwParser{file: f}

	// Find the Schema method and determine the response parameter name
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name.Name != "Schema" {
			continue
		}
		respParamName := schemaResponseParamName(fn)
		if respParamName == "" {
			continue
		}

		var attrs []Attribute
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
				return true
			}
			sel, ok := assign.Lhs[0].(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != respParamName || sel.Sel.Name != "Schema" {
				return true
			}
			switch rhs := assign.Rhs[0].(type) {
			case *ast.CompositeLit:
				attrs = p.parseSchemaLiteral(rhs, "")
			case *ast.Ident:
				attrs = p.findSchemaVar(rhs.Name)
			}
			return true
		})
		if len(attrs) > 0 {
			return attrs, nil
		}
	}

	return nil, nil
}

// fwParser holds file context for resolving function calls during schema parsing.
type fwParser struct {
	file *ast.File
}

// schemaResponseParamName returns the name of the *resource.SchemaResponse parameter,
// or "" if the function doesn't match the expected signature.
func schemaResponseParamName(fn *ast.FuncDecl) string {
	if fn.Type.Params == nil {
		return ""
	}
	for _, field := range fn.Type.Params.List {
		star, ok := field.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		sel, ok := star.X.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if sel.Sel.Name == "SchemaResponse" {
			if len(field.Names) > 0 {
				return field.Names[0].Name
			}
		}
	}
	return ""
}

// findSchemaVar finds a variable assignment like `s := schema.Schema{...}` in the file.
func (p *fwParser) findSchemaVar(varName string) []Attribute {
	var attrs []Attribute
	ast.Inspect(p.file, func(n ast.Node) bool {
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
		attrs = p.parseSchemaLiteral(lit, "")
		return false
	})
	return attrs
}

// parseSchemaLiteral parses a schema.Schema{} composite literal.
func (p *fwParser) parseSchemaLiteral(lit *ast.CompositeLit, prefix string) []Attribute {
	var attrs []Attribute
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := identName(kv.Key)
		if key == "Attributes" {
			attrs = append(attrs, p.parseAttributesMap(kv.Value, prefix)...)
		} else if key == "Blocks" {
			attrs = append(attrs, p.parseBlocksMap(kv.Value, prefix)...)
		}
	}
	return attrs
}

// parseAttributesMap parses map[string]schema.Attribute{...}
func (p *fwParser) parseAttributesMap(expr ast.Expr, prefix string) []Attribute {
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

		if name == "timeouts" {
			continue
		}

		attr := Attribute{Path: path}
		if attrLit, ok := kv.Value.(*ast.CompositeLit); ok {
			attr.Required, attr.Optional, attr.Computed = parseAttrProperties(attrLit)
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

// parseBlocksMap parses map[string]schema.Block{...}
func (p *fwParser) parseBlocksMap(expr ast.Expr, prefix string) []Attribute {
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

		if name == "timeouts" {
			continue
		}

		attrs = append(attrs, Attribute{Path: path})

		// Parse nested attributes — handle both inline literals and function calls
		switch v := kv.Value.(type) {
		case *ast.CompositeLit:
			attrs = append(attrs, p.parseBlockLiteral(v, path)...)
		case *ast.CallExpr:
			if blockLit := p.resolveBlockFunc(v); blockLit != nil {
				attrs = append(attrs, p.parseBlockLiteral(blockLit, path)...)
			}
		}
	}
	return attrs
}

// resolveBlockFunc resolves a function call like exportDataQuerySchema(ctx) to its
// returned composite literal by finding the function in the file.
func (p *fwParser) resolveBlockFunc(call *ast.CallExpr) *ast.CompositeLit {
	funcName := ""
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		funcName = fn.Name
	case *ast.SelectorExpr:
		return nil // external package call, can't resolve
	}
	if funcName == "" {
		return nil
	}

	for _, decl := range p.file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != funcName || fn.Body == nil {
			continue
		}
		// Find the return statement with a composite literal
		var result *ast.CompositeLit
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			ret, ok := n.(*ast.ReturnStmt)
			if !ok || len(ret.Results) == 0 {
				return true
			}
			if lit, ok := ret.Results[0].(*ast.CompositeLit); ok {
				result = lit
				return false
			}
			return true
		})
		return result
	}
	return nil
}

// parseBlockLiteral parses a block composite literal (e.g., schema.ListNestedBlock{...})
func (p *fwParser) parseBlockLiteral(lit *ast.CompositeLit, prefix string) []Attribute {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := identName(kv.Key)
		if key == "NestedObject" {
			if nestedLit, ok := kv.Value.(*ast.CompositeLit); ok {
				return p.parseNestedBlockObject(nestedLit, prefix)
			}
		}
	}
	return nil
}

// parseNestedBlockObject parses schema.NestedBlockObject{Attributes: ..., Blocks: ...}
func (p *fwParser) parseNestedBlockObject(lit *ast.CompositeLit, prefix string) []Attribute {
	var attrs []Attribute
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := identName(kv.Key)
		if key == "Attributes" {
			attrs = append(attrs, p.parseAttributesMap(kv.Value, prefix)...)
		} else if key == "Blocks" {
			attrs = append(attrs, p.parseBlocksMap(kv.Value, prefix)...)
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
