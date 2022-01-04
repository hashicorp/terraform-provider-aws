package devicefarm

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDevicepoolByArn(conn *devicefarm.DeviceFarm, arn string) (*devicefarm.DevicePool, error) {

	input := &devicefarm.GetDevicePoolInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetDevicePool(input)

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DevicePool == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DevicePool, nil
}

func FindProjectByArn(conn *devicefarm.DeviceFarm, arn string) (*devicefarm.Project, error) {

	input := &devicefarm.GetProjectInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetProject(input)

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Project == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Project, nil
}
