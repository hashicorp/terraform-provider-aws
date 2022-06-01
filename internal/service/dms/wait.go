package dms

import (
	"time"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	replicationTaskRunningTimeout = 5 * time.Minute
)

func waitEndpointDeleted(conn *dms.DatabaseMigrationService, id string, timeout time.Duration) (*dms.Endpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{endpointStatusDeleting},
		Target:  []string{},
		Refresh: statusEndpoint(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dms.Endpoint); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationTaskDeleted(conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusDeleting},
		Target:     []string{},
		Refresh:    statusReplicationTask(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForState()

	return err
}

func waitReplicationTaskModified(conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusModifying},
		Target:     []string{replicationTaskStatusReady, replicationTaskStatusStopped, replicationTaskStatusFailed},
		Refresh:    statusReplicationTask(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForState()

	return err
}

func waitReplicationTaskReady(conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusCreating},
		Target:     []string{replicationTaskStatusReady},
		Refresh:    statusReplicationTask(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForState()

	return err
}

func waitReplicationTaskRunning(conn *dms.DatabaseMigrationService, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusStarting},
		Target:     []string{replicationTaskStatusRunning},
		Refresh:    statusReplicationTask(conn, id),
		Timeout:    replicationTaskRunningTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForState()

	return err
}

func waitReplicationTaskStopped(conn *dms.DatabaseMigrationService, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusStopping},
		Target:     []string{replicationTaskStatusStopped},
		Refresh:    statusReplicationTask(conn, id),
		Timeout:    replicationTaskRunningTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      60 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForState()

	return err
}
