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
	awsv2.Register("aws_bedrockagentcore_agent_runtime", sweepAgentRuntimes)
	awsv2.Register("aws_bedrockagentcore_oauth2_credential_provider", sweepOAuth2CredentialProviders)
	awsv2.Register("aws_bedrockagentcore_gateway", sweepGateways)
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
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceAgentRuntime, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.AgentRuntimeId))),
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
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceOAuth2CredentialProvider, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name))),
			)
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
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceGateway, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.GatewayId))),
			)
		}
	}

	return sweepResources, nil
}
