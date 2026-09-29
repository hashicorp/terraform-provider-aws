// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testAutonomousDatabaseSecretARN  = "arn:aws:secretsmanager:us-east-1:123456789012:secret:admin-password" //lintignore:AWSAT003,AWSAT005
	testAutonomousDatabaseIAMRoleARN = "arn:aws:iam::123456789012:role/SecretManagerReadRole"                //lintignore:AWSAT005
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
	value := fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseScheduledOperationModel{
		{
			DayOfWeek:          fwtypes.StringEnumValue(odbtypes.DayOfWeekNameMonday),
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

	var result fwtypes.ListNestedObjectValueOf[autonomousDatabaseScheduledOperationModel]
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

func TestAutonomousDatabaseAdminPasswordSourceRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	source := testAutonomousDatabaseAdminPasswordSource(ctx, testAutonomousDatabaseSecretARN, testAutonomousDatabaseIAMRoleARN, odbtypes.ExternalIdTypeCompartmentOcid)

	var response resource.CreateResponse
	passwordSource, configuration := expandAutonomousDatabaseAdminPasswordSource(ctx, source, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		t.Fatalf("expanding admin password source: %v", response.Diagnostics)
	}
	if got, want := passwordSource, odbtypes.AdminPasswordSourceCustomerManagedAwsSecret; got != want {
		t.Fatalf("admin password source = %q, want %q", got, want)
	}
	secret, ok := configuration.(*odbtypes.AdminPasswordSourceConfigurationInputMemberCustomerManagedAwsSecret)
	if !ok {
		t.Fatalf("admin password source configuration = %T, want customer-managed AWS secret", configuration)
	}
	if got, want := aws.ToString(secret.Value.SecretId), testAutonomousDatabaseSecretARN; got != want {
		t.Fatalf("secret_id = %q, want %q", got, want)
	}
	if got, want := aws.ToString(secret.Value.IamRoleArn), testAutonomousDatabaseIAMRoleARN; got != want {
		t.Fatalf("iam_role_arn = %q, want %q", got, want)
	}
	if got, want := secret.Value.ExternalIdType, odbtypes.ExternalIdTypeCompartmentOcid; got != want {
		t.Fatalf("external_id_type = %q, want %q", got, want)
	}

	var result fwtypes.ListNestedObjectValueOf[autonomousDatabaseAdminPasswordSourceModel]
	diags := flattenAutonomousDatabaseAdminPasswordSource(ctx, &odbtypes.AdminPasswordSourceSummary{
		AdminPasswordSource: odbtypes.AdminPasswordSourceCustomerManagedAwsSecret,
		AdminPasswordSourceConfiguration: &odbtypes.AdminPasswordSourceConfigurationMemberCustomerManagedAwsSecret{
			Value: odbtypes.CustomerManagedAwsSecretConfiguration{
				ExternalIdType: odbtypes.ExternalIdTypeCompartmentOcid,
				IamRoleArn:     aws.String(testAutonomousDatabaseIAMRoleARN),
				SecretId:       aws.String(testAutonomousDatabaseSecretARN),
			},
		},
	}, &result)
	if diags.HasError() {
		t.Fatalf("flattening admin password source: %v", diags)
	}
	if !source.Equal(result) {
		t.Fatalf("admin password source = %#v, want %#v", result, source)
	}
}

func TestAutonomousDatabaseUpdateInputAdminPasswordSource(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	state, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	state.AutonomousDatabaseID = types.StringValue("adb-123")
	plan := state
	plan.AdminPasswordSource = testAutonomousDatabaseAdminPasswordSource(ctx, testAutonomousDatabaseSecretARN, testAutonomousDatabaseIAMRoleARN, odbtypes.ExternalIdTypeCompartmentOcid)

	var response resource.UpdateResponse
	input := expandAutonomousDatabaseUpdateInput(ctx, plan, state, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		t.Fatalf("expanding update input: %v", response.Diagnostics)
	}
	if got, want := input.AdminPasswordSource, odbtypes.AdminPasswordSourceCustomerManagedAwsSecret; got != want {
		t.Fatalf("admin_password_source = %q, want %q", got, want)
	}
	if _, ok := input.AdminPasswordSourceConfiguration.(*odbtypes.AdminPasswordSourceConfigurationInputMemberCustomerManagedAwsSecret); !ok {
		t.Fatalf("admin_password_source_configuration = %T, want customer-managed AWS secret", input.AdminPasswordSourceConfiguration)
	}

	state.AdminPasswordSource = plan.AdminPasswordSource
	plan.AdminPasswordSource = fwtypes.NewListNestedObjectValueOfNull[autonomousDatabaseAdminPasswordSourceModel](ctx)
	input = expandAutonomousDatabaseUpdateInput(ctx, plan, state, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		t.Fatalf("expanding removal update input: %v", response.Diagnostics)
	}
	if got, want := input.AdminPasswordSource, odbtypes.AdminPasswordSourceApiRequestParameter; got != want {
		t.Fatalf("removal admin_password_source = %q, want %q", got, want)
	}
	if input.AdminPasswordSourceConfiguration != nil {
		t.Fatalf("removal admin_password_source_configuration = %T, want nil", input.AdminPasswordSourceConfiguration)
	}
}

func testAutonomousDatabaseAdminPasswordSource(ctx context.Context, secretARN, iamRoleARN string, externalIDType odbtypes.ExternalIdType) fwtypes.ListNestedObjectValueOf[autonomousDatabaseAdminPasswordSourceModel] {
	return fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseAdminPasswordSourceModel{
		{
			CustomerManagedAWSSecret: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseCustomerManagedAWSSecretModel{
				{
					ExternalIDType: fwtypes.StringEnumValue(externalIDType),
					IAMRoleARN:     types.StringValue(iamRoleARN),
					SecretARN:      types.StringValue(secretARN),
				},
			}),
		},
	})
}

