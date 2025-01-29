// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_bedrockagent_agent", sweepAgents)
	awsv2.Register("aws_bedrockagent_data_source", sweepDataSources)
	awsv2.Register("aws_bedrockagent_knowledge_base", sweepKnowledgeBases, "aws_bedrockagent_agent", "aws_bedrockagent_data_source")
}

func sweepAgents(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &bedrockagent.ListAgentsInput{}
	conn := client.BedrockAgentClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := bedrockagent.NewListAgentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AgentSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newAgentResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.AgentId)), framework.NewAttribute("skip_resource_in_use_check", true)))
		}
	}

	return sweepResources, nil
}

func sweepDataSources(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &bedrockagent.ListKnowledgeBasesInput{}
	conn := client.BedrockAgentClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := bedrockagent.NewListKnowledgeBasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.KnowledgeBaseSummaries {
			input := &bedrockagent.ListDataSourcesInput{
				KnowledgeBaseId: v.KnowledgeBaseId,
			}

			pages := bedrockagent.NewListDataSourcesPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return nil, err
				}

				for _, v := range page.DataSourceSummaries {
					sweepResources = append(sweepResources, framework.NewSweepResource(newDataSourceResource, client,
						framework.NewAttribute("data_source_id", aws.ToString(v.DataSourceId)), framework.NewAttribute("knowledge_base_id", aws.ToString(v.KnowledgeBaseId))))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepKnowledgeBases(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &bedrockagent.ListKnowledgeBasesInput{}
	conn := client.BedrockAgentClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := bedrockagent.NewListKnowledgeBasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.KnowledgeBaseSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newKnowledgeBaseResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.KnowledgeBaseId))))
		}
	}

	return sweepResources, nil
}
