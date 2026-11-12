// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_networkmanager_site_to_site_vpn_attachment", name="Site To Site VPN Attachment")
func newSiteToSiteVPNAttachmentDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &siteToSiteVPNAttachmentDataSource{}, nil
}

type siteToSiteVPNAttachmentDataSource struct {
	framework.DataSourceWithModel[siteToSiteVPNAttachmentDataSourceModel]
}

func (d *siteToSiteVPNAttachmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	conflictsWithAttachmentID := path.MatchRoot("attachment_id")

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"attachment_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("core_network_id")),
				},
			},
			"attachment_policy_rule_number": schema.Int64Attribute{
				Computed: true,
			},
			"attachment_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AttachmentType](),
				Computed:   true,
			},
			"core_network_arn": schema.StringAttribute{
				Computed: true,
			},
			"core_network_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"edge_location": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(conflictsWithAttachmentID),
				},
			},
			"last_modification_errors": framework.DataSourceComputedListOfObjectAttribute[attachmentErrorModel](ctx),
			"nat_gateway_support": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.NatGatewaySupportState](),
				Computed:   true,
			},
			"nat_system_route_support": schema.BoolAttribute{
				Computed: true,
			},
			"network_function_group_name": schema.StringAttribute{
				Computed: true,
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
			},
			"proposed_network_function_group_change": framework.DataSourceComputedListOfObjectAttribute[proposedNetworkFunctionGroupChangeModel](ctx),
			"proposed_segment_change":                framework.DataSourceComputedListOfObjectAttribute[proposedSegmentChangeModel](ctx),
			names.AttrResourceARN: schema.StringAttribute{
				Computed: true,
			},
			"routing_policy_label": schema.StringAttribute{
				Computed: true,
			},
			"segment_name": schema.StringAttribute{
				Computed: true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AttachmentState](),
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(conflictsWithAttachmentID),
				},
			},
			names.AttrTags: schema.MapAttribute{
				CustomType:  tftags.MapType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Map{
					mapvalidator.ConflictsWith(conflictsWithAttachmentID),
				},
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"vpn_connection_arn": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(conflictsWithAttachmentID),
				},
			},
		},
	}
}

func (d *siteToSiteVPNAttachmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().NetworkManagerClient(ctx)

	var data siteToSiteVPNAttachmentDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	attachmentID := data.AttachmentID.ValueString()
	if attachmentID == "" {
		input := networkmanager.ListAttachmentsInput{
			AttachmentType: awstypes.AttachmentTypeSiteToSiteVpn,
			CoreNetworkId:  fwflex.StringFromFramework(ctx, data.CoreNetworkID),
			EdgeLocation:   fwflex.StringFromFramework(ctx, data.EdgeLocation),
			State:          data.State.ValueEnum(),
		}
		tagsToMatch := tftags.New(ctx, data.Tags).IgnoreAWS().IgnoreConfig(d.Meta().IgnoreTagsConfig(ctx))

		attachment, err := findSiteToSiteVPNAttachmentByFilters(ctx, conn, &input, tagsToMatch, data.VPNConnectionARN.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, tfresource.SingularDataSourceFindError("Network Manager Site To Site VPN Attachment", err), smerr.ID, data.CoreNetworkID.String())
			return
		}

		attachmentID = aws.ToString(attachment.AttachmentId)
	}

	out, err := findSiteToSiteVPNAttachmentByID(ctx, conn, attachmentID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, tfresource.SingularDataSourceFindError("Network Manager Site To Site VPN Attachment", err), smerr.ID, attachmentID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data), smerr.ID, attachmentID)
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.Attachment, &data), smerr.ID, attachmentID)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ARN = types.StringValue(ownedAttachmentARN(ctx, d.Meta(), out.Attachment))
	data.Tags = attachmentTags(ctx, out.Attachment.Tags)

	routingPolicyLabel, err := findAttachmentRoutingPolicyAssociationLabelByTwoPartKey(ctx, conn, aws.ToString(out.Attachment.CoreNetworkId), attachmentID)
	if err != nil && !retry.NotFound(err) {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, attachmentID)
		return
	}
	data.RoutingPolicyLabel = fwflex.StringToFramework(ctx, routingPolicyLabel)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, attachmentID)
}

func findSiteToSiteVPNAttachmentByFilters(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListAttachmentsInput, tags tftags.KeyValueTags, vpnConnectionARN string) (*awstypes.Attachment, error) {
	output, err := findAttachmentsByTags(ctx, conn, input, tags)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if vpnConnectionARN != "" {
		output = tfslices.Filter(output, func(v awstypes.Attachment) bool {
			return aws.ToString(v.ResourceArn) == vpnConnectionARN
		})
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

type siteToSiteVPNAttachmentDataSourceModel struct {
	ARN                                types.String                                                             `tfsdk:"arn"`
	AttachmentID                       types.String                                                             `tfsdk:"attachment_id"`
	AttachmentPolicyRuleNumber         types.Int64                                                              `tfsdk:"attachment_policy_rule_number"`
	AttachmentType                     fwtypes.StringEnum[awstypes.AttachmentType]                              `tfsdk:"attachment_type"`
	CoreNetworkARN                     types.String                                                             `tfsdk:"core_network_arn"`
	CoreNetworkID                      types.String                                                             `tfsdk:"core_network_id"`
	CreatedAt                          timetypes.RFC3339                                                        `tfsdk:"created_at"`
	EdgeLocation                       types.String                                                             `tfsdk:"edge_location"`
	LastModificationErrors             fwtypes.ListNestedObjectValueOf[attachmentErrorModel]                    `tfsdk:"last_modification_errors"`
	NATGatewaySupport                  fwtypes.StringEnum[awstypes.NatGatewaySupportState]                      `tfsdk:"nat_gateway_support"`
	NATSystemRouteSupport              types.Bool                                                               `tfsdk:"nat_system_route_support"`
	NetworkFunctionGroupName           types.String                                                             `tfsdk:"network_function_group_name"`
	OwnerAccountID                     types.String                                                             `tfsdk:"owner_account_id"`
	ProposedNetworkFunctionGroupChange fwtypes.ListNestedObjectValueOf[proposedNetworkFunctionGroupChangeModel] `tfsdk:"proposed_network_function_group_change"`
	ProposedSegmentChange              fwtypes.ListNestedObjectValueOf[proposedSegmentChangeModel]              `tfsdk:"proposed_segment_change"`
	ResourceARN                        types.String                                                             `tfsdk:"resource_arn"`
	RoutingPolicyLabel                 types.String                                                             `tfsdk:"routing_policy_label"`
	SegmentName                        types.String                                                             `tfsdk:"segment_name"`
	State                              fwtypes.StringEnum[awstypes.AttachmentState]                             `tfsdk:"state"`
	Tags                               tftags.Map                                                               `tfsdk:"tags"`
	UpdatedAt                          timetypes.RFC3339                                                        `tfsdk:"updated_at"`
	VPNConnectionARN                   types.String                                                             `tfsdk:"vpn_connection_arn"`
}
