package devicefarm

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDevicePoolByARN(conn *devicefarm.DeviceFarm, arn string) (*devicefarm.DevicePool, error) {

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

func FindProjectByARN(conn *devicefarm.DeviceFarm, arn string) (*devicefarm.Project, error) {

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

func FindUploadByARN(conn *devicefarm.DeviceFarm, arn string) (*devicefarm.Upload, error) {

	input := &devicefarm.GetUploadInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetUpload(input)

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Upload == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Upload, nil
}

func FindNetworkProfileByARN(conn *devicefarm.DeviceFarm, arn string) (*devicefarm.NetworkProfile, error) {

	input := &devicefarm.GetNetworkProfileInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetNetworkProfile(input)

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.NetworkProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.NetworkProfile, nil
}

func FindInstanceProfileByARN(conn *devicefarm.DeviceFarm, arn string) (*devicefarm.InstanceProfile, error) {

	input := &devicefarm.GetInstanceProfileInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetInstanceProfile(input)

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.InstanceProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.InstanceProfile, nil
}

func FindTestGridProjectByARN(conn *devicefarm.DeviceFarm, arn string) (*devicefarm.TestGridProject, error) {

	input := &devicefarm.GetTestGridProjectInput{
		ProjectArn: aws.String(arn),
	}
	output, err := conn.GetTestGridProject(input)

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TestGridProject == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TestGridProject, nil
}
