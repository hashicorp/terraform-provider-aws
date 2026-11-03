// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)

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

	output, err := findAttachments(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, coreNetworkID)
		return
	}

	// ListAttachments has no tag filter, so tags are matched client-side.
	tagsToMatch := tftags.New(ctx, data.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	ids := make([]string, 0, len(output))
	attachments := make([]attachmentModel, 0, len(output))
	for _, v := range output {
		if len(tagsToMatch) > 0 && !keyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch) {
			continue
		}

		attachmentID := aws.ToString(v.AttachmentId)
		var attachment attachmentModel
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, v, &attachment,
			fwflex.WithIgnoredFieldNamesAppend("ProposedNetworkFunctionGroupChange"),
			fwflex.WithIgnoredFieldNamesAppend("ProposedSegmentChange"),
		), smerr.ID, attachmentID)
		if resp.Diagnostics.HasError() {
			return
		}

		// ListAttachments doesn't return the ARN. Attachments can belong to other accounts sharing the core network.
		attachment.ARN = types.StringValue(d.Meta().GlobalARNWithAccount(ctx, "networkmanager", aws.ToString(v.OwnerAccountId), "attachment/"+attachmentID))
		attachment.Tags = tftags.FlattenStringValueMap(ctx, keyValueTags(ctx, v.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())
		attachment.ProposedNetworkFunctionGroupChange = fwtypes.NewListNestedObjectValueOfNull[proposedNetworkFunctionGroupChangeModel](ctx)
		if c := v.ProposedNetworkFunctionGroupChange; c != nil {
			attachment.ProposedNetworkFunctionGroupChange = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &proposedNetworkFunctionGroupChangeModel{
				AttachmentPolicyRuleNumber: fwflex.Int32ToFrameworkInt64(ctx, c.AttachmentPolicyRuleNumber),
				NetworkFunctionGroupName:   fwflex.StringToFramework(ctx, c.NetworkFunctionGroupName),
				Tags:                       tftags.FlattenStringValueMap(ctx, keyValueTags(ctx, c.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()),
			})
		}
		attachment.ProposedSegmentChange = fwtypes.NewListNestedObjectValueOfNull[proposedSegmentChangeModel](ctx)
		if c := v.ProposedSegmentChange; c != nil {
			attachment.ProposedSegmentChange = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &proposedSegmentChangeModel{
				AttachmentPolicyRuleNumber: fwflex.Int32ToFrameworkInt64(ctx, c.AttachmentPolicyRuleNumber),
				SegmentName:                fwflex.StringToFramework(ctx, c.SegmentName),
				Tags:                       tftags.FlattenStringValueMap(ctx, keyValueTags(ctx, c.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()),
			})
		}

		ids = append(ids, attachmentID)
		attachments = append(attachments, attachment)
	}

	data.Attachments = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, attachments)
	data.IDs = fwflex.FlattenFrameworkStringValueListOfString(ctx, ids)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, coreNetworkID)
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
