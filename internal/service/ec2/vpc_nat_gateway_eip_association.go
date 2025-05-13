// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_nat_gateway_eip_association", name="VPC NAT Gateway EIP Association")
func newResourceNATGatewayEIPAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceNATGatewayEIPAssociation{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCNATGatewayEIPAssociation = "VPC NAT Gateway EIP Association"
)

type resourceNATGatewayEIPAssociation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceNATGatewayEIPAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allocation_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"association_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"nat_gateway_id": schema.StringAttribute{
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

func (r *resourceNATGatewayEIPAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCNATGatewayEIPAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := plan.setID()
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCNATGatewayEIPAssociation, "", err),
			err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(id)

	input := ec2.AssociateNatGatewayAddressInput{
		NatGatewayId:  plan.NATGatewayID.ValueStringPointer(),
		AllocationIds: []string{plan.AllocationID.ValueString()},
	}

	_, err = conn.AssociateNatGatewayAddress(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCNATGatewayEIPAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitNATGatewayAddressAssociated(ctx, conn, plan.NATGatewayID.ValueString(), plan.AllocationID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCNATGatewayEIPAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	plan.AssociationID = types.StringPointerValue(out.AssociationId)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceNATGatewayEIPAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCNATGatewayEIPAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := state.InitFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	out, err := findNATGatewayAddressByNATGatewayIDAndAllocationIDSucceeded(ctx, conn, state.NATGatewayID.ValueString(), state.AllocationID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, ResNameVPCNATGatewayEIPAssociation, state.ID.String(), err),
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

func (r *resourceNATGatewayEIPAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceNATGatewayEIPAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCNATGatewayEIPAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DisassociateNatGatewayAddressInput{
		NatGatewayId:   state.NATGatewayID.ValueStringPointer(),
		AssociationIds: []string{state.AssociationID.ValueString()},
	}

	_, err := conn.DisassociateNatGatewayAddress(ctx, &input)

	if tfawserr.ErrCodeEquals(err, "InvalidParameter") {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCNATGatewayEIPAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitNATGatewayAddressDisassociated(ctx, conn, state.NATGatewayID.ValueString(), state.AllocationID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameVPCNATGatewayEIPAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceNATGatewayEIPAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

type resourceVPCNATGatewayEIPAssociationModel struct {
	AllocationID  types.String   `tfsdk:"allocation_id"`
	AssociationID types.String   `tfsdk:"association_id"`
	ID            types.String   `tfsdk:"id"`
	NATGatewayID  types.String   `tfsdk:"nat_gateway_id"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

const (
	natGatewayEIPAssociationParts = 2
)

func (m *resourceVPCNATGatewayEIPAssociationModel) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := fwflex.ExpandResourceId(id, natGatewayEIPAssociationParts, false)

	if err != nil {
		return err
	}

	m.NATGatewayID = types.StringValue(parts[0])
	m.AllocationID = types.StringValue(parts[1])

	return nil
}

func (m resourceVPCNATGatewayEIPAssociationModel) setID() (string, error) {
	parts := []string{
		m.NATGatewayID.ValueString(),
		m.AllocationID.ValueString(),
	}

	return fwflex.FlattenResourceId(parts, natGatewayEIPAssociationParts, false)
}
