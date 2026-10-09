// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pricingplanmanager

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricingplanmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pricingplanmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_pricingplanmanager_subscription")
func newSubscriptionResourceAsListResource() list.ListResourceWithConfigure {
	return &subscriptionListResource{}
}

var _ list.ListResource = &subscriptionListResource{}

type subscriptionListResource struct {
	subscriptionResource
	framework.WithList
}

func (l *subscriptionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PricingPlanManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input pricingplanmanager.ListSubscriptionsInput
		for item, err := range listSubscriptions(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}
			arn := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			var data subscriptionResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				output := pricingplanmanager.GetSubscriptionOutput{
					ETag: item.ETag,
					Subscription: &awstypes.Subscription{
						Arn:             item.Arn,
						CreatedAt:       item.CreatedAt,
						PlanFamily:      item.PlanFamily,
						PlanTier:        item.PlanTier,
						ResourceArns:    item.ResourceArns,
						ScheduledChange: item.ScheduledChange,
						Status:          item.Status,
						StatusReason:    item.StatusReason,
						UpdatedAt:       item.UpdatedAt,
						UsageLevel:      item.UsageLevel,
					},
				}
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &output, &data), smerr.ID, arn)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = arn
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listSubscriptions(ctx context.Context, conn *pricingplanmanager.Client, input *pricingplanmanager.ListSubscriptionsInput, optFns ...func(*pricingplanmanager.Options)) iter.Seq2[awstypes.SubscriptionSummary, error] {
	return tfiter.ConcatValuesWithError(listSubscriptionPages(ctx, conn, input, optFns...))
}

func listSubscriptionPages(ctx context.Context, conn *pricingplanmanager.Client, input *pricingplanmanager.ListSubscriptionsInput, optFns ...func(*pricingplanmanager.Options)) iter.Seq2[[]awstypes.SubscriptionSummary, error] {
	return func(yield func([]awstypes.SubscriptionSummary, error) bool) {
		pages := pricingplanmanager.NewListSubscriptionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Pricing Plan Manager Subscriptions: %w", err))
				return
			}

			if !yield(page.SubscriptionSummaries, nil) {
				return
			}
		}
	}
}
