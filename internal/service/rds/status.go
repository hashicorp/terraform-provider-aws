package rds

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// ProxyEndpoint NotFound
	proxyEndpointStatusNotFound = "NotFound"

	// ProxyEndpoint Unknown
	proxyEndpointStatusUnknown = "Unknown"
)

func statusEventSubscription(conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEventSubscriptionByID(conn, id)

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
func statusDBProxyEndpoint(conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBProxyEndpoint(conn, id)

		if err != nil {
			return nil, proxyEndpointStatusUnknown, err
		}

		if output == nil {
			return nil, proxyEndpointStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusDBClusterRole(conn *rds.RDS, dbClusterID, roleARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBClusterRoleByDBClusterIDAndRoleARN(conn, dbClusterID, roleARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusDBInstance(conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBInstanceByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DBInstanceStatus), nil
	}
}

func statusDBClusterActivityStream(conn *rds.RDS, dbClusterArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBClusterWithActivityStream(conn, dbClusterArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.ActivityStreamStatus), nil
	}
}

func statusDBInstanceAutomatedBackup(conn *rds.RDS, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBInstanceAutomatedBackupByARN(conn, arn)

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
func statusDBInstanceHasAutomatedBackup(conn *rds.RDS, dbInstanceID, dbInstanceAutomatedBackupsARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBInstanceByID(conn, dbInstanceID)

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

func statusDBProxy(conn *rds.RDS, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBProxyByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
