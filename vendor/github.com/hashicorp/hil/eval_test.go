package hil

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/hil/ast"
)

func TestEval(t *testing.T) {
	cases := []struct {
		Input      string
		Scope      *ast.BasicScope
		Error      bool
		Result     interface{}
		ResultType EvalType
	}{
		{
			Input:      "Hello World",
			Scope:      nil,
			Result:     "Hello World",
			ResultType: TypeString,
		},
		{
			Input:      `${"foo\\bar"}`,
			Scope:      nil,
			Result:     `foo\bar`,
			ResultType: TypeString,
		},
		{
			Input:      `${"foo\\\\bar"}`,
			Scope:      nil,
			Result:     `foo\\bar`,
			ResultType: TypeString,
		},
		{
			"${var.alist}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
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
			[]interface{}{"Hello", "World"},
			TypeList,
		},
		{
			"${var.alist[1]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
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
			"World",
			TypeString,
		},
		{
			`${var.alist["1"]}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
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
			"World",
			TypeString,
		},
		{
			"${var.alist[1]} ${var.alist[0]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
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
			"World Hello",
			TypeString,
		},
		{
			"${var.alist[2-1]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
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
			"World",
			TypeString,
		},
		{
			"${var.alist[1]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
						Type:  ast.TypeUnknown,
						Value: UnknownValue,
					},
				},
			},
			false,
			UnknownValue,
			TypeUnknown,
		},
		{
			"${var.alist[var.index]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeString,
								Value: "Hello",
							},
						},
					},
					"var.index": ast.Variable{
						Type:  ast.TypeUnknown,
						Value: UnknownValue,
					},
				},
			},
			false,
			UnknownValue,
			TypeUnknown,
		},
		{
			"${var.alist} ${var.alist}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
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
			true,
			nil,
			TypeInvalid,
		},
		{
			"${var.alist[1]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeUnknown,
								Value: UnknownValue,
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
			"World",
			TypeString,
		},
		{
			`${foo}`,
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
				},
			},
			false,
			map[string]interface{}{
				"foo": "hello",
				"bar": "world",
			},
			TypeMap,
		},
		{
			`${foo["bar"]}`,
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
				},
			},
			false,
			"world",
			TypeString,
		},
		{
			`${foo["foo"]}`,
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
								Type: ast.TypeUnknown,
							},
						},
					},
				},
			},
			false,
			"hello",
			TypeString,
		},
		{
			`${foo["bar"]}`,
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
								Type: ast.TypeUnknown,
							},
						},
					},
				},
			},
			false,
			UnknownValue,
			TypeUnknown,
		},
		{
			`${foo["foo"]} foo`,
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
								Type: ast.TypeUnknown,
							},
						},
					},
				},
			},
			false,
			"hello foo",
			TypeString,
		},
		{
			`${foo["bar"]} foo`,
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
								Type: ast.TypeUnknown,
							},
						},
					},
				},
			},
			false,
			UnknownValue,
			TypeUnknown,
		},

		{
			`${foo[3]}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeMap,
						Value: map[string]ast.Variable{
							"3": ast.Variable{
								Type:  ast.TypeString,
								Value: "world",
							},
						},
					},
				},
			},
			false,
			"world",
			TypeString,
		},
		{
			`${foo["bar"]} ${foo["foo"]}`,
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
				},
			},
			false,
			"world hello",
			TypeString,
		},
		{
			`${foo} ${foo}`,
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
				},
			},
			true,
			nil,
			TypeInvalid,
		},
		{
			`${foo} ${bar}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type:  ast.TypeString,
						Value: "Hello",
					},
					"bar": ast.Variable{
						Type:  ast.TypeString,
						Value: "World",
					},
				},
			},
			false,
			"Hello World",
			TypeString,
		},
		{
			`${foo} ${bar}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type:  ast.TypeString,
						Value: "Hello",
					},
					"bar": ast.Variable{
						Type:  ast.TypeInt,
						Value: 4,
					},
				},
			},
			false,
			"Hello 4",
			TypeString,
		},
		{
			`${foo} ${bar}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type:  ast.TypeString,
						Value: "Hello",
					},
					"bar": ast.Variable{
						Type: ast.TypeUnknown,
					},
				},
			},
			false,
			UnknownValue,
			TypeUnknown,
		},
		{
			`${foo}`,
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
				},
			},
			false,
			map[string]interface{}{
				"foo": "hello",
				"bar": "world",
			},
			TypeMap,
		},
		{
			"${var.alist}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
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
			[]interface{}{
				"Hello",
				"World",
			},
			TypeList,
		},
		{
			"${var.alist[0] + var.alist[1]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeUnknown,
								Value: UnknownValue,
							},
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 2,
							},
						},
					},
				},
			},
			false,
			UnknownValue,
			TypeUnknown,
		},
		{
			// Unknowns can short-circuit bits of our type checking
			// AST transform, such as the promotion of arithmetic to
			// functions. This test ensures that the evaluator and the
			// type checker co-operate to ensure that this doesn't cause
			// raw arithmetic nodes to be evaluated (which is not supported).
			"${var.alist[0 + var.unknown]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 2,
							},
						},
					},
					"var.unknown": ast.Variable{
						Type:  ast.TypeUnknown,
						Value: UnknownValue,
					},
				},
			},
			false,
			UnknownValue,
			TypeUnknown,
		},
		{
			"${join(var.alist)}",
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"join": ast.Function{
						ArgTypes:   []ast.Type{ast.TypeList},
						ReturnType: ast.TypeString,
						Callback: func(args []interface{}) (interface{}, error) {
							return nil, fmt.Errorf("should never actually be called")
						},
					},
				},
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeUnknown,
								Value: UnknownValue,
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
			UnknownValue,
			TypeUnknown,
		},
		{
			"${upper(var.alist[1])}",
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"upper": ast.Function{
						ArgTypes:   []ast.Type{ast.TypeString},
						ReturnType: ast.TypeString,
						Callback: func(args []interface{}) (interface{}, error) {
							return strings.ToUpper(args[0].(string)), nil
						},
					},
				},
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeUnknown,
								Value: UnknownValue,
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
			"WORLD",
			TypeString,
		},
		{
			`${foo[upper(bar)]}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"upper": ast.Function{
						ArgTypes:   []ast.Type{ast.TypeString},
						ReturnType: ast.TypeString,
						Callback: func(args []interface{}) (interface{}, error) {
							return strings.ToUpper(args[0].(string)), nil
						},
					},
				},
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeMap,
						Value: map[string]ast.Variable{
							"KEY": ast.Variable{
								Type:  ast.TypeString,
								Value: "value",
							},
						},
					},
					"bar": ast.Variable{
						Value: "key",
						Type:  ast.TypeString,
					},
				},
			},
			false,
			"value",
			TypeString,
		},
		{
			`${foo[upper(bar)]}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"upper": ast.Function{
						ArgTypes:   []ast.Type{ast.TypeString},
						ReturnType: ast.TypeString,
						Callback: func(args []interface{}) (interface{}, error) {
							return strings.ToUpper(args[0].(string)), nil
						},
					},
				},
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeMap,
						Value: map[string]ast.Variable{
							"KEY": ast.Variable{
								Type:  ast.TypeString,
								Value: "value",
							},
						},
					},
					"bar": ast.Variable{
						Type: ast.TypeUnknown,
					},
				},
			},
			false,
			UnknownValue,
			TypeUnknown,
		},
		{
			`${upper(foo)}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"upper": ast.Function{
						ArgTypes:   []ast.Type{ast.TypeMap},
						ReturnType: ast.TypeString,
						Callback: func(args []interface{}) (interface{}, error) {
							return "foo", nil
						},
					},
				},
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeMap,
						Value: map[string]ast.Variable{
							"KEY": ast.Variable{
								Type: ast.TypeUnknown,
							},
						},
					},
				},
			},
			false,
			UnknownValue,
			TypeUnknown,
		},
		{
			Input:      `${"foo\\"}`,
			Scope:      nil,
			Result:     `foo\`,
			ResultType: TypeString,
		},
		{
			Input:      `${"foo\\\\"}`,
			Scope:      nil,
			Result:     `foo\\`,
			ResultType: TypeString,
		},
		{
			`${second("foo", "\\", "/", "bar")}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"second": {
						ArgTypes:   []ast.Type{ast.TypeString, ast.TypeString, ast.TypeString, ast.TypeString},
						ReturnType: ast.TypeString,
						Callback: func(args []interface{}) (interface{}, error) {
							return args[1].(string), nil
						},
					},
				},
			},
			false,
			`\`,
			TypeString,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.Input), func(t *testing.T) {
			node, err := Parse(tc.Input)
			if err != nil {
				t.Fatalf("Error: %s\n\nInput: %s", err, tc.Input)
			}

			result, err := Eval(node, &EvalConfig{GlobalScope: tc.Scope})
			if err != nil != tc.Error {
				t.Fatalf("Error: %s\n\nInput: %s", err, tc.Input)
			}
			if tc.ResultType != TypeInvalid && result.Type != tc.ResultType {
				t.Fatalf("Bad: %s\n\nInput: %s", result.Type, tc.Input)
			}
			if !reflect.DeepEqual(result.Value, tc.Result) {
				t.Fatalf("\n     Got: %#v\nExpected: %#v\n\n   Input: %s\n", result.Value, tc.Result, tc.Input)
			}
		})
	}
}

func TestEvalInternal(t *testing.T) {
	cases := []struct {
		Input      string
		Scope      *ast.BasicScope
		Error      bool
		Result     interface{}
		ResultType ast.Type
	}{
		{
			"foo",
			nil,
			false,
			"foo",
			ast.TypeString,
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
			"foo baz",
			ast.TypeString,
		},

		{
			"${var.alist}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.alist": ast.Variable{
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
			[]ast.Variable{
				ast.Variable{
					Type:  ast.TypeString,
					Value: "Hello",
				},
				ast.Variable{
					Type:  ast.TypeString,
					Value: "World",
				},
			},
			ast.TypeList,
		},

		{
			"foo ${-29}",
			nil,
			false,
			"foo -29",
			ast.TypeString,
		},

		{
			"foo ${42+1}",
			nil,
			false,
			"foo 43",
			ast.TypeString,
		},

		{
			"foo ${42-1}",
			nil,
			false,
			"foo 41",
			ast.TypeString,
		},

		{
			"foo ${42*2}",
			nil,
			false,
			"foo 84",
			ast.TypeString,
		},

		{
			"foo ${42/2}",
			nil,
			false,
			"foo 21",
			ast.TypeString,
		},

		{
			"foo ${42/0}",
			nil,
			true,
			"foo ",
			ast.TypeInvalid,
		},

		{
			"foo ${42%4}",
			nil,
			false,
			"foo 2",
			ast.TypeString,
		},

		{
			"foo ${42%0}",
			nil,
			true,
			"foo ",
			ast.TypeInvalid,
		},

		{
			"foo ${42.0+1.0}",
			nil,
			false,
			"foo 43",
			ast.TypeString,
		},

		{
			"foo ${42.0+1}",
			nil,
			false,
			"foo 43",
			ast.TypeString,
		},

		{
			"foo ${42+1.0}",
			nil,
			false,
			"foo 43",
			ast.TypeString,
		},

		{
			"foo ${0.5 * 75}",
			nil,
			false,
			"foo 37.5",
			ast.TypeString,
		},

		{
			"foo ${75 * 0.5}",
			nil,
			false,
			"foo 37.5",
			ast.TypeString,
		},

		{
			"foo ${42+2*2}",
			nil,
			false,
			"foo 46",
			ast.TypeString,
		},

		{
			"foo ${42+(2*2)}",
			nil,
			false,
			"foo 46",
			ast.TypeString,
		},

		{
			"foo ${true && false}",
			nil,
			false,
			"foo false",
			ast.TypeString,
		},

		{
			"foo ${false || true}",
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${"true" || true}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${true || "true"}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${"true" || "true"}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			"foo ${1 == 2}",
			nil,
			false,
			"foo false",
			ast.TypeString,
		},

		{
			"foo ${1 == 1}",
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			"foo ${1 > 2}",
			nil,
			false,
			"foo false",
			ast.TypeString,
		},

		{
			"foo ${2 > 1}",
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${"hello" == "hello"}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${"hello" == "goodbye"}`,
			nil,
			false,
			"foo false",
			ast.TypeString,
		},

		{
			`foo ${1 == "2"}`,
			nil,
			false,
			"foo false",
			ast.TypeString,
		},

		{
			`foo ${1 == "1"}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${"1" == 1}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${1.2 == 1}`,
			nil,
			false,
			"foo false",
			ast.TypeString,
		},

		{
			// implicit conversion of float to int makes this equal
			`foo ${1 == 1.2}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${true == false}`,
			nil,
			false,
			"foo false",
			ast.TypeString,
		},

		{
			`foo ${false == false}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${"true" == true}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${true == "true"}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			`foo ${! true}`,
			nil,
			false,
			"foo false",
			ast.TypeString,
		},

		{
			`foo ${! false}`,
			nil,
			false,
			"foo true",
			ast.TypeString,
		},

		{
			"foo ${true ? 5 : 7}",
			nil,
			false,
			"foo 5",
			ast.TypeString,
		},

		{
			"foo ${false ? 5 : 7}",
			nil,
			false,
			"foo 7",
			ast.TypeString,
		},

		{
			`foo ${"true" ? 5 : 7}`,
			nil,
			false,
			"foo 5",
			ast.TypeString,
		},

		{
			// false expression is type-converted to match true expression
			`foo ${false ? 5 : 6.5}`,
			nil,
			false,
			"foo 6",
			ast.TypeString,
		},

		{
			// true expression is type-converted to match false expression
			// if the true expression is string
			`foo ${false ? "12" : 16}`,
			nil,
			false,
			"foo 16",
			ast.TypeString,
		},

		{
			"foo ${3 > 2 ? 5 : 7}",
			nil,
			false,
			"foo 5",
			ast.TypeString,
		},

		{
			"${var.do_it ? 5 : 7}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.do_it": ast.Variable{
						Value: UnknownValue,
						Type:  ast.TypeUnknown,
					},
				},
			},
			false,
			UnknownValue,
			ast.TypeUnknown,
		},

		{
			// false expression can be unknown, and is returned
			`foo ${false ? "12" : unknown}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"unknown": ast.Variable{
						Value: UnknownValue,
						Type:  ast.TypeUnknown,
					},
				},
			},
			false,
			UnknownValue,
			ast.TypeUnknown,
		},

		{
			// false expression can be unknown, and result is unknown even
			// if it's not selected.
			// (Ideally this would not be true, but we're accepting this
			// for now since this assumption is built in to the core evaluator)
			`foo ${true ? "12" : unknown}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"unknown": ast.Variable{
						Value: UnknownValue,
						Type:  ast.TypeUnknown,
					},
				},
			},
			false,
			UnknownValue,
			ast.TypeUnknown,
		},

		{
			// true expression can be unknown, and is returned
			`foo ${false ? unknown : "bar"}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"unknown": ast.Variable{
						Value: UnknownValue,
						Type:  ast.TypeUnknown,
					},
				},
			},
			false,
			UnknownValue,
			ast.TypeUnknown,
		},

		{
			// true expression can be unknown, and result is unknown even
			// if it's not selected.
			// (Ideally this would not be true, but we're accepting this
			// for now since this assumption is built in to the core evaluator)
			`foo ${false ? unknown : "bar"}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"unknown": ast.Variable{
						Value: UnknownValue,
						Type:  ast.TypeUnknown,
					},
				},
			},
			false,
			UnknownValue,
			ast.TypeUnknown,
		},

		{
			// both values can be unknown
			`foo ${false ? unknown : unknown}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"unknown": ast.Variable{
						Value: UnknownValue,
						Type:  ast.TypeUnknown,
					},
				},
			},
			false,
			UnknownValue,
			ast.TypeUnknown,
		},

		{
			// condition can be unknown, and result is unknown
			`foo ${unknown ? "baz" : "bar"}`,
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"unknown": ast.Variable{
						Value: UnknownValue,
						Type:  ast.TypeUnknown,
					},
				},
			},
			false,
			UnknownValue,
			ast.TypeUnknown,
		},

		{
			"foo ${-bar}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: 41,
						Type:  ast.TypeInt,
					},
				},
			},
			false,
			"foo -41",
			ast.TypeString,
		},

		{
			"foo ${bar+1}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: 41,
						Type:  ast.TypeInt,
					},
				},
			},
			false,
			"foo 42",
			ast.TypeString,
		},

		{
			"foo ${bar+1}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: "41",
						Type:  ast.TypeString,
					},
				},
			},
			false,
			"foo 42",
			ast.TypeString,
		},

		{
			"foo ${bar+baz}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: "41",
						Type:  ast.TypeString,
					},
					"baz": ast.Variable{
						Value: "1",
						Type:  ast.TypeString,
					},
				},
			},
			false,
			"foo 42",
			ast.TypeString,
		},

		{
			"foo ${bar+baz}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: 0.001,
						Type:  ast.TypeFloat,
					},
					"baz": ast.Variable{
						Value: "0.002",
						Type:  ast.TypeString,
					},
				},
			},
			false,
			"foo 0.003",
			ast.TypeString,
		},

		{
			"foo ${bar+baz}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: UnknownValue,
						Type:  ast.TypeUnknown,
					},
					"baz": ast.Variable{
						Value: 1,
						Type:  ast.TypeInt,
					},
				},
			},
			false,
			UnknownValue,
			ast.TypeUnknown,
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
			"foo 42",
			ast.TypeString,
		},

		{
			`foo ${rand("foo", "bar")}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"rand": ast.Function{
						ReturnType:   ast.TypeString,
						Variadic:     true,
						VariadicType: ast.TypeString,
						Callback: func(args []interface{}) (interface{}, error) {
							var result string
							for _, a := range args {
								result += a.(string)
							}
							return result, nil
						},
					},
				},
			},
			false,
			"foo foobar",
			ast.TypeString,
		},

		{
			`${foo["bar"]}`,
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
				},
			},
			false,
			"world",
			ast.TypeString,
		},

		{
			`${foo[var.key]}`,
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
					"var.key": ast.Variable{
						Type:  ast.TypeString,
						Value: "bar",
					},
				},
			},
			false,
			"world",
			ast.TypeString,
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
			"world",
			ast.TypeString,
		},

		{
			`${foo["bar"]} ${bar[1]}`,
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
								Type:  ast.TypeInt,
								Value: 10,
							},
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 20,
							},
						},
					},
				},
			},
			false,
			"world 20",
			ast.TypeString,
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
								Value: "hello",
							},
							ast.Variable{
								Type:  ast.TypeString,
								Value: "world",
							},
						},
					},
				},
			},
			false,
			"hello",
			ast.TypeString,
		},

		{
			"${foo[bar]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeString,
								Value: "hello",
							},
							ast.Variable{
								Type:  ast.TypeString,
								Value: "world",
							},
						},
					},
					"bar": ast.Variable{
						Type:  ast.TypeInt,
						Value: 1,
					},
				},
			},
			false,
			"world",
			ast.TypeString,
		},

		{
			"${foo[bar[1]]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeString,
								Value: "hello",
							},
							ast.Variable{
								Type:  ast.TypeString,
								Value: "world",
							},
						},
					},
					"bar": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 1,
							},
							ast.Variable{
								Type:  ast.TypeInt,
								Value: 0,
							},
						},
					},
				},
			},
			false,
			"hello",
			ast.TypeString,
		},

		{
			"aaa ${foo} aaa",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"foo": ast.Variable{
						Type:  ast.TypeInt,
						Value: 42,
					},
				},
			},
			false,
			"aaa 42 aaa",
			ast.TypeString,
		},

		{
			"aaa ${foo[1]} aaa",
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
								Value: 24,
							},
						},
					},
				},
			},
			false,
			"aaa 24 aaa",
			ast.TypeString,
		},

		{
			"aaa ${foo[1]} - ${foo[0]}",
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
								Value: 24,
							},
						},
					},
				},
			},
			false,
			"aaa 24 - 42",
			ast.TypeString,
		},

		{
			"${var.foo} ${var.foo[0]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.foo": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeString,
								Value: "hello",
							},
							ast.Variable{
								Type:  ast.TypeString,
								Value: "world",
							},
						},
					},
				},
			},
			true,
			nil,
			ast.TypeInvalid,
		},

		{
			"${var.foo[0]} ${var.foo[1]}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.foo": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeString,
								Value: "hello",
							},
							ast.Variable{
								Type:  ast.TypeString,
								Value: "world",
							},
						},
					},
				},
			},
			false,
			"hello world",
			ast.TypeString,
		},

		{
			"${foo[1]} ${foo[0]}",
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
								Value: 24,
							},
						},
					},
				},
			},
			false,
			"24 42",
			ast.TypeString,
		},

		{
			"${foo[1-3]}",
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
								Value: 24,
							},
						},
					},
				},
			},
			true,
			nil,
			ast.TypeInvalid,
		},

		{
			"${foo[2]}",
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
								Value: 24,
							},
						},
					},
				},
			},
			true,
			nil,
			ast.TypeInvalid,
		},

		// Testing implicit type conversions

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
			"foo 42",
			ast.TypeString,
		},

		{
			`foo ${foo("42")}`,
			&ast.BasicScope{
				FuncMap: map[string]ast.Function{
					"foo": ast.Function{
						ArgTypes:   []ast.Type{ast.TypeInt},
						ReturnType: ast.TypeString,
						Callback: func(args []interface{}) (interface{}, error) {
							return strconv.FormatInt(int64(args[0].(int)), 10), nil
						},
					},
				},
			},
			false,
			"foo 42",
			ast.TypeString,
		},

		// Multiline
		{
			"foo ${42+\n1.0}",
			nil,
			false,
			"foo 43",
			ast.TypeString,
		},

		// String vars should be able to implictly convert to floats
		{
			"${1.5 * var.foo}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.foo": ast.Variable{
						Value: "42",
						Type:  ast.TypeString,
					},
				},
			},
			false,
			"63",
			ast.TypeString,
		},

		// Unary
		{
			"foo ${-46}",
			nil,
			false,
			"foo -46",
			ast.TypeString,
		},

		{
			"foo ${-46 + 5}",
			nil,
			false,
			"foo -41",
			ast.TypeString,
		},

		{
			"foo ${46 + -5}",
			nil,
			false,
			"foo 41",
			ast.TypeString,
		},

		{
			"foo ${-bar}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: 41,
						Type:  ast.TypeInt,
					},
				},
			},
			false,
			"foo -41",
			ast.TypeString,
		},

		{
			"foo ${5 + -bar}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"bar": ast.Variable{
						Value: 41,
						Type:  ast.TypeInt,
					},
				},
			},
			false,
			"foo -36",
			ast.TypeString,
		},

		{
			"${var.foo > 1 ? 5 : 0}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.foo": ast.Variable{
						Type:  ast.TypeString,
						Value: "3",
					},
				},
			},
			false,
			"5",
			ast.TypeString,
		},

		{
			"${var.foo > 1.5 ? 5 : 0}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.foo": ast.Variable{
						Type:  ast.TypeString,
						Value: "3",
					},
				},
			},
			false,
			"5",
			ast.TypeString,
		},

		{
			"${var.foo > 1.5 ? 5 : 0}",
			&ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.foo": ast.Variable{
						Type:  ast.TypeString,
						Value: "1.2",
					},
				},
			},
			false,
			"0",
			ast.TypeString,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.Input), func(t *testing.T) {
			node, err := Parse(tc.Input)
			if err != nil {
				t.Fatalf("Error: %s\n\nInput: %s", err, tc.Input)
			}

			out, outType, err := internalEval(node, &EvalConfig{GlobalScope: tc.Scope})
			if err != nil != tc.Error {
				t.Fatalf("Error: %s\nInput: %s", err, tc.Input)
			}
			if tc.ResultType != ast.TypeInvalid && outType != tc.ResultType {
				t.Fatalf("Wrong result type\nInput: %s\nGot:   %#s\nWant:  %s", tc.Input, outType, tc.ResultType)
			}
			if !reflect.DeepEqual(out, tc.Result) {
				t.Fatalf("Wrong result value\nInput: %s\nGot:   %#s\nWant:  %s", tc.Input, out, tc.Result)
			}
		})
	}
}
