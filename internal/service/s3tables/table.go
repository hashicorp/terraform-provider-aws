// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table", name="Table")
func newResourceTable(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTable{}, nil
}

const (
	ResNameTable = "Table"
)

type resourceTable struct {
	framework.ResourceWithConfigure
}

func (r *resourceTable) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrFormat: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OpenTableFormat](),
				Required:   true,
				// TODO: Only one format is currently supported. When a new value is added, we can determine if `format` can be changed in-place or must recreate the resource
			},
			// TODO: Once Protocol v6 is supported, convert this to a `schema.SingleNestedAttribute` with full schema information
			// Validations needed:
			// * iceberg_compaction.settings.target_file_size_mb:  int32validator.Between(64, 512)
			// * iceberg_snapshot_management.settings.max_snapshot_age_hours: int32validator.AtLeast(1)
			// * iceberg_snapshot_management.settings.min_snapshots_to_keep:  int32validator.AtLeast(1)
			"maintenance_configuration": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[tableMaintenanceConfigurationModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata_location": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"modified_by": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required:   true,
				Validators: tableNameValidator,
			},
			names.AttrNamespace: schema.StringAttribute{
				Required:   true,
				Validators: namespaceNameValidator,
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"table_bucket_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TableType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version_token": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"warehouse_location": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceTable) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan resourceTableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3tables.CreateTableInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateTable(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTable, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	if !plan.MaintenanceConfiguration.IsUnknown() && !plan.MaintenanceConfiguration.IsNull() {
		mc, d := plan.MaintenanceConfiguration.ToPtr(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !mc.IcebergCompaction.IsNull() {
			input := s3tables.PutTableMaintenanceConfigurationInput{
				Name:           plan.Name.ValueStringPointer(),
				Namespace:      plan.Namespace.ValueStringPointer(),
				TableBucketARN: plan.TableBucketARN.ValueStringPointer(),
				Type:           awstypes.TableMaintenanceTypeIcebergCompaction,
			}

			value, d := expandTableMaintenanceIcebergCompaction(ctx, mc.IcebergCompaction)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableMaintenanceConfiguration(ctx, &input)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), err),
					err.Error(),
				)
				return
			}
		}

		if !mc.IcebergSnapshotManagement.IsNull() {
			input := s3tables.PutTableMaintenanceConfigurationInput{
				Name:           plan.Name.ValueStringPointer(),
				Namespace:      plan.Namespace.ValueStringPointer(),
				TableBucketARN: plan.TableBucketARN.ValueStringPointer(),
				Type:           awstypes.TableMaintenanceTypeIcebergSnapshotManagement,
			}

			value, d := expandTableMaintenanceIcebergSnapshotManagement(ctx, mc.IcebergSnapshotManagement)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableMaintenanceConfiguration(ctx, &input)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), err),
					err.Error(),
				)
				return
			}
		}
	}

	table, err := findTable(ctx, conn, plan.TableBucketARN.ValueString(), plan.Namespace.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTable, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, table, &plan, flex.WithFieldNamePrefix("Table"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Namespace = types.StringValue(table.Namespace[0])

	awsMaintenanceConfig, err := conn.GetTableMaintenanceConfiguration(ctx, &s3tables.GetTableMaintenanceConfigurationInput{
		Name:           plan.Name.ValueStringPointer(),
		Namespace:      plan.Namespace.ValueStringPointer(),
		TableBucketARN: plan.TableBucketARN.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), err),
			err.Error(),
		)
	}
	maintenanceConfiguration, d := flattenTableMaintenanceConfiguration(ctx, awsMaintenanceConfig)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.MaintenanceConfiguration = maintenanceConfiguration

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTable) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTable(ctx, conn, state.TableBucketARN.ValueString(), state.Namespace.ValueString(), state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionReading, ResNameTable, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Table"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Namespace = types.StringValue(out.Namespace[0])

	awsMaintenanceConfig, err := conn.GetTableMaintenanceConfiguration(ctx, &s3tables.GetTableMaintenanceConfigurationInput{
		Name:           state.Name.ValueStringPointer(),
		Namespace:      state.Namespace.ValueStringPointer(),
		TableBucketARN: state.TableBucketARN.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionReading, resNameTableBucket, state.Name.String(), err),
			err.Error(),
		)
	}
	maintenanceConfiguration, d := flattenTableMaintenanceConfiguration(ctx, awsMaintenanceConfig)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.MaintenanceConfiguration = maintenanceConfiguration

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTable) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan, state resourceTableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) || !plan.Namespace.Equal(state.Namespace) {
		input := s3tables.RenameTableInput{
			TableBucketARN: state.TableBucketARN.ValueStringPointer(),
			Namespace:      state.Namespace.ValueStringPointer(),
			Name:           state.Name.ValueStringPointer(),
		}

		if !plan.Name.Equal(state.Name) {
			input.NewName = plan.Name.ValueStringPointer()
		}

		if !plan.Namespace.Equal(state.Namespace) {
			input.NewNamespaceName = plan.Namespace.ValueStringPointer()
		}

		_, err := conn.RenameTable(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.S3Tables, create.ErrActionUpdating, ResNameTable, state.Name.String(), err),
				err.Error(),
			)
		}
	}

	if !plan.MaintenanceConfiguration.Equal(state.MaintenanceConfiguration) {
		planMC, d := plan.MaintenanceConfiguration.ToPtr(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		stateMC, d := state.MaintenanceConfiguration.ToPtr(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !planMC.IcebergCompaction.Equal(stateMC.IcebergCompaction) {
			input := s3tables.PutTableMaintenanceConfigurationInput{
				Name:           plan.Name.ValueStringPointer(),
				Namespace:      plan.Namespace.ValueStringPointer(),
				TableBucketARN: plan.TableBucketARN.ValueStringPointer(),
				Type:           awstypes.TableMaintenanceTypeIcebergCompaction,
			}

			value, d := expandTableMaintenanceIcebergCompaction(ctx, planMC.IcebergCompaction)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableMaintenanceConfiguration(ctx, &input)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.S3Tables, create.ErrActionUpdating, resNameTableBucket, plan.Name.String(), err),
					err.Error(),
				)
				return
			}
		}

		if !planMC.IcebergSnapshotManagement.Equal(stateMC.IcebergSnapshotManagement) {
			input := s3tables.PutTableMaintenanceConfigurationInput{
				Name:           plan.Name.ValueStringPointer(),
				Namespace:      plan.Namespace.ValueStringPointer(),
				TableBucketARN: plan.TableBucketARN.ValueStringPointer(),
				Type:           awstypes.TableMaintenanceTypeIcebergSnapshotManagement,
			}

			value, d := expandTableMaintenanceIcebergSnapshotManagement(ctx, planMC.IcebergSnapshotManagement)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableMaintenanceConfiguration(ctx, &input)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.S3Tables, create.ErrActionUpdating, resNameTableBucket, plan.Name.String(), err),
					err.Error(),
				)
				return
			}
		}
	}

	table, err := findTable(ctx, conn, plan.TableBucketARN.ValueString(), plan.Namespace.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionUpdating, ResNameTable, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, table, &plan, flex.WithFieldNamePrefix("Table"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Namespace = types.StringValue(table.Namespace[0])

	awsMaintenanceConfig, err := conn.GetTableMaintenanceConfiguration(ctx, &s3tables.GetTableMaintenanceConfigurationInput{
		Name:           plan.Name.ValueStringPointer(),
		Namespace:      plan.Namespace.ValueStringPointer(),
		TableBucketARN: plan.TableBucketARN.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionUpdating, resNameTableBucket, plan.Name.String(), err),
			err.Error(),
		)
	}
	maintenanceConfiguration, d := flattenTableMaintenanceConfiguration(ctx, awsMaintenanceConfig)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.MaintenanceConfiguration = maintenanceConfiguration

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTable) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := s3tables.DeleteTableInput{
		Name:           state.Name.ValueStringPointer(),
		Namespace:      state.Namespace.ValueStringPointer(),
		TableBucketARN: state.TableBucketARN.ValueStringPointer(),
	}

	_, err := conn.DeleteTable(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionDeleting, ResNameTable, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTable) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	identifier, err := parseTableIdentifier(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import IDs for S3 Tables Tables must use the format <table bucket ARN>"+tableIDSeparator+"<namespace>"+tableIDSeparator+"<table name>.\n"+
				fmt.Sprintf("Had %q", req.ID),
		)
		return
	}

	identifier.PopulateState(ctx, &resp.State, &resp.Diagnostics)
}

