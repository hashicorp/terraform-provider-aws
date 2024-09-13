// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_ram_resource_share", &resource.Sweeper{
		Name: "aws_ram_resource_share",
		F:    sweepResourceShares,
	})
}

func sweepResourceShares(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RAMClient(ctx)
	input := &ram.GetResourceSharesInput{
		ResourceOwner: awstypes.ResourceOwnerSelf,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ram.NewGetResourceSharesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RAM Resource Share sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing RAM Resource Shares (%s): %w", region, err)
		}

		for _, v := range page.ResourceShares {
			if v.Status == awstypes.ResourceShareStatusDeleted {
				continue
			}

			r := resourceResourceShare()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ResourceShareArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RAM Resource Shares (%s): %w", region, err)
	}

	return nil
}
