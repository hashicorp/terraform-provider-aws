package acctestcase

import (
	"go/ast"
	"reflect"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "acctestcase",
	Doc:  "find github.com/hashicorp/terraform-plugin-sdk/helper/resource.TestCase literals for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*terraformtype.HelperResourceTestCaseInfo{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}
	var result []*terraformtype.HelperResourceTestCaseInfo

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.CompositeLit)

		if !isResourceTestCase(pass, x) {
			return
		}

		result = append(result, terraformtype.NewHelperResourceTestCaseInfo(x, pass.TypesInfo))
	})

	return result, nil
}

func isResourceTestCase(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	switch v := cl.Type.(type) {
	default:
		return false
	case *ast.SelectorExpr:
		return terraformtype.IsHelperResourceTypeTestCase(pass.TypesInfo.TypeOf(v))
	}
}
