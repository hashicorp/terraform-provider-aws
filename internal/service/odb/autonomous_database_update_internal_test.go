// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAutonomousDatabaseUpdateRemoveBlocks(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"scheduled_operations", "customer_contacts_to_send_to_oci", "long_term_backup_schedule", "resource_pool_summary"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			state, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
			if diags.HasError() {
				t.Fatal(diags)
			}
			state.AutonomousDatabaseID = types.StringValue("adb-test")
			plan := state
			var key, want, apiFields string
			switch name {
			case "scheduled_operations":
				state.ScheduledOperations = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseScheduledOperationModel{{DayOfWeek: fwtypes.StringEnumValue(odbtypes.DayOfWeekNameMonday), ScheduledStartTime: types.StringValue("08:00"), ScheduledStopTime: types.StringValue("18:00")}})
				key, want = "scheduledOperations", "[]"
			case "customer_contacts_to_send_to_oci":
				state.CustomerContactsToSendToOCI = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseCustomerContactModel{{Email: types.StringValue("terraform@example.test")}})
				key, want = "customerContactsToSendToOCI", "[]"
			case "long_term_backup_schedule":
				state.LongTermBackupSchedule = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{{IsDisabled: types.BoolValue(false)}})
				key, want = "longTermBackupSchedule", `{"isDisabled":true}`
				apiFields = `,"longTermBackupSchedule":{"isDisabled":true}`
			case "resource_pool_summary":
				state.ResourcePoolSummary = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseResourcePoolSummaryModel{{IsDisabled: types.BoolValue(false)}})
				key, want = "resourcePoolSummary", `{"isDisabled":true}`
				apiFields = `,"resourcePoolSummary":{"isDisabled":true}`
			}
			result, payload := testAutonomousDatabaseUpdate(t, plan, state, plan, apiFields)
			if got := string(payload[key]); got != want {
				t.Errorf("%s JSON = %s, want %s", key, got, want)
			}
			var length int
			switch name {
			case "scheduled_operations":
				length = result.ScheduledOperations.Length(fwtypes.CollectionLengthUnhandledAsZero)
			case "customer_contacts_to_send_to_oci":
				length = result.CustomerContactsToSendToOCI.Length(fwtypes.CollectionLengthUnhandledAsZero)
			case "long_term_backup_schedule":
				length = result.LongTermBackupSchedule.Length(fwtypes.CollectionLengthUnhandledAsZero)
			case "resource_pool_summary":
				length = result.ResourcePoolSummary.Length(fwtypes.CollectionLengthUnhandledAsZero)
			}
			if length != 0 {
				t.Errorf("removed block %s has %d elements in final state", name, length)
			}
		})
	}
}

func TestAutonomousDatabaseUpdateOmittedBackupSchedule(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	state, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	state.AutonomousDatabaseID = types.StringValue("adb-test")
	state.LongTermBackupSchedule = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{{
		IsDisabled: types.BoolValue(false), RepeatCadence: fwtypes.StringEnumValue(odbtypes.RepeatCadenceWeekly), RetentionPeriodInDays: types.Int32Value(90), TimeOfBackup: timetypes.NewRFC3339Null(),
	}})
	plan := state
	plan.DisplayName = types.StringValue("updated")
	plan.LongTermBackupSchedule = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{{
		IsDisabled: types.BoolValue(true), RepeatCadence: fwtypes.StringEnumUnknown[odbtypes.RepeatCadence](), RetentionPeriodInDays: types.Int32Unknown(), TimeOfBackup: timetypes.NewRFC3339Unknown(),
	}})
	config := plan
	config.LongTermBackupSchedule = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{{
		IsDisabled: types.BoolValue(true), RepeatCadence: fwtypes.StringEnumNull[odbtypes.RepeatCadence](), RetentionPeriodInDays: types.Int32Null(), TimeOfBackup: timetypes.NewRFC3339Null(),
	}})
	result, _ := testAutonomousDatabaseUpdate(t, plan, state, config, "")
	schedule, diags := result.LongTermBackupSchedule.ToPtr(ctx)
	if diags.HasError() || schedule == nil {
		t.Fatalf("reading schedule: %v", diags)
	}
	if !schedule.IsDisabled.Equal(types.BoolValue(true)) {
		t.Errorf("configured disabled value = %v, want true", schedule.IsDisabled)
	}
	if !schedule.RepeatCadence.Equal(fwtypes.StringEnumValue(odbtypes.RepeatCadenceWeekly)) {
		t.Errorf("repeat cadence = %v, want WEEKLY", schedule.RepeatCadence)
	}
	if !schedule.RetentionPeriodInDays.Equal(types.Int32Value(90)) {
		t.Errorf("retention days = %v, want 90", schedule.RetentionPeriodInDays)
	}
	if schedule.TimeOfBackup.IsUnknown() {
		t.Error("backup time must not remain unknown after apply")
	}
}

