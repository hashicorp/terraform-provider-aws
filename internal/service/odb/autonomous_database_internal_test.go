// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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

func TestAutonomousDatabaseUpdateInputChangesOnly(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	state, diags := inttypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	state.AutonomousDatabaseID = types.StringValue("adb-123")
	state.ComputeCount = types.Float64Value(2)
	state.DisplayName = types.StringValue("example")
	plan := state
	plan.ComputeCount = types.Float64Value(4)
	config := plan

	var response resource.UpdateResponse
	input := expandAutonomousDatabaseUpdateInput(ctx, plan, state, config, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		t.Fatalf("expanding update input: %v", response.Diagnostics)
	}
	if got, want := aws.ToString(input.AutonomousDatabaseId), "adb-123"; got != want {
		t.Fatalf("autonomous_database_id = %q, want %q", got, want)
	}
	if got, want := aws.ToFloat64(input.ComputeCount), float64(4); got != want {
		t.Fatalf("compute_count = %g, want %g", got, want)
	}
	if input.DisplayName != nil {
		t.Fatalf("display_name = %q, want nil", aws.ToString(input.DisplayName))
	}
	if !autonomousDatabaseUpdateInputHasChanges(input) {
		t.Fatal("compute_count change must call UpdateAutonomousDatabase")
	}

	input = expandAutonomousDatabaseUpdateInput(ctx, state, state, state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		t.Fatalf("expanding no-op update input: %v", response.Diagnostics)
	}
	if autonomousDatabaseUpdateInputHasChanges(input) {
		t.Fatalf("no-op update input has service changes: %#v", input)
	}
}

func TestAutonomousDatabasePostCreateUpdateInput(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	plan, diags := inttypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	plan.DbName = types.StringValue("TESTDB")
	plan.DisplayName = types.StringValue("example")
	plan.LongTermBackupSchedule = inttypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{
		{
			IsDisabled: types.BoolValue(true),
		},
	})

	var response resource.CreateResponse
	input := expandAutonomousDatabasePostCreateUpdateInput(ctx, "adb-123", plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		t.Fatalf("expanding post-create update input: %v", response.Diagnostics)
	}
	if got, want := aws.ToString(input.AutonomousDatabaseId), "adb-123"; got != want {
		t.Fatalf("autonomous_database_id = %q, want %q", got, want)
	}
	if input.DbName != nil {
		t.Fatalf("db_name = %q, want nil", aws.ToString(input.DbName))
	}
	if input.DisplayName != nil {
		t.Fatalf("display_name = %q, want nil", aws.ToString(input.DisplayName))
	}
	if input.LongTermBackupSchedule == nil {
		t.Fatal("long_term_backup_schedule = nil, want configured schedule")
	}
	if got, want := aws.ToBool(input.LongTermBackupSchedule.IsDisabled), true; got != want {
		t.Fatalf("long_term_backup_schedule.is_disabled = %t, want %t", got, want)
	}
}

func TestAutonomousDatabaseNumericFlatten(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	model, diags := inttypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}

	flattenAutonomousDatabase(ctx, &odbtypes.AutonomousDatabase{
		ByolComputeCountLimit: aws.Int32(2),
		DataStorageSizeInTBs:  aws.Float64(1),
	}, &model, &diags)
	if diags.HasError() {
		t.Fatalf("flattening Autonomous Database: %v", diags)
	}
	if got, want := model.ByolComputeCountLimit.ValueFloat64(), float64(2); got != want {
		t.Fatalf("byol_compute_count_limit = %g, want %g", got, want)
	}
	if got, want := model.DataStorageSizeInTBs.ValueInt32(), int32(1); got != want {
		t.Fatalf("data_storage_size_in_tbs = %d, want %d", got, want)
	}
}

func TestAutonomousDatabaseEncryptionFlatten(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	model, diags := inttypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	model.KMSKeyID = types.StringUnknown()

	flattenAutonomousDatabase(ctx, &odbtypes.AutonomousDatabase{}, &model, &diags)
	if diags.HasError() {
		t.Fatalf("flattening Oracle-managed encryption: %v", diags)
	}
	if !model.KMSKeyID.IsNull() {
		t.Fatalf("kms_key_id = %s, want null", model.KMSKeyID)
	}

	flattenAutonomousDatabase(ctx, &odbtypes.AutonomousDatabase{
		EncryptionSummary: &odbtypes.EncryptionSummary{
			EncryptionKeyProvider: odbtypes.EncryptionKeyProviderAwsKms,
			EncryptionKeyConfiguration: &odbtypes.EncryptionKeyConfigurationMemberAwsEncryptionKey{
				Value: odbtypes.AwsEncryptionKeyConfiguration{KmsKeyId: aws.String("example-key")},
			},
		},
	}, &model, &diags)
	if diags.HasError() {
		t.Fatalf("flattening AWS KMS encryption: %v", diags)
	}
	if got, want := model.KMSKeyID.ValueString(), "example-key"; got != want {
		t.Fatalf("kms_key_id = %q, want %q", got, want)
	}
}

