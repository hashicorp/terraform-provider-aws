// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesisanalytics

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kinesisanalytics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalytics/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitApplicationDeleted(ctx context.Context, conn *kinesisanalytics.Client, name string) (*awstypes.ApplicationDetail, error) {
	const (
		applicationDeletedTimeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusDeleting),
		Target:  []string{},
		Refresh: statusApplication(conn, name),
		Timeout: applicationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

func waitApplicationStarted(ctx context.Context, conn *kinesisanalytics.Client, name string) (*awstypes.ApplicationDetail, error) {
	const (
		applicationStartedTimeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusStarting),
		Target:  enum.Slice(awstypes.ApplicationStatusRunning),
		Refresh: statusApplication(conn, name),
		Timeout: applicationStartedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

func waitApplicationStopped(ctx context.Context, conn *kinesisanalytics.Client, name string) (*awstypes.ApplicationDetail, error) {
	const (
		applicationStoppedTimeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusStopping),
		Target:  enum.Slice(awstypes.ApplicationStatusReady),
		Refresh: statusApplication(conn, name),
		Timeout: applicationStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

func waitApplicationUpdated(ctx context.Context, conn *kinesisanalytics.Client, name string) (*awstypes.ApplicationDetail, error) { //nolint:unparam
	const (
		applicationUpdatedTimeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusUpdating),
		Target:  enum.Slice(awstypes.ApplicationStatusReady, awstypes.ApplicationStatusRunning),
		Refresh: statusApplication(conn, name),
		Timeout: applicationUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

func waitIAMPropagation(ctx context.Context, f func() (any, error)) (any, error) {
	var output any

	err := tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		var err error

		output, err = f()

		// Kinesis Stream: https://github.com/hashicorp/terraform-provider-aws/issues/7032
		if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "Kinesis Analytics service doesn't have sufficient privileges") {
			return tfresource.RetryableError(err)
		}

		// Kinesis Firehose: https://github.com/hashicorp/terraform-provider-aws/issues/7394
		if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "Kinesis Analytics doesn't have sufficient privileges") {
			return tfresource.RetryableError(err)
		}

		// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
		if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "does not provide Invoke permissions on the Lambda resource") {
			return tfresource.RetryableError(err)
		}

		// S3: https://github.com/hashicorp/terraform-provider-aws/issues/16104
		if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "Please check the role provided or validity of S3 location you provided") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
