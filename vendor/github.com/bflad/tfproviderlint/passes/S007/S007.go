// Package S007 defines an Analyzer that checks for
// Schema with Required enabled and ConflictsWith configured
package S007

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema with Required enabled and ConflictsWith configured

The S007 analyzer reports cases of schemas which enables Required
and configures ConflictsWith, which will fail provider schema validation.`

const analyzerName = "S007"

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

		if !schema.Schema.Required || schema.Schema.ConflictsWith == nil {
			continue
		}

		switch t := schema.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(schema.AstCompositeLit.Lbrace, "%s: schema should not enable Required and configure ConflictsWith", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema should not enable Required and configure ConflictsWith", analyzerName)
		}
	}

	return nil, nil
}
