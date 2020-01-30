package terraformtype

import (
	"go/ast"
	"strconv"
	"strings"
)

func astBoolValue(e ast.Expr) *bool {
	switch v := e.(type) {
	case *ast.Ident:
		stringValue := v.Name
		boolValue, err := strconv.ParseBool(stringValue)

		if err != nil {
			return nil
		}

		return &boolValue
	default:
		return nil
	}
}

func astCompositeLitField(cl *ast.CompositeLit, fieldName string) *ast.KeyValueExpr {
	for _, elt := range cl.Elts {
		switch e := elt.(type) {
		case *ast.KeyValueExpr:
			if e.Key.(*ast.Ident).Name != fieldName {
				continue
			}

			return e
		}
	}

	return nil
}

func astCompositeLitFields(cl *ast.CompositeLit) map[string]*ast.KeyValueExpr {
	result := make(map[string]*ast.KeyValueExpr, len(cl.Elts))

	for _, elt := range cl.Elts {
		switch e := elt.(type) {
		case *ast.KeyValueExpr:
			result[e.Key.(*ast.Ident).Name] = e
		}
	}

	return result
}

func astCompositeLitFieldBoolValue(cl *ast.CompositeLit, fieldName string) *bool {
	kvExpr := astCompositeLitField(cl, fieldName)

	if kvExpr == nil {
		return nil
	}

	return astBoolValue(kvExpr.Value)
}

func astCompositeLitFieldExprValue(cl *ast.CompositeLit, fieldName string) *ast.Expr {
	kvExpr := astCompositeLitField(cl, fieldName)

	if kvExpr == nil {
		return nil
	}

	return astExprValue(kvExpr.Value)
}

func astCompositeLitFieldIntValue(cl *ast.CompositeLit, fieldName string) *int {
	kvExpr := astCompositeLitField(cl, fieldName)

	if kvExpr == nil {
		return nil
	}

	return astIntValue(kvExpr.Value)
}

func astCompositeLitContainsAnyField(cl *ast.CompositeLit, fieldNames ...string) bool {
	for _, elt := range cl.Elts {
		switch e := elt.(type) {
		case *ast.KeyValueExpr:
			name := e.Key.(*ast.Ident).Name

			for _, field := range fieldNames {
				if name == field {
					return true
				}
			}
		}
	}

	return false
}

func astExprValue(e ast.Expr) *ast.Expr {
	switch v := e.(type) {
	case *ast.Ident:
		if v.Name == "nil" {
			return nil
		}

		return &e
	default:
		return &e
	}
}

func astIntValue(e ast.Expr) *int {
	switch v := e.(type) {
	case *ast.BasicLit:
		stringValue := strings.Trim(v.Value, `"`)
		intValue, err := strconv.Atoi(stringValue)

		if err != nil {
			return nil
		}

		return &intValue
	default:
		return nil
	}
}

func astStringValue(e ast.Expr) *string {
	switch v := e.(type) {
	case *ast.BasicLit:
		stringValue := strings.Trim(v.Value, `"`)

		return &stringValue
	default:
		return nil
	}
}
