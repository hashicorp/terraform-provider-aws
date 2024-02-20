// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalytics

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	applicationDeletedTimeout = 5 * time.Minute
	applicationStartedTimeout = 5 * time.Minute
	applicationStoppedTimeout = 5 * time.Minute
	applicationUpdatedTimeout = 5 * time.Minute
)

// waitApplicationDeleted waits for an Application to return Deleted
func waitApplicationDeleted(ctx context.Context, conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalytics.ApplicationStatusDeleting},
		Target:  []string{},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: applicationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalytics.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationStarted waits for an Application to start
func waitApplicationStarted(ctx context.Context, conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalytics.ApplicationStatusStarting},
		Target:  []string{kinesisanalytics.ApplicationStatusRunning},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: applicationStartedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalytics.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationStopped waits for an Application to stop
func waitApplicationStopped(ctx context.Context, conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalytics.ApplicationStatusStopping},
		Target:  []string{kinesisanalytics.ApplicationStatusReady},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: applicationStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalytics.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationUpdated waits for an Application to update
func waitApplicationUpdated(ctx context.Context, conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalytics.ApplicationStatusUpdating},
		Target:  []string{kinesisanalytics.ApplicationStatusReady, kinesisanalytics.ApplicationStatusRunning},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: applicationUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalytics.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitIAMPropagation retries the specified function if the returned error indicates an IAM eventual consistency issue.
// If the retries time out the specified function is called one last time.
func waitIAMPropagation(ctx context.Context, f func() (interface{}, error)) (interface{}, error) {
	var output interface{}

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		output, err = f()

		// Kinesis Stream: https://github.com/hashicorp/terraform-provider-aws/issues/7032
		if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
			return retry.RetryableError(err)
		}

		// Kinesis Firehose: https://github.com/hashicorp/terraform-provider-aws/issues/7394
		if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Kinesis Analytics doesn't have sufficient privileges") {
			return retry.RetryableError(err)
		}

		// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
		if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
			return retry.RetryableError(err)
		}

		// S3: https://github.com/hashicorp/terraform-provider-aws/issues/16104
		if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Please check the role provided or validity of S3 location you provided") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
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
