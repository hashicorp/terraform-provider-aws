// Package S019 defines an Analyzer that checks for
// Schema that should omit Computed, Optional, or Required
// set to false
package S019

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema that should omit Computed, Optional, or Required set to false

The S019 analyzer reports cases of schema that use Computed: false, Optional: false, or
Required: false that should be removed.`

const analyzerName = "S019"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		schemaschema.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemas := pass.ResultOf[schemaschema.Analyzer].([]*ast.CompositeLit)
	for _, schema := range schemas {
		if ignorer.ShouldIgnore(analyzerName, schema) {
			continue
		}

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				if name != "Computed" && name != "Optional" && name != "Required" {
					continue
				}

				switch v := v.Value.(type) {
				default:
					continue
				case *ast.Ident:
					if v.Name == "false" {
						pass.Reportf(v.Pos(), "%s: schema should omit Computed, Optional, or Required set to false", analyzerName)
					}
				}
			}
		}
	}

	return nil, nil
}