func TestAutonomousDatabaseUpdateInputChangesOnly(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	state, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	state.AutonomousDatabaseID = types.StringValue("adb-123")
	state.AdminPassword = types.StringValue("old-password-123")
	state.ComputeCount = types.Float64Value(2)
	state.DisplayName = types.StringValue("example")
	plan := state
	plan.AdminPassword = types.StringValue("new-password-123")
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
	if got, want := aws.ToString(input.AdminPassword), "new-password-123"; got != want {
		t.Fatalf("admin_password = %q, want %q", got, want)
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
	plan, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	plan.DbName = types.StringValue("TESTDB")
	plan.DisplayName = types.StringValue("example")
	plan.LongTermBackupSchedule = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{
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
	model, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
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
	model, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
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
	model, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	model.DbToolsDetails = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseToolModel{})
	model.ResourcePoolSummary = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseResourcePoolSummaryModel{})

	flattenAutonomousDatabase(ctx, &odbtypes.AutonomousDatabase{
		DbToolsDetails: []odbtypes.DatabaseTool{{Name: aws.String("APEX")}},
		ResourcePoolSummary: &odbtypes.ResourcePoolSummary{
			PoolSize: aws.Int32(1),
		},
	}, &model, &diags)
	if diags.HasError() {
		t.Fatalf("flattening Autonomous Database defaults: %v", diags)
	}
	if got := model.DbToolsDetails.Length(fwtypes.CollectionLengthUnhandledAsZero); got != 0 {
		t.Fatalf("db_tools_details block count = %d, want 0", got)
	}
	if got := model.ResourcePoolSummary.Length(fwtypes.CollectionLengthUnhandledAsZero); got != 0 {
		t.Fatalf("resource_pool_summary block count = %d, want 0", got)
	}
}

func TestAutonomousDatabaseConfiguredBlocksFlatten(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	model, diags := fwtypes.Nullified[autonomousDatabaseResourceModel](ctx)
	if diags.HasError() {
		t.Fatalf("constructing null resource model: %v", diags)
	}
	model.CustomerContactsToSendToOCI = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseCustomerContactModel{
		{
			Email: types.StringValue("terraform@example.test"),
		},
	})
	model.LongTermBackupSchedule = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseLongTermBackupScheduleModel{
		{
			IsDisabled: types.BoolValue(true),
		},
	})

	flattenAutonomousDatabase(ctx, &odbtypes.AutonomousDatabase{}, &model, &diags)
	if diags.HasError() {
		t.Fatalf("flattening Autonomous Database: %v", diags)
	}
	if got, want := model.CustomerContactsToSendToOCI.Length(fwtypes.CollectionLengthUnhandledAsZero), 1; got != want {
		t.Fatalf("customer_contacts_to_send_to_oci block count = %d, want %d", got, want)
	}
	if got, want := model.LongTermBackupSchedule.Length(fwtypes.CollectionLengthUnhandledAsZero), 1; got != want {
		t.Fatalf("long_term_backup_schedule block count = %d, want %d", got, want)
	}
}

func TestAutonomousDatabaseCloneTableSpaceListExpansion(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	tableSpaceList := fwtypes.NewListValueOfMust[types.Int32](ctx, []attr.Value{
		types.Int32Value(1),
		types.Int32Value(2),
	})

	t.Run("point in time restore", func(t *testing.T) {
		t.Parallel()

		model := autonomousDatabasePointInTimeRestoreModel{
			CloneTableSpaceList:        tableSpaceList,
			CloneType:                  fwtypes.StringEnumValue(odbtypes.CloneTypeFull),
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
			CloneType:                  fwtypes.StringEnumValue(odbtypes.CloneTypeFull),
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
		"__type":          "ResourceNotFoundException",
		names.AttrMessage: "Autonomous Database not found",
	}))

	if err := waitAutonomousDatabaseDeleted(t.Context(), conn, "adb-does-not-exist", 5*time.Second); err != nil {
		t.Fatalf("waiting for deleted Autonomous Database: %v", err)
	}
}

