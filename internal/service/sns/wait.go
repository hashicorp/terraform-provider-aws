package sns

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	subscriptionCreateTimeout              = 2 * time.Minute
	subscriptionPendingConfirmationTimeout = 2 * time.Minute
	subscriptionDeleteTimeout              = 2 * time.Minute
)

func waitSubscriptionConfirmed(conn *sns.SNS, id, expectedValue string, timeout time.Duration) (*sns.GetSubscriptionAttributesOutput, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{expectedValue},
		Refresh: statusSubscriptionPendingConfirmation(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sns.GetSubscriptionAttributesOutput); ok {
		return output, err
	}

	return nil, err
}

func waitSubscriptionDeleted(conn *sns.SNS, id string) (*sns.GetSubscriptionAttributesOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"false", "true"},
		Target:  []string{},
		Refresh: statusSubscriptionPendingConfirmation(conn, id),
		Timeout: subscriptionDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sns.GetSubscriptionAttributesOutput); ok {
		return output, err
	}

	return nil, err
}
