// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_local_gateway_route_table_virtual_interface_group_association", name="Local Gateway Route Table Virtual Interface Group Association")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func newLocalGatewayRouteTableVirtualInterfaceGroupAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &localGatewayRouteTableVirtualInterfaceGroupAssociationResource{}, nil
}

type localGatewayRouteTableVirtualInterfaceGroupAssociationResource struct {
	framework.ResourceWithModel[localGatewayRouteTableVIFGroupAssociationModel]
	framework.WithImportByID
	framework.WithNoUpdate
}

func (r *localGatewayRouteTableVirtualInterfaceGroupAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"local_gateway_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"local_gateway_route_table_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"local_gateway_route_table_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"local_gateway_virtual_interface_group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *localGatewayRouteTableVirtualInterfaceGroupAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data localGatewayRouteTableVIFGroupAssociationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.CreateLocalGatewayRouteTableVirtualInterfaceGroupAssociationInput{
		LocalGatewayRouteTableId:            fwflex.StringFromFramework(ctx, data.LocalGatewayRouteTableId),
		LocalGatewayVirtualInterfaceGroupId: fwflex.StringFromFramework(ctx, data.LocalGatewayVirtualInterfaceGroupId),
		TagSpecifications:                   getTagSpecificationsIn(ctx, awstypes.ResourceTypeLocalGatewayRouteTableVirtualInterfaceGroupAssociation),
	}

	output, err := conn.CreateLocalGatewayRouteTableVirtualInterfaceGroupAssociation(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 Local Gateway Route Table Virtual Interface Group Association", err.Error())
		return
	}

	id := aws.ToString(output.LocalGatewayRouteTableVirtualInterfaceGroupAssociation.LocalGatewayRouteTableVirtualInterfaceGroupAssociationId)
	data.ID = types.StringValue(id)

	association, err := waitLocalGatewayRouteTableVIFGroupAssociationAssociated(ctx, conn, id)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Local Gateway Route Table Virtual Interface Group Association (%s) create", id), err.Error())
		return
	}

	data.LocalGatewayId = types.StringPointerValue(association.LocalGatewayId)
	data.LocalGatewayRouteTableArn = types.StringPointerValue(association.LocalGatewayRouteTableArn)
	data.OwnerID = types.StringPointerValue(association.OwnerId)
	data.State = types.StringPointerValue(association.State)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *localGatewayRouteTableVirtualInterfaceGroupAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data localGatewayRouteTableVIFGroupAssociationModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := data.ID.ValueString()
	association, err := findLocalGatewayRouteTableVIFGroupAssociationByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Local Gateway Route Table Virtual Interface Group Association (%s)", id), err.Error())
		return
	}

	data.LocalGatewayId = types.StringPointerValue(association.LocalGatewayId)
	data.LocalGatewayRouteTableArn = types.StringPointerValue(association.LocalGatewayRouteTableArn)
	data.LocalGatewayRouteTableId = types.StringPointerValue(association.LocalGatewayRouteTableId)
	data.LocalGatewayVirtualInterfaceGroupId = types.StringPointerValue(association.LocalGatewayVirtualInterfaceGroupId)
	data.OwnerID = types.StringPointerValue(association.OwnerId)
	data.State = types.StringPointerValue(association.State)

	setTagsOut(ctx, association.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *localGatewayRouteTableVirtualInterfaceGroupAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data localGatewayRouteTableVIFGroupAssociationModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := data.ID.ValueString()

	input := ec2.DeleteLocalGatewayRouteTableVirtualInterfaceGroupAssociationInput{
		LocalGatewayRouteTableVirtualInterfaceGroupAssociationId: aws.String(id),
	}
	_, err := conn.DeleteLocalGatewayRouteTableVirtualInterfaceGroupAssociation(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidLocalGatewayRouteTableVIFGroupAssociationIDNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Local Gateway Route Table Virtual Interface Group Association (%s)", id), err.Error())
		return
	}

	if _, err := waitLocalGatewayRouteTableVIFGroupAssociationDisassociated(ctx, conn, id); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Local Gateway Route Table Virtual Interface Group Association (%s) delete", id), err.Error())
		return
	}
}

type localGatewayRouteTableVIFGroupAssociationModel struct {
	framework.WithRegionModel
	ID                                  types.String `tfsdk:"id"`
	LocalGatewayId                      types.String `tfsdk:"local_gateway_id"`
	LocalGatewayRouteTableArn           types.String `tfsdk:"local_gateway_route_table_arn"`
	LocalGatewayRouteTableId            types.String `tfsdk:"local_gateway_route_table_id"`
	LocalGatewayVirtualInterfaceGroupId types.String `tfsdk:"local_gateway_virtual_interface_group_id"`
	OwnerID                             types.String `tfsdk:"owner_id"`
	State                               types.String `tfsdk:"state"`
	Tags                                tftags.Map   `tfsdk:"tags"`
	TagsAll                             tftags.Map   `tfsdk:"tags_all"`
}
