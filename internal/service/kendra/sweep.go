//go:build sweep
// +build sweep

package kendra

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_kendra_index", &resource.Sweeper{
		Name: "aws_kendra_index",
		F:    sweepIndex,
	})
}

func sweepIndex(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}

	ctx := context.Background()
	conn := client.(*conns.AWSClient).KendraConn
	sweepResources := make([]*sweep.SweepResource, 0)
	in := &kendra.ListIndicesInput{}
	var errs *multierror.Error

	pages := kendra.NewListIndicesPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if sweep.SkipSweepError(err) {
			log.Println("[WARN] Skipping Kendra Indices sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving Kendra Indices: %w", err)
		}

		for _, index := range page.IndexConfigurationSummaryItems {
			id := aws.ToString(index.Id)
			log.Printf("[INFO] Deleting Kendra Index: %s", id)

			r := ResourceIndex()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Kendra Indices for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Kendra Indices sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
