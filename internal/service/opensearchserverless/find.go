package opensearchserverless

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findSecurityConfigByID(ctx context.Context, conn *opensearchserverless.Client, id string) (*types.SecurityConfigDetail, error) {
	in := &opensearchserverless.GetSecurityConfigInput{
		Id: aws.String(id),
	}
	out, err := conn.GetSecurityConfig(ctx, in)
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

	if out == nil || out.SecurityConfigDetail == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.SecurityConfigDetail, nil
}
