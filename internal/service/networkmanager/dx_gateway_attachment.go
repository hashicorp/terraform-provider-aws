// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package networkmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkmanager_dx_gateway_attachment", name="Direct Connect Gateway Attachment")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networkmanager/types;awstypes;awstypes.DirectConnectGatewayAttachment")
// @Testing(skipEmptyTags=true, skipNullTags=true)
// @Testing(importIgnore="state")
func newDirectConnectGatewayAttachmentResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &directConnectGatewayAttachmentResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type directConnectGatewayAttachmentResource struct {
	framework.ResourceWithModel[directConnectGatewayAttachmentResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *directConnectGatewayAttachmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"attachment_policy_rule_number": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"attachment_type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"core_network_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"core_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direct_connect_gateway_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"edge_locations": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Required:    true,
				ElementType: types.StringType,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"routing_policy_label": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
			"segment_name": schema.StringAttribute{
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
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *directConnectGatewayAttachmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data directConnectGatewayAttachmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	var input networkmanager.CreateDirectConnectGatewayAttachmentInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateDirectConnectGatewayAttachment(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Network Manager Direct Connect Gateway Attachment", err.Error())

		return
	}

	// Set values for unknowns.
	id := aws.ToString(output.DirectConnectGatewayAttachment.Attachment.AttachmentId)
	data.ID = fwflex.StringValueToFramework(ctx, id)

	dxgwAttachment, err := waitDirectConnectGatewayAttachmentCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Manager Direct Connect Gateway Attachment (%s) create", id), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, dxgwAttachment.Attachment, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ARN = fwflex.StringValueToFramework(ctx, attachmentARN(ctx, r.Meta(), id))

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *directConnectGatewayAttachmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data directConnectGatewayAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	dxgwAttachment, err := findDirectConnectGatewayAttachmentByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Network Manager Direct Connect Gateway Attachment (%s)", id), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, dxgwAttachment.Attachment, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ARN = fwflex.StringValueToFramework(ctx, attachmentARN(ctx, r.Meta(), id))
	data.DirectConnectGatewayARN = fwflex.StringToFrameworkARN(ctx, dxgwAttachment.DirectConnectGatewayArn)
	if routingPolicyLabel, err := findAttachmentRoutingPolicyAssociationLabelByTwoPartKey(ctx, conn, data.CoreNetworkID.ValueString(), id); err != nil && !retry.NotFound(err) {
		response.Diagnostics.AddError(fmt.Sprintf("reading Network Manager Direct Connect Gateway Attachment (%s) routing policy label", id), err.Error())
		return
	} else {
		data.RoutingPolicyLabel = fwflex.StringToFramework(ctx, routingPolicyLabel)
	}

	setTagsOut(ctx, dxgwAttachment.Attachment.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *directConnectGatewayAttachmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new directConnectGatewayAttachmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	// Attachment must be in an Available state to be modified
	// Only edge locations can be modified
	if !new.EdgeLocations.Equal(old.EdgeLocations) {
		id := fwflex.StringValueFromFramework(ctx, new.ID)
		input := networkmanager.UpdateDirectConnectGatewayAttachmentInput{
			AttachmentId:  aws.String(id),
			EdgeLocations: fwflex.ExpandFrameworkStringValueList(ctx, new.EdgeLocations),
		}

		_, err := conn.UpdateDirectConnectGatewayAttachment(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Network Manager Direct Connect Gateway Attachment (%s)", id), err.Error())

			return
		}

		dxgwAttachment, err := waitDirectConnectGatewayAttachmentUpdated(ctx, conn, id, r.UpdateTimeout(ctx, new.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Manager Direct Connect Gateway Attachment (%s) update", id), err.Error())

			return
		}

		// Set values for unknowns.
		new.State = fwflex.StringValueToFramework(ctx, dxgwAttachment.Attachment.State)
	} else {
		new.State = old.State
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *directConnectGatewayAttachmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data directConnectGatewayAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	dxgwAttachment, err := findDirectConnectGatewayAttachmentByID(ctx, conn, id)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Network Manager Direct Connect Gateway Attachment (%s)", id), err.Error())

		return
	}

	// If attachment state is pending acceptance, reject the attachment before deleting.
	if state := dxgwAttachment.Attachment.State; state == awstypes.AttachmentStatePendingAttachmentAcceptance {
		input := networkmanager.RejectAttachmentInput{
			AttachmentId: aws.String(id),
		}

		_, err := conn.RejectAttachment(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("rejecting Network Manager Direct Connect Gateway Attachment (%s)", id), err.Error())

			return
		}

		if _, err := waitDirectConnectGatewayAttachmentRejected(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Manager Direct Connect Gateway Attachment (%s) reject", id), err.Error())

			return
		}
	}

	input := networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(id),
	}
	_, err = conn.DeleteAttachment(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Network Manager Direct Connect Gateway Attachment (%s)", id), err.Error())

		return
	}

	if _, err := waitDirectConnectGatewayAttachmentDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Manager Direct Connect Gateway Attachment (%s) delete", id), err.Error())

		return
	}
}

func findDirectConnectGatewayAttachmentByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.DirectConnectGatewayAttachment, error) {
	input := networkmanager.GetDirectConnectGatewayAttachmentInput{
		AttachmentId: aws.String(id),
	}

	return findDirectConnectGatewayAttachment(ctx, conn, &input)
}

