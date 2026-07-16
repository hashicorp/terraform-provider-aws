// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"testing"

	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAutonomousDatabaseSchemas(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	rawResource, err := newResourceAutonomousDatabase(ctx)
	if err != nil {
		t.Fatalf("creating resource: %s", err)
	}
	resourceResponse := resource.SchemaResponse{}
	rawResource.Schema(ctx, resource.SchemaRequest{}, &resourceResponse)
	resourceDiagnostics := resourceResponse.Schema.ValidateImplementation(ctx)
	if resourceDiagnostics.HasError() {
		t.Fatalf("validating resource schema: %v", resourceDiagnostics)
	}

	rawDataSource, err := newDataSourceAutonomousDatabase(ctx)
	if err != nil {
		t.Fatalf("creating data source: %s", err)
	}
	dataSourceResponse := datasource.SchemaResponse{}
	rawDataSource.Schema(ctx, datasource.SchemaRequest{}, &dataSourceResponse)
	dataSourceDiagnostics := dataSourceResponse.Schema.ValidateImplementation(ctx)
	if dataSourceDiagnostics.HasError() {
		t.Fatalf("validating data source schema: %v", dataSourceDiagnostics)
	}
}

func TestAutonomousDatabaseScheduledOperationsRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	value := inttypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseScheduledOperationModel{
		{
			DayOfWeek:          inttypes.StringEnumValue(odbtypes.DayOfWeekNameMonday),
			ScheduledStartTime: types.StringValue("08:00"),
			ScheduledStopTime:  types.StringValue("18:00"),
		},
	})

	var diagnostics resource.CreateResponse
	apiObjects := expandAutonomousDatabaseScheduledOperations(ctx, value, &diagnostics.Diagnostics)
	if diagnostics.Diagnostics.HasError() {
		t.Fatalf("expanding scheduled operations: %v", diagnostics.Diagnostics)
	}
	if got, want := len(apiObjects), 1; got != want {
		t.Fatalf("scheduled operations length = %d, want %d", got, want)
	}
	if apiObjects[0].DayOfWeek == nil || apiObjects[0].DayOfWeek.Name != odbtypes.DayOfWeekNameMonday {
		t.Fatalf("day of week = %#v, want MONDAY", apiObjects[0].DayOfWeek)
	}

	var result inttypes.ListNestedObjectValueOf[autonomousDatabaseScheduledOperationModel]
	diags := flattenAutonomousDatabaseScheduledOperations(ctx, apiObjects, &result)
	if diags.HasError() {
		t.Fatalf("flattening scheduled operations: %v", diags)
	}
	models, diags := result.ToSlice(ctx)
	if diags.HasError() {
		t.Fatalf("reading flattened scheduled operations: %v", diags)
	}
	if got, want := len(models), 1; got != want {
		t.Fatalf("flattened scheduled operations length = %d, want %d", got, want)
	}
	if got := models[0].DayOfWeek.ValueEnum(); got != odbtypes.DayOfWeekNameMonday {
		t.Fatalf("flattened day of week = %q, want %q", got, odbtypes.DayOfWeekNameMonday)
	}
}

func TestAutonomousDatabaseUpdateRequired(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	state, diags := inttypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	state.DisplayName = types.StringValue("example")
	state.Tags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMap(ctx, map[string]string{"Environment": "test"}))
	plan := state
	plan.Tags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMap(ctx, map[string]string{"Environment": "updated"}))

	if autonomousDatabaseUpdateRequired(plan, state) {
		t.Fatal("tag-only change must not call UpdateAutonomousDatabase")
	}

	plan.DisplayName = types.StringValue("updated")
	if !autonomousDatabaseUpdateRequired(plan, state) {
		t.Fatal("display_name change must call UpdateAutonomousDatabase")
	}
}
