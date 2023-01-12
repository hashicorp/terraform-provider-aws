package inspector

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector"
)

func findSubscriptionsByAssessmentTemplateARN(conn *inspector.Inspector, arn string) ([]*inspector.Subscription, error) {
	input := &inspector.ListEventSubscriptionsInput{
		ResourceArn: aws.String(arn),
	}

	var results []*inspector.Subscription

	err := conn.ListEventSubscriptionsPages(input, func(page *inspector.ListEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, subscription := range page.Subscriptions {
			if subscription == nil {
				continue
			}

			if aws.StringValue(subscription.ResourceArn) == arn {
				results = append(results, subscription)
			}
		}

		return !lastPage
	})

	return results, err
}
