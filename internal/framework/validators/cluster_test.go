package validators_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestClusterIdentifierValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
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
			val:         types.StringValue("11-not-valid"),
			expectError: true,
		},
		"contains special character": {
			val:         types.StringValue("not-valid!-identifier"),
			expectError: true,
		},
		"contains uppercase": {
			val:         types.StringValue("Invalid-identifier"),
			expectError: true,
		},
		"contains consecutive hyphens": {
			val:         types.StringValue("invalid--identifier"),
			expectError: true,
		},
		"ends with hyphens": {
			val:         types.StringValue("invalid-identifier--"),
			expectError: true,
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
			fwvalidators.ClusterIdentifier().ValidateString(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

func TestClusterIdentifierPrefixValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
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
		"contains special character": {
			val:         types.StringValue("not-valid!-identifier"),
			expectError: true,
		},
		"contains uppercase": {
			val:         types.StringValue("InValid-identifier"),
			expectError: true,
		},
		"contains consecutive hyphens": {
			val:         types.StringValue("invalid--identifier"),
			expectError: true,
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

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

func TestClusterFinalSnapshotIdentifierValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
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
		},
		"contains special character": {
			val:         types.StringValue("not-valid!-identifier"),
			expectError: true,
		},
		"contains uppercase": {
			val: types.StringValue("Valid-identifier"),
		},
		"contains consecutive hyphens": {
			val:         types.StringValue("invalid--identifier"),
			expectError: true,
		},
		"ends with hyphens": {
			val:         types.StringValue("invalid-identifier-"),
			expectError: true,
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

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
