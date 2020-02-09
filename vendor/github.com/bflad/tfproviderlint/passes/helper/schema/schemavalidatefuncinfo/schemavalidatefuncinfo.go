package schemavalidatefuncinfo

import (
	"go/ast"
	"reflect"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "schemavalidatefuncinfo",
	Doc:  "find github.com/hashicorp/terraform-plugin-sdk/helper/schema SchemaValidateFunc declarations for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*schema.SchemaValidateFuncInfo{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}
	var result []*schema.SchemaValidateFuncInfo

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl, funcDeclOk := n.(*ast.FuncDecl)
		funcLit, funcLitOk := n.(*ast.FuncLit)

		var funcType *ast.FuncType

		if funcDeclOk && funcDecl != nil {
			funcType = funcDecl.Type
		} else if funcLitOk && funcLit != nil {
			funcType = funcLit.Type
		} else {
			return
		}

		params := funcType.Params

		if params == nil || len(params.List) != 2 {
			return
		}

		if !astutils.IsFunctionParameterTypeInterface(params.List[0].Type) {
			return
		}

		if !astutils.IsFunctionParameterTypeString(params.List[1].Type) {
			return
		}

		results := funcType.Results

		if results == nil || len(results.List) != 2 {
			return
		}

		if !astutils.IsFunctionParameterTypeArrayString(results.List[0].Type) {
			return
		}

		if !astutils.IsFunctionParameterTypeArrayError(results.List[1].Type) {
			return
		}

		result = append(result, schema.NewSchemaValidateFuncInfo(funcDecl, funcLit, pass.TypesInfo))
	})

	return result, nil
}
