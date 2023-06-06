package sts

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCallerIdentity(ctx context.Context, conn *sts.STS) (*sts.GetCallerIdentityOutput, error) {
	input := &sts.GetCallerIdentityInput{}

	output, err := conn.GetCallerIdentityWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
