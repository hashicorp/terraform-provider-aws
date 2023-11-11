// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDirectoryIDValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}
	tests := map[string]testCase{
		"unknown String": {
			val: types.StringUnknown(),
		},
		"null String": {
			val: types.StringNull(),
		},
		"valid directory ID": {
			val: types.StringValue("d-a3b15b67b8"),
		},
		"invalid 1": {
			val: types.StringValue("a3b15b67b8"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value Match",
					`Attribute test must be a valid Directory Service Directory ID, got: a3b15b67b8`,
				),
			},
		},
		"invalid 2": {
			val: types.StringValue("d-abcdefghij"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value Match",
					`Attribute test must be a valid Directory Service Directory ID, got: d-abcdefghij`,
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			directoryIDValidator.ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestDomainWithTrailingDotValidatorValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}
	tests := map[string]testCase{
		"unknown String": {
			val: types.StringUnknown(),
		},
		"null String": {
			val: types.StringNull(),
		},
		"valid non-terminated": {
			val: types.StringValue("directory.test"),
		},
		"valid terminated": {
			val: types.StringValue("directory.test."),
		},
		"invalid 1": {
			val: types.StringValue("test"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value Match",
					`Attribute test must be a fully qualified domain name and may end with a trailing period, got: test`,
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			domainWithTrailingDotValidator.ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestTrustPasswordValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}
	tests := map[string]testCase{
		"unknown String": {
			val: types.StringUnknown(),
		},
		"null String": {
			val: types.StringNull(),
		},
		"valid password": {
			val: types.StringValue("(Super-)Secret! Passw0rd"),
		},
		"invalid 1": {
			val: types.StringValue("pass\nword"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value Match",
					"Attribute test can contain upper- and lower-case letters, numbers, and punctuation characters, got: pass\nword",
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			trustPasswordValidator.ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
