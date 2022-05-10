package rds

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindDBProxyTarget returns matching FindDBProxyTarget.
func FindDBProxyTarget(conn *rds.RDS, dbProxyName, targetGroupName, targetType, rdsResourceId string) (*rds.DBProxyTarget, error) {
	input := &rds.DescribeDBProxyTargetsInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String(targetGroupName),
	}
	var dbProxyTarget *rds.DBProxyTarget

	err := conn.DescribeDBProxyTargetsPages(input, func(page *rds.DescribeDBProxyTargetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, target := range page.Targets {
			if aws.StringValue(target.Type) == targetType && aws.StringValue(target.RdsResourceId) == rdsResourceId {
				dbProxyTarget = target
				return false
			}
		}

		return !lastPage
	})

	return dbProxyTarget, err
}

// FindDBProxyEndpoint returns matching FindDBProxyEndpoint.
func FindDBProxyEndpoint(conn *rds.RDS, id string) (*rds.DBProxyEndpoint, error) {
	dbProxyName, dbProxyEndpointName, err := ProxyEndpointParseID(id)
	if err != nil {
		return nil, err
	}

	input := &rds.DescribeDBProxyEndpointsInput{
		DBProxyName:         aws.String(dbProxyName),
		DBProxyEndpointName: aws.String(dbProxyEndpointName),
	}
	var dbProxyEndpoint *rds.DBProxyEndpoint

	err = conn.DescribeDBProxyEndpointsPages(input, func(page *rds.DescribeDBProxyEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, endpoint := range page.DBProxyEndpoints {
			if aws.StringValue(endpoint.DBProxyEndpointName) == dbProxyEndpointName &&
				aws.StringValue(endpoint.DBProxyName) == dbProxyName {
				dbProxyEndpoint = endpoint
				return false
			}
		}

		return !lastPage
	})

	return dbProxyEndpoint, err
}

func FindDBClusterRoleByDBClusterIDAndRoleARN(conn *rds.RDS, dbClusterID, roleARN string) (*rds.DBClusterRole, error) {
	dbCluster, err := FindDBClusterByID(conn, dbClusterID)

	if err != nil {
		return nil, err
	}

	for _, associatedRole := range dbCluster.AssociatedRoles {
		if aws.StringValue(associatedRole.RoleArn) == roleARN {
			if status := aws.StringValue(associatedRole.Status); status == ClusterRoleStatusDeleted {
				return nil, &resource.NotFoundError{
					Message: status,
				}
			}

			return associatedRole, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindDBClusterByID(conn *rds.RDS, id string) (*rds.DBCluster, error) {
	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(id),
	}

	output, err := conn.DescribeDBClusters(input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || len(output.DBClusters) == 0 || output.DBClusters[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	dbCluster := output.DBClusters[0]

	// Eventual consistency check.
	if aws.StringValue(dbCluster.DBClusterIdentifier) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return dbCluster, nil
}

func FindDBClusterWithActivityStream(conn *rds.RDS, dbClusterArn string) (*rds.DBCluster, error) {
	log.Printf("[DEBUG] Calling conn.DescribeDBCClusters(input) with DBClusterIdentifier set to %s", dbClusterArn)
	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(dbClusterArn),
	}

	output, err := conn.DescribeDBClusters(input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || len(output.DBClusters) == 0 || output.DBClusters[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	dbCluster := output.DBClusters[0]

	// Eventual consistency check.
	if aws.StringValue(dbCluster.DBClusterArn) != dbClusterArn {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	if status := aws.StringValue(dbCluster.ActivityStreamStatus); status == rds.ActivityStreamStatusStopped {
		return nil, &resource.NotFoundError{
			Message: status,
		}
	}

	return dbCluster, nil
}

func FindDBInstanceByID(conn *rds.RDS, id string) (*rds.DBInstance, error) {
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(id),
	}

	output, err := conn.DescribeDBInstances(input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || len(output.DBInstances) == 0 || output.DBInstances[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	dbInstance := output.DBInstances[0]

	// Eventual consistency check.
	if aws.StringValue(dbInstance.DBInstanceIdentifier) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return dbInstance, nil
}

func FindDBProxyByName(conn *rds.RDS, name string) (*rds.DBProxy, error) {
	input := &rds.DescribeDBProxiesInput{
		DBProxyName: aws.String(name),
	}

	output, err := conn.DescribeDBProxies(input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || len(output.DBProxies) == 0 || output.DBProxies[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	dbProxy := output.DBProxies[0]

	// Eventual consistency check.
	if aws.StringValue(dbProxy.DBProxyName) != name {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return dbProxy, nil
}

func FindEventSubscriptionByID(conn *rds.RDS, id string) (*rds.EventSubscription, error) {
	input := &rds.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(id),
	}

	output, err := conn.DescribeEventSubscriptions(input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeSubscriptionNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || len(output.EventSubscriptionsList) == 0 || output.EventSubscriptionsList[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.EventSubscriptionsList[0], nil
}

func FindDBInstanceAutomatedBackupByARN(conn *rds.RDS, arn string) (*rds.DBInstanceAutomatedBackup, error) {
	input := &rds.DescribeDBInstanceAutomatedBackupsInput{
		DBInstanceAutomatedBackupsArn: aws.String(arn),
	}

	output, err := findDBInstanceAutomatedBackup(conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status == InstanceAutomatedBackupStatusRetained {
		// If the automated backup is retained, the replication is stopped.
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.DBInstanceAutomatedBackupsArn) != arn {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBInstanceAutomatedBackup(conn *rds.RDS, input *rds.DescribeDBInstanceAutomatedBackupsInput) (*rds.DBInstanceAutomatedBackup, error) {
	output, err := findDBInstanceAutomatedBackups(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func findDBInstanceAutomatedBackups(conn *rds.RDS, input *rds.DescribeDBInstanceAutomatedBackupsInput) ([]*rds.DBInstanceAutomatedBackup, error) {
	var output []*rds.DBInstanceAutomatedBackup

	err := conn.DescribeDBInstanceAutomatedBackupsPages(input, func(page *rds.DescribeDBInstanceAutomatedBackupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBInstanceAutomatedBackups {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceAutomatedBackupNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
