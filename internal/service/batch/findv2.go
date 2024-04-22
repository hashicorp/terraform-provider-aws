// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ListJobDefinitionsV2ByNameWithStatus(ctx context.Context, conn *batch.Client, input *batch.DescribeJobDefinitionsInput) ([]types.JobDefinition, error) {
	var out []types.JobDefinition

	pages := batch.NewDescribeJobDefinitionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}
		out = append(out, page.JobDefinitions...)
	}

	if len(out) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out, nil
}
