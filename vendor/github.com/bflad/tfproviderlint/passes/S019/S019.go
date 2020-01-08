// Package S019 defines an Analyzer that checks for
// Schema that should omit Computed, Optional, or Required
// set to false
package S019

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema that should omit Computed, Optional, or Required set to false

The S019 analyzer reports cases of schema that use Computed: false, Optional: false, or
Required: false that should be removed.`

const analyzerName = "S019"

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

		for _, field := range []string{terraformtype.SchemaFieldComputed, terraformtype.SchemaFieldOptional, terraformtype.SchemaFieldRequired} {
			if schema.DeclaresBoolFieldWithZeroValue(field) {
				pass.Reportf(schema.Fields[field].Value.Pos(), "%s: schema should omit Computed, Optional, or Required set to false", analyzerName)
			}
		}
	}

	return nil, nil
}
