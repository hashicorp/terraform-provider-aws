package lightsail

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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
