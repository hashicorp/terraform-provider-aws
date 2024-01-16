// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// waitApplicationDeleted waits for an Application to return Deleted
func waitApplicationDeleted(ctx context.Context, conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string, timeout time.Duration) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusDeleting},
		Target:  []string{},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationStarted waits for an Application to start
func waitApplicationStarted(ctx context.Context, conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string, timeout time.Duration) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusStarting},
		Target:  []string{kinesisanalyticsv2.ApplicationStatusRunning},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationStopped waits for an Application to stop
func waitApplicationStopped(ctx context.Context, conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string, timeout time.Duration) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusForceStopping, kinesisanalyticsv2.ApplicationStatusStopping},
		Target:  []string{kinesisanalyticsv2.ApplicationStatusReady},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// waitApplicationUpdated waits for an Application to return Deleted
func waitApplicationUpdated(ctx context.Context, conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string, timeout time.Duration) (*kinesisanalyticsv2.ApplicationDetail, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusUpdating},
		Target:  []string{kinesisanalyticsv2.ApplicationStatusReady, kinesisanalyticsv2.ApplicationStatusRunning},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
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
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
			return retry.RetryableError(err)
		}

		// Kinesis Firehose: https://github.com/hashicorp/terraform-provider-aws/issues/7394
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics doesn't have sufficient privileges") {
			return retry.RetryableError(err)
		}

		// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
			return retry.RetryableError(err)
		}

		// S3: https://github.com/hashicorp/terraform-provider-aws/issues/16104
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Please check the role provided or validity of S3 location you provided") {
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

// waitSnapshotCreated waits for a Snapshot to return Created
func waitSnapshotCreated(ctx context.Context, conn *kinesisanalyticsv2.KinesisAnalyticsV2, applicationName, snapshotName string, timeout time.Duration) (*kinesisanalyticsv2.SnapshotDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.SnapshotStatusCreating},
		Target:  []string{kinesisanalyticsv2.SnapshotStatusReady},
		Refresh: statusSnapshotDetails(ctx, conn, applicationName, snapshotName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalyticsv2.SnapshotDetails); ok {
		return v, err
	}

	return nil, err
}

// waitSnapshotDeleted waits for a Snapshot to return Deleted
func waitSnapshotDeleted(ctx context.Context, conn *kinesisanalyticsv2.KinesisAnalyticsV2, applicationName, snapshotName string, timeout time.Duration) (*kinesisanalyticsv2.SnapshotDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.SnapshotStatusDeleting},
		Target:  []string{},
		Refresh: statusSnapshotDetails(ctx, conn, applicationName, snapshotName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*kinesisanalyticsv2.SnapshotDetails); ok {
		return v, err
	}

	return nil, err
}
