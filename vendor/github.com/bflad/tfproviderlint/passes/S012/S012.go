// Package S012 defines an Analyzer that checks for
// Schema that Type is configured
package S012

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema that Type is configured

The S012 analyzer reports cases of schemas which are missing Type,
which will fail provider schema validation.`

const analyzerName = "S012"

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

		if schema.DeclaresField(terraformtype.SchemaFieldType) {
			continue
		}

		switch t := schema.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(schema.AstCompositeLit.Lbrace, "%s: schema should configure Type", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema should configure Type", analyzerName)
		}
	}

	return nil, nil
}
