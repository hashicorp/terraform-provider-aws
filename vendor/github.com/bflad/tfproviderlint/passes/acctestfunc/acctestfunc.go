package acctestfunc

import (
	"go/ast"
	"reflect"
	"strings"

	"github.com/bflad/tfproviderlint/passes/testfunc"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "acctestfunc",
	Doc:  "find function names starting with TestAcc for later passes",
	Requires: []*analysis.Analyzer{
		testfunc.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*ast.FuncDecl{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	testFuncs := pass.ResultOf[testfunc.Analyzer].([]*ast.FuncDecl)

	var result []*ast.FuncDecl

	for _, testFunc := range testFuncs {
		if strings.HasPrefix(testFunc.Name.Name, "TestAcc") {
			result = append(result, testFunc)
		}
	}

	return result, nil
}
