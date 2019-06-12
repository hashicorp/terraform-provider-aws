// Package S001 defines an Analyzer that checks for
// Schema of TypeList or TypeSet missing Elem
package S001

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema of TypeList or TypeSet missing Elem

The S001 analyzer reports cases of TypeList or TypeSet schemas missing Elem,
which will fail schema validation.`

const analyzerName = "S001"

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

		var elemFound, typeListOrSet bool

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				if name == "Elem" {
					elemFound = true
					continue
				}

				if name != "Type" {
					continue
				}

				switch v := v.Value.(type) {
				default:
					continue
				case *ast.SelectorExpr:
					// Use AST over TypesInfo here as schema uses ValueType
					if v.Sel.Name != "TypeList" && v.Sel.Name != "TypeSet" {
						continue
					}

					switch t := pass.TypesInfo.TypeOf(v).(type) {
					default:
						continue
					case *types.Named:
						// HasSuffix here due to vendoring
						if !strings.HasSuffix(t.Obj().Pkg().Path(), "github.com/hashicorp/terraform/helper/schema") {
							continue
						}

						typeListOrSet = true
					}
				}
			}
		}

		if typeListOrSet && !elemFound {
			switch t := schema.Type.(type) {
			default:
				pass.Reportf(schema.Lbrace, "%s: schema of TypeList or TypeSet should include Elem", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema of TypeList or TypeSet should include Elem", analyzerName)
			}
		}
	}

	return nil, nil
}
