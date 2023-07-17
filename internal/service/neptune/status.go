// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// EventSubscription NotFound
	EventSubscriptionStatusNotFound = "NotFound"

	// EventSubscription Unknown
	EventSubscriptionStatusUnknown = "Unknown"

	// DBClusterEndpoint Unknown
	DBClusterEndpointStatusUnknown = "Unknown"
)

// StatusEventSubscription fetches the EventSubscription and its Status
func StatusEventSubscription(ctx context.Context, conn *neptune.Neptune, subscriptionName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &neptune.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(subscriptionName),
		}

		output, err := conn.DescribeEventSubscriptionsWithContext(ctx, input)

		if err != nil {
			return nil, EventSubscriptionStatusUnknown, err
		}

		if len(output.EventSubscriptionsList) == 0 {
			return nil, EventSubscriptionStatusNotFound, nil
		}

		return output.EventSubscriptionsList[0], aws.StringValue(output.EventSubscriptionsList[0].Status), nil
	}
}

// StatusDBClusterEndpoint fetches the DBClusterEndpoint and its Status
func StatusDBClusterEndpoint(ctx context.Context, conn *neptune.Neptune, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, DBClusterEndpointStatusUnknown, err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