func TestAutonomousDatabaseIdentityImport(t *testing.T) {
	t.Parallel()

	const accountID = testAutonomousDatabaseAccountID
	const region = endpoints.UsEast1RegionID
	const databaseID = "adb-test"
	testCases := map[string]struct {
		id         string
		identity   map[string]string
		wantRegion string
		wantError  string
	}{
		"legacy ID":                            {id: databaseID, wantRegion: region},
		"identity defaults":                    {identity: map[string]string{names.AttrID: databaseID}, wantRegion: region},
		"identity explicit account and region": {identity: map[string]string{names.AttrID: databaseID, names.AttrAccountID: accountID, names.AttrRegion: region}, wantRegion: region},
		"identity region override":             {identity: map[string]string{names.AttrID: databaseID, names.AttrRegion: endpoints.UsWest2RegionID}, wantRegion: endpoints.UsWest2RegionID},
		"different account rejected":           {identity: map[string]string{names.AttrID: databaseID, names.AttrAccountID: "111111111111"}, wantError: "account"},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := importer.Context(t.Context(), testAutonomousDatabaseImportClient{accountID: accountID, region: region})
			var registration *inttypes.ServicePackageFrameworkResource
			for _, candidate := range (&servicePackage{}).FrameworkResources(ctx) {
				if candidate.TypeName == "aws_odb_autonomous_database" {
					registration = candidate
					break
				}
			}
			if registration == nil {
				t.Fatal("Autonomous Database registration missing")
			}
			wantIdentity := inttypes.RegionalSingleParameterIdentity(inttypes.StringIdentityAttribute(names.AttrID, true))
			if !reflect.DeepEqual(registration.Identity, wantIdentity) {
				t.Fatalf("identity registration = %#v, want %#v", registration.Identity, wantIdentity)
			}
			raw, err := registration.Factory(ctx)
			if err != nil {
				t.Fatal(err)
			}
			identityResource, ok := raw.(framework.ImportByIdentityer)
			if !ok {
				t.Fatal("Autonomous Database must support identity import")
			}
			identityResource.SetIdentitySpec(registration.Identity)
			identityResource.SetImportSpec(registration.Import)
			var schemaResponse resource.SchemaResponse
			raw.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			schemaResponse.Schema.Attributes[names.AttrRegion] = resourceattribute.Region()
			identitySchema := identity.NewIdentitySchema(registration.Identity)
			resourceIdentity := &tfsdk.ResourceIdentity{Raw: tftypes.NewValue(identitySchema.Type().TerraformType(ctx), nil), Schema: &identitySchema}
			for key, value := range testCase.identity {
				if diags := resourceIdentity.SetAttribute(ctx, path.Root(key), value); diags.HasError() {
					t.Fatal(diags)
				}
			}
			request := resource.ImportStateRequest{ID: testCase.id, Identity: resourceIdentity}
			response := resource.ImportStateResponse{
				State:    tfsdk.State{Raw: tftypes.NewValue(schemaResponse.Schema.Type().TerraformType(ctx), nil), Schema: schemaResponse.Schema},
				Identity: resourceIdentity,
			}
			// The provider's region interceptor seeds this before invoking the resource importer.
			if diags := response.State.SetAttribute(ctx, path.Root(names.AttrRegion), region); diags.HasError() {
				t.Fatal(diags)
			}
			raw.(resource.ResourceWithImportState).ImportState(ctx, request, &response)
			if testCase.wantError != "" {
				if !response.Diagnostics.HasError() || !strings.Contains(strings.ToLower(response.Diagnostics.Errors()[0].Detail()), testCase.wantError) {
					t.Fatalf("diagnostics = %v, want %q error", response.Diagnostics, testCase.wantError)
				}
				return
			}
			if response.Diagnostics.HasError() {
				t.Fatal(response.Diagnostics)
			}
			for key, want := range map[string]string{names.AttrID: databaseID, names.AttrRegion: testCase.wantRegion} {
				var got string
				if diags := response.State.GetAttribute(ctx, path.Root(key), &got); diags.HasError() {
					t.Fatal(diags)
				}
				if got != want {
					t.Errorf("state %s = %q, want %q", key, got, want)
				}
			}
			for key, want := range map[string]string{names.AttrID: databaseID, names.AttrAccountID: accountID, names.AttrRegion: testCase.wantRegion} {
				var got string
				if diags := response.Identity.GetAttribute(ctx, path.Root(key), &got); diags.HasError() {
					t.Fatal(diags)
				}
				if got != want {
					t.Errorf("identity %s = %q, want %q", key, got, want)
				}
			}
		})
	}
}

type testAutonomousDatabaseImportClient struct{ accountID, region string }

func (c testAutonomousDatabaseImportClient) AccountID(context.Context) string { return c.accountID }
func (c testAutonomousDatabaseImportClient) Region(context.Context) string    { return c.region }
