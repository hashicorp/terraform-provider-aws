// Package S021 defines an Analyzer that checks for
// Schema that should omit ComputedWhen
package S021

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema that should omit ComputedWhen

The S021 analyzer reports cases of schema that declare ComputedWhen that should
be removed.`

const analyzerName = "S021"

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

		field := terraformtype.SchemaFieldComputedWhen

		if schema.DeclaresField(field) {
			pass.Reportf(schema.Fields[field].Value.Pos(), "%s: schema should omit ComputedWhen", analyzerName)
		}
	}

	return nil, nil
}
