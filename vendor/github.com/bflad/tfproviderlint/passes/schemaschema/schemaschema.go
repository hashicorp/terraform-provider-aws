package schemaschema

import (
	"go/ast"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/schemamap"
)

var Analyzer = &analysis.Analyzer{
	Name: "schemaschema",
	Doc:  "find github.com/hashicorp/terraform-plugin-sdk/helper/schema.Schema literals for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		schemamap.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*terraformtype.HelperSchemaSchemaInfo{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	schemamaps := pass.ResultOf[schemamap.Analyzer].([]*ast.CompositeLit)
	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}
	var result []*terraformtype.HelperSchemaSchemaInfo

	for _, smap := range schemamaps {
		for _, schema := range schemamap.GetSchemaAttributes(smap) {
			result = append(result, terraformtype.NewHelperSchemaSchemaInfo(schema, pass.TypesInfo))
		}
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.CompositeLit)

		if !isSchemaSchema(pass, x) {
			return
		}

		result = append(result, terraformtype.NewHelperSchemaSchemaInfo(x, pass.TypesInfo))
	})

	return result, nil
}

func isSchemaSchema(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	switch v := cl.Type.(type) {
	default:
		return false
	case *ast.SelectorExpr:
		return terraformtype.IsHelperSchemaTypeSchema(pass.TypesInfo.TypeOf(v))
	}
}
