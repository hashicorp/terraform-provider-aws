// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_networkmanager_attachments", name="Attachments")
func newAttachmentsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &attachmentsDataSource{}, nil
}

type attachmentsDataSource struct {
	framework.DataSourceWithModel[attachmentsDataSourceModel]
}

func (d *attachmentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"attachment_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AttachmentType](),
				Optional:   true,
			},
			"attachments": framework.DataSourceComputedListOfObjectAttribute[attachmentModel](ctx),
			"core_network_id": schema.StringAttribute{
				Required: true,
			},
			"edge_location": schema.StringAttribute{
				Optional: true,
			},
			names.AttrIDs: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AttachmentState](),
				Optional:   true,
			},
			names.AttrTags: tftags.TagsAttribute(),
		},
	}
}

func (d *attachmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().NetworkManagerClient(ctx)

	var data attachmentsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	coreNetworkID := data.CoreNetworkID.ValueString()
	input := networkmanager.ListAttachmentsInput{
		AttachmentType: data.AttachmentType.ValueEnum(),
		CoreNetworkId:  aws.String(coreNetworkID),
		EdgeLocation:   fwflex.StringFromFramework(ctx, data.EdgeLocation),
		State:          data.State.ValueEnum(),
	}
	tagsToMatch := tftags.New(ctx, data.Tags).IgnoreAWS().IgnoreConfig(d.Meta().IgnoreTagsConfig(ctx))

	output, err := findAttachmentsByTags(ctx, conn, &input, tagsToMatch)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, coreNetworkID)
		return
	}

	ids := make([]string, 0, len(output))
	attachments := make([]attachmentModel, 0, len(output))
	for _, v := range output {
		attachmentID := aws.ToString(v.AttachmentId)
		var attachment attachmentModel
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, v, &attachment), smerr.ID, attachmentID)
		if resp.Diagnostics.HasError() {
			return
		}
		attachment.ARN = types.StringValue(ownedAttachmentARN(ctx, d.Meta(), &v))
		attachment.Tags = attachmentTags(ctx, v.Tags)

		ids = append(ids, attachmentID)
		attachments = append(attachments, attachment)
	}

	data.Attachments = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, attachments)
	data.IDs = fwflex.FlattenFrameworkStringValueListOfString(ctx, ids)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, coreNetworkID)
}

// findAttachmentsByTags lists attachments and keeps those having all the given tags.
// ListAttachments has no tag filter, so the tags are matched client-side.
func findAttachmentsByTags(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListAttachmentsInput, tags tftags.KeyValueTags) ([]awstypes.Attachment, error) {
	output, err := findAttachments(ctx, conn, input)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if len(tags) == 0 {
		return output, nil
	}

	return tfslices.Filter(output, func(v awstypes.Attachment) bool {
		return keyValueTags(ctx, v.Tags).ContainsAll(tags)
	}), nil
}

// ownedAttachmentARN returns the attachment's ARN in its owner's account.
// Neither ListAttachments nor the Get*Attachment operations return the ARN, and an attachment
// can belong to another account sharing the core network.
func ownedAttachmentARN(ctx context.Context, c *conns.AWSClient, v *awstypes.Attachment) string {
	return c.GlobalARNWithAccount(ctx, "networkmanager", aws.ToString(v.OwnerAccountId), "attachment/"+aws.ToString(v.AttachmentId))
}

// attachmentTags returns the tags without AWS-reserved keys or keys ignored by the provider configuration.
func attachmentTags(ctx context.Context, tags []awstypes.Tag) tftags.Map {
	kvTags := keyValueTags(ctx, tags).IgnoreAWS()
	if inContext, ok := tftags.FromContext(ctx); ok {
		kvTags = kvTags.IgnoreConfig(inContext.IgnoreConfig)
	}
	return tftags.FlattenStringValueMap(ctx, kvTags.Map())
}

type attachmentsDataSourceModel struct {
	AttachmentType fwtypes.StringEnum[awstypes.AttachmentType]      `tfsdk:"attachment_type"`
	Attachments    fwtypes.ListNestedObjectValueOf[attachmentModel] `tfsdk:"attachments"`
	CoreNetworkID  types.String                                     `tfsdk:"core_network_id"`
	EdgeLocation   types.String                                     `tfsdk:"edge_location"`
	IDs            fwtypes.ListOfString                             `tfsdk:"ids"`
	State          fwtypes.StringEnum[awstypes.AttachmentState]     `tfsdk:"state"`
	Tags           tftags.Map                                       `tfsdk:"tags"`
}

