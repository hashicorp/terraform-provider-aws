// Package S004 defines an Analyzer that checks for
// Schema with Required enabled and Default configured
package S004

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema with Required enabled and Default configured

The S004 analyzer reports cases of schemas which enables Required
and configures Default, which will fail provider schema validation.`

const analyzerName = "S004"

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

		var defaultConfigured, requiredEnabled bool

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				if name != "Default" && name != "Required" {
					continue
				}

				switch v := v.Value.(type) {
				default:
					if name == "Default" {
						defaultConfigured = true
					}

					continue
				case *ast.Ident:
					value := v.Name

					if name == "Default" && value != "nil" {
						defaultConfigured = true
						continue
					}

					if value != "true" {
						continue
					}

					requiredEnabled = true
				}
			}
		}

		if defaultConfigured && requiredEnabled {
			switch t := schema.Type.(type) {
			default:
				pass.Reportf(schema.Lbrace, "%s: schema should not enable Required and configure Default", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema should not enable Required and configure Default", analyzerName)
			}
		}
	}

	return nil, nil
}
