package acctestfunc

import (
	"go/ast"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "acctestfunc",
	Doc:  "find function names starting with TestAcc for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*ast.FuncDecl{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}
	var result []*ast.FuncDecl

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.FuncDecl)

		if !strings.HasPrefix(x.Name.Name, "TestAcc") {
			return
		}

		result = append(result, x)
	})

	return result, nil
}
