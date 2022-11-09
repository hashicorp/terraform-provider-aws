//go:build sweep
// +build sweep

package neptune

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_neptune_event_subscription", &resource.Sweeper{
		Name: "aws_neptune_event_subscription",
		F:    sweepEventSubscriptions,
	})
}

func sweepEventSubscriptions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).NeptuneConn
	var sweeperErrs *multierror.Error

	err = conn.DescribeEventSubscriptionsPages(&neptune.DescribeEventSubscriptionsInput{}, func(page *neptune.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventSubscription := range page.EventSubscriptionsList {
			name := aws.StringValue(eventSubscription.CustSubscriptionId)

			log.Printf("[INFO] Deleting Neptune Event Subscription: %s", name)
			_, err = conn.DeleteEventSubscription(&neptune.DeleteEventSubscriptionInput{
				SubscriptionName: aws.String(name),
			})
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeSubscriptionNotFoundFault) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("deleting Neptune Event Subscription (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = WaitEventSubscriptionDeleted(conn, name)
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeSubscriptionNotFoundFault) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("waiting for Neptune Event Subscription (%s) deletion: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Neptune Event Subscriptions sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving Neptune Event Subscriptions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
