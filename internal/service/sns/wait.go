package sns

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	subscriptionCreateTimeout              = 2 * time.Minute
	subscriptionPendingConfirmationTimeout = 2 * time.Minute
	subscriptionDeleteTimeout              = 2 * time.Minute
)

func waitSubscriptionConfirmed(conn *sns.SNS, arn string, timeout time.Duration) (map[string]string, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"true"},
		Target:  []string{"false"},
		Refresh: statusSubscriptionPendingConfirmation(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(map[string]string); ok {
		return output, err
	}

	return nil, err
}

func waitSubscriptionDeleted(conn *sns.SNS, arn string) (map[string]string, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"false", "true"},
		Target:  []string{},
		Refresh: statusSubscriptionPendingConfirmation(conn, arn),
		Timeout: subscriptionDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(map[string]string); ok {
		return output, err
	}

	return nil, err
}

const (
	topicPutAttributeTimeout = 2 * time.Minute
)
