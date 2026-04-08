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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3files_mount_target", name="Mount Target")
// @IdentityAttribute("mount_target_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetMountTargetOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
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
			"ipv4_address": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ipv6_address": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrSecurityGroups: schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"mount_target_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zone_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_interface_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			"status_message": schema.StringAttribute{
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
		SubnetId:     aws.String(data.SubnetID.ValueString()),
	}

	if !data.Ipv4Address.IsNull() && !data.Ipv4Address.IsUnknown() {
		input.Ipv4Address = aws.String(data.Ipv4Address.ValueString())
	}
	if !data.Ipv6Address.IsNull() && !data.Ipv6Address.IsUnknown() {
		input.Ipv6Address = aws.String(data.Ipv6Address.ValueString())
	}
	if !data.SecurityGroups.IsNull() && !data.SecurityGroups.IsUnknown() {
		response.Diagnostics.Append(data.SecurityGroups.ElementsAs(ctx, &input.SecurityGroups, false)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	output, err := conn.CreateMountTarget(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Files Mount Target for file system (%s)", data.FileSystemId.ValueString()), err.Error())
		return
	}

	data.ID = types.StringPointerValue(output.MountTargetId)
	flattenMountTargetOutput(ctx, &data, output.MountTargetId, output.OwnerId, output.SubnetId,
		output.AvailabilityZoneId, output.FileSystemId, output.Ipv4Address, output.Ipv6Address,
		output.NetworkInterfaceId, output.SecurityGroups, output.Status, output.StatusMessage, output.VpcId)

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

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Files Mount Target (%s)", data.ID.ValueString()), err.Error())
		return
	}

	flattenMountTargetOutput(ctx, &data, output.MountTargetId, output.OwnerId, output.SubnetId,
		output.AvailabilityZoneId, output.FileSystemId, output.Ipv4Address, output.Ipv6Address,
		output.NetworkInterfaceId, output.SecurityGroups, output.Status, output.StatusMessage, output.VpcId)

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
		response.Diagnostics.Append(data.SecurityGroups.ElementsAs(ctx, &input.SecurityGroups, false)...)
		if response.Diagnostics.HasError() {
			return
		}
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func flattenMountTargetOutput(ctx context.Context, data *mountTargetResourceModel, mountTargetId, ownerId, subnetId, azId, fsId, ipv4, ipv6, niId *string, sgs []string, status awstypes.LifeCycleState, statusMsg *string, vpcId *string) {
	data.MountTargetId = types.StringPointerValue(mountTargetId)
	data.OwnerID = types.StringPointerValue(ownerId)
	data.SubnetID = types.StringPointerValue(subnetId)
	data.AvailabilityZoneId = types.StringPointerValue(azId)
	data.FileSystemId = types.StringPointerValue(fsId)
	data.Ipv4Address = types.StringPointerValue(ipv4)
	data.Ipv6Address = types.StringPointerValue(ipv6)
	data.NetworkInterfaceId = types.StringPointerValue(niId)
	data.SecurityGroups = fwflex.FlattenFrameworkStringValueList(ctx, sgs)
	data.Status = types.StringValue(string(status))
	data.StatusMessage = types.StringPointerValue(statusMsg)
	data.VpcId = types.StringPointerValue(vpcId)
}

func findMountTargetByID(findCtx context.Context, conn *s3files.Client, id string) (*s3files.GetMountTargetOutput, error) {
	output, err := conn.GetMountTarget(findCtx, &s3files.GetMountTargetInput{
		MountTargetId: aws.String(id),
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

// stringSliceToCommaSeparated is unused but kept for reference.

type mountTargetResourceModel struct {
	AvailabilityZoneId types.String `tfsdk:"availability_zone_id"`
	FileSystemId       types.String `tfsdk:"file_system_id"`
	ID                 types.String `tfsdk:"id"`
	Ipv4Address        types.String `tfsdk:"ipv4_address"`
	Ipv6Address        types.String `tfsdk:"ipv6_address"`
	MountTargetId      types.String `tfsdk:"mount_target_id"`
	NetworkInterfaceId types.String `tfsdk:"network_interface_id"`
	OwnerID            types.String `tfsdk:"owner_id"`
	SecurityGroups     types.List   `tfsdk:"security_groups"`
	Status             types.String `tfsdk:"status"`
	StatusMessage      types.String `tfsdk:"status_message"`
	SubnetID           types.String `tfsdk:"subnet_id"`
	VpcId              types.String `tfsdk:"vpc_id"`
}
