// Package AT005 defines an Analyzer that checks for
// acceptance tests prefixed with Test but not TestAcc
package AT005

import (
	"go/ast"
	"strings"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/testfunc"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for acceptance test function names missing TestAcc prefix

The AT005 analyzer reports test function names (Test prefix) that contain
resource.Test() or resource.ParallelTest(), which should be named with
the TestAcc prefix.`

const analyzerName = "AT005"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		testfunc.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	testFuncs := pass.ResultOf[testfunc.Analyzer].([]*ast.FuncDecl)

	for _, testFunc := range testFuncs {
		if strings.HasPrefix(testFunc.Name.Name, "TestAcc") {
			continue
		}

		ast.Inspect(testFunc.Body, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)

			if !ok {
				return true
			}

			isResourceTest := terraformtype.IsHelperResourceFunc(callExpr.Fun, pass.TypesInfo, terraformtype.FuncNameTest)
			isResourceParallelTest := terraformtype.IsHelperResourceFunc(callExpr.Fun, pass.TypesInfo, terraformtype.FuncNameParallelTest)

			if !isResourceTest && !isResourceParallelTest {
				return true
			}

			pass.Reportf(testFunc.Pos(), "%s: acceptance test function name should begin with TestAcc", analyzerName)
			return true
		})

	}

	return nil, nil
}