func findTable(ctx context.Context, conn *s3tables.Client, bucketARN, namespace, name string) (*s3tables.GetTableOutput, error) {
	in := s3tables.GetTableInput{
		Name:           aws.String(name),
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(bucketARN),
	}

	out, err := conn.GetTable(ctx, &in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceTableModel struct {
	ARN                      types.String                                              `tfsdk:"arn"`
	CreatedAt                timetypes.RFC3339                                         `tfsdk:"created_at"`
	CreatedBy                types.String                                              `tfsdk:"created_by"`
	Format                   fwtypes.StringEnum[awstypes.OpenTableFormat]              `tfsdk:"format"`
	MaintenanceConfiguration fwtypes.ObjectValueOf[tableMaintenanceConfigurationModel] `tfsdk:"maintenance_configuration" autoflex:"-"`
	MetadataLocation         types.String                                              `tfsdk:"metadata_location"`
	ModifiedAt               timetypes.RFC3339                                         `tfsdk:"modified_at"`
	ModifiedBy               types.String                                              `tfsdk:"modified_by"`
	Name                     types.String                                              `tfsdk:"name"`
	Namespace                types.String                                              `tfsdk:"namespace" autoflex:",noflatten"` // On read, Namespace is an array
	OwnerAccountID           types.String                                              `tfsdk:"owner_account_id"`
	TableBucketARN           fwtypes.ARN                                               `tfsdk:"table_bucket_arn"`
	Type                     fwtypes.StringEnum[awstypes.TableType]                    `tfsdk:"type"`
	VersionToken             types.String                                              `tfsdk:"version_token"`
	WarehouseLocation        types.String                                              `tfsdk:"warehouse_location"`
}

type tableMaintenanceConfigurationModel struct {
	IcebergCompaction         fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergCompactionSettingsModel]]         `tfsdk:"iceberg_compaction"`
	IcebergSnapshotManagement fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergSnapshotManagementSettingsModel]] `tfsdk:"iceberg_snapshot_management"`
}

