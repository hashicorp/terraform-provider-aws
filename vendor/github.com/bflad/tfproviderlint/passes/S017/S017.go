// Package S017 defines an Analyzer that checks for
// Schema including MaxItems or MinItems without TypeList,
// TypeMap, or TypeSet
package S017

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
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
	schemas := pass.ResultOf[schemaschema.Analyzer].([]*terraformtype.HelperSchemaSchemaInfo)
	for _, schema := range schemas {
		if ignorer.ShouldIgnore(analyzerName, schema.AstCompositeLit) {
			continue
		}

		if !schema.DeclaresField(terraformtype.SchemaFieldMaxItems) && !schema.DeclaresField(terraformtype.SchemaFieldMinItems) {
			continue
		}

		if schema.IsOneOfTypes(terraformtype.SchemaValueTypeList, terraformtype.SchemaValueTypeMap, terraformtype.SchemaValueTypeSet) {
			continue
		}

		switch t := schema.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(schema.AstCompositeLit.Lbrace, "%s: schema MaxItems or MinItems should only be included for TypeList, TypeMap, or TypeSet", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema MaxItems or MinItems should only be included for TypeList, TypeMap, or TypeSet", analyzerName)
		}
	}

	return nil, nil
}
