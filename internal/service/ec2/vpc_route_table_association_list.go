// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strings"

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

// @SDKListResource("aws_route_table_association")
func newRouteTableAssociationResourceAsListResource() inttypes.ListResourceForSDK {
	l := routeTableAssociationListResource{}
	l.SetResourceSchema(resourceRouteTableAssociation())
	return &l
}

var _ list.ListResource = &routeTableAssociationListResource{}

type routeTableAssociationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *routeTableAssociationListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"route_table_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the Route Table to list Associations from.",
			},
		},
	}
}

func (l *routeTableAssociationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	var query listRouteTableAssociationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	routeTableID := query.RouteTableID.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("route_table_id"): routeTableID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		routeTable, err := findRouteTableByID(ctx, conn, routeTableID)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		var routeTableDisplayName string
		routeTableTags := keyValueTags(ctx, routeTable.Tags)
		if v, ok := routeTableTags["Name"]; ok {
			routeTableDisplayName = v.ValueString()
		} else {
			routeTableDisplayName = routeTableID
		}

		subnetNames := make(map[string]string)
		subnetIDs := make([]string, 0, len(routeTable.Associations))
		for _, item := range routeTable.Associations {
			if item.SubnetId != nil {
				subnetIDs = append(subnetIDs, aws.ToString(item.SubnetId))
			}
		}
		if len(subnetIDs) > 0 {
			input := ec2.DescribeSubnetsInput{
				SubnetIds: subnetIDs,
			}
			subnets, err := findSubnets(ctx, conn, &input)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}
			for _, subnet := range subnets {
				tags := keyValueTags(ctx, subnet.Tags)
				if v, ok := tags["Name"]; ok {
					subnetNames[aws.ToString(subnet.SubnetId)] = v.ValueString()
				}
			}
		}

		// There can be at most one Internet Gateway for a Route Table,
		// because a VPC can have at most one, and a Route Table is associated with a single VPC.
		var internetGatewayID string
		gatewayIDs := make([]string, 0, len(routeTable.Associations))
		for _, item := range routeTable.Associations {
			if gatewayID := aws.ToString(item.GatewayId); gatewayID != "" {
				if strings.HasPrefix(gatewayID, "igw-") {
					internetGatewayID = gatewayID
					continue
				}
				gatewayIDs = append(gatewayIDs, aws.ToString(item.GatewayId))
			}
		}
		gatewayNames := make(map[string]string)
		if internetGatewayID != "" {
			internetGateway, err := findInternetGatewayByID(ctx, conn, internetGatewayID)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("finding Internet Gateway (%s): %w", internetGatewayID, err))
				yield(result)
				return
			}
			tags := keyValueTags(ctx, internetGateway.Tags)
			if v, ok := tags["Name"]; ok {
				gatewayNames[internetGatewayID] = v.ValueString()
			}
		}
		if len(gatewayIDs) > 0 {
			input := ec2.DescribeVpnGatewaysInput{
				VpnGatewayIds: gatewayIDs,
			}
			gateways, err := findVPNGateways(ctx, conn, &input)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("finding VPN Gateways: %w", err))
				yield(result)
				return
			}
			for _, gateway := range gateways {
				tags := keyValueTags(ctx, gateway.Tags)
				if v, ok := tags["Name"]; ok {
					gatewayNames[aws.ToString(gateway.VpnGatewayId)] = v.ValueString()
				}
			}
		}

		for _, item := range routeTable.Associations {
			if item.AssociationState.State == awstypes.RouteTableAssociationStateCodeDisassociated || item.AssociationState.State == awstypes.RouteTableAssociationStateCodeDisassociating {
				continue
			}

			id := aws.ToString(item.RouteTableAssociationId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := resourceRouteTableAssociationFlatten(&item, rd); err != nil {
					tflog.Error(ctx, "Reading EC2 Route Table Association", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			var targetDisplayName string
			if item.SubnetId != nil {
				targetDisplayName = aws.ToString(item.SubnetId)
				if subnetName, ok := subnetNames[aws.ToString(item.SubnetId)]; ok {
					targetDisplayName = subnetName
				}
			} else if item.GatewayId != nil {
				targetDisplayName = aws.ToString(item.GatewayId)
				if gatewayName, ok := gatewayNames[aws.ToString(item.GatewayId)]; ok {
					targetDisplayName = gatewayName
				}
			} else {
				targetDisplayName = "<unknown target type>"
			}

			result.DisplayName = fmt.Sprintf("%s / %s (%s)", targetDisplayName, routeTableDisplayName, aws.ToString(item.RouteTableAssociationId))

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

type listRouteTableAssociationModel struct {
	framework.WithRegionModel
	RouteTableID types.String `tfsdk:"route_table_id"`
}
