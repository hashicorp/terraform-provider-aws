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

// @FrameworkResource("aws_vpc_route_server_propagation", name="VPC Route Server Propagation")
func newVPCRouteServerPropagationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcRouteServerPropagationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type vpcRouteServerPropagationResource struct {
	framework.ResourceWithModel[vpcRouteServerPropagationResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *vpcRouteServerPropagationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"route_server_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"route_table_id": schema.StringAttribute{
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

func (r *vpcRouteServerPropagationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcRouteServerPropagationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	routeServerID, routeTableID := fwflex.StringValueFromFramework(ctx, data.RouteServerID), fwflex.StringValueFromFramework(ctx, data.RouteTableID)
	input := ec2.EnableRouteServerPropagationInput{
		RouteServerId: aws.String(routeServerID),
		RouteTableId:  aws.String(routeTableID),
	}

	_, err := conn.EnableRouteServerPropagation(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating VPC Route Server (%s) Propagation (%s)", routeServerID, routeTableID), err.Error())

		return
	}

	if _, err := waitRouteServerPropagationCreated(ctx, conn, routeServerID, routeTableID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server (%s) Propagation (%s) create", routeServerID, routeTableID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *vpcRouteServerPropagationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcRouteServerPropagationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	routeServerID, routeTableID := fwflex.StringValueFromFramework(ctx, data.RouteServerID), fwflex.StringValueFromFramework(ctx, data.RouteTableID)
	_, err := findRouteServerPropagationByTwoPartKey(ctx, conn, routeServerID, routeTableID)
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Route Server (%s) Propagation (%s)", routeServerID, routeTableID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcRouteServerPropagationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpcRouteServerPropagationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	routeServerID, routeTableID := fwflex.StringValueFromFramework(ctx, data.RouteServerID), fwflex.StringValueFromFramework(ctx, data.RouteTableID)
	input := ec2.DisableRouteServerPropagationInput{
		RouteServerId: aws.String(routeServerID),
		RouteTableId:  aws.String(routeTableID),
	}

	_, err := conn.DisableRouteServerPropagation(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteServerIdNotPropagated, errCodeInvalidRouteServerIdNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Route Server (%s) Propagation (%s)", routeServerID, routeTableID), err.Error())

		return
	}

	if _, err := waitRouteServerPropagationDeleted(ctx, conn, routeServerID, routeTableID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server (%s) Propagation (%s) delete", routeServerID, routeTableID), err.Error())

		return
	}
}

func (r *vpcRouteServerPropagationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		vpcRouteServerPropagationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, vpcRouteServerPropagationIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("route_server_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("route_table_id"), parts[1])...)
}

type vpcRouteServerPropagationResourceModel struct {
	framework.WithRegionModel
	RouteServerID types.String   `tfsdk:"route_server_id"`
	RouteTableID  types.String   `tfsdk:"route_table_id"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}
