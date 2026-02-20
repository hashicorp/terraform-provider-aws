// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_networkmanager_attachment_routing_policy_label", name="Attachment Routing Policy Label")
func newAttachmentRoutingPolicyLabelResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &attachmentRoutingPolicyLabelResource{}, nil
}

const (
	attachmentRoutingPolicyLabelAvailableTimeout = 20 * time.Minute
)

type attachmentRoutingPolicyLabelResource struct {
	framework.ResourceWithModel[attachmentRoutingPolicyLabelResourceModel]
	framework.WithNoUpdate
}

type attachmentRoutingPolicyLabelResourceModel struct {
	AttachmentID       types.String `tfsdk:"attachment_id"`
	CoreNetworkID      types.String `tfsdk:"core_network_id"`
	RoutingPolicyLabel types.String `tfsdk:"routing_policy_label"`
}

func (r *attachmentRoutingPolicyLabelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"attachment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"core_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"routing_policy_label": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *attachmentRoutingPolicyLabelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan attachmentRoutingPolicyLabelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	coreNetworkID, attachmentID := fwflex.StringValueFromFramework(ctx, plan.CoreNetworkID), fwflex.StringValueFromFramework(ctx, plan.AttachmentID)
	input := networkmanager.PutAttachmentRoutingPolicyLabelInput{
		AttachmentId:       aws.String(attachmentID),
		ClientToken:        aws.String(sdkid.UniqueId()),
		CoreNetworkId:      aws.String(coreNetworkID),
		RoutingPolicyLabel: fwflex.StringFromFramework(ctx, plan.RoutingPolicyLabel),
	}
	_, err := conn.PutAttachmentRoutingPolicyLabel(ctx, &input)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("creating Network Manager Attachment Routing Policy Label (%s/%s)", coreNetworkID, attachmentID),
			err.Error(),
		)
		return
	}

	if _, err := waitAttachmentAvailable(ctx, conn, coreNetworkID, attachmentID, attachmentRoutingPolicyLabelAvailableTimeout); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("waiting for Network Manager Attachment (%s) to become available after applying routing policy label", attachmentID),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *attachmentRoutingPolicyLabelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state attachmentRoutingPolicyLabelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	coreNetworkID, attachmentID := fwflex.StringValueFromFramework(ctx, state.CoreNetworkID), fwflex.StringValueFromFramework(ctx, state.AttachmentID)
	label, err := findAttachmentRoutingPolicyAssociationLabelByTwoPartKey(ctx, conn, coreNetworkID, attachmentID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("reading Network Manager Attachment Routing Policy Label (%s/%s)", coreNetworkID, attachmentID),
			err.Error(),
		)
		return
	}

	state.RoutingPolicyLabel = fwflex.StringToFramework(ctx, label)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *attachmentRoutingPolicyLabelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state attachmentRoutingPolicyLabelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	coreNetworkID, attachmentID := fwflex.StringValueFromFramework(ctx, state.CoreNetworkID), fwflex.StringValueFromFramework(ctx, state.AttachmentID)
	if _, err := waitAttachmentAvailable(ctx, conn, coreNetworkID, attachmentID, attachmentRoutingPolicyLabelAvailableTimeout); err != nil {
		// If the attachment itself is gone, nothing to delete.
		if !retry.NotFound(err) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("waiting for Network Manager Attachment (%s) to become available before removing routing policy label", attachmentID),
				err.Error(),
			)
		}
		return
	}

	input := networkmanager.RemoveAttachmentRoutingPolicyLabelInput{
		AttachmentId:  aws.String(attachmentID),
		CoreNetworkId: aws.String(coreNetworkID),
	}
	_, err := conn.RemoveAttachmentRoutingPolicyLabel(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("deleting Network Manager Attachment Routing Policy Label (%s/%s)", coreNetworkID, attachmentID),
			err.Error(),
		)
		return
	}

	if _, err := waitAttachmentAvailable(ctx, conn, coreNetworkID, attachmentID, attachmentRoutingPolicyLabelAvailableTimeout); err != nil {
		// If the attachment itself is gone after remove, that's fine.
		if !retry.NotFound(err) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("waiting for Network Manager Attachment (%s) to become available after removing routing policy label", attachmentID),
				err.Error(),
			)
		}
		return
	}
}

func (r *attachmentRoutingPolicyLabelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	const (
		attachmentRoutingPolicyLabelIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(req.ID, attachmentRoutingPolicyLabelIDParts, true)

	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("core_network_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attachment_id"), parts[1])...)
}

func findAttachments(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListAttachmentsInput) ([]awstypes.Attachment, error) {
	var output []awstypes.Attachment

	pages := networkmanager.NewListAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Attachments...)
	}

	return output, nil
}

func findAttachmentByTwoPartKey(ctx context.Context, conn *networkmanager.Client, coreNetworkID, attachmentID string) (*awstypes.Attachment, error) {
	input := networkmanager.ListAttachmentsInput{
		CoreNetworkId: aws.String(coreNetworkID),
	}
	output, err := findAttachments(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output, func(v awstypes.Attachment) bool {
		return aws.ToString(v.AttachmentId) == attachmentID
	}))
}

func statusAttachment(conn *networkmanager.Client, coreNetworkID, attachmentID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findAttachmentByTwoPartKey(ctx, conn, coreNetworkID, attachmentID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitAttachmentAvailable(ctx context.Context, conn *networkmanager.Client, coreNetworkID, attachmentID string, timeout time.Duration) (*awstypes.Attachment, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatePendingNetworkUpdate, awstypes.AttachmentStateUpdating),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable),
		Timeout: timeout,
		Refresh: statusAttachment(conn, coreNetworkID, attachmentID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Attachment); ok {
		return output, err
	}

	return nil, err
}
