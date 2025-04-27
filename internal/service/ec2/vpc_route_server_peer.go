// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_vpc_route_server_peer", name="VPC Route Server Peer")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func newVPCRouteServerPeerResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCRouteServerPeer{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCRouteServerPeer = "VPC Route Server Peer"
)

type resourceVPCRouteServerPeer struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
	framework.WithNoUpdate
}

func (r *resourceVPCRouteServerPeer) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"route_server_endpoint_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"peer_address": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"route_server_id": schema.StringAttribute{
				Computed: true,
			},
			"endpoint_eni_address": schema.StringAttribute{
				Computed: true,
			},
			"endpoint_eni_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrSubnetID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"bgp_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceVPCRouteServerPeerBgpOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"peer_asn": schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 4294967295),
							},
						},
						"peer_liveness_detection": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.RouteServerPeerLivenessMode](),
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

func (r *resourceVPCRouteServerPeer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCRouteServerPeerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.CreateRouteServerPeerInput{
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeRouteServerPeer),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateRouteServerPeer(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServerPeer, plan.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.RouteServerPeer == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServerPeer, plan.RouteServerId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.RouteServerPeer, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitVPCRouteServerPeerCreated(ctx, conn, plan.RouteServerPeerId.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCRouteServerPeer, plan.RouteServerPeerId.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVPCRouteServerPeer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerPeerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCRouteServerPeerByID(ctx, conn, state.RouteServerPeerId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, ResNameVPCRouteServerPeer, state.RouteServerPeerId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVPCRouteServerPeer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerPeerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := findVPCRouteServerPeerByID(ctx, conn, state.RouteServerPeerId.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCRouteServerPeer, state.RouteServerPeerId.String(), err),
			err.Error(),
		)
		return
	}
	input := ec2.DeleteRouteServerPeerInput{
		RouteServerPeerId: state.RouteServerPeerId.ValueStringPointer(),
	}

	_, err = conn.DeleteRouteServerPeer(ctx, &input)

	if tfresource.NotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCRouteServerPeer, state.RouteServerPeerId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCRouteServerPeerDeleted(ctx, conn, state.RouteServerPeerId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameVPCRouteServerPeer, state.RouteServerPeerId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCRouteServerPeer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitVPCRouteServerPeerCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteServerPeer, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RouteServerPeerStatePending),
		Target:                    enum.Slice(awstypes.RouteServerPeerStateAvailable),
		Refresh:                   statusVPCRouteServerPeer(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RouteServerPeer); ok {
		return out, err
	}

	return nil, err
}

func waitVPCRouteServerPeerDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteServerPeer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteServerStateDeleting),
		Target:  []string{},
		Refresh: statusVPCRouteServerPeer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RouteServerPeer); ok {
		return out, err
	}

	return nil, err
}

func statusVPCRouteServerPeer(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findVPCRouteServerPeerByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func findVPCRouteServerPeerByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.RouteServerPeer, error) {
	input := ec2.DescribeRouteServerPeersInput{
		RouteServerPeerIds: []string{id},
	}
	var routeServerPeers []awstypes.RouteServerPeer
	paginator := ec2.NewDescribeRouteServerPeersPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if tfresource.NotFound(err) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: &input,
				}
			}
			return nil, err
		}
		if page != nil && len(page.RouteServerPeers) > 0 {
			for _, routeServerPeer := range page.RouteServerPeers {
				if routeServerPeer.State == awstypes.RouteServerPeerStateDeleted {
					continue
				}
				routeServerPeers = append(routeServerPeers, routeServerPeer)
			}
		}
	}
	if len(routeServerPeers) == 0 {
		return nil, &retry.NotFoundError{
			LastError:   errors.New("route server not found"),
			LastRequest: &input,
		}
	} else {
		return &routeServerPeers[0], nil
	}
}

type resourceVPCRouteServerPeerModel struct {
	BgpOptions            fwtypes.ListNestedObjectValueOf[resourceVPCRouteServerPeerBgpOptionsModel] `tfsdk:"bgp_options"`
	EndpointEniAddress    types.String                                                               `tfsdk:"endpoint_eni_address"`
	EndpointEniId         types.String                                                               `tfsdk:"endpoint_eni_id"`
	PeerAddress           types.String                                                               `tfsdk:"peer_address"`
	RouteServerEndpointId types.String                                                               `tfsdk:"route_server_endpoint_id"`
	RouteServerId         types.String                                                               `tfsdk:"route_server_id"`
	RouteServerPeerId     types.String                                                               `tfsdk:"id"`
	SubnetId              types.String                                                               `tfsdk:"subnet_id"`
	Tags                  tftags.Map                                                                 `tfsdk:"tags"`
	TagsAll               tftags.Map                                                                 `tfsdk:"tags_all"`
	Timeouts              timeouts.Value                                                             `tfsdk:"timeouts"`
	VpcId                 types.String                                                               `tfsdk:"vpc_id"`
}

type resourceVPCRouteServerPeerBgpOptionsModel struct {
	PeerAsn               types.Int64                                              `tfsdk:"peer_asn"`
	PeerLivenessDetection fwtypes.StringEnum[awstypes.RouteServerPeerLivenessMode] `tfsdk:"peer_liveness_detection"`
}
