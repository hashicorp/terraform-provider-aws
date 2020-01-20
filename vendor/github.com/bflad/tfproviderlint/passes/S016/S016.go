// Package S016 defines an Analyzer that checks for
// Schema including Set without TypeSet
package S016

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema including Set without TypeSet

The S016 analyzer reports cases of schema including Set without TypeSet,
which will fail schema validation.`

const analyzerName = "S016"

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

		if !schema.DeclaresField(terraformtype.SchemaFieldSet) {
			continue
		}

		if schema.IsType(terraformtype.SchemaValueTypeSet) {
			continue
		}

		switch t := schema.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(schema.AstCompositeLit.Lbrace, "%s: schema Set should only be included for TypeSet", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema Set should only be included for TypeSet", analyzerName)
		}
	}

	return nil, nil
}
