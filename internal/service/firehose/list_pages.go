// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package firehose

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
)

// Custom Kinesis Firehose service lister functions using the same format as generated code.

func listDeliveryStreamsPages(ctx context.Context, conn *firehose.Client, input *firehose.ListDeliveryStreamsInput, fn func(*firehose.ListDeliveryStreamsOutput, bool) bool) error {
	for {
		output, err := conn.ListDeliveryStreams(ctx, input)
		if err != nil {
			return err
		}

		lastPage := !aws.ToBool(output.HasMoreDeliveryStreams)
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.ExclusiveStartDeliveryStreamName = aws.String(output.DeliveryStreamNames[len(output.DeliveryStreamNames)-1])
	}
	return nil
}
