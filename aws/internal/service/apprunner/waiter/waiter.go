package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AutoScalingConfigurationStatusActive   = "active"
	AutoScalingConfigurationStatusInactive = "inactive"

	AutoScalingConfigurationCreateTimeout = 2 * time.Minute
	AutoScalingConfigurationDeleteTimeout = 2 * time.Minute
)

func AutoScalingConfigurationActive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{AutoScalingConfigurationStatusActive},
		Refresh: AutoScalingConfigurationStatus(ctx, conn, arn),
		Timeout: AutoScalingConfigurationCreateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func AutoScalingConfigurationInactive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{AutoScalingConfigurationStatusActive},
		Target:  []string{AutoScalingConfigurationStatusInactive},
		Refresh: AutoScalingConfigurationStatus(ctx, conn, arn),
		Timeout: AutoScalingConfigurationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}
