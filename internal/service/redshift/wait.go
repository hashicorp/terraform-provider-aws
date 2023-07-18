// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	clusterInvalidClusterStateFaultTimeout = 15 * time.Minute

	clusterRelocationStatusResolvedTimeout = 1 * time.Minute

	snapshotScheduleAssociationActivatedTimeout = 75 * time.Minute
	snapshotScheduleAssociationDestroyedTimeout = 75 * time.Minute
)

func waitClusterCreated(ctx context.Context, conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterAvailabilityStatusModifying, clusterAvailabilityStatusUnavailable},
		Target:     []string{clusterAvailabilityStatusAvailable},
		Refresh:    statusClusterAvailability(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterAvailabilityStatusMaintenance, clusterAvailabilityStatusModifying},
		Target:  []string{},
		Refresh: statusClusterAvailability(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterUpdated(ctx context.Context, conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterAvailabilityStatusMaintenance, clusterAvailabilityStatusModifying, clusterAvailabilityStatusUnavailable},
		Target:  []string{clusterAvailabilityStatusAvailable},
		Refresh: statusClusterAvailability(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterRelocationStatusResolved(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.Cluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: clusterAvailabilityZoneRelocationStatus_PendingValues(),
		Target:  clusterAvailabilityZoneRelocationStatus_TerminalValues(),
		Refresh: statusClusterAvailabilityZoneRelocation(ctx, conn, id),
		Timeout: clusterRelocationStatusResolvedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterRebooted(ctx context.Context, conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterStatusRebooting, clusterStatusModifying},
		Target:     []string{clusterStatusAvailable},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterAquaApplied(ctx context.Context, conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{redshift.AquaStatusApplying},
		Target:     []string{redshift.AquaStatusDisabled, redshift.AquaStatusEnabled},
		Refresh:    statusClusterAqua(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func WaitScheduleAssociationActive(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.ClusterAssociatedToSchedule, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{redshift.ScheduleStateModifying},
		Target:     []string{redshift.ScheduleStateActive},
		Refresh:    statusScheduleAssociation(ctx, conn, id),
		Timeout:    snapshotScheduleAssociationActivatedTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.ClusterAssociatedToSchedule); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ScheduleAssociationState)))

		return output, err
	}

	return nil, err
}

func waitScheduleAssociationDeleted(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.ClusterAssociatedToSchedule, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{redshift.ScheduleStateModifying, redshift.ScheduleStateActive},
		Target:     []string{},
		Refresh:    statusScheduleAssociation(ctx, conn, id),
		Timeout:    snapshotScheduleAssociationDestroyedTimeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.ClusterAssociatedToSchedule); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ScheduleAssociationState)))
		return output, err
	}

	return nil, err
}

func waitEndpointAccessActive(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.EndpointAccess, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{endpointAccessStatusCreating, endpointAccessStatusModifying},
		Target:     []string{endpointAccessStatusActive},
		Refresh:    statusEndpointAccess(ctx, conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.EndpointAccess); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.EndpointStatus)))

		return output, err
	}

	return nil, err
}

func waitEndpointAccessDeleted(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.EndpointAccess, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{endpointAccessStatusDeleting},
		Target:     []string{},
		Refresh:    statusEndpointAccess(ctx, conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.EndpointAccess); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.EndpointStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterSnapshotCreated(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterSnapshotStatusCreating},
		Target:     []string{clusterSnapshotStatusAvailable},
		Refresh:    statusClusterSnapshot(ctx, conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.Snapshot); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status)))

		return output, err
	}

	return nil, err
}

func waitClusterSnapshotDeleted(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterSnapshotStatusAvailable},
		Target:     []string{},
		Refresh:    statusClusterSnapshot(ctx, conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.Snapshot); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status)))

		return output, err
	}

	return nil, err
}
