// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockframework

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockframework"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockframework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findFooByID(ctx context.Context, conn *mockframework.Client, id string) (*awstypes.Foo, error) {
	input := mockframework.DescribeFooInput{
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

func findQuotaByName(ctx context.Context, conn *mockframework.Client, name string) (*awstypes.Quota, error) {
	input := mockframework.GetQuotaInput{
		QuotaName: aws.String(name),
	}

	output, err := conn.GetQuota(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Quota == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Quota, nil
}

func findBarByID(ctx context.Context, conn *mockframework.Client, id string) (*awstypes.Bar, error) {
	input := mockframework.GetBarInput{
		BarId: aws.String(id),
	}

	output, err := conn.GetBar(ctx, &input)

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

func findBarByName(ctx context.Context, conn *mockframework.Client, name string) (*awstypes.BarSummary, error) {
	var input mockframework.ListBarsInput

	pages := mockframework.NewListBarsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, bar := range page.Bars {
			if aws.ToString(bar.Name) == name {
				return &bar, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

func findBazByID(ctx context.Context, conn *mockframework.Client, id string) (*awstypes.Baz, error) {
	input := mockframework.DescribeBazInput{
		BazId: aws.String(id),
	}

	output, err := conn.DescribeBaz(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Baz == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Baz, nil
}

func findQuxByID(ctx context.Context, conn *mockframework.Client, id string) (*awstypes.Qux, error) {
	input := mockframework.DescribeQuxInput{
		QuxId: aws.String(id),
	}

	output, err := conn.DescribeQux(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Qux == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Qux, nil
}
