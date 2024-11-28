// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestListOfStringFromTerraform(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"values": {
			val: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "red"),
				tftypes.NewValue(tftypes.String, "blue"),
				tftypes.NewValue(tftypes.String, "green"),
			}),
			expected: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
				types.StringValue("red"),
				types.StringValue("blue"),
				types.StringValue("green"),
			}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			val, err := fwtypes.ListOfStringType.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(val, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
