package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// TODO Remove?
	// ClusterCreatedTimeout = 120 * time.Minute
	// ClusterUpdatedTimeout = 120 * time.Minute
	// ClusterDeletedTimeout = 120 * time.Minute

	ConfigurationDeletedTimeout = 5 * time.Minute
)

func ClusterDeleted(conn *kafka.Kafka, arn string, timeout time.Duration) (*kafka.ClusterInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ClusterStateDeleting},
		Target:  []string{},
		Refresh: ClusterState(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafka.ClusterInfo); ok {
		// Ideally we would look at output.StateInfo if the state was FAILED
		//  https://docs.aws.amazon.com/msk/1.0/apireference/clusters.html#clusters-model-stateinfo
		// but for some reason it's not exposed in the SDK model.
		return output, err
	}

	return nil, err
}

func ConfigurationDeleted(conn *kafka.Kafka, arn string) (*kafka.DescribeConfigurationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ConfigurationStateDeleting},
		Target:  []string{},
		Refresh: ConfigurationState(conn, arn),
		Timeout: ConfigurationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafka.DescribeConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}
