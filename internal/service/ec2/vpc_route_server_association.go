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
// @FrameworkResource("aws_vpc_route_server_association", name="VPC Route Server Association")
func newVPCRouteServerAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCRouteServerAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCRouteServerAssociation          = "VPC Route Server Association"
	attributeVPCRouteServerAssociationIDParts = 2
)

type resourceVPCRouteServerAssociation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
	framework.WithNoUpdate
}

func (r *resourceVPCRouteServerAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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

func (r *resourceVPCRouteServerAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCRouteServerAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.AssociateRouteServerInput

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.AssociateRouteServer(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServerAssociation, plan.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.RouteServerAssociation == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServerAssociation, plan.RouteServerId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitVPCRouteServerAssociationCreated(ctx, conn, plan.RouteServerId.ValueString(), plan.VpcId.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCRouteServerAssociation, plan.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVPCRouteServerAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCRouteServerAssociationByTwoPartKey(ctx, conn, state.RouteServerId.ValueString(), state.VpcId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, ResNameVPCRouteServerAssociation, state.RouteServerId.String(), err),
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

func (r *resourceVPCRouteServerAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the resource is already deleted
	_, err := findVPCRouteServerAssociationByTwoPartKey(ctx, conn, state.RouteServerId.ValueString(), state.VpcId.ValueString())
	if err != nil {
		if tfresource.NotFound(err) {
			resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCRouteServer, state.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}

	input := ec2.DisassociateRouteServerInput{
		RouteServerId: state.RouteServerId.ValueStringPointer(),
		VpcId:         state.VpcId.ValueStringPointer(),
	}

	_, err = conn.DisassociateRouteServer(ctx, &input)
	if err != nil {
		if tfresource.NotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCRouteServerAssociation, state.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCRouteServerAssociationDeleted(ctx, conn, state.RouteServerId.ValueString(), state.VpcId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameVPCRouteServerAssociation, state.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCRouteServerAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, attributeVPCRouteServerAssociationIDParts, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: route_server_id,vpc_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("route_server_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrVPCID), parts[1])...)
}

func waitVPCRouteServerAssociationCreated(ctx context.Context, conn *ec2.Client, id string, vpcId string, timeout time.Duration) (*awstypes.RouteServerAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RouteServerAssociationStateAssociating),
		Target:                    enum.Slice(awstypes.RouteServerAssociationStateAssociated),
		Refresh:                   statusVPCRouteServerAssociation(ctx, conn, id, vpcId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RouteServerAssociation); ok {
		return out, err
	}

	return nil, err
}

func waitVPCRouteServerAssociationDeleted(ctx context.Context, conn *ec2.Client, id string, vpcId string, timeout time.Duration) (*awstypes.RouteServerAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteServerAssociationStateDisassociating),
		Target:  []string{},
		Refresh: statusVPCRouteServerAssociation(ctx, conn, id, vpcId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RouteServerAssociation); ok {
		return out, err
	}

	return nil, err
}

func statusVPCRouteServerAssociation(ctx context.Context, conn *ec2.Client, id string, vpcId string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findVPCRouteServerAssociationByTwoPartKey(ctx, conn, id, vpcId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func findVPCRouteServerAssociationByTwoPartKey(ctx context.Context, conn *ec2.Client, id string, vpcId string) (*awstypes.RouteServerAssociation, error) {
	input := ec2.GetRouteServerAssociationsInput{
		RouteServerId: aws.String(id),
	}

	out, err := conn.GetRouteServerAssociations(ctx, &input)
	if err != nil {
		if tfresource.NotFound(err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
	}
	if out == nil || out.RouteServerAssociations == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}
	//loop through the associations and find the one with the matching vpcId
	for _, association := range out.RouteServerAssociations {
		if aws.ToString(association.VpcId) == vpcId {
			return &association, nil
		}
	}
	return nil, tfresource.NewEmptyResultError(&input)
}

type resourceVPCRouteServerAssociationModel struct {
	RouteServerId types.String   `tfsdk:"route_server_id"`
	VpcId         types.String   `tfsdk:"vpc_id"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}
