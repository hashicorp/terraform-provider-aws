// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_network_interface_permission", name="Network Interface Permission")
func newNetworkInterfacePermissionResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &networkInterfacePermissionResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type networkInterfacePermissionResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithTimeouts
}

func (r *networkInterfacePermissionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrNetworkInterfaceID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_interface_permission_id": framework.IDAttribute(),
			"permission": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InterfacePermissionType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *networkInterfacePermissionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data networkInterfacePermissionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.CreateNetworkInterfacePermissionInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateNetworkInterfacePermission(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating EC2 Network Interface (%s) Permission", data.NetworkInterfaceID.ValueString()), err.Error())

		return
	}

	data.NetworkInterfacePermissionID = fwflex.StringToFramework(ctx, output.InterfacePermission.NetworkInterfacePermissionId)

	if _, err := waitNetworkInterfacePermissionCreated(ctx, conn, data.NetworkInterfacePermissionID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Network Interface Permission (%s) create", data.NetworkInterfacePermissionID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *networkInterfacePermissionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data networkInterfacePermissionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	output, err := findNetworkInterfacePermissionByID(ctx, conn, data.NetworkInterfacePermissionID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Network Interface Permission (%s)", data.NetworkInterfacePermissionID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *networkInterfacePermissionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data networkInterfacePermissionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.DeleteNetworkInterfacePermissionInput{
		NetworkInterfacePermissionId: fwflex.StringFromFramework(ctx, data.NetworkInterfacePermissionID),
	}

	_, err := conn.DeleteNetworkInterfacePermission(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPermissionIDNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Network Interface Permission (%s)", data.NetworkInterfacePermissionID.ValueString()), err.Error())

		return
	}

	if _, err := waitNetworkInterfacePermissionDeleted(ctx, conn, data.NetworkInterfacePermissionID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Network Interface Permission (%s) delete", data.NetworkInterfacePermissionID.ValueString()), err.Error())

		return
	}
}

func (r *networkInterfacePermissionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("network_interface_permission_id"), request, response)
}

type networkInterfacePermissionResourceModel struct {
	AWSAccountID                 types.String                                         `tfsdk:"aws_account_id"`
	NetworkInterfaceID           types.String                                         `tfsdk:"network_interface_id"`
	NetworkInterfacePermissionID types.String                                         `tfsdk:"network_interface_permission_id"`
	Permission                   fwtypes.StringEnum[awstypes.InterfacePermissionType] `tfsdk:"permission"`
	Timeouts                     timeouts.Value                                       `tfsdk:"timeouts"`
}
