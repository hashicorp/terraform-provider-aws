// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_route_server_vpc_association", name="VPC Route Server VPC Association")
func newVPCRouteServerVPCAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcRouteServerVPCAssociationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type vpcRouteServerVPCAssociationResource struct {
	framework.ResourceWithModel[vpcRouteServerVPCAssociationResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *vpcRouteServerVPCAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"route_server_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
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

func (r *vpcRouteServerVPCAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcRouteServerVPCAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	routeServerID, vpcID := fwflex.StringValueFromFramework(ctx, data.RouteServerID), fwflex.StringValueFromFramework(ctx, data.VpcID)
	input := ec2.AssociateRouteServerInput{
		RouteServerId: aws.String(routeServerID),
		VpcId:         aws.String(vpcID),
	}

	_, err := conn.AssociateRouteServer(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating VPC Route Server (%s) VPC (%s) Association", routeServerID, vpcID), err.Error())

		return
	}

	if _, err := waitRouteServerAssociationCreated(ctx, conn, routeServerID, vpcID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server (%s) VPC (%s) Association create", routeServerID, vpcID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *vpcRouteServerVPCAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcRouteServerVPCAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	routeServerID, vpcID := fwflex.StringValueFromFramework(ctx, data.RouteServerID), fwflex.StringValueFromFramework(ctx, data.VpcID)
	_, err := findRouteServerAssociationByTwoPartKey(ctx, conn, routeServerID, vpcID)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Route Server (%s) VPC (%s) Association", routeServerID, vpcID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcRouteServerVPCAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpcRouteServerVPCAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	routeServerID, vpcID := fwflex.StringValueFromFramework(ctx, data.RouteServerID), fwflex.StringValueFromFramework(ctx, data.VpcID)
	input := ec2.DisassociateRouteServerInput{
		RouteServerId: aws.String(routeServerID),
		VpcId:         aws.String(vpcID),
	}
	_, err := conn.DisassociateRouteServer(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteServerIdNotAssociated, errCodeInvalidRouteServerIdNotFound, errCodeInvalidVPCIDNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Route Server (%s) VPC (%s) Association", routeServerID, vpcID), err.Error())

		return
	}

	if _, err := waitRouteServerAssociationDeleted(ctx, conn, routeServerID, vpcID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server (%s) VPC (%s) Association delete", routeServerID, vpcID), err.Error())

		return
	}
}

func (r *vpcRouteServerVPCAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		vpcRouteServerVPCAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, vpcRouteServerVPCAssociationIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("route_server_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrVPCID), parts[1])...)
}

type vpcRouteServerVPCAssociationResourceModel struct {
	framework.WithRegionModel
	RouteServerID types.String   `tfsdk:"route_server_id"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
	VpcID         types.String   `tfsdk:"vpc_id"`
}
