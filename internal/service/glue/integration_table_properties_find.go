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

func findIntegrationTableProperties(ctx context.Context, conn *glue.Client, resourceArn, tableName string) (*glue.GetIntegrationTablePropertiesOutput, error) {
	input := glue.GetIntegrationTablePropertiesInput{
		ResourceArn: aws.String(resourceArn),
		TableName:   aws.String(tableName),
	}

	output, err := conn.GetIntegrationTableProperties(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{LastError: err, LastRequest: input}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
