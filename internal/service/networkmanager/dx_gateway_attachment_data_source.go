// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"slices"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

// @FrameworkDataSource("aws_networkmanager_dx_gateway_attachment", name="Direct Connect Gateway Attachment")
func newDirectConnectGatewayAttachmentDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &directConnectGatewayAttachmentDataSource{}, nil
}

type directConnectGatewayAttachmentDataSource struct {
	framework.DataSourceWithModel[directConnectGatewayAttachmentDataSourceModel]
}

func (d *directConnectGatewayAttachmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"direct_connect_gateway_arn": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(conflictsWithAttachmentID),
				},
			},
			"edge_locations": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					listvalidator.ConflictsWith(conflictsWithAttachmentID),
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
		},
	}
}

func (d *directConnectGatewayAttachmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().NetworkManagerClient(ctx)

	var data directConnectGatewayAttachmentDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var out *awstypes.DirectConnectGatewayAttachment
	var err error
	if attachmentID := data.AttachmentID.ValueString(); attachmentID != "" {
		out, err = findDirectConnectGatewayAttachmentByID(ctx, conn, attachmentID)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, tfresource.SingularDataSourceFindError("Network Manager Direct Connect Gateway Attachment", err), smerr.ID, attachmentID)
			return
		}
	} else {
		// The edge locations aren't sent to ListAttachments: its EdgeLocation filter takes a single Region,
		// and these attachments span a list of edges. They are matched client-side instead.
		input := networkmanager.ListAttachmentsInput{
			AttachmentType: awstypes.AttachmentTypeDirectConnectGateway,
			CoreNetworkId:  fwflex.StringFromFramework(ctx, data.CoreNetworkID),
			State:          data.State.ValueEnum(),
		}
		tagsToMatch := tftags.New(ctx, data.Tags).IgnoreAWS().IgnoreConfig(d.Meta().IgnoreTagsConfig(ctx))
		edgeLocations := fwflex.ExpandFrameworkStringValueList(ctx, data.EdgeLocations)

		out, err = findDirectConnectGatewayAttachmentByFilters(ctx, conn, &input, tagsToMatch, edgeLocations, data.DirectConnectGatewayARN.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, tfresource.SingularDataSourceFindError("Network Manager Direct Connect Gateway Attachment", err), smerr.ID, data.CoreNetworkID.String())
			return
		}
	}
	attachmentID := aws.ToString(out.Attachment.AttachmentId)

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

// findDirectConnectGatewayAttachmentByFilters returns the single attachment matching the filters.
// An attachment matches the edge locations when it covers all of them; it may cover more.
// ListAttachments doesn't return the Direct Connect gateway ARN, so when that filter is set
// each remaining candidate is read with GetDirectConnectGatewayAttachment and compared.
func findDirectConnectGatewayAttachmentByFilters(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListAttachmentsInput, tags tftags.KeyValueTags, edgeLocations []string, directConnectGatewayARN string) (*awstypes.DirectConnectGatewayAttachment, error) {
	output, err := findAttachmentsByTags(ctx, conn, input, tags)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if len(edgeLocations) > 0 {
		output = tfslices.Filter(output, func(v awstypes.Attachment) bool {
			return !slices.ContainsFunc(edgeLocations, func(edgeLocation string) bool {
				return !slices.Contains(v.EdgeLocations, edgeLocation)
			})
		})
	}

	if directConnectGatewayARN == "" {
		attachment, err := tfresource.AssertSingleValueResult(output)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		return smarterr.Assert(findDirectConnectGatewayAttachmentByID(ctx, conn, aws.ToString(attachment.AttachmentId)))
	}

	var matches []awstypes.DirectConnectGatewayAttachment
	for _, v := range output {
		attachment, err := findDirectConnectGatewayAttachmentByID(ctx, conn, aws.ToString(v.AttachmentId))
		if retry.NotFound(err) {
			continue
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		if aws.ToString(attachment.DirectConnectGatewayArn) != directConnectGatewayARN {
			continue
		}

		matches = append(matches, *attachment)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(matches))
}

type directConnectGatewayAttachmentDataSourceModel struct {
	ARN                                types.String                                                             `tfsdk:"arn"`
	AttachmentID                       types.String                                                             `tfsdk:"attachment_id"`
	AttachmentPolicyRuleNumber         types.Int64                                                              `tfsdk:"attachment_policy_rule_number"`
	AttachmentType                     fwtypes.StringEnum[awstypes.AttachmentType]                              `tfsdk:"attachment_type"`
	CoreNetworkARN                     types.String                                                             `tfsdk:"core_network_arn"`
	CoreNetworkID                      types.String                                                             `tfsdk:"core_network_id"`
	CreatedAt                          timetypes.RFC3339                                                        `tfsdk:"created_at"`
	DirectConnectGatewayARN            types.String                                                             `tfsdk:"direct_connect_gateway_arn"`
	EdgeLocations                      fwtypes.ListOfString                                                     `tfsdk:"edge_locations"`
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
}
