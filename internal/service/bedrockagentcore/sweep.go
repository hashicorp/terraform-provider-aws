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
	awsv2.Register("aws_bedrockagentcore_workload_identity", sweepWorkloadIdentities)
	awsv2.Register("aws_bedrockagentcore_code_interpreter", sweepCodeInterpreters)
	awsv2.Register("aws_bedrockagentcore_browser", sweepBrowsers)
	awsv2.Register("aws_bedrockagentcore_api_key_credential_provider", sweepAPIKeyCredentialProviders)
	awsv2.Register("aws_bedrockagentcore_oauth2_credential_provider", sweepOAuth2CredentialProviders)
	awsv2.Register("aws_bedrockagentcore_gateway", sweepGateways, "aws_bedrockagentcore_gateway_target")
	awsv2.Register("aws_bedrockagentcore_gateway_target", sweepGatewayTargets)
	awsv2.Register("aws_bedrockagentcore_memory", sweepMemories)
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

func sweepBrowsers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListBrowsersInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListBrowsersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.BrowserSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newBrowserResource, client,
				framework.NewAttribute("browser_id", aws.ToString(v.BrowserId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepAPIKeyCredentialProviders(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListApiKeyCredentialProvidersInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListApiKeyCredentialProvidersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.CredentialProviders {
			sweepResources = append(sweepResources, framework.NewSweepResource(newAPIKeyCredentialProviderResource, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name))),
			)
		}
	}

	return sweepResources, nil
}

func sweepOAuth2CredentialProviders(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListOauth2CredentialProvidersInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListOauth2CredentialProvidersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.CredentialProviders {
			sweepResources = append(sweepResources, framework.NewSweepResource(newOAuth2CredentialProviderResource, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name))),
			)
		}
	}

	return sweepResources, nil
}

func sweepCodeInterpreters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListCodeInterpretersInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListCodeInterpretersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.CodeInterpreterSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newCodeInterpreterResource, client,
				framework.NewAttribute("code_interpreter_id", aws.ToString(v.CodeInterpreterId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepWorkloadIdentities(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListWorkloadIdentitiesInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListWorkloadIdentitiesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.WorkloadIdentities {
			sweepResources = append(sweepResources, framework.NewSweepResource(newWorkloadIdentityResource, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name))),
			)
		}
	}

	return sweepResources, nil
}

func sweepMemories(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListMemoriesInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListMemoriesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Memories {
			sweepResources = append(sweepResources, framework.NewSweepResource(newMemoryResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id))),
			)
		}
	}

	return sweepResources, nil
}
