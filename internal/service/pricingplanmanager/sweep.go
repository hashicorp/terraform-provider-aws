// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pricingplanmanager

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricingplanmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pricingplanmanager/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_pricingplanmanager_subscription", sweepSubscriptions)
}

func sweepSubscriptions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := pricingplanmanager.ListSubscriptionsInput{}
	conn := client.PricingPlanManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := pricingplanmanager.NewListSubscriptionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.SubscriptionSummaries {
			// Cancellation of an active subscription is already terminal:
			// it takes effect at the end of the current billing period.
			if sc := v.ScheduledChange; sc != nil && sc.ChangeType == awstypes.ScheduledChangeTypeCancellation {
				continue
			}

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newSubscriptionResource, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}
