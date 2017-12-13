package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestConcat(t *testing.T) {
	tests := []struct {
		Input []cty.Value
		Want  cty.Value
	}{
		{
			[]cty.Value{
				cty.ListValEmpty(cty.Number),
			},
			cty.ListValEmpty(cty.Number),
		},
		{
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
					cty.NumberIntVal(3),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
				cty.NumberIntVal(3),
			}),
		},
		{
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.NumberIntVal(1),
				}),
				cty.ListVal([]cty.Value{
					cty.NumberIntVal(2),
					cty.NumberIntVal(3),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
				cty.NumberIntVal(3),
			}),
		},
		{
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.NumberIntVal(1),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("foo"),
				}),
				cty.ListVal([]cty.Value{
					cty.True,
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("1"),
				cty.StringVal("foo"),
				cty.StringVal("true"),
			}),
		},
		{
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.NumberIntVal(1),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("foo"),
					cty.StringVal("bar"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("1"),
				cty.StringVal("foo"),
				cty.StringVal("bar"),
			}),
		},
		{
			[]cty.Value{
				cty.EmptyTupleVal,
			},
			cty.EmptyTupleVal,
		},
		{
			[]cty.Value{
				cty.TupleVal([]cty.Value{
					cty.NumberIntVal(1),
					cty.True,
					cty.NumberIntVal(3),
				}),
			},
			cty.TupleVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.True,
				cty.NumberIntVal(3),
			}),
		},
		{
			[]cty.Value{
				cty.TupleVal([]cty.Value{
					cty.NumberIntVal(1),
				}),
				cty.TupleVal([]cty.Value{
					cty.True,
					cty.NumberIntVal(3),
				}),
			},
			cty.TupleVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.True,
				cty.NumberIntVal(3),
			}),
		},
		{
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.NumberIntVal(1),
				}),
				cty.TupleVal([]cty.Value{
					cty.True,
					cty.NumberIntVal(3),
				}),
			},
			cty.TupleVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.True,
				cty.NumberIntVal(3),
			}),
		},
		{
			[]cty.Value{
				cty.TupleVal([]cty.Value{
					cty.NumberIntVal(1),
					cty.True,
				}),
				cty.ListVal([]cty.Value{
					cty.NumberIntVal(3),
				}),
			},
			cty.TupleVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.True,
				cty.NumberIntVal(3),
			}),
		},
		{
			// Two lists with unconvertable element types become a tuple.
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.NumberIntVal(1),
				}),
				cty.ListVal([]cty.Value{
					cty.ListValEmpty(cty.Bool),
				}),
			},
			cty.TupleVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.ListValEmpty(cty.Bool),
			}),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Concat(%#v...)", test.Input), func(t *testing.T) {
			got, err := Concat(test.Input...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
