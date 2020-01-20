// Package R006 defines an Analyzer that checks for
// RetryFunc that omit retryable errors
package R006

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/retryfunc"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for RetryFunc that omit retryable errors

The R006 analyzer reports when RetryFunc declarations are missing
retryable errors and should not be used as RetryFunc.`

const analyzerName = "R006"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		retryfunc.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	retryFuncs := pass.ResultOf[retryfunc.Analyzer].([]*terraformtype.HelperResourceRetryFuncInfo)

	for _, retryFunc := range retryFuncs {
		if ignorer.ShouldIgnore(analyzerName, retryFunc.Node) {
			continue
		}

		var retryableErrorFound bool

		ast.Inspect(retryFunc.Body, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)

			if !ok {
				return true
			}

			if terraformtype.IsHelperResourceFunc(callExpr.Fun, pass.TypesInfo, terraformtype.FuncNameRetryableError) {
				retryableErrorFound = true
				return false
			}

			return true
		})

		if !retryableErrorFound {
			pass.Reportf(retryFunc.Pos, "%s: RetryFunc should include RetryableError() handling or be removed", analyzerName)
		}
	}

	return nil, nil
}
