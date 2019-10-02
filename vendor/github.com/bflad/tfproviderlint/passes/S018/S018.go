// Package S018 defines an Analyzer that checks for
// Schema that should prefer TypeList with MaxItems 1
package S018

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema that should prefer TypeList with MaxItems 1

The S018 analyzer reports cases of schema including MaxItems 1 and TypeSet
that should be simplified.`

const analyzerName = "S018"

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

		var maxItemsOne, typeSetFound bool

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				if name != "MaxItems" && name != "Type" {
					continue
				}

				switch v := v.Value.(type) {
				default:
					continue
				case *ast.BasicLit:
					if name != "MaxItems" {
						continue
					}

					value := strings.Trim(v.Value, `"`)

					if value != "1" {
						continue
					}

					maxItemsOne = true
				case *ast.SelectorExpr:
					if name != "Type" {
						continue
					}

					// Use AST over TypesInfo here as schema uses ValueType
					if v.Sel.Name != "TypeSet" {
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

						typeSetFound = true
					}
				}
			}
		}

		if maxItemsOne && typeSetFound {
			switch t := schema.Type.(type) {
			default:
				pass.Reportf(schema.Lbrace, "%s: schema should use TypeList with MaxItems 1", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema should use TypeList with MaxItems 1", analyzerName)
			}
		}
	}

	return nil, nil
}
