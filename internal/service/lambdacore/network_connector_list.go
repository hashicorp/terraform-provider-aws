// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdacore

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdacore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdacore/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_lambdacore_network_connector")
func newNetworkConnectorResourceAsListResource() list.ListResourceWithConfigure {
	return &networkConnectorListResource{}
}

var _ list.ListResource = &networkConnectorListResource{}

type networkConnectorListResource struct {
	networkConnectorResource
	framework.WithList
}

func (l *networkConnectorListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LambdaCoreClient(ctx)

	var query listNetworkConnectorModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input lambdacore.ListNetworkConnectorsInput
		for item, err := range listNetworkConnectors(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var out *lambdacore.GetNetworkConnectorOutput
			if request.IncludeResource {
				out, err = findNetworkConnectorByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data networkConnectorResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, out, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listNetworkConnectorModel struct {
	framework.WithRegionModel
}

func listNetworkConnectors(ctx context.Context, conn *lambdacore.Client, input *lambdacore.ListNetworkConnectorsInput, optFns ...func(*lambdacore.Options)) iter.Seq2[awstypes.NetworkConnectorSummary, error] {
	return tfiter.ConcatValuesWithError(listNetworkConnectorPages(ctx, conn, input, optFns...))
}

func listNetworkConnectorPages(ctx context.Context, conn *lambdacore.Client, input *lambdacore.ListNetworkConnectorsInput, optFns ...func(*lambdacore.Options)) iter.Seq2[[]awstypes.NetworkConnectorSummary, error] {
	return func(yield func([]awstypes.NetworkConnectorSummary, error) bool) {
		pages := lambdacore.NewListNetworkConnectorsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Lambda Core Network Connectors: %w", err))
				return
			}

			if !yield(page.NetworkConnectors, nil) {
				return
			}
		}
	}
}
