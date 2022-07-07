package rum

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAppMonitorByName(conn *cloudwatchrum.CloudWatchRUM, name string) (*cloudwatchrum.AppMonitor, error) {
	input := cloudwatchrum.GetAppMonitorInput{
		Name: aws.String(name),
	}

	output, err := conn.GetAppMonitor(&input)

	if tfawserr.ErrCodeEquals(err, cloudwatchrum.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AppMonitor == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AppMonitor, nil
}
