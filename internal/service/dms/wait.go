package dms

import (
	"context"
	"time"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	propagationTimeout            = 2 * time.Minute
	replicationTaskRunningTimeout = 5 * time.Minute
)

func waitEndpointDeleted(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{endpointStatusDeleting},
		Target:  []string{},
		Refresh: statusEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskDeleted(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusDeleting},
		Target:     []string{},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskModified(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusModifying},
		Target:     []string{replicationTaskStatusReady, replicationTaskStatusStopped, replicationTaskStatusFailed},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskReady(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusCreating},
		Target:     []string{replicationTaskStatusReady},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskRunning(ctx context.Context, conn *dms.DatabaseMigrationService, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusStarting},
		Target:     []string{replicationTaskStatusRunning},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    replicationTaskRunningTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskStopped(ctx context.Context, conn *dms.DatabaseMigrationService, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{replicationTaskStatusStopping, replicationTaskStatusRunning},
		Target:     []string{replicationTaskStatusStopped},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    replicationTaskRunningTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      60 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
