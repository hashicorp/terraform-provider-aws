package stdlib

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestUpper(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.StringVal("HELLO"),
		},
		{
			cty.StringVal("HELLO"),
			cty.StringVal("HELLO"),
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
		},
		{
			cty.StringVal("1"),
			cty.StringVal("1"),
		},
		{
			cty.StringVal("жж"),
			cty.StringVal("ЖЖ"),
		},
		{
			cty.StringVal("noël"),
			cty.StringVal("NOËL"),
		},
		{
			// Go's case conversions don't handle this ligature, which is
			// unfortunate but is now a compatibility constraint since it
			// would be potentially-breaking to behave differently here in
			// future.
			cty.StringVal("baﬄe"),
			cty.StringVal("BAﬄE"),
		},
		{
			cty.StringVal("😸😾"),
			cty.StringVal("😸😾"),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.String),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Upper(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLower(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("HELLO"),
			cty.StringVal("hello"),
		},
		{
			cty.StringVal("hello"),
			cty.StringVal("hello"),
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
		},
		{
			cty.StringVal("1"),
			cty.StringVal("1"),
		},
		{
			cty.StringVal("ЖЖ"),
			cty.StringVal("жж"),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.String),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Lower(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.StringVal("olleh"),
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
		},
		{
			cty.StringVal("1"),
			cty.StringVal("1"),
		},
		{
			cty.StringVal("Живой Журнал"),
			cty.StringVal("ланруЖ йовиЖ"),
		},
		{
			// note that the dieresis here is intentionally a combining
			// ligature.
			cty.StringVal("noël"),
			cty.StringVal("lëon"),
		},
		{
			// The Es in this string has three combining acute accents.
			// This tests something that NFC-normalization cannot collapse
			// into a single precombined codepoint, since otherwise we might
			// be cheating and relying on the single-codepoint forms.
			cty.StringVal("wé́́é́́é́́!"),
			cty.StringVal("!é́́é́́é́́w"),
		},
		{
			// Go's normalization forms don't handle this ligature, so we
			// will produce the wrong result but this is now a compatibility
			// constraint and so we'll test it.
			cty.StringVal("baﬄe"),
			cty.StringVal("eﬄab"),
		},
		{
			cty.StringVal("😸😾"),
			cty.StringVal("😾😸"),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.String),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Reverse(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestStrlen(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(5),
		},
		{
			cty.StringVal(""),
			cty.NumberIntVal(0),
		},
		{
			cty.StringVal("1"),
			cty.NumberIntVal(1),
		},
		{
			cty.StringVal("Живой Журнал"),
			cty.NumberIntVal(12),
		},
		{
			// note that the dieresis here is intentionally a combining
			// ligature.
			cty.StringVal("noël"),
			cty.NumberIntVal(4),
		},
		{
			// The Es in this string has three combining acute accents.
			// This tests something that NFC-normalization cannot collapse
			// into a single precombined codepoint, since otherwise we might
			// be cheating and relying on the single-codepoint forms.
			cty.StringVal("wé́́é́́é́́!"),
			cty.NumberIntVal(5),
		},
		{
			// Go's normalization forms don't handle this ligature, so we
			// will produce the wrong result but this is now a compatibility
			// constraint and so we'll test it.
			cty.StringVal("baﬄe"),
			cty.NumberIntVal(4),
		},
		{
			cty.StringVal("😸😾"),
			cty.NumberIntVal(2),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Strlen(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestSubstr(t *testing.T) {
	tests := []struct {
		Input  cty.Value
		Offset cty.Value
		Length cty.Value
		Want   cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(0),
			cty.NumberIntVal(2),
			cty.StringVal("he"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(3),
			cty.StringVal("ell"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(-1),
			cty.StringVal("ello"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(-10), // not documented, but <0 is the same as -1
			cty.StringVal("ello"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(10),
			cty.StringVal("ello"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(-3),
			cty.NumberIntVal(-1),
			cty.StringVal("llo"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(-3),
			cty.NumberIntVal(2),
			cty.StringVal("ll"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(10),
			cty.NumberIntVal(10),
			cty.StringVal(""),
		},
		{
			cty.StringVal("noël"),
			cty.NumberIntVal(0),
			cty.NumberIntVal(3),
			cty.StringVal("noë"),
		},
		{
			cty.StringVal("noël"),
			cty.NumberIntVal(3),
			cty.NumberIntVal(-1),
			cty.StringVal("l"),
		},
		{
			cty.StringVal("wé́́é́́é́́!"),
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.StringVal("é́́é́́"),
		},
		{
			cty.StringVal("wé́́é́́é́́!"),
			cty.NumberIntVal(3),
			cty.NumberIntVal(2),
			cty.StringVal("é́́!"),
		},
		{
			cty.StringVal("wé́́é́́é́́!"),
			cty.NumberIntVal(-2),
			cty.NumberIntVal(-1),
			cty.StringVal("é́́!"),
		},
		{
			cty.StringVal("noël"),
			cty.NumberIntVal(-2),
			cty.NumberIntVal(-1),
			cty.StringVal("ël"),
		},
		{
			cty.StringVal("😸😾"),
			cty.NumberIntVal(0),
			cty.NumberIntVal(1),
			cty.StringVal("😸"),
		},
		{
			cty.StringVal("😸😾"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(1),
			cty.StringVal("😾"),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Substr(test.Input, test.Offset, test.Length)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
