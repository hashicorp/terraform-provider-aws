package ast

import (
	"testing"
)

func TestOutput_type(t *testing.T) {
	testCases := []struct {
		Name        string
		Output      *Output
		Scope       Scope
		ReturnType  Type
		ShouldError bool
	}{
		{
			Name:       "No expressions, for backward compatibility",
			Output:     &Output{},
			Scope:      nil,
			ReturnType: TypeString,
		},
		{
			Name: "Single string expression",
			Output: &Output{
				Exprs: []Node{
					&LiteralNode{
						Value: "Whatever",
						Typex: TypeString,
					},
				},
			},
			Scope:      nil,
			ReturnType: TypeString,
		},
		{
			Name: "Single list expression of strings",
			Output: &Output{
				Exprs: []Node{
					&VariableAccess{
						Name: "testvar",
					},
				},
			},
			Scope: &BasicScope{
				VarMap: map[string]Variable{
					"testvar": Variable{
						Type: TypeList,
						Value: []Variable{
							Variable{
								Type:  TypeString,
								Value: "Hello",
							},
							Variable{
								Type:  TypeString,
								Value: "World",
							},
						},
					},
				},
			},
			ReturnType: TypeList,
		},
		{
			Name: "Single map expression",
			Output: &Output{
				Exprs: []Node{
					&VariableAccess{
						Name: "testvar",
					},
				},
			},
			Scope: &BasicScope{
				VarMap: map[string]Variable{
					"testvar": Variable{
						Type: TypeMap,
						Value: map[string]Variable{
							"key1": Variable{
								Type:  TypeString,
								Value: "Hello",
							},
							"key2": Variable{
								Type:  TypeString,
								Value: "World",
							},
						},
					},
				},
			},
			ReturnType: TypeMap,
		},
		{
			Name: "Multiple map expressions",
			Output: &Output{
				Exprs: []Node{
					&VariableAccess{
						Name: "testvar",
					},
					&VariableAccess{
						Name: "testvar",
					},
				},
			},
			Scope: &BasicScope{
				VarMap: map[string]Variable{
					"testvar": Variable{
						Type: TypeMap,
						Value: map[string]Variable{
							"key1": Variable{
								Type:  TypeString,
								Value: "Hello",
							},
							"key2": Variable{
								Type:  TypeString,
								Value: "World",
							},
						},
					},
				},
			},
			ShouldError: true,
			ReturnType:  TypeInvalid,
		},
		{
			Name: "Multiple list expressions",
			Output: &Output{
				Exprs: []Node{
					&VariableAccess{
						Name: "testvar",
					},
					&VariableAccess{
						Name: "testvar",
					},
				},
			},
			Scope: &BasicScope{
				VarMap: map[string]Variable{
					"testvar": Variable{
						Type: TypeList,
						Value: []Variable{
							Variable{
								Type:  TypeString,
								Value: "Hello",
							},
							Variable{
								Type:  TypeString,
								Value: "World",
							},
						},
					},
				},
			},
			ShouldError: true,
			ReturnType:  TypeInvalid,
		},
		{
			Name: "Multiple string expressions",
			Output: &Output{
				Exprs: []Node{
					&VariableAccess{
						Name: "testvar",
					},
					&VariableAccess{
						Name: "testvar",
					},
				},
			},
			Scope: &BasicScope{
				VarMap: map[string]Variable{
					"testvar": Variable{
						Type:  TypeString,
						Value: "Hello",
					},
				},
			},
			ReturnType: TypeString,
		},
		{
			Name: "Multiple string expressions with coercion",
			Output: &Output{
				Exprs: []Node{
					&VariableAccess{
						Name: "testvar",
					},
					&VariableAccess{
						Name: "testint",
					},
				},
			},
			Scope: &BasicScope{
				VarMap: map[string]Variable{
					"testvar": Variable{
						Type:  TypeString,
						Value: "Hello",
					},
					"testint": Variable{
						Type:  TypeInt,
						Value: 2,
					},
				},
			},
			ReturnType: TypeString,
		},
	}

	for _, v := range testCases {
		actual, err := v.Output.Type(v.Scope)
		if err != nil && !v.ShouldError {
			t.Fatalf("case: %s\nerr: %s", v.Name, err)
		}
		if actual != v.ReturnType {
			t.Fatalf("case: %s\n     bad: %s\nexpected: %s\n", v.Name, actual, v.ReturnType)
		}
	}
}
