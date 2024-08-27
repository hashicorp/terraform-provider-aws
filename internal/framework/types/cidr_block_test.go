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

func TestCIDRBlockTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.CIDRBlockNull(),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.CIDRBlockUnknown(),
		},
		"valid CIDR block": {
			val:      tftypes.NewValue(tftypes.String, "0.0.0.0/0"),
			expected: fwtypes.CIDRBlockValue("0.0.0.0/0"),
		},
		"invalid CIDR block": {
			val:      tftypes.NewValue(tftypes.String, "not ok"),
			expected: fwtypes.CIDRBlockUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.CIDRBlockType.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(val, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestCIDRBlockValidateAttrbute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         fwtypes.CIDRBlock
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val: fwtypes.CIDRBlockUnknown(),
		},
		"null": {
			val: fwtypes.CIDRBlockNull(),
		},
		"valid IPv4": {
			val: fwtypes.CIDRBlockValue("10.2.2.0/24"),
		},
		"invalid IPv4": {
			val:         fwtypes.CIDRBlockValue("10.2.2.2/24"),
			expectError: true,
		},
		"valid IPv6": {
			val: fwtypes.CIDRBlockValue("2000::/15"),
		},
		"invalid IPv6": {
			val:         fwtypes.CIDRBlockValue("2001::/15"),
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

func TestCIDRBlockToStringValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		cidrBlock fwtypes.CIDRBlock
		expected  types.String
	}{
		"value": {
			cidrBlock: fwtypes.CIDRBlockValue("10.2.2.0/24"),
			expected:  types.StringValue("10.2.2.0/24"),
		},
		"null": {
			cidrBlock: fwtypes.CIDRBlockNull(),
			expected:  types.StringNull(),
		},
		"unknown": {
			cidrBlock: fwtypes.CIDRBlockUnknown(),
			expected:  types.StringUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			s, _ := test.cidrBlock.ToStringValue(ctx)

			if !test.expected.Equal(s) {
				t.Fatalf("expected %#v to equal %#v", s, test.expected)
			}
		})
	}
}
