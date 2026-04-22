// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3tables

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table_bucket", name="Table Bucket")
// @ArnIdentity
// @ArnFormat("bucket/{name}")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3tables;s3tables.GetTableBucketOutput")
// @Testing(importStateIdAttribute="arn")
// @Testing(importStateIdFunc="testAccTableBucketImportStateIdFunc")
// @Testing(preCheck="testAccPreCheck")
// @Testing(preIdentityVersion="6.19.0")
func newTableBucketResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &tableBucketResource{}, nil
}

type tableBucketResource struct {
	framework.ResourceWithModel[tableBucketResourceModel]
	framework.WithImportByIdentity
}

func (r *tableBucketResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEncryptionConfiguration: schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[encryptionConfigurationModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrForceDestroy: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
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
					tfstringvalidator.ContainsOnlyLowerCaseLettersNumbersHyphens,
					tfstringvalidator.StartsWithLetterOrNumber,
					tfstringvalidator.EndsWithLetterOrNumber,
					tfstringvalidator.PrefixNoneOf(
						"xn--",
						"sthree-",
						"sthree-configurator",
						"amzn-s3-demo-",
					),
					tfstringvalidator.SuffixNoneOf(
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
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *tableBucketResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tableBucketResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input s3tables.CreateTableBucketInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	outputCTB, err := conn.CreateTableBucket(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Tables Table Bucket (%s)", name), err.Error())

		return
	}

	tableBucketARN := aws.ToString(outputCTB.Arn)
	if !data.MaintenanceConfiguration.IsUnknown() && !data.MaintenanceConfiguration.IsNull() {
		mc, diags := data.MaintenanceConfiguration.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		if !mc.IcebergUnreferencedFileRemovalSettings.IsNull() {
			typ := awstypes.TableBucketMaintenanceTypeIcebergUnreferencedFileRemoval
			input := s3tables.PutTableBucketMaintenanceConfigurationInput{
				TableBucketARN: aws.String(tableBucketARN),
				Type:           typ,
			}

			value, diags := expandTableBucketMaintenanceIcebergUnreferencedFileRemoval(ctx, mc.IcebergUnreferencedFileRemovalSettings)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableBucketMaintenanceConfiguration(ctx, &input)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("putting S3 Tables Table Bucket (%s) maintenance configuration (%s)", name, typ), err.Error())

				return
			}
		}
	}

	outputGTB, err := findTableBucketByARN(ctx, conn, tableBucketARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGTB, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	outputGTBMC, err := findTableBucketMaintenanceConfigurationByARN(ctx, conn, tableBucketARN)

	switch {
	case retry.NotFound(err):
	case err != nil:
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket (%s) maintenance configuration", name), err.Error())

		return
	default:
		value, diags := flattenTableBucketMaintenanceConfiguration(ctx, outputGTBMC)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		data.MaintenanceConfiguration = value
	}

	awsEncryptionConfig, err := findTableBucketEncryptionConfigurationByARN(ctx, conn, tableBucketARN)

	switch {
	case retry.NotFound(err):
	case err != nil:
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket (%s) encryption", name), err.Error())

		return
	default:
		var encryptionConfiguration encryptionConfigurationModel
		response.Diagnostics.Append(fwflex.Flatten(ctx, awsEncryptionConfig, &encryptionConfiguration)...)
		if response.Diagnostics.HasError() {
			return
		}
		var diags diag.Diagnostics
		data.EncryptionConfiguration, diags = fwtypes.NewObjectValueOf(ctx, &encryptionConfiguration)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *tableBucketResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tableBucketResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Name), fwflex.StringValueFromFramework(ctx, data.ARN)
	outputGTB, err := findTableBucketByARN(ctx, conn, tableBucketARN)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGTB, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	outputGTBMC, err := findTableBucketMaintenanceConfigurationByARN(ctx, conn, tableBucketARN)

	switch {
	case retry.NotFound(err):
	case err != nil:
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket (%s) maintenance configuration", name), err.Error())

		return
	default:
		value, diags := flattenTableBucketMaintenanceConfiguration(ctx, outputGTBMC)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		data.MaintenanceConfiguration = value
	}

	awsEncryptionConfig, err := findTableBucketEncryptionConfigurationByARN(ctx, conn, tableBucketARN)

	switch {
	case retry.NotFound(err):
	case err != nil:
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket (%s) encryption", name), err.Error())

		return
	default:
		var encryptionConfiguration encryptionConfigurationModel
		response.Diagnostics.Append(fwflex.Flatten(ctx, awsEncryptionConfig, &encryptionConfiguration)...)
		if response.Diagnostics.HasError() {
			return
		}
		var diags diag.Diagnostics
		data.EncryptionConfiguration, diags = fwtypes.NewObjectValueOf(ctx, &encryptionConfiguration)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *tableBucketResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new tableBucketResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name, tableBucketARN := fwflex.StringValueFromFramework(ctx, new.Name), fwflex.StringValueFromFramework(ctx, new.ARN)

	if !new.EncryptionConfiguration.Equal(old.EncryptionConfiguration) {
		ec, diags := new.EncryptionConfiguration.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		input := s3tables.PutTableBucketEncryptionInput{
			TableBucketARN: aws.String(tableBucketARN),
		}

		var encryptionConfiguration awstypes.EncryptionConfiguration
		response.Diagnostics.Append(fwflex.Expand(ctx, ec, &encryptionConfiguration)...)
		if response.Diagnostics.HasError() {
			return
		}
		input.EncryptionConfiguration = &encryptionConfiguration

		_, err := conn.PutTableBucketEncryption(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("putting S3 Tables Table Bucket (%s) encryption configuration", name), err.Error())

			return
		}
	}

	if !old.MaintenanceConfiguration.Equal(new.MaintenanceConfiguration) {
		mc, d := new.MaintenanceConfiguration.ToPtr(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}

		if !mc.IcebergUnreferencedFileRemovalSettings.IsNull() {
			typ := awstypes.TableBucketMaintenanceTypeIcebergUnreferencedFileRemoval
			input := s3tables.PutTableBucketMaintenanceConfigurationInput{
				TableBucketARN: aws.String(tableBucketARN),
				Type:           typ,
			}

			value, d := expandTableBucketMaintenanceIcebergUnreferencedFileRemoval(ctx, mc.IcebergUnreferencedFileRemovalSettings)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableBucketMaintenanceConfiguration(ctx, &input)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("putting S3 Tables Table Bucket (%s) maintenance configuration (%s)", name, typ), err.Error())

				return
			}
		}

		outputGTBMC, err := findTableBucketMaintenanceConfigurationByARN(ctx, conn, tableBucketARN)

		switch {
		case retry.NotFound(err):
		case err != nil:
			response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket (%s) maintenance configuration", name), err.Error())

			return
		default:
			value, d := flattenTableBucketMaintenanceConfiguration(ctx, outputGTBMC)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
			new.MaintenanceConfiguration = value
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *tableBucketResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tableBucketResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Name), fwflex.StringValueFromFramework(ctx, data.ARN)
	input := s3tables.DeleteTableBucketInput{
		TableBucketARN: aws.String(tableBucketARN),
	}
	_, err := conn.DeleteTableBucket(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	// If deletion fails due to bucket not being empty and force_destroy is enabled.
	if err != nil && data.ForceDestroy.ValueBool() {
		// Check if the error indicates the bucket is not empty.
		if errs.IsA[*awstypes.ConflictException](err) || errs.IsA[*awstypes.BadRequestException](err) {
			tflog.Debug(ctx, "Table bucket not empty, attempting to empty it", map[string]any{
				"table_bucket_arn": data.ARN.ValueString(),
			})

			// Empty the table bucket by deleting all tables and namespaces.
			if err := emptyTableBucket(ctx, conn, tableBucketARN); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Tables Table Bucket (%s) (force_destroy = true)", name), err.Error())

				return
			}

			// Retry deletion after emptying.
			_, err = conn.DeleteTableBucket(ctx, &input)
		}
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Tables Table Bucket (%s)", name), err.Error())

		return
	}
}

