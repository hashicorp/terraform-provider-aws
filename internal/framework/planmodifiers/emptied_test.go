// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package planmodifiers

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListEmptied(t *testing.T) {
	t.Parallel()

	type testCase struct {
		old       types.List
		new       types.List
		want      bool
		wantPanic bool
	}

	empty := types.ListValueMust(types.StringType, []attr.Value{})
	nonEmpty := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("value")})
	nonEmptyOther := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other-value")})
	null := types.ListNull(types.StringType)
	unknown := types.ListUnknown(types.StringType)

	tests := map[string]testCase{
		"non-empty list to empty list": {
			old:  nonEmpty,
			new:  empty,
			want: true,
		},
		"non-empty list to null list": {
			old:  nonEmpty,
			new:  null,
			want: true,
		},
		"non-empty list to different non-empty list": {
			old: nonEmpty,
			new: nonEmptyOther,
		},
		"empty list to empty list": {
			old: empty,
			new: empty,
		},
		"empty list to null list": {
			old: empty,
			new: null,
		},
		"null list to empty list": {
			old: null,
			new: empty,
		},
		"null list to non-empty list": {
			old: null,
			new: nonEmpty,
		},
		"null list to null list": {
			old: null,
			new: null,
		},
		"unknown old list": {
			old:       unknown,
			new:       empty,
			wantPanic: true,
		},
		"unknown new list": {
			old:       nonEmpty,
			new:       unknown,
			wantPanic: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got bool
			var gotPanic bool

			func() {
				defer func() {
					gotPanic = recover() != nil
				}()

				got, _ = ListEmptied(t.Context(), test.old, test.new)
			}()

			if gotPanic != test.wantPanic {
				t.Errorf("panic = %t, want %t", gotPanic, test.wantPanic)
			}

			if !gotPanic && got != test.want {
				t.Errorf("ListEmptied() = %t, want %t", got, test.want)
			}
		})
	}
}
