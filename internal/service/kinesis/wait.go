package kinesis

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	streamConsumerCreatedTimeout = 5 * time.Minute
	streamConsumerDeletedTimeout = 5 * time.Minute
)

// waitStreamConsumerCreated waits for an Stream Consumer to return Active
func waitStreamConsumerCreated(conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesis.ConsumerStatusCreating},
		Target:  []string{kinesis.ConsumerStatusActive},
		Refresh: statusStreamConsumer(conn, arn),
		Timeout: streamConsumerCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesis.ConsumerDescription); ok {
		return v, err
	}

	return nil, err
}

// waitStreamConsumerDeleted waits for a Stream Consumer to be deleted
func waitStreamConsumerDeleted(conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesis.ConsumerStatusDeleting},
		Target:  []string{},
		Refresh: statusStreamConsumer(conn, arn),
		Timeout: streamConsumerDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesis.ConsumerDescription); ok {
		return v, err
	}

	return nil, err
}
