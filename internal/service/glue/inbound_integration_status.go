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
)

func statusInboundIntegration(ctx context.Context, conn *glue.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := glue.DescribeInboundIntegrationsInput{IntegrationArn: aws.String(arn)}
		output, err := conn.DescribeInboundIntegrations(ctx, &input)

		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil, "NotFound", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.InboundIntegrations) == 0 {
			return nil, "NotFound", nil
		}

		v := output.InboundIntegrations[0]
		status := string(v.Status)
		return &v, status, nil
	}
}
