package backup

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for Backup changes to propagate
	propagationTimeout = 2 * time.Minute
)

func WaitJobCompleted(ctx context.Context, conn *backup.Backup, id string, timeout time.Duration) (*backup.DescribeBackupJobOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{backup.JobStateCreated, backup.JobStatePending, backup.JobStateRunning, backup.JobStateAborting},
		Target:  []string{backup.JobStateCompleted},
		Refresh: statusJobState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.DescribeBackupJobOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitFrameworkCreated(ctx context.Context, conn *backup.Backup, id string, timeout time.Duration) (*backup.DescribeFrameworkOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{frameworkStatusCreationInProgress},
		Target:  []string{frameworkStatusCompleted, frameworkStatusFailed},
		Refresh: statusFramework(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.DescribeFrameworkOutput); ok {
		return output, err
	}

	return nil, err
}

func waitFrameworkUpdated(ctx context.Context, conn *backup.Backup, id string, timeout time.Duration) (*backup.DescribeFrameworkOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{frameworkStatusUpdateInProgress},
		Target:  []string{frameworkStatusCompleted, frameworkStatusFailed},
		Refresh: statusFramework(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.DescribeFrameworkOutput); ok {
		return output, err
	}

	return nil, err
}

func waitFrameworkDeleted(ctx context.Context, conn *backup.Backup, id string, timeout time.Duration) (*backup.DescribeFrameworkOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{frameworkStatusDeletionInProgress},
		Target:  []string{backup.ErrCodeResourceNotFoundException},
		Refresh: statusFramework(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.DescribeFrameworkOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRecoveryPointDeleted(ctx context.Context, conn *backup.Backup, backupVaultName, recoveryPointARN string, timeout time.Duration) (*backup.DescribeRecoveryPointOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{backup.RecoveryPointStatusDeleting},
		Target:  []string{},
		Refresh: statusRecoveryPoint(ctx, conn, backupVaultName, recoveryPointARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.DescribeRecoveryPointOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
