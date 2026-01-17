// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_firewall_transit_gateway_attachment_accepter", name="Firewall Transit Gateway Attachment Accepter")
func newFirewallTransitGatewayAttachmentAccepterResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &firewallTransitGatewayAttachmentAccepterResource{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(60 * time.Minute)

	return r, nil
}

type firewallTransitGatewayAttachmentAccepterResource struct {
	framework.ResourceWithModel[firewallTransitGatewayAttachmentAccepterResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *firewallTransitGatewayAttachmentAccepterResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrTransitGatewayAttachmentID: schema.StringAttribute{
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

func (r *firewallTransitGatewayAttachmentAccepterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data firewallTransitGatewayAttachmentAccepterResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFirewallClient(ctx)

	tgwAttachmentID := fwflex.StringValueFromFramework(ctx, data.TransitGatewayAttachmentID)
	input := networkfirewall.AcceptNetworkFirewallTransitGatewayAttachmentInput{
		TransitGatewayAttachmentId: aws.String(tgwAttachmentID),
	}

	_, err := conn.AcceptNetworkFirewallTransitGatewayAttachment(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("accepting NetworkFirewall Firewall Transit Gateway Attachment (%s)", tgwAttachmentID), err.Error())

		return
	}

	if _, err := tfec2.WaitTransitGatewayAttachmentAccepted(ctx, r.Meta().EC2Client(ctx), tgwAttachmentID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall Firewall Transit Gateway Attachment (%s) accept", tgwAttachmentID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *firewallTransitGatewayAttachmentAccepterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data firewallTransitGatewayAttachmentAccepterResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tgwAttachmentID := fwflex.StringValueFromFramework(ctx, data.TransitGatewayAttachmentID)
	output, err := tfec2.FindTransitGatewayAttachmentByID(ctx, r.Meta().EC2Client(ctx), tgwAttachmentID)

	if err == nil && output.State == ec2types.TransitGatewayAttachmentStateDeleted {
		err = tfresource.NewEmptyResultError()
	}

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading NetworkFirewall Firewall Transit Gateway Attachment (%s)", tgwAttachmentID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *firewallTransitGatewayAttachmentAccepterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data firewallTransitGatewayAttachmentAccepterResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFirewallClient(ctx)

	tgwAttachmentID := fwflex.StringValueFromFramework(ctx, data.TransitGatewayAttachmentID)
	input := networkfirewall.DeleteNetworkFirewallTransitGatewayAttachmentInput{
		TransitGatewayAttachmentId: aws.String(tgwAttachmentID),
	}

	_, err := conn.DeleteNetworkFirewallTransitGatewayAttachment(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting NetworkFirewall Firewall Transit Gateway Attachment (%s)", tgwAttachmentID), err.Error())

		return
	}

	if _, err := tfec2.WaitTransitGatewayAttachmentDeleted(ctx, r.Meta().EC2Client(ctx), tgwAttachmentID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall Firewall Transit Gateway Attachment (%s) delete", tgwAttachmentID), err.Error())

		return
	}
}

func (r *firewallTransitGatewayAttachmentAccepterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrTransitGatewayAttachmentID), request, response)
}

type firewallTransitGatewayAttachmentAccepterResourceModel struct {
	framework.WithRegionModel
	TransitGatewayAttachmentID types.String   `tfsdk:"transit_gateway_attachment_id"`
	Timeouts                   timeouts.Value `tfsdk:"timeouts"`
}
