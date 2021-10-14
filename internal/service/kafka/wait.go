package kafka

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	clusterCreateTimeout = 120 * time.Minute
	clusterUpdateTimeout = 120 * time.Minute
	clusterDeleteTimeout = 120 * time.Minute
)

const (
	configurationDeletedTimeout = 5 * time.Minute
)

func waitClusterCreated(conn *kafka.Kafka, arn string, timeout time.Duration) (*kafka.ClusterInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ClusterStateCreating},
		Target:  []string{kafka.ClusterStateActive},
		Refresh: statusClusterState(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafka.ClusterInfo); ok {
		if state, stateInfo := aws.StringValue(output.State), output.StateInfo; state == kafka.ClusterStateFailed && stateInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateInfo.Code), aws.StringValue(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(conn *kafka.Kafka, arn string, timeout time.Duration) (*kafka.ClusterInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ClusterStateDeleting},
		Target:  []string{},
		Refresh: statusClusterState(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafka.ClusterInfo); ok {
		if state, stateInfo := aws.StringValue(output.State), output.StateInfo; state == kafka.ClusterStateFailed && stateInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateInfo.Code), aws.StringValue(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitClusterOperationCompleted(conn *kafka.Kafka, arn string, timeout time.Duration) (*kafka.ClusterOperationInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ClusterOperationStatePending, ClusterOperationStateUpdateInProgress},
		Target:  []string{ClusterOperationStateUpdateComplete},
		Refresh: statusClusterOperationState(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafka.ClusterOperationInfo); ok {
		if state, errorInfo := aws.StringValue(output.OperationState), output.ErrorInfo; state == ClusterOperationStateUpdateFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(errorInfo.ErrorCode), aws.StringValue(errorInfo.ErrorString)))
		}

		return output, err
	}

	return nil, err
}

func waitConfigurationDeleted(conn *kafka.Kafka, arn string) (*kafka.DescribeConfigurationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ConfigurationStateDeleting},
		Target:  []string{},
		Refresh: statusConfigurationState(conn, arn),
		Timeout: configurationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafka.DescribeConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}
