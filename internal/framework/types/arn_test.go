// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestARNTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.ARNNull(),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.ARNUnknown(),
		},
		"valid ARN": {
			val: tftypes.NewValue(tftypes.String, "arn:aws:rds:us-east-1:123456789012:db:test"), // lintignore:AWSAT003,AWSAT005
			expected: fwtypes.ARNValue(arn.ARN{
				Partition: "aws",
				Service:   "rds",
				Region:    "us-east-1", // lintignore:AWSAT003
				AccountID: "123456789012",
				Resource:  "db:test",
			}),
		},
		"invalid ARN": {
			val:      tftypes.NewValue(tftypes.String, "not ok"),
			expected: fwtypes.ARNUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.ARNType.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(val, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestARNTypeValidate(t *testing.T) {
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
		"valid string": {
			val: tftypes.NewValue(tftypes.String, "arn:aws:rds:us-east-1:123456789012:db:test"), // lintignore:AWSAT003,AWSAT005
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

			diags := fwtypes.ARNType.Validate(ctx, test.val, path.Root("test"))

			if !diags.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if diags.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %#v", diags)
			}
		})
	}
}

func TestARNToStringValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		arn      fwtypes.ARN
		expected types.String
	}{
		"value": {
			arn:      arnFromString(t, "arn:aws:rds:us-east-1:123456789012:db:test"),
			expected: types.StringValue("arn:aws:rds:us-east-1:123456789012:db:test"),
		},
		"null": {
			arn:      fwtypes.ARNNull(),
			expected: types.StringNull(),
		},
		"unknown": {
			arn:      fwtypes.ARNUnknown(),
			expected: types.StringUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			s, _ := test.arn.ToStringValue(ctx)

			if !test.expected.Equal(s) {
				t.Fatalf("expected %#v to equal %#v", s, test.expected)
			}
		})
	}
}

func arnFromString(t *testing.T, s string) fwtypes.ARN {
	ctx := context.Background()

	val := tftypes.NewValue(tftypes.String, s)

	attr, err := fwtypes.ARNType.ValueFromTerraform(ctx, val)
	if err != nil {
		t.Fatalf("setting ARN: %s", err)
	}

	return attr.(fwtypes.ARN)
}