type tableMaintenanceConfigurationValueModel[T any] struct {
	Settings fwtypes.ObjectValueOf[T]                       `tfsdk:"settings"`
	Status   fwtypes.StringEnum[awstypes.MaintenanceStatus] `tfsdk:"status"`
}

type icebergCompactionSettingsModel struct {
	TargetFileSizeMB types.Int32 `tfsdk:"target_file_size_mb"`
}

type icebergSnapshotManagementSettingsModel struct {
	MaxSnapshotAgeHours types.Int32 `tfsdk:"max_snapshot_age_hours"`
	MinSnapshotsToKeep  types.Int32 `tfsdk:"min_snapshots_to_keep"`
}

func flattenTableMaintenanceConfiguration(ctx context.Context, in *s3tables.GetTableMaintenanceConfigurationOutput) (result fwtypes.ObjectValueOf[tableMaintenanceConfigurationModel], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	compactionConfig := in.Configuration[string(awstypes.TableMaintenanceTypeIcebergCompaction)]
	compactionConfigModel, d := flattenTableMaintenanceIcebergCompaction(ctx, &compactionConfig)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	snapshotManagementConfig := in.Configuration[string(awstypes.TableMaintenanceTypeIcebergSnapshotManagement)]
	snapshotManagementConfigModel, d := flattenTableMaintenanceIcebergSnapshotManagement(ctx, &snapshotManagementConfig)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	model := tableMaintenanceConfigurationModel{
		IcebergCompaction:         compactionConfigModel,
		IcebergSnapshotManagement: snapshotManagementConfigModel,
	}

	result, d = fwtypes.NewObjectValueOf(ctx, &model)
	diags.Append(d...)
	return result, diags
}

func expandTableMaintenanceIcebergCompaction(ctx context.Context, in fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergCompactionSettingsModel]]) (result awstypes.TableMaintenanceConfigurationValue, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	settings, d := expandIcebergCompactionSettings(ctx, model.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	result.Settings = settings
	result.Status = model.Status.ValueEnum()

	return result, diags
}

func flattenTableMaintenanceIcebergCompaction(ctx context.Context, in *awstypes.TableMaintenanceConfigurationValue) (result fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergCompactionSettingsModel]], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	iceberg, d := flattenIcebergCompactionSettings(ctx, in.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	model := tableMaintenanceConfigurationValueModel[icebergCompactionSettingsModel]{
		Settings: iceberg,
		Status:   fwtypes.StringEnumValue(in.Status),
	}

	result, d = fwtypes.NewObjectValueOf(ctx, &model)
	diags.Append(d...)
	return result, diags
}

func expandIcebergCompactionSettings(ctx context.Context, in fwtypes.ObjectValueOf[icebergCompactionSettingsModel]) (result *awstypes.TableMaintenanceSettingsMemberIcebergCompaction, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	var value awstypes.IcebergCompactionSettings

	diags.Append(flex.Expand(ctx, model, &value)...)

	return &awstypes.TableMaintenanceSettingsMemberIcebergCompaction{
		Value: value,
	}, diags
}

