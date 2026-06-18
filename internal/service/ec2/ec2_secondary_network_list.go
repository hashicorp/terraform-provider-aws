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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkListResource("aws_ec2_secondary_network")
func newSecondaryNetworkResourceAsListResource() list.ListResourceWithConfigure {
	return &secondaryNetworkListResource{}
}

var _ list.ListResource = &secondaryNetworkListResource{}

type secondaryNetworkListResource struct {
	secondaryNetworkResource
	framework.WithList
}

func (r *secondaryNetworkListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().EC2Client(ctx)

	var query listSecondaryNetworkModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input ec2.DescribeSecondaryNetworksInput
		for item, err := range listSecondaryNetworks(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data secondaryNetworkResourceModel

			r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
				if diags := fwflex.Flatten(ctx, item, &data, fwflex.WithFieldNamePrefix("SecondaryNetwork")); diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				id := aws.ToString(item.SecondaryNetworkId)
				data.ID = fwflex.StringValueToFramework(ctx, id)
				result.DisplayName = id

				// Fields with mismatched names missed by AutoFlex
				data.NetworkType = fwflex.StringValueToFramework(ctx, item.Type)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listSecondaryNetworkModel struct {
	framework.WithRegionModel
}

func listSecondaryNetworks(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecondaryNetworksInput) iter.Seq2[awstypes.SecondaryNetwork, error] {
	return func(yield func(awstypes.SecondaryNetwork, error) bool) {
		pages := ec2.NewDescribeSecondaryNetworksPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.SecondaryNetwork{}, fmt.Errorf("listing EC2 Secondary Network resources: %w", err))
				return
			}

			for _, item := range page.SecondaryNetworks {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
