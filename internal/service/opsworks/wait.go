package opsworks

import (
	"time"

	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	InstanceDeleteTimeout = 2 * time.Minute
)

func waitInstanceDeleted(conn *opsworks.OpsWorks, instanceId string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{instanceStatusStopped, instanceStatusTerminating, instanceStatusTerminated},
		Target:     []string{},
		Refresh:    InstanceStatus(conn, instanceId),
		Timeout:    InstanceDeleteTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return err
}

func waitInstanceStarted(conn *opsworks.OpsWorks, instanceId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{instanceStatusRequested, instanceStatusPending, instanceStatusBooting, instanceStatusRunningSetup},
		Target:     []string{instanceStatusOnline},
		Refresh:    InstanceStatus(conn, instanceId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func waitInstanceStopped(conn *opsworks.OpsWorks, instanceId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{instanceStatusStopping, instanceStatusTerminating, instanceStatusShuttingDown, instanceStatusTerminated},
		Target:     []string{instanceStatusStopped},
		Refresh:    InstanceStatus(conn, instanceId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}
