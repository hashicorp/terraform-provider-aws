// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_ec2_transit_gateway_metering_policy")
func newTransitGatewayMeteringPolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &transitGatewayMeteringPolicyListResource{}
}

var _ list.ListResource = &transitGatewayMeteringPolicyListResource{}

type transitGatewayMeteringPolicyListResource struct {
	transitGatewayMeteringPolicyResource
	framework.WithList
}

func (l *transitGatewayMeteringPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	c := l.Meta()
	conn := c.EC2Client(ctx)

	var query listTransitGatewayMeteringPolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing EC2 Transit Gateway Metering Policies")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input ec2.DescribeTransitGatewayMeteringPoliciesInput
		for item, err := range listTransitGatewayMeteringPolicies(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.TransitGatewayMeteringPolicyId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data transitGatewayMeteringPolicyResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, c, &item, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = fwflex.StringValueFromFramework(ctx, data.ARN)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listTransitGatewayMeteringPolicyModel struct {
	framework.WithRegionModel
}
