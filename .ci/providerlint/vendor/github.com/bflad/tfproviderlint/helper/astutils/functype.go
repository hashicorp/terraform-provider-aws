// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package astutils

import (
	"go/ast"
)

func FuncTypeFromNode(node ast.Node) *ast.FuncType {
	switch node := node.(type) {
	case *ast.FuncDecl:
		return node.Type
	case *ast.FuncLit:
		return node.Type
	}

	return nil
}
