// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestStringEnumTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.StringEnumNull[awstypes.AclPermission](),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.StringEnumUnknown[awstypes.AclPermission](),
		},
		"valid enum": {
			val:      tftypes.NewValue(tftypes.String, string(awstypes.AclPermissionRead)),
			expected: fwtypes.StringEnumValue(awstypes.AclPermissionRead),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.StringEnumType[awstypes.AclPermission]().ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(val, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringEnumValidateAttribute(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         fwtypes.StringEnum[awstypes.AclPermission]
		expectError bool
	}{
		"null value": {
			val: fwtypes.StringEnumNull[awstypes.AclPermission](),
		},
		"unknown value": {
			val: fwtypes.StringEnumUnknown[awstypes.AclPermission](),
		},
		"zero value enum": {
			val: fwtypes.StringEnumValue(awstypes.AclPermission("")),
		},
		"valid enum": {
			val: fwtypes.StringEnumValue(awstypes.AclPermissionRead),
		},
		"invalid enum": {
			val:         fwtypes.StringEnumValue(awstypes.AclPermission("invalid")),
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

func TestStringEnumZeroValue(t *testing.T) {
	t.Parallel()

	var x fwtypes.StringEnum[awstypes.AclPermission]
	if got, want := x.IsNull(), true; got != want {
		t.Errorf("IsNull = %t, want %t", got, want)
	}
}
