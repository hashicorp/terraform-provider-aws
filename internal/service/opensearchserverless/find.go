package opensearchserverless

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findVPCEndpointByID(ctx context.Context, conn *opensearchserverless.Client, id string) (*types.VpcEndpointDetail, error) {
	in := &opensearchserverless.BatchGetVpcEndpointInput{
		Ids: []string{id},
	}
	out, err := conn.BatchGetVpcEndpoint(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.VpcEndpointDetails == nil || len(out.VpcEndpointDetails) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.VpcEndpointDetails[0], nil
}
