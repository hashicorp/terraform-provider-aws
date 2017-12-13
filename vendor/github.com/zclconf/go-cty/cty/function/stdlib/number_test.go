package stdlib

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestAbsolute(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.NumberIntVal(15),
			cty.NumberIntVal(15),
		},
		{
			cty.NumberIntVal(-15),
			cty.NumberIntVal(15),
		},
		{
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
		},
		{
			cty.PositiveInfinity,
			cty.PositiveInfinity,
		},
		{
			cty.NegativeInfinity,
			cty.PositiveInfinity,
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Absolute(%#v)", test.Input), func(t *testing.T) {
			got, err := Absolute(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.NumberIntVal(3),
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Add(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Add(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Subtract(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Subtract(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(5),
			cty.NumberIntVal(2),
			cty.NumberIntVal(10),
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Multiply(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Multiply(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(5),
			cty.NumberIntVal(2),
			cty.NumberFloatVal(2.5),
		},
		{
			cty.NumberIntVal(5),
			cty.NumberIntVal(0),
			cty.PositiveInfinity,
		},
		{
			cty.NumberIntVal(-5),
			cty.NumberIntVal(0),
			cty.NegativeInfinity,
		},
		{
			cty.NumberIntVal(1),
			cty.PositiveInfinity,
			cty.Zero,
		},
		{
			cty.NumberIntVal(1),
			cty.NegativeInfinity,
			cty.Zero,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Divide(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Divide(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestModulo(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(15),
			cty.NumberIntVal(10),
			cty.NumberIntVal(5),
		},
		{
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
		},
		{
			cty.PositiveInfinity,
			cty.NumberIntVal(1),
			cty.PositiveInfinity,
		},
		{
			cty.NegativeInfinity,
			cty.NumberIntVal(1),
			cty.NegativeInfinity,
		},
		{
			cty.NumberIntVal(1),
			cty.PositiveInfinity,
			cty.PositiveInfinity,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Modulo(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Modulo(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestNegate(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.NumberIntVal(15),
			cty.NumberIntVal(-15),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Negate(%#v)", test.Input), func(t *testing.T) {
			got, err := Negate(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLessThan(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.True,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(1),
			cty.False,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.NumberIntVal(1),
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
		t.Run(fmt.Sprintf("LessThan(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := LessThan(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLessThanOrEqualTo(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.True,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(1),
			cty.False,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.True,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.NumberIntVal(1),
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
		t.Run(fmt.Sprintf("LessThanOrEqualTo(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := LessThanOrEqualTo(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(1),
			cty.True,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.NumberIntVal(1),
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
		t.Run(fmt.Sprintf("GreaterThan(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := GreaterThan(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestGreaterThanOrEqualTo(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(1),
			cty.True,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.True,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.NumberIntVal(1),
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
		t.Run(fmt.Sprintf("GreaterThanOrEqualTo(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := GreaterThanOrEqualTo(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		Inputs []cty.Value
		Want   cty.Value
	}{
		{
			[]cty.Value{cty.NumberIntVal(0)},
			cty.NumberIntVal(0),
		},
		{
			[]cty.Value{cty.NumberIntVal(-12)},
			cty.NumberIntVal(-12),
		},
		{
			[]cty.Value{cty.NumberIntVal(12)},
			cty.NumberIntVal(12),
		},
		{
			[]cty.Value{cty.NumberIntVal(-12), cty.NumberIntVal(0), cty.NumberIntVal(2)},
			cty.NumberIntVal(-12),
		},
		{
			[]cty.Value{cty.NegativeInfinity, cty.NumberIntVal(0)},
			cty.NegativeInfinity,
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.NumberIntVal(0)},
			cty.NumberIntVal(0),
		},
		{
			[]cty.Value{cty.NegativeInfinity},
			cty.NegativeInfinity,
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.UnknownVal(cty.Number)},
			cty.UnknownVal(cty.Number),
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.DynamicVal},
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Inputs), func(t *testing.T) {
			got, err := Min(test.Inputs...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		Inputs []cty.Value
		Want   cty.Value
	}{
		{
			[]cty.Value{cty.NumberIntVal(0)},
			cty.NumberIntVal(0),
		},
		{
			[]cty.Value{cty.NumberIntVal(-12)},
			cty.NumberIntVal(-12),
		},
		{
			[]cty.Value{cty.NumberIntVal(12)},
			cty.NumberIntVal(12),
		},
		{
			[]cty.Value{cty.NumberIntVal(-12), cty.NumberIntVal(0), cty.NumberIntVal(2)},
			cty.NumberIntVal(2),
		},
		{
			[]cty.Value{cty.NegativeInfinity, cty.NumberIntVal(0)},
			cty.NumberIntVal(0),
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.NumberIntVal(0)},
			cty.PositiveInfinity,
		},
		{
			[]cty.Value{cty.NegativeInfinity},
			cty.NegativeInfinity,
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.UnknownVal(cty.Number)},
			cty.UnknownVal(cty.Number),
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.DynamicVal},
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Inputs), func(t *testing.T) {
			got, err := Max(test.Inputs...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
		},
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(1),
		},
		{
			cty.NumberIntVal(-1),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberFloatVal(1.3),
			cty.NumberIntVal(1),
		},
		{
			cty.NumberFloatVal(-1.7),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberFloatVal(-1.3),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberFloatVal(-1.7),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberVal(mustParseFloat("999999999999999999999999999999999999999999999999999999999999.7")),
			cty.NumberVal(mustParseFloat("999999999999999999999999999999999999999999999999999999999999")),
		},
		{
			cty.NumberVal(mustParseFloat("-999999999999999999999999999999999999999999999999999999999999.7")),
			cty.NumberVal(mustParseFloat("-999999999999999999999999999999999999999999999999999999999999")),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Int(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func mustParseFloat(s string) *big.Float {
	ret, _, err := big.ParseFloat(s, 10, 0, big.AwayFromZero)
	if err != nil {
		panic(err)
	}
	return ret
}
