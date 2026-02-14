// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_route")
func routeResourceAsListResource() inttypes.ListResourceForSDK {
	l := routeListResource{}
	l.SetResourceSchema(resourceRoute())
	return &l
}

type routeListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type routeListResourceModel struct {
	framework.WithRegionModel
	RouteTableID types.String `tfsdk:"route_table_id"`
}

func (l *routeListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"route_table_id": listschema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *routeListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query routeListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	routeTableID := query.RouteTableID.ValueString()

	tflog.Info(ctx, "Listing routes", map[string]any{
		"route_table_id": routeTableID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ec2.DescribeRouteTablesInput{
			RouteTableIds: []string{routeTableID},
		}

		output, err := conn.DescribeRouteTables(ctx, input)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		if len(output.RouteTables) == 0 {
			return
		}

		routeTable := output.RouteTables[0]

		for _, route := range routeTable.Routes {
			var destination string
			var destinationKey string

			if route.DestinationCidrBlock != nil {
				destination = aws.ToString(route.DestinationCidrBlock)
				destinationKey = routeDestinationCIDRBlock
			} else if route.DestinationIpv6CidrBlock != nil {
				destination = aws.ToString(route.DestinationIpv6CidrBlock)
				destinationKey = routeDestinationIPv6CIDRBlock
			} else if route.DestinationPrefixListId != nil {
				destination = aws.ToString(route.DestinationPrefixListId)
				destinationKey = routeDestinationPrefixListID
			} else {
				continue
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), destination)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(routeCreateID(routeTableID, destination))
			rd.Set("route_table_id", routeTableID)
			rd.Set(destinationKey, destination)

			diags := resourceRouteRead(ctx, rd, awsClient)
			if diags.HasError() || rd.Id() == "" {
				tflog.Error(ctx, "Reading route", map[string]any{
					names.AttrID: destination,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}

			result.DisplayName = destination

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				tflog.Error(ctx, "Setting result", map[string]any{
					names.AttrID: destination,
					"diags":      result.Diagnostics,
				})
				continue
			}

			if !yield(result) {
				return
			}
		}
	}
}
