//go:build !generate
// +build !generate

package firehose

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
)

// Custom Kinesis Firehose service lister functions using the same format as generated code.

func listDeliveryStreamsPages(conn *firehose.Firehose, input *firehose.ListDeliveryStreamsInput, fn func(*firehose.ListDeliveryStreamsOutput, bool) bool) error { //nolint:deadcode // This function is called from a sweeper.
	return listDeliveryStreamsPagesWithContext(context.Background(), conn, input, fn)
}

func listDeliveryStreamsPagesWithContext(ctx context.Context, conn *firehose.Firehose, input *firehose.ListDeliveryStreamsInput, fn func(*firehose.ListDeliveryStreamsOutput, bool) bool) error {
	for {
		output, err := conn.ListDeliveryStreamsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := !aws.BoolValue(output.HasMoreDeliveryStreams)
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.ExclusiveStartDeliveryStreamName = output.DeliveryStreamNames[len(output.DeliveryStreamNames)-1]
	}
	return nil
}
