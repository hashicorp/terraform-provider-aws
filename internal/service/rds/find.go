package rds

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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

	return output.DBProxies[0], nil
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
