// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package attrnames

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// consts is a lookup table of constant name → value from attr_consts_gen.go.
var consts map[string]string

// LoadConsts parses a Go file with const declarations and adds them to the lookup table.
func LoadConsts(filename string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	if consts == nil {
		consts = make(map[string]string)
	}
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
				continue
			}
			name := vs.Names[0].Name
			if lit, ok := vs.Values[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				consts[name] = strings.Trim(lit.Value, `"`)
			}
		}
	}
	return nil
}

// Resolve converts a constant name like "AttrARN" to its string value "arn".
// Uses the lookup table if loaded, falls back to camelToSnake heuristic.
func Resolve(name string) string {
	if consts != nil {
		if v, ok := consts[name]; ok {
			return v
		}
	}
	if !strings.HasPrefix(name, "Attr") {
		return ""
	}
	trimmed := strings.TrimPrefix(name, "Attr")
	return CamelToSnake(trimmed)
}

func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := s[i-1]
				if prev >= 'a' && prev <= 'z' {
					result.WriteByte('_')
				} else if i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z' {
					result.WriteByte('_')
				}
			}
			result.WriteByte(byte(r) + 32)
		} else {
			result.WriteByte(byte(r))
		}
	}
	return result.String()
}
