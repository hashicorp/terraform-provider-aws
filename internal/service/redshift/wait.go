// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	clusterInvalidClusterStateFaultTimeout = 15 * time.Minute

	clusterRelocationStatusResolvedTimeout = 1 * time.Minute
)

func waitClusterCreated(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterAvailabilityStatusModifying, clusterAvailabilityStatusUnavailable},
		Target:     []string{clusterAvailabilityStatusAvailable},
		Refresh:    statusClusterAvailability(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterAvailabilityStatusMaintenance, clusterAvailabilityStatusModifying},
		Target:  []string{},
		Refresh: statusClusterAvailability(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterUpdated(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterAvailabilityStatusMaintenance, clusterAvailabilityStatusModifying, clusterAvailabilityStatusUnavailable},
		Target:  []string{clusterAvailabilityStatusAvailable},
		Refresh: statusClusterAvailability(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterRelocationStatusResolved(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Cluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: clusterAvailabilityZoneRelocationStatus_PendingValues(),
		Target:  clusterAvailabilityZoneRelocationStatus_TerminalValues(),
		Refresh: statusClusterAvailabilityZoneRelocation(ctx, conn, id),
		Timeout: clusterRelocationStatusResolvedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterRebooted(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterStatusRebooting, clusterStatusModifying},
		Target:     []string{clusterStatusAvailable},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterAquaApplied(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.AquaStatusApplying),
		Target:     enum.Slice(awstypes.AquaStatusDisabled, awstypes.AquaStatusEnabled),
		Refresh:    statusClusterAqua(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitEndpointAccessActive(ctx context.Context, conn *redshift.Client, id string) (*awstypes.EndpointAccess, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{endpointAccessStatusCreating, endpointAccessStatusModifying},
		Target:     []string{endpointAccessStatusActive},
		Refresh:    statusEndpointAccess(ctx, conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.EndpointAccess); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.EndpointStatus)))

		return output, err
	}

	return nil, err
}

func waitEndpointAccessDeleted(ctx context.Context, conn *redshift.Client, id string) (*awstypes.EndpointAccess, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{endpointAccessStatusDeleting},
		Target:     []string{},
		Refresh:    statusEndpointAccess(ctx, conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.EndpointAccess); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.EndpointStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterSnapshotCreated(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterSnapshotStatusCreating},
		Target:     []string{clusterSnapshotStatusAvailable},
		Refresh:    statusClusterSnapshot(ctx, conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status)))

		return output, err
	}

	return nil, err
}

func waitClusterSnapshotDeleted(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterSnapshotStatusAvailable},
		Target:     []string{},
		Refresh:    statusClusterSnapshot(ctx, conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status)))

		return output, err
	}

	return nil, err
}
