// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_s3files_synchronization_configuration", name="Synchronization Configuration")
// @IdentityAttribute("file_system_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetSynchronizationConfigurationOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
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
			"expiration_days_after_last_access": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"import_size_less_than": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"import_prefix": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
	input := buildSyncConfigInput(&data)

	_, err := conn.PutSynchronizationConfiguration(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Files Synchronization Configuration (%s)", fsId), err.Error())
		return
	}

	output, err := findSyncConfigByID(ctx, conn, fsId)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Files Synchronization Configuration (%s)", fsId), err.Error())
		return
	}

	flattenSyncConfig(&data, output)

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

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Files Synchronization Configuration (%s)", fsId), err.Error())
		return
	}

	flattenSyncConfig(&data, output)

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
	input := buildSyncConfigInput(&data)

	_, err := conn.PutSynchronizationConfiguration(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Files Synchronization Configuration (%s)", fsId), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *synchronizationConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Synchronization configuration cannot be deleted, only updated.
}

func (r *synchronizationConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("file_system_id"), request, response)
}

func buildSyncConfigInput(data *syncConfigResourceModel) *s3files.PutSynchronizationConfigurationInput {
	prefix := ""
	if !data.ImportPrefix.IsNull() && !data.ImportPrefix.IsUnknown() {
		prefix = data.ImportPrefix.ValueString()
	}

	return &s3files.PutSynchronizationConfigurationInput{
		FileSystemId: aws.String(data.FileSystemId.ValueString()),
		ExpirationDataRules: []awstypes.ExpirationDataRule{
			{
				DaysAfterLastAccess: aws.Int32(int32(data.ExpirationDaysAfterLastAccess.ValueInt64())),
			},
		},
		ImportDataRules: []awstypes.ImportDataRule{
			{
				Prefix:       aws.String(prefix),
				SizeLessThan: aws.Int64(data.ImportSizeLessThan.ValueInt64()),
				Trigger:      awstypes.ImportTriggerOnFileAccess,
			},
		},
	}
}

func flattenSyncConfig(data *syncConfigResourceModel, output *s3files.GetSynchronizationConfigurationOutput) {
	if len(output.ExpirationDataRules) > 0 {
		data.ExpirationDaysAfterLastAccess = types.Int64Value(int64(aws.ToInt32(output.ExpirationDataRules[0].DaysAfterLastAccess)))
	}
	if len(output.ImportDataRules) > 0 {
		data.ImportSizeLessThan = types.Int64Value(aws.ToInt64(output.ImportDataRules[0].SizeLessThan))
		data.ImportPrefix = types.StringPointerValue(output.ImportDataRules[0].Prefix)
	}
}

func findSyncConfigByID(ctx context.Context, conn *s3files.Client, fsId string) (*s3files.GetSynchronizationConfigurationOutput, error) {
	output, err := conn.GetSynchronizationConfiguration(ctx, &s3files.GetSynchronizationConfigurationInput{
		FileSystemId: aws.String(fsId),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

type syncConfigResourceModel struct {
	ExpirationDaysAfterLastAccess types.Int64  `tfsdk:"expiration_days_after_last_access"`
	FileSystemId                  types.String `tfsdk:"file_system_id"`
	ImportPrefix                  types.String `tfsdk:"import_prefix"`
	ImportSizeLessThan            types.Int64  `tfsdk:"import_size_less_than"`
}
