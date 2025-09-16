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
	awsv2.Register("aws_bedrockagentcore_gateway", sweepGateways)
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
			gatewayIdentifier := aws.ToString(v.GatewayId)
			sweepTargets, err := sweepGatewayTargets(ctx, client, gatewayIdentifier)
			if err != nil {
				return nil, smarterr.NewError(err)
			}
			sweepResources = append(sweepResources, sweepTargets...)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceGateway, client,
				framework.NewAttribute("gateway_id", gatewayIdentifier)),
			)
		}
	}

	return sweepResources, nil
}

func sweepGatewayTargets(ctx context.Context, client *conns.AWSClient, gatewayIdentifier string) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListGatewayTargetsInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListGatewayTargetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceGatewayTarget, client,
				framework.NewAttribute("gateway_identifier", gatewayIdentifier), framework.NewAttribute(names.AttrID, aws.ToString(v.TargetId))),
			)
		}
	}

	return sweepResources, nil
}
