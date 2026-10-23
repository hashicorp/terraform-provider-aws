// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_batch_job_definition")
func newJobDefinitionResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceJobDefinition(), framework.SDKListSpec[awstypes.JobDefinition, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.JobDefinition, error], diag.Diagnostics) {
			var input batch.DescribeJobDefinitionsInput

			// Deregistered revisions linger in the API but are not resources
			// anyone can manage.
			return tfiter.FilterValues(listJobDefinitions(ctx, c.BatchClient(ctx), &input), func(jd awstypes.JobDefinition) bool {
				return aws.ToString(jd.Status) != jobDefinitionStatusInactive
			}), nil
		},
		ID: func(jd awstypes.JobDefinition, _ framework.RegionQuery) string {
			return aws.ToString(jd.JobDefinitionArn)
		},
		Flatten: resourceJobDefinitionFlatten,
		DisplayName: func(jd awstypes.JobDefinition) string {
			return aws.ToString(jd.JobDefinitionName)
		},
	})
}

func listJobDefinitions(ctx context.Context, conn *batch.Client, input *batch.DescribeJobDefinitionsInput) iter.Seq2[awstypes.JobDefinition, error] {
	return func(yield func(awstypes.JobDefinition, error) bool) {
		pages := batch.NewDescribeJobDefinitionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.JobDefinition{}, fmt.Errorf("listing Batch Job Definitions: %w", err))
				return
			}

			for _, item := range page.JobDefinitions {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
