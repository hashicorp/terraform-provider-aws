// Package S017 defines an Analyzer that checks for
// Schema including MaxItems or MinItems without TypeList,
// TypeMap, or TypeSet
package S017

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema including MaxItems or MinItems without proper Type

The S017 analyzer reports cases of schema including MaxItems or MinItems without
TypeList, TypeMap, or TypeSet, which will fail schema validation.`

const analyzerName = "S017"

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

		var maxOrMinItemsFound, typeListOrMapOrSetFound bool

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				if name == "MaxItems" || name == "MinItems" {
					maxOrMinItemsFound = true
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
					if v.Sel.Name != "TypeList" && v.Sel.Name != "TypeMap" && v.Sel.Name != "TypeSet" {
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

						typeListOrMapOrSetFound = true
					}
				}
			}
		}

		if maxOrMinItemsFound && !typeListOrMapOrSetFound {
			switch t := schema.Type.(type) {
			default:
				pass.Reportf(schema.Lbrace, "%s: schema MaxItems or MinItems should only be included for TypeList, TypeMap, or TypeSet", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema MaxItems or MinItems should only be included for TypeList, TypeMap, or TypeSet", analyzerName)
			}
		}
	}

	return nil, nil
}
