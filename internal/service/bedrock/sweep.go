// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_bedrock_guardrail", sweepGuardrails)
	awsv2.Register("aws_bedrock_inference_profile", sweepInferenceProfiles)
	awsv2.Register("aws_bedrock_prompt_router", sweepPromptRouters)
}

func sweepGuardrails(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &bedrock.ListGuardrailsInput{}
	conn := client.BedrockClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := bedrock.NewListGuardrailsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Guardrails {
			sweepResources = append(sweepResources, framework.NewSweepResource(newGuardrailResource, client,
				framework.NewAttribute("guardrail_id", aws.ToString(v.Id))))
		}
	}

	return sweepResources, nil
}

func sweepInferenceProfiles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &bedrock.ListInferenceProfilesInput{}
	conn := client.BedrockClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := bedrock.NewListInferenceProfilesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.InferenceProfileSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newInferenceProfileResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.InferenceProfileId))))
		}
	}

	return sweepResources, nil
}

func sweepPromptRouters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &bedrock.ListPromptRoutersInput{}
	conn := client.BedrockClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := bedrock.NewListPromptRoutersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.PromptRouterSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newPromptRouterResource, client,
				framework.NewAttribute("prompt_router_arn", aws.ToString(v.PromptRouterArn))))
		}
	}

	return sweepResources, nil
}
