// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findInboundIntegrationByARN(ctx context.Context, conn *glue.Client, arn string) (*awstypes.InboundIntegration, error) {
	input := glue.DescribeInboundIntegrationsInput{
		IntegrationArn: aws.String(arn),
	}

	output, err := conn.DescribeInboundIntegrations(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.InboundIntegrations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	v := output.InboundIntegrations[0]
	return &v, nil
}
