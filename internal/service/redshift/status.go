// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusClusterAvailability(conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.ClusterAvailabilityStatus), nil
	}
}

func statusClusterAvailabilityZoneRelocation(conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.AvailabilityZoneRelocationStatus), nil
	}
}

func statusCluster(conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.ClusterStatus), nil
	}
}

func statusClusterAqua(conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AquaConfiguration.AquaStatus), nil
	}
}

func statusClusterRestoration(conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.RestoreStatus == nil {
			return nil, "", nil
		}

		return output, aws.ToString(output.RestoreStatus.Status), nil
	}
}

func statusEndpointAccess(conn *redshift.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findEndpointAccessByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.EndpointStatus), nil
	}
}

func statusClusterSnapshot(conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClusterSnapshotByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func statusIntegration(conn *redshift.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIntegrationByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusSnapshotScheduleAssociation(conn *redshift.Client, clusterIdentifier, scheduleIdentifier string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSnapshotScheduleAssociationByTwoPartKey(ctx, conn, clusterIdentifier, scheduleIdentifier)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ScheduleAssociationState), nil
	}
}
