// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_prometheus_scraper_logging_configuration")
func newScraperLoggingConfigurationResourceAsListResource() list.ListResourceWithConfigure {
	return &scraperLoggingConfigurationListResource{}
}

var _ list.ListResource = &scraperLoggingConfigurationListResource{}

type scraperLoggingConfigurationListResource struct {
	scraperLoggingConfigurationResource
	framework.WithList
}

func (l *scraperLoggingConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().AMPClient(ctx)

	var query listScraperLoggingConfigurationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input amp.ListScrapersInput
		for scraper, err := range listScrapers(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			scraperID := aws.ToString(scraper.ScraperId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), scraperID)

			item, err := findScraperLoggingConfigurationByID(ctx, conn, scraperID)
			if retry.NotFound(err) {
				tflog.Debug(ctx, "Scraper has no Logging Configuration, skipping")
				continue
			}
			if err != nil {
				tflog.Error(ctx, "Reading Prometheus Scraper Logging Configuration", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			result := request.NewListResult(ctx)

			var data scraperLoggingConfigurationResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ScraperID = fwflex.StringValueToFramework(ctx, scraperID)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, item, &data)...)
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

type listScraperLoggingConfigurationModel struct {
	framework.WithRegionModel
}
