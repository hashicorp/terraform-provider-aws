// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_vpc_route_server_propagation", name="VPC Route Server Propagation")
func newVPCRouteServerPropagationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCRouteServerPropagation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCRouteServerPropagation          = "VPC Route Server Propagation"
	attributeVPCRouteServerPropagationIDParts = 2
)

type resourceVPCRouteServerPropagation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
	framework.WithNoUpdate
}

func (r *resourceVPCRouteServerPropagation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceVPCRouteServerPropagation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCRouteServerPropagationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.EnableRouteServerPropagationInput

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.EnableRouteServerPropagation(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServerPropagation, plan.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.RouteServerPropagation == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServerPropagation, plan.RouteServerId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitVPCRouteServerPropagationCreated(ctx, conn, plan.RouteServerId.ValueString(), plan.RouteTableId.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCRouteServerPropagation, plan.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVPCRouteServerPropagation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerPropagationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCRouteServerPropagationByTwoPartKey(ctx, conn, state.RouteServerId.ValueString(), state.RouteTableId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, ResNameVPCRouteServerPropagation, state.RouteServerId.String(), err),
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

func (r *resourceVPCRouteServerPropagation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerPropagationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the resource is already deleted
	_, err := findVPCRouteServerPropagationByTwoPartKey(ctx, conn, state.RouteServerId.ValueString(), state.RouteTableId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	input := ec2.DisableRouteServerPropagationInput{
		RouteServerId: aws.String(state.RouteServerId.ValueString()),
		RouteTableId:  aws.String(state.RouteTableId.ValueString()),
	}

	_, err = conn.DisableRouteServerPropagation(ctx, &input)
	if err != nil {
		if tfresource.NotFound(err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCRouteServerPropagation, state.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCRouteServerPropagationDeleted(ctx, conn, state.RouteServerId.ValueString(), state.RouteTableId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameVPCRouteServerPropagation, state.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCRouteServerPropagation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, attributeVPCRouteServerPropagationIDParts, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: route_server_id,route_table_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("route_server_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("route_table_id"), parts[1])...)
}

func waitVPCRouteServerPropagationCreated(ctx context.Context, conn *ec2.Client, id string, routeTableId string, timeout time.Duration) (*awstypes.RouteServerPropagation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RouteServerPropagationStatePending),
		Target:                    enum.Slice(awstypes.RouteServerPropagationStateAvailable),
		Refresh:                   statusVPCRouteServerPropagation(ctx, conn, id, routeTableId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RouteServerPropagation); ok {
		return out, err
	}

	return nil, err
}

func waitVPCRouteServerPropagationDeleted(ctx context.Context, conn *ec2.Client, id string, routeTableId string, timeout time.Duration) (*awstypes.RouteServerPropagation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteServerPropagationStateDeleting),
		Target:  []string{},
		Refresh: statusVPCRouteServerPropagation(ctx, conn, id, routeTableId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RouteServerPropagation); ok {
		return out, err
	}

	return nil, err
}

func statusVPCRouteServerPropagation(ctx context.Context, conn *ec2.Client, id string, routeTableId string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findVPCRouteServerPropagationByTwoPartKey(ctx, conn, id, routeTableId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func findVPCRouteServerPropagationByTwoPartKey(ctx context.Context, conn *ec2.Client, id string, routeTableId string) (*awstypes.RouteServerPropagation, error) {
	input := ec2.GetRouteServerPropagationsInput{
		RouteServerId: aws.String(id),
		RouteTableId:  aws.String(routeTableId),
	}

	out, err := conn.GetRouteServerPropagations(ctx, &input)
	if err != nil {
		if tfresource.NotFound(err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
	}
	if out == nil || out.RouteServerPropagations == nil || len(out.RouteServerPropagations) == 0 {
		return nil, tfresource.NewEmptyResultError(&input)
	} else {
		return &out.RouteServerPropagations[0], nil
	}
}

type resourceVPCRouteServerPropagationModel struct {
	RouteServerId types.String   `tfsdk:"route_server_id"`
	RouteTableId  types.String   `tfsdk:"route_table_id"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}
