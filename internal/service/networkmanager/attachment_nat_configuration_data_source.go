// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkDataSource("aws_networkmanager_attachment_nat_configuration", name="Attachment NAT Configuration")
func newAttachmentNATConfigurationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &attachmentNATConfigurationDataSource{}, nil
}

type attachmentNATConfigurationDataSource struct {
	framework.DataSourceWithModel[attachmentNATConfigurationDataSourceModel]
}

func (d *attachmentNATConfigurationDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"attachment_id": schema.StringAttribute{
				Required: true,
			},
			"nat_gateway_rule": framework.DataSourceComputedListOfObjectAttribute[attachmentNATRuleDataSourceModel](ctx),
			"nat_gateway_support": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.NatGatewaySupportState](),
				Computed:   true,
			},
			"nat_system_route_support": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *attachmentNATConfigurationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data attachmentNATConfigurationDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().NetworkManagerClient(ctx)

	attachmentID := fwflex.StringValueFromFramework(ctx, data.AttachmentID)
	input := networkmanager.ListAttachmentNatConfigurationsInput{
		AttachmentId: aws.String(attachmentID),
	}
	// Unlike the resource, NAT disabled with no rules is a valid result, so findAttachmentNATConfigurationByID isn't used.
	output, err := findAttachmentNATConfiguration(ctx, conn, &input)

	switch {
	case errors.Is(err, tfresource.ErrEmptyResult):
		// A missing attachment is reported with ResourceNotFoundException.
		// A successful response without an entry means the attachment has no NAT configuration.
		output = &awstypes.AttachmentNatConfiguration{
			AttachmentId:      aws.String(attachmentID),
			NatGatewaySupport: awstypes.NatGatewaySupportStateDisabled,
		}
	case err != nil:
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data), smerr.ID, attachmentID)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data), smerr.ID, attachmentID)
}

type attachmentNATConfigurationDataSourceModel struct {
	AttachmentID          types.String                                                      `tfsdk:"attachment_id"`
	NATGatewayRules       fwtypes.ListNestedObjectValueOf[attachmentNATRuleDataSourceModel] `tfsdk:"nat_gateway_rule"`
	NATGatewaySupport     fwtypes.StringEnum[awstypes.NatGatewaySupportState]               `tfsdk:"nat_gateway_support"`
	NATSystemRouteSupport types.Bool                                                        `tfsdk:"nat_system_route_support"`
}

type attachmentNATRuleDataSourceModel struct {
	MatchSourceCIDR types.String                              `tfsdk:"match_source_cidr"`
	NATCIDR         types.String                              `tfsdk:"nat_cidr"`
	RuleNumber      types.Int32                               `tfsdk:"rule_number"`
	State           fwtypes.StringEnum[awstypes.NatRuleState] `tfsdk:"state"`
}
