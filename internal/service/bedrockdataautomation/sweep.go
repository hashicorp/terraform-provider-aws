// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockdataautomation

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_bedrockdataautomation_blueprint", sweepBlueprints)
}

func sweepBlueprints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockdataautomation.ListBlueprintsInput{
		ResourceOwner: awstypes.ResourceOwnerAccount,
	}
	conn := client.BedrockDataAutomationClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockdataautomation.NewListBlueprintsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Blueprints {
			sweepResources = append(sweepResources, framework.NewSweepResource(newBlueprintResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.BlueprintArn))),
			)
		}
	}

	return sweepResources, nil
}