type attachmentModel struct {
	ARN                                types.String                                                             `tfsdk:"arn"`
	AttachmentID                       types.String                                                             `tfsdk:"attachment_id"`
	AttachmentPolicyRuleNumber         types.Int64                                                              `tfsdk:"attachment_policy_rule_number"`
	AttachmentType                     fwtypes.StringEnum[awstypes.AttachmentType]                              `tfsdk:"attachment_type"`
	CoreNetworkARN                     types.String                                                             `tfsdk:"core_network_arn"`
	CoreNetworkID                      types.String                                                             `tfsdk:"core_network_id"`
	CreatedAt                          timetypes.RFC3339                                                        `tfsdk:"created_at"`
	EdgeLocation                       types.String                                                             `tfsdk:"edge_location"`
	EdgeLocations                      fwtypes.ListOfString                                                     `tfsdk:"edge_locations"`
	LastModificationErrors             fwtypes.ListNestedObjectValueOf[attachmentErrorModel]                    `tfsdk:"last_modification_errors"`
	NATGatewaySupport                  fwtypes.StringEnum[awstypes.NatGatewaySupportState]                      `tfsdk:"nat_gateway_support"`
	NATSystemRouteSupport              types.Bool                                                               `tfsdk:"nat_system_route_support"`
	NetworkFunctionGroupName           types.String                                                             `tfsdk:"network_function_group_name"`
	OwnerAccountID                     types.String                                                             `tfsdk:"owner_account_id"`
	ProposedNetworkFunctionGroupChange fwtypes.ListNestedObjectValueOf[proposedNetworkFunctionGroupChangeModel] `tfsdk:"proposed_network_function_group_change"`
	ProposedSegmentChange              fwtypes.ListNestedObjectValueOf[proposedSegmentChangeModel]              `tfsdk:"proposed_segment_change"`
	ResourceARN                        types.String                                                             `tfsdk:"resource_arn"`
	SegmentName                        types.String                                                             `tfsdk:"segment_name"`
	State                              fwtypes.StringEnum[awstypes.AttachmentState]                             `tfsdk:"state"`
	Tags                               tftags.Map                                                               `tfsdk:"tags"`
	UpdatedAt                          timetypes.RFC3339                                                        `tfsdk:"updated_at"`
}

type attachmentErrorModel struct {
	Code        types.String `tfsdk:"code"`
	Message     types.String `tfsdk:"message"`
	RequestID   types.String `tfsdk:"request_id"`
	ResourceARN types.String `tfsdk:"resource_arn"`
}

type proposedNetworkFunctionGroupChangeModel struct {
	AttachmentPolicyRuleNumber types.Int64  `tfsdk:"attachment_policy_rule_number"`
	NetworkFunctionGroupName   types.String `tfsdk:"network_function_group_name"`
	Tags                       tftags.Map   `tfsdk:"tags"`
}

type proposedSegmentChangeModel struct {
	AttachmentPolicyRuleNumber types.Int64  `tfsdk:"attachment_policy_rule_number"`
	SegmentName                types.String `tfsdk:"segment_name"`
	Tags                       tftags.Map   `tfsdk:"tags"`
}

var (
	_ fwflex.Flattener = &proposedNetworkFunctionGroupChangeModel{}
	_ fwflex.Flattener = &proposedSegmentChangeModel{}
)

// Flatten implements fwflex.Flattener. AutoFlex skips fields named Tags, so the model is set explicitly.
func (m *proposedNetworkFunctionGroupChangeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ProposedNetworkFunctionGroupChange:
		m.AttachmentPolicyRuleNumber = fwflex.Int32ToFrameworkInt64(ctx, t.AttachmentPolicyRuleNumber)
		m.NetworkFunctionGroupName = fwflex.StringToFramework(ctx, t.NetworkFunctionGroupName)
		m.Tags = attachmentTags(ctx, t.Tags)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("proposedNetworkFunctionGroupChangeModel.Flatten: %T", v))
	}

	return diags
}

// Flatten implements fwflex.Flattener. AutoFlex skips fields named Tags, so the model is set explicitly.
func (m *proposedSegmentChangeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ProposedSegmentChange:
		m.AttachmentPolicyRuleNumber = fwflex.Int32ToFrameworkInt64(ctx, t.AttachmentPolicyRuleNumber)
		m.SegmentName = fwflex.StringToFramework(ctx, t.SegmentName)
		m.Tags = attachmentTags(ctx, t.Tags)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("proposedSegmentChangeModel.Flatten: %T", v))
	}

	return diags
}
