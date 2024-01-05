// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestJSONTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.JSONNull(),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.JSONUnknown(),
		},
		"valid JSON": {
			val:      tftypes.NewValue(tftypes.String, "{\"test\": \"value\"}"),
			expected: fwtypes.JSONValue("{\"test\":\"value\"}"),
		},
		"invalid value": {
			val:      tftypes.NewValue(tftypes.String, "not ok"),
			expected: fwtypes.JSONUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.JSONType.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if !test.expected.Equal(val) {
				t.Errorf("unexpected diff\nwanted: %s\ngot:    %s", test.expected, val)
			}
		})
	}
}

func TestJSONTypeValidate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         tftypes.Value
		expectError bool
	}
	tests := map[string]testCase{
		"not a string": {
			val:         tftypes.NewValue(tftypes.Bool, true),
			expectError: true,
		},
		"unknown string": {
			val: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"null string": {
			val: tftypes.NewValue(tftypes.String, nil),
		},
		"valid JSON": {
			val: tftypes.NewValue(tftypes.String, "{\"test\": \"value\"}"),
		},
		"invalid string": {
			val:         tftypes.NewValue(tftypes.String, "not ok"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			diags := fwtypes.JSONType.Validate(ctx, test.val, path.Root("test"))

			if !diags.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if diags.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %#v", diags)
			}
		})
	}
}
