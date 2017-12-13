package stdlib

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestBytesLen(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			BytesVal([]byte{}),
			cty.NumberIntVal(0),
		},
		{
			BytesVal([]byte{'a'}),
			cty.NumberIntVal(1),
		},
		{
			BytesVal([]byte{'a', 'b', 'c'}),
			cty.NumberIntVal(3),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := BytesLen(test.Input)

			if err != nil {
				t.Fatal(err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf(
					"wrong result\ninput: %#v\ngot:   %#v\nwant:  %#v",
					test.Input, got, test.Want,
				)
			}
		})
	}
}

func TestBytesSlice(t *testing.T) {
	tests := []struct {
		Input  cty.Value
		Offset cty.Value
		Length cty.Value
		Want   cty.Value
	}{
		{
			BytesVal([]byte{}),
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
			BytesVal([]byte{}),
		},
		{
			BytesVal([]byte{'a'}),
			cty.NumberIntVal(0),
			cty.NumberIntVal(1),
			BytesVal([]byte{'a'}),
		},
		{
			BytesVal([]byte{'a', 'b', 'c'}),
			cty.NumberIntVal(0),
			cty.NumberIntVal(2),
			BytesVal([]byte{'a', 'b'}),
		},
		{
			BytesVal([]byte{'a', 'b', 'c'}),
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			BytesVal([]byte{'b', 'c'}),
		},
		{
			BytesVal([]byte{'a', 'b', 'c'}),
			cty.NumberIntVal(0),
			cty.NumberIntVal(3),
			BytesVal([]byte{'a', 'b', 'c'}),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := BytesSlice(test.Input, test.Offset, test.Length)

			if err != nil {
				t.Fatal(err)
			}

			gotBytes := *(got.EncapsulatedValue().(*[]byte))
			wantBytes := *(test.Want.EncapsulatedValue().(*[]byte))

			if !reflect.DeepEqual(gotBytes, wantBytes) {
				t.Errorf(
					"wrong result\ninput: %#v, %#v,  %#v\ngot:   %#v\nwant:  %#v",
					test.Input, test.Offset, test.Length, got, test.Want,
				)
			}
		})
	}
}
