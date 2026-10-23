// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_devopsagent_agent_space", sweepAgentSpaces, "aws_devopsagent_asset")
	awsv2.Register("aws_devopsagent_asset", sweepAssets)
}

func sweepAssets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DevOpsAgentClient(ctx)
	var sweepResources []sweep.Sweepable
	pages := devopsagent.NewListAgentSpacesPaginator(conn, &devopsagent.ListAgentSpacesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}
		for _, space := range page.AgentSpaces {
			input := devopsagent.ListAssetsInput{AgentSpaceId: space.AgentSpaceId}
			assets := devopsagent.NewListAssetsPaginator(conn, &input)
			for assets.HasMorePages() {
				page, err := assets.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}
				for _, asset := range page.Items {
					sweepResources = append(sweepResources, framework.NewSweepResource(newAssetResource, client,
						framework.NewAttribute("agent_space_id", aws.ToString(space.AgentSpaceId)),
						framework.NewAttribute("asset_id", aws.ToString(asset.AssetId))))
				}
			}
		}
	}
	return sweepResources, nil
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
