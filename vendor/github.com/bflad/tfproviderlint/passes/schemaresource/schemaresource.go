package schemaresource

import (
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "schemaresource",
	Doc:  "find github.com/hashicorp/terraform/helper/schema.Resource literals for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*ast.CompositeLit{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}
	var result []*ast.CompositeLit

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.CompositeLit)

		if !isSchemaResource(pass, x) {
			return
		}

		result = append(result, x)
	})

	return result, nil
}

func isSchemaResource(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	switch v := cl.Type.(type) {
	default:
		return false
	case *ast.SelectorExpr:
		switch t := pass.TypesInfo.TypeOf(v).(type) {
		default:
			return false
		case *types.Named:
			if t.Obj().Name() != "Resource" {
				return false
			}
			// HasSuffix here due to vendoring
			if !strings.HasSuffix(t.Obj().Pkg().Path(), "github.com/hashicorp/terraform/helper/schema") {
				return false
			}
		}
	}
	return true
}
