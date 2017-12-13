package hil_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hil"
	"github.com/hashicorp/hil/ast"
)

func Example_functions() {
	input := "${lower(var.test)} - ${6 + 2}"

	tree, err := hil.Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	lowerCase := ast.Function{
		ArgTypes:   []ast.Type{ast.TypeString},
		ReturnType: ast.TypeString,
		Variadic:   false,
		Callback: func(inputs []interface{}) (interface{}, error) {
			input := inputs[0].(string)
			return strings.ToLower(input), nil
		},
	}

	config := &hil.EvalConfig{
		GlobalScope: &ast.BasicScope{
			VarMap: map[string]ast.Variable{
				"var.test": ast.Variable{
					Type:  ast.TypeString,
					Value: "TEST STRING",
				},
			},
			FuncMap: map[string]ast.Function{
				"lower": lowerCase,
			},
		},
	}

	result, err := hil.Eval(tree, config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Type: %s\n", result.Type)
	fmt.Printf("Value: %s\n", result.Value)
	// Output:
	// Type: TypeString
	// Value: test string - 8
}
