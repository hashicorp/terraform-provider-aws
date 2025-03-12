// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table_bucket", name="Table Bucket")
func newResourceTableBucket(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTableBucket{}, nil
}

const (
	resNameTableBucket = "Table Bucket"
)

type resourceTableBucket struct {
	framework.ResourceWithConfigure
}

func (r *resourceTableBucket) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			// TODO: Once Protocol v6 is supported, convert this to a `schema.SingleNestedAttribute` with full schema information
			// Validations needed:
			// * iceberg_unreferenced_file_removal.settings.non_current_days:  int32validator.AtLeast(1)
			// * iceberg_unreferenced_file_removal.settings.unreferenced_days: int32validator.AtLeast(1)
			"maintenance_configuration": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[tableBucketMaintenanceConfigurationModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					stringMustContainLowerCaseLettersNumbersHypens,
					stringMustStartWithLetterOrNumber,
					stringMustEndWithLetterOrNumber,
					validators.PrefixNoneOf(
						"xn--",
						"sthree-",
						"sthree-configurator",
						"amzn-s3-demo-",
					),
					validators.SuffixNoneOf(
						"-s3alias",
						"--ol-s3",
						".mrap",
						"--x-s3",
					),
				},
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceTableBucket) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan resourceTableBucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3tables.CreateTableBucketInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateTableBucket(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Arn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	if !plan.MaintenanceConfiguration.IsUnknown() && !plan.MaintenanceConfiguration.IsNull() {
		mc, d := plan.MaintenanceConfiguration.ToPtr(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !mc.IcebergUnreferencedFileRemovalSettings.IsNull() {
			input := s3tables.PutTableBucketMaintenanceConfigurationInput{
				TableBucketARN: out.Arn,
				Type:           awstypes.TableBucketMaintenanceTypeIcebergUnreferencedFileRemoval,
			}

			value, d := expandTableBucketMaintenanceIcebergUnreferencedFileRemoval(ctx, mc.IcebergUnreferencedFileRemovalSettings)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			input.Value = &value

			_, err := conn.PutTableBucketMaintenanceConfiguration(ctx, &input)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), err),
					err.Error(),
				)
				return
			}
		}
	}

	bucket, err := findTableBucket(ctx, conn, aws.ToString(out.Arn))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), err),
			err.Error(),
		)
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, bucket, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsMaintenanceConfig, err := conn.GetTableBucketMaintenanceConfiguration(ctx, &s3tables.GetTableBucketMaintenanceConfigurationInput{
		TableBucketARN: bucket.Arn,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), err),
			err.Error(),
		)
	}
	maintenanceConfiguration, d := flattenTableBucketMaintenanceConfiguration(ctx, awsMaintenanceConfig)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.MaintenanceConfiguration = maintenanceConfiguration

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTableBucket) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableBucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTableBucket(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionReading, resNameTableBucket, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsMaintenanceConfig, err := conn.GetTableBucketMaintenanceConfiguration(ctx, &s3tables.GetTableBucketMaintenanceConfigurationInput{
		TableBucketARN: state.ARN.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionReading, resNameTableBucket, state.Name.String(), err),
			err.Error(),
		)
	}
	maintenanceConfiguration, d := flattenTableBucketMaintenanceConfiguration(ctx, awsMaintenanceConfig)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.MaintenanceConfiguration = maintenanceConfiguration

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTableBucket) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan resourceTableBucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.MaintenanceConfiguration.Equal(plan.MaintenanceConfiguration) {
		conn := r.Meta().S3TablesClient(ctx)

		mc, d := plan.MaintenanceConfiguration.ToPtr(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !mc.IcebergUnreferencedFileRemovalSettings.IsNull() {
			input := s3tables.PutTableBucketMaintenanceConfigurationInput{
				TableBucketARN: state.ARN.ValueStringPointer(),
				Type:           awstypes.TableBucketMaintenanceTypeIcebergUnreferencedFileRemoval,
			}

			value, d := expandTableBucketMaintenanceIcebergUnreferencedFileRemoval(ctx, mc.IcebergUnreferencedFileRemovalSettings)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			input.Value = &value

			_, err := conn.PutTableBucketMaintenanceConfiguration(ctx, &input)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.S3Tables, create.ErrActionUpdating, resNameTableBucket, plan.Name.String(), err),
					err.Error(),
				)
				return
			}
		}

		awsMaintenanceConfig, err := conn.GetTableBucketMaintenanceConfiguration(ctx, &s3tables.GetTableBucketMaintenanceConfigurationInput{
			TableBucketARN: state.ARN.ValueStringPointer(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.S3Tables, create.ErrActionUpdating, resNameTableBucket, plan.Name.String(), err),
				err.Error(),
			)
		}
		maintenanceConfiguration, d := flattenTableBucketMaintenanceConfiguration(ctx, awsMaintenanceConfig)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.MaintenanceConfiguration = maintenanceConfiguration
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTableBucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableBucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &s3tables.DeleteTableBucketInput{
		TableBucketARN: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteTableBucket(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionDeleting, resNameTableBucket, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTableBucket) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

func findTableBucket(ctx context.Context, conn *s3tables.Client, arn string) (*s3tables.GetTableBucketOutput, error) {
	in := s3tables.GetTableBucketInput{
		TableBucketARN: aws.String(arn),
	}

	out, err := conn.GetTableBucket(ctx, &in)
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

type resourceTableBucketModel struct {
	ARN                      types.String                                                    `tfsdk:"arn"`
	CreatedAt                timetypes.RFC3339                                               `tfsdk:"created_at"`
	MaintenanceConfiguration fwtypes.ObjectValueOf[tableBucketMaintenanceConfigurationModel] `tfsdk:"maintenance_configuration" autoflex:"-"`
	Name                     types.String                                                    `tfsdk:"name"`
	OwnerAccountID           types.String                                                    `tfsdk:"owner_account_id"`
}

type tableBucketMaintenanceConfigurationModel struct {
	IcebergUnreferencedFileRemovalSettings fwtypes.ObjectValueOf[tableBucketMaintenanceConfigurationValueModel[icebergUnreferencedFileRemovalSettingsModel]] `tfsdk:"iceberg_unreferenced_file_removal"`
}

type tableBucketMaintenanceConfigurationValueModel[T any] struct {
	Settings fwtypes.ObjectValueOf[T]                       `tfsdk:"settings"`
	Status   fwtypes.StringEnum[awstypes.MaintenanceStatus] `tfsdk:"status"`
}

type icebergUnreferencedFileRemovalSettingsModel struct {
	NonCurrentDays   types.Int32 `tfsdk:"non_current_days"`
	UnreferencedDays types.Int32 `tfsdk:"unreferenced_days"`
}

func flattenTableBucketMaintenanceConfiguration(ctx context.Context, in *s3tables.GetTableBucketMaintenanceConfigurationOutput) (result fwtypes.ObjectValueOf[tableBucketMaintenanceConfigurationModel], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	unreferencedFileRemovalConfig := in.Configuration[string(awstypes.TableBucketMaintenanceTypeIcebergUnreferencedFileRemoval)]
	unreferencedFileRemovalConfigModel, d := flattenTableBucketMaintenanceIcebergUnreferencedFileRemoval(ctx, &unreferencedFileRemovalConfig)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	model := tableBucketMaintenanceConfigurationModel{
		IcebergUnreferencedFileRemovalSettings: unreferencedFileRemovalConfigModel,
	}

	result, d = fwtypes.NewObjectValueOf(ctx, &model)
	diags.Append(d...)
	return result, diags
}

func expandTableBucketMaintenanceIcebergUnreferencedFileRemoval(ctx context.Context, in fwtypes.ObjectValueOf[tableBucketMaintenanceConfigurationValueModel[icebergUnreferencedFileRemovalSettingsModel]]) (result awstypes.TableBucketMaintenanceConfigurationValue, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	settings, d := expandIcebergUnreferencedFileRemovalSettings(ctx, model.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	result.Settings = settings
	result.Status = model.Status.ValueEnum()

	return result, diags
}

func flattenTableBucketMaintenanceIcebergUnreferencedFileRemoval(ctx context.Context, in *awstypes.TableBucketMaintenanceConfigurationValue) (result fwtypes.ObjectValueOf[tableBucketMaintenanceConfigurationValueModel[icebergUnreferencedFileRemovalSettingsModel]], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	iceberg, d := flattenIcebergUnreferencedFileRemovalSettings(ctx, in.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	model := tableBucketMaintenanceConfigurationValueModel[icebergUnreferencedFileRemovalSettingsModel]{
		Settings: iceberg,
		Status:   fwtypes.StringEnumValue(in.Status),
	}

	result, d = fwtypes.NewObjectValueOf(ctx, &model)
	diags.Append(d...)
	return result, diags
}

func expandIcebergUnreferencedFileRemovalSettings(ctx context.Context, in fwtypes.ObjectValueOf[icebergUnreferencedFileRemovalSettingsModel]) (result *awstypes.TableBucketMaintenanceSettingsMemberIcebergUnreferencedFileRemoval, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	var value awstypes.IcebergUnreferencedFileRemovalSettings

	diags.Append(flex.Expand(ctx, model, &value)...)

	return &awstypes.TableBucketMaintenanceSettingsMemberIcebergUnreferencedFileRemoval{
		Value: value,
	}, diags
}

func flattenIcebergUnreferencedFileRemovalSettings(ctx context.Context, in awstypes.TableBucketMaintenanceSettings) (result fwtypes.ObjectValueOf[icebergUnreferencedFileRemovalSettingsModel], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	switch t := in.(type) {
	case *awstypes.TableBucketMaintenanceSettingsMemberIcebergUnreferencedFileRemoval:
		var model icebergUnreferencedFileRemovalSettingsModel
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
