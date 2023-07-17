// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package firehose

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitDeliveryStreamCreated(ctx context.Context, conn *firehose.Firehose, name string, timeout time.Duration) (*firehose.DeliveryStreamDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamStatusCreating},
		Target:  []string{firehose.DeliveryStreamStatusActive},
		Refresh: statusDeliveryStream(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*firehose.DeliveryStreamDescription); ok {
		if status, failureDescription := aws.StringValue(output.DeliveryStreamStatus), output.FailureDescription; status == firehose.DeliveryStreamStatusCreatingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(failureDescription.Type), aws.StringValue(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func waitDeliveryStreamDeleted(ctx context.Context, conn *firehose.Firehose, name string, timeout time.Duration) (*firehose.DeliveryStreamDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamStatusDeleting},
		Target:  []string{},
		Refresh: statusDeliveryStream(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*firehose.DeliveryStreamDescription); ok {
		if status, failureDescription := aws.StringValue(output.DeliveryStreamStatus), output.FailureDescription; status == firehose.DeliveryStreamStatusDeletingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(failureDescription.Type), aws.StringValue(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func waitDeliveryStreamEncryptionEnabled(ctx context.Context, conn *firehose.Firehose, name string, timeout time.Duration) (*firehose.DeliveryStreamEncryptionConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamEncryptionStatusEnabling},
		Target:  []string{firehose.DeliveryStreamEncryptionStatusEnabled},
		Refresh: statusDeliveryStreamEncryptionConfiguration(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*firehose.DeliveryStreamEncryptionConfiguration); ok {
		if status, failureDescription := aws.StringValue(output.Status), output.FailureDescription; status == firehose.DeliveryStreamEncryptionStatusEnablingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(failureDescription.Type), aws.StringValue(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func waitDeliveryStreamEncryptionDisabled(ctx context.Context, conn *firehose.Firehose, name string, timeout time.Duration) (*firehose.DeliveryStreamEncryptionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamEncryptionStatusDisabling},
		Target:  []string{firehose.DeliveryStreamEncryptionStatusDisabled},
		Refresh: statusDeliveryStreamEncryptionConfiguration(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*firehose.DeliveryStreamEncryptionConfiguration); ok {
		if status, failureDescription := aws.StringValue(output.Status), output.FailureDescription; status == firehose.DeliveryStreamEncryptionStatusDisablingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(failureDescription.Type), aws.StringValue(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}
