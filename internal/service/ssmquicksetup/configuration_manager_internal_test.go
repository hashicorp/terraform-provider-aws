// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssmquicksetup

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestRestoreConfigurationDefinitionParameters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	planned := fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*configurationDefinitionModel{
		{
			Type: types.StringValue("AWSQuickSetupType-ResourceExplorer"),
			Parameters: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"SelectedAggregatorRegion": types.StringValue("us-east-1"), //lintignore:AWSAT003
			}),
		},
	})

	// API response includes an AWS-injected "QSForceUpdateParam" key.
	apiResponse := fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*configurationDefinitionModel{
		{
			ID:          types.StringValue("cd-1234567890"),
			Type:        types.StringValue("AWSQuickSetupType-ResourceExplorer"),
			TypeVersion: types.StringValue("1.0"),
			Parameters: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"SelectedAggregatorRegion": types.StringValue("us-east-1"), //lintignore:AWSAT003
				"QSForceUpdateParam":       types.StringValue("abc123"),
			}),
		},
	})

	got, diags := restoreConfigurationDefinitionParameters(ctx, planned, apiResponse)
	if diags.HasError() {
		t.Fatalf("restoreConfigurationDefinitionParameters() unexpected error: %v", diags)
	}

	gotSlice, diags := got.ToSlice(ctx)
	if diags.HasError() {
		t.Fatalf("ToSlice() unexpected error: %v", diags)
	}
	if len(gotSlice) != 1 {
		t.Fatalf("got %d configuration definitions, want 1", len(gotSlice))
	}

	wantParameters := fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
		"SelectedAggregatorRegion": types.StringValue("us-east-1"), //lintignore:AWSAT003
	})
	if !gotSlice[0].Parameters.Equal(wantParameters) {
		t.Errorf("Parameters = %v, want %v (QSForceUpdateParam should have been dropped)", gotSlice[0].Parameters, wantParameters)
	}

	// API-owned fields are preserved.
	if got, want := gotSlice[0].ID.ValueString(), "cd-1234567890"; got != want {
		t.Errorf("ID = %q, want %q", got, want)
	}
	if got, want := gotSlice[0].TypeVersion.ValueString(), "1.0"; got != want {
		t.Errorf("TypeVersion = %q, want %q", got, want)
	}
}

func TestRestoreConfigurationDefinitionParameters_mismatchedLengths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	planned := fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*configurationDefinitionModel{})

	apiResponse := fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*configurationDefinitionModel{
		{
			ID:   types.StringValue("cd-1234567890"),
			Type: types.StringValue("AWSQuickSetupType-ResourceExplorer"),
			Parameters: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"QSForceUpdateParam": types.StringValue("abc123"),
			}),
		},
	})

	got, diags := restoreConfigurationDefinitionParameters(ctx, planned, apiResponse)
	if diags.HasError() {
		t.Fatalf("restoreConfigurationDefinitionParameters() unexpected error: %v", diags)
	}

	gotSlice, diags := got.ToSlice(ctx)
	if diags.HasError() {
		t.Fatalf("ToSlice() unexpected error: %v", diags)
	}
	if len(gotSlice) != 1 {
		t.Fatalf("got %d configuration definitions, want 1", len(gotSlice))
	}

	// No planned element to restore from; API value passes through unchanged.
	if got, want := gotSlice[0].ID.ValueString(), "cd-1234567890"; got != want {
		t.Errorf("ID = %q, want %q", got, want)
	}
}
