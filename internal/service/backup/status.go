package backup

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func statusFramework(conn *backup.Backup, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &backup.DescribeFrameworkInput{
			FrameworkName: aws.String(id),
		}

		output, err := conn.DescribeFramework(input)

		if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
			return output, backup.ErrCodeResourceNotFoundException, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DeploymentStatus), nil
	}
}
