package ssmincidents

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindReplicationSetByID(context context.Context, client *ssmincidents.Client, arn string) (*types.ReplicationSet, error) {
	in := &ssmincidents.GetReplicationSetInput{
		Arn: aws.String(arn),
	}
	out, err := client.GetReplicationSet(context, in)
	if err != nil {
		var notFoundError *types.ResourceNotFoundException
		if errors.As(err, &notFoundError) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ReplicationSet == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ReplicationSet, nil
}
