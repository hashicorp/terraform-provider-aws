package acctestcase

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
	Name: "acctestcase",
	Doc:  "find github.com/hashicorp/terraform/helper/resource.TestCase literals for later passes",
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

		if !isResourceTestCase(pass, x) {
			return
		}

		result = append(result, x)
	})

	return result, nil
}

func isResourceTestCase(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	switch v := cl.Type.(type) {
	default:
		return false
	case *ast.SelectorExpr:
		switch t := pass.TypesInfo.TypeOf(v).(type) {
		default:
			return false
		case *types.Named:
			if t.Obj().Name() != "TestCase" {
				return false
			}
			// HasSuffix here due to vendoring
			if !strings.HasSuffix(t.Obj().Pkg().Path(), "github.com/hashicorp/terraform/helper/resource") {
				return false
			}
		}
	}
	return true
}
