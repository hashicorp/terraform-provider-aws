package lightsail

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusContainerService(ctx context.Context, conn *lightsail.Lightsail, serviceName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		containerService, err := FindContainerServiceByName(ctx, conn, serviceName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return containerService, aws.StringValue(containerService.State), nil
	}
}

func statusContainerServiceDeploymentVersion(ctx context.Context, conn *lightsail.Lightsail, serviceName string, version int) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		deployment, err := FindContainerServiceDeploymentByVersion(ctx, conn, serviceName, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return deployment, aws.StringValue(deployment.State), nil
	}
}
