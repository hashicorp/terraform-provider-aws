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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3files_mount_target", name="Mount Target")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetMountTargetOutput")
// @Testing(preCheck="testAccPreCheck")
func newMountTargetResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &mountTargetResource{}, nil
}

type mountTargetResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *mountTargetResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3files_mount_target"
}

func (r *mountTargetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"file_system_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSubnetID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrIPAddress: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrSecurityGroups: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mount_target_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrAvailabilityZoneID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zone_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dns_name": schema.StringAttribute{
				Computed: true,
			},
			"network_interface_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *mountTargetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data mountTargetResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := &s3files.CreateMountTargetInput{
		FileSystemId: aws.String(data.FileSystemId.ValueString()),
		SubnetId:     aws.String(data.SubnetId.ValueString()),
	}

	if !data.IpAddress.IsNull() && !data.IpAddress.IsUnknown() {
		input.IpAddress = aws.String(data.IpAddress.ValueString())
	}
	if !data.SecurityGroups.IsNull() && !data.SecurityGroups.IsUnknown() {
		input.SecurityGroups = aws.String(data.SecurityGroups.ValueString())
	}

	output, err := conn.CreateMountTarget(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Files Mount Target for file system (%s)", data.FileSystemId.ValueString()), err.Error())
		return
	}

	data.ID = types.StringPointerValue(output.MountTargetId)
	data.MountTargetId = types.StringPointerValue(output.MountTargetId)
	data.AvailabilityZoneId = types.StringPointerValue(output.AvailabilityZoneId)
	data.AvailabilityZoneName = types.StringPointerValue(output.AvailabilityZoneName)
	data.DnsName = types.StringPointerValue(output.DnsName)
	data.IpAddress = types.StringPointerValue(output.IpAddress)
	data.NetworkInterfaceId = types.StringPointerValue(output.NetworkInterfaceId)
	data.OwnerID = types.StringPointerValue(output.OwnerId)
	data.SecurityGroups = types.StringPointerValue(output.SecurityGroups)
	data.Status = types.StringValue(string(output.Status))
	data.VpcId = types.StringPointerValue(output.VpcId)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *mountTargetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data mountTargetResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	output, err := findMountTargetByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Files Mount Target (%s)", data.ID.ValueString()), err.Error())
		return
	}

	data.AvailabilityZoneId = types.StringPointerValue(output.AvailabilityZoneId)
	data.AvailabilityZoneName = types.StringPointerValue(output.AvailabilityZoneName)
	data.DnsName = types.StringPointerValue(output.DnsName)
	data.FileSystemId = types.StringPointerValue(output.FileSystemId)
	data.IpAddress = types.StringPointerValue(output.IpAddress)
	data.MountTargetId = types.StringPointerValue(output.MountTargetId)
	data.NetworkInterfaceId = types.StringPointerValue(output.NetworkInterfaceId)
	data.OwnerID = types.StringPointerValue(output.OwnerId)
	data.SecurityGroups = types.StringPointerValue(output.SecurityGroups)
	data.Status = types.StringValue(string(output.Status))
	data.VpcId = types.StringPointerValue(output.VpcId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *mountTargetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data mountTargetResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := &s3files.UpdateMountTargetInput{
		MountTargetId: aws.String(data.ID.ValueString()),
	}

	if !data.SecurityGroups.IsNull() && !data.SecurityGroups.IsUnknown() {
		input.SecurityGroups = aws.String(data.SecurityGroups.ValueString())
	}

	_, err := conn.UpdateMountTarget(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Files Mount Target (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *mountTargetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data mountTargetResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	_, err := conn.DeleteMountTarget(ctx, &s3files.DeleteMountTargetInput{
		MountTargetId: aws.String(data.ID.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Files Mount Target (%s)", data.ID.ValueString()), err.Error())
	}
}

func (r *mountTargetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, fwflex.StringValuePath("id"), request, response)
}

func findMountTargetByID(ctx context.Context, conn *s3files.Client, id string) (*s3files.GetMountTargetOutput, error) {
	input := &s3files.GetMountTargetInput{
		MountTargetId: aws.String(id),
	}

	output, err := conn.GetMountTarget(ctx, input)

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

type mountTargetResourceModel struct {
	AvailabilityZoneId   types.String `tfsdk:"availability_zone_id"`
	AvailabilityZoneName types.String `tfsdk:"availability_zone_name"`
	DnsName              types.String `tfsdk:"dns_name"`
	FileSystemId         types.String `tfsdk:"file_system_id"`
	ID                   types.String `tfsdk:"id"`
	IpAddress            types.String `tfsdk:"ip_address"`
	MountTargetId        types.String `tfsdk:"mount_target_id"`
	NetworkInterfaceId   types.String `tfsdk:"network_interface_id"`
	OwnerID              types.String `tfsdk:"owner_id"`
	SecurityGroups       types.String `tfsdk:"security_groups"`
	Status               types.String `tfsdk:"status"`
	VpcId                types.String `tfsdk:"vpc_id"`
}
