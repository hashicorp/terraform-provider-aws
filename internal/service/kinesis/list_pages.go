// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
)

// Custom Kinesis service lister functions using the same format as generated code.

func listShardsPages(ctx context.Context, conn *kinesis.Client, input *kinesis.ListShardsInput, fn func(*kinesis.ListShardsOutput, bool) bool, optFns ...func(*kinesis.Options)) error {
	for {
		output, err := conn.ListShards(ctx, input, optFns...)
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

func listTagsForStreamPages(ctx context.Context, conn *kinesis.Client, input *kinesis.ListTagsForStreamInput, fn func(*kinesis.ListTagsForStreamOutput, bool) bool, optFns ...func(*kinesis.Options)) error {
	for {
		output, err := conn.ListTagsForStream(ctx, input, optFns...)
		if err != nil {
			return err
		}

		lastPage := !aws.ToBool(output.HasMoreTags)
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.ExclusiveStartTagKey = output.Tags[len(output.Tags)-1].Key
	}
	return nil
}
