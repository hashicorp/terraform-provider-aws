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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3files_access_point", name="Access Point")
// @Tags(identifierAttribute="access_point_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetAccessPointOutput")
// @Testing(preCheck="testAccPreCheck")
func newAccessPointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &accessPointResource{}, nil
}

type accessPointResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *accessPointResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3files_access_point"
}

func (r *accessPointResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"file_system_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"posix_user": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"root_directory": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"access_point_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_point_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrOwnerID: schema.StringAttribute{
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

func (r *accessPointResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accessPointResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := &s3files.CreateAccessPointInput{
		FileSystemId: aws.String(data.FileSystemId.ValueString()),
		Tags:         getTagsIn(ctx),
	}

	output, err := conn.CreateAccessPoint(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Files Access Point for file system (%s)", data.FileSystemId.ValueString()), err.Error())
		return
	}

	data.ID = types.StringPointerValue(output.AccessPointId)
	data.AccessPointArn = types.StringPointerValue(output.AccessPointArn)
	data.AccessPointId = types.StringPointerValue(output.AccessPointId)
	data.OwnerID = types.StringPointerValue(output.OwnerId)
	data.Status = types.StringValue(string(output.Status))

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *accessPointResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accessPointResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	output, err := findAccessPointByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Files Access Point (%s)", data.ID.ValueString()), err.Error())
		return
	}

	data.AccessPointArn = types.StringPointerValue(output.AccessPointArn)
	data.AccessPointId = types.StringPointerValue(output.AccessPointId)
	data.FileSystemId = types.StringPointerValue(output.FileSystemId)
	data.OwnerID = types.StringPointerValue(output.OwnerId)
	data.Status = types.StringValue(string(output.Status))

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessPointResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data accessPointResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Tags only - all other attributes require replacement.

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessPointResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accessPointResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	_, err := conn.DeleteAccessPoint(ctx, &s3files.DeleteAccessPointInput{
		AccessPointId: aws.String(data.ID.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Files Access Point (%s)", data.ID.ValueString()), err.Error())
	}
}

func (r *accessPointResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, fwflex.StringValuePath("id"), request, response)
}

func findAccessPointByID(ctx context.Context, conn *s3files.Client, id string) (*s3files.GetAccessPointOutput, error) {
	input := &s3files.GetAccessPointInput{
		AccessPointId: aws.String(id),
	}

	output, err := conn.GetAccessPoint(ctx, input)

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

type accessPointResourceModel struct {
	AccessPointArn types.String `tfsdk:"access_point_arn"`
	AccessPointId  types.String `tfsdk:"access_point_id"`
	FileSystemId   types.String `tfsdk:"file_system_id"`
	ID             types.String `tfsdk:"id"`
	OwnerID        types.String `tfsdk:"owner_id"`
	PosixUser      types.String `tfsdk:"posix_user"`
	RootDirectory  types.String `tfsdk:"root_directory"`
	Status         types.String `tfsdk:"status"`
	Tags           tftags.Map   `tfsdk:"tags"`
	TagsAll        tftags.Map   `tfsdk:"tags_all"`
}
