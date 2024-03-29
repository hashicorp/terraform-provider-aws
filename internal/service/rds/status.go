// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusDBClusterRole(ctx context.Context, conn *rds.RDS, dbClusterID, roleARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBClusterRoleByDBClusterIDAndRoleARN(ctx, conn, dbClusterID, roleARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusReservedInstance(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindReservedDBInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func statusDBSnapshot(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBSnapshotByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
