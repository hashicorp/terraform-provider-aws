package schemamap

import (
	"go/ast"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
)

var Analyzer = &analysis.Analyzer{
	Name: "schema",
	Doc:  "find map[string]*schema.Schema literals for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*ast.CompositeLit{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}
	var result []*ast.CompositeLit

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.CompositeLit)

		if !isSchemaMap(pass, x) {
			return
		}

		result = append(result, x)
	})

	return result, nil
}

func isSchemaMap(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	switch v := cl.Type.(type) {
	default:
		return false
	case *ast.MapType:
		switch k := v.Key.(type) {
		default:
			return false
		case *ast.Ident:
			if k.Name != "string" {
				return false
			}
		}

		switch mv := v.Value.(type) {
		default:
			return false
		case *ast.StarExpr:
			return terraformtype.IsTypeHelperSchema(pass.TypesInfo.TypeOf(mv.X))
		}
	}
	return true
}

// GetSchemaAttributes returns all attributes held in a map[string]*schema.Schema
func GetSchemaAttributes(schemamap *ast.CompositeLit) []*ast.CompositeLit {
	var result []*ast.CompositeLit

	for _, elt := range schemamap.Elts {
		switch v := elt.(type) {
		default:
			continue
		case *ast.KeyValueExpr:
			switch v := v.Value.(type) {
			default:
				continue
			case *ast.CompositeLit:
				result = append(result, v)
			}
		}
	}
	return result
}

// GetSchemaAttributeNames returns all attribute names held in a map[string]*schema.Schema
func GetSchemaAttributeNames(schemamap *ast.CompositeLit) []ast.Expr {
	var result []ast.Expr

	for _, elt := range schemamap.Elts {
		switch v := elt.(type) {
		default:
			continue
		case *ast.KeyValueExpr:
			result = append(result, v.Key)
		}
	}
	return result
}
