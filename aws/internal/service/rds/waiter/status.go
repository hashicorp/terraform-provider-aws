package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	// ProxyEndpoint NotFound
	ProxyEndpointStatusNotFound = "NotFound"

	// ProxyEndpoint Unknown
	ProxyEndpointStatusUnknown = "Unknown"
)

func EventSubscriptionStatus(conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.EventSubscriptionByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// DBProxyEndpointStatus fetches the ProxyEndpoint and its Status
func DBProxyEndpointStatus(conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.DBProxyEndpoint(conn, id)

		if err != nil {
			return nil, ProxyEndpointStatusUnknown, err
		}

		if output == nil {
			return nil, ProxyEndpointStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func DBClusterRoleStatus(conn *rds.RDS, dbClusterID, roleARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.DBClusterRoleByDBClusterIDAndRoleARN(conn, dbClusterID, roleARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
