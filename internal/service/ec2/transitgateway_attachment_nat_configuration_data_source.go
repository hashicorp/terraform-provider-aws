// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_transit_gateway_attachment_nat_configuration", name="Transit Gateway Attachment NAT Configuration")
func newTransitGatewayAttachmentNATConfigurationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &transitGatewayAttachmentNATConfigurationDataSource{}, nil
}

type transitGatewayAttachmentNATConfigurationDataSource struct {
	framework.DataSourceWithModel[transitGatewayAttachmentNATConfigurationDataSourceModel]
}

func (d *transitGatewayAttachmentNATConfigurationDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"nat_gateway_rule": framework.DataSourceComputedListOfObjectAttribute[transitGatewayNATRuleDataSourceModel](ctx),
			"nat_gateway_support": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TransitGatewayAttachmentNatGatewaySupportState](),
				Computed:   true,
			},
			"nat_system_route_support": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TransitGatewayAttachmentNatSystemRouteSupportValue](),
				Computed:   true,
			},
			names.AttrTransitGatewayAttachmentID: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *transitGatewayAttachmentNATConfigurationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data transitGatewayAttachmentNATConfigurationDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.TransitGatewayAttachmentID)
	input := ec2.GetTransitGatewayAttachmentNatConfigurationInput{
		TransitGatewayAttachmentId: aws.String(id),
	}

	// Unlike the resource, an attachment without NAT configured is a valid result, so findTransitGatewayAttachmentNATConfigurationByID isn't used.
	output, err := findTransitGatewayAttachmentNATConfiguration(ctx, conn, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data), smerr.ID, id)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data), smerr.ID, id)
}

type transitGatewayAttachmentNATConfigurationDataSourceModel struct {
	framework.WithRegionModel
	NATGatewayRules            fwtypes.ListNestedObjectValueOf[transitGatewayNATRuleDataSourceModel]           `tfsdk:"nat_gateway_rule"`
	NATGatewaySupport          fwtypes.StringEnum[awstypes.TransitGatewayAttachmentNatGatewaySupportState]     `tfsdk:"nat_gateway_support"`
	NATSystemRouteSupport      fwtypes.StringEnum[awstypes.TransitGatewayAttachmentNatSystemRouteSupportValue] `tfsdk:"nat_system_route_support"`
	TransitGatewayAttachmentID types.String                                                                    `tfsdk:"transit_gateway_attachment_id"`
}

type transitGatewayNATRuleDataSourceModel struct {
	MatchSourceCIDR types.String                                            `tfsdk:"match_source_cidr"`
	NATCIDR         types.String                                            `tfsdk:"nat_cidr"`
	RuleNumber      types.Int32                                             `tfsdk:"rule_number"`
	State           fwtypes.StringEnum[awstypes.TransitGatewayNatRuleState] `tfsdk:"state"`
}
