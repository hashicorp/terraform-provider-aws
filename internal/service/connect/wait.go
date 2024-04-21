// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

func waitPhoneNumberCreated(ctx context.Context, conn *connect.Client, timeout time.Duration, phoneNumberId string) (*connect.DescribePhoneNumberOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PhoneNumberWorkflowStatusInProgress),
		Target:  enum.Slice(awstypes.PhoneNumberWorkflowStatusClaimed),
		Refresh: statusPhoneNumber(ctx, conn, phoneNumberId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*connect.DescribePhoneNumberOutput); ok {
		if output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Status == awstypes.PhoneNumberWorkflowStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitPhoneNumberUpdated(ctx context.Context, conn *connect.Client, timeout time.Duration, phoneNumberId string) (*connect.DescribePhoneNumberOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PhoneNumberWorkflowStatusInProgress),
		Target:  enum.Slice(awstypes.PhoneNumberWorkflowStatusClaimed),
		Refresh: statusPhoneNumber(ctx, conn, phoneNumberId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*connect.DescribePhoneNumberOutput); ok {
		if output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Status == awstypes.PhoneNumberWorkflowStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitPhoneNumberDeleted(ctx context.Context, conn *connect.Client, timeout time.Duration, phoneNumberId string) (*connect.DescribePhoneNumberOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PhoneNumberWorkflowStatusInProgress),
		Target:  []string{},
		Refresh: statusPhoneNumber(ctx, conn, phoneNumberId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribePhoneNumberOutput); ok {
		return v, err
	}

	return nil, err
}

func waitVocabularyCreated(ctx context.Context, conn *connect.Client, timeout time.Duration, instanceId, vocabularyId string) (*connect.DescribeVocabularyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VocabularyStateCreationInProgress),
		Target:  []string{},
		Refresh: statusVocabulary(ctx, conn, instanceId, vocabularyId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribeVocabularyOutput); ok {
		return v, err
	}

	return nil, err
}

func waitVocabularyDeleted(ctx context.Context, conn *connect.Client, timeout time.Duration, instanceId, vocabularyId string) (*connect.DescribeVocabularyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VocabularyStateDeleteInProgress),
		Target:  []string{},
		Refresh: statusVocabulary(ctx, conn, instanceId, vocabularyId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*connect.DescribeVocabularyOutput); ok {
		return v, err
	}

	return nil, err
}