func (r *tableBucketResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	r.WithImportByIdentity.ImportState(ctx, request, response)

	// Set force_destroy to false on import to prevent accidental deletion
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrForceDestroy), types.BoolValue(false))...)
}

func findTableBucketByARN(ctx context.Context, conn *s3tables.Client, arn string) (*s3tables.GetTableBucketOutput, error) {
	input := s3tables.GetTableBucketInput{
		TableBucketARN: aws.String(arn),
	}

	return findTableBucket(ctx, conn, &input)
}

func findTableBucket(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableBucketInput) (*s3tables.GetTableBucketOutput, error) {
	output, err := conn.GetTableBucket(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findTableBucketEncryptionConfigurationByARN(ctx context.Context, conn *s3tables.Client, arn string) (*awstypes.EncryptionConfiguration, error) {
	input := s3tables.GetTableBucketEncryptionInput{
		TableBucketARN: aws.String(arn),
	}

	return findTableBucketEncryptionConfiguration(ctx, conn, &input)
}

func findTableBucketEncryptionConfiguration(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableBucketEncryptionInput) (*awstypes.EncryptionConfiguration, error) {
	output, err := conn.GetTableBucketEncryption(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.EncryptionConfiguration == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.EncryptionConfiguration, nil
}

func findTableBucketMaintenanceConfigurationByARN(ctx context.Context, conn *s3tables.Client, arn string) (*s3tables.GetTableBucketMaintenanceConfigurationOutput, error) {
	input := s3tables.GetTableBucketMaintenanceConfigurationInput{
		TableBucketARN: aws.String(arn),
	}

	return findTableBucketMaintenanceConfiguration(ctx, conn, &input)
}

func findTableBucketMaintenanceConfiguration(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableBucketMaintenanceConfigurationInput) (*s3tables.GetTableBucketMaintenanceConfigurationOutput, error) {
	output, err := conn.GetTableBucketMaintenanceConfiguration(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type tableBucketResourceModel struct {
	framework.WithRegionModel
	ARN                      types.String                                                    `tfsdk:"arn"`
	CreatedAt                timetypes.RFC3339                                               `tfsdk:"created_at"`
	EncryptionConfiguration  fwtypes.ObjectValueOf[encryptionConfigurationModel]             `tfsdk:"encryption_configuration"`
	ForceDestroy             types.Bool                                                      `tfsdk:"force_destroy"`
	MaintenanceConfiguration fwtypes.ObjectValueOf[tableBucketMaintenanceConfigurationModel] `tfsdk:"maintenance_configuration" autoflex:"-"`
	Name                     types.String                                                    `tfsdk:"name"`
	OwnerAccountID           types.String                                                    `tfsdk:"owner_account_id"`
	Tags                     tftags.Map                                                      `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                      `tfsdk:"tags_all"`
}
type encryptionConfigurationModel struct {
	KMSKeyARN    fwtypes.ARN                               `tfsdk:"kms_key_arn"`
	SSEAlgorithm fwtypes.StringEnum[awstypes.SSEAlgorithm] `tfsdk:"sse_algorithm"`
}

func (m *encryptionConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.EncryptionConfiguration:
		m.SSEAlgorithm = fwtypes.StringEnumValue(t.SseAlgorithm)
		if t.SseAlgorithm == awstypes.SSEAlgorithmAes256 {
			m.KMSKeyARN = fwtypes.ARNNull()
		} else {
			m.KMSKeyARN = fwtypes.ARNValue(aws.ToString(t.KmsKeyArn))
		}
	}
	return diags
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

	diags.Append(fwflex.Expand(ctx, model, &value)...)

	return &awstypes.TableBucketMaintenanceSettingsMemberIcebergUnreferencedFileRemoval{
		Value: value,
	}, diags
}

func flattenIcebergUnreferencedFileRemovalSettings(ctx context.Context, in awstypes.TableBucketMaintenanceSettings) (result fwtypes.ObjectValueOf[icebergUnreferencedFileRemovalSettingsModel], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	switch t := in.(type) {
	case *awstypes.TableBucketMaintenanceSettingsMemberIcebergUnreferencedFileRemoval:
		var model icebergUnreferencedFileRemovalSettingsModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &model)...)
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

// emptyTableBucket deletes all tables in all namespaces within the specified table bucket.
// This is used when force_destroy is enabled to allow deletion of non-empty table buckets.
func emptyTableBucket(ctx context.Context, conn *s3tables.Client, tableBucketARN string) error {
	tflog.Debug(ctx, "Starting to empty table bucket", map[string]any{
		"table_bucket_arn": tableBucketARN,
	})

	// First, list all namespaces in the table bucket.
	input := s3tables.ListNamespacesInput{
		TableBucketARN: aws.String(tableBucketARN),
	}
	pages := s3tables.NewListNamespacesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return fmt.Errorf("listing S3 Tables Table Bucket (%s) namespaces: %w", tableBucketARN, err)
		}

		// For each namespace, list and delete all tables.
		for _, v := range page.Namespaces {
			namespace := v.Namespace[0]
			tflog.Debug(ctx, "Processing namespace", map[string]any{
				names.AttrNamespace: namespace,
			})

			inputLT := s3tables.ListTablesInput{
				Namespace:      aws.String(namespace),
				TableBucketARN: aws.String(tableBucketARN),
			}
			pages := s3tables.NewListTablesPaginator(conn, &inputLT)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return fmt.Errorf("listing S3 Tables Table Bucket (%s,%s) tables: %w", tableBucketARN, namespace, err)
				}

				// Delete each table.
				for _, v := range page.Tables {
					name := aws.ToString(v.Name)
					tflog.Debug(ctx, "Deleting table", map[string]any{
						names.AttrName:      name,
						names.AttrNamespace: namespace,
					})

					input := s3tables.DeleteTableInput{
						Name:           aws.String(name),
						Namespace:      aws.String(namespace),
						TableBucketARN: aws.String(tableBucketARN),
					}
					_, err := conn.DeleteTable(ctx, &input)

					if errs.IsA[*awstypes.NotFoundException](err) {
						continue
					}

					if err != nil {
						return fmt.Errorf("deleting S3 Tables Table Bucket (%s,%s) table (%s): %w", tableBucketARN, namespace, name, err)
					}
				}
			}

			// After deleting all tables in the namespace, delete the namespace itself.
			tflog.Debug(ctx, "Deleting namespace", map[string]any{
				names.AttrNamespace: namespace,
			})

			inputDN := s3tables.DeleteNamespaceInput{
				Namespace:      aws.String(namespace),
				TableBucketARN: aws.String(tableBucketARN),
			}
			_, err = conn.DeleteNamespace(ctx, &inputDN)

			if errs.IsA[*awstypes.NotFoundException](err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("deleting S3 Tables Table Bucket (%s) namespace (%s): %w", tableBucketARN, namespace, err)
			}
		}
	}

	tflog.Debug(ctx, "Successfully emptied table bucket", map[string]any{
		"table_bucket_arn": tableBucketARN,
	})

	return nil
}
