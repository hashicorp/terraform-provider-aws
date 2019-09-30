// Package S009 defines an Analyzer that checks for
// Schema of TypeList or TypeSet with ValidateFunc configured
package S009

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema of TypeList or TypeSet with ValidateFunc configured

The S009 analyzer reports cases of TypeList or TypeSet schemas with ValidateFunc configured,
which will fail schema validation.`

const analyzerName = "S009"

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

		var validateFuncFound, typeListOrSet bool

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				if name == "ValidateFunc" {
					validateFuncFound = true
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

		if typeListOrSet && validateFuncFound {
			switch t := schema.Type.(type) {
			default:
				pass.Reportf(schema.Lbrace, "%s: schema of TypeList or TypeSet should not include top level ValidateFunc", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema of TypeList or TypeSet should not include top level ValidateFunc", analyzerName)
			}
		}
	}

	return nil, nil
}
