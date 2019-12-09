// Package S010 defines an Analyzer that checks for
// Schema with only Computed enabled and ValidateFunc configured
package S010

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema with only Computed enabled and ValidateFunc configured

The S010 analyzer reports cases of schemas which only enables Computed
and configures ValidateFunc, which will fail provider schema validation.`

const analyzerName = "S010"

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

		var computedEnabled, optionalOrRequiredEnabled, validateFuncConfigured bool

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				switch v := v.Value.(type) {
				default:
					if name == "ValidateFunc" {
						validateFuncConfigured = true
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
					case "ValidateFunc":
						if value != "nil" {
							validateFuncConfigured = true
							continue
						}
					}
				}
			}
		}

		if computedEnabled && !optionalOrRequiredEnabled && validateFuncConfigured {
			switch t := schema.Type.(type) {
			default:
				pass.Reportf(schema.Lbrace, "%s: schema should not only enable Computed and configure ValidateFunc", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema should not only enable Computed and configure ValidateFunc", analyzerName)
			}
		}
	}

	return nil, nil
}
