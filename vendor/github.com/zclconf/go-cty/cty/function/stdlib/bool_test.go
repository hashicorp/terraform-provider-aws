package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestNot(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.True,
			cty.False,
		},
		{
			cty.False,
			cty.True,
		},
		{
			cty.UnknownVal(cty.Bool),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Not(%#v)", test.Input), func(t *testing.T) {
			got, err := Not(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestAnd(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.False,
			cty.False,
			cty.False,
		},
		{
			cty.False,
			cty.True,
			cty.False,
		},
		{
			cty.True,
			cty.False,
			cty.False,
		},
		{
			cty.True,
			cty.True,
			cty.True,
		},
		{
			cty.True,
			cty.UnknownVal(cty.Bool),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Bool),
			cty.UnknownVal(cty.Bool),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.True,
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
		t.Run(fmt.Sprintf("And(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := And(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestOr(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.False,
			cty.False,
			cty.False,
		},
		{
			cty.False,
			cty.True,
			cty.True,
		},
		{
			cty.True,
			cty.False,
			cty.True,
		},
		{
			cty.True,
			cty.True,
			cty.True,
		},
		{
			cty.True,
			cty.UnknownVal(cty.Bool),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Bool),
			cty.UnknownVal(cty.Bool),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.True,
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
		t.Run(fmt.Sprintf("Or(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Or(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
