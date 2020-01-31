// Package S023 defines an Analyzer that checks for
// Schema that should omit Elem with incompatible Type
package S023

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema that should omit Elem with incompatible Type

The S023 analyzer reports cases of schema that declare Elem that should
be removed with incompatible Type.`

const analyzerName = "S023"

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

		if schema.IsOneOfTypes(terraformtype.SchemaValueTypeList, terraformtype.SchemaValueTypeMap, terraformtype.SchemaValueTypeSet) {
			continue
		}

		if !schema.DeclaresField(terraformtype.SchemaFieldElem) {
			continue
		}

		pass.Reportf(schema.Fields[terraformtype.SchemaFieldElem].Value.Pos(), "%s: schema should not include Elem with incompatible Type", analyzerName)
	}

	return nil, nil
}
