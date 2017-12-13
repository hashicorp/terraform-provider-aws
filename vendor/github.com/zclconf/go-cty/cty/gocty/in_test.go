package gocty

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/set"
)

func TestIn(t *testing.T) {
	capsuleANative := &capsuleType1Native{"capsuleA"}

	tests := []struct {
		GoValue interface{}
		Type    cty.Type
		Want    cty.Value
	}{
		// Bool
		{
			GoValue: true,
			Type:    cty.Bool,
			Want:    cty.True,
		},
		{
			GoValue: (*bool)(nil),
			Type:    cty.Bool,
			Want:    cty.NullVal(cty.Bool),
		},
		{
			GoValue: ptrToBool(true),
			Type:    cty.Bool,
			Want:    cty.True,
		},

		// String
		{
			GoValue: "hello",
			Type:    cty.String,
			Want:    cty.StringVal("hello"),
		},
		{
			GoValue: ptrToString("hello"),
			Type:    cty.String,
			Want:    cty.StringVal("hello"),
		},
		{
			GoValue: ptrToPtrToString("hello"),
			Type:    cty.String,
			Want:    cty.StringVal("hello"),
		},
		{
			GoValue: (*string)(nil),
			Type:    cty.String,
			Want:    cty.NullVal(cty.String),
		},
		{
			GoValue: nil, // any nil is convertable to a null of any type
			Type:    cty.String,
			Want:    cty.NullVal(cty.String),
		},
		{
			GoValue: (*bool)(nil), // any nil is convertable to a null of any type
			Type:    cty.String,
			Want:    cty.NullVal(cty.String),
		},

		// Number
		{
			GoValue: int(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: int8(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: int16(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: int32(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: int64(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: uint(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: uint8(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: uint16(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: uint32(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: uint64(1),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(1),
		},
		{
			GoValue: float32(1.5),
			Type:    cty.Number,
			Want:    cty.NumberFloatVal(1.5),
		},
		{
			GoValue: float64(1.5),
			Type:    cty.Number,
			Want:    cty.NumberFloatVal(1.5),
		},
		{
			GoValue: big.NewFloat(1.5),
			Type:    cty.Number,
			Want:    cty.NumberFloatVal(1.5),
		},
		{
			GoValue: big.NewInt(5),
			Type:    cty.Number,
			Want:    cty.NumberIntVal(5),
		},
		{
			GoValue: (*int)(nil),
			Type:    cty.Number,
			Want:    cty.NullVal(cty.Number),
		},

		// Lists
		{
			GoValue: []int{},
			Type:    cty.List(cty.Number),
			Want:    cty.ListValEmpty(cty.Number),
		},
		{
			GoValue: []int{1, 2},
			Type:    cty.List(cty.Number),
			Want: cty.ListVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
			}),
		},
		{
			GoValue: &[]int{1, 2},
			Type:    cty.List(cty.Number),
			Want: cty.ListVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
			}),
		},
		{
			GoValue: []int(nil),
			Type:    cty.List(cty.Number),
			Want:    cty.NullVal(cty.List(cty.Number)),
		},
		{
			GoValue: (*[]int)(nil),
			Type:    cty.List(cty.Number),
			Want:    cty.NullVal(cty.List(cty.Number)),
		},
		{
			GoValue: [2]int{1, 2},
			Type:    cty.List(cty.Number),
			Want: cty.ListVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
			}),
		},
		{
			GoValue: [0]int{},
			Type:    cty.List(cty.Number),
			Want:    cty.ListValEmpty(cty.Number),
		},
		{
			GoValue: []int{},
			Type:    cty.Set(cty.Number),
			Want:    cty.SetValEmpty(cty.Number),
		},

		// Sets
		{
			GoValue: []int{1, 2},
			Type:    cty.Set(cty.Number),
			Want: cty.SetVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
			}),
		},
		{
			GoValue: []int{2, 2},
			Type:    cty.Set(cty.Number),
			Want: cty.SetVal([]cty.Value{
				cty.NumberIntVal(2),
			}),
		},
		{
			GoValue: &[]int{1, 2},
			Type:    cty.Set(cty.Number),
			Want: cty.SetVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
			}),
		},
		{
			GoValue: []int(nil),
			Type:    cty.Set(cty.Number),
			Want:    cty.NullVal(cty.Set(cty.Number)),
		},
		{
			GoValue: (*[]int)(nil),
			Type:    cty.Set(cty.Number),
			Want:    cty.NullVal(cty.Set(cty.Number)),
		},
		{
			GoValue: [2]int{1, 2},
			Type:    cty.Set(cty.Number),
			Want: cty.SetVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
			}),
		},
		{
			GoValue: [0]int{},
			Type:    cty.Set(cty.Number),
			Want:    cty.SetValEmpty(cty.Number),
		},
		{
			GoValue: set.NewSet(&testSetRules{}),
			Type:    cty.Set(cty.Number),
			Want:    cty.SetValEmpty(cty.Number),
		},
		{
			GoValue: set.NewSetFromSlice(&testSetRules{}, []interface{}{1, 2}),
			Type:    cty.Set(cty.Number),
			Want: cty.SetVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
			}),
		},

		// Maps
		{
			GoValue: map[string]int{},
			Type:    cty.Map(cty.Number),
			Want:    cty.MapValEmpty(cty.Number),
		},
		{
			GoValue: map[string]int{"one": 1, "two": 2},
			Type:    cty.Map(cty.Number),
			Want: cty.MapVal(map[string]cty.Value{
				"one": cty.NumberIntVal(1),
				"two": cty.NumberIntVal(2),
			}),
		},

		// Objects
		{
			GoValue: struct{}{},
			Type:    cty.EmptyObject,
			Want:    cty.EmptyObjectVal,
		},
		{
			GoValue: struct{ Ignored int }{1},
			Type:    cty.EmptyObject,
			Want:    cty.EmptyObjectVal,
		},
		{
			GoValue: struct{}{},
			Type: cty.Object(map[string]cty.Type{
				"name": cty.String,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"name": cty.NullVal(cty.String),
			}),
		},
		{
			GoValue: struct {
				Name   string `cty:"name"`
				Number int    `cty:"number"`
			}{"Steven", 1},
			Type: cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"name":   cty.StringVal("Steven"),
				"number": cty.NumberIntVal(1),
			}),
		},
		{
			GoValue: struct {
				Name   string `cty:"name"`
				Number int
			}{"Steven", 1},
			Type: cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"name":   cty.StringVal("Steven"),
				"number": cty.NullVal(cty.Number),
			}),
		},
		{
			GoValue: map[string]interface{}{
				"name":   "Steven",
				"number": 1,
			},
			Type: cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"name":   cty.StringVal("Steven"),
				"number": cty.NumberIntVal(1),
			}),
		},
		{
			GoValue: map[string]interface{}{
				"number": 1,
			},
			Type: cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"name":   cty.NullVal(cty.String),
				"number": cty.NumberIntVal(1),
			}),
		},

		// Tuples
		{
			GoValue: []interface{}{},
			Type:    cty.EmptyTuple,
			Want:    cty.EmptyTupleVal,
		},
		{
			GoValue: struct{}{},
			Type:    cty.EmptyTuple,
			Want:    cty.EmptyTupleVal,
		},
		{
			GoValue: testTupleStruct{"Stephen", 23},
			Type:    cty.Tuple([]cty.Type{cty.String, cty.Number}),
			Want: cty.TupleVal([]cty.Value{
				cty.StringVal("Stephen"),
				cty.NumberIntVal(23),
			}),
		},
		{
			GoValue: []interface{}{1, 2, 3},
			Type: cty.Tuple([]cty.Type{
				cty.Number,
				cty.Number,
				cty.Number,
			}),
			Want: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
				cty.NumberIntVal(3),
			}),
		},
		{
			GoValue: []interface{}{1, "hello", 3},
			Type: cty.Tuple([]cty.Type{
				cty.Number,
				cty.String,
				cty.Number,
			}),
			Want: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.StringVal("hello"),
				cty.NumberIntVal(3),
			}),
		},
		{
			GoValue: []interface{}(nil),
			Type:    cty.Tuple([]cty.Type{cty.Number}),
			Want:    cty.NullVal(cty.Tuple([]cty.Type{cty.Number})),
		},

		// Capsules
		{
			GoValue: capsuleANative,
			Type:    capsuleType1,
			Want:    cty.CapsuleVal(capsuleType1, capsuleANative),
		},

		// Dynamic
		{
			GoValue: cty.NumberIntVal(2),
			Type:    cty.DynamicPseudoType,
			Want:    cty.NumberIntVal(2),
		},
		{
			GoValue: []cty.Value{cty.NumberIntVal(2)},
			Type:    cty.List(cty.DynamicPseudoType),
			Want:    cty.ListVal([]cty.Value{cty.NumberIntVal(2)}),
		},
		{
			GoValue: map[string]cty.Value{"number": cty.NumberIntVal(2)},
			Type:    cty.Map(cty.DynamicPseudoType),
			Want:    cty.MapVal(map[string]cty.Value{"number": cty.NumberIntVal(2)}),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v into %#v", test.GoValue, test.Type), func(t *testing.T) {
			got, err := ToCtyValue(test.GoValue, test.Type)
			if err != nil {
				t.Fatalf("ToCtyValue returned error: %s", err)
			}

			if got == cty.NilVal {
				t.Fatalf("ToCtyValue returned NilVal with no error")
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ninput:       %#v\ntarget type: %#v\ngot:         %#v\nwant:        %#v", test.GoValue, test.Type, got, test.Want)
			}
		})
	}
}

func ptrToBool(val bool) *bool {
	return &val
}

func ptrToString(val string) *string {
	return &val
}

func ptrToInt(val int) *int {
	return &val
}

func ptrToPtrToString(val string) **string {
	pval := &val
	return &pval
}

type testSetRules struct{}

func (r testSetRules) Hash(v interface{}) int {
	return v.(int)
}

func (r testSetRules) Equivalent(v1 interface{}, v2 interface{}) bool {
	return v1 == v2
}

type capsuleType1Native struct {
	name string
}

var capsuleType1 = cty.Capsule("capsule type 1", reflect.TypeOf(capsuleType1Native{}))
