package opsworks

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	InstanceDeleteTimeout = 2 * time.Minute
)

func waitInstanceDeleted(ctx context.Context, conn *opsworks.OpsWorks, instanceId string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{instanceStatusStopped, instanceStatusTerminating, instanceStatusTerminated},
		Target:     []string{},
		Refresh:    InstanceStatus(ctx, conn, instanceId),
		Timeout:    InstanceDeleteTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func waitInstanceStarted(ctx context.Context, conn *opsworks.OpsWorks, instanceId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{instanceStatusRequested, instanceStatusPending, instanceStatusBooting, instanceStatusRunningSetup},
		Target:     []string{instanceStatusOnline},
		Refresh:    InstanceStatus(ctx, conn, instanceId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func waitInstanceStopped(ctx context.Context, conn *opsworks.OpsWorks, instanceId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{instanceStatusStopping, instanceStatusTerminating, instanceStatusShuttingDown, instanceStatusTerminated},
		Target:     []string{instanceStatusStopped},
		Refresh:    InstanceStatus(ctx, conn, instanceId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