func flattenIcebergCompactionSettings(ctx context.Context, in awstypes.TableMaintenanceSettings) (result fwtypes.ObjectValueOf[icebergCompactionSettingsModel], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	switch t := in.(type) {
	case *awstypes.TableMaintenanceSettingsMemberIcebergCompaction:
		var model icebergCompactionSettingsModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		result = fwtypes.NewObjectValueOfMust(ctx, &model)

	case *awstypes.UnknownUnionMember:
		tflog.Warn(ctx, "Unexpected tagged union member", map[string]any{
			"tag": t.Tag,
		})

	default:
		tflog.Warn(ctx, "Unexpected nil tagged union value")
	}
	return result, diags
}

func expandTableMaintenanceIcebergSnapshotManagement(ctx context.Context, in fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergSnapshotManagementSettingsModel]]) (result awstypes.TableMaintenanceConfigurationValue, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	settings, d := expandIcebergSnapshotManagementSettings(ctx, model.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	result.Settings = settings
	result.Status = model.Status.ValueEnum()

	return result, diags
}

func flattenTableMaintenanceIcebergSnapshotManagement(ctx context.Context, in *awstypes.TableMaintenanceConfigurationValue) (result fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergSnapshotManagementSettingsModel]], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	iceberg, d := flattenIcebergSnapshotManagementSettings(ctx, in.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	model := tableMaintenanceConfigurationValueModel[icebergSnapshotManagementSettingsModel]{
		Settings: iceberg,
		Status:   fwtypes.StringEnumValue(in.Status),
	}

	result, d = fwtypes.NewObjectValueOf(ctx, &model)
	diags.Append(d...)
	return result, diags
}

func expandIcebergSnapshotManagementSettings(ctx context.Context, in fwtypes.ObjectValueOf[icebergSnapshotManagementSettingsModel]) (result *awstypes.TableMaintenanceSettingsMemberIcebergSnapshotManagement, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	var value awstypes.IcebergSnapshotManagementSettings

	diags.Append(flex.Expand(ctx, model, &value)...)

	return &awstypes.TableMaintenanceSettingsMemberIcebergSnapshotManagement{
		Value: value,
	}, diags
}

func flattenIcebergSnapshotManagementSettings(ctx context.Context, in awstypes.TableMaintenanceSettings) (result fwtypes.ObjectValueOf[icebergSnapshotManagementSettingsModel], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	switch t := in.(type) {
	case *awstypes.TableMaintenanceSettingsMemberIcebergSnapshotManagement:
		var model icebergSnapshotManagementSettingsModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		result = fwtypes.NewObjectValueOfMust(ctx, &model)

	case *awstypes.UnknownUnionMember:
		tflog.Warn(ctx, "Unexpected tagged union member", map[string]any{
			"tag": t.Tag,
		})

	default:
		tflog.Warn(ctx, "Unexpected nil tagged union value")
	}
	return result, diags
}

func tableIDFromTableARN(s string) (string, error) {
	arn, err := arn.Parse(s)
	if err != nil {
		return "", err
	}

	return tableIDFromTableARNResource(arn.Resource), nil
}

func tableIDFromTableARNResource(s string) string {
	parts := strings.SplitN(s, "/", 4)
	return parts[3]
}

type tableIdentifier struct {
	TableBucketARN string
	Namespace      string
	Name           string
}

const (
	tableIDSeparator = ";"
	tableIDParts     = 3
)

func parseTableIdentifier(s string) (tableIdentifier, error) {
	parts := strings.Split(s, tableIDSeparator)
	if len(parts) != tableIDParts {
		return tableIdentifier{}, errors.New("not enough parts")
	}
	for i := range tableIDParts {
		if parts[i] == "" {
			return tableIdentifier{}, errors.New("empty part")
		}
	}

	return tableIdentifier{
		TableBucketARN: parts[0],
		Namespace:      parts[1],
		Name:           parts[2],
	}, nil
}

func (id tableIdentifier) String() string {
	return id.TableBucketARN + tableIDSeparator +
		id.Namespace + tableIDSeparator +
		id.Name
}

func (id tableIdentifier) PopulateState(ctx context.Context, s *tfsdk.State, diags *diag.Diagnostics) {
	diags.Append(s.SetAttribute(ctx, path.Root("table_bucket_arn"), id.TableBucketARN)...)
	diags.Append(s.SetAttribute(ctx, path.Root(names.AttrNamespace), id.Namespace)...)
	diags.Append(s.SetAttribute(ctx, path.Root(names.AttrName), id.Name)...)
}

var tableNameValidator = []validator.String{
	stringvalidator.LengthBetween(1, 255),
	stringMustContainLowerCaseLettersNumbersUnderscores,
	stringMustStartWithLetterOrNumber,
	stringMustEndWithLetterOrNumber,
}
