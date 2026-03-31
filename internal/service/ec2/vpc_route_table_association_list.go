// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
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

			// TODO: Name tag display values
			// TODO: Gateway display values
			subnetDisplayName := aws.ToString(item.SubnetId)
			routeTableDisplayName := aws.ToString(item.RouteTableId)

			result.DisplayName = fmt.Sprintf("%s / %s (%s)", subnetDisplayName, routeTableDisplayName, aws.ToString(item.RouteTableAssociationId))

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
