// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
)

// Custom MQ service lister functions using the same format as generated code.

func describeBrokerInstanceOptionsPages(ctx context.Context, conn *mq.Client, input *mq.DescribeBrokerInstanceOptionsInput, fn func(*mq.DescribeBrokerInstanceOptionsOutput, bool) bool, optFns ...func(*mq.Options)) error {
	for {
		output, err := conn.DescribeBrokerInstanceOptions(ctx, input, optFns...)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
