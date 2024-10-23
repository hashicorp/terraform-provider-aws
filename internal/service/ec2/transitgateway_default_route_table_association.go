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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_transit_gateway_default_route_table_association", name="Transit Gateway Default Route Table Association")
func newTransitGatewayDefaultRouteTableAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &transitGatewayDefaultRouteTableAssociationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type transitGatewayDefaultRouteTableAssociationResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (*transitGatewayDefaultRouteTableAssociationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_ec2_transit_gateway_default_route_table_association"
}

func (r *transitGatewayDefaultRouteTableAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"original_default_route_table_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *transitGatewayDefaultRouteTableAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data transitGatewayDefaultRouteTableAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	tgwID := data.TransitGatewayID.ValueString()
	tgw, err := findTransitGatewayByID(ctx, conn, tgwID)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Transit Gateway (%s)", tgwID), err.Error())

		return
	}

	input := &ec2.ModifyTransitGatewayInput{
		Options: &awstypes.ModifyTransitGatewayOptions{
			AssociationDefaultRouteTableId: fwflex.StringFromFramework(ctx, data.RouteTableID),
		},
		TransitGatewayId: aws.String(tgwID),
	}

	_, err = conn.ModifyTransitGateway(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating EC2 Transit Gateway Default Route Table Association (%s)", tgwID), err.Error())

		return
	}

	// Set unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, tgwID)
	data.OriginalDefaultRouteTableID = fwflex.StringToFramework(ctx, tgw.Options.AssociationDefaultRouteTableId)

	if _, err := waitTransitGatewayUpdated(ctx, conn, tgwID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Default Route Table Association (%s) create", tgwID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *transitGatewayDefaultRouteTableAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data transitGatewayDefaultRouteTableAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	tgw, err := findTransitGatewayByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Transit Gateway Default Route Table Association (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.RouteTableID = fwflex.StringToFramework(ctx, tgw.Options.AssociationDefaultRouteTableId)
	data.TransitGatewayID = fwflex.StringToFramework(ctx, tgw.TransitGatewayId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *transitGatewayDefaultRouteTableAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old transitGatewayDefaultRouteTableAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.ModifyTransitGatewayInput{
		Options: &awstypes.ModifyTransitGatewayOptions{
			AssociationDefaultRouteTableId: fwflex.StringFromFramework(ctx, new.RouteTableID),
		},
		TransitGatewayId: fwflex.StringFromFramework(ctx, new.TransitGatewayID),
	}

	_, err := conn.ModifyTransitGateway(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating EC2 Transit Gateway Default Route Table Association (%s)", new.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitTransitGatewayUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Default Route Table Association (%s) update", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *transitGatewayDefaultRouteTableAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data transitGatewayDefaultRouteTableAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	_, err := conn.ModifyTransitGateway(ctx, &ec2.ModifyTransitGatewayInput{
		Options: &awstypes.ModifyTransitGatewayOptions{
			AssociationDefaultRouteTableId: fwflex.StringFromFramework(ctx, data.OriginalDefaultRouteTableID),
		},
		TransitGatewayId: fwflex.StringFromFramework(ctx, data.TransitGatewayID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeIncorrectState) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Transit Gateway Default Route Table Association (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitTransitGatewayUpdated(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Default Route Table Association (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

type transitGatewayDefaultRouteTableAssociationResourceModel struct {
	ID                          types.String   `tfsdk:"id"`
	OriginalDefaultRouteTableID types.String   `tfsdk:"original_default_route_table_id"`
	RouteTableID                types.String   `tfsdk:"transit_gateway_route_table_id"`
	Timeouts                    timeouts.Value `tfsdk:"timeouts"`
	TransitGatewayID            types.String   `tfsdk:"transit_gateway_id"`
}
