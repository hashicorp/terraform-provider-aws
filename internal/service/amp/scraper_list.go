// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_prometheus_scraper")
func newScraperResourceAsListResource() list.ListResourceWithConfigure {
	return &scraperListResource{}
}

var _ list.ListResource = &scraperListResource{}

type scraperListResource struct {
	scraperResource
	framework.WithList
}

func (l *scraperListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"filters": listschema.MapAttribute{
				CustomType:  fwtypes.MapOfListOfStringType,
				Optional:    true,
				Description: "List of key-value pairs to filter the list of scrapers returned.",
			},
		},
	}
}

func (l *scraperListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().AMPClient(ctx)

	var query listScraperModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := amp.ListScrapersInput{
			Filters: fwflex.ExpandFrameworkStringValueListMap(ctx, query.Filters),
		}
		for item, err := range listScrapers(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), aws.ToString(item.Arn))

			scraperID := aws.ToString(item.ScraperId)
			result := request.NewListResult(ctx)

			var data scraperResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = fwflex.StringValueToFramework(ctx, scraperID)

				if request.IncludeResource {
					scraper, err := findScraperByID(ctx, conn, scraperID)
					if retry.NotFound(err) {
						return
					}
					if err != nil {
						result = fwdiag.NewListResultErrorDiagnostic(err)
						return
					}

					result.Diagnostics.Append(l.flatten(ctx, scraper, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = scraperID
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listScraperModel struct {
	framework.WithRegionModel
	Filters fwtypes.MapOfListOfString `tfsdk:"filters"`
}

func listScrapers(ctx context.Context, conn *amp.Client, input *amp.ListScrapersInput, optFns ...func(*amp.Options)) iter.Seq2[awstypes.ScraperSummary, error] {
	return tfiter.ConcatValuesWithError(listScraperPages(ctx, conn, input, optFns...))
}

func listScraperPages(ctx context.Context, conn *amp.Client, input *amp.ListScrapersInput, optFns ...func(*amp.Options)) iter.Seq2[[]awstypes.ScraperSummary, error] {
	return func(yield func([]awstypes.ScraperSummary, error) bool) {
		pages := amp.NewListScrapersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Prometheus Scrapers: %w", err))
				return
			}

			if !yield(page.Scrapers, nil) {
				return
			}
		}
	}
}
