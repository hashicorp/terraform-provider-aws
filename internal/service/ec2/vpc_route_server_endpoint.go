// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_route_server_endpoint", name="VPC Route Server Endpoint")
// @Tags(identifierAttribute="route_server_endpoint_id")
// @Testing(tagsTest=false)
func newVPCRouteServerEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcRouteServerEndpointResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type vpcRouteServerEndpointResource struct {
	framework.ResourceWithModel[vpcRouteServerEndpointResourceModel]
	framework.WithTimeouts
}

func (r *vpcRouteServerEndpointResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"eni_address": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"eni_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"route_server_endpoint_id": framework.IDAttribute(),
			"route_server_id": schema.StringAttribute{
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
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *vpcRouteServerEndpointResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcRouteServerEndpointResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	var input ec2.CreateRouteServerEndpointInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeRouteServerEndpoint)

	output, err := conn.CreateRouteServerEndpoint(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating VPC Route Server Endpoint", err.Error())

		return
	}

	// Set values for unknowns.
	rse := output.RouteServerEndpoint
	id := aws.ToString(rse.RouteServerEndpointId)
	data.ARN = r.routeServerEndpointARN(ctx, id)
	data.RouteServerEndpointID = fwflex.StringValueToFramework(ctx, id)

	rse, err = waitRouteServerEndpointCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server Endpoint (%s) create", id), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, rse, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *vpcRouteServerEndpointResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcRouteServerEndpointResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.RouteServerEndpointID)
	rse, err := findRouteServerEndpointByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Route Server Endpoint (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, rse, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ARN = r.routeServerEndpointARN(ctx, id)
	setTagsOut(ctx, rse.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcRouteServerEndpointResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpcRouteServerEndpointResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.RouteServerEndpointID)
	input := ec2.DeleteRouteServerEndpointInput{
		RouteServerEndpointId: aws.String(id),
	}
	_, err := conn.DeleteRouteServerEndpoint(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteServerEndpointIdNotFound, errCodeIncorrectState) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Route Server Endpoint (%s)", id), err.Error())

		return
	}

	if _, err := waitRouteServerEndpointDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server Endpoint (%s) delete", id), err.Error())

		return
	}
}

func (r *vpcRouteServerEndpointResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("route_server_endpoint_id"), request, response)
}

func (r *vpcRouteServerEndpointResource) routeServerEndpointARN(ctx context.Context, id string) types.String {
	return fwflex.StringValueToFramework(ctx, r.Meta().RegionalARN(ctx, names.EC2, "route-server-endpoint/"+id))
}

type vpcRouteServerEndpointResourceModel struct {
	framework.WithRegionModel
	ARN                   types.String   `tfsdk:"arn"`
	EniAddress            types.String   `tfsdk:"eni_address"`
	EniID                 types.String   `tfsdk:"eni_id"`
	RouteServerEndpointID types.String   `tfsdk:"route_server_endpoint_id"`
	RouteServerID         types.String   `tfsdk:"route_server_id"`
	SubnetID              types.String   `tfsdk:"subnet_id"`
	Tags                  tftags.Map     `tfsdk:"tags"`
	TagsAll               tftags.Map     `tfsdk:"tags_all"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
	VpcID                 types.String   `tfsdk:"vpc_id"`
}
