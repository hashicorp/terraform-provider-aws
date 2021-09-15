package waiter

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfkafka "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kafka"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	// TODO Remove?
	// ClusterCreatedTimeout = 120 * time.Minute
	// ClusterUpdatedTimeout = 120 * time.Minute
	// ClusterDeletedTimeout = 120 * time.Minute

	ConfigurationDeletedTimeout = 5 * time.Minute
)

func ClusterCreated(conn *kafka.Kafka, arn string, timeout time.Duration) (*kafka.ClusterInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ClusterStateCreating},
		Target:  []string{kafka.ClusterStateActive},
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

func ClusterOperationCompleted(conn *kafka.Kafka, arn string, timeout time.Duration) (*kafka.ClusterOperationInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{tfkafka.ClusterOperationStatePending, tfkafka.ClusterOperationStateUpdateInProgress},
		Target:  []string{tfkafka.ClusterOperationStateUpdateComplete},
		Refresh: ClusterOperationState(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafka.ClusterOperationInfo); ok {
		if state, errorInfo := aws.StringValue(output.OperationState), output.ErrorInfo; state == tfkafka.ClusterOperationStateUpdateFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(errorInfo.ErrorCode), aws.StringValue(errorInfo.ErrorString)))
		}

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
