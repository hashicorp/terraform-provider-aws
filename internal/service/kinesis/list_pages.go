// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
)

// Custom Kinesis service lister functions using the same format as generated code.

func listShardsPages(ctx context.Context, conn *kinesis.Client, input *kinesis.ListShardsInput, fn func(*kinesis.ListShardsOutput, bool) bool) error {
	for {
		output, err := conn.ListShards(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
		// "Don't specify StreamName or StreamCreationTimestamp if you specify NextToken because the latter unambiguously identifies the stream".
		input.StreamCreationTimestamp = nil
		input.StreamName = nil
	}
	return nil
}
