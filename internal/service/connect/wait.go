// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	botAssociationCreateTimeout = 5 * time.Minute

	phoneNumberCreatedTimeout = 2 * time.Minute
	phoneNumberUpdatedTimeout = 2 * time.Minute
	phoneNumberDeletedTimeout = 2 * time.Minute

	vocabularyCreatedTimeout = 5 * time.Minute
	// It takes about 90 minutes for Amazon Connect to delete a vocabulary.
	// https://docs.aws.amazon.com/connect/latest/adminguide/add-custom-vocabulary.html
	vocabularyDeletedTimeout = 100 * time.Minute
)

func waitPhoneNumberCreated(ctx context.Context, conn *connect.Connect, timeout time.Duration, phoneNumberId string) (*connect.DescribePhoneNumberOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{connect.PhoneNumberWorkflowStatusInProgress},
		Target:  []string{connect.PhoneNumberWorkflowStatusClaimed},
		Refresh: statusPhoneNumber(ctx, conn, phoneNumberId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*connect.DescribePhoneNumberOutput); ok {
		if aws.StringValue(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Status) == connect.PhoneNumberWorkflowStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitPhoneNumberUpdated(ctx context.Context, conn *connect.Connect, timeout time.Duration, phoneNumberId string) (*connect.DescribePhoneNumberOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{connect.PhoneNumberWorkflowStatusInProgress},
		Target:  []string{connect.PhoneNumberWorkflowStatusClaimed},
		Refresh: statusPhoneNumber(ctx, conn, phoneNumberId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*connect.DescribePhoneNumberOutput); ok {
		if aws.StringValue(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Status) == connect.PhoneNumberWorkflowStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitPhoneNumberDeleted(ctx context.Context, conn *connect.Connect, timeout time.Duration, phoneNumberId string) (*connect.DescribePhoneNumberOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{connect.PhoneNumberWorkflowStatusInProgress},
		Target:  []string{connect.ErrCodeResourceNotFoundException},
		Refresh: statusPhoneNumber(ctx, conn, phoneNumberId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribePhoneNumberOutput); ok {
		return v, err
	}

	return nil, err
}

func waitVocabularyCreated(ctx context.Context, conn *connect.Connect, timeout time.Duration, instanceId, vocabularyId string) (*connect.DescribeVocabularyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{connect.VocabularyStateCreationInProgress},
		Target:  []string{connect.VocabularyStateActive, connect.VocabularyStateCreationFailed},
		Refresh: statusVocabulary(ctx, conn, instanceId, vocabularyId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribeVocabularyOutput); ok {
		return v, err
	}

	return nil, err
}

func waitVocabularyDeleted(ctx context.Context, conn *connect.Connect, timeout time.Duration, instanceId, vocabularyId string) (*connect.DescribeVocabularyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{connect.VocabularyStateDeleteInProgress},
		Target:  []string{connect.ErrCodeResourceNotFoundException},
		Refresh: statusVocabulary(ctx, conn, instanceId, vocabularyId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribeVocabularyOutput); ok {
		return v, err
	}

	return nil, err
}
