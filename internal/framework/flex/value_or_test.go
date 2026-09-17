// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestInt64ValueOr(t *testing.T) {
	t.Parallel()

	var defaultValue int64 = 600
	type testCase struct {
		input    types.Int64
		expected int64
	}
	tests := map[string]testCase{
		"null": {
			input:    types.Int64Null(),
			expected: defaultValue,
		},
		"unknown": {
			input:    types.Int64Unknown(),
			expected: defaultValue,
		},
		"value": {
			input:    types.Int64Value(30),
			expected: 30,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int64ValueOr(t.Context(), test.input, defaultValue)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringValueOr(t *testing.T) {
	t.Parallel()

	defaultValue := "THE-DEFAULT"
	type testCase struct {
		input    types.String
		expected string
	}
	tests := map[string]testCase{
		"null": {
			input:    types.StringNull(),
			expected: defaultValue,
		},
		"unknown": {
			input:    types.StringUnknown(),
			expected: defaultValue,
		},
		"value": {
			input:    types.StringValue("THE-VALUE"),
			expected: "THE-VALUE",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.StringValueOr(t.Context(), test.input, defaultValue)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringEnumValueOr(t *testing.T) {
	t.Parallel()

	defaultValue := awstypes.AclPermissionRead
	type testCase struct {
		input    fwtypes.StringEnum[awstypes.AclPermission]
		expected awstypes.AclPermission
	}
	tests := map[string]testCase{
		"null": {
			input:    fwtypes.StringEnumNull[awstypes.AclPermission](),
			expected: defaultValue,
		},
		"unknown": {
			input:    fwtypes.StringEnumUnknown[awstypes.AclPermission](),
			expected: defaultValue,
		},
		"value": {
			input:    fwtypes.StringEnumValue(awstypes.AclPermissionWrite),
			expected: awstypes.AclPermissionWrite,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.StringEnumValueOr(t.Context(), test.input, defaultValue)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
