package kinesisanalytics

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	applicationDeletedTimeout = 5 * time.Minute
	applicationStartedTimeout = 5 * time.Minute
	applicationStoppedTimeout = 5 * time.Minute
	applicationUpdatedTimeout = 5 * time.Minute
)

// waitApplicationDeleted waits for an Application to return Deleted
func waitApplicationDeleted(conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalytics.ApplicationStatusDeleting},
		Target:  []string{},
		Refresh: statusApplication(conn, name),
		Timeout: applicationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalytics.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationStarted waits for an Application to start
func waitApplicationStarted(conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalytics.ApplicationStatusStarting},
		Target:  []string{kinesisanalytics.ApplicationStatusRunning},
		Refresh: statusApplication(conn, name),
		Timeout: applicationStartedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalytics.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationStopped waits for an Application to stop
func waitApplicationStopped(conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalytics.ApplicationStatusStopping},
		Target:  []string{kinesisanalytics.ApplicationStatusReady},
		Refresh: statusApplication(conn, name),
		Timeout: applicationStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalytics.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationUpdated waits for an Application to update
func waitApplicationUpdated(conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalytics.ApplicationStatusUpdating},
		Target:  []string{kinesisanalytics.ApplicationStatusReady, kinesisanalytics.ApplicationStatusRunning},
		Refresh: statusApplication(conn, name),
		Timeout: applicationUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalytics.ApplicationDetail); ok {
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
		if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
			return resource.RetryableError(err)
		}

		// Kinesis Firehose: https://github.com/hashicorp/terraform-provider-aws/issues/7394
		if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Kinesis Analytics doesn't have sufficient privileges") {
			return resource.RetryableError(err)
		}

		// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
		if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
			return resource.RetryableError(err)
		}

		// S3: https://github.com/hashicorp/terraform-provider-aws/issues/16104
		if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Please check the role provided or validity of S3 location you provided") {
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
