// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	vocabularyCreatedTimeout = 5 * time.Minute
	// It takes about 90 minutes for Amazon Connect to delete a vocabulary.
	// https://docs.aws.amazon.com/connect/latest/adminguide/add-custom-vocabulary.html
	vocabularyDeletedTimeout = 100 * time.Minute
)

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
