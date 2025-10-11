// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_bedrockagentcore_agent_runtime", sweepAgentRuntimes, "aws_bedrockagentcore_agent_runtime_endpoint")
	awsv2.Register("aws_bedrockagentcore_agent_runtime_endpoint", sweepAgentRuntimeEndpoints)
	awsv2.Register("aws_bedrockagentcore_gateway", sweepGateways, "aws_bedrockagentcore_gateway_target")
	awsv2.Register("aws_bedrockagentcore_gateway_target", sweepGatewayTargets)
}

func sweepAgentRuntimes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListAgentRuntimesInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListAgentRuntimesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.AgentRuntimes {
			sweepResources = append(sweepResources, framework.NewSweepResource(newAgentRuntimeResource, client,
				framework.NewAttribute("agent_runtime_id", aws.ToString(v.AgentRuntimeId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepAgentRuntimeEndpoints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListAgentRuntimesInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListAgentRuntimesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.AgentRuntimes {
			agentRuntimeID := aws.ToString(v.AgentRuntimeId)
			input := bedrockagentcorecontrol.ListAgentRuntimeEndpointsInput{
				AgentRuntimeId: aws.String(agentRuntimeID),
			}

			pages := bedrockagentcorecontrol.NewListAgentRuntimeEndpointsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.RuntimeEndpoints {
					sweepResources = append(sweepResources, framework.NewSweepResource(newAgentRuntimeEndpointResource, client,
						framework.NewAttribute("agent_runtime_id", agentRuntimeID),
						framework.NewAttribute(names.AttrName, aws.ToString(v.Name)),
					),
					)
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepGateways(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListGatewaysInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListGatewaysPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newGatewayResource, client,
				framework.NewAttribute("gateway_id", aws.ToString(v.GatewayId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepGatewayTargets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListGatewaysInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListGatewaysPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			gatewayID := aws.ToString(v.GatewayId)
			input := bedrockagentcorecontrol.ListGatewayTargetsInput{
				GatewayIdentifier: aws.String(gatewayID),
			}

			pages := bedrockagentcorecontrol.NewListGatewayTargetsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.Items {
					sweepResources = append(sweepResources, framework.NewSweepResource(newGatewayTargetResource, client,
						framework.NewAttribute("gateway_identifier", gatewayID),
						framework.NewAttribute("target_id", aws.ToString(v.TargetId))),
					)
				}
			}
		}
	}

	return sweepResources, nil
}