func findDirectConnectGatewayAttachment(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetDirectConnectGatewayAttachmentInput) (*awstypes.DirectConnectGatewayAttachment, error) {
	output, err := conn.GetDirectConnectGatewayAttachment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DirectConnectGatewayAttachment == nil || output.DirectConnectGatewayAttachment.Attachment == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.DirectConnectGatewayAttachment, nil
}

func statusDirectConnectGatewayAttachment(conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDirectConnectGatewayAttachmentByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Attachment.State), nil
	}
}

func waitDirectConnectGatewayAttachmentCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:                    enum.Slice(awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingAttachmentAcceptance),
		Refresh:                   statusDirectConnectGatewayAttachment(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitDirectConnectGatewayAttachmentUpdated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateUpdating, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingTagAcceptance),
		Refresh: statusDirectConnectGatewayAttachment(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitDirectConnectGatewayAttachmentDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.AttachmentStateDeleting, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:         []string{},
		Refresh:        statusDirectConnectGatewayAttachment(conn, id),
		Timeout:        timeout,
		Delay:          2 * time.Minute,
		PollInterval:   10 * time.Second,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitDirectConnectGatewayAttachmentAvailable(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate, awstypes.AttachmentStatePendingAttachmentAcceptance, awstypes.AttachmentStateUpdating),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable),
		Refresh: statusDirectConnectGatewayAttachment(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitDirectConnectGatewayAttachmentRejected(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatePendingAttachmentAcceptance, awstypes.AttachmentStateAvailable),
		Target:  enum.Slice(awstypes.AttachmentStateRejected),
		Refresh: statusDirectConnectGatewayAttachment(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

type directConnectGatewayAttachmentResourceModel struct {
	ARN                        types.String         `tfsdk:"arn"`
	AttachmentPolicyRuleNumber types.Int64          `tfsdk:"attachment_policy_rule_number"`
	AttachmentType             types.String         `tfsdk:"attachment_type"`
	CoreNetworkARN             types.String         `tfsdk:"core_network_arn"`
	CoreNetworkID              types.String         `tfsdk:"core_network_id"`
	DirectConnectGatewayARN    fwtypes.ARN          `tfsdk:"direct_connect_gateway_arn"`
	EdgeLocations              fwtypes.ListOfString `tfsdk:"edge_locations"`
	ID                         types.String         `tfsdk:"id"`
	OwnerAccountId             types.String         `tfsdk:"owner_account_id"`
	RoutingPolicyLabel         types.String         `tfsdk:"routing_policy_label"`
	SegmentName                types.String         `tfsdk:"segment_name"`
	State                      types.String         `tfsdk:"state"`
	Tags                       tftags.Map           `tfsdk:"tags"`
	TagsAll                    tftags.Map           `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value       `tfsdk:"timeouts"`
}
