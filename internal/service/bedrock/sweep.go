// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_bedrock_evaluation_job", sweepEvaluationJobs)
}

func sweepEvaluationJobs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &bedrock.ListEvaluationJobsInput{}
	conn := client.BedrockClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := bedrock.NewListEvaluationJobsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.JobSummaries {
			// StopEvaluationJob only supports jobs that are In-Progress.
			// All other status values are terminal states.
			if v.Status == awstypes.EvaluationJobStatusInProgress {
				sweepResources = append(sweepResources, framework.NewSweepResource(newEvaluationJobResource, client,
					framework.NewAttribute("job_arn", aws.ToString(v.JobArn))),
				)
			}
		}
	}

	return sweepResources, nil
}
