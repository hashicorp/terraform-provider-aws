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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_nat_gateway_eip_association", name="VPC NAT Gateway EIP Association")
func newNATGatewayEIPAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &natGatewayEIPAssociationResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type natGatewayEIPAssociationResource struct {
	framework.ResourceWithModel[natGatewayEIPAssociationResourceModel]
	framework.WithNoUpdate
	framework.WithTimeouts
}

func (r *natGatewayEIPAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allocation_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrAssociationID: schema.StringAttribute{
				Computed: true,
			},
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
				Delete: true,
			}),
		},
	}
}

func (r *natGatewayEIPAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data natGatewayEIPAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	natGatewayID, allocationID := fwflex.StringValueFromFramework(ctx, data.NATGatewayID), fwflex.StringValueFromFramework(ctx, data.AllocationID)
	input := ec2.AssociateNatGatewayAddressInput{
		AllocationIds: []string{allocationID},
		NatGatewayId:  aws.String(natGatewayID),
	}

	_, err := conn.AssociateNatGatewayAddress(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating VPC NAT Gateway (%s) EIP (%s) Association", natGatewayID, allocationID), err.Error())

		return
	}

	output, err := waitNATGatewayAddressAssociated(ctx, conn, natGatewayID, allocationID, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC NAT Gateway (%s) EIP (%s) Association create", natGatewayID, allocationID), err.Error())

		return
	}

	// Set values for unknowns.
	data.AssociationID = fwflex.StringToFramework(ctx, output.AssociationId)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *natGatewayEIPAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data natGatewayEIPAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	natGatewayID, allocationID := fwflex.StringValueFromFramework(ctx, data.NATGatewayID), fwflex.StringValueFromFramework(ctx, data.AllocationID)
	output, err := findNATGatewayAddressByNATGatewayIDAndAllocationIDSucceeded(ctx, conn, natGatewayID, allocationID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC NAT Gateway (%s) EIP (%s) Association", natGatewayID, allocationID), err.Error())

		return
	}

	// Set attributes for import.
	data.AssociationID = fwflex.StringToFramework(ctx, output.AssociationId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *natGatewayEIPAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data natGatewayEIPAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	natGatewayID, allocationID, associationID := fwflex.StringValueFromFramework(ctx, data.NATGatewayID), fwflex.StringValueFromFramework(ctx, data.AllocationID), fwflex.StringValueFromFramework(ctx, data.AssociationID)
	input := ec2.DisassociateNatGatewayAddressInput{
		AssociationIds: []string{associationID},
		NatGatewayId:   aws.String(natGatewayID),
	}
	_, err := conn.DisassociateNatGatewayAddress(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidParameter) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC NAT Gateway (%s) EIP (%s) Association (%s)", natGatewayID, allocationID, associationID), err.Error())

		return
	}

	if _, err := waitNATGatewayAddressDisassociated(ctx, conn, data.NATGatewayID.ValueString(), data.AllocationID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC NAT Gateway (%s) EIP (%s) Association delete", natGatewayID, allocationID), err.Error())

		return
	}
}

func (r *natGatewayEIPAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		natGatewayEIPAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, natGatewayEIPAssociationIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("nat_gateway_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("allocation_id"), parts[1])...)
}

type natGatewayEIPAssociationResourceModel struct {
	framework.WithRegionModel
	AllocationID  types.String   `tfsdk:"allocation_id"`
	AssociationID types.String   `tfsdk:"association_id"`
	NATGatewayID  types.String   `tfsdk:"nat_gateway_id"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}
