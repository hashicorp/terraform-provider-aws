package diff

import (
	"fmt"
	"testing"

	"reflect"

	"github.com/zclconf/go-cty/cty"
	"github.com/kylelemons/godebug/pretty"
)

func TestDiffListsShallow(t *testing.T) {
	tests := []struct {
		Old  []cty.Value
		New  []cty.Value
		Want Diff
	}{
		{
			[]cty.Value{},
			[]cty.Value{},
			Diff(nil),
		},
		{
			[]cty.Value{cty.NumberIntVal(1)},
			[]cty.Value{},
			Diff{
				DeleteChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					OldValue: cty.NumberIntVal(1),
				},
			},
		},
		{
			[]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)},
			[]cty.Value{},
			Diff{
				DeleteChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					OldValue: cty.NumberIntVal(1),
				},
				DeleteChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					OldValue: cty.NumberIntVal(2),
				},
			},
		},
		{
			[]cty.Value{},
			[]cty.Value{cty.NumberIntVal(1)},
			Diff{
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					NewValue:    cty.NumberIntVal(1),
					BeforeValue: cty.NullVal(cty.Number),
				},
			},
		},
		{
			[]cty.Value{},
			[]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)},
			Diff{
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					NewValue:    cty.NumberIntVal(1),
					BeforeValue: cty.NullVal(cty.Number),
				},
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(1)},
					},
					NewValue:    cty.NumberIntVal(2),
					BeforeValue: cty.NullVal(cty.Number),
				},
			},
		},
		{
			[]cty.Value{cty.NumberIntVal(1)},
			[]cty.Value{cty.NumberIntVal(1)},
			Diff{
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					WantValue: cty.NumberIntVal(1),
				},
			},
		},
		{
			[]cty.Value{cty.NumberIntVal(2)},
			[]cty.Value{cty.NumberIntVal(1)},
			Diff{
				DeleteChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					OldValue: cty.NumberIntVal(2),
				},
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					NewValue:    cty.NumberIntVal(1),
					BeforeValue: cty.NullVal(cty.Number),
				},
			},
		},
		{
			[]cty.Value{cty.NumberIntVal(2)},
			[]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)},
			Diff{
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					NewValue:    cty.NumberIntVal(1),
					BeforeValue: cty.NumberIntVal(2),
				},
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(1)},
					},
					WantValue: cty.NumberIntVal(2),
				},
			},
		},
		{
			[]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)},
			[]cty.Value{cty.NumberIntVal(2)},
			Diff{
				DeleteChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					OldValue: cty.NumberIntVal(1),
				},
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					WantValue: cty.NumberIntVal(2),
				},
			},
		},
		{
			[]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(4),
				cty.NumberIntVal(6),
			},
			[]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
				cty.NumberIntVal(3),
				cty.NumberIntVal(4),
				cty.NumberIntVal(5),
				cty.NumberIntVal(6),
			},
			Diff{
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					WantValue: cty.NumberIntVal(1),
				},
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(1)},
					},
					NewValue:    cty.NumberIntVal(2),
					BeforeValue: cty.NumberIntVal(4),
				},
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(2)},
					},
					NewValue:    cty.NumberIntVal(3),
					BeforeValue: cty.NumberIntVal(4),
				},
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(3)},
					},
					WantValue: cty.NumberIntVal(4),
				},
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(4)},
					},
					NewValue:    cty.NumberIntVal(5),
					BeforeValue: cty.NumberIntVal(6),
				},
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(5)},
					},
					WantValue: cty.NumberIntVal(6),
				},
			},
		},
		{
			[]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(4),
				cty.NumberIntVal(6),
			},
			[]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
				cty.NumberIntVal(3),
				cty.NumberIntVal(5),
				cty.NumberIntVal(6),
			},
			Diff{
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					WantValue: cty.NumberIntVal(1),
				},
				DeleteChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(1)},
					},
					OldValue: cty.NumberIntVal(4),
				},
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(1)},
					},
					NewValue:    cty.NumberIntVal(2),
					BeforeValue: cty.NumberIntVal(6),
				},
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(2)},
					},
					NewValue:    cty.NumberIntVal(3),
					BeforeValue: cty.NumberIntVal(6),
				},
				InsertChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(3)},
					},
					NewValue:    cty.NumberIntVal(5),
					BeforeValue: cty.NumberIntVal(6),
				},
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(4)},
					},
					WantValue: cty.NumberIntVal(6),
				},
			},
		},
		{
			[]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
				cty.NumberIntVal(2),
				cty.NumberIntVal(2),
				cty.NumberIntVal(3),
			},
			[]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
				cty.NumberIntVal(3),
			},
			Diff{
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(0)},
					},
					WantValue: cty.NumberIntVal(1),
				},
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(1)},
					},
					WantValue: cty.NumberIntVal(2),
				},
				DeleteChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(2)},
					},
					OldValue: cty.NumberIntVal(2),
				},
				DeleteChange{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(2)},
					},
					OldValue: cty.NumberIntVal(2),
				},
				Context{
					Path: cty.Path{
						cty.IndexStep{Key: cty.NumberIntVal(2)},
					},
					WantValue: cty.NumberIntVal(3),
				},
			},
		},
	}

	pr := &pretty.Config{
		Diffable: true,
		Formatter: map[reflect.Type]interface{}{
			reflect.TypeOf(cty.NilVal): func(val cty.Value) string {
				return val.GoString()
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v,%#v", test.Old, test.New), func(t *testing.T) {
			var ov cty.Value
			var nv cty.Value
			if len(test.Old) == 0 {
				ov = cty.ListValEmpty(cty.Number)
			} else {
				ov = cty.ListVal(test.Old)
			}
			if len(test.New) == 0 {
				nv = cty.ListValEmpty(cty.Number)
			} else {
				nv = cty.ListVal(test.New)
			}

			got := diffListsShallow(ov, nv, cty.Path(nil))

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\n%s", pr.Compare(test.Want, got))
			}
		})
	}
}
