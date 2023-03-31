package synthetics

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCanaryByName(ctx context.Context, conn *synthetics.Synthetics, name string) (*synthetics.Canary, error) {
	input := &synthetics.GetCanaryInput{
		Name: aws.String(name),
	}

	output, err := conn.GetCanaryWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, synthetics.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Canary == nil || output.Canary.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Canary, nil
}
