// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_network_firewall_transit_gateway_attachment_accepter", name="Network Firewall Transit Gateway Attachment Accepter")
func newResourceNetworkFirewallTransitGatewayAttachmentAccepter(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceNetworkFirewallTransitGatewayAttachmentAccepter{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(60 * time.Minute)

	return r, nil
}

const (
	ResNameNetworkFirewallTransitGatewayAttachmentAccepter = "Network Firewall Transit Gateway Attachment Accepter"
)

type resourceNetworkFirewallTransitGatewayAttachmentAccepter struct {
	framework.ResourceWithModel[resourceNetworkFirewallTransitGatewayAttachmentAccepterModel]
	framework.WithTimeouts
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *resourceNetworkFirewallTransitGatewayAttachmentAccepter) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
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

func (r *resourceNetworkFirewallTransitGatewayAttachmentAccepter) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	conn := r.Meta().NetworkFirewallClient(ctx)
	ec2Conn := r.Meta().EC2Client(ctx)

	var plan resourceNetworkFirewallTransitGatewayAttachmentAccepterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input networkfirewall.AcceptNetworkFirewallTransitGatewayAttachmentInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.AcceptNetworkFirewallTransitGatewayAttachment(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionCreating, ResNameNetworkFirewallTransitGatewayAttachmentAccepter, plan.TransitGatewayAttachmentId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.TransitGatewayAttachmentId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionCreating, ResNameNetworkFirewallTransitGatewayAttachmentAccepter, plan.TransitGatewayAttachmentId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitNetworkFirewallTransitGatewayAttachmentAccepterCreated(ctx, ec2Conn, fwflex.StringValueFromFramework(ctx, plan.TransitGatewayAttachmentId), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForCreation, ResNameNetworkFirewallTransitGatewayAttachmentAccepter, plan.TransitGatewayAttachmentId.String(), err),
			err.Error(),
		)
		return
	}
	plan.ID = flex.StringToFramework(ctx, out.TransitGatewayAttachmentId)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceNetworkFirewallTransitGatewayAttachmentAccepter) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ec2Conn := r.Meta().EC2Client(ctx)

	var state resourceNetworkFirewallTransitGatewayAttachmentAccepterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNetworkFirewallTransitGatewayAttachmentAccepterById(ctx, ec2Conn, fwflex.StringValueFromFramework(ctx, state.TransitGatewayAttachmentId))
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionReading, ResNameNetworkFirewallTransitGatewayAttachmentAccepter, fwflex.StringValueFromFramework(ctx, state.TransitGatewayAttachmentId), err),
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

func (r *resourceNetworkFirewallTransitGatewayAttachmentAccepter) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)
	ec2Conn := r.Meta().EC2Client(ctx)

	var state resourceNetworkFirewallTransitGatewayAttachmentAccepterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := networkfirewall.DeleteNetworkFirewallTransitGatewayAttachmentInput{
		TransitGatewayAttachmentId: aws.String(state.TransitGatewayAttachmentId.ValueString()),
	}

	_, err := conn.DeleteNetworkFirewallTransitGatewayAttachment(ctx, &input)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionDeleting, ResNameNetworkFirewallTransitGatewayAttachmentAccepter, fwflex.StringValueFromFramework(ctx, state.TransitGatewayAttachmentId), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitNetworkFirewallTransitGatewayAttachmentAccepterDeleted(ctx, ec2Conn, fwflex.StringValueFromFramework(ctx, state.TransitGatewayAttachmentId), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForDeletion, ResNameNetworkFirewallTransitGatewayAttachmentAccepter, fwflex.StringValueFromFramework(ctx, state.TransitGatewayAttachmentId), err),
			err.Error(),
		)
		return
	}
}

func waitNetworkFirewallTransitGatewayAttachmentAccepterCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*ec2awstypes.TransitGatewayAttachmentState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(ec2awstypes.TransitGatewayAttachmentStatePending, ec2awstypes.TransitGatewayAttachmentStatePendingAcceptance),
		Target:                    enum.Slice(ec2awstypes.TransitGatewayAttachmentStateAvailable),
		Refresh:                   statusNetworkFirewallTransitGatewayAttachmentAccepter(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ec2awstypes.TransitGatewayAttachment); ok {
		return &out.State, err
	}

	return nil, err
}

func waitNetworkFirewallTransitGatewayAttachmentAccepterDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayAttachmentStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(ec2awstypes.TransitGatewayAttachmentStateDeleting, ec2awstypes.TransitGatewayAttachmentStateAvailable),
		Target:  enum.Slice(ec2awstypes.TransitGatewayAttachmentStateDeleted),
		Refresh: statusNetworkFirewallTransitGatewayAttachmentAccepter(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok {
		return &out.FirewallStatus.TransitGatewayAttachmentSyncState.TransitGatewayAttachmentStatus, err
	}

	return nil, err
}

func statusNetworkFirewallTransitGatewayAttachmentAccepter(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(context.Context) (any, string, error) {

		out, err := findNetworkFirewallTransitGatewayAttachmentAccepterById(ctx, conn, id)
		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func findNetworkFirewallTransitGatewayAttachmentAccepterById(ctx context.Context, conn *ec2.Client, id string) (*ec2awstypes.TransitGatewayAttachment, error) {
	out, err := tfec2.FindTransitGatewayAttachmentByID(ctx, conn, id)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, &retry.NotFoundError{
			Message: "Network Firewall Transit Gateway Attachment Accepter not found",
		}
	}

	return out, nil
}

type resourceNetworkFirewallTransitGatewayAttachmentAccepterModel struct {
	framework.WithRegionModel
	ID                         types.String   `tfsdk:"id"`
	TransitGatewayAttachmentId types.String   `tfsdk:"transit_gateway_attachment_id"`
	Timeouts                   timeouts.Value `tfsdk:"timeouts"`
}
