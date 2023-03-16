//go:build sweep
// +build sweep

package resourceexplorer2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_resourceexplorer2_index", &resource.Sweeper{
		Name: "aws_resourceexplorer2_index",
		F:    sweepIndexes,
	})
}

func sweepIndexes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ResourceExplorer2Client()
	input := &resourceexplorer2.ListIndexesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := resourceexplorer2.NewListIndexesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Resource Explorer Index sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Resource Explorer Indexes (%s): %w", region, err)
		}

		for _, v := range page.Indexes {
			sweepResources = append(sweepResources, sweep.NewSweepFrameworkResource(newResourceIndex, aws.ToString(v.Arn), client))
		}
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Resource Explorer Indexes (%s): %w", region, err)
	}

	return nil
}
