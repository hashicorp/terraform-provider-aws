// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestRFC3339DurationTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.RFC3339DurationNull(),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.RFC3339DurationUnknown(),
		},
		"valid duration": {
			val:      tftypes.NewValue(tftypes.String, "P2Y"),
			expected: fwtypes.RFC3339DurationValue("P2Y"),
		},
		"invalid duration": {
			val:      tftypes.NewValue(tftypes.String, "not ok"),
			expected: fwtypes.RFC3339DurationUnknown(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.RFC3339DurationType.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(val, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestRFC3339DurationValidateAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         fwtypes.RFC3339Duration
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val: fwtypes.RFC3339DurationUnknown(),
		},
		"null": {
			val: fwtypes.RFC3339DurationNull(),
		},
		"valid": {
			val: fwtypes.RFC3339DurationValue("P2Y"),
		},
		"invalid": {
			val:         fwtypes.RFC3339DurationValue("not ok"),
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			req := xattr.ValidateAttributeRequest{}
			resp := xattr.ValidateAttributeResponse{}

			test.val.ValidateAttribute(ctx, req, &resp)
			if resp.Diagnostics.HasError() != test.expectError {
				t.Errorf("resp.Diagnostics.HasError() = %t, want = %t", resp.Diagnostics.HasError(), test.expectError)
			}
		})
	}
}

func TestRFC3339DurationToStringValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		duration fwtypes.RFC3339Duration
		expected types.String
	}{
		"value": {
			duration: fwtypes.RFC3339DurationValue("P2Y"),
			expected: types.StringValue("P2Y"),
		},
		"null": {
			duration: fwtypes.RFC3339DurationNull(),
			expected: types.StringNull(),
		},
		"unknown": {
			duration: fwtypes.RFC3339DurationUnknown(),
			expected: types.StringUnknown(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			s, _ := test.duration.ToStringValue(ctx)

			if !test.expected.Equal(s) {
				t.Fatalf("expected %#v to equal %#v", s, test.expected)
			}
		})
	}
}
