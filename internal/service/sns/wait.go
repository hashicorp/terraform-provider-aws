package sns

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	subscriptionCreateTimeout              = 2 * time.Minute
	subscriptionPendingConfirmationTimeout = 2 * time.Minute
	subscriptionDeleteTimeout              = 2 * time.Minute
)

func waitSubscriptionConfirmed(ctx context.Context, conn *sns.SNS, arn string, timeout time.Duration) (map[string]string, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"true"},
		Target:  []string{"false"},
		Refresh: statusSubscriptionPendingConfirmation(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(map[string]string); ok {
		return output, err
	}

	return nil, err
}

func waitSubscriptionDeleted(ctx context.Context, conn *sns.SNS, arn string) (map[string]string, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"false", "true"},
		Target:  []string{},
		Refresh: statusSubscriptionPendingConfirmation(ctx, conn, arn),
		Timeout: subscriptionDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(map[string]string); ok {
		return output, err
	}

	return nil, err
}

const (
	topicPutAttributeTimeout = 2 * time.Minute
)
