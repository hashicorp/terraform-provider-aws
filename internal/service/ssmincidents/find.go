package ssmincidents

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindReplicationSetByID(context context.Context, client *ssmincidents.Client, arn string) (*types.ReplicationSet, error) {
	input := &ssmincidents.GetReplicationSetInput{
		Arn: aws.String(arn),
	}
	output, err := client.GetReplicationSet(context, input)
	if err != nil {
		var notFoundError *types.ResourceNotFoundException
		if errors.As(err, &notFoundError) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if output == nil || output.ReplicationSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ReplicationSet, nil
}
