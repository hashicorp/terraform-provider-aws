package redshift

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	clusterInvalidClusterStateFaultTimeout = 15 * time.Minute

	clusterRelocationStatusResolvedTimeout = 1 * time.Minute
)

func waitClusterCreated(conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{clusterAvailabilityStatusModifying, clusterAvailabilityStatusUnavailable},
		Target:     []string{clusterAvailabilityStatusAvailable},
		Refresh:    statusClusterAvailability(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{clusterAvailabilityStatusModifying},
		Target:  []string{},
		Refresh: statusClusterAvailability(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterUpdated(conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{clusterAvailabilityStatusMaintenance, clusterAvailabilityStatusModifying, clusterAvailabilityStatusUnavailable},
		Target:  []string{clusterAvailabilityStatusAvailable},
		Refresh: statusClusterAvailability(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterRelocationStatusResolved(conn *redshift.Redshift, id string) (*redshift.Cluster, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: clusterAvailabilityZoneRelocationStatus_PendingValues(),
		Target:  clusterAvailabilityZoneRelocationStatus_TerminalValues(),
		Refresh: statusClusterAvailabilityZoneRelocation(conn, id),
		Timeout: clusterRelocationStatusResolvedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterRebooted(conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{clusterStatusRebooting, clusterStatusModifying},
		Target:     []string{clusterStatusAvailable},
		Refresh:    statusCluster(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}

func waitClusterAquaApplied(conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{redshift.AquaStatusApplying},
		Target:     []string{redshift.AquaStatusDisabled, redshift.AquaStatusEnabled},
		Refresh:    statusClusterAqua(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}

	return nil, err
}
