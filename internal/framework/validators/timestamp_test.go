package validators_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestUTCTimestampValidator(t *testing.T) {
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
		"valid timestamp": {
			val: types.StringValue("2023-02-28T14:04:05Z"),
		},
		"invalid timestamp": {
			val:         types.StringValue("02 Jan 06 15:04 -0700"),
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
			fwvalidators.UTCTimestamp().ValidateString(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
