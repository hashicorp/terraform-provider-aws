// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/ec2/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
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
