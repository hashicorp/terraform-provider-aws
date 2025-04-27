// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_vpc_route_server_endpoint", name="VPC Route Server Endpoint")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func newVPCRouteServerEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCRouteServerEndpoint{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCRouteServerEndpoint = "VPC Route Server Endpoint"
)

type resourceVPCRouteServerEndpoint struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
	framework.WithNoUpdate
}

func (r *resourceVPCRouteServerEndpoint) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
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
			"eni_id": schema.StringAttribute{
				Computed: true,
			},
			"eni_address": schema.StringAttribute{
				Computed: true,
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceVPCRouteServerEndpoint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCRouteServerEndpointModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.CreateRouteServerEndpointInput{
		ClientToken:       aws.String(sdkid.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeRouteServerEndpoint),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateRouteServerEndpoint(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServerEndpoint, plan.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.RouteServerEndpoint == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServerEndpoint, plan.RouteServerId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.RouteServerEndpoint, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	createdEndpoint, err := waitVPCRouteServerEndpointCreated(ctx, conn, plan.RouteServerEndpointId.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCRouteServerEndpoint, plan.RouteServerEndpointId.String(), err),
			err.Error(),
		)
		return
	}
	//Flatten the created endpoint again to the plan to include eni_id and eni_address
	resp.Diagnostics.Append(flex.Flatten(ctx, createdEndpoint, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVPCRouteServerEndpoint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerEndpointModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCRouteServerEndpointByID(ctx, conn, state.RouteServerEndpointId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, ResNameVPCRouteServerEndpoint, state.RouteServerEndpointId.String(), err),
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

func (r *resourceVPCRouteServerEndpoint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerEndpointModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := findVPCRouteServerEndpointByID(ctx, conn, state.RouteServerEndpointId.ValueString())
	if err != nil {
		if tfresource.NotFound(err) {
			resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCRouteServerEndpoint, state.RouteServerEndpointId.String(), err),
			err.Error(),
		)
		return
	}
	input := ec2.DeleteRouteServerEndpointInput{
		RouteServerEndpointId: state.RouteServerEndpointId.ValueStringPointer(),
	}

	_, err = conn.DeleteRouteServerEndpoint(ctx, &input)
	if err != nil {
		if tfresource.NotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCRouteServerEndpoint, state.RouteServerEndpointId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCRouteServerEndpointDeleted(ctx, conn, state.RouteServerEndpointId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameVPCRouteServerEndpoint, state.RouteServerEndpointId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCRouteServerEndpoint) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitVPCRouteServerEndpointCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteServerEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RouteServerEndpointStatePending),
		Target:                    enum.Slice(awstypes.RouteServerEndpointStateAvailable),
		Refresh:                   statusVPCRouteServerEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RouteServerEndpoint); ok {
		return out, err
	}

	return nil, err
}

func waitVPCRouteServerEndpointDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteServerEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteServerStateDeleting),
		Target:  []string{},
		Refresh: statusVPCRouteServerEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RouteServerEndpoint); ok {
		return out, err
	}

	return nil, err
}

func statusVPCRouteServerEndpoint(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findVPCRouteServerEndpointByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func findVPCRouteServerEndpointByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.RouteServerEndpoint, error) {
	input := ec2.DescribeRouteServerEndpointsInput{
		RouteServerEndpointIds: []string{id},
	}
	var routeServerEndpoints []awstypes.RouteServerEndpoint
	paginator := ec2.NewDescribeRouteServerEndpointsPaginator(conn, &input)
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
		if page != nil && len(page.RouteServerEndpoints) > 0 {
			// until filters are implemented by API, this will check if each routeServerEndpoint in routeServerEndpoint state is not deleted
			for _, routeServerEndpoint := range page.RouteServerEndpoints {
				if routeServerEndpoint.State == awstypes.RouteServerEndpointStateDeleted {
					continue
				}
				routeServerEndpoints = append(routeServerEndpoints, routeServerEndpoint)
			}
		}
	}
	if len(routeServerEndpoints) == 0 {
		return nil, &retry.NotFoundError{
			LastError:   errors.New("route server not found"),
			LastRequest: &input,
		}
	} else {
		return &routeServerEndpoints[0], nil
	}
}

type resourceVPCRouteServerEndpointModel struct {
	RouteServerEndpointId types.String   `tfsdk:"id"`
	RouteServerId         types.String   `tfsdk:"route_server_id"`
	EniAddress            types.String   `tfsdk:"eni_address"`
	EniId                 types.String   `tfsdk:"eni_id"`
	SubnetId              types.String   `tfsdk:"subnet_id"`
	VPCId                 types.String   `tfsdk:"vpc_id"`
	Tags                  tftags.Map     `tfsdk:"tags"`
	TagsAll               tftags.Map     `tfsdk:"tags_all"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}
