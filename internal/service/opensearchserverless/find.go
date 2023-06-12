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

func FindAccessPolicyByNameAndType(ctx context.Context, conn *opensearchserverless.Client, id, policyType string) (*types.AccessPolicyDetail, error) {
	in := &opensearchserverless.GetAccessPolicyInput{
		Name: aws.String(id),
		Type: types.AccessPolicyType(policyType),
	}
	out, err := conn.GetAccessPolicy(ctx, in)
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

	if out == nil || out.AccessPolicyDetail == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.AccessPolicyDetail, nil
}
