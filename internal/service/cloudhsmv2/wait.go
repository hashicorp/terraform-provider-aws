package cloudhsmv2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitClusterActive(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cloudhsmv2.ClusterStateCreateInProgress,
			cloudhsmv2.ClusterStateInitializeInProgress,
		},
		Target:     []string{cloudhsmv2.ClusterStateActive},
		Refresh:    statusClusterState(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*cloudhsmv2.Cluster); ok {
		return v, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.ClusterStateDeleteInProgress},
		Target:     []string{cloudhsmv2.ClusterStateDeleted},
		Refresh:    statusClusterState(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*cloudhsmv2.Cluster); ok {
		return v, err
	}

	return nil, err
}

func waitClusterUninitialized(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cloudhsmv2.ClusterStateCreateInProgress,
			cloudhsmv2.ClusterStateInitializeInProgress,
		},
		Target:     []string{cloudhsmv2.ClusterStateUninitialized},
		Refresh:    statusClusterState(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*cloudhsmv2.Cluster); ok {
		return v, err
	}

	return nil, err
}

func waitHSMActive(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Hsm, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.HsmStateCreateInProgress},
		Target:     []string{cloudhsmv2.HsmStateActive},
		Refresh:    statusHSMState(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*cloudhsmv2.Hsm); ok {
		return v, err
	}

	return nil, err
}

func waitHSMDeleted(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Hsm, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.HsmStateDeleteInProgress},
		Target:     []string{},
		Refresh:    statusHSMState(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*cloudhsmv2.Hsm); ok {
		return v, err
	}

	return nil, err
}
