// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
)

func findSubscriptionsByAssessmentTemplateARN(ctx context.Context, conn *inspector.Client, arn string) ([]awstypes.Subscription, error) {
	input := &inspector.ListEventSubscriptionsInput{
		ResourceArn: aws.String(arn),
	}

	var results []awstypes.Subscription

	pages := inspector.NewListEventSubscriptionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, subscription := range page.Subscriptions {
			if aws.ToString(subscription.ResourceArn) == arn {
				results = append(results, subscription)
			}
		}
	}

	return results, nil
}
