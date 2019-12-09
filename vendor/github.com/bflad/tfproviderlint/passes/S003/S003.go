// Package S003 defines an Analyzer that checks for
// Schema with both Required and Computed enabled
package S003

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema with both Required and Computed enabled

The S003 analyzer reports cases of schemas which enables both Required
and Computed, which will fail provider schema validation.`

const analyzerName = "S003"

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

		var computedEnabled, requiredEnabled bool

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				if name != "Computed" && name != "Required" {
					continue
				}

				switch v := v.Value.(type) {
				default:
					continue
				case *ast.Ident:
					value := v.Name

					if value != "true" {
						continue
					}

					if name == "Computed" {
						computedEnabled = true
						continue
					}

					requiredEnabled = true
				}
			}
		}

		if computedEnabled && requiredEnabled {
			switch t := schema.Type.(type) {
			default:
				pass.Reportf(schema.Lbrace, "%s: schema should not enable Required and Computed", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema should not enable Required and Computed", analyzerName)
			}
		}
	}

	return nil, nil
}
