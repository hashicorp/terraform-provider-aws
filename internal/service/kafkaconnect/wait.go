package kafkaconnect

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitCustomPluginCreated(conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeCustomPluginOutput, error) {
	stateconf := &resource.StateChangeConf{
		Pending: []string{kafkaconnect.CustomPluginStateCreating},
		Target:  []string{kafkaconnect.CustomPluginStateActive},
		Refresh: statusCustomPluginState(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateconf.WaitForState()

	if output, ok := outputRaw.(*kafkaconnect.DescribeCustomPluginOutput); ok {
		return output, err
	}

	return nil, err
}

func waitConnectorCreated(conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateconf := &resource.StateChangeConf{
		Pending: []string{kafkaconnect.ConnectorStateCreating},
		Target:  []string{kafkaconnect.ConnectorStateRunning},
		Refresh: statusConnectorState(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateconf.WaitForState()

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		return output, err
	}

	return nil, err
}

func waitConnectorDeleted(conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateconf := &resource.StateChangeConf{
		Pending: []string{kafkaconnect.ConnectorStateDeleting},
		Target:  []string{},
		Refresh: statusConnectorState(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateconf.WaitForState()

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		return output, err
	}

	return nil, err
}
