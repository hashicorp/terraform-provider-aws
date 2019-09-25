// Package S011 defines an Analyzer that checks for
// Schema with only Computed enabled and DiffSuppressFunc configured
package S011

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema with only Computed enabled and DiffSuppressFunc configured

The S011 analyzer reports cases of schemas which only enables Computed
and configures DiffSuppressFunc, which will fail provider schema validation.`

const analyzerName = "S011"

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

		var computedEnabled, optionalOrRequiredEnabled, diffSuppressFuncConfigured bool

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				switch v := v.Value.(type) {
				default:
					if name == "DiffSuppressFunc" {
						diffSuppressFuncConfigured = true
					}

					continue
				case *ast.Ident:
					value := v.Name

					switch name {
					case "Computed":
						if value == "true" {
							computedEnabled = true
							continue
						}
					case "Optional", "Required":
						if value == "true" {
							optionalOrRequiredEnabled = true
							break
						}
					case "DiffSuppressFunc":
						if value != "nil" {
							diffSuppressFuncConfigured = true
							continue
						}
					}
				}
			}
		}

		if computedEnabled && !optionalOrRequiredEnabled && diffSuppressFuncConfigured {
			switch t := schema.Type.(type) {
			default:
				pass.Reportf(schema.Lbrace, "%s: schema should not only enable Computed and configure DiffSuppressFunc", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema should not only enable Computed and configure DiffSuppressFunc", analyzerName)
			}
		}
	}

	return nil, nil
}
