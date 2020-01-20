// Package AT004 defines an Analyzer that checks for
// TestStep Config containing provider configuration
package AT004

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for TestStep Config containing provider configuration

The AT004 analyzer reports likely incorrect uses of TestStep
Config which define a provider configuration. Provider configurations should
be handled outside individual test configurations (e.g. environment variables).`

const analyzerName = "AT004"

var Analyzer = &analysis.Analyzer{
	Name:     analyzerName,
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BasicLit)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.BasicLit)

		if x.Kind != token.STRING {
			return
		}

		if !strings.Contains(x.Value, `provider "`) {
			return
		}

		pass.Reportf(x.ValuePos, "%s: provider declaration should be omitted", analyzerName)
	})
	return nil, nil
}
