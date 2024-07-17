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

func TestDurationTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.DurationNull(),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.DurationUnknown(),
		},
		"valid duration": {
			val:      tftypes.NewValue(tftypes.String, "2h"),
			expected: fwtypes.DurationValue("2h"),
		},
		"invalid duration": {
			val:      tftypes.NewValue(tftypes.String, "not ok"),
			expected: fwtypes.DurationUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.DurationType.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(val, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestDurationValidateAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         fwtypes.Duration
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val: fwtypes.DurationUnknown(),
		},
		"null": {
			val: fwtypes.DurationNull(),
		},
		"valid": {
			val: fwtypes.DurationValue("2h"),
		},
		"invalid": {
			val:         fwtypes.DurationValue("not ok"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
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

func TestDurationToStringValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		duration fwtypes.Duration
		expected types.String
	}{
		"value": {
			duration: fwtypes.DurationValue("2h"),
			expected: types.StringValue("2h"),
		},
		"null": {
			duration: fwtypes.DurationNull(),
			expected: types.StringNull(),
		},
		"unknown": {
			duration: fwtypes.DurationUnknown(),
			expected: types.StringUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
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
