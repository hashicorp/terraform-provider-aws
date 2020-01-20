// Package S008 defines an Analyzer that checks for
// Schema of TypeList or TypeSet with Default configured
package S008

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema of TypeList or TypeSet with Default configured

The S008 analyzer reports cases of TypeList or TypeSet schemas with Default configured,
which will fail schema validation.`

const analyzerName = "S008"

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

		if !schema.DeclaresField(terraformtype.SchemaFieldDefault) {
			continue
		}

		if !schema.IsOneOfTypes(terraformtype.SchemaValueTypeList, terraformtype.SchemaValueTypeSet) {
			continue
		}

		switch t := schema.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(schema.AstCompositeLit.Lbrace, "%s: schema of TypeList or TypeSet should not include Default", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema of TypeList or TypeSet should not include Default", analyzerName)
		}
	}

	return nil, nil
}
