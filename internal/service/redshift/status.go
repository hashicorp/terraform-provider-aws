// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusClusterAvailability(ctx context.Context, conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.ClusterAvailabilityStatus), nil
	}
}

func statusClusterAvailabilityZoneRelocation(ctx context.Context, conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.AvailabilityZoneRelocationStatus), nil
	}
}

func statusCluster(ctx context.Context, conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.ClusterStatus), nil
	}
}

func statusClusterAqua(ctx context.Context, conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AquaConfiguration.AquaStatus), nil
	}
}

func statusEndpointAccess(ctx context.Context, conn *redshift.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findEndpointAccessByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.EndpointStatus), nil
	}
}

func statusClusterSnapshot(ctx context.Context, conn *redshift.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findClusterSnapshotByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}
