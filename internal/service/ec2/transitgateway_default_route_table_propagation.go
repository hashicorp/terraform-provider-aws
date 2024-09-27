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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_ec2_transit_gateway_default_route_table_propagation", name="Transit Gateway Default Route Table Propagation")
func newResourceTransitGatewayDefaultRouteTablePropagation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTransitGatewayDefaultRouteTablePropagation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameTransitGatewayDefaultRouteTablePropagation = "Transit Gateway Default Route Table Propagation"
)

type resourceTransitGatewayDefaultRouteTablePropagation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceTransitGatewayDefaultRouteTablePropagation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ec2_transit_gateway_default_route_table_propagation"
}

func (r *resourceTransitGatewayDefaultRouteTablePropagation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"original_route_table_id": schema.StringAttribute{
				Computed: true,
			},
			"transit_gateway_route_table_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrTransitGatewayID: schema.StringAttribute{
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

func (r *resourceTransitGatewayDefaultRouteTablePropagation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan transitgatewayDefaultRouteTablePropagationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tgw, err := findTransitGatewayByID(ctx, conn, plan.TransitGatewayId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameTransitGatewayDefaultRouteTablePropagation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	in := &ec2.ModifyTransitGatewayInput{
		TransitGatewayId: flex.StringFromFramework(ctx, plan.TransitGatewayId),
		Options: &awstypes.ModifyTransitGatewayOptions{
			PropagationDefaultRouteTableId: flex.StringFromFramework(ctx, plan.RouteTableId),
		},
	}

	out, err := conn.ModifyTransitGateway(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameTransitGatewayDefaultRouteTablePropagation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameTransitGatewayDefaultRouteTablePropagation, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.TransitGateway.TransitGatewayId)
	plan.OriginalRouteTableId = flex.StringToFramework(ctx, tgw.Options.PropagationDefaultRouteTableId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitTransitGatewayUpdated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameTransitGatewayDefaultRouteTablePropagation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTransitGatewayDefaultRouteTablePropagation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state transitgatewayDefaultRouteTablePropagationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTransitGatewayByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameTransitGatewayDefaultRouteTablePropagation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.TransitGatewayId)
	state.TransitGatewayId = flex.StringToFramework(ctx, out.TransitGatewayId)
	state.RouteTableId = flex.StringToFramework(ctx, out.Options.PropagationDefaultRouteTableId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTransitGatewayDefaultRouteTablePropagation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan, state transitgatewayDefaultRouteTablePropagationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.RouteTableId.Equal(state.RouteTableId) {
		in := &ec2.ModifyTransitGatewayInput{
			TransitGatewayId: state.ID.ValueStringPointer(),
			Options: &awstypes.ModifyTransitGatewayOptions{
				PropagationDefaultRouteTableId: plan.RouteTableId.ValueStringPointer(),
			},
		}

		out, err := conn.ModifyTransitGateway(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionUpdating, ResNameTransitGatewayDefaultRouteTablePropagation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionUpdating, ResNameTransitGatewayDefaultRouteTablePropagation, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitTransitGatewayUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForUpdate, ResNameTransitGatewayDefaultRouteTablePropagation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTransitGatewayDefaultRouteTablePropagation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state transitgatewayDefaultRouteTablePropagationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ec2.ModifyTransitGatewayInput{
		TransitGatewayId: flex.StringFromFramework(ctx, state.TransitGatewayId),
		Options: &awstypes.ModifyTransitGatewayOptions{
			PropagationDefaultRouteTableId: flex.StringFromFramework(ctx, state.OriginalRouteTableId),
		},
	}

	_, err := conn.ModifyTransitGateway(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameTransitGatewayDefaultRouteTablePropagation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitTransitGatewayUpdated(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameTransitGatewayDefaultRouteTablePropagation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

type transitgatewayDefaultRouteTablePropagationResourceModel struct {
	ID                   types.String   `tfsdk:"id"`
	OriginalRouteTableId types.String   `tfsdk:"original_route_table_id"`
	RouteTableId         types.String   `tfsdk:"transit_gateway_route_table_id"`
	TransitGatewayId     types.String   `tfsdk:"transit_gateway_id"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}
