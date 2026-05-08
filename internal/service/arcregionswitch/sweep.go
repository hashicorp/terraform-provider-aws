// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_arcregionswitch_plan", sweepPlans)
}

func sweepPlans(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ARCRegionSwitchClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := arcregionswitch.NewListPlansPaginator(conn, &arcregionswitch.ListPlansInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, plan := range page.Plans {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourcePlan, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(plan.Arn)),
			))
		}
	}

	return sweepResources, nil
}
