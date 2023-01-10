package stringplanmodifier

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDefaultValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plannedValue  types.String
		currentValue  types.String
		defaultValue  string
		expectedValue types.String
		expectError   bool
	}
	tests := map[string]testCase{
		"non-default non-Null string": {
			plannedValue:  types.StringValue("gamma"),
			currentValue:  types.StringValue("beta"),
			defaultValue:  "alpha",
			expectedValue: types.StringValue("gamma"),
		},
		"non-default non-Null string, current Null": {
			plannedValue:  types.StringValue("gamma"),
			currentValue:  types.StringNull(),
			defaultValue:  "alpha",
			expectedValue: types.StringValue("gamma"),
		},
		"non-default Null string, current Null": {
			plannedValue:  types.StringNull(),
			currentValue:  types.StringValue("beta"),
			defaultValue:  "alpha",
			expectedValue: types.StringValue("alpha"),
		},
		"default string": {
			plannedValue:  types.StringNull(),
			currentValue:  types.StringValue("alpha"),
			defaultValue:  "alpha",
			expectedValue: types.StringValue("alpha"),
		},
		"default string on create": {
			plannedValue:  types.StringNull(),
			currentValue:  types.StringNull(),
			defaultValue:  "alpha",
			expectedValue: types.StringValue("alpha"),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			request := planmodifier.StringRequest{
				Path:       path.Root("test"),
				PlanValue:  test.plannedValue,
				StateValue: test.currentValue,
			}
			response := planmodifier.StringResponse{
				PlanValue: request.PlanValue,
			}
			DefaultValue(test.defaultValue).PlanModifyString(ctx, request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}

			if diff := cmp.Diff(response.PlanValue, test.expectedValue); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
