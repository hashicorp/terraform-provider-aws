// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_s3files_synchronization_configuration", name="Synchronization Configuration")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetSynchronizationConfigurationOutput")
// @Testing(preCheck="testAccPreCheck")
func newSynchronizationConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &synchronizationConfigurationResource{}, nil
}

type synchronizationConfigurationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *synchronizationConfigurationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3files_synchronization_configuration"
}

func (r *synchronizationConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"file_system_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_size_threshold_bytes": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"expiration_days": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"import_file_size_threshold_bytes": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *synchronizationConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data syncConfigResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	fsId := data.FileSystemId.ValueString()
	input := &s3files.PutSynchronizationConfigurationInput{
		FileSystemId: aws.String(fsId),
	}

	if !data.FileSizeThresholdBytes.IsNull() && !data.FileSizeThresholdBytes.IsUnknown() {
		input.FileSizeThresholdBytes = aws.Int32(int32(data.FileSizeThresholdBytes.ValueInt64()))
	}
	if !data.ExpirationDays.IsNull() && !data.ExpirationDays.IsUnknown() {
		input.ExpirationDataRule = &awstypes.ExpirationDataRule{
			Days: aws.Int32(int32(data.ExpirationDays.ValueInt64())),
		}
	}
	if !data.ImportFileSizeThresholdBytes.IsNull() && !data.ImportFileSizeThresholdBytes.IsUnknown() {
		input.ImportDataRule = &awstypes.ImportDataRule{
			FileSizeThresholdBytes: aws.Int32(int32(data.ImportFileSizeThresholdBytes.ValueInt64())),
		}
	}

	_, err := conn.PutSynchronizationConfiguration(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Files Synchronization Configuration (%s)", fsId), err.Error())
		return
	}

	// Read back to get computed values.
	output, err := findSyncConfigByID(ctx, conn, fsId)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Files Synchronization Configuration (%s)", fsId), err.Error())
		return
	}

	data.FileSizeThresholdBytes = types.Int64Value(int64(aws.ToInt32(output.FileSizeThresholdBytes)))
	if output.ExpirationDataRule != nil {
		data.ExpirationDays = types.Int64Value(int64(aws.ToInt32(output.ExpirationDataRule.Days)))
	}
	if output.ImportDataRule != nil {
		data.ImportFileSizeThresholdBytes = types.Int64Value(int64(aws.ToInt32(output.ImportDataRule.FileSizeThresholdBytes)))
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *synchronizationConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data syncConfigResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	fsId := data.FileSystemId.ValueString()
	output, err := findSyncConfigByID(ctx, conn, fsId)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Files Synchronization Configuration (%s)", fsId), err.Error())
		return
	}

	data.FileSizeThresholdBytes = types.Int64Value(int64(aws.ToInt32(output.FileSizeThresholdBytes)))
	if output.ExpirationDataRule != nil {
		data.ExpirationDays = types.Int64Value(int64(aws.ToInt32(output.ExpirationDataRule.Days)))
	}
	if output.ImportDataRule != nil {
		data.ImportFileSizeThresholdBytes = types.Int64Value(int64(aws.ToInt32(output.ImportDataRule.FileSizeThresholdBytes)))
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *synchronizationConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data syncConfigResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	fsId := data.FileSystemId.ValueString()
	input := &s3files.PutSynchronizationConfigurationInput{
		FileSystemId: aws.String(fsId),
	}

	if !data.FileSizeThresholdBytes.IsNull() {
		input.FileSizeThresholdBytes = aws.Int32(int32(data.FileSizeThresholdBytes.ValueInt64()))
	}
	if !data.ExpirationDays.IsNull() {
		input.ExpirationDataRule = &awstypes.ExpirationDataRule{
			Days: aws.Int32(int32(data.ExpirationDays.ValueInt64())),
		}
	}
	if !data.ImportFileSizeThresholdBytes.IsNull() {
		input.ImportDataRule = &awstypes.ImportDataRule{
			FileSizeThresholdBytes: aws.Int32(int32(data.ImportFileSizeThresholdBytes.ValueInt64())),
		}
	}

	_, err := conn.PutSynchronizationConfiguration(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Files Synchronization Configuration (%s)", fsId), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *synchronizationConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Synchronization configuration cannot be deleted, only updated.
	// Removing from state is sufficient.
}

func (r *synchronizationConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, fwflex.StringValuePath("file_system_id"), request, response)
}

func findSyncConfigByID(ctx context.Context, conn *s3files.Client, fsId string) (*s3files.GetSynchronizationConfigurationOutput, error) {
	input := &s3files.GetSynchronizationConfigurationInput{
		FileSystemId: aws.String(fsId),
	}

	output, err := conn.GetSynchronizationConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &tfresource.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type syncConfigResourceModel struct {
	ExpirationDays               types.Int64  `tfsdk:"expiration_days"`
	FileSizeThresholdBytes       types.Int64  `tfsdk:"file_size_threshold_bytes"`
	FileSystemId                 types.String `tfsdk:"file_system_id"`
	ImportFileSizeThresholdBytes types.Int64  `tfsdk:"import_file_size_threshold_bytes"`
}
