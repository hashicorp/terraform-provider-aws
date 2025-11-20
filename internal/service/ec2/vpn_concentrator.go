// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpn_concentrator",name="VPN Concentrator")
// @Tags(identifierAttribute="id")
func newVPNConcentratorResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &vpnConcentratorResource{}, nil
}

type vpnConcentratorResource struct {
	framework.ResourceWithModel[vpnConcentratorModel]
	framework.WithImportByID
}

func (r *vpnConcentratorResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrState: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"transit_gateway_attachment_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"transit_gateway_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VpnConcentratorType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *vpnConcentratorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpnConcentratorModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.CreateVpnConcentratorInput{
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpnConcentrator),
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateVpnConcentrator(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 VPN Concentrator", err.Error())
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.VpnConcentrator.VpnConcentratorId)

	vpnConcentrator, err := waitVPNConcentratorAvailable(ctx, conn, data.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 VPN Concentrator (%s) create", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, vpnConcentrator, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpnConcentratorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpnConcentratorModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	vpnConcentrator, err := findVPNConcentratorByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 VPN Concentrator (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, vpnConcentrator, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpnConcentratorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpnConcentratorModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	_, err := conn.DeleteVpnConcentrator(ctx, &ec2.DeleteVpnConcentratorInput{
		VpnConcentratorId: fwflex.StringFromFramework(ctx, data.ID),
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 VPN Concentrator (%s)", data.ID.ValueString()), err.Error())
		return
	}

	if err := waitVPNConcentratorDeleted(ctx, conn, data.ID.ValueString(), data.TransitGatewayAttachmentID.ValueString()); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 VPN Concentrator (%s) delete", data.ID.ValueString()), err.Error())
		return
	}
}

func findVPNConcentratorByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnConcentrator, error) {
	input := &ec2.DescribeVpnConcentratorsInput{
		VpnConcentratorIds: []string{id},
	}

	output, err := conn.DescribeVpnConcentrators(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpnConcentrators) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	vpnConcentrator := &output.VpnConcentrators[0]

	if state := aws.ToString(vpnConcentrator.State); state == "deleted" {
		return nil, &tfresource.EmptyResultError{
			LastRequest: input,
		}
	}

	return vpnConcentrator, nil
}

type vpnConcentratorModel struct {
	framework.WithRegionModel
	ID                         types.String                                     `tfsdk:"id"`
	State                      types.String                                     `tfsdk:"state"`
	Tags                       tftags.Map                                       `tfsdk:"tags"`
	TagsAll                    tftags.Map                                       `tfsdk:"tags_all"`
	TransitGatewayAttachmentID types.String                                     `tfsdk:"transit_gateway_attachment_id"`
	TransitGatewayID           types.String                                     `tfsdk:"transit_gateway_id"`
	Type                       fwtypes.StringEnum[awstypes.VpnConcentratorType] `tfsdk:"type"`
}
