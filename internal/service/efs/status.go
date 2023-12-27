// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusAccessPointLifeCycleState fetches the Access Point and its LifecycleState
func statusAccessPointLifeCycleState(ctx context.Context, conn *efs.EFS, accessPointId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &efs.DescribeAccessPointsInput{
			AccessPointId: aws.String(accessPointId),
		}

		output, err := conn.DescribeAccessPointsWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.AccessPoints) == 0 || output.AccessPoints[0] == nil {
			return nil, "", nil
		}

		mt := output.AccessPoints[0]

		return mt, aws.StringValue(mt.LifeCycleState), nil
	}
}

func statusBackupPolicy(ctx context.Context, conn *efs.EFS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindBackupPolicyByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusReplicationConfiguration(ctx context.Context, conn *efs.EFS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindReplicationConfigurationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Destinations[0].Status), nil
	}
}
