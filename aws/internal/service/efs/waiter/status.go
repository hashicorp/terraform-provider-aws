package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/efs/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// AccessPointLifeCycleState fetches the Access Point and its LifecycleState
func AccessPointLifeCycleState(conn *efs.EFS, accessPointId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &efs.DescribeAccessPointsInput{
			AccessPointId: aws.String(accessPointId),
		}

		output, err := conn.DescribeAccessPoints(input)

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

func BackupPolicyStatus(conn *efs.EFS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.BackupPolicyByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func FileSystemLifeCycleState(conn *efs.EFS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FileSystemByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.LifeCycleState), nil
	}
}
