package astutils

import (
	"go/ast"
)

// IsFunctionParameterTypeArrayString returns true if the expression matches []string
func IsFunctionParameterTypeArrayString(e ast.Expr) bool {
	arrayType, ok := e.(*ast.ArrayType)

	return ok && IsFunctionParameterTypeString(arrayType.Elt)
}

// IsFunctionParameterTypeArrayError returns true if the expression matches []error
func IsFunctionParameterTypeArrayError(e ast.Expr) bool {
	arrayType, ok := e.(*ast.ArrayType)

	return ok && IsFunctionParameterTypeError(arrayType.Elt)
}

// IsFunctionParameterTypeError returns true if the expression matches string
func IsFunctionParameterTypeError(e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)

	return ok && ident.Name == "error"
}

// IsFunctionParameterTypeInterface returns true if the expression matches interface{}
func IsFunctionParameterTypeInterface(e ast.Expr) bool {
	_, ok := e.(*ast.InterfaceType)

	return ok
}

// IsFunctionParameterTypeString returns true if the expression matches string
func IsFunctionParameterTypeString(e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)

	return ok && ident.Name == "string"
}
