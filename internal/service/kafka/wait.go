package kafka

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	clusterCreateTimeout = 120 * time.Minute
	clusterUpdateTimeout = 120 * time.Minute
	clusterDeleteTimeout = 120 * time.Minute
)

const (
	// Maximum amount of time to wait for an Configuration to return Deleted
	configurationDeletedTimeout = 5 * time.Minute
)

// waitConfigurationDeleted waits for an Configuration to return Deleted
func waitConfigurationDeleted(conn *kafka.Kafka, arn string) (*kafka.DescribeConfigurationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ConfigurationStateDeleting},
		Target:  []string{configurationStateDeleted},
		Refresh: statusConfigurationState(conn, arn),
		Timeout: configurationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafka.DescribeConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}
