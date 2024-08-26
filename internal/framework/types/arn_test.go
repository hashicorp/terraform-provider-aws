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
			val:      tftypes.NewValue(tftypes.String, "arn:aws:rds:us-east-1:123456789012:db:test"), // lintignore:AWSAT003,AWSAT005
			expected: fwtypes.ARNValue("arn:aws:rds:us-east-1:123456789012:db:test"),                 // lintignore:AWSAT003,AWSAT005
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

func TestARNValidateAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         fwtypes.ARN
		expectError bool
	}
	tests := map[string]testCase{
		"null value": {
			val: fwtypes.ARNNull(),
		},
		"unknown value": {
			val: fwtypes.ARNUnknown(),
		},
		"valid arn": {
			val: fwtypes.ARNValue("arn:aws:rds:us-east-1:123456789012:db:test"), // lintignore:AWSAT003,AWSAT005
		},
		"invalid arn": {
			val:         fwtypes.ARNValue("not ok"), // lintignore:AWSAT003,AWSAT005
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
