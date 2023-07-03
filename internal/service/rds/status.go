// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// ProxyEndpoint NotFound
	proxyEndpointStatusNotFound = "NotFound"

	// ProxyEndpoint Unknown
	proxyEndpointStatusUnknown = "Unknown"
)

func statusEventSubscription(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEventSubscriptionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// statusDBProxyEndpoint fetches the ProxyEndpoint and its Status
func statusDBProxyEndpoint(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBProxyEndpoint(ctx, conn, id)
		if err != nil {
			return nil, proxyEndpointStatusUnknown, err
		}

		if output == nil {
			return nil, proxyEndpointStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

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

func statusDBInstanceAutomatedBackup(ctx context.Context, conn *rds.RDS, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBInstanceAutomatedBackupByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// statusDBInstanceHasAutomatedBackup returns whether or not a database instance has a specified automated backup.
// The connection must be valid for the database instance's Region.
func statusDBInstanceHasAutomatedBackup(ctx context.Context, conn *rds.RDS, dbInstanceID, dbInstanceAutomatedBackupsARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBInstanceByIDSDKv1(ctx, conn, dbInstanceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, v := range output.DBInstanceAutomatedBackupsReplications {
			if aws.StringValue(v.DBInstanceAutomatedBackupsArn) == dbInstanceAutomatedBackupsARN {
				return output, strconv.FormatBool(true), nil
			}
		}

		return output, strconv.FormatBool(false), nil
	}
}

func statusDBProxy(ctx context.Context, conn *rds.RDS, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBProxyByName(ctx, conn, name)

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
