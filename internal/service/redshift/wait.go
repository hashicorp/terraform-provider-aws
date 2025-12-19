// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

func waitClusterCreated(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterAvailabilityStatusModifying, clusterAvailabilityStatusUnavailable},
		Target:     []string{clusterAvailabilityStatusAvailable},
		Refresh:    statusClusterAvailability(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterAvailabilityStatusMaintenance, clusterAvailabilityStatusModifying},
		Target:  []string{},
		Refresh: statusClusterAvailability(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterUpdated(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterAvailabilityStatusMaintenance, clusterAvailabilityStatusModifying, clusterAvailabilityStatusUnavailable},
		Target:  []string{clusterAvailabilityStatusAvailable},
		Refresh: statusClusterAvailability(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterRelocationStatusResolved(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Cluster, error) { //nolint:unparam
	const (
		timeout = 1 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: clusterAvailabilityZoneRelocationStatus_PendingValues(),
		Target:  clusterAvailabilityZoneRelocationStatus_TerminalValues(),
		Refresh: statusClusterAvailabilityZoneRelocation(conn, id),
		Timeout: timeout,
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
		Refresh:    statusCluster(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterAquaApplied(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.AquaStatusApplying),
		Target:     enum.Slice(awstypes.AquaStatusDisabled, awstypes.AquaStatusEnabled),
		Refresh:    statusClusterAqua(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterRestored(ctx context.Context, conn *redshift.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterRestoreStatusStarting, clusterRestoreStatusRestoring},
		Target:     []string{clusterRestoreStatusCompleted},
		Refresh:    statusClusterRestoration(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitEndpointAccessActive(ctx context.Context, conn *redshift.Client, id string) (*awstypes.EndpointAccess, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{endpointAccessStatusCreating, endpointAccessStatusModifying},
		Target:     []string{endpointAccessStatusActive},
		Refresh:    statusEndpointAccess(conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.EndpointAccess); ok {
		return output, err
	}

	return nil, err
}

func waitEndpointAccessDeleted(ctx context.Context, conn *redshift.Client, id string) (*awstypes.EndpointAccess, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{endpointAccessStatusDeleting},
		Target:     []string{},
		Refresh:    statusEndpointAccess(conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.EndpointAccess); ok {
		return output, err
	}

	return nil, err
}

func waitClusterSnapshotCreated(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterSnapshotStatusCreating},
		Target:     []string{clusterSnapshotStatusAvailable},
		Refresh:    statusClusterSnapshot(conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitClusterSnapshotDeleted(ctx context.Context, conn *redshift.Client, id string) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterSnapshotStatusAvailable},
		Target:     []string{},
		Refresh:    statusClusterSnapshot(conn, id),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitIntegrationCreated(ctx context.Context, conn *redshift.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ZeroETLIntegrationStatusCreating, awstypes.ZeroETLIntegrationStatusModifying),
		Target:  enum.Slice(awstypes.ZeroETLIntegrationStatusActive),
		Refresh: statusIntegration(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Integration); ok {
		retry.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.Errors, integrationError)...))

		return output, err
	}

	return nil, err
}

func waitIntegrationUpdated(ctx context.Context, conn *redshift.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ZeroETLIntegrationStatusModifying),
		Target:  enum.Slice(awstypes.ZeroETLIntegrationStatusActive),
		Refresh: statusIntegration(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Integration); ok {
		retry.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.Errors, integrationError)...))

		return output, err
	}

	return nil, err
}

func waitIntegrationDeleted(ctx context.Context, conn *redshift.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ZeroETLIntegrationStatusDeleting, awstypes.ZeroETLIntegrationStatusActive),
		Target:  []string{},
		Refresh: statusIntegration(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Integration); ok {
		retry.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.Errors, integrationError)...))

		return output, err
	}

	return nil, err
}

func waitSnapshotScheduleAssociationCreated(ctx context.Context, conn *redshift.Client, clusterIdentifier, scheduleIdentifier string) (*awstypes.ClusterAssociatedToSchedule, error) {
	const (
		timeout = 75 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ScheduleStateModifying),
		Target:     enum.Slice(awstypes.ScheduleStateActive),
		Refresh:    statusSnapshotScheduleAssociation(conn, clusterIdentifier, scheduleIdentifier),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.ClusterAssociatedToSchedule); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotScheduleAssociationDeleted(ctx context.Context, conn *redshift.Client, clusterIdentifier, scheduleIdentifier string) (*awstypes.ClusterAssociatedToSchedule, error) { //nolint:unparam
	const (
		timeout = 75 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ScheduleStateModifying, awstypes.ScheduleStateActive),
		Target:     []string{},
		Refresh:    statusSnapshotScheduleAssociation(conn, clusterIdentifier, scheduleIdentifier),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.ClusterAssociatedToSchedule); ok {
		return output, err
	}

	return nil, err
}
