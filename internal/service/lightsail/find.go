package lightsail

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindContainerServiceByName(ctx context.Context, conn *lightsail.Lightsail, serviceName string) (*lightsail.ContainerService, error) {
	input := &lightsail.GetContainerServicesInput{
		ServiceName: aws.String(serviceName),
	}

	output, err := conn.GetContainerServicesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ContainerServices) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ContainerServices); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ContainerServices[0], nil
}

func FindContainerServiceDeploymentByVersion(ctx context.Context, conn *lightsail.Lightsail, serviceName string, version int) (*lightsail.ContainerServiceDeployment, error) {
	input := &lightsail.GetContainerServiceDeploymentsInput{
		ServiceName: aws.String(serviceName),
	}

	output, err := conn.GetContainerServiceDeploymentsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Deployments) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	var result *lightsail.ContainerServiceDeployment

	for _, deployment := range output.Deployments {
		if deployment == nil {
			continue
		}

		if int(aws.Int64Value(deployment.Version)) == version {
			result = deployment
			break
		}
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return result, nil
}
