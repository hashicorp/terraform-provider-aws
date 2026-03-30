// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkListResource("aws_lambda_capacity_provider")
func newCapacityProviderResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceCapacityProvider{}
}

var _ list.ListResource = &listResourceCapacityProvider{}

type listResourceCapacityProvider struct {
	resourceCapacityProvider
	framework.WithList
}

func (r *listResourceCapacityProvider) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().LambdaClient(ctx)
	var query capacityProviderListModel

	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input lambda.ListCapacityProvidersInput
		for capacityProvider, err := range listCapacityProviders(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			if capacityProvider.State == awstypes.CapacityProviderStateDeleting {
				continue
			}

			var data resourceCapacityProviderModel
			r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
				if diags := flex.Flatten(ctx, capacityProvider, &data, flex.WithFieldNamePrefix(capacityProviderNamePrefix)); diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				cpARN, err := arn.Parse(data.ARN.ValueString())
				if err != nil {
					result = fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}

				name := strings.TrimPrefix(cpARN.Resource, "capacity-provider:")
				data.Name = flex.StringValueToFramework(ctx, name)
				result.DisplayName = name
			})

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

type capacityProviderListModel struct {
	framework.WithRegionModel
}

func listCapacityProviders(ctx context.Context, conn *lambda.Client, input *lambda.ListCapacityProvidersInput) iter.Seq2[awstypes.CapacityProvider, error] {
	return func(yield func(awstypes.CapacityProvider, error) bool) {
		pages := lambda.NewListCapacityProvidersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.CapacityProvider{}, fmt.Errorf("listing Lambda Capacity Providers: %w", err))
				return
			}

			for _, cp := range page.CapacityProviders {
				if !yield(cp, nil) {
					return
				}
			}
		}
	}
}
