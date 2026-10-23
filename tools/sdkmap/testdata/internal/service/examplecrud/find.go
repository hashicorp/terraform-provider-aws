// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplecrud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplecrud"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplecrud/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findFooByID(ctx context.Context, conn *examplecrud.Client, id string) (*awstypes.Foo, error) {
	input := examplecrud.DescribeFooInput{
		FooId: aws.String(id),
	}

	output, err := conn.DescribeFoo(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Foo == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Foo, nil
}

func findBarByID(ctx context.Context, conn *examplecrud.Client, id string) (*awstypes.Bar, error) {
	input := examplecrud.DescribeBarInput{
		BarId: aws.String(id),
	}

	output, err := conn.DescribeBar(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Bar == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Bar, nil
}