func TestAutonomousDatabaseDefaultBlocksFlatten(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	model, diags := inttypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	model.DbToolsDetails = inttypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseToolModel{})
	model.ResourcePoolSummary = inttypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseResourcePoolSummaryModel{})

	flattenAutonomousDatabase(ctx, &odbtypes.AutonomousDatabase{
		DbToolsDetails: []odbtypes.DatabaseTool{{Name: aws.String("APEX")}},
		ResourcePoolSummary: &odbtypes.ResourcePoolSummary{
			PoolSize: aws.Int32(1),
		},
	}, &model, &diags)
	if diags.HasError() {
		t.Fatalf("flattening Autonomous Database defaults: %v", diags)
	}
	if got := len(model.DbToolsDetails.Elements()); got != 0 {
		t.Fatalf("db_tools_details block count = %d, want 0", got)
	}
	if got := len(model.ResourcePoolSummary.Elements()); got != 0 {
		t.Fatalf("resource_pool_summary block count = %d, want 0", got)
	}
}

func TestAutonomousDatabaseConfiguredBlocksFlatten(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	model, diags := inttypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	model.CustomerContactsToSendToOCI = inttypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseCustomerContactModel{
		{
			Email: types.StringValue("terraform@example.com"),
		},
	})
	model.LongTermBackupSchedule = inttypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{
		{
			IsDisabled: types.BoolValue(true),
		},
	})

	flattenAutonomousDatabase(ctx, &odbtypes.AutonomousDatabase{}, &model, &diags)
	if diags.HasError() {
		t.Fatalf("flattening Autonomous Database: %v", diags)
	}
	if got, want := len(model.CustomerContactsToSendToOCI.Elements()), 1; got != want {
		t.Fatalf("customer_contacts_to_send_to_oci block count = %d, want %d", got, want)
	}
	if got, want := len(model.LongTermBackupSchedule.Elements()), 1; got != want {
		t.Fatalf("long_term_backup_schedule block count = %d, want %d", got, want)
	}
}

func TestAutonomousDatabaseCloneTableSpaceListExpansion(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	tableSpaceList := inttypes.NewListValueOfMust[types.Int32](ctx, []attr.Value{
		types.Int32Value(1),
		types.Int32Value(2),
	})

	t.Run("point in time restore", func(t *testing.T) {
		t.Parallel()

		model := autonomousDatabasePointInTimeRestoreModel{
			CloneTableSpaceList:        tableSpaceList,
			CloneType:                  inttypes.StringEnumValue(odbtypes.CloneTypeFull),
			SourceAutonomousDatabaseId: types.StringValue("adb-source"),
		}
		var apiObject odbtypes.PointInTimeRestoreConfiguration
		diags := fwflex.Expand(ctx, model, &apiObject)
		if diags.HasError() {
			t.Fatalf("expanding point-in-time restore: %v", diags)
		}
		if got, want := apiObject.CloneTableSpaceList, []int32{1, 2}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("clone_table_space_list = %v, want %v", got, want)
		}
	})

	t.Run("restore from backup", func(t *testing.T) {
		t.Parallel()

		model := autonomousDatabaseRestoreFromBackupModel{
			AutonomousDatabaseBackupId: types.StringValue("backup-source"),
			CloneTableSpaceList:        tableSpaceList,
			CloneType:                  inttypes.StringEnumValue(odbtypes.CloneTypeFull),
		}
		var apiObject odbtypes.RestoreFromBackupConfiguration
		diags := fwflex.Expand(ctx, model, &apiObject)
		if diags.HasError() {
			t.Fatalf("expanding restore from backup: %v", diags)
		}
		if got, want := apiObject.CloneTableSpaceList, []int32{1, 2}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("clone_table_space_list = %v, want %v", got, want)
		}
	})
}

func TestWaitAutonomousDatabaseDeletedNotFound(t *testing.T) {
	t.Parallel()

	conn := newTestClient(t, jsonHandler(http.StatusBadRequest, map[string]any{
		"__type":  "ResourceNotFoundException",
		"message": "Autonomous Database not found",
	}))

	if err := waitAutonomousDatabaseDeleted(t.Context(), conn, "adb-does-not-exist", 5*time.Second); err != nil {
		t.Fatalf("waiting for deleted Autonomous Database: %v", err)
	}
}
