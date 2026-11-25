// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_route53_health_check_source")
func newHealthCheckSourceResourceAsListResource() list.ListResourceWithConfigure {
	return &healthCheckSourceListResource{}
}

var _ list.ListResource = &healthCheckSourceListResource{}

type healthCheckSourceListResource struct {
	healthCheckSourceResource
	framework.WithList
}

func (l *healthCheckSourceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().Route53Client(ctx)

	tflog.Info(ctx, "Listing Route 53 Health Check Sources")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input route53.ListHealthCheckSourcesInput
		for item, err := range listHealthCheckSources(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// Route 53 keeps listing a source for a while after it is deleted.
			if item.HealthCheckSourceStatus == awstypes.HealthCheckSourceStatusDeleteCompleted {
				continue
			}

			id := aws.ToString(item.HealthCheckSourceId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data healthCheckSourceResourceModel
			data.ID = fwflex.StringValueToFramework(ctx, id)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data), smerr.ID, id)
				if result.Diagnostics.HasError() {
					return
				}

				// A health check source has no name, so its ID is the most useful label.
				result.DisplayName = id
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listHealthCheckSources(ctx context.Context, conn *route53.Client, input *route53.ListHealthCheckSourcesInput) iter.Seq2[awstypes.HealthCheckSource, error] {
	return func(yield func(awstypes.HealthCheckSource, error) bool) {
		pages := route53.NewListHealthCheckSourcesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.HealthCheckSource{}, fmt.Errorf("listing Route 53 Health Check Sources: %w", err))
				return
			}

			for _, item := range page.HealthCheckSources {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
