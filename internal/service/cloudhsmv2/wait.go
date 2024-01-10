// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitClusterActive(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{cloudhsmv2.ClusterStateCreateInProgress, cloudhsmv2.ClusterStateInitializeInProgress},
		Target:     []string{cloudhsmv2.ClusterStateActive},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudhsmv2.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{cloudhsmv2.ClusterStateDeleteInProgress},
		Target:     []string{},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudhsmv2.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func waitClusterUninitialized(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{cloudhsmv2.ClusterStateCreateInProgress, cloudhsmv2.ClusterStateInitializeInProgress},
		Target:     []string{cloudhsmv2.ClusterStateUninitialized},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudhsmv2.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func waitHSMCreated(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Hsm, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{cloudhsmv2.HsmStateCreateInProgress},
		Target:     []string{cloudhsmv2.HsmStateActive},
		Refresh:    statusHSM(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudhsmv2.Hsm); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func waitHSMDeleted(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Hsm, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{cloudhsmv2.HsmStateDeleteInProgress},
		Target:     []string{},
		Refresh:    statusHSM(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudhsmv2.Hsm); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))

		return output, err
	}

	return nil, err
}
