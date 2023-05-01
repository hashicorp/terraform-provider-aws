package opensearchserverless

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findCollectionByID(ctx context.Context, conn *opensearchserverless.Client, id string) (*types.CollectionDetail, error) {
	in := &opensearchserverless.BatchGetCollectionInput{
		Ids: []string{id},
	}
	out, err := conn.BatchGetCollection(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.CollectionDetails == nil || len(out.CollectionDetails) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.CollectionDetails[0], nil
}
