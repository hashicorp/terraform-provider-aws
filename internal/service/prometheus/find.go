package prometheus

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAlertManagerDefinitionByID(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	input := &prometheusservice.DescribeAlertManagerDefinitionInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeAlertManagerDefinitionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AlertManagerDefinition == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AlertManagerDefinition, nil
}
