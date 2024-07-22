// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDBClusterRoleByDBClusterIDAndRoleARN(ctx context.Context, conn *rds.RDS, dbClusterID, roleARN string) (*rds.DBClusterRole, error) {
	dbCluster, err := FindDBClusterByID(ctx, conn, dbClusterID)
	if err != nil {
		return nil, err
	}

	for _, associatedRole := range dbCluster.AssociatedRoles {
		if aws.StringValue(associatedRole.RoleArn) == roleARN {
			if status := aws.StringValue(associatedRole.Status); status == clusterRoleStatusDeleted {
				return nil, &retry.NotFoundError{
					Message: status,
				}
			}

			return associatedRole, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindReservedDBInstanceByID(ctx context.Context, conn *rds.RDS, id string) (*rds.ReservedDBInstance, error) {
	input := &rds.DescribeReservedDBInstancesInput{
		ReservedDBInstanceId: aws.String(id),
	}

	output, err := conn.DescribeReservedDBInstancesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeReservedDBInstanceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ReservedDBInstances) == 0 || output.ReservedDBInstances[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ReservedDBInstances); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ReservedDBInstances[0], nil
}
