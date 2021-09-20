package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/rds/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// EventSubscription NotFound
	EventSubscriptionStatusNotFound = "NotFound"

	// EventSubscription Unknown
	EventSubscriptionStatusUnknown = "Unknown"

	// ProxyEndpoint NotFound
	ProxyEndpointStatusNotFound = "NotFound"

	// ProxyEndpoint Unknown
	ProxyEndpointStatusUnknown = "Unknown"
)

// EventSubscriptionStatus fetches the EventSubscription and its Status
func EventSubscriptionStatus(conn *rds.RDS, subscriptionName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &rds.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(subscriptionName),
		}

		output, err := conn.DescribeEventSubscriptions(input)

		if err != nil {
			return nil, EventSubscriptionStatusUnknown, err
		}

		if len(output.EventSubscriptionsList) == 0 {
			return nil, EventSubscriptionStatusNotFound, nil
		}

		return output.EventSubscriptionsList[0], aws.StringValue(output.EventSubscriptionsList[0].Status), nil
	}
}

// DBProxyEndpointStatus fetches the ProxyEndpoint and its Status
func DBProxyEndpointStatus(conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.DBProxyEndpoint(conn, id)

		if err != nil {
			return nil, ProxyEndpointStatusUnknown, err
		}

		if output == nil {
			return nil, ProxyEndpointStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func DBClusterRoleStatus(conn *rds.RDS, dbClusterID, roleARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.DBClusterRoleByDBClusterIDAndRoleARN(conn, dbClusterID, roleARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
