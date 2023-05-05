package boolplanmodifier

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
		plannedValue  types.Bool
		currentValue  types.Bool
		defaultValue  bool
		expectedValue types.Bool
		expectError   bool
	}
	tests := map[string]testCase{
		"default bool": {
			plannedValue:  types.BoolNull(),
			currentValue:  types.BoolValue(true),
			defaultValue:  true,
			expectedValue: types.BoolValue(true),
		},
		"default bool on create": {
			plannedValue:  types.BoolNull(),
			currentValue:  types.BoolNull(),
			defaultValue:  true,
			expectedValue: types.BoolValue(true),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			request := planmodifier.BoolRequest{
				Path:       path.Root("test"),
				PlanValue:  test.plannedValue,
				StateValue: test.currentValue,
			}
			response := planmodifier.BoolResponse{
				PlanValue: request.PlanValue,
			}
			DefaultValue(test.defaultValue).PlanModifyBool(ctx, request, &response)

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
