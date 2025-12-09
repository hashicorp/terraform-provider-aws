// Copyright IBM Corp. 2014, 2025
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
	"github.com/hashicorp/terraform-plugin-framework/path"
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
// @Tags(identifierAttribute="vpn_concentrator_id")
// @Testing(tagsTest=false)
func newVPNConcentratorResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &vpnConcentratorResource{}, nil
}

type vpnConcentratorResource struct {
	framework.ResourceWithModel[vpnConcentratorModel]
}

func (r *vpnConcentratorResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrTransitGatewayAttachmentID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTransitGatewayID: schema.StringAttribute{
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
			"vpn_concentrator_id": framework.IDAttribute(),
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

	input := ec2.CreateVpnConcentratorInput{
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpnConcentrator),
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateVpnConcentrator(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 VPN Concentrator", err.Error())
		return
	}

	id := aws.ToString(output.VpnConcentrator.VpnConcentratorId)
	data.VPNConcentratorID = fwflex.StringValueToFramework(ctx, id)

	vpnConcentrator, err := waitVPNConcentratorAvailable(ctx, conn, id)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 VPN Concentrator (%s) create", id), err.Error())
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

	id := fwflex.StringValueFromFramework(ctx, data.VPNConcentratorID)
	output, err := findVPNConcentratorByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 VPN Concentrator (%s)", id), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
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

	id := fwflex.StringValueFromFramework(ctx, data.VPNConcentratorID)
	input := ec2.DeleteVpnConcentratorInput{
		VpnConcentratorId: aws.String(id),
	}
	_, err := conn.DeleteVpnConcentrator(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNConcentratorIDNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 VPN Concentrator (%s)", id), err.Error())
		return
	}

	if _, err := waitVPNConcentratorDeleted(ctx, conn, id); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 VPN Concentrator (%s) delete", id), err.Error())
		return
	}

	if attachmentID := fwflex.StringValueFromFramework(ctx, data.TransitGatewayAttachmentID); attachmentID != "" {
		const (
			timeout = 10 * time.Minute
		)
		if _, err := waitTransitGatewayAttachmentDeleted(ctx, conn, attachmentID, timeout); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Attachment (%s) delete", attachmentID), err.Error())
			return
		}
	}
}

func (r *vpnConcentratorResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("vpn_concentrator_id"), request, response)
}

type vpnConcentratorModel struct {
	framework.WithRegionModel
	Tags                       tftags.Map                                       `tfsdk:"tags"`
	TagsAll                    tftags.Map                                       `tfsdk:"tags_all"`
	TransitGatewayAttachmentID types.String                                     `tfsdk:"transit_gateway_attachment_id"`
	TransitGatewayID           types.String                                     `tfsdk:"transit_gateway_id"`
	Type                       fwtypes.StringEnum[awstypes.VpnConcentratorType] `tfsdk:"type"`
	VPNConcentratorID          types.String                                     `tfsdk:"vpn_concentrator_id"`
}
