// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

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

// @SDKListResource("aws_ec2_transit_gateway_route_table_propagation", name="Transit Gateway Route Table Propagation")
func newTransitGatewayRouteTablePropagationResourceAsListResource() inttypes.ListResourceForSDK {
	l := transitGatewayRouteTablePropagationListResource{}
	l.SetResourceSchema(resourceTransitGatewayRouteTablePropagation())
	return &l
}

var _ list.ListResource = &transitGatewayRouteTablePropagationListResource{}

type transitGatewayRouteTablePropagationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *transitGatewayRouteTablePropagationListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"transit_gateway_route_table_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the Transit Gateway route table to list propagations for.",
			},
		},
	}
}

func (l *transitGatewayRouteTablePropagationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	var query listTransitGatewayRouteTablePropagationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	routeTableID := query.TransitGatewayRouteTableID.ValueString()
	tflog.Info(ctx, "Listing EC2 Transit Gateway Route Table Propagation resources", map[string]any{
		logging.ResourceAttributeKey("transit_gateway_route_table_id"): routeTableID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		for propagation, err := range listTransitGatewayRouteTablePropagations(ctx, conn, routeTableID) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			if propagation.State == awstypes.TransitGatewayPropagationStateDisabled {
				continue
			}

			attachmentID := aws.ToString(propagation.TransitGatewayAttachmentId)
			if attachmentID == "" {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing EC2 Transit Gateway Route Table Propagation returned an empty transit gateway attachment ID"))
				yield(result)
				return
			}

			itemCtx := tflog.SetField(ctx, logging.ResourceAttributeKey("transit_gateway_route_table_id"), routeTableID)
			itemCtx = tflog.SetField(itemCtx, logging.ResourceAttributeKey(names.AttrTransitGatewayAttachmentID), attachmentID)
			result := request.NewListResult(itemCtx)

			rd := l.ResourceData()
			rd.SetId(transitGatewayRouteTablePropagationCreateResourceID(routeTableID, attachmentID))
			rd.Set(names.AttrTransitGatewayAttachmentID, attachmentID)
			rd.Set("transit_gateway_route_table_id", routeTableID)

			if request.IncludeResource {
				if err := resourceTransitGatewayRouteTablePropagationFlatten(&propagation, routeTableID, rd); err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}
			}

			result.DisplayName = fmt.Sprintf("%s (%s)", routeTableID, attachmentID)

			l.SetResult(itemCtx, l.Meta(), request.IncludeResource, rd, &result)
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

type listTransitGatewayRouteTablePropagationModel struct {
	framework.WithRegionModel
	TransitGatewayRouteTableID types.String `tfsdk:"transit_gateway_route_table_id"`
}

func listTransitGatewayRouteTablePropagations(ctx context.Context, conn *ec2.Client, routeTableID string) iter.Seq2[awstypes.TransitGatewayRouteTablePropagation, error] {
	return func(yield func(awstypes.TransitGatewayRouteTablePropagation, error) bool) {
		input := &ec2.GetTransitGatewayRouteTablePropagationsInput{
			TransitGatewayRouteTableId: aws.String(routeTableID),
		}
		pages := ec2.NewGetTransitGatewayRouteTablePropagationsPaginator(conn, input)

		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.TransitGatewayRouteTablePropagation{}, fmt.Errorf("listing EC2 Transit Gateway Route Table Propagation resources for route table (%s): %w", routeTableID, err))
				return
			}

			for _, propagation := range page.TransitGatewayRouteTablePropagations {
				if !yield(propagation, nil) {
					return
				}
			}
		}
	}
}
