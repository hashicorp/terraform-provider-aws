// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ec2_transit_gateway_route")
func newTransitGatewayRouteResourceAsListResource() inttypes.ListResourceForSDK {
	l := transitGatewayRouteListResource{}
	l.SetResourceSchema(resourceTransitGatewayRoute())
	return &l
}

var _ list.ListResource = &transitGatewayRouteListResource{}

type transitGatewayRouteListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *transitGatewayRouteListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"transit_gateway_route_table_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the transit gateway route table.",
			},
		},
	}
}

func (l *transitGatewayRouteListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	var query listTransitGatewayRouteModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	transitGatewayRouteTableID := query.TransitGatewayRouteTableID.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("transit_gateway_route_table_id"): transitGatewayRouteTableID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ec2.SearchTransitGatewayRoutesInput{
			Filters: newAttributeFilterList(map[string]string{
				names.AttrType: string(awstypes.TransitGatewayRouteTypeStatic),
			}),
			TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
		}
		paginator := ec2.NewSearchTransitGatewayRoutesPaginator(conn, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, route := range page.Routes {
				if route.State == awstypes.TransitGatewayRouteStateDeleted {
					continue
				}

				destination := inttypes.CanonicalCIDRBlock(aws.ToString(route.DestinationCidrBlock))
				route.DestinationCidrBlock = aws.String(destination)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("destination_cidr_block"), destination)

				result := request.NewListResult(ctx)
				rd := l.ResourceData()
				rd.SetId(transitGatewayRouteCreateResourceID(transitGatewayRouteTableID, destination))

				if err := resourceTransitGatewayRouteFlatten(rd, &route, transitGatewayRouteTableID); err != nil {
					tflog.Error(ctx, "Flattening EC2 Transit Gateway Route", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				result.DisplayName = destination

				l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
				if result.Diagnostics.HasError() {
					yield(result)
					return
				}

				if !yield(result) {
					return
				}
			}
		}
	}
}

type listTransitGatewayRouteModel struct {
	framework.WithRegionModel
	TransitGatewayRouteTableID types.String `tfsdk:"transit_gateway_route_table_id"`
}
