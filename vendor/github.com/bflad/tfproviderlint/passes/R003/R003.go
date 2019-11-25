// Package R003 defines an Analyzer that checks for
// Resource having Exists functions
package R003

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaresource"
)

const Doc = `check for Resource having Exists functions

The R003 analyzer reports likely extraneous uses of Exists
functions for a resource. Exists logic can be handled inside the Read function
to prevent logic duplication.`

const analyzerName = "R003"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		schemaresource.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	resources := pass.ResultOf[schemaresource.Analyzer].([]*ast.CompositeLit)
	for _, resource := range resources {
		if ignorer.ShouldIgnore(analyzerName, resource) {
			continue
		}

		for _, elt := range resource.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				if v.Key.(*ast.Ident).Name == "Exists" {
					pass.Reportf(v.Key.Pos(), "%s: resource should not include Exists function", analyzerName)
					break
				}
			}
		}
	}

	return nil, nil
}
