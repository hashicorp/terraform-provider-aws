package evidently

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func statusProject(conn *cloudwatchevidently.CloudWatchEvidently, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindProjectByNameOrARN(context.Background(), conn, id)

		if tfawserr.ErrCodeEquals(err, cloudwatchevidently.ErrCodeResourceNotFoundException) {
			return output, cloudwatchevidently.ErrCodeResourceNotFoundException, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
