// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"

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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_network_interface_permission", name="Network Interface Permission")
// @Tags(identifierAttribute="id")
func newNetworkInterfacePermission(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &networkInterfacePermission{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameNetworkInterfacePermission = "Network Interface Permission"
)

type networkInterfacePermission struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *networkInterfacePermission) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_network_interface_permission"
}

func (r *networkInterfacePermission) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrAccountID: schema.StringAttribute{
				Required: true,
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
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *networkInterfacePermission) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceNetworkInterfacePermission
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.CreateNetworkInterfacePermissionInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateNetworkInterfacePermission(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameNetworkInterfacePermission, data.NetworkInterfaceID.String(), err),
			err.Error(),
		)
		return
	}
	if output == nil || output.InterfacePermission == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameNetworkInterfacePermission, data.NetworkInterfaceID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	data.NetworkInterfacePermissionID = fwflex.StringToFramework(ctx, output.InterfacePermission.NetworkInterfacePermissionId)

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	_, err = waitNetworkInterfacePermissionCreated(ctx, conn, data.NetworkInterfacePermissionID.ValueString(), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameNetworkInterfacePermission, data.NetworkInterfaceID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *networkInterfacePermission) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceNetworkInterfacePermission
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
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameNetworkInterfacePermission, data.NetworkInterfacePermissionID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *networkInterfacePermission) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceNetworkInterfacePermission
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
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameNetworkInterfacePermission, data.NetworkInterfacePermissionID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	_, err = waitNetworkInterfacePermissionDeleted(ctx, conn, data.NetworkInterfacePermissionID.ValueString(), deleteTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameNetworkInterfacePermission, data.NetworkInterfacePermissionID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *networkInterfacePermission) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

type resourceNetworkInterfacePermission struct {
	NetworkInterfacePermissionID types.String                                         `tfsdk:"id"`
	NetworkInterfaceID           types.String                                         `tfsdk:"network_interface_id"`
	AWSAccountID                 types.String                                         `tfsdk:"account_id"`
	Permission                   fwtypes.StringEnum[awstypes.InterfacePermissionType] `tfsdk:"permission"`
	Timeouts                     timeouts.Value                                       `tfsdk:"timeouts"`
}
