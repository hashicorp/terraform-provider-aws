package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// AccessPointStatus fetches the Access Point and its Status
func AccessPointStatus(conn *efs.EFS, accessPointId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &efs.DescribeAccessPointsInput{
			AccessPointId: aws.String(accessPointId),
		}

		output, err := conn.DescribeAccessPoints(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil && len(output.AccessPoints) == 0 {
			return nil, "", nil
		}

		mt := output.AccessPoints[0]

		return mt, aws.StringValue(mt.LifeCycleState), nil
	}
}
