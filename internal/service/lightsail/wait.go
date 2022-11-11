package lightsail

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// OperationTimeout is the Timout Value for Operations
	OperationTimeout = 20 * time.Minute
	// OperationDelay is the Delay Value for Operations
	OperationDelay = 5 * time.Second
	// OperationMinTimeout is the MinTimout Value for Operations
	OperationMinTimeout = 3 * time.Second

	// DatabaseStateModifying is a state value for a Relational Database undergoing a modification
	DatabaseStateModifying = "modifying"
	// DatabaseStateAvailable is a state value for a Relational Database available for modification
	DatabaseStateAvailable = "available"

	// DatabaseTimeout is the Timout Value for Relational Database Modifications
	DatabaseTimeout = 20 * time.Minute
	// DatabaseDelay is the Delay Value for Relational Database Modifications
	DatabaseDelay = 5 * time.Second
	// DatabaseMinTimeout is the MinTimout Value for Relational Database Modifications
	DatabaseMinTimeout = 3 * time.Second
)

// waitOperation waits for an Operation to return Succeeded or Compleated
func waitOperation(conn *lightsail.Lightsail, oid *string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.OperationStatusStarted},
		Target:     []string{lightsail.OperationStatusCompleted, lightsail.OperationStatusSucceeded},
		Refresh:    statusOperation(conn, oid),
		Timeout:    OperationTimeout,
		Delay:      OperationDelay,
		MinTimeout: OperationMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if _, ok := outputRaw.(*lightsail.GetOperationOutput); ok {
		return err
	}

	return err
}

// waitDatabaseModified waits for a Modified Database return available
func waitDatabaseModified(conn *lightsail.Lightsail, db *string) (*lightsail.GetRelationalDatabaseOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{DatabaseStateModifying},
		Target:     []string{DatabaseStateAvailable},
		Refresh:    statusDatabase(conn, db),
		Timeout:    DatabaseTimeout,
		Delay:      DatabaseDelay,
		MinTimeout: DatabaseMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lightsail.GetRelationalDatabaseOutput); ok {
		return output, err
	}

	return nil, err
}

// waitDatabaseBackupRetentionModified waits for a Modified  BackupRetention on Database return available

func waitDatabaseBackupRetentionModified(conn *lightsail.Lightsail, db *string, status *bool) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.FormatBool(!aws.BoolValue(status))},
		Target:     []string{strconv.FormatBool(aws.BoolValue(status))},
		Refresh:    statusDatabaseBackupRetention(conn, db),
		Timeout:    DatabaseTimeout,
		Delay:      DatabaseDelay,
		MinTimeout: DatabaseMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if _, ok := outputRaw.(*lightsail.GetRelationalDatabaseOutput); ok {
		return err
	}

	return err
}

func waitContainerServiceCreated(ctx context.Context, conn *lightsail.Lightsail, serviceName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceStatePending},
		Target:     []string{lightsail.ContainerServiceStateReady},
		Refresh:    statusContainerService(ctx, conn, serviceName),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerService); ok {
		if detail := output.StateDetail; detail != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(detail.Code), aws.StringValue(detail.Message)))
		}

		return err
	}

	return err
}

func waitContainerServiceDisabled(ctx context.Context, conn *lightsail.Lightsail, serviceName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceStateUpdating},
		Target:     []string{lightsail.ContainerServiceStateDisabled},
		Refresh:    statusContainerService(ctx, conn, serviceName),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerService); ok {
		if detail := output.StateDetail; detail != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(detail.Code), aws.StringValue(detail.Message)))
		}

		return err
	}

	return err
}

func waitContainerServiceUpdated(ctx context.Context, conn *lightsail.Lightsail, serviceName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceStateUpdating},
		Target:     []string{lightsail.ContainerServiceStateReady, lightsail.ContainerServiceStateRunning},
		Refresh:    statusContainerService(ctx, conn, serviceName),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerService); ok {
		if detail := output.StateDetail; detail != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(detail.Code), aws.StringValue(detail.Message)))
		}

		return err
	}

	return err
}

func waitContainerServiceDeleted(ctx context.Context, conn *lightsail.Lightsail, serviceName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceStateDeleting},
		Target:     []string{},
		Refresh:    statusContainerService(ctx, conn, serviceName),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerService); ok {
		if detail := output.StateDetail; detail != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(detail.Code), aws.StringValue(detail.Message)))
		}

		return err
	}

	return err
}

func waitContainerServiceDeploymentVersionActive(ctx context.Context, conn *lightsail.Lightsail, serviceName string, version int, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceDeploymentStateActivating},
		Target:     []string{lightsail.ContainerServiceDeploymentStateActive},
		Refresh:    statusContainerServiceDeploymentVersion(ctx, conn, serviceName, version),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerServiceDeployment); ok {
		if aws.StringValue(output.State) == lightsail.ContainerServiceDeploymentStateFailed {
			tfresource.SetLastError(err, errors.New("The deployment failed. Use the GetContainerLog action to view the log events for the containers in the deployment to try to determine the reason for the failure."))
		}

		return err
	}

	return err
}

func waitInstanceStateWithContext(ctx context.Context, conn *lightsail.Lightsail, id *string) (*lightsail.GetInstanceStateOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending", "stopping"},
		Target:     []string{"stopped", "running"},
		Refresh:    statusInstance(conn, id),
		Timeout:    OperationTimeout,
		Delay:      OperationDelay,
		MinTimeout: OperationMinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*lightsail.GetInstanceStateOutput); ok {
		return out, err
	}

	return nil, err
}