func TestAutonomousDatabaseUpdateBlocksWithoutRemoval(t *testing.T) {
	t.Parallel()
	for _, unknown := range []bool{false, true} {
		name := "absent blocks"
		if unknown {
			name = "unknown planned blocks"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			state, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
			if diags.HasError() {
				t.Fatal(diags)
			}
			state.AutonomousDatabaseID = types.StringValue("adb-test")
			plan := state
			if unknown {
				state.CustomerContactsToSendToOCI = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseCustomerContactModel{{Email: types.StringValue("terraform@example.test")}})
				state.ScheduledOperations = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseScheduledOperationModel{{DayOfWeek: fwtypes.StringEnumValue(odbtypes.DayOfWeekNameMonday)}})
				state.LongTermBackupSchedule = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{{IsDisabled: types.BoolValue(false)}})
				state.ResourcePoolSummary = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseResourcePoolSummaryModel{{IsDisabled: types.BoolValue(false)}})
				plan.CustomerContactsToSendToOCI = fwtypes.NewListNestedObjectValueOfUnknown[autonomousDatabaseCustomerContactModel](ctx)
				plan.ScheduledOperations = fwtypes.NewListNestedObjectValueOfUnknown[autonomousDatabaseScheduledOperationModel](ctx)
				plan.LongTermBackupSchedule = fwtypes.NewListNestedObjectValueOfUnknown[autonomousDatabaseLongTermBackupScheduleModel](ctx)
				plan.ResourcePoolSummary = fwtypes.NewListNestedObjectValueOfUnknown[autonomousDatabaseResourcePoolSummaryModel](ctx)
			}
			input := expandAutonomousDatabaseUpdateInput(ctx, plan, state, plan, &diags)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if input.CustomerContactsToSendToOCI != nil || input.ScheduledOperations != nil || input.LongTermBackupSchedule != nil || input.ResourcePoolSummary != nil {
				t.Fatalf("blocks without a known removal must be omitted from update: %#v", input)
			}
			if autonomousDatabaseUpdateInputHasChanges(input) {
				t.Fatal("blocks without a known removal must not trigger an API update")
			}
		})
	}
}

func TestAutonomousDatabaseFlattenDisabledBackupSchedule(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	nullSchedule := fwtypes.NewListNestedObjectValueOfNull[autonomousDatabaseLongTermBackupScheduleModel](ctx)
	configuredSchedule := fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{{IsDisabled: types.BoolValue(true)}})
	testCases := map[string]struct {
		before     fwtypes.ListNestedObjectValueOf[autonomousDatabaseLongTermBackupScheduleModel]
		disabled   bool
		wantLength int
	}{
		"absent disabled schedule stays absent":         {before: nullSchedule, disabled: true},
		"configured disabled schedule stays configured": {before: configuredSchedule, disabled: true, wantLength: 1},
		"unknown schedule adopts API value":             {before: fwtypes.NewListNestedObjectValueOfUnknown[autonomousDatabaseLongTermBackupScheduleModel](ctx), disabled: true, wantLength: 1},
		"enabled API schedule remains visible":          {before: nullSchedule, wantLength: 1},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			model, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
			if diags.HasError() {
				t.Fatal(diags)
			}
			model.LongTermBackupSchedule = testCase.before
			flattenAutonomousDatabase(ctx, &odbtypes.AutonomousDatabase{LongTermBackupSchedule: &odbtypes.LongTermBackupSchedule{IsDisabled: aws.Bool(testCase.disabled)}}, &model, &diags)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if got := model.LongTermBackupSchedule.Length(fwtypes.CollectionLengthUnhandledAsZero); got != testCase.wantLength {
				t.Fatalf("schedule length = %d, want %d", got, testCase.wantLength)
			}
			if testCase.wantLength > 0 {
				schedule, diags := model.LongTermBackupSchedule.ToPtr(ctx)
				if diags.HasError() {
					t.Fatal(diags)
				}
				if !schedule.IsDisabled.Equal(types.BoolValue(testCase.disabled)) {
					t.Errorf("is_disabled = %v, want %t", schedule.IsDisabled, testCase.disabled)
				}
			}
		})
	}
}

