package kinesisanalyticsv2

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// waitApplicationDeleted waits for an Application to return Deleted
func waitApplicationDeleted(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string, timeout time.Duration) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusDeleting},
		Target:  []string{},
		Refresh: statusApplication(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationStarted waits for an Application to start
func waitApplicationStarted(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string, timeout time.Duration) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusStarting},
		Target:  []string{kinesisanalyticsv2.ApplicationStatusRunning},
		Refresh: statusApplication(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationStopped waits for an Application to stop
func waitApplicationStopped(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string, timeout time.Duration) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusForceStopping, kinesisanalyticsv2.ApplicationStatusStopping},
		Target:  []string{kinesisanalyticsv2.ApplicationStatusReady},
		Refresh: statusApplication(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationUpdated waits for an Application to return Deleted
func waitApplicationUpdated(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string, timeout time.Duration) (*kinesisanalyticsv2.ApplicationDetail, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusUpdating},
		Target:  []string{kinesisanalyticsv2.ApplicationStatusReady, kinesisanalyticsv2.ApplicationStatusRunning},
		Refresh: statusApplication(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitIAMPropagation retries the specified function if the returned error indicates an IAM eventual consistency issue.
// If the retries time out the specified function is called one last time.
func waitIAMPropagation(f func() (interface{}, error)) (interface{}, error) {
	var output interface{}

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		output, err = f()

		// Kinesis Stream: https://github.com/hashicorp/terraform-provider-aws/issues/7032
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
			return resource.RetryableError(err)
		}

		// Kinesis Firehose: https://github.com/hashicorp/terraform-provider-aws/issues/7394
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics doesn't have sufficient privileges") {
			return resource.RetryableError(err)
		}

		// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
			return resource.RetryableError(err)
		}

		// S3: https://github.com/hashicorp/terraform-provider-aws/issues/16104
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Please check the role provided or validity of S3 location you provided") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = f()
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// waitSnapshotCreated waits for a Snapshot to return Created
func waitSnapshotCreated(conn *kinesisanalyticsv2.KinesisAnalyticsV2, applicationName, snapshotName string, timeout time.Duration) (*kinesisanalyticsv2.SnapshotDetails, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.SnapshotStatusCreating},
		Target:  []string{kinesisanalyticsv2.SnapshotStatusReady},
		Refresh: statusSnapshotDetails(conn, applicationName, snapshotName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.SnapshotDetails); ok {
		return v, err
	}

	return nil, err
}

// waitSnapshotDeleted waits for a Snapshot to return Deleted
func waitSnapshotDeleted(conn *kinesisanalyticsv2.KinesisAnalyticsV2, applicationName, snapshotName string, timeout time.Duration) (*kinesisanalyticsv2.SnapshotDetails, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.SnapshotStatusDeleting},
		Target:  []string{},
		Refresh: statusSnapshotDetails(conn, applicationName, snapshotName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.SnapshotDetails); ok {
		return v, err
	}

	return nil, err
}
