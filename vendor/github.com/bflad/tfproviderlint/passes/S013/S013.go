// Package S013 defines an Analyzer that checks for
// Schema that one of Computed, Optional, or Required
// is not configured
package S013

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemamap"
)

const Doc = `check for Schema that are missing required fields

The S013 analyzer reports cases of schemas which one of Computed,
Optional, or Required is not configured, which will fail provider
schema validation.`

const analyzerName = "S013"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		schemamap.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemamaps := pass.ResultOf[schemamap.Analyzer].([]*ast.CompositeLit)

	for _, smap := range schemamaps {
		for _, schemaCompositeLit := range schemamap.GetSchemaAttributes(smap) {
			schema := terraformtype.NewHelperSchemaSchemaInfo(schemaCompositeLit, pass.TypesInfo)

			if ignorer.ShouldIgnore(analyzerName, schema.AstCompositeLit) {
				continue
			}

			if schema.Schema.Computed || schema.Schema.Optional || schema.Schema.Required {
				continue
			}

			switch t := schema.AstCompositeLit.Type.(type) {
			default:
				pass.Reportf(schema.AstCompositeLit.Lbrace, "%s: schema should configure one of Computed, Optional, or Required", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema should configure one of Computed, Optional, or Required", analyzerName)
			}
		}
	}

	return nil, nil
}
