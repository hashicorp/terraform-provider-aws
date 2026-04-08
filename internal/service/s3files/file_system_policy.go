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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3files_file_system_policy", name="File System Policy")
// @IdentityAttribute("file_system_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetFileSystemPolicyOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newFileSystemPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &fileSystemPolicyResource{}, nil
}

type fileSystemPolicyResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *fileSystemPolicyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3files_file_system_policy"
}

func (r *fileSystemPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"file_system_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
		},
	}
}

func (r *fileSystemPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data fileSystemPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	fsId := data.FileSystemId.ValueString()
	_, err := conn.PutFileSystemPolicy(ctx, &s3files.PutFileSystemPolicyInput{
		FileSystemId: aws.String(fsId),
		Policy:       aws.String(data.Policy.ValueString()),
	})
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Files File System Policy (%s)", fsId), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *fileSystemPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data fileSystemPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	fsId := data.FileSystemId.ValueString()
	output, err := findFileSystemPolicyByID(ctx, conn, fsId)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Files File System Policy (%s)", fsId), err.Error())
		return
	}

	data.Policy = fwtypes.IAMPolicyValue(aws.ToString(output.Policy))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *fileSystemPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data fileSystemPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	fsId := data.FileSystemId.ValueString()
	_, err := conn.PutFileSystemPolicy(ctx, &s3files.PutFileSystemPolicyInput{
		FileSystemId: aws.String(fsId),
		Policy:       aws.String(data.Policy.ValueString()),
	})
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Files File System Policy (%s)", fsId), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *fileSystemPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data fileSystemPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	_, err := conn.DeleteFileSystemPolicy(ctx, &s3files.DeleteFileSystemPolicyInput{
		FileSystemId: aws.String(data.FileSystemId.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Files File System Policy (%s)", data.FileSystemId.ValueString()), err.Error())
	}
}

func (r *fileSystemPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("file_system_id"), request, response)
}

func findFileSystemPolicyByID(ctx context.Context, conn *s3files.Client, fsId string) (*s3files.GetFileSystemPolicyOutput, error) {
	output, err := conn.GetFileSystemPolicy(ctx, &s3files.GetFileSystemPolicyInput{
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

	if output == nil || aws.ToString(output.Policy) == "" {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type fileSystemPolicyResourceModel struct {
	FileSystemId types.String      `tfsdk:"file_system_id"`
	Policy       fwtypes.IAMPolicy `tfsdk:"policy"`
}
