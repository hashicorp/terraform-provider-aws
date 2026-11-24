// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_continuum_integration")
func newIntegrationResourceAsListResource() list.ListResourceWithConfigure {
	return &integrationListResource{}
}

var _ list.ListResource = &integrationListResource{}

type integrationListResource struct {
	integrationResource
	framework.WithList
}

func (l *integrationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ContinuumClient(ctx)

	tflog.Info(ctx, "Listing Continuum Integrations")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input continuum.ListIntegrationsInput
		for item, err := range listIntegrations(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			id := aws.ToString(item.IntegrationId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data integrationResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, data.flatten(ctx, &item, item.ProviderDetails), smerr.ID, id)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listIntegrations(ctx context.Context, conn *continuum.Client, input *continuum.ListIntegrationsInput) iter.Seq2[awstypes.IntegrationSummary, error] {
	return func(yield func(awstypes.IntegrationSummary, error) bool) {
		pages := continuum.NewListIntegrationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.IntegrationSummary{}, smarterr.NewError(err))
				return
			}

			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
