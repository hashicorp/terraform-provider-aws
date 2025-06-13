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
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_route_server_peer", name="VPC Route Server Peer")
// @Tags(identifierAttribute="route_server_peer_id")
// @Testing(tagsTest=false)
func newVPCRouteServerPeerResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcRouteServerPeerResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type vpcRouteServerPeerResource struct {
	framework.ResourceWithModel[vpcRouteServerPeerResourceModel]
	framework.WithTimeouts
}

func (r *vpcRouteServerPeerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"endpoint_eni_address": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint_eni_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"peer_address": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"route_server_endpoint_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"route_server_peer_id": framework.IDAttribute(),
			"route_server_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrSubnetID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"bgp_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[routeServerBGPOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"peer_asn": schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 4294967295),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						"peer_liveness_detection": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RouteServerPeerLivenessMode](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIfConfigured(),
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *vpcRouteServerPeerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcRouteServerPeerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	var input ec2.CreateRouteServerPeerInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeRouteServerPeer)

	output, err := conn.CreateRouteServerPeer(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating VPC Route Server Peer", err.Error())

		return
	}

	// Set values for unknowns.
	rsp := output.RouteServerPeer
	id := aws.ToString(rsp.RouteServerPeerId)
	data.ARN = r.routeServerPeerARN(ctx, id)
	data.RouteServerPeerID = fwflex.StringValueToFramework(ctx, id)

	rsp, err = waitRouteServerPeerCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server Peer (%s) create", id), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, rsp, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *vpcRouteServerPeerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcRouteServerPeerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.RouteServerPeerID)
	rsp, err := findRouteServerPeerByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Route Server Peer (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, rsp, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ARN = r.routeServerPeerARN(ctx, id)
	setTagsOut(ctx, rsp.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcRouteServerPeerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpcRouteServerPeerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.RouteServerPeerID)
	input := ec2.DeleteRouteServerPeerInput{
		RouteServerPeerId: aws.String(id),
	}
	_, err := conn.DeleteRouteServerPeer(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteServerPeerIdNotFound, errCodeIncorrectState) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Route Server Peer (%s)", id), err.Error())

		return
	}

	if _, err := waitRouteServerPeerDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server Peer (%s) delete", id), err.Error())

		return
	}
}

func (r *vpcRouteServerPeerResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("route_server_peer_id"), request, response)
}

func (r *vpcRouteServerPeerResource) routeServerPeerARN(ctx context.Context, id string) types.String {
	return fwflex.StringValueToFramework(ctx, r.Meta().RegionalARN(ctx, names.EC2, "route-server-peer/"+id))
}

type vpcRouteServerPeerResourceModel struct {
	framework.WithRegionModel
	ARN                   types.String                                                `tfsdk:"arn"`
	BGPOptions            fwtypes.ListNestedObjectValueOf[routeServerBGPOptionsModel] `tfsdk:"bgp_options"`
	EndpointEniAddress    types.String                                                `tfsdk:"endpoint_eni_address"`
	EndpointEniID         types.String                                                `tfsdk:"endpoint_eni_id"`
	PeerAddress           types.String                                                `tfsdk:"peer_address"`
	RouteServerEndpointID types.String                                                `tfsdk:"route_server_endpoint_id"`
	RouteServerID         types.String                                                `tfsdk:"route_server_id"`
	RouteServerPeerID     types.String                                                `tfsdk:"route_server_peer_id"`
	SubnetID              types.String                                                `tfsdk:"subnet_id"`
	Tags                  tftags.Map                                                  `tfsdk:"tags"`
	TagsAll               tftags.Map                                                  `tfsdk:"tags_all"`
	Timeouts              timeouts.Value                                              `tfsdk:"timeouts"`
	VpcID                 types.String                                                `tfsdk:"vpc_id"`
}

type routeServerBGPOptionsModel struct {
	PeerASN               types.Int64                                              `tfsdk:"peer_asn"`
	PeerLivenessDetection fwtypes.StringEnum[awstypes.RouteServerPeerLivenessMode] `tfsdk:"peer_liveness_detection"`
}
