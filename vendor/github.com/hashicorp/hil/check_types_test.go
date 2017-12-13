package hil

import (
	"testing"

	"github.com/hashicorp/hil/ast"
)

func TestTypeCheck(t *testing.T) {
	cases := []struct {
		Input string
		Scope ast.Scope
		Error bool
	}{
		{
			"foo",
			&ast.BasicScope{},
			false,
		},

		{
			"foo ${bar}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: "baz",
						Type:  ast.TypeString,
					},
				},
			},
			false,
		},

		{
			"foo ${rand()}",
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"rand": ast.Function{
						ReturnType: ast.TypeString,
						Callback: func([]interface{}) (interface{}, error) {
							return "42", nil
						},
					},
				},
			},
			false,
		},

		{
			`foo ${rand("42")}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"rand": ast.Function{
						ArgTypes:   []ast.Type{ast.TypeString},
						ReturnType: ast.TypeString,
						Callback: func([]interface{}) (interface{}, error) {
							return "42", nil
						},
					},
				},
			},
			false,
		},

		{
			`foo ${rand(42)}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"rand": ast.Function{
						ArgTypes:   []ast.Type{ast.TypeString},
						ReturnType: ast.TypeString,
						Callback: func([]interface{}) (interface{}, error) {
							return "42", nil
						},
					},
				},
			},
			true,
		},

		{
			`foo ${rand()}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"rand": ast.Function{
						ArgTypes:     nil,
						ReturnType:   ast.TypeString,
						Variadic:     true,
						VariadicType: ast.TypeString,
						Callback: func([]interface{}) (interface{}, error) {
							return "42", nil
						},
					},
				},
			},
			false,
		},

		{
			`foo ${rand("42")}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"rand": ast.Function{
						ArgTypes:     nil,
						ReturnType:   ast.TypeString,
						Variadic:     true,
						VariadicType: ast.TypeString,
						Callback: func([]interface{}) (interface{}, error) {
							return "42", nil
						},
					},
				},
			},
			false,
		},

		{
			`foo ${rand("42", 42)}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"rand": ast.Function{
						ArgTypes:     nil,
						ReturnType:   ast.TypeString,
						Variadic:     true,
						VariadicType: ast.TypeString,
						Callback: func([]interface{}) (interface{}, error) {
							return "42", nil
						},
					},
				},
			},
			true,
		},

		{
			"${foo[0]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeString,
								Value: "Hello",
							},
							ast.Variable{
								Type:  ast.TypeString,
								Value: "World",
							},
						},
					},
				},
			},
			false,
		},

		{
			"${foo[0]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 3,
							},
							ast.Variable{
								Type:  ast.TypeString,
								Value: "World",
							},
						},
					},
				},
			},
			true,
		},

		{
			"${foo[0]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type:  ast.TypeString,
						Value: "Hello World",
					},
				},
			},
			true,
		},

		{
			"foo ${bar}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: 42,
						Type:  ast.TypeInt,
					},
				},
			},
			true,
		},

		{
			"foo ${rand()}",
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"rand": ast.Function{
						ReturnType: ast.TypeInt,
						Callback: func([]interface{}) (interface{}, error) {
							return 42, nil
						},
					},
				},
			},
			true,
		},

		{
			`foo ${true ? "foo" : "bar"}`,
			&ast.BasicScope{},
			false,
		},

		{
			// can't use different types for true and false expressions
			`foo ${true ? 1 : "baz"}`,
			&ast.BasicScope{},
			true,
		},

		{
			// condition must be boolean
			`foo ${"foo" ? 1 : 5}`,
			&ast.BasicScope{},
			true,
		},

		{
			// conditional with unknown value is permitted
			`foo ${true ? known : unknown}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"known": ast.Variable{
						Type:  ast.TypeString,
						Value: "bar",
					},
					"unknown": ast.Variable{
						Type:  ast.TypeUnknown,
						Value: UnknownValue,
					},
				},
			},
			false,
		},

		{
			// conditional with unknown value the other way permitted too
			`foo ${true ? unknown : known}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"known": ast.Variable{
						Type:  ast.TypeString,
						Value: "bar",
					},
					"unknown": ast.Variable{
						Type:  ast.TypeUnknown,
						Value: UnknownValue,
					},
				},
			},
			false,
		},

		{
			// conditional with two unknowns is allowed
			`foo ${true ? unknown : unknown}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"unknown": ast.Variable{
						Type:  ast.TypeUnknown,
						Value: UnknownValue,
					},
				},
			},
			false,
		},

		{
			// conditional with unknown condition is allowed
			`foo ${unknown ? 1 : 2}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"unknown": ast.Variable{
						Type:  ast.TypeUnknown,
						Value: UnknownValue,
					},
				},
			},
			false,
		},

		{
			// currently lists are not allowed at all
			`foo ${true ? arr1 : arr2}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"arr1": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 3,
							},
						},
					},
					"arr2": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 4,
							},
						},
					},
				},
			},
			true,
		},

		{
			// mismatching element types are invalid
			`foo ${true ? arr1 : arr2}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"arr1": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 3,
							},
						},
					},
					"arr2": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeString,
								Value: "foo",
							},
						},
					},
				},
			},
			true,
		},
	}

	for _, tc := range cases {
		node, err := Parse(tc.Input)
		if err != nil {
			t.Fatalf("Error: %s\n\nInput: %s", err, tc.Input)
		}

		visitor := &TypeCheck{Scope: tc.Scope}
		err = visitor.Visit(node)
		if err != nil != tc.Error {
			t.Fatalf("Error: %s\n\nInput: %s", err, tc.Input)
		}
	}
}

func TestTypeCheck_implicit(t *testing.T) {
	implicitMap := map[ast.Type]map[ast.Type]string{
		ast.TypeInt: {
			ast.TypeString: "intToString",
		},
	}

	cases := []struct {
		Input string
		Scope *ast.BasicScope
		Error bool
	}{
		{
			"foo ${bar}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: 42,
						Type:  ast.TypeInt,
					},
				},
			},
			false,
		},

		{
			"foo ${foo(42)}",
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"foo": ast.Function{
						ArgTypes:   []ast.Type{ast.TypeString},
						ReturnType: ast.TypeString,
					},
				},
			},
			false,
		},

		{
			`foo ${foo("42", 42)}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"foo": ast.Function{
						ArgTypes:     []ast.Type{ast.TypeString},
						Variadic:     true,
						VariadicType: ast.TypeString,
						ReturnType:   ast.TypeString,
					},
				},
			},
			false,
		},

		{
			"${foo[1]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 42,
							},
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 23,
							},
						},
					},
				},
			},
			false,
		},

		{
			"${foo[1]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 42,
							},
							ast.Variable{
								Type: ast.TypeUnknown,
							},
						},
					},
				},
			},
			false,
		},

		{
			`${foo[bar[var.keyint]]}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeMap,
						Value: map[string]ast.Variable{
							"foo": ast.Variable{
								Type:  ast.TypeString,
								Value: "hello",
							},
							"bar": ast.Variable{
								Type:  ast.TypeString,
								Value: "world",
							},
						},
					},
					"bar": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeString,
								Value: "i dont exist",
							},
							ast.Variable{
								Type:  ast.TypeString,
								Value: "bar",
							},
						},
					},
					"var.keyint": ast.Variable{
						Type:  ast.TypeInt,
						Value: 1,
					},
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Input, func(t *testing.T) {
			node, err := Parse(tc.Input)
			if err != nil {
				t.Fatalf("Error: %s\n\nInput: %s", err, tc.Input)
			}

			// Modify the scope to add our conversion functions.
			if tc.Scope.FuncMap == nil {
				tc.Scope.FuncMap = make(map[string]ast.Function)
			}
			tc.Scope.FuncMap["intToString"] = ast.Function{
				ArgTypes:   []ast.Type{ast.TypeInt},
				ReturnType: ast.TypeString,
			}

			// Do the first pass...
			visitor := &TypeCheck{Scope: tc.Scope, Implicit: implicitMap}
			err = visitor.Visit(node)
			if err != nil != tc.Error {
				t.Fatalf("Error: %s\n\nInput: %s", err, tc.Input)
			}
			if err != nil {
				return
			}

			// If we didn't error, then the next type check should not fail
			// WITHOUT implicits.
			visitor = &TypeCheck{Scope: tc.Scope}
			err = visitor.Visit(node)
			if err != nil {
				t.Fatalf("Error: %s\n\nInput: %s", err, tc.Input)
			}
		})
	}
}
