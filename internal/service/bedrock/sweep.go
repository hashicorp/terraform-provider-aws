// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_bedrock_evaluation_job", sweepEvaluationJobs)
	awsv2.Register("aws_bedrock_guardrail_resource_policy", sweepGuardrailResourcePolicies)
	awsv2.Register("aws_bedrock_model_invocation_job", sweepModelInvocationJobs)
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

func sweepModelInvocationJobs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &bedrock.ListModelInvocationJobsInput{}
	conn := client.BedrockClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := bedrock.NewListModelInvocationJobsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.InvocationJobSummaries {
			// StopModelInvocationJob only supports jobs that are In-Progress.
			// All other status values are terminal states.
			if v.Status == awstypes.ModelInvocationJobStatusInProgress {
				sweepResources = append(sweepResources, framework.NewSweepResource(newEvaluationJobResource, client,
					framework.NewAttribute("job_arn", aws.ToString(v.JobArn))),
				)
			}
		}
	}

	return sweepResources, nil
}

func sweepGuardrailResourcePolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.BedrockClient(ctx)
	var sweepResources []sweep.Sweepable

	resourceARNs := make([]string, 0)
	var input bedrock.ListGuardrailsInput
	for item, err := range listGuardrails(ctx, conn, &input) {
		if err != nil {
			return nil, err
		}

		resourceARNs = append(resourceARNs, aws.ToString(item.Arn))
	}

	// The serialized guardrail profile acceptance test attaches a policy to the "us" guardrail profile.
	if strings.HasPrefix(client.Region(ctx), "us-") {
		resourceARNs = append(resourceARNs, client.RegionalARN(ctx, "bedrock", "guardrail-profile/us.guardrail.v1:0"))
	}

	for _, resourceARN := range resourceARNs {
		_, err := findGuardrailResourcePolicyByARN(ctx, conn, resourceARN)
		if retry.NotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newGuardrailResourcePolicyResource, client,
			framework.NewAttribute(names.AttrResourceARN, resourceARN)),
		)
	}

	return sweepResources, nil
}
