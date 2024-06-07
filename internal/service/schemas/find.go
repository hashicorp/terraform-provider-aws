// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindSchemaByNameAndRegistryName(ctx context.Context, conn *schemas.Schemas, name, registryName string) (*schemas.DescribeSchemaOutput, error) {
	input := &schemas.DescribeSchemaInput{
		RegistryName: aws.String(registryName),
		SchemaName:   aws.String(name),
	}

	output, err := conn.DescribeSchemaWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
