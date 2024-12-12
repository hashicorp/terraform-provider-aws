// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="DX Gateway Attachment")
func newResourceDXGatewayAttachment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDXGatewayAttachment{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDXGatewayAttachment = "DX Gateway Attachment"
)

type resourceDXGatewayAttachment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDXGatewayAttachment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_networkmanager_dx_gateway_attachment"
}

func (r *resourceDXGatewayAttachment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"attachment_policy_rule_number": schema.Int64Attribute{
				Computed: true,
			},
			"attachment_type": schema.StringAttribute{
				Computed: true,
			},
			"core_network_arn": schema.StringAttribute{
				Computed: true,
			},
			"core_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 50),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^core-network-([0-9a-f]{8,17})$`),
						"must be a valid core network id",
					),
				},
			},
			"direct_connect_gateway_arn": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 500),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^arn:[^:]{1,63}:directconnect::[^:]{0,63}:dx-gateway\/[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}$`),
						"must be a valid Direct Connect gateway ARN",
					),
				},
			},
			"edge_locations": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"id": framework.IDAttribute(),
			"network_function_group_name": schema.StringAttribute{
				Computed: true,
			},
			"owner_account_id": schema.StringAttribute{
				Computed: true,
			},
			"segment_name": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDXGatewayAttachment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkManagerClient(ctx)

	var plan resourceDXGatewayAttachmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &networkmanager.CreateDirectConnectGatewayAttachmentInput{
		CoreNetworkId:           aws.String(plan.CoreNetworkID.ValueString()),
		DirectConnectGatewayArn: aws.String(plan.DirectConnectGatewayARN.ValueString()),
		EdgeLocations:           flex.ExpandFrameworkStringValueList(ctx, plan.EdgeLocations),
	}

	out, err := conn.CreateDirectConnectGatewayAttachment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkManager, create.ErrActionCreating, ResNameDXGatewayAttachment, plan.DirectConnectGatewayARN.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.DirectConnectGatewayAttachment == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkManager, create.ErrActionCreating, ResNameDXGatewayAttachment, plan.DirectConnectGatewayARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	a := out.DirectConnectGatewayAttachment.Attachment

	plan.ARN = flex.StringToFramework(ctx, a.ResourceArn)
	plan.ID = flex.StringToFramework(ctx, a.AttachmentId)
	plan.AttachmentType = flex.StringToFramework(ctx, (*string)(&a.AttachmentType))
	plan.AttachmentPolicyRuleNumber = flex.Int32ToFramework(ctx, a.AttachmentPolicyRuleNumber)
	plan.CoreNetworkARN = flex.StringToFramework(ctx, a.CoreNetworkArn)
	plan.CoreNetworkID = flex.StringToFramework(ctx, a.CoreNetworkId)
	plan.NetworkFunctionGroupName = flex.StringToFramework(ctx, a.NetworkFunctionGroupName)
	plan.OwnerAccountId = flex.StringToFramework(ctx, a.OwnerAccountId)
	plan.SegmentName = flex.StringToFramework(ctx, a.SegmentName)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	status, waitErr := waitDXGatewayAttachmentCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if waitErr != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkManager, create.ErrActionWaitingForCreation, ResNameDXGatewayAttachment, plan.DirectConnectGatewayARN.String(), waitErr),
			waitErr.Error(),
		)
		return
	}

	// Set state attribute once resource creation has completed
	plan.State = flex.StringToFramework(ctx, (*string)(&status.Attachment.State))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDXGatewayAttachment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkManagerClient(ctx)

	var state resourceDXGatewayAttachmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDXGatewayAttachmentByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkManager, create.ErrActionSetting, ResNameDXGatewayAttachment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	a := out.Attachment

	state.ARN = flex.StringToFramework(ctx, a.ResourceArn)
	state.AttachmentPolicyRuleNumber = flex.Int32ToFramework(ctx, a.AttachmentPolicyRuleNumber)
	state.AttachmentType = flex.StringToFramework(ctx, (*string)(&a.AttachmentType))
	state.CoreNetworkARN = flex.StringToFramework(ctx, a.CoreNetworkArn)
	state.CoreNetworkID = flex.StringToFramework(ctx, a.CoreNetworkId)
	state.DirectConnectGatewayARN = flex.StringToFramework(ctx, out.DirectConnectGatewayArn)
	state.EdgeLocations = flex.FlattenFrameworkStringValueList(ctx, a.EdgeLocations)
	state.ID = flex.StringToFramework(ctx, a.AttachmentId)
	state.NetworkFunctionGroupName = flex.StringToFramework(ctx, a.NetworkFunctionGroupName)
	state.OwnerAccountId = flex.StringToFramework(ctx, a.OwnerAccountId)
	state.SegmentName = flex.StringToFramework(ctx, a.SegmentName)
	state.State = flex.StringToFramework(ctx, (*string)(&a.State))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDXGatewayAttachment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkManagerClient(ctx)

	var plan, state resourceDXGatewayAttachmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Attachment must be in an Available state to be modified
	// Only edge locations can be modified
	if !plan.EdgeLocations.Equal(state.EdgeLocations) {

		in := &networkmanager.UpdateDirectConnectGatewayAttachmentInput{
			AttachmentId:  aws.String(plan.ID.ValueString()),
			EdgeLocations: flex.ExpandFrameworkStringValueList(ctx, plan.EdgeLocations),
		}

		out, err := conn.UpdateDirectConnectGatewayAttachment(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkManager, create.ErrActionUpdating, ResNameDXGatewayAttachment, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.DirectConnectGatewayAttachment == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkManager, create.ErrActionUpdating, ResNameDXGatewayAttachment, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.DirectConnectGatewayAttachment.Attachment.ResourceArn)
		plan.ID = flex.StringToFramework(ctx, out.DirectConnectGatewayAttachment.Attachment.AttachmentId)

		resp.Diagnostics.Append(flex.Flatten(ctx, out.DirectConnectGatewayAttachment.Attachment, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	status, err := waitDXGatewayAttachmentUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkManager, create.ErrActionWaitingForUpdate, ResNameDXGatewayAttachment, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Set state attribute once resource update has completed
	plan.State = flex.StringToFramework(ctx, (*string)(&status.Attachment.State))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDXGatewayAttachment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkManagerClient(ctx)

	var state resourceDXGatewayAttachmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	output, stateErr := findDXGatewayAttachmentByID(ctx, conn, state.ID.ValueString())
	if errs.IsA[*awstypes.ResourceNotFoundException](stateErr) {
		return
	}
	if stateErr != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkManager, create.ErrActionReading, ResNameDXGatewayAttachment, state.ID.String(), stateErr),
			stateErr.Error(),
		)
		return
	}

	// Get current attachment state attribute
	state.State = flex.StringToFramework(ctx, (*string)(&output.Attachment.State))

	// If attachment state is pending acceptance, reject the attachment before deleting
	if attachmentState := awstypes.AttachmentState(state.State.ValueString()); attachmentState == awstypes.AttachmentStatePendingAttachmentAcceptance || attachmentState == awstypes.AttachmentStatePendingTagAcceptance {
		_, err := conn.RejectAttachment(ctx, &networkmanager.RejectAttachmentInput{
			AttachmentId: aws.String(state.ID.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkManager, "detaching", ResNameDXGatewayAttachment, state.ID.String(), err),
				err.Error(),
			)
		}

		if _, err := waitDXGatewayAttachmentRejected(ctx, conn, state.ID.ValueString(), deleteTimeout); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkManager, "waiting for attachment to be rejected", ResNameDXGatewayAttachment, state.ID.String(), err),
				err.Error(),
			)
		}
	}

	const (
		timeout = 5 * time.Minute
	)

	in := &networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteAttachment(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkManager, create.ErrActionDeleting, ResNameDXGatewayAttachment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	_, err = waitDXGatewayAttachmentDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkManager, create.ErrActionWaitingForDeletion, ResNameDXGatewayAttachment, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDXGatewayAttachment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitDXGatewayAttachmentCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:                    enum.Slice(awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingAttachmentAcceptance),
		Refresh:                   statusDXGatewayAttachment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		return out, err
	}

	return nil, err
}

func waitDXGatewayAttachmentUpdated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AttachmentStateUpdating, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:                    enum.Slice(awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingTagAcceptance),
		Refresh:                   statusDXGatewayAttachment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		return out, err
	}

	return nil, err
}

func waitDXGatewayAttachmentDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateDeleting),
		Target:  []string{},
		Refresh: statusDXGatewayAttachment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		return out, err
	}

	return nil, err
}

func waitDXGatewayAttachmentAvailable(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate, awstypes.AttachmentStatePendingAttachmentAcceptance, awstypes.AttachmentStateUpdating),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable),
		Refresh: statusDXGatewayAttachment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		return out, err
	}

	return nil, err
}

func waitDXGatewayAttachmentRejected(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatePendingAttachmentAcceptance, awstypes.AttachmentStateAvailable),
		Target:  enum.Slice(awstypes.AttachmentStateRejected),
		Refresh: statusDXGatewayAttachment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		return out, err
	}

	return nil, err
}

func statusDXGatewayAttachment(ctx context.Context, conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDXGatewayAttachmentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Attachment.State)), nil
	}
}

func findDXGatewayAttachmentByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.DirectConnectGatewayAttachment, error) {
	in := &networkmanager.GetDirectConnectGatewayAttachmentInput{
		AttachmentId: aws.String(id),
	}

	out, err := conn.GetDirectConnectGatewayAttachment(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.DirectConnectGatewayAttachment.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.DirectConnectGatewayAttachment, nil
}

type resourceDXGatewayAttachmentData struct {
	ARN                        types.String   `tfsdk:"arn"`
	AttachmentPolicyRuleNumber types.Int64    `tfsdk:"attachment_policy_rule_number"`
	AttachmentType             types.String   `tfsdk:"attachment_type"`
	CoreNetworkARN             types.String   `tfsdk:"core_network_arn"`
	CoreNetworkID              types.String   `tfsdk:"core_network_id"`
	DirectConnectGatewayARN    types.String   `tfsdk:"direct_connect_gateway_arn"`
	EdgeLocations              types.List     `tfsdk:"edge_locations"`
	ID                         types.String   `tfsdk:"id"`
	NetworkFunctionGroupName   types.String   `tfsdk:"network_function_group_name"`
	OwnerAccountId             types.String   `tfsdk:"owner_account_id"`
	SegmentName                types.String   `tfsdk:"segment_name"`
	State                      types.String   `tfsdk:"state"`
	Timeouts                   timeouts.Value `tfsdk:"timeouts"`
}

type complexArgumentData struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}

var complexArgumentAttrTypes = map[string]attr.Type{
	"nested_required": types.StringType,
	"nested_optional": types.StringType,
}
