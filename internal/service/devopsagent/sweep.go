// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_devopsagent_agent_space", sweepAgentSpaces)
}

func sweepAgentSpaces(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DevOpsAgentClient(ctx)
	input := devopsagent.ListAgentSpacesInput{}
	var sweepResources []sweep.Sweepable

	pages := devopsagent.NewListAgentSpacesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.AgentSpaces {
			sweepResources = append(sweepResources, framework.NewSweepResource(newAgentSpaceResource, client,
				framework.NewAttribute("agent_space_id", aws.ToString(v.AgentSpaceId))))
		}
	}

	return sweepResources, nil
}
