package validators_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestClusterIdentifierValidator(t *testing.T) {
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
		"valid identifier": {
			val: types.StringValue("valid-cluster-identifier"),
		},
		"begins with number": {
			val: types.StringValue("11-not-valid"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Cluster Identifier: first character must be a letter, got: 11-not-valid`,
				),
			},
		},
		"contains special character": {
			val: types.StringValue("not-valid!-identifier"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Cluster Identifier: must contain only lowercase alphanumeric characters and hyphens, got: not-valid!-identifier`,
				),
			},
		},
		"contains uppercase": {
			val: types.StringValue("Invalid-identifier"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Cluster Identifier: must contain only lowercase alphanumeric characters and hyphens, got: Invalid-identifier`,
				),
			},
		},
		"contains consecutive hyphens": {
			val: types.StringValue("invalid--identifier"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Cluster Identifier: cannot contain two consecutive hyphens, got: invalid--identifier`,
				),
			},
		},
		"ends with hyphen": {
			val: types.StringValue("invalid-identifier-"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Cluster Identifier: cannot end with a hyphen, got: invalid-identifier-`,
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
			fwvalidators.ClusterIdentifier().ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestClusterIdentifierPrefixValidator(t *testing.T) {
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
		"valid identifier": {
			val: types.StringValue("valid-cluster-identifier"),
		},
		"begins with number": {
			val: types.StringValue("11-not-valid"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Cluster Identifier Prefix: first character must be a letter, got: 11-not-valid`,
				),
			},
		},
		"contains special character": {
			val: types.StringValue("not-valid!-identifier"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Cluster Identifier Prefix: must contain only lowercase alphanumeric characters and hyphens, got: not-valid!-identifier`,
				),
			},
		},
		"contains uppercase": {
			val: types.StringValue("InValid-identifier"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Cluster Identifier Prefix: must contain only lowercase alphanumeric characters and hyphens, got: InValid-identifier`,
				),
			},
		},
		"contains consecutive hyphens": {
			val: types.StringValue("invalid--identifier"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Cluster Identifier Prefix: cannot contain two consecutive hyphens, got: invalid--identifier`,
				),
			},
		},
		"ends with hyphen": {
			val: types.StringValue("valid-identifier-"),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.ClusterIdentifierPrefix().ValidateString(context.Background(), request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestClusterFinalSnapshotIdentifierValidator(t *testing.T) {
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
		"valid identifier": {
			val: types.StringValue("valid-cluster-identifier"),
		},
		"begins with number": {
			val: types.StringValue("11-not-valid"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Final Snapshot Identifier: first character must be a letter, got: 11-not-valid`,
				),
			},
		},
		"contains special character": {
			val: types.StringValue("not-valid!-identifier"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Final Snapshot Identifier: must contain only alphanumeric characters and hyphens, got: not-valid!-identifier`,
				),
			},
		},
		"contains uppercase": {
			val: types.StringValue("Valid-identifier"),
		},
		"contains consecutive hyphens": {
			val: types.StringValue("invalid--identifier"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Final Snapshot Identifier: cannot contain two consecutive hyphens, got: invalid--identifier`,
				),
			},
		},
		"ends with hyphens": {
			val: types.StringValue("invalid-identifier-"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid Final Snapshot Identifier: cannot end with a hyphen, got: invalid-identifier-`,
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.ClusterFinalSnapshotIdentifier().ValidateString(context.Background(), request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