func testAutonomousDatabaseUpdate(t *testing.T, plan, state, config autonomousDatabaseResourceModel, apiFields string) (autonomousDatabaseResourceModel, map[string]json.RawMessage) {
	t.Helper()
	ctx := t.Context()
	var payload map[string]json.RawMessage
	client := new(conns.AWSClient)
	client.SetHTTPClient(ctx, &http.Client{Transport: autonomousDatabaseSweepTransport(func(request *http.Request) (*http.Response, error) {
		var body string
		switch request.Header.Get("X-Amz-Target") {
		case "Odb.UpdateAutonomousDatabase":
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			body = `{"autonomousDatabaseId":"adb-test"}`
		case "Odb.GetAutonomousDatabase":
			body = `{"autonomousDatabase":{"autonomousDatabaseId":"adb-test","status":"AVAILABLE","displayName":"updated"` + apiFields + `}}`
		default:
			t.Fatalf("unexpected AWS operation: %s", request.Header.Get("X-Amz-Target"))
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/x-amz-json-1.0"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})})
	client.SetServicePackages(ctx, map[string]conns.ServicePackage{names.ODB: &servicePackage{}})
	providerConfig := conns.Config{AccessKey: "test", SecretKey: "test", Region: endpoints.UsEast1RegionID, SkipCredsValidation: true, SkipRequestingAccountId: true, MaxRetries: 0, SharedConfigFiles: []string{}, SharedCredentialsFiles: []string{}}
	client, diagnostics := providerConfig.ConfigureProvider(ctx, client)
	if diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	raw, err := newResourceAutonomousDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var configureResponse resource.ConfigureResponse
	raw.Configure(ctx, resource.ConfigureRequest{ProviderData: client}, &configureResponse)
	if configureResponse.Diagnostics.HasError() {
		t.Fatal(configureResponse.Diagnostics)
	}
	var schemaResponse resource.SchemaResponse
	raw.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	schemaResponse.Schema.Attributes[names.AttrRegion] = resourceattribute.Region()
	timeoutsValue := timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType})}
	plan.Timeouts, state.Timeouts, config.Timeouts = timeoutsValue, timeoutsValue, timeoutsValue
	request := resource.UpdateRequest{Plan: tfsdk.Plan{Schema: schemaResponse.Schema}, State: tfsdk.State{Schema: schemaResponse.Schema}, Config: tfsdk.Config{Schema: schemaResponse.Schema}}
	if diags := request.Plan.Set(ctx, plan); diags.HasError() {
		t.Fatal(diags)
	}
	if diags := request.State.Set(ctx, state); diags.HasError() {
		t.Fatal(diags)
	}
	configPlan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := configPlan.Set(ctx, config); diags.HasError() {
		t.Fatal(diags)
	}
	request.Config.Raw = configPlan.Raw
	response := resource.UpdateResponse{State: tfsdk.State{Schema: schemaResponse.Schema, Raw: request.Plan.Raw}}
	raw.Update(ctx, request, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var result autonomousDatabaseResourceModel
	if diags := response.State.Get(ctx, &result); diags.HasError() {
		t.Fatal(diags)
	}
	return result, payload
}
