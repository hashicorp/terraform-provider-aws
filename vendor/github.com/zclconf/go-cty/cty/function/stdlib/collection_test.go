package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestHasIndex(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Key        cty.Value
		Want       cty.Value
	}{
		{
			cty.ListValEmpty(cty.Number),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.MapValEmpty(cty.Bool),
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.StringVal("hello"),
			cty.True,
		},
		{
			cty.EmptyTupleVal,
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.EmptyTupleVal,
			cty.NumberIntVal(0),
			cty.False,
		},
		{
			cty.TupleVal([]cty.Value{cty.True}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("HasIndex(%#v,%#v)", test.Collection, test.Key), func(t *testing.T) {
			got, err := HasIndex(test.Collection, test.Key)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Key        cty.Value
		Want       cty.Value
	}{
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.StringVal("hello"),
			cty.True,
		},
		{
			cty.TupleVal([]cty.Value{cty.True, cty.StringVal("hello")}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.TupleVal([]cty.Value{cty.True, cty.StringVal("hello")}),
			cty.NumberIntVal(1),
			cty.StringVal("hello"),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.MapValEmpty(cty.Number),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.StringVal("hello"),
			cty.DynamicVal,
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.DynamicVal,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Index(%#v,%#v)", test.Collection, test.Key), func(t *testing.T) {
			got, err := Index(test.Collection, test.Key)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Want       cty.Value
	}{
		{
			cty.ListValEmpty(cty.Number),
			cty.NumberIntVal(0),
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.SetValEmpty(cty.Number),
			cty.NumberIntVal(0),
		},
		{
			cty.SetVal([]cty.Value{cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.MapValEmpty(cty.Bool),
			cty.NumberIntVal(0),
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.EmptyTupleVal,
			cty.NumberIntVal(0),
		},
		{
			cty.TupleVal([]cty.Value{cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Length(%#v)", test.Collection), func(t *testing.T) {
			got, err := Length(test.Collection)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
