// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func statusPhoneNumber(ctx context.Context, conn *connect.Client, phoneNumberId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &connect.DescribePhoneNumberInput{
			PhoneNumberId: aws.String(phoneNumberId),
		}

		output, err := conn.DescribePhoneNumber(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return output, err.Error(), nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Status), nil
	}
}

func statusVocabulary(ctx context.Context, conn *connect.Client, instanceId, vocabularyId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &connect.DescribeVocabularyInput{
			InstanceId:   aws.String(instanceId),
			VocabularyId: aws.String(vocabularyId),
		}

		output, err := conn.DescribeVocabulary(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return output, err.Error(), nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Vocabulary.State), nil
	}
}
