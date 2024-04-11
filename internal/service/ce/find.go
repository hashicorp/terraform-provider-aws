// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCostAllocationTagByKey(ctx context.Context, conn *costexplorer.Client, key string) (*awstypes.CostAllocationTag, error) {
	in := &costexplorer.ListCostAllocationTagsInput{
		TagKeys:    []string{key},
		MaxResults: aws.Int32(1),
	}

	out, err := conn.ListCostAllocationTags(ctx, in)

	if errs.IsA[*awstypes.UnknownMonitorException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.CostAllocationTags) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.CostAllocationTags[0], nil
}

func FindCostCategoryByARN(ctx context.Context, conn *costexplorer.Client, arn string) (*awstypes.CostCategory, error) {
	in := &costexplorer.DescribeCostCategoryDefinitionInput{
		CostCategoryArn: aws.String(arn),
	}

	out, err := conn.DescribeCostCategoryDefinition(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.CostCategory == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.CostCategory, nil
}
