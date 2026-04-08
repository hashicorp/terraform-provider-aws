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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3files_file_system", name="File System")
// @IdentityAttribute("file_system_id")
// @Tags(identifierAttribute="file_system_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetFileSystemOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newFileSystemResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &fileSystemResource{}, nil
}

type fileSystemResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *fileSystemResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3files_file_system"
}

func (r *fileSystemResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"bucket": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"kms_key_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"accept_bucket_warning": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"file_system_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file_system_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			"status_message": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *fileSystemResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data fileSystemResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := &s3files.CreateFileSystemInput{
		Bucket:  aws.String(data.Bucket.ValueString()),
		RoleArn: aws.String(data.RoleArn.ValueString()),
		Tags:    getTagsIn(ctx),
	}

	if !data.Prefix.IsNull() && !data.Prefix.IsUnknown() {
		input.Prefix = aws.String(data.Prefix.ValueString())
	}
	if !data.KmsKeyId.IsNull() && !data.KmsKeyId.IsUnknown() {
		input.KmsKeyId = aws.String(data.KmsKeyId.ValueString())
	}
	if !data.AcceptBucketWarning.IsNull() {
		input.AcceptBucketWarning = aws.Bool(data.AcceptBucketWarning.ValueBool())
	}

	output, err := conn.CreateFileSystem(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Files File System (%s)", data.Bucket.ValueString()), err.Error())
		return
	}

	data.ID = types.StringPointerValue(output.FileSystemId)
	data.FileSystemArn = types.StringPointerValue(output.FileSystemArn)
	data.FileSystemId = types.StringPointerValue(output.FileSystemId)
	data.OwnerID = types.StringPointerValue(output.OwnerId)
	data.Status = types.StringValue(string(output.Status))
	data.StatusMessage = types.StringPointerValue(output.StatusMessage)
	data.KmsKeyId = types.StringPointerValue(output.KmsKeyId)
	data.Prefix = types.StringPointerValue(output.Prefix)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *fileSystemResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data fileSystemResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	output, err := findFileSystemByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Files File System (%s)", data.ID.ValueString()), err.Error())
		return
	}

	data.Bucket = types.StringPointerValue(output.Bucket)
	data.FileSystemArn = types.StringPointerValue(output.FileSystemArn)
	data.FileSystemId = types.StringPointerValue(output.FileSystemId)
	data.KmsKeyId = types.StringPointerValue(output.KmsKeyId)
	data.OwnerID = types.StringPointerValue(output.OwnerId)
	data.Prefix = types.StringPointerValue(output.Prefix)
	data.RoleArn = fwtypes.ARNValue(aws.ToString(output.RoleArn))
	data.Status = types.StringValue(string(output.Status))
	data.StatusMessage = types.StringPointerValue(output.StatusMessage)

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *fileSystemResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data fileSystemResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Tags only - all other attributes require replacement.

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *fileSystemResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data fileSystemResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	_, err := conn.DeleteFileSystem(ctx, &s3files.DeleteFileSystemInput{
		FileSystemId: aws.String(data.ID.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Files File System (%s)", data.ID.ValueString()), err.Error())
	}
}

func (r *fileSystemResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func findFileSystemByID(ctx context.Context, conn *s3files.Client, id string) (*s3files.GetFileSystemOutput, error) {
	output, err := conn.GetFileSystem(ctx, &s3files.GetFileSystemInput{
		FileSystemId: aws.String(id),
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

type fileSystemResourceModel struct {
	AcceptBucketWarning types.Bool   `tfsdk:"accept_bucket_warning"`
	Bucket              types.String `tfsdk:"bucket"`
	FileSystemArn       types.String `tfsdk:"file_system_arn"`
	FileSystemId        types.String `tfsdk:"file_system_id"`
	ID                  types.String `tfsdk:"id"`
	KmsKeyId            types.String `tfsdk:"kms_key_id"`
	OwnerID             types.String `tfsdk:"owner_id"`
	Prefix              types.String `tfsdk:"prefix"`
	RoleArn             fwtypes.ARN  `tfsdk:"role_arn"`
	Status              types.String `tfsdk:"status"`
	StatusMessage       types.String `tfsdk:"status_message"`
	Tags                tftags.Map   `tfsdk:"tags"`
	TagsAll             tftags.Map   `tfsdk:"tags_all"`
}
