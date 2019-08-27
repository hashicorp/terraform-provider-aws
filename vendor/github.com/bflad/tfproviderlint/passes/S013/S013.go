// Package S013 defines an Analyzer that checks for
// Schema that one of Computed, Optional, or Required
// is not configured
package S013

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemamap"
)

const Doc = `check for Schema that are missing required fields

The S013 analyzer reports cases of schemas which one of Computed,
Optional, or Required is not configured, which will fail provider
schema validation.`

const analyzerName = "S013"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		schemamap.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemamaps := pass.ResultOf[schemamap.Analyzer].([]*ast.CompositeLit)

	for _, smap := range schemamaps {
		for _, schema := range schemamap.GetSchemaAttributes(smap) {
			if ignorer.ShouldIgnore(analyzerName, schema) {
				continue
			}

			var computedOrOptionalOrRequiredEnabled bool

			for _, elt := range schema.Elts {
				switch v := elt.(type) {
				default:
					continue
				case *ast.KeyValueExpr:
					name := v.Key.(*ast.Ident).Name

					switch v := v.Value.(type) {
					case *ast.Ident:
						value := v.Name

						switch name {
						case "Computed", "Optional", "Required":
							if value == "true" {
								computedOrOptionalOrRequiredEnabled = true
								break
							}
						}
					}
				}
			}

			if !computedOrOptionalOrRequiredEnabled {
				switch t := schema.Type.(type) {
				default:
					pass.Reportf(schema.Lbrace, "%s: schema should configure one of Computed, Optional, or Required", analyzerName)
				case *ast.SelectorExpr:
					pass.Reportf(t.Sel.Pos(), "%s: schema should configure one of Computed, Optional, or Required", analyzerName)
				}
			}
		}
	}

	return nil, nil
}
