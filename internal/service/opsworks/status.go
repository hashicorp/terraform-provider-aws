package opsworks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func InstanceStatus(conn *opsworks.OpsWorks, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeInstances(&opsworks.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(instanceID)},
		})

		if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if resp == nil || len(resp.Instances) == 0 || resp.Instances[0] == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		i := resp.Instances[0]
		return i, aws.StringValue(i.Status), nil
	}
}
