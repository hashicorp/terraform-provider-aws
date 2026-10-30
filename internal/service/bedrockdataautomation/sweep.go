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
	awsv2.Register("aws_bedrockdataautomation_project", sweepProjects)
}

func sweepBlueprints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockdataautomation.ListBlueprintsInput{
		ResourceOwner: awstypes.ResourceOwnerAccount,
	}
	conn := client.BedrockDataAutomationClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listBlueprints(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newBlueprintResource, client,
			framework.NewAttribute(names.AttrARN, aws.ToString(v.BlueprintArn))),
		)
	}

	return sweepResources, nil
}

func sweepProjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockdataautomation.ListDataAutomationProjectsInput{
		ResourceOwner: awstypes.ResourceOwnerAccount,
	}
	conn := client.BedrockDataAutomationClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listProjects(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newProjectResource, client,
			framework.NewAttribute(names.AttrARN, aws.ToString(v.ProjectArn))),
		)
	}

	return sweepResources, nil
}
