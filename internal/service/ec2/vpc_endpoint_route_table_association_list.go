// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_vpc_endpoint_route_table_association")
func newVPCEndpointRouteTableAssociationResourceAsListResource() inttypes.ListResourceForSDK {
	l := vpcEndpointRouteTableAssociationListResource{}
	l.SetResourceSchema(resourceVPCEndpointRouteTableAssociation())
	return &l
}

var _ list.ListResource = &vpcEndpointRouteTableAssociationListResource{}

type vpcEndpointRouteTableAssociationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *vpcEndpointRouteTableAssociationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	tflog.Info(ctx, "Listing Resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ec2.DescribeVpcEndpointsInput{}
		for endpoint, err := range listVPCEndpoints(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			// Only Gateway-type endpoints have route table associations.
			if endpoint.VpcEndpointType != awstypes.VpcEndpointTypeGateway {
				continue
			}

			endpointID := aws.ToString(endpoint.VpcEndpointId)

			for _, routeTableID := range endpoint.RouteTableIds {
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrVPCEndpointID), endpointID)
				ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("route_table_id"), routeTableID)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.SetId(vpcEndpointRouteTableAssociationCreateID(endpointID, routeTableID))
				rd.Set(names.AttrVPCEndpointID, endpointID)
				rd.Set("route_table_id", routeTableID)

				if request.IncludeResource { //nolint:revive,staticcheck // Be explicit about IncludeResource handling.
					// No-op, all readable attributes are already populated above.
				}

				result.DisplayName = fmt.Sprintf("%s / %s", endpointID, routeTableID)

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
