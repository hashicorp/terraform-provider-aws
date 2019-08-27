// Package AT001 defines an Analyzer that checks for
// TestCase missing CheckDestroy
package AT001

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/acctestcase"
	"github.com/bflad/tfproviderlint/passes/commentignore"
)

const Doc = `check for TestCase missing CheckDestroy

The AT001 analyzer reports likely incorrect uses of TestCase
which do not define a CheckDestroy function. CheckDestroy is used to verify
that test infrastructure has been removed at the end of an acceptance test.
Ignores file names beginning with data_source_.

More information can be found at:
https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy`

const analyzerName = "AT001"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		acctestcase.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	testCases := pass.ResultOf[acctestcase.Analyzer].([]*ast.CompositeLit)
	for _, testCase := range testCases {
		fileName := filepath.Base(pass.Fset.File(testCase.Pos()).Name())

		if strings.HasPrefix(fileName, "data_source_") {
			continue
		}

		if ignorer.ShouldIgnore(analyzerName, testCase) {
			continue
		}

		var found bool

		for _, elt := range testCase.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				if v.Key.(*ast.Ident).Name == "CheckDestroy" {
					found = true
					break
				}
			}
		}

		if !found {
			pass.Reportf(testCase.Type.(*ast.SelectorExpr).Sel.Pos(), "%s: missing CheckDestroy", analyzerName)
		}
	}

	return nil, nil
}
